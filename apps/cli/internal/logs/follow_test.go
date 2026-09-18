package logs

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"beep/internal/daemon"
)

// safeBuffer wraps bytes.Buffer with a mutex to allow concurrent reads and writes.
type safeBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (sb *safeBuffer) Write(p []byte) (int, error) {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	return sb.buf.Write(p)
}

func (sb *safeBuffer) String() string {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	return sb.buf.String()
}

var _ io.Writer = (*safeBuffer)(nil)

func TestFollowReadsAppendedLines(t *testing.T) {
	ws := t.TempDir()
	if err := os.MkdirAll(filepath.Join(ws, "logs"), 0o755); err != nil {
		t.Fatal(err)
	}
	today := time.Now().Format("2006-01-02")
	path := daemon.DailyLogPath(ws, daemon.ServiceRunner, today)
	if err := os.WriteFile(path, []byte("old\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var buf safeBuffer
	done := make(chan error, 1)
	go func() {
		done <- Follow(ctx, ws, []string{daemon.ServiceRunner}, nil, time.Now, &buf, false, false)
	}()

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o600)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = f.WriteString("fresh\n")
		_ = f.Close()
		time.Sleep(250 * time.Millisecond)
		if strings.Contains(buf.String(), "runner | fresh") {
			cancel()
			<-done
			if strings.Contains(buf.String(), "old") {
				t.Fatalf("history leaked into follow: %q", buf.String())
			}
			return
		}
	}
	cancel()
	<-done
	t.Fatalf("did not see appended line, got %q", buf.String())
}

func TestFollowAcrossMidnight(t *testing.T) {
	ws := t.TempDir()
	if err := os.MkdirAll(filepath.Join(ws, "logs"), 0o755); err != nil {
		t.Fatal(err)
	}

	day1 := "2026-09-18"
	day2 := "2026-09-19"
	t1 := time.Date(2026, 9, 18, 23, 59, 59, 0, time.UTC)
	t2 := time.Date(2026, 9, 19, 0, 0, 1, 0, time.UTC)

	var currentMu sync.Mutex
	currentTime := t1
	simNow := func() time.Time {
		currentMu.Lock()
		defer currentMu.Unlock()
		return currentTime
	}

	path1 := daemon.DailyLogPath(ws, daemon.ServiceRunner, day1)
	if err := os.WriteFile(path1, []byte("day1 existing\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var buf safeBuffer
	done := make(chan error, 1)
	go func() {
		done <- Follow(ctx, ws, []string{daemon.ServiceRunner}, nil, simNow, &buf, false, false)
	}()

	// Wait for initial tick to finish and skip day1
	time.Sleep(250 * time.Millisecond)

	// Advance time to day 2 and write lines before and after
	path2 := daemon.DailyLogPath(ws, daemon.ServiceRunner, day2)
	if err := os.WriteFile(path2, []byte("day2 early morning line\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	currentMu.Lock()
	currentTime = t2
	currentMu.Unlock()

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if strings.Contains(buf.String(), "runner | day2 early morning line") {
			cancel()
			<-done
			if strings.Contains(buf.String(), "day1 existing") {
				t.Fatalf("day1 leaked into follow: %q", buf.String())
			}
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	cancel()
	<-done
	t.Fatalf("did not see midnight rolled-over lines, got %q", buf.String())
}
