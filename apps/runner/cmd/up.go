package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"beep-runner/internal/config"
	"beep-runner/internal/daemon"
	"beep-runner/internal/ui"
	"beep-runner/internal/workspace"

	"github.com/spf13/cobra"
)

var (
	flagConcurrency  int
	flagPollInterval time.Duration
	flagDaemon       bool
)

var upCmd = &cobra.Command{
	Use:     "up",
	Aliases: []string{"run"},
	Short:   "Start the runner daemon to poll and execute scheduled tasks",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runUp(cmd, args)
	},
}

func init() {
	upCmd.Flags().IntVarP(&flagConcurrency, "concurrency", "c", 0, "Max concurrent jobs (default 5)")
	upCmd.Flags().DurationVarP(&flagPollInterval, "poll-interval", "i", 0, "Poll interval (default 3s)")
	upCmd.Flags().BoolVarP(&flagDaemon, "daemon", "d", false, "Run runner daemon in background")
}

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

	ws, err := workspace.Open(cfg.Workspace)
	if err != nil {
		return fmt.Errorf("workspace error: %w", err)
	}

	isChild := os.Getenv("BEEP_DAEMON_CHILD") == "1"

	// If background daemon mode requested and not already the spawned child:
	if flagDaemon && !isChild {
		return startBackgroundDaemon(cfg)
	}

	// Single instance control via Unix domain socket
	sock, err := daemon.AcquireSocket(cfg.Workspace)
	if err != nil {
		return err
	}
	defer sock.Close()

	// Daily rotating logger in $WORKSPACE/logs
	logWriter, _, err := daemon.SetupLogger(cfg.Workspace, !isChild)
	if err != nil {
		return fmt.Errorf("failed to setup daily logger: %w", err)
	}
	defer logWriter.Close()

	d := daemon.New(cfg, ws)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println(ui.Dim("[beep-runner] Received termination signal..."))
		cancel()
	}()

	return d.Start(ctx)
}

func startBackgroundDaemon(cfg *config.Config) error {
	running, pid, _ := daemon.CheckRunning(cfg.Workspace)
	if running {
		return fmt.Errorf("runner daemon is already running (PID: %d, socket: %s)", pid, daemon.SocketPath(cfg.Workspace))
	}

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to determine executable path: %w", err)
	}

	// Filter out daemon flags from args so child doesn't think it needs to spawn again
	childArgs := stripDaemonFlags(os.Args[1:])

	subcommands := map[string]bool{"up": true, "run": true, "ping": true, "version": true, "config": true, "job": true}
	hasSubcommand := false
	for _, a := range childArgs {
		if !strings.HasPrefix(a, "-") {
			if subcommands[a] {
				hasSubcommand = true
			}
			break
		}
	}
	if !hasSubcommand {
		childArgs = append([]string{"up"}, childArgs...)
	}

	cmd := exec.Command(exe, childArgs...)
	cmd.Env = append(os.Environ(), "BEEP_DAEMON_CHILD=1")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start background daemon: %w", err)
	}

	// Wait up to 3s for daemon to acquire socket
	started := false
	for i := 0; i < 30; i++ {
		time.Sleep(100 * time.Millisecond)
		if isRunning, childPID, _ := daemon.CheckRunning(cfg.Workspace); isRunning {
			started = true
			if childPID == 0 {
				childPID = cmd.Process.Pid
			}
			today := time.Now().Format("2006-01-02")
			logFile := filepath.Join(cfg.Workspace, "logs", fmt.Sprintf("beep-runner-%s.log", today))

			fmt.Printf("%s %s (PID: %s)\n",
				ui.Green("✓"),
				ui.Bold("Beep runner started in background"),
				ui.Cyan(fmt.Sprintf("%d", childPID)),
			)
			fmt.Printf("  %s %s\n", ui.Dim("Workspace:"), cfg.Workspace)
			fmt.Printf("  %s %s\n", ui.Dim("Logs:     "), logFile)
			fmt.Printf("  %s %s\n", ui.Dim("Socket:   "), daemon.SocketPath(cfg.Workspace))
			return nil
		}
	}

	if !started {
		return fmt.Errorf("daemon failed to initialize socket (check logs in %s/logs)", cfg.Workspace)
	}

	return nil
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
