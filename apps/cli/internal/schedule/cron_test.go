package schedule

import "testing"

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
