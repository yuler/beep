package ui

import (
	"testing"
	"time"

	"beep/internal/client"
)

func TestDetectBeepFailedFields(t *testing.T) {
	tests := []struct {
		name     string
		errList  []string
		expected BeepFailedFields
	}{
		{
			name:     "title only",
			errList:  []string{"Title can't be blank"},
			expected: BeepFailedFields{Title: true},
		},
		{
			name:     "cron schedule only",
			errList:  []string{"Cron is not a valid cron expression"},
			expected: BeepFailedFields{Schedule: true},
		},
		{
			name:     "run_at schedule only",
			errList:  []string{"Run at must be in the future"},
			expected: BeepFailedFields{Schedule: true},
		},
		{
			name:     "timezone only",
			errList:  []string{"Timezone is not a valid IANA timezone"},
			expected: BeepFailedFields{Timezone: true},
		},
		{
			name:     "channels only",
			errList:  []string{"Notification channels contain unknown channels: sms"},
			expected: BeepFailedFields{Channels: true},
		},
		{
			name:     "multiple fields: title and cron",
			errList:  []string{"Title can't be blank", "Cron is not a valid cron expression"},
			expected: BeepFailedFields{Title: true, Schedule: true},
		},
		{
			name:     "no matching fields",
			errList:  []string{"Internal server error occurred"},
			expected: BeepFailedFields{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := DetectBeepFailedFields(tc.errList)
			if got.Title != tc.expected.Title {
				t.Errorf("Title: expected %v, got %v", tc.expected.Title, got.Title)
			}
			if got.Body != tc.expected.Body {
				t.Errorf("Body: expected %v, got %v", tc.expected.Body, got.Body)
			}
			if got.Schedule != tc.expected.Schedule {
				t.Errorf("Schedule: expected %v, got %v", tc.expected.Schedule, got.Schedule)
			}
			if got.Timezone != tc.expected.Timezone {
				t.Errorf("Timezone: expected %v, got %v", tc.expected.Timezone, got.Timezone)
			}
			if got.Channels != tc.expected.Channels {
				t.Errorf("Channels: expected %v, got %v", tc.expected.Channels, got.Channels)
			}
		})
	}
}

func TestDetectBeeperFailedFields(t *testing.T) {
	app := &client.BeeperApp{
		Slug: "http",
		Name: "HTTP Monitor",
		Inputs: []client.BeeperAppInput{
			{Name: "target_url", Required: true},
			{Name: "method", Required: false},
		},
	}

	tests := []struct {
		name     string
		errList  []string
		expected BeeperFailedFields
	}{
		{
			name:     "title error",
			errList:  []string{"Title can't be blank"},
			expected: BeeperFailedFields{Title: true, Configs: map[string]bool{}},
		},
		{
			name:     "cron error",
			errList:  []string{"Cron expression is invalid"},
			expected: BeeperFailedFields{Cron: true, Configs: map[string]bool{}},
		},
		{
			name:     "specific config input error",
			errList:  []string{"Config target_url is required"},
			expected: BeeperFailedFields{Configs: map[string]bool{"target_url": true}},
		},
		{
			name:     "timezone error",
			errList:  []string{"Timezone is invalid"},
			expected: BeeperFailedFields{Timezone: true, Configs: map[string]bool{}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := DetectBeeperFailedFields(tc.errList, app)
			if got.Title != tc.expected.Title {
				t.Errorf("Title: expected %v, got %v", tc.expected.Title, got.Title)
			}
			if got.Cron != tc.expected.Cron {
				t.Errorf("Cron: expected %v, got %v", tc.expected.Cron, got.Cron)
			}
			if got.Timezone != tc.expected.Timezone {
				t.Errorf("Timezone: expected %v, got %v", tc.expected.Timezone, got.Timezone)
			}
			for k, v := range tc.expected.Configs {
				if got.Configs[k] != v {
					t.Errorf("Configs[%s]: expected %v, got %v", k, v, got.Configs[k])
				}
			}
		})
	}
}

func TestCommonIANATimezonesAreValid(t *testing.T) {
	zones := CommonIANATimezones()
	if len(zones) == 0 {
		t.Fatal("expected CommonIANATimezones to not be empty")
	}
	for _, z := range zones {
		if _, err := time.LoadLocation(z); err != nil {
			t.Errorf("zone %q in CommonIANATimezones is not loadable: %v", z, err)
		}
	}
}
