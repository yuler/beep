package runner

import (
	"fmt"
	"os"
	"time"

	"beep/internal/cliservice"
	"beep/internal/cmdutil"

	"github.com/spf13/cobra"
)

// NewCmdStart creates the 'runner start' subcommand.
func NewCmdStart() *cobra.Command {
	var (
		concurrency  int
		pollInterval time.Duration
		daemonMode   bool
	)

	cmd := &cobra.Command{
		Use:     "start",
		Aliases: []string{"up"},
		Short:   "Start runner daemon to execute scheduled jobs",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig(cmd)
			if err != nil {
				return err
			}
			if concurrency > 0 {
				cfg.Concurrency = concurrency
			}
			if pollInterval > 0 {
				cfg.PollInterval = pollInterval
			}
			if err := cfg.Validate(); err != nil {
				return fmt.Errorf("configuration error: %w", err)
			}
			return cliservice.RunRunnerService(cfg, daemonMode, os.Args[1:])
		},
	}

	cmd.Flags().IntVarP(&concurrency, "concurrency", "c", 0, "Max concurrent jobs (default 5)")
	cmd.Flags().DurationVarP(&pollInterval, "poll-interval", "i", 0, "Poll interval (default 3s)")
	cmd.Flags().BoolVarP(&daemonMode, "daemon", "d", false, "Run runner in background")
	return cmd
}

// NewCmdUp provides backward compatibility for 'runner up'.
func NewCmdUp() *cobra.Command {
	return NewCmdStart()
}
