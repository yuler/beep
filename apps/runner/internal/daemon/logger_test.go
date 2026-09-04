package daemon

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDailyLogWriter(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "beep-log-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	w, err := NewDailyLogWriter(tempDir, "test-runner")
	if err != nil {
		t.Fatalf("failed to create DailyLogWriter: %v", err)
	}
	defer w.Close()

	// Write message with ANSI colors
	coloredMsg := "\x1b[32m[OK]\x1b[0m Job finished successfully\n"
	n, err := w.Write([]byte(coloredMsg))
	if err != nil {
		t.Fatalf("unexpected write error: %v", err)
	}
	if n != len(coloredMsg) {
		t.Fatalf("expected write len %d, got %d", len(coloredMsg), n)
	}

	today := time.Now().Format("2006-01-02")
	expectedFile := filepath.Join(tempDir, "test-runner-"+today+".log")

	data, err := os.ReadFile(expectedFile)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	content := string(data)
	if strings.Contains(content, "\x1b[32m") {
		t.Errorf("expected ANSI codes to be stripped, got: %q", content)
	}
	if !strings.Contains(content, "[OK] Job finished successfully\n") {
		t.Errorf("expected clean message in log, got: %q", content)
	}

	info, err := os.Stat(expectedFile)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("expected log file mode 0600, got %o", info.Mode().Perm())
	}
}
