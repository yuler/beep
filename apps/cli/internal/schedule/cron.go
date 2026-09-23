// Package schedule validates cron expressions compatible with Rails Fugit
// (classic 5/6-field cron and Fugit::Nat semantic phrases like "every 5 minutes").
package schedule

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
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

var (
	monthToNum = map[string]int{
		"jan": 1, "feb": 2, "mar": 3, "apr": 4, "may": 5, "jun": 6,
		"jul": 7, "aug": 8, "sep": 9, "oct": 10, "nov": 11, "dec": 12,
	}
	dowToNum = map[string]int{
		"sun": 0, "mon": 1, "tue": 2, "wed": 3, "thu": 4, "fri": 5, "sat": 6,
	}
)

func parseCronToken(tok string, isDow, isMonth bool) (int, error) {
	tok = strings.ToLower(strings.TrimSpace(tok))
	if isMonth {
		if n, ok := monthToNum[tok]; ok {
			return n, nil
		}
	}
	if isDow {
		if n, ok := dowToNum[tok]; ok {
			return n, nil
		}
	}
	return strconv.Atoi(tok)
}

// MatchField reports whether val matches field specification (supports *, comma, ranges, steps).
func MatchField(field string, val int, min, max int, isDow, isMonth bool) bool {
	if field == "*" {
		return true
	}
	for _, part := range strings.Split(field, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		base, stepStr, _ := splitStep(part)
		step := 1
		if stepStr != "" {
			var err error
			step, err = strconv.Atoi(stepStr)
			if err != nil || step <= 0 {
				continue
			}
		}
		if base == "*" {
			if (val-min)%step == 0 {
				return true
			}
			continue
		}
		loStr, hiStr, _ := splitRange(base)
		lo, err1 := parseCronToken(loStr, isDow, isMonth)
		if err1 != nil {
			continue
		}
		hi := lo
		if hiStr != "" {
			var err2 error
			hi, err2 = parseCronToken(hiStr, isDow, isMonth)
			if err2 != nil {
				continue
			}
		}
		v := val
		if isDow {
			if lo == 7 {
				lo = 0
			}
			if hi == 7 {
				hi = 0
			}
			if v == 7 {
				v = 0
			}
		}
		if lo <= hi {
			if v >= lo && v <= hi && (v-lo)%step == 0 {
				return true
			}
		} else {
			if (v >= lo || v <= hi) {
				return true
			}
		}
	}
	return false
}

func stripClassicTimezone(expr string) string {
	fields := strings.Fields(expr)
	if len(fields) >= 6 {
		last := fields[len(fields)-1]
		if looksLikeTimezone(last) {
			return strings.Join(fields[:len(fields)-1], " ")
		}
	}
	return expr
}

// NextRun calculates the next run time after from in timezone loc.
func NextRun(expr string, from time.Time, loc *time.Location) (time.Time, bool) {
	if loc == nil {
		loc = time.Local
	}
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return time.Time{}, false
	}

	clean := stripClassicTimezone(expr)
	fields := strings.Fields(clean)
	if len(fields) == 5 && validClassic(clean) {
		return nextRunClassic5(fields, from, loc)
	}

	return nextRunNat(expr, from, loc)
}

func nextRunClassic5(fields []string, from time.Time, loc *time.Location) (time.Time, bool) {
	t := from.In(loc).Truncate(time.Minute).Add(time.Minute)
	// Iterate through candidate times, advancing coarse units when they do not match
	for i := 0; i < 5*366*24*60; i++ {
		// 1. Month
		if !MatchField(fields[3], int(t.Month()), 1, 12, false, true) {
			t = time.Date(t.Year(), t.Month()+1, 1, 0, 0, 0, 0, loc)
			continue
		}
		// 2. Day of Month & Day of Week
		domStar := fields[2] == "*"
		dowStar := fields[4] == "*"
		domMatch := MatchField(fields[2], t.Day(), 1, 31, false, false)
		dowMatch := MatchField(fields[4], int(t.Weekday()), 0, 6, true, false)
		dayMatches := false
		if domStar && dowStar {
			dayMatches = true
		} else if !domStar && !dowStar {
			dayMatches = domMatch || dowMatch
		} else if !domStar {
			dayMatches = domMatch
		} else {
			dayMatches = dowMatch
		}
		if !dayMatches {
			t = time.Date(t.Year(), t.Month(), t.Day()+1, 0, 0, 0, 0, loc)
			continue
		}
		// 3. Hour
		if !MatchField(fields[1], t.Hour(), 0, 23, false, false) {
			t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour()+1, 0, 0, 0, loc)
			continue
		}
		// 4. Minute
		if !MatchField(fields[0], t.Minute(), 0, 59, false, false) {
			t = t.Add(time.Minute)
			continue
		}
		return t, true
	}
	return time.Time{}, false
}

func nextRunNat(expr string, from time.Time, loc *time.Location) (time.Time, bool) {
	clean := stripNatTimezone(expr)
	lower := strings.ToLower(strings.TrimSpace(clean))

	m := regexp.MustCompile(`(?i)^every\s+(\d+)\s*(s|sec|secs|second|seconds|m|min|mins|minute|minutes|h|hr|hrs|hour|hours|d|day|days)$`).FindStringSubmatch(lower)
	if len(m) == 3 {
		n, err := strconv.Atoi(m[1])
		if err == nil && n > 0 {
			unit := m[2]
			switch {
			case strings.HasPrefix(unit, "s"):
				return from.In(loc).Add(time.Duration(n) * time.Second), true
			case strings.HasPrefix(unit, "m"):
				return from.In(loc).Add(time.Duration(n) * time.Minute), true
			case strings.HasPrefix(unit, "h"):
				return from.In(loc).Add(time.Duration(n) * time.Hour), true
			case strings.HasPrefix(unit, "d"):
				return from.In(loc).Add(time.Duration(n) * 24 * time.Hour), true
			}
		}
	}
	if lower == "every minute" {
		return from.In(loc).Add(time.Minute), true
	}
	if lower == "every hour" {
		return from.In(loc).Add(time.Hour), true
	}
	if lower == "every day" {
		return from.In(loc).Add(24 * time.Hour), true
	}

	if lower == "every day at noon" {
		target := time.Date(from.Year(), from.Month(), from.Day(), 12, 0, 0, 0, loc)
		if !target.After(from.In(loc)) {
			target = target.AddDate(0, 0, 1)
		}
		return target, true
	}
	if lower == "every day at midnight" {
		target := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, 1)
		return target, true
	}

	timeMatch := regexp.MustCompile(`(?i)^every\s+day\s+at\s+(\d{1,2}):(\d{2})$`).FindStringSubmatch(lower)
	if len(timeMatch) == 3 {
		h, _ := strconv.Atoi(timeMatch[1])
		min, _ := strconv.Atoi(timeMatch[2])
		target := time.Date(from.Year(), from.Month(), from.Day(), h, min, 0, 0, loc)
		if !target.After(from.In(loc)) {
			target = target.AddDate(0, 0, 1)
		}
		return target, true
	}

	return time.Time{}, false
}

// Describe returns a natural language description for the cron expression.
func Describe(expr string) string {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return ""
	}
	lower := strings.ToLower(expr)
	if strings.HasPrefix(lower, "every ") {
		return strings.ToUpper(expr[:1]) + expr[1:]
	}

	clean := stripClassicTimezone(expr)
	fields := strings.Fields(clean)
	if len(fields) != 5 {
		return ""
	}

	m, h, dom, mon, dow := fields[0], fields[1], fields[2], fields[3], fields[4]

	// 1. Every N minutes: "*/5 * * * *"
	if strings.HasPrefix(m, "*/") && h == "*" && dom == "*" && mon == "*" && dow == "*" {
		step := m[2:]
		return fmt.Sprintf("Every %s minutes", step)
	}

	// 2. Hourly: "0 * * * *"
	if m == "0" && h == "*" && dom == "*" && mon == "*" && dow == "*" {
		return "Every hour"
	}
	if !strings.ContainsAny(m, "*/-,") && h == "*" && dom == "*" && mon == "*" && dow == "*" {
		minNum, err := strconv.Atoi(m)
		if err == nil {
			return fmt.Sprintf("Every hour at minute %d", minNum)
		}
	}

	// 3. Daily / Specific time: m and h are numbers, dom and mon are *
	minNum, errM := strconv.Atoi(m)
	hourNum, errH := strconv.Atoi(h)
	if errM == nil && errH == nil && dom == "*" && mon == "*" {
		timeStr := fmt.Sprintf("%02d:%02d", hourNum, minNum)
		switch dow {
		case "*":
			if hourNum == 0 && minNum == 0 {
				return "Every day at midnight (00:00)"
			}
			if hourNum == 12 && minNum == 0 {
				return "Every day at noon (12:00)"
			}
			return fmt.Sprintf("Every day at %s", timeStr)
		case "1-5":
			return fmt.Sprintf("Every weekday at %s", timeStr)
		case "0,6", "6,0", "7,6", "6,7":
			return fmt.Sprintf("Every weekend at %s", timeStr)
		case "1", "mon", "Mon":
			return fmt.Sprintf("Every Monday at %s", timeStr)
		case "2", "tue", "Tue":
			return fmt.Sprintf("Every Tuesday at %s", timeStr)
		case "3", "wed", "Wed":
			return fmt.Sprintf("Every Wednesday at %s", timeStr)
		case "4", "thu", "Thu":
			return fmt.Sprintf("Every Thursday at %s", timeStr)
		case "5", "fri", "Fri":
			return fmt.Sprintf("Every Friday at %s", timeStr)
		case "6", "sat", "Sat":
			return fmt.Sprintf("Every Saturday at %s", timeStr)
		case "0", "7", "sun", "Sun":
			return fmt.Sprintf("Every Sunday at %s", timeStr)
		}
	}

	return ""
}

