package runner

import (
	cmdconfig "beep/cmd/config"
	cmdjob "beep/cmd/runner/job"
	"beep/internal/ui"

	"github.com/spf13/cobra"
)

// NewCmdRunner creates and returns the parent 'runner' command.
func NewCmdRunner() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "runner",
		Short: "Manage Beep self-hosted runner daemon and workspace",
		Long: ui.Bold(ui.Cyan("Beep Self-Hosted Runner")) + `
A runner executes scheduled jobs locally in your workspace and reports logs/results to Beep Core.
Use runner job create / push / pull to manage local check scripts.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(NewCmdConnect())
	cmd.AddCommand(NewCmdDisconnect())
	cmd.AddCommand(NewCmdUp())
	cmd.AddCommand(NewCmdStop())
	cmd.AddCommand(NewCmdStatus())
	cmd.AddCommand(cmdconfig.NewCmdConfig())
	cmd.AddCommand(cmdjob.NewCmdJob())

	return cmd
}
