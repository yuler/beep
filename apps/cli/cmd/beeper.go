package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"beep/internal/client"
	"beep/internal/schedule"
	"beep/internal/ui"
	"beep/internal/workspace"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var (
	flagBeeperApp      string
	flagBeeperTitle    string
	flagBeeperBody     string
	flagBeeperCron     string
	flagBeeperTimezone string
	flagBeeperChannels string
	flagBeeperConfigs  []string
)

var beeperCmd = &cobra.Command{
	Use:     "beeper",
	Aliases: []string{"beepers"},
	Short:   "Manage monitor probe beepers",
	Long:    ui.Bold(ui.Cyan("Beeper Management")) + ` - Manage monitoring probes, view catalog apps, and check probe runs.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var beeperListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List monitor beepers",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
		defer cancel()

		if _, err := ensureLoggedIn(ctx, cfg); err != nil {
			return err
		}

		c := client.New(cfg)
		beepers, err := c.ListBeepers(ctx)
		if err != nil {
			return err
		}

		if flagJSON {
			data, err := json.MarshalIndent(beepers, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return nil
		}

		accountDisplay := cfg.AccountSlug
		if accountDisplay == "" {
			accountDisplay = "personal"
		}
		fmt.Printf("%s %s\n\n", ui.Bold(ui.Cyan("Beepers")), ui.Dim(fmt.Sprintf("(account: %s)", accountDisplay)))

		if len(beepers) == 0 {
			fmt.Println(ui.Dim("  No beepers found. Create one with 'beep beeper create'."))
			return nil
		}

		// Print table header
		fmt.Printf("  %-10s  %-24s  %-16s  %-10s  %-12s  %-16s  %s\n",
			ui.Dim("ID"),
			ui.Dim("TITLE"),
			ui.Dim("APP"),
			ui.Dim("STATUS"),
			ui.Dim("STATE"),
			ui.Dim("CRON"),
			ui.Dim("RECENT RUNS"),
		)

		for _, b := range beepers {
			title := b.Title
			if len(title) > 24 {
				title = title[:21] + "..."
			}

			appName := "-"
			if b.BeeperApp != nil {
				appName = b.BeeperApp.Slug
			}

			statsStr := "-"
			if b.RunStats != nil {
				statsStr = fmt.Sprintf("%d/%d ok", b.RunStats.Succeeded, b.RunStats.Total)
			}

			fmt.Printf("  %-10s  %-24s  %-16s  %-10s  %-12s  %-16s  %s\n",
				b.ID,
				title,
				appName,
				formatBeeperStatus(b.Status),
				formatAlertState(b.AlertState),
				b.Cron,
				statsStr,
			)
		}
		fmt.Println()
		return nil
	},
}

var beeperShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show details of a monitor beeper",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
		defer cancel()

		if _, err := ensureLoggedIn(ctx, cfg); err != nil {
			return err
		}

		c := client.New(cfg)
		b, err := c.GetBeeper(ctx, args[0])
		if err != nil {
			return err
		}

		if flagJSON {
			data, err := json.MarshalIndent(b, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return nil
		}

		fmt.Println()
		fmt.Printf("  %s %s\n", ui.Bold("Beeper:"), ui.Cyan(b.Title))
		fmt.Println(ui.KeyValue("ID", b.ID))
		if b.BeeperApp != nil {
			fmt.Println(ui.KeyValue("App", fmt.Sprintf("%s (%s)", b.BeeperApp.Name, b.BeeperApp.Slug)))
		}
		fmt.Println(ui.KeyValue("Status", formatBeeperStatus(b.Status)))
		fmt.Println(ui.KeyValue("Alert State", formatAlertState(b.AlertState)))
		fmt.Println(ui.KeyValue("Cron", b.Cron))
		if b.Timezone != "" {
			fmt.Println(ui.KeyValue("Timezone", b.Timezone))
		}
		if b.ConsecutiveFailures > 0 {
			fmt.Println(ui.KeyValue("Failures", fmt.Sprintf("%d consecutive", b.ConsecutiveFailures)))
		}
		if b.PingToken != "" {
			fmt.Println(ui.KeyValue("Ping Token", ui.Yellow(b.PingToken)))
		}
		if b.LastPingAt != "" {
			fmt.Println(ui.KeyValue("Last Ping", b.LastPingAt))
		}
		if b.NextRunAt != "" {
			fmt.Println(ui.KeyValue("Next Run", b.NextRunAt))
		}
		if b.LastRunAt != "" {
			fmt.Println(ui.KeyValue("Last Run", b.LastRunAt))
		}
		if len(b.NotificationChannels) > 0 {
			fmt.Println(ui.KeyValue("Channels", strings.Join(b.NotificationChannels, ", ")))
		}

		if len(b.Config) > 0 {
			fmt.Println()
			fmt.Println(ui.Section("  Configuration:"))
			for k, v := range b.Config {
				fmt.Printf("    %-20s %v\n", ui.Cyan(k+":"), v)
			}
		}

		if b.RunStats != nil {
			fmt.Println()
			fmt.Println(ui.Section("  Run Statistics:"))
			fmt.Println(ui.KeyValue("Total Runs", fmt.Sprintf("%d", b.RunStats.Total)))
			fmt.Println(ui.KeyValue("Succeeded", fmt.Sprintf("%d", b.RunStats.Succeeded)))
		}

		if len(b.Runs) > 0 {
			fmt.Println()
			fmt.Println(ui.Section("  Recent Runs:"))
			fmt.Printf("    %-10s  %-24s  %-10s  %-12s\n",
				ui.Dim("RUN ID"),
				ui.Dim("SCHEDULED FOR"),
				ui.Dim("STATUS"),
				ui.Dim("SIGNAL"),
			)
			for _, r := range b.Runs {
				fmt.Printf("    %-10s  %-24s  %-10s  %-12s\n",
					r.ID,
					r.ScheduledFor,
					formatRunStatus(r.Status),
					r.SignalStatus,
				)
			}
		}
		fmt.Println()
		return nil
	},
}

var beeperAppsCmd = &cobra.Command{
	Use:     "apps",
	Aliases: []string{"catalog"},
	Short:   "List available monitor probe app templates",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
		defer cancel()

		c := client.New(cfg)
		apps, err := c.ListBeeperApps(ctx)
		if err != nil {
			return err
		}

		if flagJSON {
			data, err := json.MarshalIndent(apps, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return nil
		}

		fmt.Printf("%s\n\n", ui.Bold(ui.Cyan("Available Beeper Probe Apps")))
		fmt.Printf("  %-20s  %-24s  %-14s  %s\n",
			ui.Dim("SLUG"),
			ui.Dim("NAME"),
			ui.Dim("DEFAULT CRON"),
			ui.Dim("DESCRIPTION"),
		)

		for _, a := range apps {
			fmt.Printf("  %-20s  %-24s  %-14s  %s\n",
				ui.Cyan(a.Slug),
				a.Name,
				a.DefaultCron,
				a.Description,
			)
		}
		fmt.Println()
		return nil
	},
}

var beeperCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create and install a monitor beeper",
	Long: `Create and install a monitor probe beeper.

Examples:
  # Create a heartbeat ping probe
  beep beeper create --app heartbeat-ping --title "Production Gateway Ping"

  # Create an HTTP probe with custom cron and config
  beep beeper create --app http-uptime --title "API Check" --cron "*/2 * * * *" --config url=https://api.example.com --config timeout=5s

  # Interactive wizard (when --app is omitted)
  beep beeper create`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
		defer cancel()

		if _, err := ensureLoggedIn(ctx, cfg); err != nil {
			return err
		}

		c := client.New(cfg)

		appSlug := strings.TrimSpace(flagBeeperApp)
		title := strings.TrimSpace(flagBeeperTitle)
		cron := strings.TrimSpace(flagBeeperCron)
		configMap := make(map[string]any)

		// Parse key=value configs
		for _, item := range flagBeeperConfigs {
			parts := strings.SplitN(item, "=", 2)
			if len(parts) == 2 {
				configMap[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}

		// Interactive fallback if missing app or title
		if appSlug == "" {
			if !flagNoInteractive && ui.IsInteractive() {
				apps, err := c.ListBeeperApps(ctx)
				if err != nil {
					return err
				}
				if len(apps) == 0 {
					return errors.New("no beeper apps available on server")
				}

				options := make([]huh.Option[string], 0, len(apps))
				for _, a := range apps {
					label := fmt.Sprintf("%s (%s)", a.Name, a.Slug)
					options = append(options, huh.NewOption(label, a.Slug))
				}

				selectErr := huh.NewSelect[string]().
					Title("Select Probe App Template").
					Options(options...).
					Value(&appSlug).
					Run()
				if selectErr != nil {
					return selectErr
				}
			} else {
				return errors.New("--app slug is required (view available apps with 'beep beeper apps')")
			}
		}

		// Fetch selected app to get default cron and input specs
		selectedApp, err := c.GetBeeperApp(ctx, appSlug)
		if err != nil {
			return fmt.Errorf("failed to fetch beeper app %s: %w", appSlug, err)
		}

		if cron == "" {
			cron = selectedApp.DefaultCron
		}

		if title == "" {
			if !flagNoInteractive && ui.IsInteractive() {
				defaultTitle := selectedApp.Name
				titleInput := defaultTitle
				err := huh.NewInput().
					Title("Beeper Title").
					Description("Name of this monitor probe").
					Value(&titleInput).
					Validate(func(s string) error {
						if strings.TrimSpace(s) == "" {
							return errors.New("title is required")
						}
						return nil
					}).
					Run()
				if err != nil {
					return err
				}
				title = strings.TrimSpace(titleInput)
			} else {
				title = selectedApp.Name
			}
		}

		// Check required inputs from selectedApp
		if !flagNoInteractive && ui.IsInteractive() {
			for _, input := range selectedApp.Inputs {
				if _, ok := configMap[input.Name]; !ok && input.Required {
					var val string
					if input.Default != nil {
						val = fmt.Sprintf("%v", input.Default)
					}
					promptErr := huh.NewInput().
						Title(fmt.Sprintf("Input: %s", input.Name)).
						Description(input.Description).
						Value(&val).
						Validate(func(s string) error {
							if input.Required && strings.TrimSpace(s) == "" {
								return fmt.Errorf("%s is required", input.Name)
							}
							return nil
						}).
						Run()
					if promptErr != nil {
						return promptErr
					}
					configMap[input.Name] = strings.TrimSpace(val)
				}
			}
		}

		if cron != "" {
			if err := schedule.Validate(cron); err != nil {
				return fmt.Errorf("invalid cron expression: %w", err)
			}
		}

		tz := flagBeeperTimezone
		if tz == "" {
			if detected, ok := workspace.DetectTimezoneOK(); ok {
				tz = detected
			} else {
				tz = "UTC"
			}
		}

		req := &client.CreateBeeperRequest{
			BeeperAppSlug: appSlug,
			Title:         title,
			Body:          flagBeeperBody,
			Cron:          cron,
			Timezone:      tz,
			Config:        configMap,
		}

		if flagBeeperChannels != "" {
			for _, ch := range strings.Split(flagBeeperChannels, ",") {
				if trimmed := strings.TrimSpace(ch); trimmed != "" {
					req.NotificationChannels = append(req.NotificationChannels, trimmed)
				}
			}
		}

		b, err := c.CreateBeeper(ctx, req)
		if err != nil {
			return err
		}

		if flagJSON {
			data, err := json.MarshalIndent(b, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return nil
		}

		fmt.Println(ui.Success("Created beeper %s (%s)", ui.Bold(b.Title), ui.Dim(b.ID)))
		fmt.Printf("  %s %s\n", ui.Dim("App:"), appSlug)
		fmt.Printf("  %s %s\n", ui.Dim("Cron:"), b.Cron)
		if b.PingToken != "" {
			fmt.Printf("  %s %s\n", ui.Dim("Ping Token:"), ui.Yellow(b.PingToken))
			fmt.Printf("  %s %s/api/v1/beeper_apps/heartbeat/pings/%s\n",
				ui.Dim("Ping URL:"),
				cfg.ServerURL,
				b.PingToken,
			)
		}
		return nil
	},
}

var beeperDeleteCmd = &cobra.Command{
	Use:     "delete <id>",
	Aliases: []string{"rm"},
	Short:   "Delete a monitor beeper",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
		defer cancel()

		if _, err := ensureLoggedIn(ctx, cfg); err != nil {
			return err
		}

		c := client.New(cfg)
		if err := c.DeleteBeeper(ctx, args[0]); err != nil {
			return err
		}

		if flagJSON {
			fmt.Printf("{\"id\":%q,\"deleted\":true}\n", args[0])
			return nil
		}

		fmt.Println(ui.Success("Deleted beeper %s", ui.Bold(args[0])))
		return nil
	},
}

var beeperPauseCmd = &cobra.Command{
	Use:   "pause <id>",
	Short: "Pause a monitor beeper",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
		defer cancel()

		if _, err := ensureLoggedIn(ctx, cfg); err != nil {
			return err
		}

		c := client.New(cfg)
		b, err := c.PauseBeeper(ctx, args[0])
		if err != nil {
			return err
		}

		if flagJSON {
			data, _ := json.MarshalIndent(b, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		fmt.Println(ui.Success("Paused beeper %s (%s)", ui.Bold(b.Title), ui.Dim(b.ID)))
		return nil
	},
}

var beeperResumeCmd = &cobra.Command{
	Use:   "resume <id>",
	Short: "Resume a monitor beeper",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
		defer cancel()

		if _, err := ensureLoggedIn(ctx, cfg); err != nil {
			return err
		}

		c := client.New(cfg)
		b, err := c.ResumeBeeper(ctx, args[0])
		if err != nil {
			return err
		}

		if flagJSON {
			data, _ := json.MarshalIndent(b, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		fmt.Println(ui.Success("Resumed beeper %s (%s)", ui.Bold(b.Title), ui.Dim(b.ID)))
		return nil
	},
}

var beeperRunCmd = &cobra.Command{
	Use:   "run <id>",
	Short: "Immediately trigger a probe run for a beeper",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
		defer cancel()

		if _, err := ensureLoggedIn(ctx, cfg); err != nil {
			return err
		}

		c := client.New(cfg)
		run, err := c.RunBeeper(ctx, args[0])
		if err != nil {
			return err
		}

		if flagJSON {
			data, _ := json.MarshalIndent(run, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		fmt.Println(ui.Success("Triggered run for beeper %s (Run ID: %s, Status: %s)", ui.Bold(args[0]), ui.Cyan(run.ID), formatRunStatus(run.Status)))
		return nil
	},
}

var beeperRunsCmd = &cobra.Command{
	Use:   "runs <id>",
	Short: "View probe execution history for a beeper",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
		defer cancel()

		if _, err := ensureLoggedIn(ctx, cfg); err != nil {
			return err
		}

		c := client.New(cfg)
		runs, err := c.ListBeeperRuns(ctx, args[0])
		if err != nil {
			return err
		}

		if flagJSON {
			data, err := json.MarshalIndent(runs, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return nil
		}

		fmt.Printf("%s %s\n\n", ui.Bold(ui.Cyan("Beeper Runs")), ui.Dim(fmt.Sprintf("(beeper ID: %s)", args[0])))

		if len(runs) == 0 {
			fmt.Println(ui.Dim("  No runs found for this beeper."))
			return nil
		}

		fmt.Printf("  %-10s  %-24s  %-10s  %-14s\n",
			ui.Dim("RUN ID"),
			ui.Dim("SCHEDULED FOR"),
			ui.Dim("STATUS"),
			ui.Dim("SIGNAL STATUS"),
		)

		for _, r := range runs {
			fmt.Printf("  %-10s  %-24s  %-10s  %-14s\n",
				r.ID,
				r.ScheduledFor,
				formatRunStatus(r.Status),
				r.SignalStatus,
			)
		}
		fmt.Println()
		return nil
	},
}

func formatBeeperStatus(s string) string {
	switch strings.ToLower(s) {
	case "active":
		return ui.Green("active")
	case "paused":
		return ui.Yellow("paused")
	default:
		return s
	}
}

func formatAlertState(s string) string {
	switch strings.ToLower(s) {
	case "ok":
		return ui.Green("OK")
	case "alerting":
		return ui.Red("ALERTING")
	case "pending":
		return ui.Yellow("PENDING")
	case "recovering":
		return ui.Cyan("RECOVERING")
	default:
		return s
	}
}

func init() {
	beeperCreateCmd.Flags().StringVar(&flagBeeperApp, "app", "", "Beeper app slug (e.g. 'heartbeat-ping')")
	beeperCreateCmd.Flags().StringVar(&flagBeeperTitle, "title", "", "Beeper probe title")
	beeperCreateCmd.Flags().StringVarP(&flagBeeperBody, "body", "b", "", "Beeper description / body")
	beeperCreateCmd.Flags().StringVarP(&flagBeeperCron, "cron", "c", "", "Cron schedule (defaults to app's default cron)")
	beeperCreateCmd.Flags().StringVarP(&flagBeeperTimezone, "timezone", "z", "", "Timezone (defaults to local timezone)")
	beeperCreateCmd.Flags().StringVar(&flagBeeperChannels, "channels", "", "Comma-separated notification channel names or IDs")
	beeperCreateCmd.Flags().StringSliceVar(&flagBeeperConfigs, "config", nil, "Configuration parameters in key=value format (repeatable)")

	beeperCmd.AddCommand(beeperListCmd)
	beeperCmd.AddCommand(beeperShowCmd)
	beeperCmd.AddCommand(beeperAppsCmd)
	beeperCmd.AddCommand(beeperCreateCmd)
	beeperCmd.AddCommand(beeperDeleteCmd)
	beeperCmd.AddCommand(beeperPauseCmd)
	beeperCmd.AddCommand(beeperResumeCmd)
	beeperCmd.AddCommand(beeperRunCmd)
	beeperCmd.AddCommand(beeperRunsCmd)
}
