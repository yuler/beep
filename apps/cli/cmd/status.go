package cmd

import (
	"fmt"
	"path/filepath"
	"time"

	"beep/internal/config"
	"beep/internal/daemon"
	"beep/internal/ui"

	"github.com/spf13/cobra"
)

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Check daemon running status and information",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(cmd, args)
		},
	}
}

var statusCmd = newStatusCmd()

func runStatus(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	status, err := daemon.GetDaemonStatus(cfg.Workspace)
	if err != nil {
		return fmt.Errorf("failed to query daemon status: %w", err)
	}

	today := time.Now().Format("2006-01-02")
	logFile := daemon.DailyLogPath(cfg.Workspace, today)
	socketFile := daemon.SocketPath(cfg.Workspace)

	fmt.Println(ui.Bold(ui.Cyan("Beep Runner Daemon Status:")))

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
		if cfg.RunnerToken != "" {
			fmt.Println(ui.KeyValue("Runner Token", ui.Yellow(config.MaskToken(cfg.RunnerToken))))
		}
		if cfg.ChannelToken != "" {
			fmt.Println(ui.KeyValue("Channel Token", ui.Yellow(config.MaskToken(cfg.ChannelToken))))
		}
	} else {
		fmt.Println(ui.KeyValue("Status", ui.Dim("stopped")+" "+ui.Dim("○")))
		fmt.Println(ui.KeyValue("Workspace", ui.Dim(cfg.Workspace)))
		fmt.Println(ui.KeyValue("Socket", ui.Dim(socketFile)))
		fmt.Println(ui.KeyValue("Logs", ui.Dim(filepath.Join(cfg.Workspace, "logs"))))
		if cfg.ServerURL != "" {
			fmt.Println(ui.KeyValue("Server", ui.Bold(cfg.ServerURL)))
		}
		if cfg.RunnerToken != "" {
			fmt.Println(ui.KeyValue("Runner Token", ui.Yellow(config.MaskToken(cfg.RunnerToken))))
		}
		if cfg.ChannelToken != "" {
			fmt.Println(ui.KeyValue("Channel Token", ui.Yellow(config.MaskToken(cfg.ChannelToken))))
		}
		fmt.Println()
		fmt.Println(ui.Section("Start commands:"))
		fmt.Printf("  Foreground: %s\n", ui.Green("beep up"))
		fmt.Printf("  Background: %s\n", ui.Cyan("beep up -d"))
	}
	return nil
}
