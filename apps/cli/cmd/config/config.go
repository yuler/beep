package config

import (
	"fmt"

	"beep/internal/config"
	"beep/internal/ui"

	"github.com/spf13/cobra"
)

// NewCmdConfig creates and returns the parent 'config' command.
func NewCmdConfig() *cobra.Command {
	var flagShowToken bool

	cmd := &cobra.Command{
		Use:   "config",
		Short: fmt.Sprintf("Manage local runner configuration (%s/config.json)", config.DefaultWorkspaceDisplay()),
		Long:  ui.Bold(ui.Cyan("Configuration Management")) + fmt.Sprintf(" - View, set, and manage local runner configuration (%s/config.json).", config.DefaultWorkspaceDisplay()),
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunConfigShow(cmd, args, flagShowToken)
		},
	}

	cmd.PersistentFlags().BoolVar(&flagShowToken, "show-token", false, "Display unmasked runner and channel tokens")

	cmd.AddCommand(NewCmdShow())
	cmd.AddCommand(NewCmdSet())
	cmd.AddCommand(NewCmdUnset())
	cmd.AddCommand(NewCmdPath())
	cmd.AddCommand(NewCmdList())

	return cmd
}
