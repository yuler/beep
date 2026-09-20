package service

import (
	"fmt"
	"os"
	"time"

	"beep/internal/cmdutil"
	"beep/internal/config"
	"beep/internal/daemon"
	intsvc "beep/internal/service"

	"github.com/spf13/cobra"
)

// NewCmdStart creates the 'service start' subcommand.
func NewCmdStart() *cobra.Command {
	var (
		concurrency  int
		pollInterval time.Duration
		daemonMode   bool
	)

	cmd := &cobra.Command{
		Use:     "start [runner|channel]",
		Aliases: []string{"up"},
		Short:   "Start daemon services to listen for notifications and execute jobs",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target, err := parseServiceArg(args)
			if err != nil {
				return err
			}

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

			switch target {
			case "runner":
				return intsvc.RunRunnerService(cfg, daemonMode, os.Args[1:])
			case "channel":
				return intsvc.RunChannelService(cfg, daemonMode, os.Args[1:])
			default:
				return runStartAll(cfg, daemonMode)
			}
		},
	}

	cmd.Flags().IntVarP(&concurrency, "concurrency", "c", 0, "Max concurrent jobs for runner (default 5)")
	cmd.Flags().DurationVarP(&pollInterval, "poll-interval", "i", 0, "Poll interval (default 3s)")
	cmd.Flags().BoolVarP(&daemonMode, "daemon", "d", false, "Run in background (systemd/LaunchAgent when available, otherwise detach)")

	return cmd
}

func runStartAll(cfg *config.Config, daemonMode bool) error {
	hasRunner := cfg.HasRunnerService()
	hasChannel := cfg.HasChannelService()

	if !hasRunner && !hasChannel {
		return fmt.Errorf("no services configured. To configure:\n  Runner:  set BEEP_RUNNER_TOKEN or configure config.json\n  Channel: run '%s channel connect'", config.BinaryName())
	}

	if daemonMode {
		if hasRunner {
			if err := intsvc.StartServiceDaemonFn(daemon.ServiceRunner, os.Args[1:], cfg); err != nil {
				return err
			}
		}
		if hasChannel {
			if err := intsvc.StartServiceDaemonFn(daemon.ServiceChannel, os.Args[1:], cfg); err != nil {
				return err
			}
		}
		return nil
	}

	var services []string
	if hasRunner {
		services = append(services, daemon.ServiceRunner)
	}
	if hasChannel {
		services = append(services, daemon.ServiceChannel)
	}
	return intsvc.RunForegroundServicesFn(cfg, services, os.Args[1:])
}
