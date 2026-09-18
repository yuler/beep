package service

import (
	"fmt"
	"time"

	"beep/internal/cmdutil"
	"beep/internal/daemon"
	intsvc "beep/internal/service"
	"beep/internal/ui"

	"github.com/spf13/cobra"
)

// NewCmdStop creates the 'service stop' subcommand.
func NewCmdStop() *cobra.Command {
	var (
		force   bool
		timeout time.Duration
	)

	cmd := &cobra.Command{
		Use:     "stop [runner|channel]",
		Aliases: []string{"down"},
		Short:   "Stop running Beep daemon services (runner and channel)",
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

			switch target {
			case "runner":
				return intsvc.StopSingleService(daemon.ServiceRunner, cfg.Workspace, timeout, force)
			case "channel":
				return intsvc.StopSingleService(daemon.ServiceChannel, cfg.Workspace, timeout, force)
			default:
				return runStopAll(cfg.Workspace, timeout, force)
			}
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Forcibly kill the daemon process if graceful stop times out")
	cmd.Flags().DurationVar(&timeout, "timeout", 10*time.Second, "Timeout waiting for daemon to stop")

	return cmd
}

func runStopAll(workspaceDir string, timeout time.Duration, force bool) error {
	runnerRunning, _, _ := daemon.CheckRunning(workspaceDir, daemon.ServiceRunner)
	channelRunning, _, _ := daemon.CheckRunning(workspaceDir, daemon.ServiceChannel)

	if !runnerRunning && !channelRunning {
		fmt.Println(ui.Info("No running Beep daemons found in workspace: %s", ui.Dim(workspaceDir)))
		return nil
	}

	if runnerRunning {
		if err := intsvc.StopSingleService(daemon.ServiceRunner, workspaceDir, timeout, force); err != nil {
			return err
		}
	}

	if channelRunning {
		if err := intsvc.StopSingleService(daemon.ServiceChannel, workspaceDir, timeout, force); err != nil {
			return err
		}
	}

	return nil
}
