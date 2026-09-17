package cmd

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"beep/internal/config"

	"github.com/spf13/cobra"
)

func mustFindCmd(t *testing.T, path ...string) *cobra.Command {
	t.Helper()
	cmd, _, err := RootCmd.Find(path)
	if err != nil {
		t.Fatalf("failed to find %q: %v", strings.Join(path, " "), err)
	}
	return cmd
}

func setupCLITestEnv(t *testing.T, handler http.HandlerFunc) (string, func()) {
	server := httptest.NewServer(handler)
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	fc := &config.FileConfig{
		ServerURL:   server.URL,
		AccessToken: "beep_at_test_token",
	}
	if err := config.SaveFile(configPath, fc); err != nil {
		t.Fatalf("failed to save test config: %v", err)
	}

	oldWs := flagWorkspace
	oldServer := flagServer
	oldAccount := flagAccount
	oldJSON := flagJSON

	flagWorkspace = tmpDir
	flagServer = server.URL
	flagAccount = ""
	flagJSON = false

	cleanup := func() {
		server.Close()
		flagWorkspace = oldWs
		flagServer = oldServer
		flagAccount = oldAccount
		flagJSON = oldJSON
	}
	return server.URL, cleanup
}

func captureStdout(fn func() error) (string, error) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := fn()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String(), err
}
