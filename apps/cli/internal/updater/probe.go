package updater

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"beep/internal/config"
	"beep/internal/ui"
	"beep/internal/version"

	"github.com/mattn/go-isatty"
)

// allowNoticeWithoutTTY lets tests exercise CheckNotice without a TTY on stderr.
var allowNoticeWithoutTTY bool

// IsCheckDisabled returns true if update checks are turned off via environment.
func IsCheckDisabled() bool {
	val := strings.ToLower(strings.TrimSpace(os.Getenv("BEEP_NO_UPDATE_CHECK")))
	if val == "1" || val == "true" || val == "yes" {
		return true
	}
	if os.Getenv("CI") != "" {
		return true
	}
	return false
}

// TriggerBackgroundCheck spawns a non-blocking goroutine to fetch the latest release
// if the local cache is older than 24 hours.
func TriggerBackgroundCheck(workspaceDir string) {
	if IsCheckDisabled() {
		return
	}

	ws := workspaceDir
	if ws == "" {
		ws = config.DefaultWorkspace()
	}
	statePath := GetStateFilePath(ws)

	state, err := LoadState(statePath)
	if err != nil || state == nil {
		state = &State{}
	}

	// Cache TTL: 24 hours
	if !state.LastCheckedAt.IsZero() && time.Since(state.LastCheckedAt) < 24*time.Hour {
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
		defer cancel()

		rel, err := FetchLatestRelease(ctx, GetRepo())
		if err != nil || rel == nil || rel.Tag == "" {
			return
		}

		// Reload state in case another process modified it
		currState, _ := LoadState(statePath)
		if currState == nil {
			currState = &State{}
		}
		currState.LastCheckedAt = time.Now()
		currState.LatestVersion = rel.Tag
		currState.LatestURL = rel.HTMLURL
		_ = SaveState(statePath, currState)
	}()
}

// CheckNotice returns a formatted notification banner if a new release is available
// and has not yet been notified to the user. It does not persist notification state;
// callers should call MarkNotified after successfully printing the notice.
func CheckNotice(workspaceDir string) string {
	if IsCheckDisabled() {
		return ""
	}

	// Do not notify during local development builds
	if version.Version == "dev" {
		return ""
	}

	// Only print notification if stderr is an interactive terminal (or tests opt in)
	if !allowNoticeWithoutTTY && !isatty.IsTerminal(os.Stderr.Fd()) {
		return ""
	}

	ws := workspaceDir
	if ws == "" {
		ws = config.DefaultWorkspace()
	}
	statePath := GetStateFilePath(ws)

	state, err := LoadState(statePath)
	if err != nil || state == nil || state.LatestVersion == "" {
		return ""
	}

	// Check if this specific version was already notified
	if state.LatestVersion == state.NotifiedVersion {
		return ""
	}

	// Check if latest version is newer than current version
	if CompareVersions(state.LatestVersion, version.Version) <= 0 {
		return ""
	}

	notice := fmt.Sprintf(
		"\n%s %s %s → %s\n  Run %s to update.\n",
		ui.Yellow("!"),
		ui.Bold("A new release of beep is available:"),
		ui.Dim(version.Version),
		ui.Bold(ui.Green(state.LatestVersion)),
		ui.Cyan("beep upgrade"),
	)
	return notice
}

// MarkNotified records that the user has been shown the notice for LatestVersion.
func MarkNotified(workspaceDir string) error {
	ws := workspaceDir
	if ws == "" {
		ws = config.DefaultWorkspace()
	}
	statePath := GetStateFilePath(ws)

	state, err := LoadState(statePath)
	if err != nil {
		return err
	}
	if state == nil {
		state = &State{}
	}
	if state.LatestVersion == "" {
		return nil
	}
	state.NotifiedVersion = state.LatestVersion
	return SaveState(statePath, state)
}

// RunDaemonUpdateProbe runs a periodic ticker in a long-running daemon (beep up)
// and logs a message if a newer version is discovered.
func RunDaemonUpdateProbe(ctx context.Context, ws string, logFn func(format string, args ...any)) {
	if IsCheckDisabled() {
		return
	}

	if ws == "" {
		ws = config.DefaultWorkspace()
	}
	statePath := GetStateFilePath(ws)

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	checkFn := func() {
		checkCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		rel, err := FetchLatestRelease(checkCtx, GetRepo())
		if err != nil || rel == nil || rel.Tag == "" {
			return
		}

		st, _ := LoadState(statePath)
		if st == nil {
			st = &State{}
		}
		st.LastCheckedAt = time.Now()
		st.LatestVersion = rel.Tag
		st.LatestURL = rel.HTMLURL

		if CompareVersions(rel.Tag, version.Version) > 0 && st.NotifiedVersion != rel.Tag {
			if logFn != nil {
				logFn("[beep] A new version of beep is available: %s -> %s (run 'beep upgrade' to update)", version.Version, rel.Tag)
			}
			st.NotifiedVersion = rel.Tag
		}
		_ = SaveState(statePath, st)
	}

	// Initial check after 15 seconds of uptime
	select {
	case <-time.After(15 * time.Second):
		checkFn()
	case <-ctx.Done():
		return
	}

	for {
		select {
		case <-ticker.C:
			checkFn()
		case <-ctx.Done():
			return
		}
	}
}
