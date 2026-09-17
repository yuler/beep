package auth

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	"beep/internal/browser"
	"beep/internal/client"
	"beep/internal/cmdutil"
	"beep/internal/config"
	"beep/internal/ui"

	"github.com/spf13/cobra"
)

// NewCmdLogin creates the 'auth login' subcommand.
func NewCmdLogin() *cobra.Command {
	var (
		flagNoBrowser  bool
		flagClientName string
	)

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Log in to Beep via web browser",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig(cmd)
			if err != nil {
				return err
			}

			if cfg.ServerURL == "" {
				return fmt.Errorf("server URL is not configured (set via BEEP_SERVER or '%s config set server <url>')", config.BinaryName())
			}

			c := client.New(cfg)
			ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
			defer cancel()

			clientName := strings.TrimSpace(flagClientName)
			if clientName == "" {
				if h, err := os.Hostname(); err == nil {
					clientName = fmt.Sprintf("CLI on %s", h)
				} else {
					clientName = "Beep CLI"
				}
			}

			var authRes *client.DeviceAuthorizationResponse
			err = ui.WithSpinner("Requesting authorization...", func() error {
				authReqCtx, authReqCancel := context.WithTimeout(ctx, 15*time.Second)
				defer authReqCancel()
				var reqErr error
				authRes, reqErr = c.RequestCliDeviceAuthorization(authReqCtx, clientName)
				return reqErr
			})
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

			shouldOpenBrowser := !flagNoBrowser && !cmdutil.IsHeadless()
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
					return fmt.Errorf("device authorization expired, please run '%s auth login' again", config.BinaryName())
				}

				pollCtx, pollCancel := context.WithTimeout(ctx, 10*time.Second)
				tokenRes, err := c.PollCliDeviceToken(pollCtx, authRes.DeviceCode)
				pollCancel()

				if err == nil && tokenRes != nil {
					workspace, _ := cmd.Root().PersistentFlags().GetString("workspace")
					configPath := config.GetConfigPath(workspace)
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
						continue
					case client.OAuthErrSlowDown:
						interval += 5 * time.Second
						continue
					case client.OAuthErrAccessDenied:
						return fmt.Errorf("authorization was denied in the web browser")
					case client.OAuthErrExpiredToken:
						return fmt.Errorf("authorization request expired, please run '%s auth login' again", config.BinaryName())
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

	cmd.Flags().BoolVar(&flagNoBrowser, "no-browser", false, "Do not open browser automatically")
	cmd.Flags().StringVar(&flagClientName, "client-name", "", "Name to identify this client/device")

	return cmd
}
