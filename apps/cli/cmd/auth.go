package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"time"

	"beep/internal/browser"
	"beep/internal/client"
	"beep/internal/config"
	"beep/internal/ui"

	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate CLI with Beep platform",
}

var (
	flagAuthNoBrowser  bool
	flagAuthClientName string
)

func isHeadless() bool {
	if os.Getenv("SSH_CONNECTION") != "" || os.Getenv("SSH_CLIENT") != "" || os.Getenv("SSH_TTY") != "" {
		return true
	}
	if runtime.GOOS == "linux" && os.Getenv("DISPLAY") == "" && os.Getenv("WAYLAND_DISPLAY") == "" {
		return true
	}
	return false
}

func ensureLoggedIn(ctx context.Context, cfg *config.Config) (*client.MeResponse, error) {
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

func resolveAccountSlug(me *client.MeResponse, explicitAccount string, cfgAccount string) (string, error) {
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

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in to Beep via web browser",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		if cfg.ServerURL == "" {
			return fmt.Errorf("server URL is not configured (set via BEEP_SERVER or 'beep config set server <url>')")
		}

		c := client.New(cfg)
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
		defer cancel()

		clientName := strings.TrimSpace(flagAuthClientName)
		if clientName == "" {
			if h, err := os.Hostname(); err == nil {
				clientName = fmt.Sprintf("CLI on %s", h)
			} else {
				clientName = "Beep CLI"
			}
		}

		fmt.Println(ui.Dim("Requesting device authorization from ") + ui.Cyan(cfg.ServerURL) + ui.Dim("..."))

		authReqCtx, authReqCancel := context.WithTimeout(ctx, 15*time.Second)
		authRes, err := c.RequestCliDeviceAuthorization(authReqCtx, clientName)
		authReqCancel()
		if err != nil {
			return fmt.Errorf("device authorization request failed: %w", err)
		}

		fmt.Println()
		fmt.Printf("! %s %s\n", ui.Bold("First copy your one-time code:"), ui.Yellow(authRes.UserCode))

		verifyURL := authRes.VerificationURIComplete
		if verifyURL == "" {
			verifyURL = authRes.VerificationURI
		}

		fmt.Printf("  %s %s\n", ui.Dim("Open URL:"), ui.Cyan(verifyURL))
		fmt.Println()

		shouldOpenBrowser := !flagAuthNoBrowser && !isHeadless()
		if shouldOpenBrowser {
			_ = browser.Open(verifyURL)
			fmt.Printf("%s Waiting for authorization in web browser...\n", ui.Dim("⏳"))
		} else {
			fmt.Printf("%s Please open the URL above in your browser and authorize the login.\n", ui.Dim("⏳"))
		}

		interval := time.Duration(authRes.Interval) * time.Second
		if interval <= 0 {
			interval = 5 * time.Second
		}

		expiresIn := authRes.ExpiresIn
		if expiresIn <= 0 {
			expiresIn = 900
		}
		deadline := time.Now().Add(time.Duration(expiresIn) * time.Second)

		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(interval):
			}

			if time.Now().After(deadline) {
				return fmt.Errorf("device authorization expired, please run 'beep auth login' again")
			}

			pollCtx, pollCancel := context.WithTimeout(ctx, 10*time.Second)
			tokenRes, err := c.PollCliDeviceToken(pollCtx, authRes.DeviceCode)
			pollCancel()

			if err == nil && tokenRes != nil {
				configPath := config.GetConfigPath(flagWorkspace)
				fc, err := config.LoadFile(configPath)
				if err != nil {
					fc = &config.FileConfig{}
				}

				fc.AccessToken = tokenRes.AccessToken
				fc.UserEmail = tokenRes.User.Email
				fc.UserName = tokenRes.User.Name

				if err := config.SaveFile(configPath, fc); err != nil {
					return fmt.Errorf("failed to save config: %w", err)
				}

				fmt.Println()
				displayName := tokenRes.User.Name
				if displayName == "" {
					displayName = tokenRes.User.Email
				} else {
					displayName = fmt.Sprintf("%s (%s)", displayName, tokenRes.User.Email)
				}

				fmt.Printf("%s %s %s\n",
					ui.Green("✓"),
					ui.Bold("Logged in as:"),
					ui.Cyan(displayName),
				)
				fmt.Printf("%s Config saved to %s\n", ui.Green("✓"), ui.Dim(configPath))
				return nil
			}

			var oauthErr *client.OAuthErrorResponse
			if errors.As(err, &oauthErr) {
				switch oauthErr.ErrorCode {
				case client.OAuthErrAuthorizationPending:
					// Still waiting for approval
					continue
				case client.OAuthErrSlowDown:
					interval += 5 * time.Second
					continue
				case client.OAuthErrAccessDenied:
					return fmt.Errorf("authorization was denied in the web browser")
				case client.OAuthErrExpiredToken:
					return fmt.Errorf("authorization request expired, please run 'beep auth login' again")
				default:
					return fmt.Errorf("authorization failed: %s", oauthErr.Error())
				}
			}

			if err != nil {
				// Temporary network error while polling; retry
				continue
			}
		}
	},
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Log out of Beep on this machine",
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath := config.GetConfigPath(flagWorkspace)
		fc, err := config.LoadFile(configPath)
		if err != nil || fc.AccessToken == "" {
			fmt.Println(ui.Dim("No active login session found in " + configPath))
			return nil
		}

		userEmail := fc.UserEmail
		fc.AccessToken = ""
		fc.UserEmail = ""
		fc.UserName = ""

		if err := config.SaveFile(configPath, fc); err != nil {
			return fmt.Errorf("failed to update config file: %w", err)
		}

		if userEmail != "" {
			fmt.Printf("%s Logged out of %s\n", ui.Green("✓"), ui.Cyan(userEmail))
		} else {
			fmt.Printf("%s Logged out successfully\n", ui.Green("✓"))
		}
		return nil
	},
}

var authStatusCmd = &cobra.Command{
	Use:     "status",
	Aliases: []string{"whoami"},
	Short:   "View current authentication status",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		if !cfg.IsLoggedIn() {
			fmt.Println(ui.Dim("Not logged in to Beep."))
			if cfg.ServerURL != "" {
				fmt.Printf("Server: %s\n", ui.Cyan(cfg.ServerURL))
			}
			fmt.Printf("\n%s Run %s to log in.\n", ui.Dim("Tip:"), ui.Cyan("beep auth login"))
			return nil
		}

		c := client.New(cfg)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		me, err := c.GetMe(ctx)
		if err != nil {
			if strings.Contains(err.Error(), "401") || strings.Contains(err.Error(), "invalid or expired") {
				fmt.Println(ui.Error("Stored login session is expired or invalid."))
				fmt.Printf("Run %s to re-authenticate.\n", ui.Cyan("beep auth login"))
				return nil
			}
			// Offline or network error
			fmt.Println(ui.Yellow("⚠ Could not verify authentication with server (network error): ") + err.Error())
			fmt.Printf("Stored token: %s\n", ui.Dim(config.MaskToken(cfg.AccessToken)))
			if cfg.UserEmail != "" {
				fmt.Printf("Logged in as: %s\n", ui.Cyan(cfg.UserEmail))
			}
			return nil
		}

		displayName := me.Identity.Name
		if displayName == "" {
			displayName = me.Identity.Email
		} else {
			displayName = fmt.Sprintf("%s (%s)", displayName, me.Identity.Email)
		}

		fmt.Printf("%s Logged in to %s\n", ui.Green("✓"), ui.Bold(cfg.ServerURL))
		fmt.Printf("  %s %s\n", ui.Dim("User:"), ui.Cyan(displayName))
		if cfg.AccountSlug != "" {
			fmt.Printf("  %s %s\n", ui.Dim("Active account:"), ui.Yellow(cfg.AccountSlug))
		} else if me.LastAccountSlug != "" {
			fmt.Printf("  %s %s\n", ui.Dim("Active account:"), ui.Yellow(me.LastAccountSlug))
		}
		fmt.Printf("  %s %s\n", ui.Dim("Token:"), ui.Dim(config.MaskToken(cfg.AccessToken)))

		if len(me.Accounts) > 1 {
			fmt.Println()
			fmt.Println(ui.Dim("Available accounts:"))
			for _, acc := range me.Accounts {
				prefix := "  • "
				if acc.Slug == cfg.AccountSlug || (cfg.AccountSlug == "" && acc.Slug == me.LastAccountSlug) {
					prefix = "  * "
				}
				fmt.Printf("%s%s (%s)\n", prefix, ui.Bold(acc.Name), ui.Dim(acc.Slug))
			}
		}

		return nil
	},
}

func init() {
	authLoginCmd.Flags().BoolVar(&flagAuthNoBrowser, "no-browser", false, "Do not open browser automatically")
	authLoginCmd.Flags().StringVar(&flagAuthClientName, "client-name", "", "Name to identify this client/device")

	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authLogoutCmd)
	authCmd.AddCommand(authStatusCmd)
}
