package logview

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"beep/internal/daemon"
	"beep/internal/ui"
)

var relativeRe = regexp.MustCompile(`^([0-9]+)([dhm])$`)

// Line is one daemon log line with its file source.
type Line struct {
	Source string `json:"source"`
	Text   string `json:"text"`
}

// ParseInstant parses a relative duration (2d, 12h, 30m) or YYYY-MM-DD as the start of that day.
func ParseInstant(s string, now time.Time) (time.Time, error) {
	return parseInstant(s, now, false)
}

// ParseUntil parses --until. A YYYY-MM-DD value is the end of that day.
func ParseUntil(s string, now time.Time) (time.Time, error) {
	return parseInstant(s, now, true)
}

func parseInstant(s string, now time.Time, untilDate bool) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, fmt.Errorf("empty time")
	}
	if m := relativeRe.FindStringSubmatch(s); m != nil {
		n := 0
		for _, c := range m[1] {
			n = n*10 + int(c-'0')
		}
		var d time.Duration
		switch m[2] {
		case "d":
			d = time.Duration(n) * 24 * time.Hour
		case "h":
			d = time.Duration(n) * time.Hour
		case "m":
			d = time.Duration(n) * time.Minute
		}
		return now.Add(-d), nil
	}
	day, err := time.ParseInLocation("2006-01-02", s, now.Location())
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid time %q (use 2d, 12h, 30m, or YYYY-MM-DD)", s)
	}
	if untilDate {
		return time.Date(day.Year(), day.Month(), day.Day(), 23, 59, 59, 1e9-1, day.Location()), nil
	}
	return day, nil
}

// StartOfDay returns local midnight for now.
func StartOfDay(now time.Time) time.Time {
	y, m, d := now.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, now.Location())
}

// DaysCovering lists YYYY-MM-DD dates from from through until, inclusive.
func DaysCovering(from, until time.Time) []string {
	if until.Before(from) {
		from, until = until, from
	}
	start := StartOfDay(from)
	end := StartOfDay(until)
	var days []string
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		days = append(days, d.Format("2006-01-02"))
	}
	return days
}

// ParseServices accepts repeated flags and comma lists. Default is runner then channel.
func ParseServices(vals []string) ([]string, error) {
	if len(vals) == 0 {
		return []string{daemon.ServiceRunner, daemon.ServiceChannel}, nil
	}
	var out []string
	seen := map[string]bool{}
	for _, v := range vals {
		for _, part := range strings.Split(v, ",") {
			part = strings.ToLower(strings.TrimSpace(part))
			if part == "" {
				continue
			}
			if part != daemon.ServiceRunner && part != daemon.ServiceChannel {
				return nil, fmt.Errorf("unknown service %q (expected runner or channel)", part)
			}
			if !seen[part] {
				seen[part] = true
				out = append(out, part)
			}
		}
	}
	if len(out) == 0 {
		return []string{daemon.ServiceRunner, daemon.ServiceChannel}, nil
	}
	return out, nil
}

// MatchGrep reports whether line matches any pattern (case-insensitive substring). Empty patterns match all.
func MatchGrep(line string, patterns []string) bool {
	if len(patterns) == 0 {
		return true
	}
	lower := strings.ToLower(line)
	for _, p := range patterns {
		if p == "" {
			continue
		}
		if strings.Contains(lower, strings.ToLower(p)) {
			return true
		}
	}
	return false
}

// ShouldFollow applies TTY heuristics and flag overrides.
func ShouldFollow(follow, noFollow, isTTY, untilInPast bool) bool {
	if untilInPast || noFollow {
		return false
	}
	if follow {
		return true
	}
	return isTTY
}

// CollectFiles returns existing prefixed log paths for the given services and days.
func CollectFiles(workspace string, services, days []string) []string {
	var files []string
	for _, day := range days {
		for _, service := range services {
			path := daemon.DailyLogPath(workspace, service, day)
			if _, err := os.Stat(path); err == nil {
				files = append(files, path)
			}
		}
	}
	return files
}

// History reads matching files in day/service order, greps, then keeps the last n lines.
func History(workspace string, services, days, patterns []string, n int) ([]Line, error) {
	if n < 0 {
		n = 0
	}
	var matched []Line
	for _, day := range days {
		for _, service := range services {
			path := daemon.DailyLogPath(workspace, service, day)
			lines, err := readFileLines(path, service, patterns)
			if err != nil {
				if os.IsNotExist(err) {
					continue
				}
				return nil, err
			}
			matched = append(matched, lines...)
		}
	}
	if n == 0 || len(matched) <= n {
		return matched, nil
	}
	return matched[len(matched)-n:], nil
}

func readFileLines(path, source string, patterns []string) ([]Line, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []Line
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		text := scanner.Text()
		if MatchGrep(text, patterns) {
			out = append(out, Line{Source: source, Text: text})
		}
	}
	return out, scanner.Err()
}

// FormatLine renders one log line as text or NDJSON. No trailing newline.
func FormatLine(line Line, asJSON, color bool) string {
	if asJSON {
		b, err := json.Marshal(line)
		if err != nil {
			return ""
		}
		return string(b)
	}
	label := line.Source
	if color {
		switch line.Source {
		case daemon.ServiceRunner:
			label = ui.Cyan(line.Source)
		case daemon.ServiceChannel:
			label = ui.Magenta(line.Source)
		}
	}
	return label + " | " + line.Text
}

// WriteLines writes formatted lines with trailing newlines.
func WriteLines(w io.Writer, lines []Line, asJSON, color bool) error {
	for _, line := range lines {
		if _, err := fmt.Fprintln(w, FormatLine(line, asJSON, color)); err != nil {
			return err
		}
	}
	return nil
}

// LogsDir is the workspace logs directory.
func LogsDir(workspace string) string {
	return filepath.Join(workspace, "logs")
}
