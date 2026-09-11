package cmd

import (
	"fmt"
	"os"

	"beep/internal/config"
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

func init() {
	runnerCmd.PersistentFlags().StringVarP(&flagWorkspace, "workspace", "w", "", fmt.Sprintf("Local job workspace directory (default %s, env: BEEP_WORKSPACE)", config.DefaultWorkspaceDisplay()))
	runnerCmd.PersistentFlags().StringVarP(&flagServer, "server", "s", "", "Beep server URL (env: BEEP_SERVER)")
	runnerCmd.PersistentFlags().StringVarP(&flagToken, "token", "t", "", "Runner authentication token (env: BEEP_RUNNER_TOKEN)")

	runnerCmd.AddCommand(upCmd)
	runnerCmd.AddCommand(stopCmd)
	runnerCmd.AddCommand(statusCmd)
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
