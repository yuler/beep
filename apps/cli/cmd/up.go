package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"time"

	"beep/internal/channel"
	"beep/internal/config"
	"beep/internal/daemon"
	"beep/internal/proc"
	"beep/internal/runner"
	"beep/internal/ui"
	"beep/internal/workspace"

	"github.com/spf13/cobra"
)

var (
	flagConcurrency  int
	flagPollInterval time.Duration
	flagDaemon       bool
)

func newUpCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "up",
		Aliases: []string{"run"},
		Short:   "Start daemon services to listen for notifications and execute tasks",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUp(cmd, args)
		},
	}
	cmd.Flags().IntVarP(&flagConcurrency, "concurrency", "c", 0, "Max concurrent jobs for runner (default 5)")
	cmd.Flags().DurationVarP(&flagPollInterval, "poll-interval", "i", 0, "Poll interval (default 3s)")
	cmd.Flags().BoolVarP(&flagDaemon, "daemon", "d", false, "Run daemon in background")
	return cmd
}

var upCmd = newUpCmd()

func runUp(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	if flagConcurrency > 0 {
		cfg.Concurrency = flagConcurrency
	}
	if flagPollInterval > 0 {
		cfg.PollInterval = flagPollInterval
	}
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	hasRunner := cfg.RunnerToken != ""
	hasChannel := cfg.ChannelToken != "" || cfg.CliToken != "" || cfg.DeviceToken != ""

	if !hasRunner && !hasChannel {
		return fmt.Errorf("no services configured. To configure:\n  Runner:  set BEEP_RUNNER_TOKEN or configure config.json\n  Channel: run 'beep channel connect'")
	}

	isChild := os.Getenv("BEEP_DAEMON_CHILD") == "1"

	// If background daemon mode requested and not already the spawned child:
	if flagDaemon && !isChild {
		if hasRunner {
			if err := startServiceBackgroundDaemon(daemon.ServiceRunner, []string{"runner", "up"}, os.Args[1:], cfg); err != nil {
				return err
			}
		}
		if hasChannel {
			if err := startServiceBackgroundDaemon(daemon.ServiceChannel, []string{"channel", "up"}, os.Args[1:], cfg); err != nil {
				return err
			}
		}
		return nil
	}

	ws, err := workspace.Open(cfg.Workspace)
	if err != nil {
		return fmt.Errorf("workspace error: %w", err)
	}

	// Foreground combined mode: acquire sockets for all active services
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
			if sockRunner != nil {
				sockRunner.SetRunning()
			}
		}
		go func() {
			defer wg.Done()
			if err := r.Run(ctx); err != nil && ctx.Err() == nil {
				runnerErr = err
				cancel()
			}
		}()
	}

	if hasChannel {
		wg.Add(1)
		ch := channel.New(cfg, ws)
		ch.OnReady = func() {
			if sockChannel != nil {
				sockChannel.SetRunning()
			}
		}
		go func() {
			defer wg.Done()
			if err := ch.Run(ctx); err != nil && ctx.Err() == nil {
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

func runRunnerService(cfg *config.Config, daemonMode bool) error {
	if cfg.RunnerToken == "" {
		return fmt.Errorf("runner token is not configured (set via BEEP_RUNNER_TOKEN or config.json)")
	}

	isChild := os.Getenv("BEEP_DAEMON_CHILD") == "1"
	if daemonMode && !isChild {
		return startServiceBackgroundDaemon(daemon.ServiceRunner, []string{"runner", "up"}, os.Args[1:], cfg)
	}

	ws, err := workspace.Open(cfg.Workspace)
	if err != nil {
		return fmt.Errorf("workspace error: %w", err)
	}

	sock, err := daemon.AcquireSocket(cfg.Workspace, daemon.ServiceRunner)
	if err != nil {
		return err
	}
	defer sock.Close()

	logWriter, _, err := daemon.SetupLogger(cfg.Workspace, daemon.ServiceRunner, !isChild)
	if err != nil {
		return fmt.Errorf("failed to setup daily logger: %w", err)
	}
	defer logWriter.Close()

	r := runner.New(cfg, ws)
	r.OnReady = func() {
		sock.SetRunning()
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, proc.ShutdownSignals...)
	go func() {
		<-sigChan
		log.Println(ui.Dim("[beep-runner] Received termination signal..."))
		cancel()
	}()

	return r.Run(ctx)
}

func runChannelService(cfg *config.Config, daemonMode bool) error {
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

	isChild := os.Getenv("BEEP_DAEMON_CHILD") == "1"
	if daemonMode && !isChild {
		return startServiceBackgroundDaemon(daemon.ServiceChannel, []string{"channel", "up"}, os.Args[1:], cfg)
	}

	ws, err := workspace.Open(cfg.Workspace)
	if err != nil {
		return fmt.Errorf("workspace error: %w", err)
	}

	sock, err := daemon.AcquireSocket(cfg.Workspace, daemon.ServiceChannel)
	if err != nil {
		return err
	}
	defer sock.Close()

	logWriter, _, err := daemon.SetupLogger(cfg.Workspace, daemon.ServiceChannel, !isChild)
	if err != nil {
		return fmt.Errorf("failed to setup daily logger: %w", err)
	}
	defer logWriter.Close()

	ch := channel.New(cfg, ws)
	ch.OnReady = func() {
		sock.SetRunning()
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, proc.ShutdownSignals...)
	go func() {
		<-sigChan
		log.Println(ui.Dim("[beep-channel] Received termination signal..."))
		cancel()
	}()

	return ch.Run(ctx)
}

var startServiceDaemonFn = startServiceBackgroundDaemon

func autoStartServiceDaemon(service string, cfg *config.Config) error {
	running, pid, _ := daemon.CheckRunning(cfg.Workspace, service)
	if running && pid > 0 {
		fmt.Printf("%s Restarting %s daemon (stopping PID: %d)...\n", ui.Cyan("●"), service, pid)
		if _, err := daemon.StopDaemon(cfg.Workspace, service, 10*time.Second, true); err != nil {
			return fmt.Errorf("failed to stop existing %s daemon: %w", service, err)
		}
	}

	var rawArgs []string
	if cfg.Workspace != "" {
		rawArgs = append(rawArgs, "--workspace", cfg.Workspace)
	}
	if flagServer != "" {
		rawArgs = append(rawArgs, "--server", flagServer)
	}

	return startServiceDaemonFn(service, []string{service, "up"}, rawArgs, cfg)
}

func startServiceBackgroundDaemon(service string, childSubcommand []string, rawArgs []string, cfg *config.Config) error {
	running, pid, _ := daemon.CheckRunning(cfg.Workspace, service)
	if running {
		return fmt.Errorf("%s daemon is already running (PID: %d, socket: %s)", service, pid, daemon.SocketPath(cfg.Workspace, service))
	}

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to determine executable path: %w", err)
	}

	childArgs := buildServiceChildArgs(service, rawArgs)

	cmd := exec.Command(exe, childArgs...)
	cmd.Env = append(os.Environ(), "BEEP_DAEMON_CHILD=1")
	if service == "channel" {
		token := cfg.ChannelToken
		if token == "" {
			token = cfg.CliToken
		}
		if token == "" {
			token = cfg.DeviceToken
		}
		if token != "" {
			cmd.Env = append(cmd.Env, "BEEP_CHANNEL_TOKEN="+token)
		}
	} else if service == "runner" && cfg.RunnerToken != "" {
		cmd.Env = append(cmd.Env, "BEEP_RUNNER_TOKEN="+cfg.RunnerToken)
	}
	proc.Detach(cmd)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start background daemon for %s: %w", service, err)
	}

	today := time.Now().Format("2006-01-02")
	logFile := daemon.DailyLogPath(cfg.Workspace, service, today)

	for i := 0; i < 50; i++ {
		time.Sleep(100 * time.Millisecond)

		if isReady, childPID, _ := daemon.CheckReady(cfg.Workspace, service); isReady {
			if childPID <= 0 {
				childPID = cmd.Process.Pid
			}

			fmt.Printf("%s %s (PID: %s)\n",
				ui.Green("✓"),
				ui.Bold(fmt.Sprintf("Beep %s started in background", service)),
				ui.Cyan(fmt.Sprintf("%d", childPID)),
			)
			fmt.Printf("  %s %s\n", ui.Dim("Workspace:"), cfg.Workspace)
			fmt.Printf("  %s %s\n", ui.Dim("Logs:     "), logFile)
			fmt.Printf("  %s %s\n", ui.Dim("Socket:   "), daemon.SocketPath(cfg.Workspace, service))
			return nil
		}

		if exited, exitStatus := proc.Exited(cmd.Process.Pid); exited {
			return fmt.Errorf("%s daemon failed to start (exited with status %d, check logs: %s)", service, exitStatus, logFile)
		}
	}

	if cmd.Process != nil {
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
	}
	return fmt.Errorf("%s daemon failed to complete handshake within 5s (check logs: %s)", service, logFile)
}

func stripDaemonFlags(args []string) []string {
	var out []string
	for _, a := range args {
		if a == "-d" || a == "--daemon" || strings.HasPrefix(a, "--daemon=") || strings.HasPrefix(a, "-d=") {
			continue
		}
		out = append(out, a)
	}
	return out
}

func buildServiceChildArgs(service string, args []string) []string {
	stripped := stripDaemonFlags(args)
	var flags []string
	takesArg := false
	skipNext := false
	for _, arg := range stripped {
		if skipNext {
			skipNext = false
			continue
		}
		if takesArg {
			flags = append(flags, arg)
			takesArg = false
			continue
		}

		if arg == "runner" || arg == "channel" || arg == "up" || arg == "run" || arg == "connect" {
			continue
		}

		if arg == "-t" || arg == "--token" {
			skipNext = true
			continue
		}
		if strings.HasPrefix(arg, "--token=") || strings.HasPrefix(arg, "-t=") {
			continue
		}

		flags = append(flags, arg)
		if arg == "-w" || arg == "--workspace" ||
			arg == "-s" || arg == "--server" ||
			arg == "-c" || arg == "--concurrency" ||
			arg == "-i" || arg == "--poll-interval" ||
			arg == "--timeout" {
			takesArg = true
		}
	}
	return append([]string{service, "up"}, flags...)
}

func buildChildDaemonArgs(args []string) []string {
	return buildServiceChildArgs("runner", args)
}
