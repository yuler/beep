package cmd

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"beep/internal/client"

	"github.com/spf13/cobra"
)

func TestBeeperCommandsRegistration(t *testing.T) {
	for _, sub := range []string{"list", "show", "apps", "create", "delete", "pause", "resume", "run", "runs"} {
		cmd, _, err := RootCmd.Find([]string{"beeper", sub})
		if err != nil {
			t.Fatalf("failed to find 'beeper %s': %v", sub, err)
		}
		if cmd.Name() != sub {
			t.Errorf("expected command name %q, got %q", sub, cmd.Name())
		}
	}
}

func setupBeeperTestEnv(t *testing.T, handler http.HandlerFunc) (string, func()) {
	return setupCLITestEnv(t, handler)
}

func findBeeperCmd(t *testing.T, sub string) *cobra.Command {
	t.Helper()
	cmd, _, err := RootCmd.Find([]string{"beeper", sub})
	if err != nil {
		t.Fatalf("failed to find 'beeper %s': %v", sub, err)
	}
	return cmd
}



func TestBeeperListCommand(t *testing.T) {
	var gotHeaderAccount string
	mockBeepers := []*client.Beeper{
		{
			ID:         "beeper_101",
			Title:      "Production Healthcheck",
			Status:     "active",
			AlertState: "ok",
			Cron:       "*/5 * * * *",
			PingToken:  "beep_pt_supersecret123",
			BeeperApp: &client.BeeperApp{
				Slug: "heartbeat-ping",
				Name: "Heartbeat Ping",
			},
			RunStats: &client.BeeperRunStats{
				Total:     10,
				Succeeded: 10,
			},
		},
	}

	_, cleanup := setupBeeperTestEnv(t, func(w http.ResponseWriter, r *http.Request) {
		gotHeaderAccount = r.Header.Get("X-Account-Slug")
		if r.URL.Path == "/api/v1/me" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"identity": map[string]any{"id": "id_1", "email": "test@example.com"},
			})
			return
		}
		if r.URL.Path == "/api/v1/beepers" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"beepers": mockBeepers,
			})
			return
		}
		http.NotFound(w, r)
	})
	defer cleanup()

	// 1. Plain list without --account (personal default, header empty)
	listCmd := findBeeperCmd(t, "list")
	out, err := captureStdout(func() error {
		return listCmd.RunE(listCmd, nil)
	})
	if err != nil {
		t.Fatalf("beeper list failed: %v", err)
	}
	if gotHeaderAccount != "" {
		t.Errorf("expected empty X-Account-Slug for default personal account, got %q", gotHeaderAccount)
	}
	if !strings.Contains(out, "Production Healthcheck") || !strings.Contains(out, "heartbeat-ping") {
		t.Errorf("expected output to contain beeper info, got: %s", out)
	}

	// 2. With --account team-alpha
	flagAccount = "team-alpha"
	_, err = captureStdout(func() error {
		return listCmd.RunE(listCmd, nil)
	})
	if err != nil {
		t.Fatalf("beeper list with account failed: %v", err)
	}
	if gotHeaderAccount != "team-alpha" {
		t.Errorf("expected X-Account-Slug header 'team-alpha', got %q", gotHeaderAccount)
	}

	// 3. With --json (verify token is masked)
	flagJSON = true
	jsonOut, err := captureStdout(func() error {
		return listCmd.RunE(listCmd, nil)
	})
	if err != nil {
		t.Fatalf("beeper list --json failed: %v", err)
	}
	var parsed []*client.Beeper
	if err := json.Unmarshal([]byte(jsonOut), &parsed); err != nil {
		t.Fatalf("expected valid JSON array, got: %v", err)
	}
	if len(parsed) != 1 || parsed[0].ID != "beeper_101" {
		t.Errorf("unexpected parsed JSON: %+v", parsed)
	}
	if parsed[0].PingToken != "beep_pt_••••••••" {
		t.Errorf("expected masked ping token in list --json, got %q", parsed[0].PingToken)
	}
}

func TestBeeperShowCommandTokenMasking(t *testing.T) {
	mockBeeper := &client.Beeper{
		ID:        "beeper_202",
		Title:     "Ping Probe",
		Status:    "active",
		PingToken: "beep_pt_supersecret123",
	}

	_, cleanup := setupBeeperTestEnv(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/me" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"identity": map[string]any{"id": "id_1", "email": "test@example.com"},
			})
			return
		}
		if r.URL.Path == "/api/v1/beepers/beeper_202" {
			_ = json.NewEncoder(w).Encode(mockBeeper)
			return
		}
		http.NotFound(w, r)
	})
	defer cleanup()

	showCmd := findBeeperCmd(t, "show")

	// 1. Default masked in human show
	out, err := captureStdout(func() error {
		return showCmd.RunE(showCmd, []string{"beeper_202"})
	})
	if err != nil {
		t.Fatalf("beeper show failed: %v", err)
	}
	if !strings.Contains(out, "beep_pt_••••••••") || strings.Contains(out, "supersecret123") {
		t.Errorf("expected masked token in show, got: %s", out)
	}

	// 2. Default masked in --json show
	flagJSON = true
	jsonOut, err := captureStdout(func() error {
		return showCmd.RunE(showCmd, []string{"beeper_202"})
	})
	flagJSON = false
	if err != nil {
		t.Fatalf("beeper show --json failed: %v", err)
	}
	var maskedBeeper client.Beeper
	if err := json.Unmarshal([]byte(jsonOut), &maskedBeeper); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}
	if maskedBeeper.PingToken != "beep_pt_••••••••" {
		t.Errorf("expected masked token in --json show, got %q", maskedBeeper.PingToken)
	}

	// 3. Unmasked with --show-token
	showCmd.Flags().Set("show-token", "true")

	out, err = captureStdout(func() error {
		return showCmd.RunE(showCmd, []string{"beeper_202"})
	})
	if err != nil {
		t.Fatalf("beeper show --show-token failed: %v", err)
	}
	if !strings.Contains(out, "beep_pt_supersecret123") {
		t.Errorf("expected raw token in show --show-token, got: %s", out)
	}

	// 4. Unmasked in --json with --show-token
	flagJSON = true
	jsonOut, err = captureStdout(func() error {
		return showCmd.RunE(showCmd, []string{"beeper_202"})
	})
	flagJSON = false
	showCmd.Flags().Set("show-token", "false")
	if err != nil {
		t.Fatalf("beeper show --show-token --json failed: %v", err)
	}
	var unmaskedBeeper client.Beeper
	if err := json.Unmarshal([]byte(jsonOut), &unmaskedBeeper); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}
	if unmaskedBeeper.PingToken != "beep_pt_supersecret123" {
		t.Errorf("expected raw token in --json show with --show-token, got %q", unmaskedBeeper.PingToken)
	}
}

func TestBeeperAppsCommand(t *testing.T) {
	mockApps := []*client.BeeperApp{
		{
			Slug:        "heartbeat-ping",
			Name:        "Heartbeat Ping",
			DefaultCron: "*/5 * * * *",
			Description: "Receive periodic ping signals",
		},
	}

	_, cleanup := setupBeeperTestEnv(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/beeper_apps" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"beeper_apps": mockApps,
			})
			return
		}
		http.NotFound(w, r)
	})
	defer cleanup()

	appsCmd := findBeeperCmd(t, "apps")
	out, err := captureStdout(func() error {
		return appsCmd.RunE(appsCmd, nil)
	})
	if err != nil {
		t.Fatalf("beeper apps failed: %v", err)
	}
	if !strings.Contains(out, "heartbeat-ping") || !strings.Contains(out, "Heartbeat Ping") {
		t.Errorf("expected catalog output, got: %s", out)
	}
}

func TestBeeperCreateCommand(t *testing.T) {
	var receivedBody client.CreateBeeperRequest

	_, cleanup := setupBeeperTestEnv(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/me" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"identity": map[string]any{"id": "id_1", "email": "test@example.com"},
			})
			return
		}
		if r.URL.Path == "/api/v1/beeper_apps/heartbeat-ping" {
			_ = json.NewEncoder(w).Encode(&client.BeeperApp{
				Slug:        "heartbeat-ping",
				Name:        "Heartbeat Ping",
				DefaultCron: "*/5 * * * *",
			})
			return
		}
		if r.URL.Path == "/api/v1/beepers" && r.Method == http.MethodPost {
			_ = json.NewDecoder(r.Body).Decode(&receivedBody)
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(&client.Beeper{
				ID:        "beeper_created",
				Title:     receivedBody.Title,
				Cron:      receivedBody.Cron,
				PingToken: "ping_token_123",
			})
			return
		}
		http.NotFound(w, r)
	})
	defer cleanup()

	flagNoInteractive = true
	defer func() { flagNoInteractive = false }()

	createCmd := findBeeperCmd(t, "create")
	createCmd.Flags().Set("app", "heartbeat-ping")
	createCmd.Flags().Set("title", "Gateway Ping")
	createCmd.Flags().Set("config", "ping_interval=300")

	out, err := captureStdout(func() error {
		return createCmd.RunE(createCmd, nil)
	})
	createCmd.Flags().Set("app", "")
	createCmd.Flags().Set("title", "")
	createCmd.Flags().Set("config", "")

	if err != nil {
		t.Fatalf("beeper create failed: %v", err)
	}
	if receivedBody.Title != "Gateway Ping" || receivedBody.BeeperAppSlug != "heartbeat-ping" {
		t.Errorf("unexpected received create request: %+v", receivedBody)
	}
	if !strings.Contains(out, "ping_token_123") {
		t.Errorf("expected ping token in output: %s", out)
	}
}

func TestBeeperActionsCommands(t *testing.T) {
	var pausedID, resumedID, runID, runsID, deletedID string

	_, cleanup := setupBeeperTestEnv(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/me" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"identity": map[string]any{"id": "id_1", "email": "test@example.com"},
			})
			return
		}
		if r.URL.Path == "/api/v1/beepers/bp1/pause" && r.Method == http.MethodPost {
			pausedID = "bp1"
			_ = json.NewEncoder(w).Encode(&client.Beeper{ID: "bp1", Status: "paused"})
			return
		}
		if r.URL.Path == "/api/v1/beepers/bp1/pause" && r.Method == http.MethodDelete {
			resumedID = "bp1"
			_ = json.NewEncoder(w).Encode(&client.Beeper{ID: "bp1", Status: "active"})
			return
		}
		if r.URL.Path == "/api/v1/beepers/bp1/runs" && r.Method == http.MethodPost {
			runID = "bp1"
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(&client.BeeperRun{ID: "brun_1", Status: "ok"})
			return
		}
		if r.URL.Path == "/api/v1/beepers/bp1/runs" && r.Method == http.MethodGet {
			runsID = "bp1"
			_ = json.NewEncoder(w).Encode(map[string]any{
				"runs": []*client.BeeperRun{
					{ID: "brun_1", Status: "ok", ScheduledFor: "2026-09-15T12:00:00Z"},
				},
			})
			return
		}
		if r.URL.Path == "/api/v1/beepers/bp1" && r.Method == http.MethodDelete {
			deletedID = "bp1"
			w.WriteHeader(http.StatusNoContent)
			return
		}
		http.NotFound(w, r)
	})
	defer cleanup()

	pauseCmd := findBeeperCmd(t, "pause")
	resumeCmd := findBeeperCmd(t, "resume")
	runCmd := findBeeperCmd(t, "run")
	runsCmd := findBeeperCmd(t, "runs")
	deleteCmd := findBeeperCmd(t, "delete")

	// Pause
	if _, err := captureStdout(func() error { return pauseCmd.RunE(pauseCmd, []string{"bp1"}) }); err != nil {
		t.Fatalf("pause failed: %v", err)
	}
	if pausedID != "bp1" {
		t.Errorf("expected pause on bp1")
	}

	// Resume
	if _, err := captureStdout(func() error { return resumeCmd.RunE(resumeCmd, []string{"bp1"}) }); err != nil {
		t.Fatalf("resume failed: %v", err)
	}
	if resumedID != "bp1" {
		t.Errorf("expected resume on bp1")
	}

	// Run
	if _, err := captureStdout(func() error { return runCmd.RunE(runCmd, []string{"bp1"}) }); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if runID != "bp1" {
		t.Errorf("expected run on bp1")
	}

	// Runs
	if _, err := captureStdout(func() error { return runsCmd.RunE(runsCmd, []string{"bp1"}) }); err != nil {
		t.Fatalf("runs failed: %v", err)
	}
	if runsID != "bp1" {
		t.Errorf("expected runs on bp1")
	}

	// Delete
	if _, err := captureStdout(func() error { return deleteCmd.RunE(deleteCmd, []string{"bp1"}) }); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if deletedID != "bp1" {
		t.Errorf("expected delete on bp1")
	}
}

func TestBeeperOmittedIDNonInteractive(t *testing.T) {
	_, cleanup := setupBeeperTestEnv(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/me" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"identity": map[string]any{"id": "id_1", "email": "test@example.com"},
			})
			return
		}
		http.NotFound(w, r)
	})
	defer cleanup()

	flagNoInteractive = true
	defer func() { flagNoInteractive = false }()

	commands := []*cobra.Command{
		findBeeperCmd(t, "show"),
		findBeeperCmd(t, "delete"),
		findBeeperCmd(t, "pause"),
		findBeeperCmd(t, "resume"),
		findBeeperCmd(t, "run"),
		findBeeperCmd(t, "runs"),
	}
	for _, c := range commands {
		err := c.RunE(c, nil)
		if err == nil {
			t.Errorf("expected error for command %s without ID in non-interactive mode", c.Name())
		} else if !strings.Contains(err.Error(), "beeper ID is required") {
			t.Errorf("expected 'beeper ID is required' error for %s, got: %v", c.Name(), err)
		}
	}
}

