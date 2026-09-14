package cmd

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
	"beep/internal/config"
	"beep/internal/daemon"
	"beep/internal/ui"
	"beep/internal/version"

	"github.com/spf13/cobra"
)

var (
	flagWorkspace string
	flagServer    string
	flagToken     string
)

var runnerCmd = &cobra.Command{
	Use:   "runner",
	Short: "Manage Beep self-hosted runner daemon and workspace",
	Long: ui.Bold(ui.Cyan("Beep self-hosted runner")) + `

A runner executes scheduled jobs locally in your workspace and reports logs/results to Beep Core.
Use runner job create / push / pull to manage local check scripts.`,
}

func newRunnerUpCmd() *cobra.Command {
	var (
		concurrency  int
		pollInterval time.Duration
		daemonMode   bool
	)
	cmd := &cobra.Command{
		Use:     "up",
		Aliases: []string{"run"},
		Short:   "Start runner daemon to execute scheduled jobs",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			if concurrency > 0 {
				cfg.Concurrency = concurrency
			}
			if pollInterval > 0 {
				cfg.PollInterval = pollInterval
			}
			if err := cfg.Validate(); err != nil {
				return fmt.Errorf("configuration error: %w", err)
			}
			return runRunnerService(cfg, daemonMode)
		},
	}
	cmd.Flags().IntVarP(&concurrency, "concurrency", "c", 0, "Max concurrent jobs (default 5)")
	cmd.Flags().DurationVarP(&pollInterval, "poll-interval", "i", 0, "Poll interval (default 3s)")
	cmd.Flags().BoolVarP(&daemonMode, "daemon", "d", false, "Run runner in background")
	return cmd
}

func newRunnerStopCmd() *cobra.Command {
	var (
		force   bool
		timeout time.Duration
	)
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop the running runner daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			return stopSingleService(daemon.ServiceRunner, cfg.Workspace, timeout, force)
		},
	}
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Forcibly kill the daemon process if graceful stop times out")
	cmd.Flags().DurationVar(&timeout, "timeout", 10*time.Second, "Timeout waiting for daemon to stop")
	return cmd
}

func newRunnerStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Check runner daemon running status and information",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			return showSingleServiceStatus(daemon.ServiceRunner, cfg)
		},
	}
}

func newRunnerConnectCmd() *cobra.Command {
	var (
		name string
		tags []string
	)
	cmd := &cobra.Command{
		Use:   "connect",
		Short: "Connect this CLI as a self-hosted runner via Web browser (RFC 8628)",
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

			fmt.Println(ui.Dim("Requesting runner device authorization from ") + ui.Cyan(cfg.ServerURL) + ui.Dim("..."))

			authReqCtx, authReqCancel := context.WithTimeout(ctx, 15*time.Second)
			authRes, err := c.RequestRunnerDeviceAuthorization(authReqCtx, runnerName, tags, metadata)
			authReqCancel()
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
					return fmt.Errorf("device authorization expired, please run 'beep runner connect' again")
				}

				pollCtx, pollCancel := context.WithTimeout(ctx, 10*time.Second)
				tokenRes, err := c.PollRunnerDeviceToken(pollCtx, authRes.DeviceCode)
				pollCancel()

				if err == nil && tokenRes != nil {
					configPath := config.GetConfigPath(flagWorkspace)
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
					fmt.Printf("\n%s To start executing scheduled tasks:\n  %s\n  %s\n",
						ui.Dim("Next:"),
						ui.Cyan("beep runner up"),
						ui.Dim("or: beep up"),
					)
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

func newRunnerDisconnectCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "disconnect",
		Short: "Disconnect CLI runner (removes server-side runner and clears local token)",
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

			if fc.RunnerToken == "" {
				fmt.Println(ui.Dim("No CLI runner token configured in " + configPath))
				return nil
			}

			if cfg.ServerURL != "" && cfg.RunnerToken != "" {
				c := client.New(cfg)
				ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
				defer cancel()

				if err := c.DisconnectRunner(ctx); err != nil {
					fmt.Printf("%s Failed to remove server-side runner: %v\n", ui.Yellow("!"), err)
				}
			}

			fc.RunnerToken = ""
			if err := config.SaveFile(configPath, fc); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Printf("%s Disconnected CLI runner from %s\n", ui.Green("✓"), ui.Dim(configPath))
			return nil
		},
	}
}

func init() {
	runnerCmd.AddCommand(newRunnerConnectCmd())
	runnerCmd.AddCommand(newRunnerDisconnectCmd())
	runnerCmd.AddCommand(newRunnerUpCmd())
	runnerCmd.AddCommand(newRunnerStopCmd())
	runnerCmd.AddCommand(newRunnerStatusCmd())
	runnerCmd.AddCommand(configCmd)
	runnerCmd.AddCommand(jobCmd)
}

func loadConfig() (*config.Config, error) {
	cfg, err := config.Load(flagWorkspace)
	if err != nil {
		return nil, err
	}
	if flagServer != "" {
		cfg.ServerURL = flagServer
	}
	if flagToken != "" {
		cfg.RunnerToken = flagToken
		fmt.Fprintln(os.Stderr, ui.Warn("passing --token on the command line exposes it in process lists; prefer config.json or BEEP_RUNNER_TOKEN"))
	}
	if flagWorkspace != "" {
		cfg.Workspace = flagWorkspace
	}
	return cfg, nil
}
