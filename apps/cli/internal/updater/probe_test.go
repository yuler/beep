package updater

import (
	"strings"
	"testing"
	"time"

	"beep/internal/version"
)

func TestProbeNotification(t *testing.T) {
	tmpDir := t.TempDir()

	oldVersion := version.Version
	version.Version = "0.2.1"
	defer func() { version.Version = oldVersion }()

	oldAllow := allowNoticeWithoutTTY
	allowNoticeWithoutTTY = true
	defer func() { allowNoticeWithoutTTY = oldAllow }()

	t.Setenv("BEEP_NO_UPDATE_CHECK", "")
	t.Setenv("CI", "")

	statePath := GetStateFilePath(tmpDir)
	state := &State{
		LastCheckedAt:   time.Now(),
		LatestVersion:   "v0.3.0",
		NotifiedVersion: "",
	}
	if err := SaveState(statePath, state); err != nil {
		t.Fatalf("failed to save state: %v", err)
	}

	notice := CheckNotice(tmpDir)
	if notice == "" {
		t.Fatalf("expected non-empty notice from CheckNotice")
	}
	if !strings.Contains(notice, "v0.3.0") {
		t.Errorf("expected notice to mention v0.3.0, got %q", notice)
	}

	// CheckNotice must not mark notified by itself
	st, err := LoadState(statePath)
	if err != nil {
		t.Fatalf("LoadState: %v", err)
	}
	if st.NotifiedVersion != "" {
		t.Errorf("CheckNotice should not set NotifiedVersion, got %q", st.NotifiedVersion)
	}

	if err := MarkNotified(tmpDir); err != nil {
		t.Fatalf("MarkNotified: %v", err)
	}

	notice2 := CheckNotice(tmpDir)
	if notice2 != "" {
		t.Errorf("expected empty notice after MarkNotified, got %q", notice2)
	}

	reloaded, _ := LoadState(statePath)
	if reloaded.NotifiedVersion != "v0.3.0" {
		t.Errorf("expected notified version v0.3.0, got %s", reloaded.NotifiedVersion)
	}
}

func TestIsCheckDisabled(t *testing.T) {
	t.Setenv("BEEP_NO_UPDATE_CHECK", "1")
	if !IsCheckDisabled() {
		t.Errorf("expected IsCheckDisabled() to be true when BEEP_NO_UPDATE_CHECK=1")
	}

	t.Setenv("BEEP_NO_UPDATE_CHECK", "0")
	t.Setenv("CI", "")
	if IsCheckDisabled() {
		t.Errorf("expected IsCheckDisabled() to be false when disabled flag is 0 and CI is empty")
	}
}

func TestDownloadClientHasNoShortTimeout(t *testing.T) {
	if downloadHTTPClient.Timeout != 0 {
		t.Errorf("downloadHTTPClient.Timeout = %v; want 0 (rely on context only)", downloadHTTPClient.Timeout)
	}
	if apiHTTPClient.Timeout == 0 {
		t.Errorf("apiHTTPClient.Timeout should be non-zero for API calls")
	}
}
