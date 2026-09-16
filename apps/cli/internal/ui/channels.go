package ui

import (
	"strings"

	"beep/internal/client"

	"github.com/charmbracelet/huh"
)

// NotificationChannelOptions returns the fixed set of selectable channels,
// mirroring account settings (email / web_push / cli).
func NotificationChannelOptions() []huh.Option[string] {
	return []huh.Option[string]{
		huh.NewOption("Email", "email"),
		huh.NewOption("Web Push", "web_push"),
		huh.NewOption("CLI", "cli"),
	}
}

// PromptNotificationChannels shows a multi-select over the existing channel
// kinds. Selected values default to the channels chosen in account settings.
// An empty selection means "use account default" and returns "".
func PromptNotificationChannels(defaultChannels []string) (string, error) {
	selected := client.SanitizeChannelDefaults(defaultChannels)

	err := huh.NewMultiSelect[string]().
		Title("Notification Channels").
		Description("Select where this fires (Space to toggle, Enter to confirm; empty = account default)").
		Options(NotificationChannelOptions()...).
		Value(&selected).
		Run()
	if err != nil {
		return "", err
	}

	// Preserve option order (email, web_push, cli) regardless of toggle order.
	ordered := make([]string, 0, len(selected))
	included := make(map[string]bool, len(selected))
	for _, s := range selected {
		included[strings.TrimSpace(s)] = true
	}
	for _, kind := range client.NotificationChannelKinds {
		if included[kind] {
			ordered = append(ordered, kind)
		}
	}
	return strings.Join(ordered, ","), nil
}
