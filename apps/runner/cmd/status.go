package cmd

import (
	"fmt"
	"path/filepath"
	"time"

	"beep-runner/internal/config"
	"beep-runner/internal/daemon"
	"beep-runner/internal/ui"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check runner daemon running status and information",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runStatus(cmd, args)
	},
}

func init() {
	RootCmd.AddCommand(statusCmd)
}

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
	logFile := filepath.Join(cfg.Workspace, "logs", fmt.Sprintf("beep-runner-%s.log", today))
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
			fmt.Println(ui.KeyValue("Token", ui.Yellow(config.MaskToken(cfg.RunnerToken))))
		}
	} else {
		fmt.Println(ui.KeyValue("Status", ui.Dim("stopped")+" "+ui.Dim("○")))
		fmt.Println(ui.KeyValue("Workspace", ui.Dim(cfg.Workspace)))
		fmt.Println(ui.KeyValue("Socket", ui.Dim(socketFile)))
		fmt.Println(ui.KeyValue("Logs", ui.Dim(filepath.Join(cfg.Workspace, "logs"))))
		if cfg.ServerURL != "" {
			fmt.Println(ui.KeyValue("Server", ui.Bold(cfg.ServerURL)))
		}
		fmt.Println()
		fmt.Println(ui.Section("Start commands:"))
		fmt.Printf("  Foreground: %s\n", ui.Green("beep-runner up"))
		fmt.Printf("  Background: %s\n", ui.Cyan("beep-runner up -d"))
	}
	return nil
}
