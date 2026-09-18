package cliservice

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"time"

	"beep/internal/channel"
	"beep/internal/config"
	"beep/internal/daemon"
	"beep/internal/proc"
	"beep/internal/runner"
	"beep/internal/ui"
	"beep/internal/updater"
	"beep/internal/workspace"
)

// StartServiceDaemonFn is the function used to start background daemons. Can be replaced in tests.
var StartServiceDaemonFn = StartServiceBackgroundDaemon

// AutoStartServiceDaemon automatically stops an existing daemon and starts a new background daemon for the service.
func AutoStartServiceDaemon(service string, cfg *config.Config, rawArgs []string) error {
	running, pid, _ := daemon.CheckRunning(cfg.Workspace, service)
	if running && pid > 0 {
		fmt.Printf("%s Restarting %s daemon (stopping PID: %d)...\n", ui.Cyan("●"), service, pid)
		if _, err := daemon.StopDaemon(cfg.Workspace, service, 10*time.Second, true); err != nil {
			return fmt.Errorf("failed to stop existing %s daemon: %w", service, err)
		}
	}

	return StartServiceDaemonFn(service, []string{service, "up"}, rawArgs, cfg)
}

// StartServiceBackgroundDaemon starts a background daemon process for the specified service.
func StartServiceBackgroundDaemon(service string, childSubcommand []string, rawArgs []string, cfg *config.Config) error {
	running, pid, _ := daemon.CheckRunning(cfg.Workspace, service)
	if running {
		return fmt.Errorf("%s daemon is already running (PID: %d, socket: %s)", service, pid, daemon.SocketPath(cfg.Workspace, service))
	}

	exe, err := osExecutable()
	if err != nil {
		return fmt.Errorf("failed to determine executable path: %w", err)
	}

	childArgs := BuildServiceChildArgs(service, rawArgs)

	cmd := exec.Command(exe, childArgs...)
	cmd.Env = childServiceEnv(service, cfg, true)
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

// StopSingleService stops a running daemon for the given service.
func StopSingleService(service, workspaceDir string, timeout time.Duration, force bool) error {
	running, pid, err := daemon.CheckRunning(workspaceDir, service)
	if err != nil {
		return fmt.Errorf("failed to check %s daemon status: %w", service, err)
	}

	if !running || pid == 0 {
		fmt.Println(ui.Info("Beep %s daemon is not running in workspace: %s", service, ui.Dim(workspaceDir)))
		return nil
	}

	fmt.Printf("%s Stopping %s daemon (PID: %s)...\n",
		ui.Cyan("●"),
		service,
		ui.Cyan(fmt.Sprintf("%d", pid)),
	)

	stoppedPid, err := daemon.StopDaemon(workspaceDir, service, timeout, force)
	if err != nil {
		return err
	}

	fmt.Printf("%s %s (PID: %s)\n",
		ui.Green("✓"),
		ui.Bold(fmt.Sprintf("Beep %s stopped", service)),
		ui.Dim(fmt.Sprintf("%d", stoppedPid)),
	)
	return nil
}

// ShowSingleServiceStatus displays the daemon status for a specific service.
func ShowSingleServiceStatus(service string, cfg *config.Config) error {
	status, err := daemon.GetDaemonStatus(cfg.Workspace, service)
	if err != nil {
		return fmt.Errorf("failed to query %s daemon status: %w", service, err)
	}

	today := time.Now().Format("2006-01-02")
	logFile := daemon.DailyLogPath(cfg.Workspace, service, today)
	socketFile := daemon.SocketPath(cfg.Workspace, service)

	title := fmt.Sprintf("Beep %s Daemon Status:", service)
	if service == daemon.ServiceRunner {
		title = "Beep Runner Daemon Status:"
	} else if service == daemon.ServiceChannel {
		title = "Beep Channel Daemon Status:"
	}
	fmt.Println(ui.Bold(ui.Cyan(title)))

	if status != nil && status.PID > 0 {
		uptime := ""
		if t, err := time.Parse(time.RFC3339, status.StartTime); err == nil {
			uptime = time.Since(t).Round(time.Second).String()
		}

		fmt.Println(ui.KeyValue("Status", ui.Green("running")+" "+ui.Green("●")))
		fmt.Println(ui.KeyValue("PID", ui.Cyan(fmt.Sprintf("%d", status.PID))))
		if status.Version != "" {
			fmt.Println(ui.KeyValue("Version", ui.Bold(status.Version)))
		}
		if uptime != "" {
			fmt.Println(ui.KeyValue("Uptime", ui.Dim(uptime)))
		}
		fmt.Println(ui.KeyValue("Workspace", ui.Dim(cfg.Workspace)))
		fmt.Println(ui.KeyValue("Socket", ui.Dim(socketFile)))
		fmt.Println(ui.KeyValue("Logs", ui.Dim(logFile)))
		if cfg.ServerURL != "" {
			fmt.Println(ui.KeyValue("Server", ui.Bold(cfg.ServerURL)))
		}
		if service == daemon.ServiceRunner && cfg.RunnerToken != "" {
			fmt.Println(ui.KeyValue("Runner Token", ui.Yellow(config.MaskToken(cfg.RunnerToken))))
		}
		if service == daemon.ServiceChannel {
			if chToken := cfg.ChannelAuthToken(); chToken != "" {
				fmt.Println(ui.KeyValue("Channel Token", ui.Yellow(config.MaskToken(chToken))))
			}
		}
	} else {
		fmt.Println(ui.KeyValue("Status", ui.Dim("stopped")+" "+ui.Dim("○")))
		fmt.Println(ui.KeyValue("Workspace", ui.Dim(cfg.Workspace)))
		fmt.Println(ui.KeyValue("Socket", ui.Dim(socketFile)))
		fmt.Println(ui.KeyValue("Logs", ui.Dim(logFile)))
		if cfg.ServerURL != "" {
			fmt.Println(ui.KeyValue("Server", ui.Bold(cfg.ServerURL)))
		}
		if service == daemon.ServiceRunner && cfg.RunnerToken != "" {
			fmt.Println(ui.KeyValue("Runner Token", ui.Yellow(config.MaskToken(cfg.RunnerToken))))
		}
		if service == daemon.ServiceChannel {
			if chToken := cfg.ChannelAuthToken(); chToken != "" {
				fmt.Println(ui.KeyValue("Channel Token", ui.Yellow(config.MaskToken(chToken))))
			}
		}
		fmt.Println()
		fmt.Println(ui.Section("Start commands:"))
		fmt.Printf("  Foreground: %s\n", ui.Green(fmt.Sprintf("%s service start %s", config.BinaryName(), service)))
		fmt.Printf("  Background: %s\n", ui.Cyan(fmt.Sprintf("%s service start %s -d", config.BinaryName(), service)))
	}
	return nil
}

// RunRunnerService runs the runner daemon either in background or foreground.
func RunRunnerService(cfg *config.Config, daemonMode bool, rawArgs []string) error {
	if cfg.RunnerToken == "" {
		return fmt.Errorf("runner token is not configured (set via BEEP_RUNNER_TOKEN or config.json)")
	}

	isChild := os.Getenv("BEEP_DAEMON_CHILD") == "1"
	if daemonMode && !isChild {
		return StartServiceDaemonFn(daemon.ServiceRunner, []string{"runner", "up"}, rawArgs, cfg)
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
	go updater.RunDaemonUpdateProbe(ctx, cfg.Workspace, log.Printf)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, proc.ShutdownSignals...)
	go func() {
		<-sigChan
		log.Println(ui.Dim("[beep-runner] Received termination signal..."))
		cancel()
	}()

	return r.Run(ctx)
}

// RunChannelService runs the channel daemon either in background or foreground.
func RunChannelService(cfg *config.Config, daemonMode bool, rawArgs []string) error {
	token := cfg.ChannelAuthToken()
	if token == "" {
		return fmt.Errorf("channel token is not configured (run '%s channel connect')", config.BinaryName())
	}

	isChild := os.Getenv("BEEP_DAEMON_CHILD") == "1"
	if daemonMode && !isChild {
		return StartServiceDaemonFn(daemon.ServiceChannel, []string{"channel", "up"}, rawArgs, cfg)
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
	go updater.RunDaemonUpdateProbe(ctx, cfg.Workspace, log.Printf)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, proc.ShutdownSignals...)
	go func() {
		<-sigChan
		log.Println(ui.Dim("[beep-channel] Received termination signal..."))
		cancel()
	}()

	return ch.Run(ctx)
}

// StripDaemonFlags strips -d and --daemon flags.
func StripDaemonFlags(args []string) []string {
	var out []string
	for _, a := range args {
		if a == "-d" || a == "--daemon" || strings.HasPrefix(a, "--daemon=") || strings.HasPrefix(a, "-d=") {
			continue
		}
		out = append(out, a)
	}
	return out
}

// BuildServiceChildArgs builds the arguments passed to the spawned child process.
// Token flags are stripped so they do not appear in the child process list;
// StartServiceBackgroundDaemon injects tokens from cfg via the environment instead.
func BuildServiceChildArgs(service string, args []string) []string {
	stripped := StripDaemonFlags(args)
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
