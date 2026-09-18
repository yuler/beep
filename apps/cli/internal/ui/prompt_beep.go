package ui

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"beep/internal/client"
	"beep/internal/schedule"
	"beep/internal/workspace"

	"github.com/charmbracelet/huh"
)

// BeepFailedFields tracks which beep fields failed validation and need correction.
type BeepFailedFields struct {
	Title    bool
	Body     bool
	Schedule bool
	Timezone bool
	Channels bool
	Intent   bool
	Metadata bool
}

// Any reports whether any field failed.
func (f BeepFailedFields) Any() bool {
	return f.Title || f.Body || f.Schedule || f.Timezone || f.Channels || f.Intent || f.Metadata
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
		if strings.Contains(low, "intent") {
			f.Intent = true
		}
		if strings.Contains(low, "metadata") {
			f.Metadata = true
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
			huh.NewOption("Intent", "intent"),
			huh.NewOption("Metadata", "metadata"),
		).
		Value(&choice).
		Run()
	if err != nil || choice == "all" || choice == "" {
		return BeepFailedFields{Title: true, Body: true, Schedule: true, Timezone: true, Channels: true, Intent: true, Metadata: true}
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
	case "intent":
		f.Intent = true
	case "metadata":
		f.Metadata = true
	}
	return f
}

// PromptBeepCreate prompts the user sequentially for beep creation parameters.
// If any parameter was already supplied via flags, its prompt is skipped.
// Optional parameters can be skipped by pressing Enter.
// defaultChannels carries the channels selected in account settings and
// pre-selects them in the channels multi-select.
func PromptBeepCreate(initial client.CreateBeepParams, defaultChannels []string) (*client.CreateBeepParams, error) {
	return promptBeepForm(initial, defaultChannels, false)
}

// PromptBeepReview shows the full create form with current values pre-filled so
// the user can edit and confirm before submit (e.g. after AI proposal fallback).
func PromptBeepReview(initial client.CreateBeepParams, defaultChannels []string) (*client.CreateBeepParams, error) {
	return promptBeepForm(initial, defaultChannels, true)
}

func promptBeepForm(initial client.CreateBeepParams, defaultChannels []string, reviewAll bool) (*client.CreateBeepParams, error) {
	res := initial

	if reviewAll || strings.TrimSpace(res.Title) == "" {
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

	if reviewAll || strings.TrimSpace(res.Body) == "" {
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

	if reviewAll || res.ScheduleKind == "" {
		if res.ScheduleKind == "" {
			res.ScheduleKind = "instant"
		}
		if err := promptBeepSchedule(&res); err != nil {
			return nil, err
		}
	}

	if reviewAll || strings.TrimSpace(res.Timezone) == "" {
		tz, err := PromptTimezone(res.Timezone)
		if err != nil {
			return nil, err
		}
		res.Timezone = tz
	}

	if reviewAll || strings.TrimSpace(res.Channels) == "" {
		chDefaults := defaultChannels
		if strings.TrimSpace(res.Channels) != "" {
			chDefaults = strings.Split(res.Channels, ",")
		}
		channels, err := PromptNotificationChannels(chDefaults)
		if err != nil {
			return nil, err
		}
		res.Channels = channels
	}
	res.Channels = strings.TrimSpace(res.Channels)

	if reviewAll || strings.TrimSpace(res.Intent) == "" {
		err := huh.NewInput().
			Title("Intent Identifier (optional)").
			Description("Optional business intent slug (e.g. lunch_break, get_off_work)").
			Placeholder("e.g. lunch_break").
			Value(&res.Intent).
			Run()
		if err != nil {
			return nil, err
		}
	}
	res.Intent = strings.TrimSpace(res.Intent)

	var metadataStr string
	if res.Metadata != nil {
		if b, err := json.Marshal(res.Metadata); err == nil {
			metadataStr = string(b)
		}
	}
	if reviewAll || metadataStr == "" {
		err := huh.NewInput().
			Title("Metadata JSON (optional)").
			Description("Optional JSON object for downstream handlers").
			Placeholder(`e.g. {"key":"value"}`).
			Value(&metadataStr).
			Validate(func(s string) error {
				s = strings.TrimSpace(s)
				if s == "" {
					return nil
				}
				var m map[string]any
				if err := json.Unmarshal([]byte(s), &m); err != nil {
					return errors.New("must be a valid JSON object or empty")
				}
				return nil
			}).
			Run()
		if err != nil {
			return nil, err
		}
	}
	metadataStr = strings.TrimSpace(metadataStr)
	if metadataStr != "" {
		var m map[string]any
		if err := json.Unmarshal([]byte(metadataStr), &m); err == nil {
			res.Metadata = m
		}
	} else {
		res.Metadata = nil
	}

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

	// 6. Intent
	if failed.Intent {
		err := huh.NewInput().
			Title("Intent Identifier (optional)").
			Description("Optional business intent slug (e.g. lunch_break, get_off_work)").
			Placeholder("e.g. lunch_break").
			Value(&res.Intent).
			Run()
		if err != nil {
			return nil, err
		}
		res.Intent = strings.TrimSpace(res.Intent)
	}

	// 7. Metadata
	if failed.Metadata {
		var metadataStr string
		if res.Metadata != nil {
			if b, err := json.Marshal(res.Metadata); err == nil {
				metadataStr = string(b)
			}
		}
		err := huh.NewInput().
			Title("Metadata JSON (optional)").
			Description("Optional JSON object for downstream handlers").
			Placeholder(`e.g. {"key":"value"}`).
			Value(&metadataStr).
			Validate(func(s string) error {
				s = strings.TrimSpace(s)
				if s == "" {
					return nil
				}
				var m map[string]any
				if err := json.Unmarshal([]byte(s), &m); err != nil {
					return errors.New("must be a valid JSON object or empty")
				}
				return nil
			}).
			Run()
		if err != nil {
			return nil, err
		}
		metadataStr = strings.TrimSpace(metadataStr)
		if metadataStr != "" {
			var m map[string]any
			if err := json.Unmarshal([]byte(metadataStr), &m); err == nil {
				res.Metadata = m
			}
		} else {
			res.Metadata = nil
		}
	}

	return &res, nil
}

// CreateSummaryItem is one row in the printed create-form snapshot.
type CreateSummaryItem struct {
	Key    string
	Value  string
	Failed bool
}

// BeepCreateSummary returns the current beep form values, marking fields that
// failed validation so the user can see what to edit.
func BeepCreateSummary(params client.CreateBeepParams, errList []string) []CreateSummaryItem {
	failed := DetectBeepFailedFields(errList)
	kind := params.ScheduleKind
	if kind == "" {
		kind = "instant"
	}
	items := []CreateSummaryItem{
		{Key: "Title", Value: params.Title, Failed: failed.Title},
		{Key: "Body", Value: params.Body, Failed: failed.Body},
		{Key: "Schedule", Value: kind, Failed: failed.Schedule},
	}
	if params.ScheduleKind != "" && params.ScheduleKind != "instant" {
		items = append(items, CreateSummaryItem{Key: "When", Value: params.ScheduleVal, Failed: failed.Schedule})
	}
	items = append(items,
		CreateSummaryItem{Key: "Timezone", Value: params.Timezone, Failed: failed.Timezone},
		CreateSummaryItem{Key: "Channels", Value: params.Channels, Failed: failed.Channels},
	)
	if params.Intent != "" {
		items = append(items, CreateSummaryItem{Key: "Intent", Value: params.Intent, Failed: failed.Intent})
	}
	if params.Metadata != nil {
		if metaBytes, err := json.Marshal(params.Metadata); err == nil {
			items = append(items, CreateSummaryItem{Key: "Metadata", Value: string(metaBytes), Failed: failed.Metadata})
		}
	}
	return items
}

// PrintBeepCreateSummary prints the current beep form values after a failure.
func PrintBeepCreateSummary(params client.CreateBeepParams, errList []string) {
	printCreateSummary("Current values:", BeepCreateSummary(params, errList))
}

func printCreateSummary(header string, items []CreateSummaryItem) {
	fmt.Println()
	fmt.Println(Bold(Cyan("  " + header)))
	for _, item := range items {
		printCreateValue(item.Key, item.Value, item.Failed)
	}
}

func printCreateValue(key, val string, failed bool) {
	if strings.TrimSpace(val) == "" {
		val = Dim("(empty)")
	}
	if failed {
		fmt.Println(KeyValue(key, Red(val)+" "+Yellow("(needs update)")))
		return
	}
	fmt.Println(KeyValue(key, val))
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
		tzName := res.Timezone
		if tzName == "" {
			if detected, ok := workspace.DetectTimezoneOK(); ok {
				tzName = detected
			}
		}
		loc, _ := time.LoadLocation(tzName)
		if loc == nil {
			loc = time.Local
		}
		if strings.TrimSpace(res.ScheduleVal) == "" {
			res.ScheduleVal = "16:30"
		} else if parsed, err := time.Parse(time.RFC3339, res.ScheduleVal); err == nil {
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
				_, err := client.ParseAtTime(s, loc)
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
