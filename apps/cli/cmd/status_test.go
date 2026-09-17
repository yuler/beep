package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"beep/cmd/service"
	"beep/internal/config"
	"beep/internal/ui"
)

func captureOutput(f func()) string {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}

func TestServiceStatusOutput(t *testing.T) {
	// Disable color for simpler assertions
	ui.SetEnabled(false)
	defer ui.SetEnabled(true)

	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.json")
	_ = config.SaveFile(cfgPath, &config.FileConfig{
		ServerURL:   "http://example.com",
		AccessToken: "beep_pat_secrettoken123456",
		UserEmail:   "alice@example.com",
		UserName:    "Alice Smith",
		AccountSlug: "alice-team",
	})

	flagWorkspace = tmpDir
	defer func() { flagWorkspace = "" }()

	cmd := service.NewCmdStatus()
	out := captureOutput(func() {
		_ = cmd.RunE(cmd, nil)
	})

	// Service status should display runner and channel service sections
	if !strings.Contains(out, "Runner Service:") {
		t.Errorf("expected output to contain 'Runner Service:', got:\n%s", out)
	}
	if !strings.Contains(out, "Channel Service:") {
		t.Errorf("expected output to contain 'Channel Service:', got:\n%s", out)
	}

	// Service status should not display Beep Status or auth/user details
	if strings.Contains(out, "Beep Status:") {
		t.Errorf("expected output not to contain 'Beep Status:', got:\n%s", out)
	}
	if strings.Contains(out, "Auth:") {
		t.Errorf("expected output not to contain 'Auth:', got:\n%s", out)
	}
	if strings.Contains(out, "Alice Smith") {
		t.Errorf("expected output not to contain user name, got:\n%s", out)
	}
	if strings.Contains(out, "alice-team") {
		t.Errorf("expected output not to contain account slug, got:\n%s", out)
	}
}
