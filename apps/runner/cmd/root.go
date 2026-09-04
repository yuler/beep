package cmd

import (
	"fmt"
	"os"

	"beep-runner/internal/config"
	"beep-runner/internal/ui"

	"github.com/spf13/cobra"
)

var (
	flagWorkspace     string
	flagServer        string
	flagToken         string
	flagNoColor       bool
	flagNoInteractive bool
)

var RootCmd = &cobra.Command{
	Use:   "beep-runner",
	Short: "Beep self-hosted runner daemon and workspace manager",
	Long: ui.Bold(ui.Cyan("Beep self-hosted runner")) + `

A runner executes scheduled jobs locally in your workspace and reports logs/results to Beep Core.
Use job create / push / pull to manage local check scripts.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if flagNoColor {
			ui.SetEnabled(false)
		}
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Default action when no subcommand is given is runUp
		return runUp(cmd, args)
	},
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, ui.Error("%v", err))
		os.Exit(1)
	}
}

func init() {
	RootCmd.PersistentFlags().StringVarP(&flagWorkspace, "workspace", "w", "", "Local job workspace directory (default ~/.beep-runner, env: BEEP_WORKSPACE)")
	RootCmd.PersistentFlags().StringVarP(&flagServer, "server", "s", "", "Beep server URL (env: BEEP_SERVER)")
	RootCmd.PersistentFlags().StringVarP(&flagToken, "token", "t", "", "Runner authentication token (env: BEEP_RUNNER_TOKEN)")
	RootCmd.PersistentFlags().BoolVar(&flagNoColor, "no-color", false, "Disable colored output")
	RootCmd.PersistentFlags().BoolVar(&flagNoInteractive, "no-interactive", false, "Disable interactive prompts")

	// Root flags for default action (up)
	RootCmd.Flags().IntVarP(&flagConcurrency, "concurrency", "c", 0, "Max concurrent jobs (default 5)")
	RootCmd.Flags().DurationVarP(&flagPollInterval, "poll-interval", "i", 0, "Poll interval (default 3s)")
	RootCmd.Flags().BoolVarP(&flagDaemon, "daemon", "d", false, "Run runner daemon in background")

	// Register subcommands
	RootCmd.AddCommand(upCmd)
	RootCmd.AddCommand(pingCmd)
	RootCmd.AddCommand(versionCmd)
	RootCmd.AddCommand(configCmd)
	RootCmd.AddCommand(jobCmd)
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
