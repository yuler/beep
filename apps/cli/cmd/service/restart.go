package service

import (
	"errors"
	"fmt"
	"os"
	"time"

	"beep/internal/cmdutil"
	"beep/internal/config"
	"beep/internal/daemon"
	intsvc "beep/internal/service"
	"beep/internal/ui"

	"github.com/charmbracelet/huh"
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
		Long: `Restart Beep daemon services (stops and relaunches in background).

When run without arguments in an interactive terminal, prompts to select which
services to restart. In non-interactive environments (e.g. CI or scripts),
restarts all configured services.`,
		Args: cobra.MaximumNArgs(1),
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
				return restartService(daemon.ServiceRunner, cfg, timeout, force)
			case "channel":
				return restartService(daemon.ServiceChannel, cfg, timeout, force)
			default:
				if cmdutil.IsInteractive(cmd) {
					return restartInteractive(cfg, timeout, force)
				}
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
		if !cfg.HasRunnerService() {
			return fmt.Errorf("runner token is not configured (set via BEEP_RUNNER_TOKEN or config.json)")
		}
	case daemon.ServiceChannel:
		if !cfg.HasChannelService() {
			return fmt.Errorf("channel token is not configured (run '%s channel connect')", config.BinaryName())
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

	return intsvc.StartServiceDaemonFn(service, []string{service, "up"}, os.Args[1:], cfg)
}

func restartAll(cfg *config.Config, timeout time.Duration, force bool) error {
	hasRunner := cfg.HasRunnerService()
	hasChannel := cfg.HasChannelService()

	if !hasRunner && !hasChannel {
		return fmt.Errorf("no services configured. To configure:\n  Runner:  set BEEP_RUNNER_TOKEN or configure config.json\n  Channel: run '%s channel connect'", config.BinaryName())
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

var promptRestartServicesFn = promptRestartServices

func promptRestartServices(options []huh.Option[string], defaultSelected []string) ([]string, error) {
	selected := defaultSelected
	err := huh.NewMultiSelect[string]().
		Title("Select services to restart").
		Description("Use Space to toggle, Enter to confirm").
		Options(options...).
		Value(&selected).
		Run()
	if err != nil {
		return nil, err
	}
	return selected, nil
}

func restartInteractive(cfg *config.Config, timeout time.Duration, force bool) error {
	hasRunner := cfg.HasRunnerService()
	hasChannel := cfg.HasChannelService()

	if !hasRunner && !hasChannel {
		return fmt.Errorf("no services configured. To configure:\n  Runner:  set BEEP_RUNNER_TOKEN or configure config.json\n  Channel: run '%s channel connect'", config.BinaryName())
	}

	var options []huh.Option[string]
	var defaultSelected []string
	var runningCount int

	if hasRunner {
		running, pid, _ := daemon.CheckRunning(cfg.Workspace, daemon.ServiceRunner)
		label := "runner (stopped)"
		if running && pid > 0 {
			label = fmt.Sprintf("runner (running, PID %d)", pid)
			defaultSelected = append(defaultSelected, daemon.ServiceRunner)
			runningCount++
		}
		options = append(options, huh.NewOption(label, daemon.ServiceRunner))
	}

	if hasChannel {
		running, pid, _ := daemon.CheckRunning(cfg.Workspace, daemon.ServiceChannel)
		label := "channel (stopped)"
		if running && pid > 0 {
			label = fmt.Sprintf("channel (running, PID %d)", pid)
			defaultSelected = append(defaultSelected, daemon.ServiceChannel)
			runningCount++
		}
		options = append(options, huh.NewOption(label, daemon.ServiceChannel))
	}

	if runningCount == 0 {
		if hasRunner {
			defaultSelected = append(defaultSelected, daemon.ServiceRunner)
		}
		if hasChannel {
			defaultSelected = append(defaultSelected, daemon.ServiceChannel)
		}
	}

	selected, err := promptRestartServicesFn(options, defaultSelected)
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return nil
		}
		return err
	}

	if len(selected) == 0 {
		fmt.Println("No services selected to restart.")
		return nil
	}

	for _, service := range selected {
		if err := restartService(service, cfg, timeout, force); err != nil {
			return err
		}
	}

	return nil
}
