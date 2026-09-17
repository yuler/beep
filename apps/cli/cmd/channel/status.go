package channel

import (
	"beep/internal/cliservice"
	"beep/internal/cmdutil"
	"beep/internal/daemon"

	"github.com/spf13/cobra"
)

// NewCmdStatus creates the 'channel status' subcommand.
func NewCmdStatus() *cobra.Command {
	return &cobra.Command{
		Use:     "status",
		Aliases: []string{"ps"},
		Short:   "Check channel daemon running status and information",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig(cmd)
			if err != nil {
				return err
			}
			return cliservice.ShowSingleServiceStatus(daemon.ServiceChannel, cfg)
		},
	}
}
