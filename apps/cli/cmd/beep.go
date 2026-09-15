package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"beep/internal/client"
	"beep/internal/schedule"
	"beep/internal/ui"
	"beep/internal/workspace"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var (
	flagBeepBody     string
	flagBeepIn       string
	flagBeepAt       string
	flagBeepCron     string
	flagBeepTimezone string
	flagBeepChannels string
	flagBeepNatural  string
)

var beepCmd = &cobra.Command{
	Use:   "beep",
	Short: "Manage reminder and notification beeps",
	Long:  ui.Bold(ui.Cyan("Beep Management")) + ` - Create, list, trigger, pause, and delete reminder beeps.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var beepListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List reminder beeps",
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
		beeps, err := c.ListBeeps(ctx)
		if err != nil {
			return err
		}

		if flagJSON {
			data, err := json.MarshalIndent(beeps, "", "  ")
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
		fmt.Printf("%s %s\n\n", ui.Bold(ui.Cyan("Beeps")), ui.Dim(fmt.Sprintf("(account: %s)", accountDisplay)))

		if len(beeps) == 0 {
			fmt.Println(ui.Dim("  No beeps found. Create one with 'beep beep create'."))
			return nil
		}

		// Print table header
		fmt.Printf("  %-10s  %-24s  %-10s  %-9s  %-22s  %s\n",
			ui.Dim("ID"),
			ui.Dim("TITLE"),
			ui.Dim("STATUS"),
			ui.Dim("KIND"),
			ui.Dim("SCHEDULE"),
			ui.Dim("CHANNELS"),
		)

		for _, b := range beeps {
			statusStr := formatBeepStatus(b.Status)
			scheduleStr := formatBeepSchedule(b)
			title := b.Title
			if len(title) > 24 {
				title = title[:21] + "..."
			}
			channels := strings.Join(b.NotificationChannels, ", ")
			if channels == "" {
				channels = ui.Dim("default")
			}

			fmt.Printf("  %-10s  %-24s  %-10s  %-9s  %-22s  %s\n",
				b.ID,
				title,
				statusStr,
				b.Kind,
				scheduleStr,
				channels,
			)
		}
		fmt.Println()
		return nil
	},
}

var beepShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show details of a reminder beep",
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
		b, err := c.GetBeep(ctx, args[0])
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
		fmt.Printf("  %s %s\n", ui.Bold("Beep:"), ui.Cyan(b.Title))
		fmt.Println(ui.KeyValue("ID", b.ID))
		fmt.Println(ui.KeyValue("Status", formatBeepStatus(b.Status)))
		fmt.Println(ui.KeyValue("Kind", b.Kind))
		if b.Cron != "" {
			fmt.Println(ui.KeyValue("Cron", b.Cron))
		}
		if b.RunAt != "" {
			fmt.Println(ui.KeyValue("Run At", b.RunAt))
		}
		if b.NextRunAt != "" {
			fmt.Println(ui.KeyValue("Next Run", b.NextRunAt))
		}
		if b.LastRunAt != "" {
			fmt.Println(ui.KeyValue("Last Run", b.LastRunAt))
		}
		if b.Timezone != "" {
			fmt.Println(ui.KeyValue("Timezone", b.Timezone))
		}
		if len(b.NotificationChannels) > 0 {
			fmt.Println(ui.KeyValue("Channels", strings.Join(b.NotificationChannels, ", ")))
		}
		if b.Body != "" {
			fmt.Println()
			fmt.Println(ui.Dim("  Message Body:"))
			for _, line := range strings.Split(b.Body, "\n") {
				fmt.Printf("    %s\n", line)
			}
		}

		if len(b.Runs) > 0 {
			fmt.Println()
			fmt.Println(ui.Section("  Recent Runs:"))
			fmt.Printf("    %-10s  %-24s  %-10s\n",
				ui.Dim("RUN ID"),
				ui.Dim("SCHEDULED FOR"),
				ui.Dim("STATUS"),
			)
			for _, r := range b.Runs {
				fmt.Printf("    %-10s  %-24s  %-10s\n",
					r.ID,
					r.ScheduledFor,
					formatRunStatus(r.Status),
				)
			}
		}
		fmt.Println()
		return nil
	},
}

var beepCreateCmd = &cobra.Command{
	Use:   "create [title]",
	Short: "Create a reminder beep",
	Long: `Create a reminder beep.

Scheduling modes (mutually exclusive):
  - Instant (default): fires immediately when no schedule flags are provided
  - Relative delay:    --in (e.g. 15m, 2h, 1d)
  - Specific datetime: --at (e.g. 16:30, "2026-10-01 10:00")
  - Recurring cron:    --cron (e.g. "0 10 * * 1-5")
  - AI natural:        --natural / -n (e.g. "remind me in 30 minutes to drink water")

Note: --cron, --in, --at, and --natural are mutually exclusive.
When using --json, interactive confirmation is skipped.
Optional --body and --channels can also be combined with --natural to supplement proposal fields.

Examples:
  # Instant reminder (fires immediately)
  beep beep create "Deploy finished"

  # Relative delay
  beep beep create "Meeting starts" --in 15m
  beep beep create "Check logs" --in 2h

  # Specific datetime
  beep beep create "Doctor appointment" --at "16:30"
  beep beep create "Release v1.0" --at "2026-10-01 10:00"

  # Recurring cron
  beep beep create "Daily Standup" --cron "0 10 * * 1-5"

  # Natural language via DeepSeek AI
  beep beep create -n "remind me in 30 minutes to drink water"`,
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

		tz := flagBeepTimezone
		if tz == "" {
			if detected, ok := workspace.DetectTimezoneOK(); ok {
				tz = detected
			} else {
				tz = "UTC"
			}
		}

		// 1. Natural language creation
		if flagBeepNatural != "" {
			return handleNaturalBeepCreate(ctx, c, flagBeepNatural, tz)
		}

		var title string
		if len(args) > 0 {
			title = strings.TrimSpace(args[0])
		}

		// Interactive fallback for missing title
		if title == "" {
			if !flagNoInteractive && ui.IsInteractive() {
				err := huh.NewInput().
					Title("Reminder Title").
					Description("Short description of what you want to be reminded about").
					Placeholder("e.g. Call dentist, Check build status").
					Value(&title).
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
				title = strings.TrimSpace(title)
			} else {
				return fmt.Errorf("reminder title is required (e.g. beep beep create \"Meeting in 10m\" --in 10m)")
			}
		}

		req := &client.CreateBeepRequest{
			Title:    title,
			Body:     flagBeepBody,
			Timezone: tz,
		}

		if flagBeepChannels != "" {
			for _, ch := range strings.Split(flagBeepChannels, ",") {
				if trimmed := strings.TrimSpace(ch); trimmed != "" {
					req.NotificationChannels = append(req.NotificationChannels, trimmed)
				}
			}
		}

		// Determine schedule
		if flagBeepCron != "" {
			if err := schedule.Validate(flagBeepCron); err != nil {
				return fmt.Errorf("invalid --cron expression: %w", err)
			}
			req.Kind = "recurring"
			req.Cron = flagBeepCron
		} else if flagBeepIn != "" {
			d, err := parseInDuration(flagBeepIn)
			if err != nil {
				return fmt.Errorf("invalid --in duration (e.g. 10m, 2h, 1d): %w", err)
			}
			req.Kind = "once"
			req.RunAt = time.Now().Add(d).Format(time.RFC3339)
		} else if flagBeepAt != "" {
			loc, err := time.LoadLocation(tz)
			if err != nil {
				loc = time.Local
			}
			t, err := parseAtTime(flagBeepAt, loc)
			if err != nil {
				return fmt.Errorf("invalid --at time (e.g. 15:30, 2026-10-01 10:00): %w", err)
			}
			req.Kind = "once"
			req.RunAt = t.Format(time.RFC3339)
		} else {
			// Instant reminder
			req.Kind = "once"
			req.RunAt = time.Now().Format(time.RFC3339)
		}

		b, err := c.CreateBeep(ctx, req)
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

		scheduleInfo := formatBeepSchedule(b)
		fmt.Println(ui.Success("Created beep %s (%s)", ui.Bold(b.Title), ui.Dim(b.ID)))
		fmt.Printf("  %s %s\n", ui.Dim("Schedule:"), scheduleInfo)
		if b.Timezone != "" {
			fmt.Printf("  %s %s\n", ui.Dim("Timezone:"), b.Timezone)
		}
		return nil
	},
}

func handleNaturalBeepCreate(ctx context.Context, c *client.Client, prompt, tz string) error {
	proposal, err := c.ProposeBeep(ctx, prompt, tz)
	if err != nil {
		return fmt.Errorf("natural language parse failed: %w", err)
	}

	if len(proposal.Errors) > 0 {
		return fmt.Errorf("could not understand reminder: %s", strings.Join(proposal.Errors, ", "))
	}

	// In interactive mode, confirm proposal unless --json or --no-interactive is passed.
	if !flagNoInteractive && !flagJSON && ui.IsInteractive() {
		fmt.Println()
		fmt.Printf("  %s %s\n", ui.Bold("AI Proposal:"), ui.Cyan(proposal.Title))
		if proposal.Body != "" {
			fmt.Printf("  %s %s\n", ui.Dim("Body:"), proposal.Body)
		}
		fmt.Printf("  %s %s\n", ui.Dim("Kind:"), proposal.Kind)
		if proposal.Cron != "" {
			fmt.Printf("  %s %s\n", ui.Dim("Cron:"), proposal.Cron)
		}
		if proposal.RunAt != "" {
			fmt.Printf("  %s %s\n", ui.Dim("Run At:"), proposal.RunAt)
		}
		fmt.Println()

		var confirmed bool = true
		confirmErr := huh.NewConfirm().
			Title("Create this reminder?").
			Value(&confirmed).
			Run()
		if confirmErr != nil {
			return confirmErr
		}
		if !confirmed {
			fmt.Println(ui.Dim("Cancelled."))
			return nil
		}
	}

	body := proposal.Body
	if flagBeepBody != "" {
		body = flagBeepBody
	}

	req := &client.CreateBeepRequest{
		Title:    proposal.Title,
		Body:     body,
		Kind:     proposal.Kind,
		Cron:     proposal.Cron,
		RunAt:    proposal.RunAt,
		Timezone: proposal.Timezone,
	}

	if flagBeepChannels != "" {
		for _, ch := range strings.Split(flagBeepChannels, ",") {
			if trimmed := strings.TrimSpace(ch); trimmed != "" {
				req.NotificationChannels = append(req.NotificationChannels, trimmed)
			}
		}
	}

	b, err := c.CreateBeep(ctx, req)
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

	fmt.Println(ui.Success("Created beep %s (%s)", ui.Bold(b.Title), ui.Dim(b.ID)))
	fmt.Printf("  %s %s\n", ui.Dim("Schedule:"), formatBeepSchedule(b))
	return nil
}

var beepDeleteCmd = &cobra.Command{
	Use:     "delete <id>",
	Aliases: []string{"rm"},
	Short:   "Delete a reminder beep",
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
		if err := c.DeleteBeep(ctx, args[0]); err != nil {
			return err
		}

		if flagJSON {
			fmt.Printf("{\"id\":%q,\"deleted\":true}\n", args[0])
			return nil
		}

		fmt.Println(ui.Success("Deleted beep %s", ui.Bold(args[0])))
		return nil
	},
}

var beepPauseCmd = &cobra.Command{
	Use:   "pause <id>",
	Short: "Pause a recurring reminder beep",
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
		b, err := c.PauseBeep(ctx, args[0])
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

		fmt.Println(ui.Success("Paused beep %s (%s)", ui.Bold(b.Title), ui.Dim(b.ID)))
		return nil
	},
}

var beepResumeCmd = &cobra.Command{
	Use:   "resume <id>",
	Short: "Resume a paused reminder beep",
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
		b, err := c.ResumeBeep(ctx, args[0])
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

		fmt.Println(ui.Success("Resumed beep %s (%s)", ui.Bold(b.Title), ui.Dim(b.ID)))
		return nil
	},
}

var beepRunCmd = &cobra.Command{
	Use:     "run <id>",
	Aliases: []string{"trigger"},
	Short:   "Immediately trigger a reminder beep",
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
		run, err := c.RunBeep(ctx, args[0])
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

		fmt.Println(ui.Success("Triggered beep %s (Run ID: %s, Status: %s)", ui.Bold(args[0]), ui.Cyan(run.ID), formatRunStatus(run.Status)))
		return nil
	},
}

func formatBeepStatus(s string) string {
	switch strings.ToLower(s) {
	case "active":
		return ui.Green("active")
	case "paused":
		return ui.Yellow("paused")
	case "firing":
		return ui.Red("firing")
	case "completed":
		return ui.Dim("completed")
	case "cancelled":
		return ui.Dim("cancelled")
	default:
		return s
	}
}

func formatRunStatus(s string) string {
	switch strings.ToLower(s) {
	case "succeeded", "ok":
		return ui.Green(s)
	case "failed", "error":
		return ui.Red(s)
	case "firing", "running":
		return ui.Yellow(s)
	default:
		return s
	}
}

func formatBeepSchedule(b *client.Beep) string {
	if b.Kind == "recurring" && b.Cron != "" {
		return "cron: " + b.Cron
	}
	if b.NextRunAt != "" {
		return b.NextRunAt
	}
	if b.RunAt != "" {
		return b.RunAt
	}
	return ui.Dim("instant")
}

func parseInDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if strings.HasSuffix(s, "d") {
		daysStr := strings.TrimSuffix(s, "d")
		days, err := strconv.Atoi(daysStr)
		if err == nil {
			return time.Duration(days) * 24 * time.Hour, nil
		}
	}
	return time.ParseDuration(s)
}

func parseAtTime(s string, loc *time.Location) (time.Time, error) {
	s = strings.TrimSpace(s)
	formats := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
	}
	for _, layout := range formats {
		if t, err := time.ParseInLocation(layout, s, loc); err == nil {
			return t, nil
		}
	}

	// Try time-only "15:04"
	if t, err := time.ParseInLocation("15:04", s, loc); err == nil {
		now := time.Now().In(loc)
		target := time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, loc)
		if target.Before(now) {
			target = target.AddDate(0, 0, 1)
		}
		return target, nil
	}

	return time.Time{}, fmt.Errorf("unrecognized time format %q", s)
}

func init() {
	beepCreateCmd.Flags().StringVarP(&flagBeepBody, "body", "b", "", "Reminder markdown body / message")
	beepCreateCmd.Flags().StringVar(&flagBeepIn, "in", "", "Delay duration before firing (e.g. 15m, 2h, 1d)")
	beepCreateCmd.Flags().StringVar(&flagBeepAt, "at", "", "Specific time to fire (e.g. 16:30, 2026-10-01 10:00)")
	beepCreateCmd.Flags().StringVarP(&flagBeepCron, "cron", "c", "", "Recurring cron schedule (e.g. '0 9 * * *')")
	beepCreateCmd.Flags().StringVarP(&flagBeepTimezone, "timezone", "z", "", "Timezone (defaults to local timezone)")
	beepCreateCmd.Flags().StringVar(&flagBeepChannels, "channels", "", "Comma-separated notification channel names or IDs")
	beepCreateCmd.Flags().StringVarP(&flagBeepNatural, "natural", "n", "", "Natural language reminder prompt parsed by AI")
	beepCreateCmd.MarkFlagsMutuallyExclusive("cron", "in", "at", "natural")

	beepCmd.AddCommand(beepListCmd)
	beepCmd.AddCommand(beepShowCmd)
	beepCmd.AddCommand(beepCreateCmd)
	beepCmd.AddCommand(beepDeleteCmd)
	beepCmd.AddCommand(beepPauseCmd)
	beepCmd.AddCommand(beepResumeCmd)
	beepCmd.AddCommand(beepRunCmd)
}
