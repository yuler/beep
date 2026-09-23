package ui

import (
	"strings"
	"time"
	_ "time/tzdata"
)

var timestampLayouts = []string{
	time.RFC3339Nano,
	time.RFC3339,
	"2006-01-02T15:04:05.999999999Z07:00",
	"2006-01-02T15:04:05.999999999-0700",
	"2006-01-02T15:04:05-0700",
	"2006-01-02T15:04:05.999999999",
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05.999999999Z07:00",
	"2006-01-02 15:04:05Z07:00",
	"2006-01-02 15:04:05 -0700",
	"2006-01-02 15:04:05.999999999-0700",
	"2006-01-02 15:04:05 MST",
	"2006-01-02 15:04:05",
}

// FormatTimestamp formats an API timestamp into a human-readable string in the given IANA timezone.
//
// Display rules:
// - Shape: YYYY-MM-DD HH:mm:ss (24-hour, no Z, no offset, no timezone suffix).
// - Empty or invalid IANA timezone falls back to UTC.
// - If the string is not a parseable timestamp, returns it unchanged.
// - Empty string stays empty.
func FormatTimestamp(raw, timezone string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	loc := parseLocation(timezone)

	t, ok := parseTimestamp(raw)
	if !ok {
		return raw
	}

	return t.In(loc).Format("2006-01-02 15:04:05")
}

func parseLocation(timezone string) *time.Location {
	timezone = strings.TrimSpace(timezone)
	if timezone == "" {
		return time.UTC
	}
	loc, err := time.LoadLocation(timezone)
	if err != nil || loc == nil {
		return time.UTC
	}
	return loc
}

func parseTimestamp(raw string) (time.Time, bool) {
	for _, layout := range timestampLayouts {
		var (
			t   time.Time
			err error
		)
		if strings.Contains(layout, "Z07:00") || strings.Contains(layout, "-0700") || strings.Contains(layout, "MST") || layout == time.RFC3339 || layout == time.RFC3339Nano {
			t, err = time.Parse(layout, raw)
		} else {
			t, err = time.ParseInLocation(layout, raw, time.UTC)
		}
		if err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}
