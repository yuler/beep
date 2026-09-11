package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"beep/internal/browser"
	"beep/internal/client"
	"beep/internal/config"
	"beep/internal/ui"

	"github.com/spf13/cobra"
)

var channelCmd = &cobra.Command{
	Use:     "channel",
	Aliases: []string{"channels"},
	Short:   "Connect this CLI as a notification channel",
}

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

		c := client.New(cfg)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

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
		authRes, err := c.RequestDeviceAuthorization(authReqCtx, channelName)
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
		if interval < 2*time.Second {
			interval = 5 * time.Second
		}

		deadline := time.Now().Add(time.Duration(authRes.ExpiresIn) * time.Second)

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

				fc.CliToken = tokenRes.AccessToken
				fc.DeviceToken = tokenRes.AccessToken
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
				fmt.Printf("\n%s To start receiving notifications:\n  %s\n",
					ui.Dim("Next:"),
					ui.Cyan("beep up"),
				)
				return nil
			}

			if oauthErr, ok := err.(*client.OAuthErrorResponse); ok {
				switch oauthErr.ErrorCode {
				case client.OAuthErrAuthorizationPending:
					// Continue polling
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
				// Network error or unexpected response, retry if context not cancelled
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

		if fc.CliToken == "" && fc.DeviceToken == "" {
			fmt.Println(ui.Dim("No CLI channel token configured in " + configPath))
			return nil
		}

		if cfg.ServerURL != "" && cfg.CliToken != "" {
			c := client.New(cfg)
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			if err := c.DisconnectChannel(ctx); err != nil {
				fmt.Printf("%s Failed to remove server-side channel: %v\n", ui.Yellow("!"), err)
			}
		}

		fc.CliToken = ""
		fc.DeviceToken = ""
		if err := config.SaveFile(configPath, fc); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Printf("%s Disconnected CLI channel from %s\n", ui.Green("✓"), ui.Dim(configPath))
		return nil
	},
}

func init() {
	channelCmd.AddCommand(channelConnectCmd)
	channelCmd.AddCommand(channelDisconnectCmd)
}
