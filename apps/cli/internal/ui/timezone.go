package ui

import (
	"fmt"
	"strings"

	"beep/internal/workspace"

	"github.com/charmbracelet/huh"
)

// CommonIANATimezones returns popular world timezones to prioritize in selection.
func CommonIANATimezones() []string {
	return []string{
		"UTC",
		"Asia/Shanghai",
		"Asia/Hong_Kong",
		"Asia/Tokyo",
		"Asia/Singapore",
		"Asia/Seoul",
		"Asia/Kolkata",
		"Asia/Dubai",
		"Europe/London",
		"Europe/Paris",
		"Europe/Berlin",
		"America/New_York",
		"America/Chicago",
		"America/Denver",
		"America/Los_Angeles",
		"Australia/Sydney",
		"Pacific/Auckland",
	}
}

// PromptTimezone prompts the user to select an IANA timezone from a selectable list.
// The list includes the current/detected timezone at the top, common timezones,
// and all system IANA timezones with type-to-filter support.
func PromptTimezone(current string) (string, error) {
	detected, _ := workspace.DetectTimezoneOK()
	if detected == "" {
		detected = "UTC"
	}

	selected := strings.TrimSpace(current)
	if selected == "" {
		selected = detected
	}

	rawZones := workspace.ListIANATimezones()
	seen := make(map[string]bool)
	var orderedZones []string

	addZone := func(z string) {
		z = strings.TrimSpace(z)
		if z != "" && !seen[z] {
			seen[z] = true
			orderedZones = append(orderedZones, z)
		}
	}

	// 1. Current / detected zone first so it's immediately accessible
	addZone(selected)
	addZone(detected)

	// 2. Common world timezones
	for _, z := range CommonIANATimezones() {
		addZone(z)
	}

	// 3. All other available IANA timezones
	for _, z := range rawZones {
		addZone(z)
	}

	options := make([]huh.Option[string], 0, len(orderedZones))
	for _, z := range orderedZones {
		label := z
		if z == detected {
			label = fmt.Sprintf("%s (host detected)", z)
		}
		options = append(options, huh.NewOption(label, z))
	}

	err := huh.NewSelect[string]().
		Title("Timezone").
		Description("Select IANA timezone (type to filter)").
		Options(options...).
		Value(&selected).
		Run()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(selected), nil
}
