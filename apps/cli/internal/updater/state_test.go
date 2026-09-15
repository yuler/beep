package updater

import (
	"path/filepath"
	"testing"
	"time"
)

func TestStateLoadSave(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "update.json")

	// 1. Initial load when file does not exist
	state, err := LoadState(filePath)
	if err != nil {
		t.Fatalf("unexpected error loading non-existent state: %v", err)
	}
	if state == nil {
		t.Fatal("expected non-nil state")
	}
	if state.LatestVersion != "" {
		t.Errorf("expected empty LatestVersion, got %q", state.LatestVersion)
	}

	// 2. Save state
	now := time.Now().Truncate(time.Second)
	state.LastCheckedAt = now
	state.LatestVersion = "v0.3.0"
	state.NotifiedVersion = "v0.2.5"

	if err := SaveState(filePath, state); err != nil {
		t.Fatalf("failed to save state: %v", err)
	}

	// 3. Reload state
	reloaded, err := LoadState(filePath)
	if err != nil {
		t.Fatalf("failed to reload state: %v", err)
	}
	if reloaded.LatestVersion != "v0.3.0" {
		t.Errorf("expected LatestVersion v0.3.0, got %q", reloaded.LatestVersion)
	}
	if reloaded.NotifiedVersion != "v0.2.5" {
		t.Errorf("expected NotifiedVersion v0.2.5, got %q", reloaded.NotifiedVersion)
	}
}
