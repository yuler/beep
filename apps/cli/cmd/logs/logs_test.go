package logs

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"beep/internal/cmdutil"
	"beep/internal/config"
	"beep/internal/daemon"
	"beep/internal/ui"
)

func TestLogsUnknownService(t *testing.T) {
	ws := t.TempDir()
	cmdutil.SetOverrideWorkspace(ws)
	t.Cleanup(func() { cmdutil.SetOverrideWorkspace("") })
	fc := &config.FileConfig{ServerURL: "http://localhost:3000"}
	if err := config.SaveFile(filepath.Join(ws, "config.json"), fc); err != nil {
		t.Fatal(err)
	}

	cmd := NewCmdLogs()
	cmd.SetArgs([]string{"--service", "service", "--no-follow"})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "unknown service") {
		t.Fatalf("expected unknown service error, got %v", err)
	}
}

func TestLogsDumpsHistoryWithoutFollow(t *testing.T) {
	ui.SetEnabled(false)
	ws := t.TempDir()
	cmdutil.SetOverrideWorkspace(ws)
	t.Cleanup(func() { cmdutil.SetOverrideWorkspace("") })

	if err := os.MkdirAll(filepath.Join(ws, "logs"), 0o755); err != nil {
		t.Fatal(err)
	}
	fc := &config.FileConfig{ServerURL: "http://localhost:3000"}
	if err := config.SaveFile(filepath.Join(ws, "config.json"), fc); err != nil {
		t.Fatal(err)
	}

	path := daemon.DailyLogPath(ws, daemon.ServiceRunner, timeNow().Format("2006-01-02"))
	if err := os.WriteFile(path, []byte("hello from runner\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	cmd := NewCmdLogs()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--no-follow"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "runner | hello from runner") {
		t.Fatalf("got %q", buf.String())
	}
}
