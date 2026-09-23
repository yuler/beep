// Package schedule validates cron expressions compatible with Rails Fugit
// (classic 5/6-field cron and Fugit::Nat semantic phrases like "every 5 minutes").
package schedule

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Validate returns nil if expr is a Fugit-compatible cron or natural-language schedule.
func Validate(expr string) error {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return fmt.Errorf("cron expression cannot be empty")
	}
	if len(expr) > 256 {
		return fmt.Errorf("cron expression is too long")
	}
	if validClassic(expr) || validNat(expr) {
		return nil
	}
	return fmt.Errorf("invalid cron (use classic like */5 * * * * or Fugit-style like every 5 minutes)")
}

func validClassic(expr string) bool {
	fields := strings.Fields(expr)
	// Optional trailing IANA / UTC timezone (Fugit allows "*/5 * * * * Asia/Shanghai")
	if len(fields) >= 6 {
		last := fields[len(fields)-1]
		if looksLikeTimezone(last) {
			fields = fields[:len(fields)-1]
		}
	}
	switch len(fields) {
	case 5:
		return validField(fields[0], 0, 59) &&
			validField(fields[1], 0, 23) &&
			validField(fields[2], 1, 31) &&
			validMonth(fields[3]) &&
			validDow(fields[4])
	case 6:
		return validField(fields[0], 0, 59) &&
			validField(fields[1], 0, 59) &&
			validField(fields[2], 0, 23) &&
			validField(fields[3], 1, 31) &&
			validMonth(fields[4]) &&
			validDow(fields[5])
	default:
		return false
	}
}

func looksLikeTimezone(s string) bool {
	if s == "UTC" || s == "Z" {
		return true
	}
	if strings.Contains(s, "/") {
		return true
	}
	if matched, _ := regexp.MatchString(`^[+-]\d{2}(:?\d{2})?$`, s); matched {
		return true
	}
	return false
}

var (
	monthNames = map[string]struct{}{
		"jan": {}, "feb": {}, "mar": {}, "apr": {}, "may": {}, "jun": {},
		"jul": {}, "aug": {}, "sep": {}, "oct": {}, "nov": {}, "dec": {},
	}
	dowNames = map[string]struct{}{
		"sun": {}, "mon": {}, "tue": {}, "wed": {}, "thu": {}, "fri": {}, "sat": {},
	}
)

func validMonth(field string) bool {
	return validNamedOrNumeric(field, 1, 12, monthNames)
}

func validDow(field string) bool {
	return validNamedOrNumeric(field, 0, 7, dowNames) // 0 and 7 = Sunday
}

func validNamedOrNumeric(field string, min, max int, names map[string]struct{}) bool {
	for _, part := range strings.Split(field, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			return false
		}
		base, step, ok := splitStep(part)
		if !ok {
			return false
		}
		if base == "*" {
			continue
		}
		lo, hi, ok := splitRange(base)
		if !ok {
			return false
		}
		if !validToken(lo, min, max, names) {
			return false
		}
		if hi != "" && !validToken(hi, min, max, names) {
			return false
		}
		_ = step
	}
	return true
}

func validField(field string, min, max int) bool {
	for _, part := range strings.Split(field, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			return false
		}
		base, step, ok := splitStep(part)
		if !ok {
			return false
		}
		if step != "" {
			if _, err := strconv.Atoi(step); err != nil {
				return false
			}
		}
		if base == "*" {
			continue
		}
		lo, hi, ok := splitRange(base)
		if !ok {
			return false
		}
		nLo, err := strconv.Atoi(lo)
		if err != nil || nLo < min || nLo > max {
			return false
		}
		if hi != "" {
			nHi, err := strconv.Atoi(hi)
			if err != nil || nHi < min || nHi > max || nHi < nLo {
				return false
			}
		}
	}
	return true
}

func splitStep(s string) (base, step string, ok bool) {
	if i := strings.IndexByte(s, '/'); i >= 0 {
		base, step = s[:i], s[i+1:]
		if base == "" || step == "" {
			return "", "", false
		}
		return base, step, true
	}
	return s, "", true
}

func splitRange(s string) (lo, hi string, ok bool) {
	if i := strings.IndexByte(s, '-'); i >= 0 {
		lo, hi = s[:i], s[i+1:]
		if lo == "" || hi == "" {
			return "", "", false
		}
		return lo, hi, true
	}
	return s, "", true
}

func validToken(tok string, min, max int, names map[string]struct{}) bool {
	if _, ok := names[strings.ToLower(tok)]; ok {
		return true
	}
	n, err := strconv.Atoi(tok)
	return err == nil && n >= min && n <= max
}

// --- Fugit::Nat subset -------------------------------------------------------

var (
	reEveryInterval = regexp.MustCompile(`(?i)^every\s+(?:(\d+)\s+)?(seconds?|secs?|s|minutes?|mins?|m|hours?|hou|h|days?|d|months?|mon)$`)
	reEveryWeekYear = regexp.MustCompile(`(?i)^every\s+(?:1\s+)?(week|year)$`)
	reEveryWeekday  = regexp.MustCompile(`(?i)^every\s+weekday$`)
	reEveryDayAt    = regexp.MustCompile(`(?i)^every\s+(day|weekday|(?:mon|tues|wednes|thurs|fri|satur|sun)day|(?:mon|tue|wed|thu|fri|sat|sun))(?:\s+or\s+(?:(?:mon|tues|wednes|thurs|fri|satur|sun)day|(?:mon|tue|wed|thu|fri|sat|sun)))*(?:\s+at\s+.+)?$`)
	reEveryAtTime   = regexp.MustCompile(`(?i)^every\s+.+\s+at\s+.+$`)
	reHasInterval   = regexp.MustCompile(`(?i)\b(second|sec|minute|min|hour|hou|day|month|week|year|weekday|noon|midnight|am|pm)s?\b|\d{1,2}:\d{2}`)
)

func validNat(expr string) bool {
	s := strings.TrimSpace(expr)
	lower := strings.ToLower(s)

	// Strip optional trailing "in|on Zone" like Fugit Nat timezone suffix.
	s = stripNatTimezone(s)
	lower = strings.ToLower(s)

	if !strings.HasPrefix(lower, "every ") && !strings.HasPrefix(lower, "every\t") {
		// Fugit Nat also allows leading "at ..." / "from ..." but Core mostly stores "every ..."
		if strings.HasPrefix(lower, "at ") || strings.HasPrefix(lower, "from ") {
			return reHasInterval.MatchString(lower)
		}
		return false
	}

	body := strings.TrimSpace(s[5:]) // after "every"
	if body == "" {
		return false
	}

	if reEveryInterval.MatchString(lower) || reEveryWeekYear.MatchString(lower) || reEveryWeekday.MatchString(lower) {
		return true
	}
	if reEveryDayAt.MatchString(lower) {
		// Require a recognizable time/interval token when "at" is present, or bare day name.
		if strings.Contains(lower, " at ") {
			return validNatTimePart(lower[strings.Index(lower, " at ")+4:])
		}
		return true
	}
	if reEveryAtTime.MatchString(lower) {
		return reHasInterval.MatchString(lower)
	}
	// "every 5 minutes starting at minute 10" and similar Fugit Nat forms.
	return looksLikeFugitNat(lower)
}

func stripNatTimezone(s string) string {
	// "every day at noon Asia/Tokyo" or "... in Asia/Tokyo"
	parts := strings.Fields(s)
	if len(parts) < 2 {
		return s
	}
	last := parts[len(parts)-1]
	if looksLikeTimezone(last) {
		return strings.Join(parts[:len(parts)-1], " ")
	}
	if len(parts) >= 3 {
		prep := strings.ToLower(parts[len(parts)-2])
		if (prep == "in" || prep == "on") && looksLikeTimezone(last) {
			return strings.Join(parts[:len(parts)-2], " ")
		}
	}
	return s
}

func validNatTimePart(part string) bool {
	part = strings.TrimSpace(strings.ToLower(part))
	if part == "" {
		return false
	}
	switch {
	case part == "noon", part == "midnight", part == "midday":
		return true
	case regexp.MustCompile(`^\d{1,2}:\d{2}(\s*(am|pm))?$`).MatchString(part):
		return true
	case regexp.MustCompile(`^\d{1,2}\s*(am|pm)$`).MatchString(part):
		return true
	case regexp.MustCompile(`^(one|two|three|four|five|six|seven|eight|nine|ten|eleven|twelve)(\s+(o'clock|thirty|fifteen))?(\s*(am|pm))?$`).MatchString(part):
		return true
	case reHasInterval.MatchString(part):
		return true
	default:
		return false
	}
}

func looksLikeFugitNat(lower string) bool {
	// Must mention a schedule unit or named day; reject "every banana".
	unit := regexp.MustCompile(`(?i)\b(\d+\s+)?(seconds?|secs?|minutes?|mins?|hours?|days?|months?|weeks?|years?|weekday|sec|min|hou|s|m|h|d)\b`)
	day := regexp.MustCompile(`(?i)\b(mon|tue|wed|thu|fri|sat|sun|monday|tuesday|wednesday|thursday|friday|saturday|sunday|day)\b`)
	return unit.MatchString(lower) || day.MatchString(lower)
}
