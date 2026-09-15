package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"time"

	"beep/internal/browser"
	"beep/internal/client"
	"beep/internal/config"
	"beep/internal/daemon"
	"beep/internal/ui"

	"github.com/spf13/cobra"
)

var channelCmd = &cobra.Command{
	Use:     "channel",
	Aliases: []string{"channels"},
	Short:   "Connect this CLI as a notification channel",
}

var (
	flagChannelAccount string
)

var channelConnectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Connect this CLI as a notification channel via Web browser (RFC 8628)",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		if cfg.ServerURL == "" {
			return fmt.Errorf("server URL is not configured (set via BEEP_SERVER or 'beep config set server <url>')")
		}

		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
		defer cancel()

		me, err := ensureLoggedIn(ctx, cfg)
		if err != nil {
			return err
		}

		accountSlug, err := resolveAccountSlug(me, flagChannelAccount, cfg.AccountSlug)
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

		fmt.Println(ui.Dim("Requesting device authorization from ") + ui.Cyan(cfg.ServerURL) + ui.Dim("..."))

		authReqCtx, authReqCancel := context.WithTimeout(ctx, 15*time.Second)
		authRes, err := c.RequestDeviceAuthorization(authReqCtx, channelName, accountSlug)
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
				return fmt.Errorf("device authorization expired, please run 'beep channel connect' again")
			}

			pollCtx, pollCancel := context.WithTimeout(ctx, 10*time.Second)
			tokenRes, err := c.PollDeviceToken(pollCtx, authRes.DeviceCode)
			pollCancel()

			if err == nil && tokenRes != nil {
				// Success! Save to config.json
				configPath := config.GetConfigPath(flagWorkspace)
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
				return autoStartServiceDaemon(daemon.ServiceChannel, cfg)
			}

			var oauthErr *client.OAuthErrorResponse
			if errors.As(err, &oauthErr) {
				// The server answered — this was not a network failure.
				consecutiveNetworkErrors = 0
				switch oauthErr.ErrorCode {
				case client.OAuthErrAuthorizationPending:
					// Continue polling
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
				// Network error or unexpected response. Back off instead of
				// spinning at poll frequency, and give up if the server never
				// answers — the 15-minute deadline alone is not feedback.
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

var channelDisconnectCmd = &cobra.Command{
	Use:   "disconnect",
	Short: "Disconnect CLI channel (removes server-side channel and clears local token)",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		configPath := config.GetConfigPath(flagWorkspace)
		fc, err := config.LoadFile(configPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if fc.ChannelToken == "" && fc.CliToken == "" && fc.DeviceToken == "" {
			fmt.Println(ui.Dim("No CLI channel token configured in " + configPath))
			return nil
		}

		if cfg.ServerURL != "" && cfg.ChannelToken != "" {
			c := client.New(cfg)
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			if err := c.DisconnectChannel(ctx); err != nil {
				fmt.Printf("%s Failed to remove server-side channel: %v\n", ui.Yellow("!"), err)
			}
		}

		fc.ChannelToken = ""
		fc.CliToken = ""
		fc.DeviceToken = ""
		if err := config.SaveFile(configPath, fc); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Printf("%s Disconnected CLI channel from %s\n", ui.Green("✓"), ui.Dim(configPath))
		return nil
	},
}

func newChannelUpCmd() *cobra.Command {
	var (
		pollInterval time.Duration
		daemonMode   bool
	)
	cmd := &cobra.Command{
		Use:     "up",
		Aliases: []string{"run"},
		Short:   "Start channel daemon to listen for notifications and execute hooks",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			if pollInterval > 0 {
				cfg.PollInterval = pollInterval
			}
			if err := cfg.Validate(); err != nil {
				return fmt.Errorf("configuration error: %w", err)
			}
			return runChannelService(cfg, daemonMode)
		},
	}
	cmd.Flags().DurationVarP(&pollInterval, "poll-interval", "i", 0, "Poll interval (default 3s)")
	cmd.Flags().BoolVarP(&daemonMode, "daemon", "d", false, "Run channel daemon in background")
	return cmd
}

func newChannelStopCmd() *cobra.Command {
	var (
		force   bool
		timeout time.Duration
	)
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop the running channel daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			return stopSingleService(daemon.ServiceChannel, cfg.Workspace, timeout, force)
		},
	}
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Forcibly kill the daemon process if graceful stop times out")
	cmd.Flags().DurationVar(&timeout, "timeout", 10*time.Second, "Timeout waiting for daemon to stop")
	return cmd
}

func newChannelStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Check channel daemon running status and information",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			return showSingleServiceStatus(daemon.ServiceChannel, cfg)
		},
	}
}

func init() {
	channelConnectCmd.Flags().StringVarP(&flagChannelAccount, "account", "a", "", "Account slug to connect to")
	channelCmd.AddCommand(newChannelUpCmd())
	channelCmd.AddCommand(newChannelStopCmd())
	channelCmd.AddCommand(newChannelStatusCmd())
	channelCmd.AddCommand(channelConnectCmd)
	channelCmd.AddCommand(channelDisconnectCmd)
}
