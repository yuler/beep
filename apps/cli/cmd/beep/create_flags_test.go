package beep

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"beep/internal/client"
	"beep/internal/cmdutil"
)

func setupTestWorkspace(t *testing.T, serverURL, token, accountSlug string) string {
	t.Helper()
	dir := t.TempDir()
	cfg := map[string]any{
		"server_url":   serverURL,
		"access_token": token,
		"account_slug": accountSlug,
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("failed to marshal config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.json"), data, 0o600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}
	cmdutil.SetOverrideWorkspace(dir)
	t.Cleanup(func() {
		cmdutil.SetOverrideWorkspace("")
	})
	return dir
}

func TestCreateBeepParamsToRequest_IntentAndMetadata(t *testing.T) {
	params := client.CreateBeepParams{
		Title:        "Lunch Time",
		Body:         "Take a break",
		ScheduleKind: "instant",
		Timezone:     "Asia/Shanghai",
		Channels:     "cli,desktop",
		Intent:       "lunch_break",
		Metadata: map[string]any{
			"first_checkin_time": "09:05:00",
		},
	}

	req, err := params.ToRequest()
	if err != nil {
		t.Fatalf("ToRequest failed: %v", err)
	}

	if req.Intent != "lunch_break" {
		t.Errorf("got Intent %q, want 'lunch_break'", req.Intent)
	}
	if req.Metadata["first_checkin_time"] != "09:05:00" {
		t.Errorf("got Metadata %v, want first_checkin_time=09:05:00", req.Metadata)
	}
	if len(req.NotificationChannels) != 2 || req.NotificationChannels[0] != "cli" || req.NotificationChannels[1] != "desktop" {
		t.Errorf("got channels %v, want ['cli', 'desktop']", req.NotificationChannels)
	}
}

func TestNewCmdCreate_FlagsAndAliases(t *testing.T) {
	var receivedReq client.CreateBeepRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/beeps" && r.Method == http.MethodPost {
			_ = json.NewDecoder(r.Body).Decode(&receivedReq)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			resp := client.Beep{
				ID:                   "beep-123",
				Title:                receivedReq.Title,
				Body:                 receivedReq.Body,
				Kind:                 receivedReq.Kind,
				RunAt:                receivedReq.RunAt,
				Intent:               receivedReq.Intent,
				Metadata:             receivedReq.Metadata,
				NotificationChannels: receivedReq.NotificationChannels,
			}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	setupTestWorkspace(t, server.URL, "token", "slug")

	cmd := NewCmdCreate()
	cmd.SetArgs([]string{
		"Get Off Work",
		"--body", "Time to leave!",
		"--run-at", "18:30",
		"--channel", "cli",
		"--intent", "get_off_work",
		"--metadata", `{"today":"2026-09-18"}`,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if receivedReq.Title != "Get Off Work" {
		t.Errorf("got Title %q, want 'Get Off Work'", receivedReq.Title)
	}
	if receivedReq.Body != "Time to leave!" {
		t.Errorf("got Body %q, want 'Time to leave!'", receivedReq.Body)
	}
	if receivedReq.Intent != "get_off_work" {
		t.Errorf("got Intent %q, want 'get_off_work'", receivedReq.Intent)
	}
	if receivedReq.Metadata["today"] != "2026-09-18" {
		t.Errorf("got Metadata %v, want today=2026-09-18", receivedReq.Metadata)
	}
	if len(receivedReq.NotificationChannels) != 1 || receivedReq.NotificationChannels[0] != "cli" {
		t.Errorf("got NotificationChannels %v, want ['cli']", receivedReq.NotificationChannels)
	}
}

func TestNewCmdCreate_InvalidMetadata(t *testing.T) {
	cmd := NewCmdCreate()
	cmd.SetArgs([]string{
		"Invalid Metadata Beep",
		"--metadata", "not-a-valid-json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error for invalid metadata JSON, got nil")
	}
	if !strings.Contains(err.Error(), "invalid --metadata") {
		t.Errorf("expected 'invalid --metadata' in error, got: %v", err)
	}
}

func TestNewCmdCreate_ConflictingFlags(t *testing.T) {
	// 1. Conflicting --at and --run-at
	cmd1 := NewCmdCreate()
	cmd1.SetArgs([]string{
		"Conflict Beep",
		"--at", "10:00",
		"--run-at", "11:00",
	})
	if err := cmd1.Execute(); err == nil {
		t.Fatalf("expected error for conflicting --at and --run-at, got nil")
	}

	// 2. Conflicting --channels and --channel
	cmd2 := NewCmdCreate()
	cmd2.SetArgs([]string{
		"Conflict Beep",
		"--channels", "cli",
		"--channel", "desktop",
	})
	if err := cmd2.Execute(); err == nil {
		t.Fatalf("expected error for conflicting --channel and --channels, got nil")
	}
}
