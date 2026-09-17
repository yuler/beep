package service

import (
	"fmt"
	"os"
	"strings"
	"time"

	"beep/internal/cliservice"
	"beep/internal/cmdutil"
	"beep/internal/config"
	"beep/internal/daemon"
	"beep/internal/ui"

	"github.com/spf13/cobra"
)

// NewCmdRestart creates the 'service restart' subcommand.
func NewCmdRestart() *cobra.Command {
	var (
		force        bool
		timeout      time.Duration
		concurrency  int
		pollInterval time.Duration
	)

	cmd := &cobra.Command{
		Use:   "restart [runner|channel]",
		Short: "Restart Beep daemon services (stops and relaunches in background)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := ""
			if len(args) > 0 {
				target = strings.ToLower(args[0])
			}
			if target != "" && target != "runner" && target != "channel" {
				return fmt.Errorf("unknown service %q (expected 'runner' or 'channel')", target)
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
				return restartService(daemon.ServiceRunner, cfg, timeout, force)
			case "channel":
				return restartService(daemon.ServiceChannel, cfg, timeout, force)
			default:
				return restartAll(cfg, timeout, force)
			}
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Forcibly kill the daemon process if graceful stop times out")
	cmd.Flags().DurationVar(&timeout, "timeout", 10*time.Second, "Timeout waiting for daemon to stop")
	cmd.Flags().IntVarP(&concurrency, "concurrency", "c", 0, "Max concurrent jobs for runner")
	cmd.Flags().DurationVarP(&pollInterval, "poll-interval", "i", 0, "Poll interval")

	return cmd
}

func restartService(service string, cfg *config.Config, timeout time.Duration, force bool) error {
	switch service {
	case daemon.ServiceRunner:
		if cfg.RunnerToken == "" {
			return fmt.Errorf("runner token is not configured (set via BEEP_RUNNER_TOKEN or config.json)")
		}
	case daemon.ServiceChannel:
		token := cfg.ChannelToken
		if token == "" {
			token = cfg.CliToken
		}
		if token == "" {
			token = cfg.DeviceToken
		}
		if token == "" {
			return fmt.Errorf("channel token is not configured (run 'beep channel connect')")
		}
	}

	running, pid, err := daemon.CheckRunning(cfg.Workspace, service)
	if err != nil {
		return fmt.Errorf("failed to check %s daemon status: %w", service, err)
	}

	if running && pid > 0 {
		fmt.Printf("%s Restarting %s daemon (stopping PID: %d)...\n", ui.Cyan("●"), service, pid)
		if _, err := daemon.StopDaemon(cfg.Workspace, service, timeout, force); err != nil {
			return fmt.Errorf("failed to stop existing %s daemon: %w", service, err)
		}
	}

	return cliservice.StartServiceDaemonFn(service, []string{service, "up"}, os.Args[1:], cfg)
}

func restartAll(cfg *config.Config, timeout time.Duration, force bool) error {
	hasRunner := cfg.RunnerToken != ""
	hasChannel := cfg.ChannelToken != "" || cfg.CliToken != "" || cfg.DeviceToken != ""

	if !hasRunner && !hasChannel {
		return fmt.Errorf("no services configured. To configure:\n  Runner:  set BEEP_RUNNER_TOKEN or configure config.json\n  Channel: run 'beep channel connect'")
	}

	if hasRunner {
		if err := restartService(daemon.ServiceRunner, cfg, timeout, force); err != nil {
			return err
		}
	}

	if hasChannel {
		if err := restartService(daemon.ServiceChannel, cfg, timeout, force); err != nil {
			return err
		}
	}

	return nil
}
