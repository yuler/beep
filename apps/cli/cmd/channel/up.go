package channel

import (
	"fmt"
	"os"
	"time"

	"beep/internal/cliservice"
	"beep/internal/cmdutil"

	"github.com/spf13/cobra"
)

// NewCmdUp creates the 'channel up' subcommand.
func NewCmdUp() *cobra.Command {
	var (
		pollInterval time.Duration
		daemonMode   bool
	)

	cmd := &cobra.Command{
		Use:     "up",
		Aliases: []string{"run"},
		Short:   "Start channel daemon to listen for notifications and execute hooks",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig(cmd)
			if err != nil {
				return err
			}
			if pollInterval > 0 {
				cfg.PollInterval = pollInterval
			}
			if err := cfg.Validate(); err != nil {
				return fmt.Errorf("configuration error: %w", err)
			}
			return cliservice.RunChannelService(cfg, daemonMode, os.Args[1:])
		},
	}

	cmd.Flags().DurationVarP(&pollInterval, "poll-interval", "i", 0, "Poll interval (default 3s)")
	cmd.Flags().BoolVarP(&daemonMode, "daemon", "d", false, "Run channel daemon in background")
	return cmd
}
