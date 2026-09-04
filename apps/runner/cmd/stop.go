package cmd

import (
	"fmt"
	"time"

	"beep-runner/internal/daemon"
	"beep-runner/internal/ui"

	"github.com/spf13/cobra"
)

var (
	flagStopForce   bool
	flagStopTimeout time.Duration
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the running runner daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runStop(cmd, args)
	},
}

func init() {
	stopCmd.Flags().BoolVarP(&flagStopForce, "force", "f", false, "Forcibly kill the daemon process if graceful stop times out")
	stopCmd.Flags().DurationVar(&flagStopTimeout, "timeout", 10*time.Second, "Timeout waiting for daemon to stop")
	RootCmd.AddCommand(stopCmd)
}

func runStop(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	running, pid, err := daemon.CheckRunning(cfg.Workspace)
	if err != nil {
		return fmt.Errorf("failed to check daemon status: %w", err)
	}

	if !running || pid == 0 {
		fmt.Println(ui.Info("Runner daemon is not running in workspace: %s", ui.Dim(cfg.Workspace)))
		return nil
	}

	fmt.Printf("%s Stopping runner daemon (PID: %s)...\n",
		ui.Cyan("●"),
		ui.Cyan(fmt.Sprintf("%d", pid)),
	)

	stoppedPid, err := daemon.StopDaemon(cfg.Workspace, flagStopTimeout, flagStopForce)
	if err != nil {
		return err
	}

	fmt.Printf("%s %s (PID: %s)\n",
		ui.Green("✓"),
		ui.Bold("Beep runner stopped"),
		ui.Dim(fmt.Sprintf("%d", stoppedPid)),
	)
	return nil
}
