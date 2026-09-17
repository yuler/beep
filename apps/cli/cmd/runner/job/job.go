package job

import (
	"beep/internal/ui"

	"github.com/spf13/cobra"
)

// NewCmdJob creates and returns the parent 'job' command.
func NewCmdJob() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "job",
		Short: "Manage workspace jobs and sync with Beep Core",
		Long:  ui.Bold(ui.Cyan("Job Workspace Management")) + ` - Manage local check scripts, push definitions to Beep Core, and pull remote jobs.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(NewCmdCreate())
	cmd.AddCommand(NewCmdList())
	cmd.AddCommand(NewCmdPush())
	cmd.AddCommand(NewCmdPull())
	cmd.AddCommand(NewCmdDelete())

	return cmd
}
