package ui

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
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
			huh.NewOption("Body", "body"),
			huh.NewOption("Intent", "intent"),
			huh.NewOption("Metadata", "metadata"),
			huh.NewOption("Schedule", "schedule"),
			huh.NewOption("Timezone", "timezone"),
			huh.NewOption("Notification Channels", "channels"),
		).
		Value(&choice).
		Run()
	if err != nil || choice == "all" || choice == "" {
		return BeepFailedFields{Title: true, Body: true, Intent: true, Metadata: true, Schedule: true, Timezone: true, Channels: true}
	}
	var f BeepFailedFields
	switch choice {
	case "title":
		f.Title = true
	case "body":
		f.Body = true
	case "intent":
		f.Intent = true
	case "metadata":
		f.Metadata = true
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
			Title("Body (optional)").
			Description("Optional details or markdown body (press Enter to skip)").
			Value(&res.Body).
			Run()
		if err != nil {
			return nil, err
		}
	}
	res.Body = strings.TrimSpace(res.Body)

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
				if err := json.Unmarshal([]byte(s), &m); err != nil || m == nil {
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
		if err := json.Unmarshal([]byte(metadataStr), &m); err == nil && m != nil {
			res.Metadata = m
		}
	} else {
		res.Metadata = nil
	}

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

	// 2. Body
	if failed.Body {
		err := huh.NewInput().
			Title("Body (optional)").
			Description("Optional details or markdown body (press Enter to skip)").
			Value(&res.Body).
			Run()
		if err != nil {
			return nil, err
		}
		res.Body = strings.TrimSpace(res.Body)
	}

	// 3. Intent
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

	// 4. Metadata
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
				if err := json.Unmarshal([]byte(s), &m); err != nil || m == nil {
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
			if err := json.Unmarshal([]byte(metadataStr), &m); err == nil && m != nil {
				res.Metadata = m
			}
		} else {
			res.Metadata = nil
		}
	}

	// 5. Schedule
	if failed.Schedule {
		if res.ScheduleKind == "" {
			res.ScheduleKind = "instant"
		}
		if err := promptBeepSchedule(&res); err != nil {
			return nil, err
		}
	}

	// 6. Timezone (selectable list)
	if failed.Timezone {
		tz, err := PromptTimezone(res.Timezone)
		if err != nil {
			return nil, err
		}
		res.Timezone = tz
	}

	// 7. Notification Channels
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
	metaStr := ""
	if params.Metadata != nil && len(params.Metadata) > 0 {
		if metaBytes, err := json.Marshal(params.Metadata); err == nil {
			metaStr = string(metaBytes)
		}
	}
	items := []CreateSummaryItem{
		{Key: "Title", Value: params.Title, Failed: failed.Title},
		{Key: "Body", Value: params.Body, Failed: failed.Body},
		{Key: "Intent", Value: params.Intent, Failed: failed.Intent},
		{Key: "Metadata", Value: metaStr, Failed: failed.Metadata},
		{Key: "Schedule", Value: kind, Failed: failed.Schedule},
	}
	if params.ScheduleKind != "" && params.ScheduleKind != "instant" {
		items = append(items, CreateSummaryItem{Key: "When", Value: params.ScheduleVal, Failed: failed.Schedule})
	}
	items = append(items,
		CreateSummaryItem{Key: "Timezone", Value: params.Timezone, Failed: failed.Timezone},
		CreateSummaryItem{Key: "Channels", Value: params.Channels, Failed: failed.Channels},
	)
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
				return nil
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
				return nil
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

func parseDelayDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return 0, errors.New("empty duration")
	}

	reDay := regexp.MustCompile(`^(\d+)\s*(?:d|days?)$`)
	if m := reDay.FindStringSubmatch(s); len(m) == 2 {
		n, err := strconv.Atoi(m[1])
		if err != nil {
			return 0, err
		}
		return time.Duration(n) * 24 * time.Hour, nil
	}

	reWeek := regexp.MustCompile(`^(\d+)\s*(?:w|weeks?)$`)
	if m := reWeek.FindStringSubmatch(s); len(m) == 2 {
		n, err := strconv.Atoi(m[1])
		if err != nil {
			return 0, err
		}
		return time.Duration(n) * 7 * 24 * time.Hour, nil
	}

	s = strings.ReplaceAll(s, "minutes", "m")
	s = strings.ReplaceAll(s, "minute", "m")
	s = strings.ReplaceAll(s, "mins", "m")
	s = strings.ReplaceAll(s, "min", "m")
	s = strings.ReplaceAll(s, "hours", "h")
	s = strings.ReplaceAll(s, "hour", "h")
	s = strings.ReplaceAll(s, "hrs", "h")
	s = strings.ReplaceAll(s, "hr", "h")
	s = strings.ReplaceAll(s, "seconds", "s")
	s = strings.ReplaceAll(s, "second", "s")
	s = strings.ReplaceAll(s, "secs", "s")
	s = strings.ReplaceAll(s, "sec", "s")
	s = strings.ReplaceAll(s, " ", "")

	return time.ParseDuration(s)
}

func formatDurationFriendly(d time.Duration) string {
	if d <= 0 {
		return "0s"
	}
	days := int(d / (24 * time.Hour))
	rem := d % (24 * time.Hour)
	hours := int(rem / time.Hour)
	rem = rem % time.Hour
	minutes := int(rem / time.Minute)
	seconds := int((rem % time.Minute) / time.Second)

	var parts []string
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%dd", days))
	}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}
	if minutes > 0 {
		parts = append(parts, fmt.Sprintf("%dm", minutes))
	}
	if seconds > 0 && days == 0 && hours == 0 {
		parts = append(parts, fmt.Sprintf("%ds", seconds))
	}
	if len(parts) == 0 {
		return d.String()
	}
	return strings.Join(parts, " ")
}

func resolveLocation(tz string) *time.Location {
	if tz != "" {
		if loc, err := time.LoadLocation(tz); err == nil && loc != nil {
			return loc
		}
	}
	if detected, ok := workspace.DetectTimezoneOK(); ok {
		if loc, err := time.LoadLocation(detected); err == nil && loc != nil {
			return loc
		}
	}
	return time.Local
}

func isSameDay(t1, t2 time.Time) bool {
	return t1.Year() == t2.Year() && t1.YearDay() == t2.YearDay()
}

// FormatScheduleHuman returns a human-friendly key and converted natural language description.
func FormatScheduleHuman(scheduleKind, scheduleVal, tz string, now time.Time) (key string, formattedVal string) {
	loc := resolveLocation(tz)
	tzName := tz
	if tzName == "" {
		tzName = loc.String()
	}

	switch scheduleKind {
	case "cron":
		key = "Cron"
		s := strings.TrimSpace(scheduleVal)
		if s == "" {
			return key, Dim("(empty)")
		}
		desc := schedule.Describe(s)
		nextRun, ok := schedule.NextRun(s, now, loc)
		if ok {
			var nextStr string
			if isSameDay(nextRun, now.In(loc)) {
				nextStr = fmt.Sprintf("Today at %s (%s %s)", nextRun.Format("15:04"), nextRun.Format("2006-01-02 15:04:05"), tzName)
			} else if isSameDay(nextRun, now.In(loc).AddDate(0, 0, 1)) {
				nextStr = fmt.Sprintf("Tomorrow at %s (%s %s)", nextRun.Format("15:04"), nextRun.Format("2006-01-02 15:04:05"), tzName)
			} else {
				nextStr = fmt.Sprintf("%s %s", nextRun.Format("2006-01-02 15:04:05"), tzName)
			}

			if desc != "" {
				formattedVal = fmt.Sprintf("%s (%s · Next: %s)", s, desc, nextStr)
			} else {
				formattedVal = fmt.Sprintf("%s (Next: %s)", s, nextStr)
			}
		} else {
			if desc != "" {
				formattedVal = fmt.Sprintf("%s (%s)", s, desc)
			} else {
				formattedVal = s
			}
		}
		return key, formattedVal

	case "at":
		key = "Run At"
		s := strings.TrimSpace(scheduleVal)
		if s == "" {
			return key, Dim("(empty)")
		}
		localNow := now.In(loc)

		reTime := regexp.MustCompile(`^(\d{1,2}):(\d{2})(?::(\d{2}))?$`)
		if m := reTime.FindStringSubmatch(s); len(m) >= 3 {
			hour, _ := strconv.Atoi(m[1])
			minute, _ := strconv.Atoi(m[2])
			second := 0
			if len(m) == 4 && m[3] != "" {
				second, _ = strconv.Atoi(m[3])
			}
			if hour >= 0 && hour <= 23 && minute >= 0 && minute <= 59 && second >= 0 && second <= 59 {
				target := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), hour, minute, second, 0, loc)
				isTomorrow := false
				if !target.After(localNow) {
					target = target.AddDate(0, 0, 1)
					isTomorrow = true
				}
				if isTomorrow {
					formattedVal = fmt.Sprintf("%s (Tomorrow at %s · %s %s)", s, target.Format("15:04"), target.Format("2006-01-02 15:04:05"), tzName)
				} else {
					formattedVal = fmt.Sprintf("%s (Today at %s · %s %s)", s, target.Format("15:04"), target.Format("2006-01-02 15:04:05"), tzName)
				}
				return key, formattedVal
			}
		}

		var target time.Time
		var parseErr error
		layouts := []string{
			"2006-01-02 15:04:05",
			"2006-01-02 15:04",
			"2006/01/02 15:04:05",
			"2006/01/02 15:04",
			time.RFC3339,
			time.RFC3339Nano,
		}
		for _, layout := range layouts {
			if parsed, err := time.ParseInLocation(layout, s, loc); err == nil {
				target = parsed
				parseErr = nil
				break
			} else {
				parseErr = err
			}
		}
		if parseErr == nil {
			if target.Before(localNow) {
				formattedVal = fmt.Sprintf("%s (in the past · %s %s)", s, target.Format("2006-01-02 15:04:05"), tzName)
			} else if isSameDay(target, localNow) {
				formattedVal = fmt.Sprintf("%s (Today at %s · %s %s)", s, target.Format("15:04"), target.Format("2006-01-02 15:04:05"), tzName)
			} else if isSameDay(target, localNow.AddDate(0, 0, 1)) {
				formattedVal = fmt.Sprintf("%s (Tomorrow at %s · %s %s)", s, target.Format("15:04"), target.Format("2006-01-02 15:04:05"), tzName)
			} else {
				diff := target.Sub(localNow)
				days := int(diff.Hours() / 24)
				formattedVal = fmt.Sprintf("%s (%s %s · in %d days)", s, target.Format("2006-01-02 15:04:05"), tzName, days)
			}
			return key, formattedVal
		}

		return key, s

	case "delay":
		key = "Delay"
		s := strings.TrimSpace(scheduleVal)
		if s == "" {
			return key, Dim("(empty)")
		}
		dur, err := parseDelayDuration(s)
		if err != nil {
			return key, s
		}
		target := now.In(loc).Add(dur)
		durFriendly := formatDurationFriendly(dur)

		if isSameDay(target, now.In(loc)) {
			formattedVal = fmt.Sprintf("%s (in %s · Today at %s · %s %s)", s, durFriendly, target.Format("15:04:05"), target.Format("2006-01-02 15:04:05"), tzName)
		} else if isSameDay(target, now.In(loc).AddDate(0, 0, 1)) {
			formattedVal = fmt.Sprintf("%s (in %s · Tomorrow at %s · %s %s)", s, durFriendly, target.Format("15:04:05"), target.Format("2006-01-02 15:04:05"), tzName)
		} else {
			formattedVal = fmt.Sprintf("%s (in %s · %s %s)", s, durFriendly, target.Format("2006-01-02 15:04:05"), tzName)
		}
		return key, formattedVal

	default: // instant
		key = "Schedule"
		return key, "instant " + Dim("(fires immediately)")
	}
}

// PrintBeepPreview prints a structured summary of the beep before creation.
func PrintBeepPreview(params client.CreateBeepParams, header string) {
	if header == "" {
		header = "Proposed Beep:"
	}
	var displayChannels []string
	if strings.TrimSpace(params.Channels) != "" {
		for _, ch := range strings.Split(params.Channels, ",") {
			if trimmed := strings.TrimSpace(ch); trimmed != "" {
				displayChannels = append(displayChannels, trimmed)
			}
		}
	}

	fmt.Println()
	fmt.Println(Bold(Cyan("  " + header)))
	fmt.Println(KeyValue("Title", params.Title))
	if strings.TrimSpace(params.Body) != "" {
		if strings.Contains(params.Body, "\n") {
			fmt.Println(KeyValue("Body", ""))
			for _, line := range strings.Split(params.Body, "\n") {
				fmt.Printf("      %s\n", line)
			}
		} else {
			fmt.Println(KeyValue("Body", params.Body))
		}
	} else {
		fmt.Println(KeyValue("Body", Dim("(empty)")))
	}

	intentVal := params.Intent
	if strings.TrimSpace(intentVal) == "" {
		intentVal = Dim("(empty)")
	}
	fmt.Println(KeyValue("Intent", intentVal))

	metaVal := Dim("(empty)")
	if params.Metadata != nil && len(params.Metadata) > 0 {
		if metaBytes, err := json.Marshal(params.Metadata); err == nil {
			metaVal = string(metaBytes)
		}
	}
	fmt.Println(KeyValue("Metadata", metaVal))

	kind := "once"
	if params.ScheduleKind == "cron" {
		kind = "recurring"
	}
	fmt.Println(KeyValue("Kind", kind))

	schedKey, schedVal := FormatScheduleHuman(params.ScheduleKind, params.ScheduleVal, params.Timezone, time.Now())
	fmt.Println(KeyValue(schedKey, schedVal))

	if params.Timezone != "" {
		fmt.Println(KeyValue("Timezone", params.Timezone))
	}
	if len(displayChannels) > 0 {
		fmt.Println(KeyValue("Channels", strings.Join(displayChannels, ", ")))
	}
	fmt.Println()
}

// PromptBeepConfirmation displays the beep preview and prompts the user to
// create, edit, or cancel. It loops if the user chooses to edit.
// Returns (updatedParams, true, nil) to proceed, (nil, false, nil) on cancel.
func PromptBeepConfirmation(params client.CreateBeepParams, defaultChannels []string, header string) (*client.CreateBeepParams, bool, error) {
	current := params
	for {
		PrintBeepPreview(current, header)

		action, actionErr := PromptBeepProposalAction()
		if actionErr != nil {
			return nil, false, actionErr
		}
		switch action {
		case "cancel":
			fmt.Println(Dim("Cancelled."))
			return nil, false, nil
		case "edit":
			prompted, pErr := PromptBeepAdjust(current, defaultChannels, nil)
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

