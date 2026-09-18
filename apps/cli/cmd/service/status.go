package service

import (
	"fmt"
	"time"

	"beep/internal/autostart"
	"beep/internal/cliservice"
	"beep/internal/cmdutil"
	"beep/internal/config"
	"beep/internal/daemon"
	"beep/internal/ui"

	"github.com/spf13/cobra"
)

// NewCmdStatus creates the 'service status' subcommand.
func NewCmdStatus() *cobra.Command {
	return &cobra.Command{
		Use:   "status [runner|channel]",
		Short: "Check running status and information of Beep services",
		Args:  cobra.MaximumNArgs(1),
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
				return cliservice.ShowSingleServiceStatus(daemon.ServiceRunner, cfg)
			case "channel":
				return cliservice.ShowSingleServiceStatus(daemon.ServiceChannel, cfg)
			default:
				return runStatusAll(cfg)
			}
		},
	}
}

func runStatusAll(cfg *config.Config) error {
	today := time.Now().Format("2006-01-02")

	runnerStatus, err := daemon.GetDaemonStatus(cfg.Workspace, daemon.ServiceRunner)
	if err != nil {
		return fmt.Errorf("failed to query runner daemon status: %w", err)
	}

	channelStatus, err := daemon.GetDaemonStatus(cfg.Workspace, daemon.ServiceChannel)
	if err != nil {
		return fmt.Errorf("failed to query channel daemon status: %w", err)
	}

	mgr := autostart.CurrentManager()
	runnerAutostart := cliservice.FormatAutostartStatus(mgr.GetStatus(daemon.ServiceRunner, config.BinaryName()))
	channelAutostart := cliservice.FormatAutostartStatus(mgr.GetStatus(daemon.ServiceChannel, config.BinaryName()))

	// Runner Section
	fmt.Println(ui.Bold("● Runner Service:"))
	if runnerStatus != nil && runnerStatus.PID > 0 {
		uptime := ""
		if t, err := time.Parse(time.RFC3339, runnerStatus.StartTime); err == nil {
			uptime = time.Since(t).Round(time.Second).String()
		}
		fmt.Printf("  %s %s (PID: %s)\n", ui.Green("● running"), ui.Dim(uptime), ui.Cyan(fmt.Sprintf("%d", runnerStatus.PID)))
	} else {
		fmt.Printf("  %s\n", ui.Dim("○ stopped"))
	}
	fmt.Printf("  %s %s\n", ui.Dim("Autostart:"), runnerAutostart)
	fmt.Printf("  %s %s\n", ui.Dim("Socket:   "), daemon.SocketPath(cfg.Workspace, daemon.ServiceRunner))
	fmt.Printf("  %s %s\n", ui.Dim("Logs:     "), daemon.DailyLogPath(cfg.Workspace, daemon.ServiceRunner, today))
	if cfg.RunnerToken != "" {
		fmt.Printf("  %s %s\n", ui.Dim("Token:    "), ui.Yellow(config.MaskToken(cfg.RunnerToken)))
	} else {
		fmt.Printf("  %s %s\n", ui.Dim("Token:    "), ui.Dim("(not configured)"))
	}
	fmt.Println()

	// Channel Section
	fmt.Println(ui.Bold("● Channel Service:"))
	if channelStatus != nil && channelStatus.PID > 0 {
		uptime := ""
		if t, err := time.Parse(time.RFC3339, channelStatus.StartTime); err == nil {
			uptime = time.Since(t).Round(time.Second).String()
		}
		fmt.Printf("  %s %s (PID: %s)\n", ui.Green("● running"), ui.Dim(uptime), ui.Cyan(fmt.Sprintf("%d", channelStatus.PID)))
	} else {
		fmt.Printf("  %s\n", ui.Dim("○ stopped"))
	}
	fmt.Printf("  %s %s\n", ui.Dim("Autostart:"), channelAutostart)
	fmt.Printf("  %s %s\n", ui.Dim("Socket:   "), daemon.SocketPath(cfg.Workspace, daemon.ServiceChannel))
	fmt.Printf("  %s %s\n", ui.Dim("Logs:     "), daemon.DailyLogPath(cfg.Workspace, daemon.ServiceChannel, today))
	chToken := cfg.ChannelAuthToken()
	if chToken != "" {
		fmt.Printf("  %s %s\n", ui.Dim("Token:    "), ui.Yellow(config.MaskToken(chToken)))
	} else {
		fmt.Printf("  %s %s\n", ui.Dim("Token:    "), ui.Dim("(not configured)"))
	}

	if (runnerStatus == nil || runnerStatus.PID <= 0) && (channelStatus == nil || channelStatus.PID <= 0) {
		fmt.Println()
		fmt.Println(ui.Section("Start commands:"))
		binName := config.BinaryName()
		fmt.Printf("  All services:  %s  (background: %s)\n", ui.Green(binName+" service start"), ui.Cyan(binName+" service start -d"))
		fmt.Printf("  Runner only:   %s  (background: %s)\n", ui.Green(binName+" service start runner"), ui.Cyan(binName+" service start runner -d"))
		fmt.Printf("  Channel only:  %s  (background: %s)\n", ui.Green(binName+" service start channel"), ui.Cyan(binName+" service start channel -d"))
	}

	return nil
}
