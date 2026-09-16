package ui

import (
	"errors"
	"strings"
	"time"

	"beep/internal/client"
	"beep/internal/schedule"
	"beep/internal/workspace"

	"github.com/charmbracelet/huh"
)

// PromptBeepCreate prompts the user sequentially for beep creation parameters.
// If any parameter was already supplied via flags, its prompt is skipped.
// Optional parameters can be skipped by pressing Enter.
func PromptBeepCreate(initial client.CreateBeepParams) (*client.CreateBeepParams, error) {
	res := initial

	// 1. Title (required)
	if strings.TrimSpace(res.Title) == "" {
		err := huh.NewInput().
			Title("Reminder Title").
			Description("Short description of what you want to be reminded about").
			Placeholder("e.g. Deploy finished, Check server logs").
			Value(&res.Title).
			Validate(func(s string) error {
				if strings.TrimSpace(s) == "" {
					return errors.New("title is required")
				}
				return nil
			}).
			Run()
		if err != nil {
			return nil, err
		}
	}
	res.Title = strings.TrimSpace(res.Title)

	// 2. Message Body (optional)
	if strings.TrimSpace(res.Body) == "" {
		err := huh.NewInput().
			Title("Message Body (optional)").
			Description("Optional details or markdown body (press Enter to skip)").
			Value(&res.Body).
			Run()
		if err != nil {
			return nil, err
		}
	}
	res.Body = strings.TrimSpace(res.Body)

	// 3. Schedule Type & Value
	// If ScheduleKind is empty, prompt user to select schedule mode
	if res.ScheduleKind == "" {
		res.ScheduleKind = "instant"
		err := huh.NewSelect[string]().
			Title("Schedule Type").
			Description("How this reminder should be scheduled").
			Options(
				huh.NewOption("Instant (fire immediately)", "instant"),
				huh.NewOption("Relative delay (e.g. 15m, 2h, 1d)", "delay"),
				huh.NewOption("Specific time (e.g. 16:30, 2026-10-01 10:00)", "at"),
				huh.NewOption("Recurring cron (e.g. 0 9 * * 1-5)", "cron"),
			).
			Value(&res.ScheduleKind).
			Run()
		if err != nil {
			return nil, err
		}
	}

	// Step-by-step detail prompt based on schedule mode
	switch res.ScheduleKind {
	case "delay":
		if strings.TrimSpace(res.ScheduleVal) == "" {
			res.ScheduleVal = "15m"
			err := huh.NewInput().
				Title("Delay Duration").
				Description("Duration before firing (e.g. 10m, 2h, 1d)").
				Value(&res.ScheduleVal).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return errors.New("delay duration is required")
					}
					_, err := client.ParseInDuration(s)
					return err
				}).
				Run()
			if err != nil {
				return nil, err
			}
		}
	case "at":
		if strings.TrimSpace(res.ScheduleVal) == "" {
			res.ScheduleVal = "16:30"
			err := huh.NewInput().
				Title("Specific Time / Date").
				Description("When to fire (e.g. 16:30, 2026-10-01 10:00)").
				Value(&res.ScheduleVal).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return errors.New("time is required")
					}
					loc, err := time.LoadLocation(res.Timezone)
					if err != nil {
						loc = time.Local
					}
					_, err = client.ParseAtTime(s, loc)
					return err
				}).
				Run()
			if err != nil {
				return nil, err
			}
		}
	case "cron":
		if strings.TrimSpace(res.ScheduleVal) == "" {
			res.ScheduleVal = "0 9 * * 1-5"
			err := huh.NewInput().
				Title("Cron Expression").
				Description("Standard 5-part cron expression (e.g. 0 9 * * 1-5)").
				Value(&res.ScheduleVal).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return errors.New("cron expression is required")
					}
					return schedule.Validate(s)
				}).
				Run()
			if err != nil {
				return nil, err
			}
		}
	}

	// 4. Timezone
	if strings.TrimSpace(res.Timezone) == "" {
		if detected, ok := workspace.DetectTimezoneOK(); ok {
			res.Timezone = detected
		} else {
			res.Timezone = "UTC"
		}
		err := huh.NewInput().
			Title("Timezone").
			Description("Timezone for scheduling (press Enter to accept default)").
			Value(&res.Timezone).
			Run()
		if err != nil {
			return nil, err
		}
	}

	// 5. Notification Channels (optional)
	if strings.TrimSpace(res.Channels) == "" {
		err := huh.NewInput().
			Title("Notification Channels (optional)").
			Description("Comma-separated channel names or IDs (press Enter to skip)").
			Value(&res.Channels).
			Run()
		if err != nil {
			return nil, err
		}
	}
	res.Channels = strings.TrimSpace(res.Channels)

	return &res, nil
}
