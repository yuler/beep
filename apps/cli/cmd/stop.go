package cmd

import (
	"fmt"
	"time"

	"beep/internal/daemon"
	"beep/internal/ui"

	"github.com/spf13/cobra"
)

var (
	flagStopForce   bool
	flagStopTimeout time.Duration
)

func newStopCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop all running Beep daemon services (runner and channel)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStop(cmd, args)
		},
	}
	cmd.Flags().BoolVarP(&flagStopForce, "force", "f", false, "Forcibly kill the daemon process if graceful stop times out")
	cmd.Flags().DurationVar(&flagStopTimeout, "timeout", 10*time.Second, "Timeout waiting for daemon to stop")
	return cmd
}

var stopCmd = newStopCmd()

func stopSingleService(service, workspaceDir string, timeout time.Duration, force bool) error {
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

func runStop(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	runnerRunning, _, _ := daemon.CheckRunning(cfg.Workspace, daemon.ServiceRunner)
	channelRunning, _, _ := daemon.CheckRunning(cfg.Workspace, daemon.ServiceChannel)

	if !runnerRunning && !channelRunning {
		fmt.Println(ui.Info("No running Beep daemons found in workspace: %s", ui.Dim(cfg.Workspace)))
		return nil
	}

	if runnerRunning {
		if err := stopSingleService(daemon.ServiceRunner, cfg.Workspace, flagStopTimeout, flagStopForce); err != nil {
			return err
		}
	}

	if channelRunning {
		if err := stopSingleService(daemon.ServiceChannel, cfg.Workspace, flagStopTimeout, flagStopForce); err != nil {
			return err
		}
	}

	return nil
}
