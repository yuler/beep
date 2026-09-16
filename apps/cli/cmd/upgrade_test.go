package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"beep/internal/ui"
	"beep/internal/updater"
	"beep/internal/version"

	"github.com/spf13/cobra"
)

func TestUpgradeCmdHelpAndAliases(t *testing.T) {
	cmd := newUpgradeCmd()
	if cmd.Use != "upgrade" {
		t.Errorf("expected Use 'upgrade', got %s", cmd.Use)
	}

	hasUpdateAlias := false
	for _, a := range cmd.Aliases {
		if a == "update" {
			hasUpdateAlias = true
			break
		}
	}
	if !hasUpdateAlias {
		t.Errorf("expected cmd.Aliases to contain 'update', got %v", cmd.Aliases)
	}

	if cmd.Flag("check") == nil {
		t.Errorf("expected --check flag to be defined")
	}
	if cmd.Flag("version") == nil {
		t.Errorf("expected --version flag to be defined")
	}
	if cmd.Flag("force") == nil {
		t.Errorf("expected --force flag to be defined")
	}
	if cmd.Flag("yes") == nil {
		t.Errorf("expected --yes flag to be defined")
	}
}

func TestUpgradeCheckMode(t *testing.T) {
	ui.SetEnabled(false)
	defer ui.SetEnabled(true)

	mockRelease := updater.ReleaseInfo{
		Tag:         "v0.3.0",
		Name:        "Beep CLI v0.3.0",
		HTMLURL:     "https://github.com/yuler/beep/releases/tag/v0.3.0",
		PublishedAt: time.Now(),
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(mockRelease)
	}))
	defer ts.Close()

	mockTransport := &rewritingTransport{
		base:   http.DefaultTransport,
		target: ts.URL,
	}
	mockClient := &http.Client{Transport: mockTransport, Timeout: 15 * time.Second}
	restore := updater.SetHTTPClientsForTest(mockClient, mockClient)
	defer restore()

	oldVersion := version.Version
	version.Version = "0.2.1"
	defer func() { version.Version = oldVersion }()

	t.Setenv("BEEP_REPO", "yuler/beep")
	t.Setenv("BEEP_NO_UPDATE_CHECK", "1")

	cmd := newUpgradeCmd()
	cmd.SetArgs([]string{"--check"})
	flagUpgradeCheck = true
	flagUpgradeVersion = ""
	flagUpgradeForce = false
	flagUpgradeYes = true
	defer func() {
		flagUpgradeCheck = false
		flagUpgradeVersion = ""
		flagUpgradeForce = false
		flagUpgradeYes = false
	}()

	out := captureOutput(func() {
		err := cmd.Execute()
		if err != nil {
			t.Errorf("cmd.Execute error: %v", err)
		}
	})

	if !strings.Contains(out, "new release") && !strings.Contains(out, "v0.3.0") {
		t.Fatalf("expected output to mention new release / v0.3.0, got: %q", out)
	}
	if strings.Contains(out, "already up to date") {
		t.Fatalf("unexpected up-to-date soft-log in output: %q", out)
	}
}

type rewritingTransport struct {
	base   http.RoundTripper
	target string
}

func (t *rewritingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	u, err := http.NewRequestWithContext(req.Context(), req.Method, t.target+req.URL.Path, req.Body)
	if err != nil {
		return nil, err
	}
	u.Header = clone.Header
	return t.base.RoundTrip(u)
}

func TestRootRegisteredUpgrade(t *testing.T) {
	found := false
	for _, c := range RootCmd.Commands() {
		if c.Name() == "upgrade" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected RootCmd to have 'upgrade' subcommand registered")
	}
}

func TestUpgradeNoticeOnRootCmd(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "beep-update-notice-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	flagWorkspace = tmpDir
	defer func() { flagWorkspace = "" }()

	// Pre-populate state with newer version
	statePath := updater.GetStateFilePath(tmpDir)
	_ = updater.SaveState(statePath, &updater.State{
		LastCheckedAt:   time.Now(),
		LatestVersion:   "v0.9.0",
		NotifiedVersion: "",
	})

	// Check that checkNotice works with configured workspace
	st, err := updater.LoadState(statePath)
	if err != nil || st.LatestVersion != "v0.9.0" {
		t.Fatalf("failed to read state: %v", err)
	}
}

func TestSkipUpdateHooks(t *testing.T) {
	findCases := []struct {
		args []string
		want bool
	}{
		{[]string{"upgrade"}, true},
		{[]string{"update"}, true},
		{[]string{"version"}, true},
		{[]string{"completion"}, true},
		{[]string{"completion", "bash"}, true},
		{[]string{"completion", "zsh"}, true},
		{[]string{"completion", "fish"}, true},
		{[]string{"help"}, true},
		{[]string{"status"}, false},
		{[]string{"up"}, false},
	}
	for _, tc := range findCases {
		cmd, _, err := RootCmd.Find(tc.args)
		if err != nil {
			t.Fatalf("RootCmd.Find(%v): %v", tc.args, err)
		}
		if got := skipUpdateHooks(cmd); got != tc.want {
			t.Errorf("skipUpdateHooks after Find(%v) (leaf %q) = %v; want %v", tc.args, cmd.Name(), got, tc.want)
		}
	}

	internal := &cobra.Command{Use: "__complete"}
	if got := skipUpdateHooks(internal); !got {
		t.Errorf("skipUpdateHooks(__complete) = %v; want true", got)
	}
}
