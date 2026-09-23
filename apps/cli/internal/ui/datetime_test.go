package ui

import (
	"testing"
)

func TestFormatTimestamp(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		timezone string
		want     string
	}{
		{
			name:     "Asia/Shanghai conversion from UTC",
			raw:      "2026-09-23T10:00:00Z",
			timezone: "Asia/Shanghai",
			want:     "2026-09-23 18:00:00",
		},
		{
			name:     "Asia/Shanghai with fractional seconds",
			raw:      "2026-09-23T10:00:00.000Z",
			timezone: "Asia/Shanghai",
			want:     "2026-09-23 18:00:00",
		},
		{
			name:     "America/New_York daylight saving time",
			raw:      "2026-09-23T10:00:00Z",
			timezone: "America/New_York",
			want:     "2026-09-23 06:00:00",
		},
		{
			name:     "empty timezone falls back to UTC",
			raw:      "2026-09-23T10:00:00Z",
			timezone: "",
			want:     "2026-09-23 10:00:00",
		},
		{
			name:     "invalid timezone falls back to UTC",
			raw:      "2026-09-23T10:00:00Z",
			timezone: "Invalid/Timezone_Name",
			want:     "2026-09-23 10:00:00",
		},
		{
			name:     "empty raw string returns empty",
			raw:      "",
			timezone: "Asia/Shanghai",
			want:     "",
		},
		{
			name:     "whitespace only raw string returns empty",
			raw:      "   ",
			timezone: "Asia/Shanghai",
			want:     "",
		},
		{
			name:     "unparseable string returned unchanged",
			raw:      "not-a-timestamp",
			timezone: "Asia/Shanghai",
			want:     "not-a-timestamp",
		},
		{
			name:     "dash placeholder returned unchanged",
			raw:      "-",
			timezone: "Asia/Shanghai",
			want:     "-",
		},
		{
			name:     "input with positive offset converted to another timezone",
			raw:      "2026-09-23T18:00:00+08:00",
			timezone: "UTC",
			want:     "2026-09-23 10:00:00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatTimestamp(tt.raw, tt.timezone)
			if got != tt.want {
				t.Errorf("FormatTimestamp(%q, %q) = %q, want %q", tt.raw, tt.timezone, got, tt.want)
			}
		})
	}
}
