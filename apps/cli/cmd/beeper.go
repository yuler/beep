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
	flagBeeperApp       string
	flagBeeperTitle     string
	flagBeeperBody      string
	flagBeeperCron      string
	flagBeeperTimezone  string
	flagBeeperChannels  string
	flagBeeperConfigs   []string
	flagBeeperShowToken bool
)

func maskToken(tok string) string {
	if tok == "" {
		return ""
	}
	if len(tok) <= 8 {
		return "••••••••"
	}
	prefix := tok[:min(len(tok), 8)]
	return prefix + "••••••••"
}

func redactBeeper(b *client.Beeper, revealToken bool) *client.Beeper {
	if b == nil {
		return nil
	}
	clone := *b
	if !revealToken {
		clone.PingToken = maskToken(clone.PingToken)
	}
	return &clone
}

func redactBeepers(list []*client.Beeper, revealToken bool) []*client.Beeper {
	res := make([]*client.Beeper, len(list))
	for i, b := range list {
		res[i] = redactBeeper(b, revealToken)
	}
	return res
}

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
			redacted := redactBeepers(beepers, false)
			data, err := json.MarshalIndent(redacted, "", "  ")
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
	Use:   "show [id]",
	Short: "Show details of a monitor beeper",
	Long: `Show details of a monitor beeper.

Note: By default, sensitive tokens like ping_token are masked (e.g. beep_pt_••••••••).
Pass --show-token to display the unmasked token in human view or --json output.`,
	Args: cobra.MaximumNArgs(1),
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
		id, err := resolveBeeperID(ctx, c, args, "show")
		if err != nil {
			return err
		}
		b, err := c.GetBeeper(ctx, id)
		if err != nil {
			return err
		}

		if flagJSON {
			data, err := json.MarshalIndent(redactBeeper(b, flagBeeperShowToken), "", "  ")
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
			tokDisplay := maskToken(b.PingToken)
			if flagBeeperShowToken {
				tokDisplay = b.PingToken
			}
			fmt.Println(ui.KeyValue("Ping Token", ui.Yellow(tokDisplay)))
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

Note: The raw ping token is displayed once upon creation so you can configure your endpoint.
Subsequent 'list' and 'show' commands mask the token by default (use 'beep beeper show <id> --show-token' to view unmasked).

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

		tz := flagBeeperTimezone
		if tz == "" {
			if detected, ok := workspace.DetectTimezoneOK(); ok {
				tz = detected
			} else {
				tz = "UTC"
			}
		}

		var b *client.Beeper
		if !flagNoInteractive && ui.IsInteractive() {
			if appSlug == "" || title == "" {
				b, err = runInteractiveBeeperCreate(ctx, c, appSlug, title, strings.TrimSpace(flagBeeperBody), cron, tz, configMap, flagBeeperChannels)
				if err != nil {
					return err
				}
			} else {
				if cron != "" {
					if err := schedule.Validate(cron); err != nil {
						return fmt.Errorf("invalid cron expression: %w", err)
					}
				}

				req := &client.CreateBeeperRequest{
					BeeperAppSlug: appSlug,
					Title:         title,
					Body:          strings.TrimSpace(flagBeeperBody),
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

				b, err = c.CreateBeeper(ctx, req)
				if err != nil {
					fmt.Println()
					fmt.Println(ui.Error("Creation failed: %s", err))
					fmt.Println(ui.Dim("Please review and adjust your inputs below:"))
					fmt.Println()
					b, err = runInteractiveBeeperCreate(ctx, c, appSlug, title, strings.TrimSpace(flagBeeperBody), cron, tz, configMap, flagBeeperChannels)
					if err != nil {
						return err
					}
				}
			}
		} else {
			if appSlug == "" {
				return errors.New("--app slug is required (view available apps with 'beep beeper apps')")
			}
			if title == "" {
				selectedApp, err := c.GetBeeperApp(ctx, appSlug)
				if err != nil {
					return fmt.Errorf("failed to fetch beeper app %s: %w", appSlug, err)
				}
				title = selectedApp.Name
				if cron == "" {
					cron = selectedApp.DefaultCron
				}
			}
			if cron != "" {
				if err := schedule.Validate(cron); err != nil {
					return fmt.Errorf("invalid cron expression: %w", err)
				}
			}

			req := &client.CreateBeeperRequest{
				BeeperAppSlug: appSlug,
				Title:         title,
				Body:          strings.TrimSpace(flagBeeperBody),
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

			b, err = c.CreateBeeper(ctx, req)
			if err != nil {
				return err
			}
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

func runInteractiveBeeperCreate(
	ctx context.Context,
	c *client.Client,
	initialAppSlug, initialTitle, initialBody, initialCron, initialTz string,
	initialConfig map[string]any,
	initialChannels string,
) (*client.Beeper, error) {
	appSlug := initialAppSlug
	title := initialTitle
	body := initialBody
	cron := initialCron
	tz := initialTz
	channels := initialChannels
	configMap := make(map[string]any)
	for k, v := range initialConfig {
		configMap[k] = v
	}

	apps, err := c.ListBeeperApps(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list beeper apps: %w", err)
	}
	if len(apps) == 0 {
		return nil, errors.New("no beeper apps available on server")
	}

	appOptions := make([]huh.Option[string], 0, len(apps))
	for _, a := range apps {
		label := fmt.Sprintf("%s (%s)", a.Name, a.Slug)
		appOptions = append(appOptions, huh.NewOption(label, a.Slug))
	}

	for {
		// 1. Select App Template
		err := huh.NewSelect[string]().
			Title("Select Probe App Template").
			Options(appOptions...).
			Value(&appSlug).
			Run()
		if err != nil {
			return nil, err
		}

		selectedApp, err := c.GetBeeperApp(ctx, appSlug)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch beeper app %s: %w", appSlug, err)
		}

		if title == "" {
			title = selectedApp.Name
		}
		if cron == "" {
			cron = selectedApp.DefaultCron
		}

		// 2. Build form fields
		basicGroup := huh.NewGroup(
			huh.NewInput().
				Title("Beeper Title").
				Description("Name of this monitor probe").
				Value(&title).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return errors.New("title is required")
					}
					return nil
				}),
			huh.NewInput().
				Title("Description / Body (optional)").
				Description("Optional markdown body or details").
				Value(&body),
			huh.NewInput().
				Title("Cron Schedule").
				Description(fmt.Sprintf("Schedule expression (default: %s)", selectedApp.DefaultCron)).
				Value(&cron).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return errors.New("cron expression is required")
					}
					return schedule.Validate(s)
				}),
		)

		var groups []*huh.Group
		groups = append(groups, basicGroup)

		// Dynamic config inputs
		configStringVals := make(map[string]*string)
		if len(selectedApp.Inputs) > 0 {
			configFields := make([]huh.Field, 0, len(selectedApp.Inputs))
			for _, input := range selectedApp.Inputs {
				val := ""
				if existing, ok := configMap[input.Name]; ok {
					val = fmt.Sprintf("%v", existing)
				} else if input.Default != nil {
					val = fmt.Sprintf("%v", input.Default)
				}
				configStringVals[input.Name] = &val

				inputName := input.Name
				desc := input.Description
				req := input.Required
				fieldTitle := fmt.Sprintf("Config: %s", inputName)
				if req {
					fieldTitle += " (required)"
				} else {
					fieldTitle += " (optional)"
				}

				configFields = append(configFields, huh.NewInput().
					Title(fieldTitle).
					Description(desc).
					Value(configStringVals[inputName]).
					Validate(func(s string) error {
						if req && strings.TrimSpace(s) == "" {
							return fmt.Errorf("%s is required", inputName)
						}
						return nil
					}),
				)
			}
			groups = append(groups, huh.NewGroup(configFields...))
		}

		// Settings group: Timezone and Channels
		settingsGroup := huh.NewGroup(
			huh.NewInput().
				Title("Timezone").
				Description("Timezone for probe schedule").
				Value(&tz),
			huh.NewInput().
				Title("Notification Channels (optional)").
				Description("Optional comma-separated channel names or IDs").
				Value(&channels),
		)
		groups = append(groups, settingsGroup)

		form := huh.NewForm(groups...)
		if err := form.Run(); err != nil {
			return nil, err
		}

		// Copy dynamic config values back to configMap
		for k, v := range configStringVals {
			if trimmed := strings.TrimSpace(*v); trimmed != "" {
				configMap[k] = trimmed
			} else {
				delete(configMap, k)
			}
		}

		req := &client.CreateBeeperRequest{
			BeeperAppSlug: appSlug,
			Title:         strings.TrimSpace(title),
			Body:          strings.TrimSpace(body),
			Cron:          strings.TrimSpace(cron),
			Timezone:      strings.TrimSpace(tz),
			Config:        configMap,
		}

		if trimmedCh := strings.TrimSpace(channels); trimmedCh != "" {
			for _, ch := range strings.Split(trimmedCh, ",") {
				if cName := strings.TrimSpace(ch); cName != "" {
					req.NotificationChannels = append(req.NotificationChannels, cName)
				}
			}
		}

		b, err := c.CreateBeeper(ctx, req)
		if err == nil {
			return b, nil
		}

		fmt.Println()
		fmt.Println(ui.Error("Creation failed: %s", err))
		fmt.Println(ui.Dim("Please review and adjust your inputs below:"))
		fmt.Println()
	}
}

func resolveBeeperID(ctx context.Context, c *client.Client, args []string, action string) (string, error) {
	if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
		return strings.TrimSpace(args[0]), nil
	}

	if flagNoInteractive || !ui.IsInteractive() {
		return "", fmt.Errorf("beeper ID is required (e.g. beep beeper %s <id>)", action)
	}

	beepers, err := c.ListBeepers(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list beepers: %w", err)
	}
	if len(beepers) == 0 {
		return "", errors.New("no beepers found in this account")
	}

	options := make([]huh.Option[string], 0, len(beepers))
	for _, b := range beepers {
		title := b.Title
		if len(title) > 30 {
			title = title[:27] + "..."
		}
		label := fmt.Sprintf("%-30s (%s - %s)", title, b.ID, b.Status)
		options = append(options, huh.NewOption(label, b.ID))
	}

	var selectedID string
	err = huh.NewSelect[string]().
		Title(fmt.Sprintf("Select beeper to %s", action)).
		Options(options...).
		Value(&selectedID).
		Run()
	if err != nil {
		return "", err
	}
	return selectedID, nil
}

var beeperDeleteCmd = &cobra.Command{
	Use:     "delete [id]",
	Aliases: []string{"rm"},
	Short:   "Delete a monitor beeper",
	Args:    cobra.MaximumNArgs(1),
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
		id, err := resolveBeeperID(ctx, c, args, "delete")
		if err != nil {
			return err
		}

		if err := c.DeleteBeeper(ctx, id); err != nil {
			return err
		}

		if flagJSON {
			fmt.Printf("{\"id\":%q,\"deleted\":true}\n", id)
			return nil
		}

		fmt.Println(ui.Success("Deleted beeper %s", ui.Bold(id)))
		return nil
	},
}

var beeperPauseCmd = &cobra.Command{
	Use:   "pause [id]",
	Short: "Pause a monitor beeper",
	Args:  cobra.MaximumNArgs(1),
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
		id, err := resolveBeeperID(ctx, c, args, "pause")
		if err != nil {
			return err
		}

		b, err := c.PauseBeeper(ctx, id)
		if err != nil {
			return err
		}

		if flagJSON {
			data, err := json.MarshalIndent(redactBeeper(b, flagBeeperShowToken), "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return nil
		}

		fmt.Println(ui.Success("Paused beeper %s (%s)", ui.Bold(b.Title), ui.Dim(b.ID)))
		return nil
	},
}

var beeperResumeCmd = &cobra.Command{
	Use:   "resume [id]",
	Short: "Resume a monitor beeper",
	Args:  cobra.MaximumNArgs(1),
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
		id, err := resolveBeeperID(ctx, c, args, "resume")
		if err != nil {
			return err
		}

		b, err := c.ResumeBeeper(ctx, id)
		if err != nil {
			return err
		}

		if flagJSON {
			data, err := json.MarshalIndent(redactBeeper(b, flagBeeperShowToken), "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return nil
		}

		fmt.Println(ui.Success("Resumed beeper %s (%s)", ui.Bold(b.Title), ui.Dim(b.ID)))
		return nil
	},
}

var beeperRunCmd = &cobra.Command{
	Use:   "run [id]",
	Short: "Immediately trigger a probe run for a beeper",
	Args:  cobra.MaximumNArgs(1),
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
		id, err := resolveBeeperID(ctx, c, args, "run")
		if err != nil {
			return err
		}

		run, err := c.RunBeeper(ctx, id)
		if err != nil {
			return err
		}

		if flagJSON {
			data, err := json.MarshalIndent(run, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return nil
		}

		fmt.Println(ui.Success("Triggered run for beeper %s (Run ID: %s, Status: %s)", ui.Bold(id), ui.Cyan(run.ID), formatRunStatus(run.Status)))
		return nil
	},
}

var beeperRunsCmd = &cobra.Command{
	Use:   "runs [id]",
	Short: "View probe execution history for a beeper",
	Args:  cobra.MaximumNArgs(1),
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
		id, err := resolveBeeperID(ctx, c, args, "view runs for")
		if err != nil {
			return err
		}

		runs, err := c.ListBeeperRuns(ctx, id)
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

		fmt.Printf("%s %s\n\n", ui.Bold(ui.Cyan("Beeper Runs")), ui.Dim(fmt.Sprintf("(beeper ID: %s)", id)))

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

	beeperShowCmd.Flags().BoolVar(&flagBeeperShowToken, "show-token", false, "Display unmasked ping token (defaults to masked for security)")

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
