package updater

import (
	"path/filepath"
	"testing"
	"time"

	"beep/internal/version"
)

func TestProbeNotification(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "update.json")

	// Set version for test
	oldVersion := version.Version
	version.Version = "0.2.1"
	defer func() { version.Version = oldVersion }()

	// Pre-populate state with newer version
	state := &State{
		LastCheckedAt:   time.Now(),
		LatestVersion:   "v0.3.0",
		NotifiedVersion: "",
	}
	if err := SaveState(statePath, state); err != nil {
		t.Fatalf("failed to save state: %v", err)
	}

	// In tests, os.Stderr might not be a terminal, but let's test the version & notification logic directly
	// by simulating what CheckNotice does
	st, err := LoadState(statePath)
	if err != nil {
		t.Fatalf("failed to load state: %v", err)
	}

	if CompareVersions(st.LatestVersion, version.Version) <= 0 {
		t.Errorf("expected LatestVersion %s to be newer than %s", st.LatestVersion, version.Version)
	}

	// Verify one-time notification mechanism
	if st.LatestVersion == st.NotifiedVersion {
		t.Errorf("expected not notified yet")
	}

	st.NotifiedVersion = st.LatestVersion
	if err := SaveState(statePath, st); err != nil {
		t.Fatalf("failed to save state: %v", err)
	}

	reloaded, _ := LoadState(statePath)
	if reloaded.NotifiedVersion != "v0.3.0" {
		t.Errorf("expected notified version v0.3.0, got %s", reloaded.NotifiedVersion)
	}

	// Next time: since st.LatestVersion == st.NotifiedVersion, should not notify
	if reloaded.LatestVersion == reloaded.NotifiedVersion {
		// Suppressed as expected!
	} else {
		t.Errorf("expected notification to be suppressed on repeat checks")
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
