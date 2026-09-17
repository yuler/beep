package auth

import (
	"beep/internal/ui"

	"github.com/spf13/cobra"
)

// NewCmdAuth creates and returns the parent 'auth' command.
func NewCmdAuth() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authenticate CLI with Beep platform",
		Long:  ui.Bold(ui.Cyan("Beep Authentication")) + ` - Manage login session, credentials, and authentication status.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(NewCmdLogin())
	cmd.AddCommand(NewCmdLogout())
	cmd.AddCommand(NewCmdStatus())

	return cmd
}
