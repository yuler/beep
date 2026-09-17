package channel

import (
	"beep/internal/ui"

	"github.com/spf13/cobra"
)

// NewCmdChannel creates and returns the parent 'channel' command.
func NewCmdChannel() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "channel",
		Short: "Connect this CLI as a notification channel",
		Long:    ui.Bold(ui.Cyan("Notification Channel Management")) + ` - Connect, run, and manage CLI notification channels.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(NewCmdConnect())
	cmd.AddCommand(NewCmdDisconnect())
	cmd.AddCommand(NewCmdStart())
	cmd.AddCommand(NewCmdStop())
	cmd.AddCommand(NewCmdStatus())

	return cmd
}
