package beep

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"beep/internal/client"
	"beep/internal/cmdutil"
	"beep/internal/config"
	"beep/internal/ui"
	"beep/internal/workspace"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

// NewCmdCreate creates the 'beep create' subcommand.
func NewCmdCreate() *cobra.Command {
	var (
		flagBody     string
		flagIn       string
		flagAt       string
		flagRunAt    string
		flagCron     string
		flagTimezone string
		flagChannels string
		flagChannel  string
		flagNatural  string
		flagIntent   string
		flagMetadata string
	)

	cmd := &cobra.Command{
		Use:   "create [title]",
		Short: "Create a beep",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if flagMetadata != "" {
				var tmp map[string]any
				if err := json.Unmarshal([]byte(flagMetadata), &tmp); err != nil || tmp == nil {
					return fmt.Errorf("invalid --metadata: must be a valid JSON object")
				}
			}
			return nil
		},
		Long: fmt.Sprintf(`Create a beep.

Scheduling modes (mutually exclusive):
  - Instant (default): fires immediately when no schedule flags are provided
  - Relative delay:    --in (e.g. 15m, 2h, 1d)
  - Specific datetime: --at (e.g. 16:30, "2026-10-01 10:00")
  - Recurring cron:    --cron (e.g. "0 10 * * 1-5")
  - AI natural:        --natural / -n (e.g. "remind me in 30 minutes to drink water")

Note: --cron, --in, --at, and --natural are mutually exclusive.
When using --json, interactive confirmation is skipped.
Optional --body and --channels can also be combined with --natural to supplement proposal fields.

Timezone precedence:
  If your Beep account has a configured User timezone, wall-clock times (e.g. --at 16:30)
  are scheduled in that user timezone. The --timezone flag / auto-detected machine timezone
  is sent as a fallback when no user timezone has been configured.

Examples:
  # Instant beep (fires immediately)
  %s create "Deploy finished"

  # Relative delay
  %s create "Meeting starts" --in 15m
  %s create "Check logs" --in 2h

  # Specific datetime
  %s create "Doctor appointment" --at "16:30"
  %s create "Release v1.0" --at "2026-10-01 10:00"

  # Recurring cron
  %s create "Daily Standup" --cron "0 10 * * 1-5"

  # Natural language via DeepSeek AI
  %s create -n "remind me in 30 minutes to drink water"`,
			config.BinaryName(), config.BinaryName(), config.BinaryName(),
			config.BinaryName(), config.BinaryName(), config.BinaryName(), config.BinaryName()),
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmdutil.RunWithClient(cmd, func(ctx context.Context, cfg *config.Config, c *client.Client) error {
				tz := flagTimezone
				if tz != "" {
					if !workspace.ValidIANATimezone(tz) {
						return fmt.Errorf("invalid --timezone %q: must be a valid IANA timezone (e.g. Asia/Shanghai, UTC, America/New_York)", tz)
					}
				} else {
					if detected, ok := workspace.DetectTimezoneOK(); ok {
						tz = detected
					} else {
						tz = "UTC"
					}
				}

				if flagChannel != "" {
					if flagChannels != "" && flagChannels != flagChannel {
						return fmt.Errorf("cannot specify both --channel and --channels")
					}
					flagChannels = flagChannel
				}

				if flagRunAt != "" {
					if flagAt != "" && flagAt != flagRunAt {
						return fmt.Errorf("cannot specify both --at and --run-at")
					}
					flagAt = flagRunAt
				}

				var metadataMap map[string]any
				if flagMetadata != "" {
					// Already validated in PreRunE; unmarshal here for use.
					_ = json.Unmarshal([]byte(flagMetadata), &metadataMap)
				}

				// 1. Natural language creation via -n / --natural flag
				if flagNatural != "" {
					err := handleNaturalCreate(ctx, c, cmd, flagNatural, flagBody, flagChannels, tz, flagIntent, metadataMap)
					if isUserAbort(ctx, err) {
						return nil
					}
					return err
				}

				var argText string
				if len(args) > 0 {
					argText = strings.TrimSpace(strings.Join(args, " "))
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
					Title:        argText,
					Body:         flagBody,
					ScheduleKind: schedKind,
					ScheduleVal:  schedVal,
					Timezone:     tz,
					Channels:     flagChannels,
					Intent:       flagIntent,
					Metadata:     metadataMap,
				}

				var b *client.Beep
				aiFallback := false
				if cmdutil.IsInteractive(cmd) {
					// Default channel selection comes from account settings.
					defaultChannels := client.DefaultNotificationChannels
					if s, err := c.GetSettings(ctx); err == nil && s != nil {
						defaultChannels = client.SanitizeChannelDefaults(s.NotificationChannels)
					}

					// When no explicit schedule flags are given in interactive mode:
					// 1. If args were provided (e.g. `beep create "natural text"`), default to AI proposal
					// 2. If no args were provided (e.g. `beep create`), directly proceed to interactive form
					if schedKind == "" && argText != "" {
						err := handleNaturalCreate(ctx, c, cmd, argText, flagBody, flagChannels, tz, flagIntent, metadataMap)
						if err == nil {
							return nil
						}
						if isUserAbort(ctx, err) {
							return nil
						}
						var proposalErr *aiProposalError
						if errors.As(err, &proposalErr) {
							// If AI proposal failed (e.g. offline/unconfigured), warn and fall back to manual form
							fmt.Println(ui.Warn("AI proposal unavailable (%v), falling back to form...", proposalErr.err))
							params.Title = argText
							aiFallback = true
							ui.PrintBeepCreateSummary(params, nil)
						} else {
							return err
						}
					}

					if shouldPromptBeepCreateForm(params.Title, params.ScheduleKind, len(args), aiFallback) {
						var prompted *client.CreateBeepParams
						var err error
						if aiFallback {
							prompted, err = ui.PromptBeepReview(params, defaultChannels)
						} else {
							prompted, err = ui.PromptBeepCreate(params, defaultChannels)
						}
						if err != nil {
							if isUserAbort(ctx, err) {
								return nil
							}
							return err
						}
						params = *prompted
					}

					// Interactive creates (form or flags) confirm via server preview.
					// --json / non-interactive skip this path entirely.
					confirmed, ok, cErr := confirmBeepWithPreview(ctx, c, params, defaultChannels, "Proposed Beep:")
					if cErr != nil {
						if isUserAbort(ctx, cErr) {
							return nil
						}
						return cErr
					}
					if !ok {
						return nil
					}
					params = *confirmed

					for {
						if params.ScheduleKind == "" {
							params.ScheduleKind = "instant"
						}
						req, err := params.ToRequest()
						if err != nil {
							errList := client.ExtractErrorList(err)
							ui.PrintErrorList("Invalid input", errList)
							ui.PrintBeepCreateSummary(params, errList)
							retry, promptErr := ui.PromptConfirm("Would you like to adjust your inputs?", true)
							if promptErr != nil {
								if isUserAbort(ctx, promptErr) {
									return nil
								}
								return promptErr
							}
							if !retry {
								return err
							}
							prompted, pErr := ui.PromptBeepAdjust(params, defaultChannels, errList)
							if pErr != nil {
								if isUserAbort(ctx, pErr) {
									return nil
								}
								return pErr
							}
							params = *prompted
							continue
						}

						_ = ui.WithSpinner("Creating beep...", func() error {
							b, err = c.CreateBeep(ctx, req)
							return err
						})
						if err == nil {
							break
						}
						if isUserAbort(ctx, err) {
							return nil
						}

						errList := client.ExtractErrorList(err)
						ui.PrintErrorList("Creation failed", errList)
						ui.PrintBeepCreateSummary(params, errList)
						retry, promptErr := ui.PromptConfirm("Would you like to adjust your inputs and retry?", true)
						if promptErr != nil {
							if isUserAbort(ctx, promptErr) {
								return nil
							}
							return promptErr
						}
						if !retry {
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
						return fmt.Errorf("beep title is required (e.g. %s create \"Meeting in 10m\" --in 10m)", config.BinaryName())
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
				if b.Intent != "" {
					fmt.Printf("  %s %s\n", ui.Dim("Intent:"), b.Intent)
				}
				if len(b.Metadata) > 0 {
					if metaBytes, err := json.Marshal(b.Metadata); err == nil {
						fmt.Printf("  %s %s\n", ui.Dim("Metadata:"), string(metaBytes))
					}
				}
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
		PostRun: func(cmd *cobra.Command, args []string) {
			flagBody = ""
			flagIn = ""
			flagAt = ""
			flagRunAt = ""
			flagCron = ""
			flagTimezone = ""
			flagChannels = ""
			flagChannel = ""
			flagNatural = ""
			flagIntent = ""
			flagMetadata = ""
		},
	}

	cmd.Flags().StringVarP(&flagBody, "body", "b", "", "Beep markdown body / message")
	cmd.Flags().StringVar(&flagIn, "in", "", "Delay duration before firing (e.g. 15m, 2h, 1d)")
	cmd.Flags().StringVar(&flagAt, "at", "", "Specific time to fire (e.g. 16:30, 2026-10-01 10:00)")
	cmd.Flags().StringVar(&flagRunAt, "run-at", "", "Alias for --at: specific time to fire")
	cmd.Flags().StringVarP(&flagCron, "cron", "c", "", "Recurring cron schedule (e.g. '0 9 * * *')")
	cmd.Flags().StringVarP(&flagTimezone, "timezone", "z", "", "Timezone fallback if not configured on user account (defaults to local timezone)")
	cmd.Flags().StringVar(&flagChannels, "channels", "", "Comma-separated notification channel names or IDs")
	cmd.Flags().StringVar(&flagChannel, "channel", "", "Alias for --channels: notification channel name or ID")
	cmd.Flags().StringVarP(&flagNatural, "natural", "n", "", "Natural language beep prompt parsed by AI")
	cmd.Flags().StringVar(&flagIntent, "intent", "", "Optional intent identifier (e.g. lunch_break)")
	cmd.Flags().StringVarP(&flagMetadata, "metadata", "m", "", "Optional metadata JSON object")
	cmd.MarkFlagsMutuallyExclusive("cron", "in", "at", "natural")
	cmd.MarkFlagsMutuallyExclusive("cron", "in", "run-at", "natural")
	cmd.MarkFlagsMutuallyExclusive("at", "run-at")
	cmd.MarkFlagsMutuallyExclusive("channels", "channel")

	return cmd
}

type aiProposalError struct {
	err error
}

func (e *aiProposalError) Error() string {
	return e.err.Error()
}

func (e *aiProposalError) Unwrap() error {
	return e.err
}

func isUserAbort(ctx context.Context, err error) bool {
	if ctx != nil && ctx.Err() != nil {
		return true
	}
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, huh.ErrUserAborted) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "user aborted") ||
		strings.Contains(msg, "context canceled") ||
		strings.Contains(msg, "canceled") ||
		strings.Contains(msg, "cancelled")
}

func shouldPromptBeepCreateForm(title, scheduleKind string, argCount int, aiFallback bool) bool {
	if aiFallback {
		return true
	}
	return strings.TrimSpace(title) == "" || (scheduleKind == "" && argCount == 0)
}

func confirmBeepWithPreview(ctx context.Context, c *client.Client, initial client.CreateBeepParams, defaultChannels []string, header string) (*client.CreateBeepParams, bool, error) {
	current := initial
	for {
		if current.ScheduleKind == "" {
			current.ScheduleKind = "instant"
		}
		req, err := current.ToRequest()
		if err != nil {
			errList := client.ExtractErrorList(err)
			ui.PrintErrorList("Invalid input", errList)
			ui.PrintBeepCreateSummary(current, errList)
			retry, promptErr := ui.PromptConfirm("Would you like to adjust your inputs?", true)
			if promptErr != nil {
				return nil, false, promptErr
			}
			if !retry {
				return nil, false, nil
			}
			prompted, pErr := ui.PromptBeepAdjust(current, defaultChannels, errList)
			if pErr != nil {
				return nil, false, pErr
			}
			current = *prompted
			continue
		}

		var preview *client.BeepPreview
		err = ui.WithSpinner("Fetching preview...", func() error {
			var pErr error
			preview, pErr = c.PreviewBeep(ctx, req)
			return pErr
		})
		if err != nil {
			if isUserAbort(ctx, err) {
				return nil, false, err
			}
			errList := client.ExtractErrorList(err)
			ui.PrintErrorList("Preview failed", errList)
			ui.PrintBeepCreateSummary(current, errList)
			retry, promptErr := ui.PromptConfirm("Would you like to adjust your inputs and retry?", true)
			if promptErr != nil {
				return nil, false, promptErr
			}
			if !retry {
				return nil, false, nil
			}
			prompted, pErr := ui.PromptBeepAdjust(current, defaultChannels, errList)
			if pErr != nil {
				return nil, false, pErr
			}
			current = *prompted
			continue
		}

		if !preview.Valid {
			ui.PrintErrorList("Invalid beep schedule", preview.Errors)
			ui.PrintBeepCreateSummary(current, preview.Errors)
			retry, promptErr := ui.PromptConfirm("Would you like to adjust your inputs?", true)
			if promptErr != nil {
				return nil, false, promptErr
			}
			if !retry {
				return nil, false, nil
			}
			prompted, pErr := ui.PromptBeepAdjust(current, defaultChannels, preview.Errors)
			if pErr != nil {
				return nil, false, pErr
			}
			current = *prompted
			continue
		}

		if current.Timezone == "" && preview.Timezone != "" {
			current.Timezone = preview.Timezone
		}

		ui.PrintBeepPreview(preview, header)

		action, actionErr := ui.PromptBeepProposalAction()
		if actionErr != nil {
			return nil, false, actionErr
		}
		switch action {
		case "cancel":
			fmt.Println(ui.Dim("Cancelled."))
			return nil, false, nil
		case "edit":
			prompted, pErr := ui.PromptBeepAdjust(current, defaultChannels, nil)
			if pErr != nil {
				return nil, false, pErr
			}
			current = *prompted
			continue
		case "create":
			return &current, true, nil
		default:
			return &current, true, nil
		}
	}
}

func handleNaturalCreate(ctx context.Context, c *client.Client, cmd *cobra.Command, prompt, bodyFlag, channelsFlag, tz, intentFlag string, metadata map[string]any) error {
	var proposal *client.BeepProposal
	err := ui.WithSpinner("Analyzing natural language prompt with AI...", func() error {
		var pErr error
		proposal, pErr = c.ProposeBeep(ctx, prompt, tz)
		return pErr
	})
	if err != nil {
		if isUserAbort(ctx, err) {
			return err
		}
		return &aiProposalError{err: fmt.Errorf("natural language parse failed: %w", err)}
	}

	if proposal.HasErrors() {
		return &aiProposalError{err: fmt.Errorf("could not understand beep: %s", strings.Join(proposal.Errors, ", "))}
	}
	if proposal.Title == "" {
		if proposal.Message != "" {
			return &aiProposalError{err: fmt.Errorf("could not understand beep: %s", proposal.Message)}
		}
		return &aiProposalError{err: fmt.Errorf("could not understand beep from prompt")}
	}

	body := proposal.Body
	if bodyFlag != "" {
		body = bodyFlag
	}

	resolvedTz := proposal.Timezone
	if resolvedTz == "" || !workspace.ValidIANATimezone(resolvedTz) {
		resolvedTz = tz
	}
	if resolvedTz == "" || !workspace.ValidIANATimezone(resolvedTz) {
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

	channels := channelsFlag
	if channels == "" && len(proposal.Channels) > 0 {
		channels = strings.Join(proposal.Channels, ",")
	}

	intent := intentFlag
	if intent == "" && proposal.Intent != "" {
		intent = proposal.Intent
	}

	metadataVal := metadata
	if metadataVal == nil && proposal.Metadata != nil {
		metadataVal = proposal.Metadata
	}

	params := client.CreateBeepParams{
		Title:        proposal.Title,
		Body:         body,
		ScheduleKind: schedKind,
		ScheduleVal:  schedVal,
		Timezone:     resolvedTz,
		Channels:     channels,
		Intent:       intent,
		Metadata:     metadataVal,
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
		confirmed, ok, cErr := confirmBeepWithPreview(ctx, c, params, defaultChannels, "Proposed Beep:")
		if cErr != nil {
			if isUserAbort(ctx, cErr) {
				return nil
			}
			return cErr
		}
		if !ok {
			return nil
		}
		params = *confirmed
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
			ui.PrintBeepCreateSummary(params, errList)
			retry, promptErr := ui.PromptConfirm("Would you like to adjust your inputs?", true)
			if promptErr != nil {
				return promptErr
			}
			if !retry {
				return err
			}
			prompted, pErr := ui.PromptBeepAdjust(params, defaultChannels, errList)
			if pErr != nil {
				return pErr
			}
			params = *prompted
			continue
		}

		_ = ui.WithSpinner("Creating beep...", func() error {
			b, err = c.CreateBeep(ctx, req)
			return err
		})
		if err != nil {
			if isUserAbort(ctx, err) {
				return err
			}
			if !cmdutil.IsInteractive(cmd) {
				return err
			}
			errList := client.ExtractErrorList(err)
			ui.PrintErrorList("Creation failed", errList)
			ui.PrintBeepCreateSummary(params, errList)
			retry, promptErr := ui.PromptConfirm("Would you like to adjust your inputs and retry?", true)
			if promptErr != nil {
				return promptErr
			}
			if !retry {
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
	if b.Intent != "" {
		fmt.Printf("  %s %s\n", ui.Dim("Intent:"), b.Intent)
	}
	if len(b.Metadata) > 0 {
		if metaBytes, err := json.Marshal(b.Metadata); err == nil {
			fmt.Printf("  %s %s\n", ui.Dim("Metadata:"), string(metaBytes))
		}
	}
	fmt.Printf("  %s %s\n", ui.Dim("Schedule:"), FormatBeepSchedule(b))
	if b.Timezone != "" {
		fmt.Printf("  %s %s\n", ui.Dim("Timezone:"), b.Timezone)
	}
	if len(b.NotificationChannels) > 0 {
		fmt.Printf("  %s %s\n", ui.Dim("Channels:"), strings.Join(b.NotificationChannels, ", "))
	}
	return nil
}
