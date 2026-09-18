package ui

import (
	"errors"
	"fmt"
	"strings"

	"beep/internal/client"
	"beep/internal/schedule"

	"github.com/charmbracelet/huh"
)

// BeeperFailedFields tracks which beeper fields failed validation and need correction.
type BeeperFailedFields struct {
	App      bool
	Title    bool
	Body     bool
	Cron     bool
	Timezone bool
	Channels bool
	Configs  map[string]bool
}

// Any reports whether any field failed.
func (f BeeperFailedFields) Any() bool {
	return f.App || f.Title || f.Body || f.Cron || f.Timezone || f.Channels || len(f.Configs) > 0
}

// DetectBeeperFailedFields inspects error messages to identify which fields failed.
func DetectBeeperFailedFields(errList []string, app *client.BeeperApp) BeeperFailedFields {
	f := BeeperFailedFields{
		Configs: make(map[string]bool),
	}
	for _, raw := range errList {
		low := strings.ToLower(raw)
		if strings.Contains(low, "title") {
			f.Title = true
		}
		if strings.Contains(low, "body") {
			f.Body = true
		}
		if strings.Contains(low, "cron") || strings.Contains(low, "schedule") {
			f.Cron = true
		}
		if strings.Contains(low, "timezone") || strings.Contains(low, "iana") {
			f.Timezone = true
		}
		if strings.Contains(low, "channel") || strings.Contains(low, "notification") {
			f.Channels = true
		}
		if strings.Contains(low, "app") {
			f.App = true
		}
		if app != nil {
			for _, input := range app.Inputs {
				inputLower := strings.ToLower(input.Name)
				if strings.Contains(low, inputLower) {
					f.Configs[input.Name] = true
				}
			}
		}
		if strings.Contains(low, "config") && len(f.Configs) == 0 && app != nil {
			for _, input := range app.Inputs {
				f.Configs[input.Name] = true
			}
		}
	}
	return f
}

func promptBeeperSelectFieldToAdjust() BeeperFailedFields {
	var choice string
	err := huh.NewSelect[string]().
		Title("Which field would you like to adjust?").
		Options(
			huh.NewOption("All fields", "all"),
			huh.NewOption("Title", "title"),
			huh.NewOption("Description / Body", "body"),
			huh.NewOption("Cron Schedule", "cron"),
			huh.NewOption("App Configuration", "config"),
			huh.NewOption("Timezone", "timezone"),
			huh.NewOption("Notification Channels", "channels"),
		).
		Value(&choice).
		Run()
	if err != nil || choice == "all" || choice == "" {
		f := BeeperFailedFields{Title: true, Body: true, Cron: true, Timezone: true, Channels: true}
		f.Configs = map[string]bool{"*": true}
		return f
	}
	f := BeeperFailedFields{Configs: make(map[string]bool)}
	switch choice {
	case "title":
		f.Title = true
	case "body":
		f.Body = true
	case "cron":
		f.Cron = true
	case "config":
		f.Configs["*"] = true
	case "timezone":
		f.Timezone = true
	case "channels":
		f.Channels = true
	}
	return f
}

// PromptBeeperCreate prompts the user sequentially for beeper creation parameters.
// If any parameter was already supplied via flags, its prompt is skipped.
// defaultChannels carries the channels selected in account settings and
// pre-selects them in the channels multi-select.
func PromptBeeperCreate(initial client.CreateBeeperParams, apps []*client.BeeperApp, defaultChannels []string) (*client.CreateBeeperParams, error) {
	if len(apps) == 0 {
		return nil, errors.New("no beeper apps available on server")
	}

	res := initial
	if res.Config == nil {
		res.Config = make(map[string]any)
	}

	var selectedApp *client.BeeperApp

	// 1. App Template Selection
	if strings.TrimSpace(res.AppSlug) != "" {
		for _, a := range apps {
			if a.Slug == res.AppSlug {
				selectedApp = a
				break
			}
		}
	}

	if selectedApp == nil {
		appOptions := make([]huh.Option[string], 0, len(apps))
		for _, a := range apps {
			label := fmt.Sprintf("%s (%s)", a.Name, a.Slug)
			appOptions = append(appOptions, huh.NewOption(label, a.Slug))
		}

		err := huh.NewSelect[string]().
			Title("Select Probe App Template").
			Options(appOptions...).
			Value(&res.AppSlug).
			Run()
		if err != nil {
			return nil, err
		}

		for _, a := range apps {
			if a.Slug == res.AppSlug {
				selectedApp = a
				break
			}
		}
		if selectedApp == nil {
			return nil, fmt.Errorf("selected app %s not found", res.AppSlug)
		}
	}

	// 2. Title
	if strings.TrimSpace(res.Title) == "" {
		res.Title = selectedApp.Name
		err := huh.NewInput().
			Title("Beeper Title").
			Description("Name of this monitor probe").
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

	// 3. Message Body / Description (optional)
	if strings.TrimSpace(res.Body) == "" {
		err := huh.NewInput().
			Title("Description / Body (optional)").
			Description("Optional details or markdown body (press Enter to skip)").
			Value(&res.Body).
			Run()
		if err != nil {
			return nil, err
		}
	}
	res.Body = strings.TrimSpace(res.Body)

	// 4. Cron Schedule
	if strings.TrimSpace(res.Cron) == "" {
		res.Cron = selectedApp.DefaultCron
		err := huh.NewInput().
			Title("Cron Schedule").
			Description(fmt.Sprintf("Schedule expression (default: %s)", selectedApp.DefaultCron)).
			Value(&res.Cron).
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
	res.Cron = strings.TrimSpace(res.Cron)

	// 5. Dynamic App Config Inputs
	for _, input := range selectedApp.Inputs {
		if _, exists := res.Config[input.Name]; exists {
			continue
		}

		defaultVal := ""
		if input.Default != nil {
			defaultVal = fmt.Sprintf("%v", input.Default)
		}

		val := defaultVal
		fieldTitle := fmt.Sprintf("Config: %s", input.Name)
		if input.Required {
			fieldTitle += " (required)"
		} else {
			fieldTitle += " (optional)"
		}

		err := huh.NewInput().
			Title(fieldTitle).
			Description(input.Description).
			Value(&val).
			Validate(func(s string) error {
				if input.Required && strings.TrimSpace(s) == "" {
					return fmt.Errorf("%s is required", input.Name)
				}
				return nil
			}).
			Run()
		if err != nil {
			return nil, err
		}

		trimmedVal := strings.TrimSpace(val)
		if trimmedVal != "" {
			res.Config[input.Name] = trimmedVal
		} else {
			delete(res.Config, input.Name)
		}
	}

	// 6. Timezone (selectable list)
	if strings.TrimSpace(res.Timezone) == "" {
		tz, err := PromptTimezone(res.Timezone)
		if err != nil {
			return nil, err
		}
		res.Timezone = tz
	}

	// 7. Notification Channels (multi-select, at least one required)
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

// PromptBeeperAdjust prompts the user to adjust only the failed content/fields.
// If errList does not contain recognizable fields, the user can choose which field to adjust.
func PromptBeeperAdjust(initial client.CreateBeeperParams, apps []*client.BeeperApp, defaultChannels []string, errList []string) (*client.CreateBeeperParams, error) {
	var selectedApp *client.BeeperApp
	for _, a := range apps {
		if a.Slug == initial.AppSlug {
			selectedApp = a
			break
		}
	}

	failed := DetectBeeperFailedFields(errList, selectedApp)
	if !failed.Any() {
		failed = promptBeeperSelectFieldToAdjust()
	}

	res := initial
	if res.Config == nil {
		res.Config = make(map[string]any)
	}

	// 1. App Template
	if failed.App || selectedApp == nil {
		appOptions := make([]huh.Option[string], 0, len(apps))
		for _, a := range apps {
			label := fmt.Sprintf("%s (%s)", a.Name, a.Slug)
			appOptions = append(appOptions, huh.NewOption(label, a.Slug))
		}

		err := huh.NewSelect[string]().
			Title("Select Probe App Template").
			Options(appOptions...).
			Value(&res.AppSlug).
			Run()
		if err != nil {
			return nil, err
		}

		selectedApp = nil
		for _, a := range apps {
			if a.Slug == res.AppSlug {
				selectedApp = a
				break
			}
		}
		if selectedApp == nil {
			return nil, fmt.Errorf("selected app %s not found", res.AppSlug)
		}
	}

	// 2. Title
	if failed.Title {
		err := huh.NewInput().
			Title("Beeper Title").
			Description("Name of this monitor probe").
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

	// 3. Description / Body
	if failed.Body {
		err := huh.NewInput().
			Title("Description / Body (optional)").
			Description("Optional details or markdown body (press Enter to skip)").
			Value(&res.Body).
			Run()
		if err != nil {
			return nil, err
		}
		res.Body = strings.TrimSpace(res.Body)
	}

	// 4. Cron Schedule
	if failed.Cron {
		defaultCron := "0 9 * * 1-5"
		if selectedApp != nil && selectedApp.DefaultCron != "" {
			defaultCron = selectedApp.DefaultCron
		}
		if strings.TrimSpace(res.Cron) == "" {
			res.Cron = defaultCron
		}
		err := huh.NewInput().
			Title("Cron Schedule").
			Description(fmt.Sprintf("Schedule expression (default: %s)", defaultCron)).
			Value(&res.Cron).
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
		res.Cron = strings.TrimSpace(res.Cron)
	}

	// 5. Dynamic App Config Inputs (only prompt failed ones)
	if selectedApp != nil && len(failed.Configs) > 0 {
		promptAllConfigs := failed.Configs["*"]
		for _, input := range selectedApp.Inputs {
			if !promptAllConfigs && !failed.Configs[input.Name] {
				continue
			}

			val := ""
			if current, exists := res.Config[input.Name]; exists && current != nil {
				val = fmt.Sprintf("%v", current)
			} else if input.Default != nil {
				val = fmt.Sprintf("%v", input.Default)
			}

			fieldTitle := fmt.Sprintf("Config: %s", input.Name)
			if input.Required {
				fieldTitle += " (required)"
			} else {
				fieldTitle += " (optional)"
			}

			err := huh.NewInput().
				Title(fieldTitle).
				Description(input.Description).
				Value(&val).
				Validate(func(s string) error {
					if input.Required && strings.TrimSpace(s) == "" {
						return fmt.Errorf("%s is required", input.Name)
					}
					return nil
				}).
				Run()
			if err != nil {
				return nil, err
			}

			trimmedVal := strings.TrimSpace(val)
			if trimmedVal != "" {
				res.Config[input.Name] = trimmedVal
			} else {
				delete(res.Config, input.Name)
			}
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

// BeeperCreateSummary returns the current beeper form values, marking fields
// that failed validation so the user can see what to edit.
func BeeperCreateSummary(params client.CreateBeeperParams, apps []*client.BeeperApp, errList []string) []CreateSummaryItem {
	var app *client.BeeperApp
	for _, a := range apps {
		if a.Slug == params.AppSlug {
			app = a
			break
		}
	}
	failed := DetectBeeperFailedFields(errList, app)
	items := []CreateSummaryItem{
		{Key: "App", Value: params.AppSlug, Failed: failed.App},
		{Key: "Title", Value: params.Title, Failed: failed.Title},
		{Key: "Body", Value: params.Body, Failed: failed.Body},
		{Key: "Cron", Value: params.Cron, Failed: failed.Cron},
	}
	if app != nil {
		for _, input := range app.Inputs {
			val := ""
			if current, exists := params.Config[input.Name]; exists && current != nil {
				val = fmt.Sprintf("%v", current)
			}
			label := strings.TrimSpace(input.Label)
			if label == "" {
				label = input.Name
			}
			items = append(items, CreateSummaryItem{
				Key:    label,
				Value:  val,
				Failed: failed.Configs[input.Name] || failed.Configs["*"],
			})
		}
	} else {
		for key, current := range params.Config {
			val := ""
			if current != nil {
				val = fmt.Sprintf("%v", current)
			}
			items = append(items, CreateSummaryItem{
				Key:    key,
				Value:  val,
				Failed: failed.Configs[key] || failed.Configs["*"],
			})
		}
	}
	items = append(items,
		CreateSummaryItem{Key: "Timezone", Value: params.Timezone, Failed: failed.Timezone},
		CreateSummaryItem{Key: "Channels", Value: params.Channels, Failed: failed.Channels},
	)
	return items
}

// PrintBeeperCreateSummary prints the current beeper form values after a failure.
func PrintBeeperCreateSummary(params client.CreateBeeperParams, apps []*client.BeeperApp, errList []string) {
	printCreateSummary("Current values:", BeeperCreateSummary(params, apps, errList))
}
