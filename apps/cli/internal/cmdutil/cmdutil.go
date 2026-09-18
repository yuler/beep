package cmdutil

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	"beep/internal/client"
	"beep/internal/config"
	"beep/internal/ui"

	"github.com/spf13/cobra"
)

var (
	overrideWorkspace string
	WorkspaceHook     func() string
)

// SetOverrideWorkspace sets an explicit workspace override (useful for tests).
func SetOverrideWorkspace(w string) {
	overrideWorkspace = w
}

// GetWorkspace gets the active workspace from flags, override, or environment.
func GetWorkspace(cmd *cobra.Command) string {
	if overrideWorkspace != "" {
		return overrideWorkspace
	}
	if WorkspaceHook != nil {
		if w := WorkspaceHook(); w != "" {
			return w
		}
	}
	if cmd != nil {
		if f := cmd.Flag("workspace"); f != nil && f.Value.String() != "" {
			return f.Value.String()
		}
		if cmd.Root() != nil && cmd.Root().PersistentFlags().Lookup("workspace") != nil {
			val, _ := cmd.Root().PersistentFlags().GetString("workspace")
			if val != "" {
				return val
			}
		}
	}
	if env := os.Getenv("BEEP_WORKSPACE"); env != "" {
		return env
	}
	return ""
}

// LoadConfig loads the client configuration taking global persistent flags into account.
func LoadConfig(cmd *cobra.Command) (*config.Config, error) {
	workspace := GetWorkspace(cmd)
	var server, account string
	if cmd != nil && cmd.Root() != nil {
		if cmd.Root().PersistentFlags().Lookup("server") != nil {
			server, _ = cmd.Root().PersistentFlags().GetString("server")
		}
		if cmd.Root().PersistentFlags().Lookup("account") != nil {
			account, _ = cmd.Root().PersistentFlags().GetString("account")
		}
	}

	cfg, err := config.Load(workspace)
	if err != nil {
		return nil, err
	}
	if server != "" {
		cfg.ServerURL = server
	}
	if workspace != "" {
		cfg.Workspace = workspace
	}
	if account != "" {
		cfg.AccountSlug = strings.TrimSpace(account)
	}
	return cfg, nil
}

// EnsureLoggedIn checks if the current configuration has an active session.
func EnsureLoggedIn(ctx context.Context, cfg *config.Config) (*client.MeResponse, error) {
	if !cfg.IsLoggedIn() {
		return nil, fmt.Errorf("not logged in. Please run '%s auth login' first", config.BinaryName())
	}

	c := client.New(cfg)
	verifyCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	me, err := c.GetMe(verifyCtx)
	if err != nil {
		if strings.Contains(err.Error(), "401") || strings.Contains(err.Error(), "invalid or expired") {
			return nil, fmt.Errorf("stored login session is invalid or expired. Please run '%s auth login' first", config.BinaryName())
		}
		// On temporary network errors during pre-flight check, proceed with stored token
		return nil, nil
	}
	return me, nil
}

// RunWithClient wraps config loading, login check, signal interruption, and client initialization.
func RunWithClient(cmd *cobra.Command, fn func(ctx context.Context, cfg *config.Config, c *client.Client) error) error {
	cfg, err := LoadConfig(cmd)
	if err != nil {
		return err
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if _, err := EnsureLoggedIn(ctx, cfg); err != nil {
		return err
	}

	c := client.New(cfg)
	return fn(ctx, cfg, c)
}

// IsJSON returns true if --json flag was passed.
func IsJSON(cmd *cobra.Command) bool {
	v, _ := cmd.Root().PersistentFlags().GetBool("json")
	return v
}

// IsInteractive returns true if interactive mode is permitted.
func IsInteractive(cmd *cobra.Command) bool {
	if IsJSON(cmd) {
		return false
	}
	noInteractive, _ := cmd.Root().PersistentFlags().GetBool("no-interactive")
	if noInteractive {
		return false
	}
	return ui.IsInteractive()
}

// IsHeadless checks if the current shell environment is headless (SSH or no DISPLAY).
func IsHeadless() bool {
	if os.Getenv("SSH_CONNECTION") != "" || os.Getenv("SSH_CLIENT") != "" || os.Getenv("SSH_TTY") != "" {
		return true
	}
	if os.Getenv("DISPLAY") == "" && os.Getenv("WAYLAND_DISPLAY") == "" {
		return true
	}
	return false
}

// ResolveAccountSlug resolves the target account slug from flags, config, or user selection.
func ResolveAccountSlug(me *client.MeResponse, explicitAccount string, cfgAccount string) (string, error) {
	if explicit := strings.TrimSpace(explicitAccount); explicit != "" {
		return explicit, nil
	}
	if cfgAccount != "" {
		return cfgAccount, nil
	}
	if me != nil && len(me.Accounts) > 0 {
		if len(me.Accounts) == 1 {
			return me.Accounts[0].Slug, nil
		}
		if ui.IsInteractive() {
			return ui.PromptAccountSelect(me.Accounts, me.LastAccountSlug)
		}
		return "", fmt.Errorf("account slug is required (set via --account <slug> or BEEP_ACCOUNT)")
	}
	if ui.IsInteractive() {
		return ui.PromptAccountSlug()
	}
	return "", fmt.Errorf("account slug is required (set via --account <slug> or BEEP_ACCOUNT)")
}
