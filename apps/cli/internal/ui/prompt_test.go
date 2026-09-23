package ui

import (
	"strings"
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
			name:     "intent only",
			errList:  []string{"Intent is not valid"},
			expected: BeepFailedFields{Intent: true},
		},
		{
			name:     "metadata only",
			errList:  []string{"Metadata must be a hash"},
			expected: BeepFailedFields{Metadata: true},
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
			if got.Intent != tc.expected.Intent {
				t.Errorf("Intent: expected %v, got %v", tc.expected.Intent, got.Intent)
			}
			if got.Metadata != tc.expected.Metadata {
				t.Errorf("Metadata: expected %v, got %v", tc.expected.Metadata, got.Metadata)
			}
		})
	}
}

func TestBeepCreateSummaryMarksFailedSchedule(t *testing.T) {
	params := client.CreateBeepParams{
		Title:        "哈哈",
		Body:         "喝水",
		ScheduleKind: "cron",
		ScheduleVal:  "not-a-cron",
		Timezone:     "Asia/Shanghai",
		Channels:     "cli",
	}
	items := BeepCreateSummary(params, []string{"Cron is not a valid cron expression"})
	got := map[string]CreateSummaryItem{}
	for _, item := range items {
		got[item.Key] = item
	}
	if got["Title"].Value != "哈哈" || got["Title"].Failed {
		t.Fatalf("Title = %+v, want value 哈哈 and not failed", got["Title"])
	}
	if !got["Schedule"].Failed || got["When"].Value != "not-a-cron" || !got["When"].Failed {
		t.Fatalf("schedule summary = Schedule%+v When%+v, want both failed with cron value", got["Schedule"], got["When"])
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

func TestRequireNotificationChannels(t *testing.T) {
	if err := requireNotificationChannels(nil); err == nil {
		t.Fatal("expected error for nil selection")
	}
	if err := requireNotificationChannels([]string{}); err == nil {
		t.Fatal("expected error for empty selection")
	}
	if err := requireNotificationChannels([]string{"  ", ""}); err == nil {
		t.Fatal("expected error for blank-only selection")
	}
	if err := requireNotificationChannels([]string{"cli"}); err != nil {
		t.Fatalf("expected cli selection to be valid, got %v", err)
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

func TestFormatScheduleHuman(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	// 2026-09-23 17:41:00 Wednesday
	now := time.Date(2026, 9, 23, 17, 41, 0, 0, loc)

	// 1. Instant
	k, v := FormatScheduleHuman("instant", "", "Asia/Shanghai", now)
	if k != "Schedule" || !strings.Contains(v, "instant") || !strings.Contains(v, "fires immediately") {
		t.Errorf("instant got key=%q val=%q", k, v)
	}

	// 2. Delay: 15m
	k, v = FormatScheduleHuman("delay", "15m", "Asia/Shanghai", now)
	if k != "Delay" || !strings.Contains(v, "15m") || !strings.Contains(v, "Today at 17:56:00") || !strings.Contains(v, "Asia/Shanghai") {
		t.Errorf("delay 15m got key=%q val=%q", k, v)
	}

	// 3. Delay: 20h (crosses midnight)
	k, v = FormatScheduleHuman("delay", "20h", "Asia/Shanghai", now)
	if k != "Delay" || !strings.Contains(v, "Tomorrow at 13:41:00") {
		t.Errorf("delay 20h got key=%q val=%q", k, v)
	}

	// 4. Delay: 1d
	k, v = FormatScheduleHuman("delay", "1d", "Asia/Shanghai", now)
	if k != "Delay" || !strings.Contains(v, "Tomorrow at 17:41:00") {
		t.Errorf("delay 1d got key=%q val=%q", k, v)
	}

	// 5. At: 16:30 (already passed at 17:41 today, must be tomorrow!)
	k, v = FormatScheduleHuman("at", "16:30", "Asia/Shanghai", now)
	if k != "Run At" || !strings.Contains(v, "Tomorrow at 16:30") || !strings.Contains(v, "2026-09-24 16:30:00") {
		t.Errorf("at 16:30 (after) got key=%q val=%q", k, v)
	}

	// 6. At: 18:30 (still in future today!)
	k, v = FormatScheduleHuman("at", "18:30", "Asia/Shanghai", now)
	if k != "Run At" || !strings.Contains(v, "Today at 18:30") || !strings.Contains(v, "2026-09-23 18:30:00") {
		t.Errorf("at 18:30 (before) got key=%q val=%q", k, v)
	}

	// 7. At: specific date
	k, v = FormatScheduleHuman("at", "2026-10-01 10:00", "Asia/Shanghai", now)
	if k != "Run At" || !strings.Contains(v, "2026-10-01 10:00:00") || !strings.Contains(v, "in 7 days") {
		t.Errorf("at specific date got key=%q val=%q", k, v)
	}

	// 8. Cron: 0 9 * * 1-5
	k, v = FormatScheduleHuman("cron", "0 9 * * 1-5", "Asia/Shanghai", now)
	if k != "Cron" || !strings.Contains(v, "Every weekday at 09:00") || !strings.Contains(v, "2026-09-24 09:00:00") {
		t.Errorf("cron 0 9 * * 1-5 got key=%q val=%q", k, v)
	}

	// 9. Cron: */5 * * * *
	k, v = FormatScheduleHuman("cron", "*/5 * * * *", "Asia/Shanghai", now)
	if k != "Cron" || !strings.Contains(v, "Every 5 minutes") || !strings.Contains(v, "2026-09-23 17:45:00") {
		t.Errorf("cron */5 * * * * got key=%q val=%q", k, v)
	}
}
