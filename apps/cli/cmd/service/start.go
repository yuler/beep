package service

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"beep/internal/channel"
	"beep/internal/cliservice"
	"beep/internal/cmdutil"
	"beep/internal/config"
	"beep/internal/daemon"
	"beep/internal/proc"
	"beep/internal/runner"
	"beep/internal/ui"
	"beep/internal/updater"
	"beep/internal/workspace"

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
		Short:   "Start daemon services to listen for notifications and execute tasks",
		Args:    cobra.MaximumNArgs(1),
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
				return cliservice.RunRunnerService(cfg, daemonMode, os.Args[1:])
			case "channel":
				return cliservice.RunChannelService(cfg, daemonMode, os.Args[1:])
			default:
				return runStartAll(cfg, daemonMode)
			}
		},
	}

	cmd.Flags().IntVarP(&concurrency, "concurrency", "c", 0, "Max concurrent jobs for runner (default 5)")
	cmd.Flags().DurationVarP(&pollInterval, "poll-interval", "i", 0, "Poll interval (default 3s)")
	cmd.Flags().BoolVarP(&daemonMode, "daemon", "d", false, "Run daemon in background")

	return cmd
}

func runStartAll(cfg *config.Config, daemonMode bool) error {
	hasRunner := cfg.RunnerToken != ""
	hasChannel := cfg.ChannelToken != "" || cfg.CliToken != "" || cfg.DeviceToken != ""

	if !hasRunner && !hasChannel {
		return fmt.Errorf("no services configured. To configure:\n  Runner:  set BEEP_RUNNER_TOKEN or configure config.json\n  Channel: run '%s channel connect'", config.BinaryName())
	}

	isChild := os.Getenv("BEEP_DAEMON_CHILD") == "1"

	// Background daemon mode: start each configured service as background child
	if daemonMode && !isChild {
		if hasRunner {
			if err := cliservice.StartServiceDaemonFn(daemon.ServiceRunner, []string{"runner", "up"}, os.Args[1:], cfg); err != nil {
				return err
			}
		}
		if hasChannel {
			if err := cliservice.StartServiceDaemonFn(daemon.ServiceChannel, []string{"channel", "up"}, os.Args[1:], cfg); err != nil {
				return err
			}
		}
		return nil
	}

	ws, err := workspace.Open(cfg.Workspace)
	if err != nil {
		return fmt.Errorf("workspace error: %w", err)
	}

	// Foreground mode: acquire sockets for all active services
	var sockRunner *daemon.SocketListener
	var sockChannel *daemon.SocketListener

	if hasRunner {
		var err error
		sockRunner, err = daemon.AcquireSocket(cfg.Workspace, daemon.ServiceRunner)
		if err != nil {
			return err
		}
		defer sockRunner.Close()
	}

	if hasChannel {
		var err error
		sockChannel, err = daemon.AcquireSocket(cfg.Workspace, daemon.ServiceChannel)
		if err != nil {
			return err
		}
		defer sockChannel.Close()
	}

	logWriter, _, err := daemon.SetupLogger(cfg.Workspace, "", !isChild)
	if err != nil {
		return fmt.Errorf("failed to setup daily logger: %w", err)
	}
	defer logWriter.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go updater.RunDaemonUpdateProbe(ctx, cfg.Workspace, log.Printf)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, proc.ShutdownSignals...)
	go func() {
		<-sigChan
		log.Println(ui.Dim("[beep] Received termination signal..."))
		cancel()
	}()

	var wg sync.WaitGroup
	var runnerErr error
	var channelErr error

	if hasRunner {
		wg.Add(1)
		r := runner.New(cfg, ws)
		r.OnReady = func() {
			sockRunner.SetRunning()
			log.Println(ui.Green("✓") + " " + ui.Bold("Runner ready") + " " + ui.Dim("(waiting for scheduled jobs)"))
		}
		go func() {
			defer wg.Done()
			if err := r.Run(ctx); err != nil && err != context.Canceled {
				runnerErr = err
				cancel()
			}
		}()
	}

	if hasChannel {
		wg.Add(1)
		ch := channel.New(cfg, ws)
		ch.OnReady = func() {
			sockChannel.SetRunning()
			log.Println(ui.Green("✓") + " " + ui.Bold("Channel connected") + " " + ui.Dim("(listening for notifications)"))
		}
		go func() {
			defer wg.Done()
			if err := ch.Run(ctx); err != nil && err != context.Canceled {
				channelErr = err
				cancel()
			}
		}()
	}

	wg.Wait()

	if runnerErr != nil {
		return fmt.Errorf("runner error: %w", runnerErr)
	}
	if channelErr != nil {
		return fmt.Errorf("channel error: %w", channelErr)
	}

	return nil
}
