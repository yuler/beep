package cmd

import (
	"fmt"
	"os"
	"time"

	"beep/internal/config"
	"beep/internal/daemon"
	"beep/internal/ui"

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

func init() {
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
