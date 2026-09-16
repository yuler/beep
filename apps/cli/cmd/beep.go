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
		return runWithClient(func(ctx context.Context, cfg *config.Config, c *client.Client) error {
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

				chans := "-"
				if len(b.NotificationChannels) > 0 {
					chans = strings.Join(b.NotificationChannels, ", ")
				}

				fmt.Printf("  %-10s  %-24s  %-10s  %-9s  %-22s  %s\n",
					b.ID,
					title,
					statusStr,
					b.Kind,
					scheduleStr,
					chans,
				)
			}
			fmt.Println()
			return nil
		})
	},
}

var beepShowCmd = &cobra.Command{
	Use:     "show [id]",
	Aliases: []string{"view", "info"},
	Short:   "Show details and recent run history of a reminder beep",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWithClient(func(ctx context.Context, cfg *config.Config, c *client.Client) error {
			id, err := resolveBeepID(ctx, c, args, "view")
			if err != nil {
				return err
			}

			b, err := c.GetBeep(ctx, id)
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
		})
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
		return runWithClient(func(ctx context.Context, cfg *config.Config, c *client.Client) error {
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

			schedKind := ""
			schedVal := ""
			if flagBeepCron != "" {
				schedKind = "cron"
				schedVal = flagBeepCron
			} else if flagBeepIn != "" {
				schedKind = "delay"
				schedVal = flagBeepIn
			} else if flagBeepAt != "" {
				schedKind = "at"
				schedVal = flagBeepAt
			}

			params := client.CreateBeepParams{
				Title:        title,
				Body:         flagBeepBody,
				ScheduleKind: schedKind,
				ScheduleVal:  schedVal,
				Timezone:     tz,
				Channels:     flagBeepChannels,
			}

			var b *client.Beep
			if !flagNoInteractive && ui.IsInteractive() {
				// If required info is missing, prompt sequentially
				isMissingInfo := params.Title == "" || (params.ScheduleKind == "" && len(args) == 0)
				if isMissingInfo {
					prompted, err := ui.PromptBeepCreate(params)
					if err != nil {
						return err
					}
					params = *prompted
				}

				for {
					if params.ScheduleKind == "" {
						params.ScheduleKind = "instant"
					}
					req, err := params.ToRequest()
					if err != nil {
						fmt.Println()
						fmt.Println(ui.Error("Invalid input: %s", err))
						retry, promptErr := ui.PromptConfirm("Would you like to adjust your inputs?", true)
						if promptErr != nil || !retry {
							return err
						}
						prompted, pErr := ui.PromptBeepCreate(params)
						if pErr != nil {
							return pErr
						}
						params = *prompted
						continue
					}

					b, err = c.CreateBeep(ctx, req)
					if err == nil {
						break
					}

					fmt.Println()
					fmt.Println(ui.Error("Creation failed: %s", err))
					retry, promptErr := ui.PromptConfirm("Would you like to adjust your inputs and retry?", true)
					if promptErr != nil || !retry {
						return err
					}
					prompted, pErr := ui.PromptBeepCreate(params)
					if pErr != nil {
						return pErr
					}
					params = *prompted
				}
			} else {
				if params.Title == "" {
					return fmt.Errorf("reminder title is required (e.g. beep beep create \"Meeting in 10m\" --in 10m)")
				}
				if params.ScheduleKind == "" {
					params.ScheduleKind = "instant"
				}
				req, err := params.ToRequest()
				if err != nil {
					return err
				}
				var errCreate error
				b, errCreate = c.CreateBeep(ctx, req)
				if errCreate != nil {
					return errCreate
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

			scheduleInfo := formatBeepSchedule(b)
			fmt.Println(ui.Success("Created beep %s (%s)", ui.Bold(b.Title), ui.Dim(b.ID)))
			fmt.Printf("  %s %s\n", ui.Dim("Schedule:"), scheduleInfo)
			if b.Timezone != "" {
				fmt.Printf("  %s %s\n", ui.Dim("Timezone:"), b.Timezone)
			}
			return nil
		})
	},
}

func resolveBeepID(ctx context.Context, c *client.Client, args []string, action string) (string, error) {
	if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
		return strings.TrimSpace(args[0]), nil
	}

	if flagNoInteractive || !ui.IsInteractive() {
		return "", fmt.Errorf("beep ID is required (e.g. beep beep %s <id>)", action)
	}

	beeps, err := c.ListBeeps(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list beeps: %w", err)
	}
	if len(beeps) == 0 {
		return "", errors.New("no beeps found in this account")
	}

	options := make([]huh.Option[string], 0, len(beeps))
	for _, b := range beeps {
		title := b.Title
		if len(title) > 30 {
			title = title[:27] + "..."
		}
		label := fmt.Sprintf("%-30s (%s - %s)", title, b.ID, b.Status)
		options = append(options, huh.NewOption(label, b.ID))
	}

	return ui.PromptSelectResource(fmt.Sprintf("Select beep to %s", action), options)
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

		confirmed, confirmErr := ui.PromptConfirm("Create this reminder?", true)
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
	Use:     "delete [id]",
	Aliases: []string{"rm"},
	Short:   "Delete a reminder beep",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWithClient(func(ctx context.Context, cfg *config.Config, c *client.Client) error {
			id, err := resolveBeepID(ctx, c, args, "delete")
			if err != nil {
				return err
			}

			if err := c.DeleteBeep(ctx, id); err != nil {
				return err
			}

			if flagJSON {
				fmt.Printf("{\"id\":%q,\"deleted\":true}\n", id)
				return nil
			}

			fmt.Println(ui.Success("Deleted beep %s", ui.Bold(id)))
			return nil
		})
	},
}

var beepPauseCmd = &cobra.Command{
	Use:   "pause [id]",
	Short: "Pause a recurring reminder beep",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWithClient(func(ctx context.Context, cfg *config.Config, c *client.Client) error {
			id, err := resolveBeepID(ctx, c, args, "pause")
			if err != nil {
				return err
			}

			b, err := c.PauseBeep(ctx, id)
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
		})
	},
}

var beepResumeCmd = &cobra.Command{
	Use:   "resume [id]",
	Short: "Resume a paused reminder beep",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWithClient(func(ctx context.Context, cfg *config.Config, c *client.Client) error {
			id, err := resolveBeepID(ctx, c, args, "resume")
			if err != nil {
				return err
			}

			b, err := c.ResumeBeep(ctx, id)
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
		})
	},
}

var beepRunCmd = &cobra.Command{
	Use:     "run [id]",
	Aliases: []string{"trigger"},
	Short:   "Immediately trigger a reminder beep",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWithClient(func(ctx context.Context, cfg *config.Config, c *client.Client) error {
			id, err := resolveBeepID(ctx, c, args, "trigger")
			if err != nil {
				return err
			}

			run, err := c.RunBeep(ctx, id)
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

			fmt.Println(ui.Success("Triggered beep %s (run id: %s, status: %s)",
				ui.Bold(id),
				ui.Dim(run.ID),
				formatRunStatus(run.Status),
			))
			return nil
		})
	},
}

func formatBeepStatus(s string) string {
	switch strings.ToLower(s) {
	case "active":
		return ui.Green("active")
	case "paused":
		return ui.Yellow("paused")
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
