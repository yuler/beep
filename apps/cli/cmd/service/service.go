package service

import (
	"github.com/spf13/cobra"
)

// NewCmdService creates the 'beep service' parent command.
func NewCmdService() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "service",
		Short: "Manage local background daemon services (runner and channel)",
		Long:    "Manage local background daemon services for executing jobs (runner) and receiving notifications (channel).",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(NewCmdStart())
	cmd.AddCommand(NewCmdStop())
	cmd.AddCommand(NewCmdRestart())
	cmd.AddCommand(NewCmdStatus())

	return cmd
}
