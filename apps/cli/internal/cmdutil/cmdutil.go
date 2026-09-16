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

// LoadConfig loads the client configuration taking global persistent flags into account.
func LoadConfig(cmd *cobra.Command) (*config.Config, error) {
	workspace, _ := cmd.Root().PersistentFlags().GetString("workspace")
	server, _ := cmd.Root().PersistentFlags().GetString("server")
	token, _ := cmd.Root().PersistentFlags().GetString("token")
	account, _ := cmd.Root().PersistentFlags().GetString("account")

	cfg, err := config.Load(workspace)
	if err != nil {
		return nil, err
	}
	if server != "" {
		cfg.ServerURL = server
	}
	if token != "" {
		cfg.RunnerToken = token
		fmt.Fprintln(os.Stderr, ui.Warn("passing --token on the command line exposes it in process lists; prefer config.json or BEEP_RUNNER_TOKEN"))
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
		return nil, fmt.Errorf("not logged in. Please run 'beep auth login' first")
	}

	c := client.New(cfg)
	verifyCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	me, err := c.GetMe(verifyCtx)
	if err != nil {
		if strings.Contains(err.Error(), "401") || strings.Contains(err.Error(), "invalid or expired") {
			return nil, fmt.Errorf("stored login session is invalid or expired. Please run 'beep auth login' first")
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
