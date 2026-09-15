package cmd

import (
	"fmt"
	"time"

	"beep/internal/config"
	"beep/internal/daemon"
	"beep/internal/ui"

	"github.com/spf13/cobra"
)

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Check running status and information of Beep services",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(cmd, args)
		},
	}
}

var statusCmd = newStatusCmd()

func showSingleServiceStatus(service string, cfg *config.Config) error {
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
			chToken := cfg.ChannelToken
			if chToken == "" {
				chToken = cfg.CliToken
			}
			if chToken == "" {
				chToken = cfg.DeviceToken
			}
			if chToken != "" {
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
			chToken := cfg.ChannelToken
			if chToken == "" {
				chToken = cfg.CliToken
			}
			if chToken == "" {
				chToken = cfg.DeviceToken
			}
			if chToken != "" {
				fmt.Println(ui.KeyValue("Channel Token", ui.Yellow(config.MaskToken(chToken))))
			}
		}
		fmt.Println()
		fmt.Println(ui.Section("Start commands:"))
		fmt.Printf("  Foreground: %s\n", ui.Green(fmt.Sprintf("beep %s up", service)))
		fmt.Printf("  Background: %s\n", ui.Cyan(fmt.Sprintf("beep %s up -d", service)))
	}
	return nil
}

func runStatus(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	today := time.Now().Format("2006-01-02")

	runnerStatus, err := daemon.GetDaemonStatus(cfg.Workspace, daemon.ServiceRunner)
	if err != nil {
		return fmt.Errorf("failed to query runner daemon status: %w", err)
	}

	channelStatus, err := daemon.GetDaemonStatus(cfg.Workspace, daemon.ServiceChannel)
	if err != nil {
		return fmt.Errorf("failed to query channel daemon status: %w", err)
	}

	fmt.Println(ui.Bold(ui.Cyan("Beep Status:")))
	fmt.Println(ui.KeyValue("Workspace", ui.Dim(cfg.Workspace)))
	if cfg.ServerURL != "" {
		fmt.Println(ui.KeyValue("Server", ui.Bold(cfg.ServerURL)))
	}
	printAuthStatus(cfg)
	fmt.Println()

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
	fmt.Printf("  %s %s\n", ui.Dim("Socket:"), daemon.SocketPath(cfg.Workspace, daemon.ServiceRunner))
	fmt.Printf("  %s %s\n", ui.Dim("Logs:  "), daemon.DailyLogPath(cfg.Workspace, daemon.ServiceRunner, today))
	if cfg.RunnerToken != "" {
		fmt.Printf("  %s %s\n", ui.Dim("Token: "), ui.Yellow(config.MaskToken(cfg.RunnerToken)))
	} else {
		fmt.Printf("  %s %s\n", ui.Dim("Token: "), ui.Dim("(not configured)"))
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
	fmt.Printf("  %s %s\n", ui.Dim("Socket:"), daemon.SocketPath(cfg.Workspace, daemon.ServiceChannel))
	fmt.Printf("  %s %s\n", ui.Dim("Logs:  "), daemon.DailyLogPath(cfg.Workspace, daemon.ServiceChannel, today))
	chToken := cfg.ChannelToken
	if chToken == "" {
		chToken = cfg.CliToken
	}
	if chToken == "" {
		chToken = cfg.DeviceToken
	}
	if chToken != "" {
		fmt.Printf("  %s %s\n", ui.Dim("Token: "), ui.Yellow(config.MaskToken(chToken)))
	} else {
		fmt.Printf("  %s %s\n", ui.Dim("Token: "), ui.Dim("(not configured)"))
	}

	if (runnerStatus == nil || runnerStatus.PID <= 0) && (channelStatus == nil || channelStatus.PID <= 0) {
		fmt.Println()
		fmt.Println(ui.Section("Start commands:"))
		fmt.Printf("  All services:  %s  (background: %s)\n", ui.Green("beep up"), ui.Cyan("beep up -d"))
		fmt.Printf("  Runner only:   %s  (background: %s)\n", ui.Green("beep runner up"), ui.Cyan("beep runner up -d"))
		fmt.Printf("  Channel only:  %s  (background: %s)\n", ui.Green("beep channel up"), ui.Cyan("beep channel up -d"))
	}

	return nil
}

func printAuthStatus(cfg *config.Config) {
	if cfg.IsLoggedIn() {
		fmt.Println(ui.KeyValue("Auth", ui.Green("logged in")+" "+ui.Green("●")))
		user := cfg.UserEmail
		if cfg.UserName != "" && cfg.UserEmail != "" {
			user = fmt.Sprintf("%s (%s)", cfg.UserName, cfg.UserEmail)
		} else if cfg.UserName != "" {
			user = cfg.UserName
		}
		if user != "" {
			fmt.Println(ui.KeyValue("User", ui.Cyan(user)))
		}
		if cfg.AccountSlug != "" {
			fmt.Println(ui.KeyValue("Account", ui.Yellow(cfg.AccountSlug)))
		}
		if cfg.AccessToken != "" {
			fmt.Println(ui.KeyValue("Token", ui.Yellow(config.MaskToken(cfg.AccessToken))))
		}
	} else {
		fmt.Println(ui.KeyValue("Auth", ui.Dim("not logged in")+" "+ui.Dim("○")))
	}
}
