package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

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

func TestStatusAuthDisplay(t *testing.T) {
	// Disable color for simpler assertions
	ui.SetEnabled(false)
	defer ui.SetEnabled(true)

	t.Run("not logged in", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "beep-status-not-logged-in-*")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		cfgPath := filepath.Join(tmpDir, "config.json")
		_ = config.SaveFile(cfgPath, &config.FileConfig{
			ServerURL: "http://example.com",
		})

		flagWorkspace = tmpDir
		defer func() { flagWorkspace = "" }()

		cmd := newStatusCmd()
		out := captureOutput(func() {
			_ = cmd.RunE(cmd, nil)
		})

		if !strings.Contains(out, "Server:          http://example.com") {
			t.Errorf("expected output to contain Server, got:\n%s", out)
		}
		if !strings.Contains(out, "Auth:            not logged in ○") {
			t.Errorf("expected output to contain 'Auth: not logged in ○', got:\n%s", out)
		}
		if strings.Contains(out, "Token:           beep_pat") {
			t.Errorf("expected output not to contain access token, got:\n%s", out)
		}
	})

	t.Run("logged in with full details", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "beep-status-logged-in-*")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

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

		cmd := newStatusCmd()
		out := captureOutput(func() {
			_ = cmd.RunE(cmd, nil)
		})

		if !strings.Contains(out, "Server:          http://example.com") {
			t.Errorf("expected output to contain Server, got:\n%s", out)
		}
		if !strings.Contains(out, "Auth:            logged in ●") {
			t.Errorf("expected output to contain 'Auth: logged in ●', got:\n%s", out)
		}
		if !strings.Contains(out, "User:            Alice Smith (alice@example.com)") {
			t.Errorf("expected output to contain User info, got:\n%s", out)
		}
		if !strings.Contains(out, "Account:         alice-team") {
			t.Errorf("expected output to contain Account slug, got:\n%s", out)
		}
		if !strings.Contains(out, "Token:           beep_pat••••••••") {
			t.Errorf("expected output to contain masked Token, got:\n%s", out)
		}

		// Verify ordering: Server should appear before Auth, Auth before User, etc.
		serverIdx := strings.Index(out, "Server:")
		authIdx := strings.Index(out, "Auth:")
		userIdx := strings.Index(out, "User:")
		accountIdx := strings.Index(out, "Account:")
		tokenIdx := strings.Index(out, "Token:")

		if !(serverIdx < authIdx && authIdx < userIdx && userIdx < accountIdx && accountIdx < tokenIdx) {
			t.Errorf("expected ordering Server < Auth < User < Account < Token, indices: server=%d auth=%d user=%d account=%d token=%d",
				serverIdx, authIdx, userIdx, accountIdx, tokenIdx)
		}
	})

	t.Run("logged in with email only and no account", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "beep-status-email-only-*")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		cfgPath := filepath.Join(tmpDir, "config.json")
		_ = config.SaveFile(cfgPath, &config.FileConfig{
			ServerURL:   "http://example.com",
			AccessToken: "beep_pat_9876543210token",
			UserEmail:   "bob@example.com",
		})

		flagWorkspace = tmpDir
		defer func() { flagWorkspace = "" }()

		cmd := newStatusCmd()
		out := captureOutput(func() {
			_ = cmd.RunE(cmd, nil)
		})

		if !strings.Contains(out, "Auth:            logged in ●") {
			t.Errorf("expected output to contain 'Auth: logged in ●', got:\n%s", out)
		}
		if !strings.Contains(out, "User:            bob@example.com") {
			t.Errorf("expected output to contain 'User: bob@example.com', got:\n%s", out)
		}
		if strings.Contains(out, "Account:") {
			t.Errorf("expected output not to contain Account when not configured, got:\n%s", out)
		}
		if !strings.Contains(out, "Token:           beep_pat••••••••") {
			t.Errorf("expected output to contain masked Token, got:\n%s", out)
		}
	})
}
