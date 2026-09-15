package updater

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"beep/internal/config"
)

// State holds cached release check results and notification tracking.
type State struct {
	LastCheckedAt   time.Time `json:"last_checked_at"`
	LatestVersion   string    `json:"latest_version"`
	LatestURL       string    `json:"latest_url,omitempty"`
	NotifiedVersion string    `json:"notified_version,omitempty"`
}

// GetStateFilePath returns the full path to the update state file within the workspace.
func GetStateFilePath(workspaceDir string) string {
	if workspaceDir == "" {
		workspaceDir = config.DefaultWorkspace()
	}
	return filepath.Join(workspaceDir, "update.json")
}

// LoadState reads the update state file, returning an empty state if the file does not exist.
func LoadState(filePath string) (*State, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &State{}, nil
		}
		return nil, err
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		// Corrupted or invalid state file; return blank state rather than failing
		return &State{}, nil
	}
	return &state, nil
}

// SaveState atomically writes the update state to the given file path.
func SaveState(filePath string, state *State) error {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("failed to create state directory %s: %w", dir, err)
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	// Write to temporary file in the same directory then atomic rename
	tmpFile := filepath.Join(dir, fmt.Sprintf(".update-%d.tmp", time.Now().UnixNano()))
	if err := os.WriteFile(tmpFile, append(data, '\n'), 0o600); err != nil {
		return err
	}

	if err := os.Rename(tmpFile, filePath); err != nil {
		_ = os.Remove(tmpFile)
		return err
	}

	return nil
}
