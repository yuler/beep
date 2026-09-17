package runner

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"time"

	"beep/internal/browser"
	"beep/internal/client"
	"beep/internal/cliservice"
	"beep/internal/cmdutil"
	"beep/internal/config"
	"beep/internal/daemon"
	"beep/internal/ui"
	"beep/internal/version"

	"github.com/spf13/cobra"
)

// NewCmdConnect creates the 'runner connect' subcommand.
func NewCmdConnect() *cobra.Command {
	var (
		name string
		tags []string
	)

	cmd := &cobra.Command{
		Use:   "connect",
		Short: "Connect this CLI as a self-hosted runner via Web browser (RFC 8628)",
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

			runnerName := name
			if runnerName == "" {
				runnerName = cfg.Hostname
			}
			if runnerName == "" {
				if h, err := os.Hostname(); err == nil {
					runnerName = h
				} else {
					runnerName = "CLI Runner"
				}
			}

			if len(tags) == 0 {
				tags = []string{"default"}
			}

			metadata := map[string]string{
				"version":  version.Version,
				"os":       runtime.GOOS,
				"arch":     runtime.GOARCH,
				"hostname": runnerName,
			}

			var authRes *client.DeviceAuthorizationResponse
			err = ui.WithSpinner("Requesting runner device authorization...", func() error {
				authReqCtx, authReqCancel := context.WithTimeout(ctx, 15*time.Second)
				defer authReqCancel()
				var reqErr error
				authRes, reqErr = c.RequestRunnerDeviceAuthorization(authReqCtx, runnerName, tags, metadata, accountSlug)
				return reqErr
			})
			if err != nil {
				return fmt.Errorf("runner device authorization request failed: %w", err)
			}

			fmt.Println()
			fmt.Printf("! %s %s\n", ui.Bold("First copy your one-time code:"), ui.Yellow(authRes.UserCode))

			verifyURL := authRes.VerificationURIComplete
			if verifyURL == "" {
				verifyURL = authRes.VerificationURI
			}

			fmt.Printf("  %s %s\n", ui.Dim("Open URL:"), ui.Cyan(verifyURL))
			fmt.Println()

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

			for {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(interval):
				}

				if time.Now().After(deadline) {
					return fmt.Errorf("device authorization expired, please run '%s runner connect' again", config.BinaryName())
				}

				pollCtx, pollCancel := context.WithTimeout(ctx, 10*time.Second)
				tokenRes, err := c.PollRunnerDeviceToken(pollCtx, authRes.DeviceCode)
				pollCancel()

				if err == nil && tokenRes != nil {
					workspace := cmdutil.GetWorkspace(cmd)
					configPath := config.GetConfigPath(workspace)
					fc, err := config.LoadFile(configPath)
					if err != nil {
						fc = &config.FileConfig{}
					}

					fc.RunnerToken = tokenRes.AccessToken
					if err := config.SaveFile(configPath, fc); err != nil {
						return fmt.Errorf("failed to save config: %w", err)
					}

					fmt.Println()
					fmt.Printf("%s %s %s (%s)\n",
						ui.Green("✓"),
						ui.Bold("Connected CLI runner:"),
						ui.Cyan(tokenRes.RunnerName),
						ui.Dim(config.MaskToken(tokenRes.AccessToken)),
					)
					fmt.Printf("%s Config saved to %s\n", ui.Green("✓"), ui.Dim(configPath))
					fmt.Println()

					cfg.RunnerToken = tokenRes.AccessToken

					var rawArgs []string
					if cfg.Workspace != "" {
						rawArgs = append(rawArgs, "--workspace", cfg.Workspace)
					}
					if cfg.ServerURL != "" {
						rawArgs = append(rawArgs, "--server", cfg.ServerURL)
					}

					return cliservice.AutoStartServiceDaemon(daemon.ServiceRunner, cfg, rawArgs)
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
						return fmt.Errorf("authorization was denied by the user")
					case client.OAuthErrExpiredToken:
						return fmt.Errorf("authorization request expired")
					default:
						return fmt.Errorf("authorization failed: %s", oauthErr.Error())
					}
				} else if err != nil {
					if ctx.Err() != nil {
						return ctx.Err()
					}
				}
			}
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Runner display name (defaults to hostname)")
	cmd.Flags().StringSliceVar(&tags, "tags", nil, "Runner tags (defaults to [\"default\"])")
	return cmd
}
