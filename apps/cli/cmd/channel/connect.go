package channel

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"time"

	"beep/internal/browser"
	"beep/internal/client"
	"beep/internal/cmdutil"
	"beep/internal/config"
	"beep/internal/daemon"
	"beep/internal/service"
	"beep/internal/ui"

	"github.com/spf13/cobra"
)

// NewCmdConnect creates the 'channel connect' subcommand.
func NewCmdConnect() *cobra.Command {
	return &cobra.Command{
		Use:   "connect",
		Short: "Connect this CLI as a notification channel via Web browser (RFC 8628)",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig(cmd)
			if err != nil {
				return err
			}

			if cfg.ServerURL == "" {
				return fmt.Errorf("server URL is not configured (set via BEEP_SERVER or '%s config set server <url>')", config.BinaryName())
			}

			ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
			defer cancel()

			me, err := cmdutil.EnsureLoggedIn(ctx, cfg)
			if err != nil {
				return err
			}

			var flagAccount string
			if cmd.Root() != nil && cmd.Root().PersistentFlags().Lookup("account") != nil {
				flagAccount, _ = cmd.Root().PersistentFlags().GetString("account")
			}
			accountSlug, err := cmdutil.ResolveAccountSlug(me, flagAccount, cfg.AccountSlug)
			if err != nil {
				return err
			}

			c := client.New(cfg)

			channelName := cfg.Hostname
			if channelName == "" {
				if h, err := os.Hostname(); err == nil {
					channelName = h
				} else {
					channelName = "CLI Device"
				}
			}

			var authRes *client.DeviceAuthorizationResponse
			err = ui.WithSpinner("Requesting device authorization...", func() error {
				authReqCtx, authReqCancel := context.WithTimeout(ctx, 15*time.Second)
				defer authReqCancel()
				var reqErr error
				authRes, reqErr = c.RequestDeviceAuthorization(authReqCtx, channelName, accountSlug)
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

			// Attempt to open browser automatically
			_ = browser.Open(verifyURL)

			fmt.Printf("%s Waiting for authorization in web browser...\n", ui.Dim("⏳"))

			interval := time.Duration(authRes.Interval) * time.Second
			if interval <= 0 {
				interval = 5 * time.Second
			}

			expiresIn := authRes.ExpiresIn
			if expiresIn <= 0 {
				expiresIn = 900
			}
			deadline := time.Now().Add(time.Duration(expiresIn) * time.Second)

			const maxPollInterval = 30 * time.Second
			consecutiveNetworkErrors := 0

			for {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(interval):
				}

				if time.Now().After(deadline) {
					return fmt.Errorf("device authorization expired, please run '%s channel connect' again", config.BinaryName())
				}

				pollCtx, pollCancel := context.WithTimeout(ctx, 10*time.Second)
				tokenRes, err := c.PollDeviceToken(pollCtx, authRes.DeviceCode)
				pollCancel()

				if err == nil && tokenRes != nil {
					workspace := cmdutil.GetWorkspace(cmd)
					configPath := config.GetConfigPath(workspace)
					fc, err := config.LoadFile(configPath)
					if err != nil {
						fc = &config.FileConfig{}
					}

					fc.ChannelToken = tokenRes.AccessToken
					fc.CliToken = ""
					fc.DeviceToken = ""
					if err := config.SaveFile(configPath, fc); err != nil {
						return fmt.Errorf("failed to save config: %w", err)
					}

					fmt.Println()
					fmt.Printf("%s %s %s (%s)\n",
						ui.Green("✓"),
						ui.Bold("Connected CLI channel:"),
						ui.Cyan(tokenRes.ChannelName),
						ui.Dim(config.MaskToken(tokenRes.AccessToken)),
					)
					fmt.Printf("%s Config saved to %s\n", ui.Green("✓"), ui.Dim(configPath))
					fmt.Println()

					cfg.ChannelToken = tokenRes.AccessToken
					cfg.CliToken = ""
					cfg.DeviceToken = ""

					var rawArgs []string
					if cfg.Workspace != "" {
						rawArgs = append(rawArgs, "--workspace", cfg.Workspace)
					}
					if cfg.ServerURL != "" {
						rawArgs = append(rawArgs, "--server", cfg.ServerURL)
					}

					return service.AutoStartServiceDaemon(daemon.ServiceChannel, cfg, rawArgs)
				}

				var oauthErr *client.OAuthErrorResponse
				if errors.As(err, &oauthErr) {
					consecutiveNetworkErrors = 0
					switch oauthErr.ErrorCode {
					case client.OAuthErrAuthorizationPending:
						continue
					case client.OAuthErrSlowDown:
						interval += 5 * time.Second
						if interval > maxPollInterval {
							interval = maxPollInterval
						}
						continue
					case client.OAuthErrAccessDenied:
						return fmt.Errorf("authorization was denied by the user")
					case client.OAuthErrExpiredToken:
						return fmt.Errorf("authorization request expired")
					default:
						return fmt.Errorf("authorization failed: %s", oauthErr.Error())
					}
				} else if err != nil {
					consecutiveNetworkErrors++
					if consecutiveNetworkErrors >= 10 {
						return fmt.Errorf("device authorization polling failed %d times in a row: %w", consecutiveNetworkErrors, err)
					}
					backoff := interval * time.Duration(consecutiveNetworkErrors)
					if backoff > maxPollInterval {
						backoff = maxPollInterval
					}
					select {
					case <-ctx.Done():
						return ctx.Err()
					case <-time.After(backoff):
					}
					continue
				}
			}
		},
	}
}
