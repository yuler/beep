package ui

import (
	"errors"
	"fmt"
	"strings"

	"beep/internal/client"
	"beep/internal/schedule"
	"beep/internal/workspace"

	"github.com/charmbracelet/huh"
)

// PromptBeeperCreate prompts the user sequentially for beeper creation parameters.
// If any parameter was already supplied via flags, its prompt is skipped.
// defaultChannels carries the channels selected in account settings and
// pre-selects them in the channels multi-select.
func PromptBeeperCreate(initial client.CreateBeeperParams, apps []*client.BeeperApp, defaultChannels []string) (*client.CreateBeeperParams, error) {
	return promptBeeper(initial, apps, defaultChannels, false)
}

// PromptBeeperAdjust prompts the user to review and adjust their inputs.
// All fields are presented with their current values pre-filled.
func PromptBeeperAdjust(initial client.CreateBeeperParams, apps []*client.BeeperApp, defaultChannels []string) (*client.CreateBeeperParams, error) {
	return promptBeeper(initial, apps, defaultChannels, true)
}

func promptBeeper(initial client.CreateBeeperParams, apps []*client.BeeperApp, defaultChannels []string, adjust bool) (*client.CreateBeeperParams, error) {
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
	if adjust || strings.TrimSpace(res.Title) == "" {
		if strings.TrimSpace(res.Title) == "" {
			res.Title = selectedApp.Name
		}
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
	if adjust || strings.TrimSpace(res.Body) == "" {
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
	if adjust || strings.TrimSpace(res.Cron) == "" {
		if strings.TrimSpace(res.Cron) == "" {
			res.Cron = selectedApp.DefaultCron
		}
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
		if !adjust {
			if _, exists := res.Config[input.Name]; exists {
				continue
			}
		}

		defaultVal := ""
		if input.Default != nil {
			defaultVal = fmt.Sprintf("%v", input.Default)
		}
		if current, exists := res.Config[input.Name]; exists && current != nil {
			defaultVal = fmt.Sprintf("%v", current)
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

	// 6. Timezone
	if adjust || strings.TrimSpace(res.Timezone) == "" {
		if strings.TrimSpace(res.Timezone) == "" {
			if detected, ok := workspace.DetectTimezoneOK(); ok {
				res.Timezone = detected
			} else {
				res.Timezone = "UTC"
			}
		}
		err := huh.NewInput().
			Title("Timezone").
			Description("Timezone for probe schedule (press Enter to accept default)").
			Value(&res.Timezone).
			Run()
		if err != nil {
			return nil, err
		}
	}

	// 7. Notification Channels (optional, multi-select with account defaults)
	if adjust || strings.TrimSpace(res.Channels) == "" {
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
