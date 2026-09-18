package logview

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

	got, err = ParseInstant("12h", now)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Equal(now.Add(-12 * time.Hour)) {
		t.Fatalf("12h: got %v", got)
	}

	got, err = ParseInstant("30m", now)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Equal(now.Add(-30 * time.Minute)) {
		t.Fatalf("30m: got %v", got)
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
