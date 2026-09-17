package channel

import (
	"time"

	"beep/internal/cliservice"
	"beep/internal/cmdutil"
	"beep/internal/daemon"

	"github.com/spf13/cobra"
)

// NewCmdStop creates the 'channel stop' subcommand.
func NewCmdStop() *cobra.Command {
	var (
		force   bool
		timeout time.Duration
	)

	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop the running channel daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig(cmd)
			if err != nil {
				return err
			}
			return cliservice.StopSingleService(daemon.ServiceChannel, cfg.Workspace, timeout, force)
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Forcibly kill the daemon process if graceful stop times out")
	cmd.Flags().DurationVar(&timeout, "timeout", 10*time.Second, "Timeout waiting for daemon to stop")
	return cmd
}
