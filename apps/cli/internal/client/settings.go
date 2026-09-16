package client

import (
	"context"
	"fmt"
)

// Notification channel kinds supported by the server (mirrors User::NOTIFICATION_CHANNELS).
var NotificationChannelKinds = []string{"email", "web_push", "cli"}

// DefaultNotificationChannels matches the server default for new users.
var DefaultNotificationChannels = []string{"email"}

type SettingsResponse struct {
	ID                   string   `json:"id"`
	Name                 string   `json:"name"`
	Slug                 string   `json:"slug"`
	Personal             bool     `json:"personal"`
	NotificationChannels []string `json:"notification_channels"`
	Timezone             *string  `json:"timezone"`
	TimezoneSource       *string  `json:"timezone_source"`
}

// GetSettings fetches account settings for the current account scope.
// The response includes notification_channels selected in account settings,
// used as defaults for interactive channel prompts.
func (c *Client) GetSettings(ctx context.Context) (*SettingsResponse, error) {
	url := fmt.Sprintf("%s/api/v1/settings", c.cfg.ServerURL)
	var res SettingsResponse
	if err := c.getAuthJSON(ctx, url, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// SanitizeChannelDefaults filters defaults down to known channel kinds.
// Unknown entries are dropped; empty result falls back to the server default.
func SanitizeChannelDefaults(defaults []string) []string {
	allowed := make(map[string]bool, len(NotificationChannelKinds))
	for _, k := range NotificationChannelKinds {
		allowed[k] = true
	}
	seen := make(map[string]bool, len(defaults))
	out := make([]string, 0, len(defaults))
	for _, d := range defaults {
		if allowed[d] && !seen[d] {
			seen[d] = true
			out = append(out, d)
		}
	}
	if len(out) == 0 {
		return append([]string(nil), DefaultNotificationChannels...)
	}
	return out
}
