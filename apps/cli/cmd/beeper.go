package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"beep/internal/client"
	"beep/internal/config"
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
		return runWithClient(func(ctx context.Context, cfg *config.Config, c *client.Client) error {
			beepers, err := c.ListBeepers(ctx)
			if err != nil {
				return err
			}

			if flagJSON {
				redacted := client.RedactBeepers(beepers, false)
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
			fmt.Printf("  %-10s  %-24s  %-12s  %-10s  %-10s  %-14s  %s\n",
				ui.Dim("ID"),
				ui.Dim("TITLE"),
				ui.Dim("APP"),
				ui.Dim("STATUS"),
				ui.Dim("ALERT"),
				ui.Dim("SCHEDULE"),
				ui.Dim("LAST PING/RUN"),
			)

			for _, b := range beepers {
				appSlug := "-"
				if b.BeeperApp != nil {
					appSlug = b.BeeperApp.Slug
				}
				statusStr := formatBeeperStatus(b.Status)
				alertStr := formatBeeperAlert(b.AlertState)
				scheduleStr := formatBeeperSchedule(b.Cron)
				lastRun := "-"
				if b.LastPingAt != "" {
					lastRun = b.LastPingAt
				} else if b.LastRunAt != "" {
					lastRun = b.LastRunAt
				}

				title := b.Title
				if len(title) > 24 {
					title = title[:21] + "..."
				}

				fmt.Printf("  %-10s  %-24s  %-12s  %-10s  %-10s  %-14s  %s\n",
					b.ID,
					title,
					appSlug,
					statusStr,
					alertStr,
					scheduleStr,
					lastRun,
				)
			}
			fmt.Println()
			return nil
		})
	},
}

var beeperShowCmd = &cobra.Command{
	Use:     "show [id]",
	Aliases: []string{"view", "info"},
	Short:   "Show details and health status of a monitor beeper",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWithClient(func(ctx context.Context, cfg *config.Config, c *client.Client) error {
			id, err := resolveBeeperID(ctx, c, args, "view")
			if err != nil {
				return err
			}

			b, err := c.GetBeeper(ctx, id)
			if err != nil {
				return err
			}

			if flagJSON {
				redacted := b.Redacted(flagBeeperShowToken)
				data, err := json.MarshalIndent(redacted, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(data))
				return nil
			}

			tokenDisplay := client.MaskToken(b.PingToken)
			if flagBeeperShowToken {
				tokenDisplay = b.PingToken
			}

			fmt.Println()
			fmt.Printf("  %s %s\n", ui.Bold("Beeper:"), ui.Cyan(b.Title))
			fmt.Println(ui.KeyValue("ID", b.ID))
			if b.BeeperApp != nil {
				fmt.Println(ui.KeyValue("App", fmt.Sprintf("%s (%s)", b.BeeperApp.Name, b.BeeperApp.Slug)))
			}
			fmt.Println(ui.KeyValue("Status", formatBeeperStatus(b.Status)))
			fmt.Println(ui.KeyValue("Alert State", formatBeeperAlert(b.AlertState)))
			if b.ConsecutiveFailures > 0 {
				fmt.Println(ui.KeyValue("Failures", ui.Red(fmt.Sprintf("%d consecutive", b.ConsecutiveFailures))))
			}
			if b.Cron != "" {
				fmt.Println(ui.KeyValue("Cron", b.Cron))
			}
			if b.Timezone != "" {
				fmt.Println(ui.KeyValue("Timezone", b.Timezone))
			}
			if tokenDisplay != "" {
				fmt.Println(ui.KeyValue("Ping Token", ui.Yellow(tokenDisplay)))
				if flagBeeperShowToken {
					fmt.Println(ui.KeyValue("Ping URL", fmt.Sprintf("%s/api/v1/beeper_apps/heartbeat/pings/%s", cfg.ServerURL, b.PingToken)))
				} else {
					fmt.Println(ui.KeyValue("Ping URL", fmt.Sprintf("%s/api/v1/beeper_apps/heartbeat/pings/%s", cfg.ServerURL, client.MaskToken(b.PingToken))))
				}
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
			if b.Body != "" {
				fmt.Println()
				fmt.Println(ui.Dim("  Description:"))
				for _, line := range strings.Split(b.Body, "\n") {
					fmt.Printf("    %s\n", line)
				}
			}

			if b.RunStats != nil {
				fmt.Println()
				fmt.Println(ui.Section("  Run Statistics:"))
				fmt.Printf("    Total runs: %d | Succeeded: %d\n", b.RunStats.Total, b.RunStats.Succeeded)
			}

			if len(b.Runs) > 0 {
				fmt.Println()
				fmt.Println(ui.Section("  Recent Runs:"))
				fmt.Printf("    %-10s  %-24s  %-10s  %s\n",
					ui.Dim("RUN ID"),
					ui.Dim("SCHEDULED FOR"),
					ui.Dim("STATUS"),
					ui.Dim("SIGNAL STATUS"),
				)
				for _, r := range b.Runs {
					sigStatus := "-"
					if r.SignalStatus != "" {
						sigStatus = r.SignalStatus
					}
					fmt.Printf("    %-10s  %-24s  %-10s  %s\n",
						r.ID,
						r.ScheduledFor,
						formatRunStatus(r.Status),
						sigStatus,
					)
				}
			}
			fmt.Println()
			return nil
		})
	},
}

var beeperAppsCmd = &cobra.Command{
	Use:     "apps",
	Aliases: []string{"templates", "catalog"},
	Short:   "List available beeper probe app templates",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWithClient(func(ctx context.Context, cfg *config.Config, c *client.Client) error {
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

			fmt.Printf("%s\n\n", ui.Bold(ui.Cyan("Available Beeper App Templates")))

			if len(apps) == 0 {
				fmt.Println(ui.Dim("  No apps available on server."))
				return nil
			}

			fmt.Printf("  %-16s  %-24s  %-16s  %s\n",
				ui.Dim("SLUG"),
				ui.Dim("NAME"),
				ui.Dim("DEFAULT CRON"),
				ui.Dim("DESCRIPTION"),
			)

			for _, a := range apps {
				name := a.Name
				if len(name) > 24 {
					name = name[:21] + "..."
				}
				desc := a.Description
				if len(desc) > 40 {
					desc = desc[:37] + "..."
				}
				fmt.Printf("  %-16s  %-24s  %-16s  %s\n",
					a.Slug,
					name,
					a.DefaultCron,
					ui.Dim(desc),
				)
			}
			fmt.Println()
			return nil
		})
	},
}

var beeperCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a monitor probe beeper",
	Long: `Create a monitor probe beeper.

Interactive mode guides you through selecting an app template and configuring its options.
In non-interactive mode or when flags are provided, options are read from flags.

Examples:
  # Interactive creation wizard
  beep beeper create

  # Heartbeat probe
  beep beeper create --app heartbeat --title "Production API Heartbeat" --cron "*/5 * * * *"

  # HTTP check probe
  beep beeper create --app http --title "Website Health" --cron "*/5 * * * *" --config url=https://example.com`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWithClient(func(ctx context.Context, cfg *config.Config, c *client.Client) error {
			appSlug := strings.TrimSpace(flagBeeperApp)
			title := strings.TrimSpace(flagBeeperTitle)
			cron := strings.TrimSpace(flagBeeperCron)
			configMap := make(map[string]any)

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

			params := client.CreateBeeperParams{
				AppSlug:  appSlug,
				Title:    title,
				Body:     strings.TrimSpace(flagBeeperBody),
				Cron:     cron,
				Timezone: tz,
				Config:   configMap,
				Channels: flagBeeperChannels,
			}

			var b *client.Beeper
			if !flagNoInteractive && ui.IsInteractive() {
				apps, err := c.ListBeeperApps(ctx)
				if err != nil {
					return fmt.Errorf("failed to list beeper apps: %w", err)
				}

				// Default channel selection comes from account settings.
				defaultChannels := client.DefaultNotificationChannels
				if s, err := c.GetSettings(ctx); err == nil && s != nil {
					defaultChannels = client.SanitizeChannelDefaults(s.NotificationChannels)
				}

				if params.AppSlug == "" || params.Title == "" {
					prompted, err := ui.PromptBeeperCreate(params, apps, defaultChannels)
					if err != nil {
						return err
					}
					params = *prompted
				}

				for {
					req, err := params.ToRequest()
					if err != nil {
						ui.PrintErrorList("Invalid input", client.ExtractErrorList(err))
						retry, promptErr := ui.PromptConfirm("Would you like to adjust your inputs?", true)
						if promptErr != nil || !retry {
							return err
						}
						prompted, pErr := ui.PromptBeeperAdjust(params, apps, defaultChannels)
						if pErr != nil {
							return pErr
						}
						params = *prompted
						continue
					}

					b, err = c.CreateBeeper(ctx, req)
					if err == nil {
						break
					}

					ui.PrintErrorList("Creation failed", client.ExtractErrorList(err))
					retry, promptErr := ui.PromptConfirm("Would you like to adjust your inputs and retry?", true)
					if promptErr != nil || !retry {
						return err
					}
					prompted, pErr := ui.PromptBeeperAdjust(params, apps, defaultChannels)
					if pErr != nil {
						return pErr
					}
					params = *prompted
				}
			} else {
				if params.AppSlug == "" {
					return errors.New("--app slug is required (view available apps with 'beep beeper apps')")
				}
				if params.Title == "" {
					selectedApp, err := c.GetBeeperApp(ctx, params.AppSlug)
					if err != nil {
						return fmt.Errorf("failed to fetch beeper app %s: %w", params.AppSlug, err)
					}
					params.Title = selectedApp.Name
					if params.Cron == "" {
						params.Cron = selectedApp.DefaultCron
					}
				}

				req, err := params.ToRequest()
				if err != nil {
					return err
				}

				var errCreate error
				b, errCreate = c.CreateBeeper(ctx, req)
				if errCreate != nil {
					return errCreate
				}
			}

			if flagJSON {
				redacted := b.Redacted(false)
				data, err := json.MarshalIndent(redacted, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(data))
				return nil
			}

			fmt.Println(ui.Success("Created beeper %s (%s)", ui.Bold(b.Title), ui.Dim(b.ID)))
			fmt.Printf("  %s %s\n", ui.Dim("App:"), params.AppSlug)
			if b.Cron != "" {
				fmt.Printf("  %s %s\n", ui.Dim("Cron:"), b.Cron)
			}
			if b.PingToken != "" {
				fmt.Printf("  %s %s\n", ui.Dim("Ping Token:"), ui.Yellow(b.PingToken))
				fmt.Printf("  %s %s/api/v1/beeper_apps/heartbeat/pings/%s\n",
					ui.Dim("Ping URL:"),
					cfg.ServerURL,
					b.PingToken,
				)
			}
			return nil
		})
	},
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

	return ui.PromptSelectResource(fmt.Sprintf("Select beeper to %s", action), options)
}

var beeperDeleteCmd = &cobra.Command{
	Use:     "delete [id]",
	Aliases: []string{"rm"},
	Short:   "Delete a monitor beeper",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWithClient(func(ctx context.Context, cfg *config.Config, c *client.Client) error {
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
		})
	},
}

var beeperPauseCmd = &cobra.Command{
	Use:   "pause [id]",
	Short: "Pause a monitor beeper",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWithClient(func(ctx context.Context, cfg *config.Config, c *client.Client) error {
			id, err := resolveBeeperID(ctx, c, args, "pause")
			if err != nil {
				return err
			}

			b, err := c.PauseBeeper(ctx, id)
			if err != nil {
				return err
			}

			if flagJSON {
				redacted := b.Redacted(false)
				data, err := json.MarshalIndent(redacted, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(data))
				return nil
			}

			fmt.Println(ui.Success("Paused beeper %s (%s)", ui.Bold(b.Title), ui.Dim(b.ID)))
			return nil
		})
	},
}

var beeperResumeCmd = &cobra.Command{
	Use:   "resume [id]",
	Short: "Resume a paused monitor beeper",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWithClient(func(ctx context.Context, cfg *config.Config, c *client.Client) error {
			id, err := resolveBeeperID(ctx, c, args, "resume")
			if err != nil {
				return err
			}

			b, err := c.ResumeBeeper(ctx, id)
			if err != nil {
				return err
			}

			if flagJSON {
				redacted := b.Redacted(false)
				data, err := json.MarshalIndent(redacted, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(data))
				return nil
			}

			fmt.Println(ui.Success("Resumed beeper %s (%s)", ui.Bold(b.Title), ui.Dim(b.ID)))
			return nil
		})
	},
}

var beeperRunCmd = &cobra.Command{
	Use:     "run [id]",
	Aliases: []string{"trigger"},
	Short:   "Immediately trigger a monitor beeper run",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWithClient(func(ctx context.Context, cfg *config.Config, c *client.Client) error {
			id, err := resolveBeeperID(ctx, c, args, "trigger")
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

			fmt.Println(ui.Success("Triggered beeper %s (run id: %s, status: %s)",
				ui.Bold(id),
				ui.Dim(run.ID),
				formatRunStatus(run.Status),
			))
			return nil
		})
	},
}

var beeperRunsCmd = &cobra.Command{
	Use:     "runs [id]",
	Aliases: []string{"history", "logs"},
	Short:   "View recent execution runs for a monitor beeper",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWithClient(func(ctx context.Context, cfg *config.Config, c *client.Client) error {
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

			fmt.Printf("%s %s\n\n", ui.Bold(ui.Cyan("Beeper Runs")), ui.Dim(fmt.Sprintf("(beeper: %s)", id)))

			if len(runs) == 0 {
				fmt.Println(ui.Dim("  No runs found for this beeper."))
				return nil
			}

			fmt.Printf("  %-10s  %-24s  %-10s  %-14s  %s\n",
				ui.Dim("ID"),
				ui.Dim("SCHEDULED FOR"),
				ui.Dim("STATUS"),
				ui.Dim("SIGNAL STATUS"),
				ui.Dim("CREATED AT"),
			)

			for _, r := range runs {
				sigStatus := "-"
				if r.SignalStatus != "" {
					sigStatus = r.SignalStatus
				}
				fmt.Printf("  %-10s  %-24s  %-10s  %-14s  %s\n",
					r.ID,
					r.ScheduledFor,
					formatRunStatus(r.Status),
					sigStatus,
					r.CreatedAt,
				)
			}
			fmt.Println()
			return nil
		})
	},
}

func formatBeeperStatus(s string) string {
	switch strings.ToLower(s) {
	case "active":
		return ui.Green("active")
	case "paused":
		return ui.Yellow("paused")
	case "disabled":
		return ui.Dim("disabled")
	default:
		return s
	}
}

func formatBeeperAlert(s string) string {
	switch strings.ToLower(s) {
	case "ok", "clear", "healthy":
		return ui.Green(s)
	case "firing", "alerting", "critical":
		return ui.Red(s)
	case "pending":
		return ui.Yellow(s)
	default:
		return ui.Dim(s)
	}
}

func formatBeeperSchedule(cron string) string {
	if cron == "" {
		return ui.Dim("webhook only")
	}
	return cron
}

func init() {
	beeperCreateCmd.Flags().StringVar(&flagBeeperApp, "app", "", "Beeper app slug (e.g. 'heartbeat-ping')")
	beeperCreateCmd.Flags().StringVar(&flagBeeperTitle, "title", "", "Beeper probe title")
	beeperCreateCmd.Flags().StringVarP(&flagBeeperBody, "body", "b", "", "Description / body for this monitor")
	beeperCreateCmd.Flags().StringVarP(&flagBeeperCron, "cron", "c", "", "Recurring cron schedule (e.g. '*/5 * * * *')")
	beeperCreateCmd.Flags().StringVarP(&flagBeeperTimezone, "timezone", "z", "", "Timezone (defaults to local timezone)")
	beeperCreateCmd.Flags().StringVar(&flagBeeperChannels, "channels", "", "Comma-separated notification channel names or IDs")
	beeperCreateCmd.Flags().StringSliceVar(&flagBeeperConfigs, "config", nil, "App config key=value pair (can be specified multiple times)")

	beeperShowCmd.Flags().BoolVar(&flagBeeperShowToken, "show-token", false, "Display the full unredacted ping token")

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
