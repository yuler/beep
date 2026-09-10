package cmd

import (
	"fmt"
	"os"

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

	// Register top-level subcommands
	RootCmd.AddCommand(runnerCmd)
	RootCmd.AddCommand(channelCmd)
	RootCmd.AddCommand(configCmd)
	RootCmd.AddCommand(versionCmd)
}
