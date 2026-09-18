package logs

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"beep/internal/daemon"
)

func TestParseSinceRelativeAndDate(t *testing.T) {
	now := time.Date(2026, 9, 18, 15, 4, 0, 0, time.UTC)

	got, err := ParseInstant("2d", now)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Equal(now.Add(-48 * time.Hour)) {
		t.Fatalf("2d: got %v", got)
	}

	got, err = ParseInstant("2026-09-17", now)
	if err != nil {
		t.Fatal(err)
	}
	want := time.Date(2026, 9, 17, 0, 0, 0, 0, now.Location())
	if !got.Equal(want) {
		t.Fatalf("date: got %v want %v", got, want)
	}
}

func TestParseSinceRejectsHoursAndMinutes(t *testing.T) {
	now := time.Now()
	for _, v := range []string{"12h", "30m", "1h", "60m"} {
		_, err := ParseInstant(v, now)
		if err == nil {
			t.Fatalf("expected %q to be rejected", v)
		}
		if !strings.Contains(err.Error(), "hours and minutes (h/m) are not supported") {
			t.Fatalf("expected h/m rejection message, got: %v", err)
		}
	}
}

func TestParseSinceRejectsGarbage(t *testing.T) {
	_, err := ParseInstant("yesterday", time.Now())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDaysCoveringRange(t *testing.T) {
	now := time.Date(2026, 9, 18, 4, 0, 0, 0, time.UTC)
	from := now.Add(-12 * time.Hour) // previous calendar day
	days := DaysCovering(from, now)
	if len(days) != 2 || days[0] != "2026-09-17" || days[1] != "2026-09-18" {
		t.Fatalf("got %v", days)
	}
}

func TestParseServices(t *testing.T) {
	got, err := ParseServices(nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0] != daemon.ServiceRunner || got[1] != daemon.ServiceChannel {
		t.Fatalf("default: %v", got)
	}

	got, err = ParseServices([]string{"channel,runner", "channel"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("dedupe: %v", got)
	}

	_, err = ParseServices([]string{"service"})
	if err == nil {
		t.Fatal("expected unknown service error")
	}
}

func TestMatchGrep(t *testing.T) {
	if !MatchGrep("Hello ERROR there", []string{"error"}) {
		t.Fatal("expected case-insensitive match")
	}
	if MatchGrep("ok", []string{"error"}) {
		t.Fatal("unexpected match")
	}
	if !MatchGrep("foo", []string{"bar", "FOO"}) {
		t.Fatal("expected OR")
	}
	if !MatchGrep("anything", nil) {
		t.Fatal("empty patterns should keep all lines")
	}
}

func TestShouldFollow(t *testing.T) {
	if !ShouldFollow(false, false, true, false) {
		t.Fatal("TTY should follow by default")
	}
	if ShouldFollow(false, false, false, false) {
		t.Fatal("pipe should not follow by default")
	}
	if !ShouldFollow(true, false, false, false) {
		t.Fatal("--follow should follow even when piped")
	}
	if ShouldFollow(false, true, true, false) {
		t.Fatal("--no-follow should win on TTY")
	}
	if ShouldFollow(true, false, true, true) {
		t.Fatal("until in the past should not follow")
	}
}

func TestCollectFilesIgnoresUnprefixed(t *testing.T) {
	ws := t.TempDir()
	logs := filepath.Join(ws, "logs")
	if err := os.MkdirAll(logs, 0o755); err != nil {
		t.Fatal(err)
	}
	today := "2026-09-18"
	unprefixed := filepath.Join(logs, today+".log")
	runner := daemon.DailyLogPath(ws, daemon.ServiceRunner, today)
	if err := os.WriteFile(unprefixed, []byte("legacy\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(runner, []byte("runner-line\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	files := CollectFiles(ws, []string{daemon.ServiceRunner, daemon.ServiceChannel}, []string{today})
	if len(files) != 1 || files[0] != runner {
		t.Fatalf("got %v", files)
	}
}

func TestHistoryFiltersThenTails(t *testing.T) {
	ws := t.TempDir()
	today := "2026-09-18"
	if err := os.MkdirAll(filepath.Join(ws, "logs"), 0o755); err != nil {
		t.Fatal(err)
	}
	runner := daemon.DailyLogPath(ws, daemon.ServiceRunner, today)
	channel := daemon.DailyLogPath(ws, daemon.ServiceChannel, today)
	if err := os.WriteFile(runner, []byte("keep one\nskip\nkeep two\nkeep three\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(channel, []byte("keep four\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	lines, err := History(ws, []string{daemon.ServiceRunner, daemon.ServiceChannel}, []string{today}, []string{"keep"}, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 2 {
		t.Fatalf("len=%d %#v", len(lines), lines)
	}
	if lines[0].Text != "keep three" || lines[0].Source != daemon.ServiceRunner {
		t.Fatalf("line0: %#v", lines[0])
	}
	if lines[1].Text != "keep four" || lines[1].Source != daemon.ServiceChannel {
		t.Fatalf("line1: %#v", lines[1])
	}
}

func TestHistoryMultiDayReverseTail(t *testing.T) {
	ws := t.TempDir()
	day1 := "2026-09-17"
	day2 := "2026-09-18"
	if err := os.MkdirAll(filepath.Join(ws, "logs"), 0o755); err != nil {
		t.Fatal(err)
	}

	d1Runner := daemon.DailyLogPath(ws, daemon.ServiceRunner, day1)
	d2Runner := daemon.DailyLogPath(ws, daemon.ServiceRunner, day2)
	d2Channel := daemon.DailyLogPath(ws, daemon.ServiceChannel, day2)

	if err := os.WriteFile(d1Runner, []byte("d1-line1\nd1-line2\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(d2Runner, []byte("d2-runner-1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(d2Channel, []byte("d2-chan-1\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	// Requesting 2 lines should only take from day 2 and keep chronological order
	lines, err := History(ws, []string{daemon.ServiceRunner, daemon.ServiceChannel}, []string{day1, day2}, nil, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %#v", len(lines), lines)
	}
	if lines[0].Text != "d2-runner-1" || lines[0].Source != daemon.ServiceRunner {
		t.Errorf("line0 got %#v", lines[0])
	}
	if lines[1].Text != "d2-chan-1" || lines[1].Source != daemon.ServiceChannel {
		t.Errorf("line1 got %#v", lines[1])
	}

	// Requesting 3 lines should reach back to day 1 for 1 line
	lines3, err := History(ws, []string{daemon.ServiceRunner, daemon.ServiceChannel}, []string{day1, day2}, nil, 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines3) != 3 {
		t.Fatalf("expected 3 lines, got %d: %#v", len(lines3), lines3)
	}
	if lines3[0].Text != "d1-line2" || lines3[1].Text != "d2-runner-1" || lines3[2].Text != "d2-chan-1" {
		t.Errorf("lines3 got %#v", lines3)
	}
}

func TestHistoryLongLine(t *testing.T) {
	ws := t.TempDir()
	today := "2026-09-18"
	if err := os.MkdirAll(filepath.Join(ws, "logs"), 0o755); err != nil {
		t.Fatal(err)
	}
	runner := daemon.DailyLogPath(ws, daemon.ServiceRunner, today)
	longText := strings.Repeat("x", 128*1024)
	if err := os.WriteFile(runner, []byte(longText+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	lines, err := History(ws, []string{daemon.ServiceRunner}, []string{today}, nil, 10)
	if err != nil {
		t.Fatalf("History failed on 128KiB line: %v", err)
	}
	if len(lines) != 1 || lines[0].Text != longText {
		t.Fatalf("unexpected line content, len=%d", len(lines))
	}
}

func TestFormatTextAndJSON(t *testing.T) {
	line := Line{Source: "runner", Text: "hello"}
	if got := FormatLine(line, false, false); got != "runner | hello" {
		t.Fatalf("text: %q", got)
	}
	raw := FormatLine(line, true, false)
	var obj map[string]string
	if err := json.Unmarshal([]byte(raw), &obj); err != nil {
		t.Fatal(err)
	}
	if obj["source"] != "runner" || obj["text"] != "hello" {
		t.Fatalf("json: %v", obj)
	}
	if strings.Contains(raw, "runner |") {
		t.Fatal("json should not include SOURCE | prefix")
	}
}

func TestEmptyHistory(t *testing.T) {
	ws := t.TempDir()
	lines, err := History(ws, []string{daemon.ServiceRunner}, []string{"2026-09-18"}, nil, 100)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 0 {
		t.Fatalf("got %v", lines)
	}
}

func TestWriteLines(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteLines(&buf, []Line{{Source: "channel", Text: "x"}}, false, false); err != nil {
		t.Fatal(err)
	}
	if buf.String() != "channel | x\n" {
		t.Fatalf("got %q", buf.String())
	}
}
