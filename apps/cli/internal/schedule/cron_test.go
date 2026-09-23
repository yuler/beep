package schedule

import (
	"testing"
	"time"
)

func TestValidateClassic(t *testing.T) {
	ok := []string{
		"*/5 * * * *",
		"* * * * *",
		"0 * * * *",
		"0 0 * * *",
		"30 4 * * *",
		"0 22 * * 1-5",
		"*/15 * * * * *",
		"0 0 1 jan *",
		"*/5 * * * * Asia/Shanghai",
	}
	for _, s := range ok {
		if err := Validate(s); err != nil {
			t.Errorf("Validate(%q) unexpected error: %v", s, err)
		}
	}

	bad := []string{
		"",
		"invalid",
		"not a cron",
		"60 * * * *",
		"* * * *",
		"* * * * * * * *",
	}
	for _, s := range bad {
		if err := Validate(s); err == nil {
			t.Errorf("Validate(%q) expected error", s)
		}
	}
}

func TestValidateNat(t *testing.T) {
	ok := []string{
		"every 5 minutes",
		"every day at noon",
		"every weekday at 9am",
		"every 3 hours",
		"every monday at 5pm",
		"every day at midnight",
		"every 15 minutes",
		"every day at 12:30",
	}
	for _, s := range ok {
		if err := Validate(s); err != nil {
			t.Errorf("Validate(%q) unexpected error: %v", s, err)
		}
	}

	bad := []string{
		"every",
		"every banana",
		"every foo bar",
	}
	for _, s := range bad {
		if err := Validate(s); err == nil {
			t.Errorf("Validate(%q) expected error", s)
		}
	}
}

func TestDescribe(t *testing.T) {
	tests := []struct {
		expr string
		want string
	}{
		{"0 9 * * 1-5", "Every weekday at 09:00"},
		{"0 9 * * *", "Every day at 09:00"},
		{"0 0 * * *", "Every day at midnight (00:00)"},
		{"0 12 * * *", "Every day at noon (12:00)"},
		{"*/5 * * * *", "Every 5 minutes"},
		{"0 * * * *", "Every hour"},
		{"0 9 * * 1", "Every Monday at 09:00"},
		{"every 5 minutes", "Every 5 minutes"},
		{"every day at noon", "Every day at noon"},
	}

	for _, tc := range tests {
		got := Describe(tc.expr)
		if got != tc.want {
			t.Errorf("Describe(%q) = %q, want %q", tc.expr, got, tc.want)
		}
	}
}

func TestNextRun(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	// 2026-09-23 17:41:00 (Wednesday)
	now := time.Date(2026, 9, 23, 17, 41, 0, 0, loc)

	// 1. "0 9 * * 1-5" (Mon-Fri at 09:00) -> next is 2026-09-24 09:00:00 (Thursday)
	next, ok := NextRun("0 9 * * 1-5", now, loc)
	if !ok {
		t.Fatalf("NextRun('0 9 * * 1-5') expected ok")
	}
	expected := time.Date(2026, 9, 24, 9, 0, 0, 0, loc)
	if !next.Equal(expected) {
		t.Errorf("NextRun('0 9 * * 1-5') = %v, want %v", next, expected)
	}

	// 2. "*/5 * * * *" -> next is 2026-09-23 17:45:00
	next5, ok := NextRun("*/5 * * * *", now, loc)
	if !ok {
		t.Fatalf("NextRun('*/5 * * * *') expected ok")
	}
	expected5 := time.Date(2026, 9, 23, 17, 45, 0, 0, loc)
	if !next5.Equal(expected5) {
		t.Errorf("NextRun('*/5 * * * *') = %v, want %v", next5, expected5)
	}

	// 3. "every 15 minutes" -> next is 17:56:00
	nextNat, ok := NextRun("every 15 minutes", now, loc)
	if !ok {
		t.Fatalf("NextRun('every 15 minutes') expected ok")
	}
	expectedNat := now.Add(15 * time.Minute)
	if !nextNat.Equal(expectedNat) {
		t.Errorf("NextRun('every 15 minutes') = %v, want %v", nextNat, expectedNat)
	}
}
