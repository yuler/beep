package logview

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
