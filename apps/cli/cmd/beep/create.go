package beep

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"beep/internal/client"
	"beep/internal/cmdutil"
	"beep/internal/config"
	"beep/internal/ui"
	"beep/internal/workspace"

	"github.com/spf13/cobra"
)

// NewCmdCreate creates the 'beep create' subcommand.
func NewCmdCreate() *cobra.Command {
	var (
		flagBody     string
		flagIn       string
		flagAt       string
		flagCron     string
		flagTimezone string
		flagChannels string
		flagNatural  string
	)

	cmd := &cobra.Command{
		Use:   "create [title]",
		Short: "Create a beep",
		Long: `Create a beep.

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
  # Instant beep (fires immediately)
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
			return cmdutil.RunWithClient(cmd, func(ctx context.Context, cfg *config.Config, c *client.Client) error {
				tz := flagTimezone
				if tz != "" {
					if !workspace.ValidIANATimezone(tz) {
						return fmt.Errorf("invalid --timezone %q: must be a valid IANA timezone (e.g. Asia/Shanghai, UTC, America/New_York)", tz)
					}
				} else if !cmdutil.IsInteractive(cmd) {
					if detected, ok := workspace.DetectTimezoneOK(); ok {
						tz = detected
					} else {
						tz = "UTC"
					}
				}

				// 1. Natural language creation
				if flagNatural != "" {
					return handleNaturalCreate(ctx, c, cmd, flagNatural, flagBody, flagChannels, tz)
				}

				var title string
				if len(args) > 0 {
					title = strings.TrimSpace(args[0])
				}

				schedKind := ""
				schedVal := ""
				if flagCron != "" {
					schedKind = "cron"
					schedVal = flagCron
				} else if flagIn != "" {
					schedKind = "delay"
					schedVal = flagIn
				} else if flagAt != "" {
					schedKind = "at"
					schedVal = flagAt
				}

				params := client.CreateBeepParams{
					Title:        title,
					Body:         flagBody,
					ScheduleKind: schedKind,
					ScheduleVal:  schedVal,
					Timezone:     tz,
					Channels:     flagChannels,
				}

				var b *client.Beep
				if cmdutil.IsInteractive(cmd) {
					// Default channel selection comes from account settings.
					defaultChannels := client.DefaultNotificationChannels
					if s, err := c.GetSettings(ctx); err == nil && s != nil {
						defaultChannels = client.SanitizeChannelDefaults(s.NotificationChannels)
					}

					isMissingInfo := params.Title == "" || (params.ScheduleKind == "" && len(args) == 0)
					if isMissingInfo {
						prompted, err := ui.PromptBeepCreate(params, defaultChannels)
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
							errList := client.ExtractErrorList(err)
							ui.PrintErrorList("Invalid input", errList)
							retry, promptErr := ui.PromptConfirm("Would you like to adjust your inputs?", true)
							if promptErr != nil || !retry {
								return err
							}
							prompted, pErr := ui.PromptBeepAdjust(params, defaultChannels, errList)
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

						errList := client.ExtractErrorList(err)
						ui.PrintErrorList("Creation failed", errList)
						retry, promptErr := ui.PromptConfirm("Would you like to adjust your inputs and retry?", true)
						if promptErr != nil || !retry {
							return err
						}
						prompted, pErr := ui.PromptBeepAdjust(params, defaultChannels, errList)
						if pErr != nil {
							return pErr
						}
						params = *prompted
					}
				} else {
					if params.Title == "" {
						return fmt.Errorf("beep title is required (e.g. beep beep create \"Meeting in 10m\" --in 10m)")
					}
					if params.Timezone == "" {
						if detected, ok := workspace.DetectTimezoneOK(); ok {
							params.Timezone = detected
						} else {
							params.Timezone = "UTC"
						}
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

				if cmdutil.IsJSON(cmd) {
					data, err := json.MarshalIndent(b, "", "  ")
					if err != nil {
						return err
					}
					fmt.Println(string(data))
					return nil
				}

				scheduleInfo := FormatBeepSchedule(b)
				fmt.Println(ui.Success("Created beep %s (%s)", ui.Bold(b.Title), ui.Dim(b.ID)))
				fmt.Printf("  %s %s\n", ui.Dim("Schedule:"), scheduleInfo)
				if b.Timezone != "" {
					fmt.Printf("  %s %s\n", ui.Dim("Timezone:"), b.Timezone)
				}
				if len(b.NotificationChannels) > 0 {
					fmt.Printf("  %s %s\n", ui.Dim("Channels:"), strings.Join(b.NotificationChannels, ", "))
				}
				return nil
			})
		},
	}

	cmd.Flags().StringVarP(&flagBody, "body", "b", "", "Beep markdown body / message")
	cmd.Flags().StringVar(&flagIn, "in", "", "Delay duration before firing (e.g. 15m, 2h, 1d)")
	cmd.Flags().StringVar(&flagAt, "at", "", "Specific time to fire (e.g. 16:30, 2026-10-01 10:00)")
	cmd.Flags().StringVarP(&flagCron, "cron", "c", "", "Recurring cron schedule (e.g. '0 9 * * *')")
	cmd.Flags().StringVarP(&flagTimezone, "timezone", "z", "", "Timezone (defaults to local timezone)")
	cmd.Flags().StringVar(&flagChannels, "channels", "", "Comma-separated notification channel names or IDs")
	cmd.Flags().StringVarP(&flagNatural, "natural", "n", "", "Natural language beep prompt parsed by AI")
	cmd.MarkFlagsMutuallyExclusive("cron", "in", "at", "natural")

	return cmd
}

func handleNaturalCreate(ctx context.Context, c *client.Client, cmd *cobra.Command, prompt, bodyFlag, channelsFlag, tz string) error {
	proposal, err := c.ProposeBeep(ctx, prompt, tz)
	if err != nil {
		return fmt.Errorf("natural language parse failed: %w", err)
	}

	if proposal.HasErrors() {
		return fmt.Errorf("could not understand beep: %s", strings.Join(proposal.Errors, ", "))
	}
	if proposal.Title == "" {
		if proposal.Message != "" {
			return fmt.Errorf("could not understand beep: %s", proposal.Message)
		}
		return fmt.Errorf("could not understand beep from prompt")
	}

	body := proposal.Body
	if bodyFlag != "" {
		body = bodyFlag
	}

	resolvedTz := proposal.Timezone
	if resolvedTz == "" {
		resolvedTz = tz
	}
	if resolvedTz == "" {
		resolvedTz = "UTC"
	}

	schedKind := "instant"
	schedVal := ""
	if proposal.Kind == "recurring" && proposal.Cron != "" {
		schedKind = "cron"
		schedVal = proposal.Cron
	} else if proposal.RunAt != "" {
		schedKind = "at"
		schedVal = proposal.RunAt
	}

	params := client.CreateBeepParams{
		Title:        proposal.Title,
		Body:         body,
		ScheduleKind: schedKind,
		ScheduleVal:  schedVal,
		Timezone:     resolvedTz,
		Channels:     channelsFlag,
	}

	defaultChannels := client.DefaultNotificationChannels
	if cmdutil.IsInteractive(cmd) {
		if s, err := c.GetSettings(ctx); err == nil && s != nil {
			defaultChannels = client.SanitizeChannelDefaults(s.NotificationChannels)
		}
		if params.Channels == "" && len(defaultChannels) > 0 {
			params.Channels = strings.Join(defaultChannels, ",")
		}
	}

	if cmdutil.IsInteractive(cmd) && !cmdutil.IsJSON(cmd) {
		for {
			var displayChannels []string
			if strings.TrimSpace(params.Channels) != "" {
				for _, ch := range strings.Split(params.Channels, ",") {
					if trimmed := strings.TrimSpace(ch); trimmed != "" {
						displayChannels = append(displayChannels, trimmed)
					}
				}
			}

			fmt.Println()
			fmt.Println(ui.Bold(ui.Cyan("  Proposed Beep:")))
			fmt.Println(ui.KeyValue("Title", params.Title))
			if params.Body != "" {
				if strings.Contains(params.Body, "\n") {
					fmt.Println(ui.KeyValue("Body", ""))
					for _, line := range strings.Split(params.Body, "\n") {
						fmt.Printf("      %s\n", line)
					}
				} else {
					fmt.Println(ui.KeyValue("Body", params.Body))
				}
			}

			kind := "once"
			if params.ScheduleKind == "cron" {
				kind = "recurring"
			}
			fmt.Println(ui.KeyValue("Kind", kind))

			switch params.ScheduleKind {
			case "cron":
				fmt.Println(ui.KeyValue("Cron", params.ScheduleVal))
			case "at":
				fmt.Println(ui.KeyValue("Run At", params.ScheduleVal))
			case "delay":
				fmt.Println(ui.KeyValue("Delay", params.ScheduleVal))
			default:
				fmt.Println(ui.KeyValue("Schedule", ui.Dim("instant")))
			}

			if params.Timezone != "" {
				fmt.Println(ui.KeyValue("Timezone", params.Timezone))
			}
			if len(displayChannels) > 0 {
				fmt.Println(ui.KeyValue("Channels", strings.Join(displayChannels, ", ")))
			}
			fmt.Println()

			action, actionErr := ui.PromptBeepProposalAction()
			if actionErr != nil {
				return actionErr
			}
			if action == "cancel" {
				fmt.Println(ui.Dim("Cancelled."))
				return nil
			}
			if action == "edit" {
				prompted, pErr := ui.PromptBeepAdjust(params, defaultChannels, nil)
				if pErr != nil {
					return pErr
				}
				params = *prompted
				continue
			}

			break
		}
	}

	var b *client.Beep
	for {
		if params.ScheduleKind == "" {
			params.ScheduleKind = "instant"
		}
		req, err := params.ToRequest()
		if err != nil {
			if !cmdutil.IsInteractive(cmd) {
				return err
			}
			errList := client.ExtractErrorList(err)
			ui.PrintErrorList("Invalid input", errList)
			retry, promptErr := ui.PromptConfirm("Would you like to adjust your inputs?", true)
			if promptErr != nil || !retry {
				return err
			}
			prompted, pErr := ui.PromptBeepAdjust(params, defaultChannels, errList)
			if pErr != nil {
				return pErr
			}
			params = *prompted
			continue
		}

		b, err = c.CreateBeep(ctx, req)
		if err != nil {
			if !cmdutil.IsInteractive(cmd) {
				return err
			}
			errList := client.ExtractErrorList(err)
			ui.PrintErrorList("Creation failed", errList)
			retry, promptErr := ui.PromptConfirm("Would you like to adjust your inputs and retry?", true)
			if promptErr != nil || !retry {
				return err
			}
			prompted, pErr := ui.PromptBeepAdjust(params, defaultChannels, errList)
			if pErr != nil {
				return pErr
			}
			params = *prompted
			continue
		}
		break
	}

	if cmdutil.IsJSON(cmd) {
		data, err := json.MarshalIndent(b, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	fmt.Println(ui.Success("Created beep %s (%s)", ui.Bold(b.Title), ui.Dim(b.ID)))
	fmt.Printf("  %s %s\n", ui.Dim("Schedule:"), FormatBeepSchedule(b))
	if b.Timezone != "" {
		fmt.Printf("  %s %s\n", ui.Dim("Timezone:"), b.Timezone)
	}
	if len(b.NotificationChannels) > 0 {
		fmt.Printf("  %s %s\n", ui.Dim("Channels:"), strings.Join(b.NotificationChannels, ", "))
	}
	return nil
}
