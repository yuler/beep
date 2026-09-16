package ui

import (
	"errors"
	"strings"
	"time"

	"beep/internal/client"
	"beep/internal/schedule"

	"github.com/charmbracelet/huh"
)

// BeepFailedFields tracks which beep fields failed validation and need correction.
type BeepFailedFields struct {
	Title    bool
	Body     bool
	Schedule bool
	Timezone bool
	Channels bool
}

// Any reports whether any field failed.
func (f BeepFailedFields) Any() bool {
	return f.Title || f.Body || f.Schedule || f.Timezone || f.Channels
}

// DetectBeepFailedFields inspects error messages to identify which fields failed.
func DetectBeepFailedFields(errList []string) BeepFailedFields {
	var f BeepFailedFields
	for _, raw := range errList {
		low := strings.ToLower(raw)
		if strings.Contains(low, "title") {
			f.Title = true
		}
		if strings.Contains(low, "body") {
			f.Body = true
		}
		if strings.Contains(low, "cron") ||
			strings.Contains(low, "run_at") ||
			strings.Contains(low, "run at") ||
			strings.Contains(low, "schedule") ||
			strings.Contains(low, "delay") ||
			strings.Contains(low, "duration") ||
			(strings.Contains(low, "time") && !strings.Contains(low, "timezone")) {
			f.Schedule = true
		}
		if strings.Contains(low, "timezone") || strings.Contains(low, "iana") {
			f.Timezone = true
		}
		if strings.Contains(low, "channel") || strings.Contains(low, "notification") {
			f.Channels = true
		}
	}
	return f
}

func promptBeepSelectFieldToAdjust() BeepFailedFields {
	var choice string
	err := huh.NewSelect[string]().
		Title("Which field would you like to adjust?").
		Options(
			huh.NewOption("All fields", "all"),
			huh.NewOption("Title", "title"),
			huh.NewOption("Message Body", "body"),
			huh.NewOption("Schedule", "schedule"),
			huh.NewOption("Timezone", "timezone"),
			huh.NewOption("Notification Channels", "channels"),
		).
		Value(&choice).
		Run()
	if err != nil || choice == "all" || choice == "" {
		return BeepFailedFields{Title: true, Body: true, Schedule: true, Timezone: true, Channels: true}
	}
	var f BeepFailedFields
	switch choice {
	case "title":
		f.Title = true
	case "body":
		f.Body = true
	case "schedule":
		f.Schedule = true
	case "timezone":
		f.Timezone = true
	case "channels":
		f.Channels = true
	}
	return f
}

// PromptBeepCreate prompts the user sequentially for beep creation parameters.
// If any parameter was already supplied via flags, its prompt is skipped.
// Optional parameters can be skipped by pressing Enter.
// defaultChannels carries the channels selected in account settings and
// pre-selects them in the channels multi-select.
func PromptBeepCreate(initial client.CreateBeepParams, defaultChannels []string) (*client.CreateBeepParams, error) {
	res := initial

	// 1. Title (required)
	if strings.TrimSpace(res.Title) == "" {
		err := huh.NewInput().
			Title("Beep Title").
			Description("Name of this beep, shown when it fires").
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
	if res.ScheduleKind == "" {
		res.ScheduleKind = "instant"
		if err := promptBeepSchedule(&res); err != nil {
			return nil, err
		}
	}

	// 4. Timezone (selectable list)
	if strings.TrimSpace(res.Timezone) == "" {
		tz, err := PromptTimezone(res.Timezone)
		if err != nil {
			return nil, err
		}
		res.Timezone = tz
	}

	// 5. Notification Channels (optional, multi-select with account defaults)
	if strings.TrimSpace(res.Channels) == "" {
		channels, err := PromptNotificationChannels(defaultChannels)
		if err != nil {
			return nil, err
		}
		res.Channels = channels
	}
	res.Channels = strings.TrimSpace(res.Channels)

	return &res, nil
}

// PromptBeepAdjust prompts the user to adjust only the failed content/fields.
// If errList does not contain recognizable fields, the user can choose which field to adjust.
func PromptBeepAdjust(initial client.CreateBeepParams, defaultChannels []string, errList []string) (*client.CreateBeepParams, error) {
	failed := DetectBeepFailedFields(errList)
	if !failed.Any() {
		failed = promptBeepSelectFieldToAdjust()
	}

	res := initial

	// 1. Title
	if failed.Title {
		err := huh.NewInput().
			Title("Beep Title").
			Description("Name of this beep, shown when it fires").
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
		res.Title = strings.TrimSpace(res.Title)
	}

	// 2. Message Body
	if failed.Body {
		err := huh.NewInput().
			Title("Message Body (optional)").
			Description("Optional details or markdown body (press Enter to skip)").
			Value(&res.Body).
			Run()
		if err != nil {
			return nil, err
		}
		res.Body = strings.TrimSpace(res.Body)
	}

	// 3. Schedule
	if failed.Schedule {
		if res.ScheduleKind == "" {
			res.ScheduleKind = "instant"
		}
		if err := promptBeepSchedule(&res); err != nil {
			return nil, err
		}
	}

	// 4. Timezone (selectable list)
	if failed.Timezone {
		tz, err := PromptTimezone(res.Timezone)
		if err != nil {
			return nil, err
		}
		res.Timezone = tz
	}

	// 5. Notification Channels
	if failed.Channels {
		chDefaults := defaultChannels
		if strings.TrimSpace(res.Channels) != "" {
			chDefaults = strings.Split(res.Channels, ",")
		}
		channels, err := PromptNotificationChannels(chDefaults)
		if err != nil {
			return nil, err
		}
		res.Channels = strings.TrimSpace(channels)
	}

	return &res, nil
}

func promptBeepSchedule(res *client.CreateBeepParams) error {
	err := huh.NewSelect[string]().
		Title("Schedule Type").
		Description("How this beep should be scheduled").
		Options(
			huh.NewOption("Instant (fire immediately)", "instant"),
			huh.NewOption("Relative delay (e.g. 15m, 2h, 1d)", "delay"),
			huh.NewOption("Specific time (e.g. 16:30, 2026-10-01 10:00)", "at"),
			huh.NewOption("Recurring cron (e.g. 0 9 * * 1-5)", "cron"),
		).
		Value(&res.ScheduleKind).
		Run()
	if err != nil {
		return err
	}

	switch res.ScheduleKind {
	case "instant":
		res.ScheduleVal = ""
	case "delay":
		if strings.TrimSpace(res.ScheduleVal) == "" {
			res.ScheduleVal = "15m"
		}
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
			return err
		}
	case "at":
		if strings.TrimSpace(res.ScheduleVal) == "" {
			res.ScheduleVal = "16:30"
		} else if parsed, err := time.Parse(time.RFC3339, res.ScheduleVal); err == nil {
			loc, locErr := time.LoadLocation(res.Timezone)
			if locErr != nil {
				loc = time.Local
			}
			res.ScheduleVal = parsed.In(loc).Format("2006-01-02 15:04")
		}
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
			return err
		}
	case "cron":
		if strings.TrimSpace(res.ScheduleVal) == "" {
			res.ScheduleVal = "0 9 * * 1-5"
		}
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
			return err
		}
	}
	return nil
}

// PromptBeepProposalAction asks the user what to do with a proposed beep: create, edit, or cancel.
func PromptBeepProposalAction() (string, error) {
	var choice string = "create"
	err := huh.NewSelect[string]().
		Title("Create this beep?").
		Options(
			huh.NewOption("Create beep", "create"),
			huh.NewOption("Edit details", "edit"),
			huh.NewOption("Cancel", "cancel"),
		).
		Value(&choice).
		Run()
	if err != nil {
		return "", err
	}
	return choice, nil
}

// PromptBeepCreateMode asks the user how they want to create the beep: natural language or form.
func PromptBeepCreateMode() (string, error) {
	var mode string = "natural"
	err := huh.NewSelect[string]().
		Title("Creation Mode").
		Description("How would you like to create this beep?").
		Options(
			huh.NewOption("Natural language (AI prompt)", "natural"),
			huh.NewOption("Interactive form (step-by-step)", "form"),
		).
		Value(&mode).
		Run()
	if err != nil {
		return "", err
	}
	return mode, nil
}

// PromptBeepNaturalPrompt prompts for a natural language description of the beep.
func PromptBeepNaturalPrompt() (string, error) {
	var promptText string
	err := huh.NewInput().
		Title("Natural Language Prompt").
		Description("Describe your beep in plain language (e.g. 'remind me in 30 minutes to drink water')").
		Placeholder("e.g. Check server logs tomorrow at 10am").
		Value(&promptText).
		Validate(func(s string) error {
			if strings.TrimSpace(s) == "" {
				return errors.New("prompt is required")
			}
			return nil
		}).
		Run()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(promptText), nil
}
