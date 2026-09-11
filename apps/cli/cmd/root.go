package cmd

import (
	"fmt"
	"os"

	"beep/internal/config"
	"beep/internal/ui"

	"github.com/spf13/cobra"
)

var (
	flagNoColor       bool
	flagNoInteractive bool
)

var RootCmd = &cobra.Command{
	Use:   "beep",
	Short: "Beep command-line interface",
	Long: ui.Bold(ui.Cyan("Beep CLI")) + ` - Command-line interface for the Beep platform.

Execute 'beep <command> --help' for detailed usage of a specific command.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if flagNoColor {
			ui.SetEnabled(false)
		}
	},
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, ui.Error("%v", err))
		os.Exit(1)
	}
}

func init() {
	RootCmd.PersistentFlags().BoolVar(&flagNoColor, "no-color", false, "Disable colored output")
	RootCmd.PersistentFlags().BoolVar(&flagNoInteractive, "no-interactive", false, "Disable interactive prompts")
	RootCmd.PersistentFlags().StringVarP(&flagWorkspace, "workspace", "w", "", fmt.Sprintf("Local job workspace directory (default %s, env: BEEP_WORKSPACE)", config.DefaultWorkspaceDisplay()))
	RootCmd.PersistentFlags().StringVarP(&flagServer, "server", "s", "", "Beep server URL (env: BEEP_SERVER)")
	RootCmd.PersistentFlags().StringVarP(&flagToken, "token", "t", "", "Runner authentication token (env: BEEP_RUNNER_TOKEN)")

	// Daemon commands
	RootCmd.AddCommand(newUpCmd())
	RootCmd.AddCommand(newStopCmd())
	RootCmd.AddCommand(newStatusCmd())

	// Top-level subcommands
	RootCmd.AddCommand(runnerCmd)
	RootCmd.AddCommand(channelCmd)
	RootCmd.AddCommand(configCmd)
	RootCmd.AddCommand(versionCmd)
}
