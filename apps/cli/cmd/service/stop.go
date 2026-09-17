package service

import (
	"fmt"
	"strings"
	"time"

	"beep/internal/cliservice"
	"beep/internal/cmdutil"
	"beep/internal/daemon"
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
			cfg, err := cmdutil.LoadConfig(cmd)
			if err != nil {
				return err
			}

			target := ""
			if len(args) > 0 {
				target = strings.ToLower(args[0])
			}

			switch target {
			case "runner":
				return cliservice.StopSingleService(daemon.ServiceRunner, cfg.Workspace, timeout, force)
			case "channel":
				return cliservice.StopSingleService(daemon.ServiceChannel, cfg.Workspace, timeout, force)
			case "":
				return runStopAll(cfg.Workspace, timeout, force)
			default:
				return fmt.Errorf("unknown service %q (expected 'runner' or 'channel')", target)
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
		if err := cliservice.StopSingleService(daemon.ServiceRunner, workspaceDir, timeout, force); err != nil {
			return err
		}
	}

	if channelRunning {
		if err := cliservice.StopSingleService(daemon.ServiceChannel, workspaceDir, timeout, force); err != nil {
			return err
		}
	}

	return nil
}
