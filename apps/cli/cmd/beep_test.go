package cmd

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"beep/internal/client"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func resetCmdFlags(cmd *cobra.Command) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		f.Changed = false
		_ = f.Value.Set(f.DefValue)
	})
}

func TestBeepCommandsRegistration(t *testing.T) {
	// 1. Root 'beep' command is in core group
	legacyBeep, _, err := RootCmd.Find([]string{"beep"})
	if err != nil {
		t.Fatalf("failed to find 'beep' command: %v", err)
	}
	if legacyBeep.GroupID != "core" {
		t.Errorf("expected 'beep' command to have GroupID 'core', got %q", legacyBeep.GroupID)
	}

	for _, sub := range []string{"list", "show", "create", "delete", "pause", "resume", "run"} {
		// Canonical top-level command
		topCmd, _, err := RootCmd.Find([]string{sub})
		if err != nil {
			t.Fatalf("failed to find canonical top-level %q: %v", sub, err)
		}
		if topCmd.Name() != sub {
			t.Errorf("expected top-level command name %q, got %q", sub, topCmd.Name())
		}
		if topCmd.GroupID != "beeps" {
			t.Errorf("expected command %q to have GroupID 'beeps', got %q", sub, topCmd.GroupID)
		}

		// Backward-compatible legacy command 'beep <subcommand>'
		legacyCmd, _, err := RootCmd.Find([]string{"beep", sub})
		if err != nil {
			t.Fatalf("failed to find legacy 'beep %s': %v", sub, err)
		}
		if legacyCmd.Name() != sub {
			t.Errorf("expected legacy command name %q, got %q", sub, legacyCmd.Name())
		}

		// Verify independent command instances (no shared pointer between Root and legacy parent)
		if topCmd == legacyCmd {
			t.Errorf("top-level %q and legacy 'beep %s' share the same *cobra.Command pointer instance", sub, sub)
		}
	}
}

func setupBeepTestEnv(t *testing.T, handler http.HandlerFunc) (string, func()) {
	return setupCLITestEnv(t, handler)
}

func findBeepCmd(t *testing.T, sub string) *cobra.Command {
	t.Helper()
	cmd, _, err := RootCmd.Find([]string{sub})
	if err != nil {
		t.Fatalf("failed to find %q: %v", sub, err)
	}
	return cmd
}

func TestBeepListCommand(t *testing.T) {
	var gotHeaderAccount string
	mockBeeps := []*client.Beep{
		{
			ID:                   "beep_123",
			Title:                "Check Deploy",
			Status:               "active",
			Kind:                 "recurring",
			Cron:                 "0 9 * * *",
			NotificationChannels: []string{"slack"},
		},
		{
			ID:                   "beep_chinese",
			Title:                "这是一个非常长的中文标题用来测试截断是否会出现乱码字符",
			Status:               "active",
			Kind:                 "once",
			NotificationChannels: []string{"email"},
		},
	}

	_, cleanup := setupBeepTestEnv(t, func(w http.ResponseWriter, r *http.Request) {
		gotHeaderAccount = r.Header.Get("X-Account-Slug")
		if r.URL.Path == "/api/v1/me" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"identity": map[string]any{"id": "id_1", "email": "test@example.com"},
			})
			return
		}
		if r.URL.Path == "/api/v1/beeps" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"beeps": mockBeeps,
			})
			return
		}
		http.NotFound(w, r)
	})
	defer cleanup()

	// 1. Plain table output without --account (personal default, header empty)
	listCmd := findBeepCmd(t, "list")
	out, err := captureStdout(func() error {
		return listCmd.RunE(listCmd, nil)
	})
	if err != nil {
		t.Fatalf("beep list failed: %v", err)
	}
	if gotHeaderAccount != "" {
		t.Errorf("expected empty X-Account-Slug header for default personal account, got %q", gotHeaderAccount)
	}
	if !strings.Contains(out, "Check Deploy") || !strings.Contains(out, "beep_123") {
		t.Errorf("expected output to contain beep info, got: %s", out)
	}

	// 2. With --account team-slug
	flagAccount = "team-slug"
	_, err = captureStdout(func() error {
		return listCmd.RunE(listCmd, nil)
	})
	if err != nil {
		t.Fatalf("beep list with account failed: %v", err)
	}
	if gotHeaderAccount != "team-slug" {
		t.Errorf("expected X-Account-Slug header 'team-slug', got %q", gotHeaderAccount)
	}

	// 3. With --json
	flagJSON = true
	jsonOut, err := captureStdout(func() error {
		return listCmd.RunE(listCmd, nil)
	})
	if err != nil {
		t.Fatalf("beep list --json failed: %v", err)
	}
	var parsed []*client.Beep
	if err := json.Unmarshal([]byte(jsonOut), &parsed); err != nil {
		t.Fatalf("expected valid JSON array, got error: %v, raw: %s", err, jsonOut)
	}
	if len(parsed) != 2 || parsed[0].ID != "beep_123" || parsed[1].ID != "beep_chinese" {
		t.Errorf("unexpected parsed JSON: %+v", parsed)
	}
}

func TestBeepShowCommand(t *testing.T) {
	mockBeep := &client.Beep{
		ID:        "beep_456",
		Title:     "Standup Reminder",
		Status:    "active",
		Kind:      "recurring",
		Cron:      "0 10 * * 1-5",
		Timezone:  "Asia/Shanghai",
		Body:      "Remember to post your update",
		NextRunAt: "2026-09-16T10:00:00+08:00",
		Runs: []client.BeepRun{
			{ID: "run_1", Status: "succeeded", ScheduledFor: "2026-09-15T10:00:00+08:00"},
		},
	}

	_, cleanup := setupBeepTestEnv(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/me" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"identity": map[string]any{"id": "id_1", "email": "test@example.com"},
			})
			return
		}
		if r.URL.Path == "/api/v1/beeps/beep_456" {
			_ = json.NewEncoder(w).Encode(mockBeep)
			return
		}
		http.NotFound(w, r)
	})
	defer cleanup()

	showCmd := findBeepCmd(t, "show")
	out, err := captureStdout(func() error {
		return showCmd.RunE(showCmd, []string{"beep_456"})
	})
	if err != nil {
		t.Fatalf("beep show failed: %v", err)
	}
	if !strings.Contains(out, "Standup Reminder") || !strings.Contains(out, "Asia/Shanghai") {
		t.Errorf("expected output to contain beep details, got: %s", out)
	}
}

func TestBeepCreateCommand(t *testing.T) {
	var receivedBody client.CreateBeepRequest

	_, cleanup := setupBeepTestEnv(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/me" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"identity": map[string]any{"id": "id_1", "email": "test@example.com"},
			})
			return
		}
		if r.URL.Path == "/api/v1/beeps" && r.Method == http.MethodPost {
			_ = json.NewDecoder(r.Body).Decode(&receivedBody)
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(&client.Beep{
				ID:       "beep_new",
				Title:    receivedBody.Title,
				Kind:     receivedBody.Kind,
				Cron:     receivedBody.Cron,
				RunAt:    receivedBody.RunAt,
				Timezone: receivedBody.Timezone,
			})
			return
		}
		if r.URL.Path == "/api/v1/beep_proposals" && r.Method == http.MethodPost {
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(&client.BeepProposal{
				Title:       "check database backup",
				Kind:        "once",
				RunAt:       "2026-09-15T18:00:00Z",
				Timezone:    "UTC",
				Confirmable: true,
			})
			return
		}
		http.NotFound(w, r)
	})
	defer cleanup()

	flagNoInteractive = true
	defer func() { flagNoInteractive = false }()

	createCmd := findBeepCmd(t, "create")

	// 1. Instant reminder (no schedule flags)
	_, err := captureStdout(func() error {
		return createCmd.RunE(createCmd, []string{"Instant Test"})
	})
	if err != nil {
		t.Fatalf("beep create instant failed: %v", err)
	}
	if receivedBody.Title != "Instant Test" || receivedBody.Kind != "once" || receivedBody.RunAt == "" {
		t.Errorf("unexpected instant create request: %+v", receivedBody)
	}
	if receivedBody.Timezone == "" {
		t.Errorf("expected auto-detected timezone to be set on instant create, got empty")
	}

	// 2. Delay with --in 15m
	createCmd.Flags().Set("in", "15m")
	_, err = captureStdout(func() error {
		return createCmd.RunE(createCmd, []string{"Delay Test"})
	})
	createCmd.Flags().Set("in", "")
	if err != nil {
		t.Fatalf("beep create with --in failed: %v", err)
	}
	if receivedBody.Title != "Delay Test" || receivedBody.Kind != "once" || receivedBody.RunAt == "" {
		t.Errorf("unexpected delay create request: %+v", receivedBody)
	}
	if receivedBody.Timezone == "" {
		t.Errorf("expected auto-detected timezone to be set on delay create, got empty")
	}

	// 3. Recurring with --cron
	createCmd.Flags().Set("cron", "*/5 * * * *")
	_, err = captureStdout(func() error {
		return createCmd.RunE(createCmd, []string{"Cron Test"})
	})
	createCmd.Flags().Set("cron", "")
	if err != nil {
		t.Fatalf("beep create with --cron failed: %v", err)
	}
	if receivedBody.Title != "Cron Test" || receivedBody.Kind != "recurring" || receivedBody.Cron != "*/5 * * * *" {
		t.Errorf("unexpected cron create request: %+v", receivedBody)
	}

	// 4. Natural create with --json, merging --body and --channels
	createCmd.Flags().Set("natural", "check database backup")
	createCmd.Flags().Set("body", "Custom body")
	createCmd.Flags().Set("channels", "slack,email")
	flagJSON = true
	_, err = captureStdout(func() error {
		return createCmd.RunE(createCmd, nil)
	})
	createCmd.Flags().Set("natural", "")
	createCmd.Flags().Set("body", "")
	createCmd.Flags().Set("channels", "")
	flagJSON = false
	if err != nil {
		t.Fatalf("natural create with --json failed: %v", err)
	}
	if receivedBody.Body != "Custom body" || len(receivedBody.NotificationChannels) != 2 {
		t.Errorf("expected merged body and channels in natural create, got: %+v", receivedBody)
	}
}

func TestBeepCreateMutuallyExclusiveFlags(t *testing.T) {
	defer resetCmdFlags(findBeepCmd(t, "create"))

	for _, args := range [][]string{
		{"beep", "create", "Conflict Test", "--cron", "0 * * * *", "--in", "10m"},
		{"create", "Conflict Test", "--cron", "0 * * * *", "--in", "10m"},
	} {
		cmd := RootCmd
		cmd.SetArgs(args)
		err := cmd.Execute()
		if err == nil {
			t.Fatalf("expected mutually exclusive error for args %v, got nil", args)
		}
		if !strings.Contains(err.Error(), "none of the others can be") && !strings.Contains(err.Error(), "mutually exclusive") {
			t.Errorf("unexpected error message for args %v: %v", args, err)
		}
	}
}

func TestBeepActionsCommands(t *testing.T) {
	var pausedID, resumedID, runID, deletedID string

	_, cleanup := setupBeepTestEnv(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/me" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"identity": map[string]any{"id": "id_1", "email": "test@example.com"},
			})
			return
		}
		if r.URL.Path == "/api/v1/beeps/b1/pause" && r.Method == http.MethodPost {
			pausedID = "b1"
			_ = json.NewEncoder(w).Encode(&client.Beep{ID: "b1", Status: "paused"})
			return
		}
		if r.URL.Path == "/api/v1/beeps/b1/pause" && r.Method == http.MethodDelete {
			resumedID = "b1"
			_ = json.NewEncoder(w).Encode(&client.Beep{ID: "b1", Status: "active"})
			return
		}
		if r.URL.Path == "/api/v1/beeps/b1/runs" && r.Method == http.MethodPost {
			runID = "b1"
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(&client.BeepRun{ID: "run_99", Status: "firing"})
			return
		}
		if r.URL.Path == "/api/v1/beeps/b1" && r.Method == http.MethodDelete {
			deletedID = "b1"
			w.WriteHeader(http.StatusNoContent)
			return
		}
		http.NotFound(w, r)
	})
	defer cleanup()

	pauseCmd := findBeepCmd(t, "pause")
	resumeCmd := findBeepCmd(t, "resume")
	runCmd := findBeepCmd(t, "run")
	deleteCmd := findBeepCmd(t, "delete")

	// Pause
	if _, err := captureStdout(func() error { return pauseCmd.RunE(pauseCmd, []string{"b1"}) }); err != nil {
		t.Fatalf("pause failed: %v", err)
	}
	if pausedID != "b1" {
		t.Errorf("expected pause on b1")
	}

	// Resume
	if _, err := captureStdout(func() error { return resumeCmd.RunE(resumeCmd, []string{"b1"}) }); err != nil {
		t.Fatalf("resume failed: %v", err)
	}
	if resumedID != "b1" {
		t.Errorf("expected resume on b1")
	}

	// Run
	if _, err := captureStdout(func() error { return runCmd.RunE(runCmd, []string{"b1"}) }); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if runID != "b1" {
		t.Errorf("expected run on b1")
	}

	// Delete
	if _, err := captureStdout(func() error { return deleteCmd.RunE(deleteCmd, []string{"b1"}) }); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if deletedID != "b1" {
		t.Errorf("expected delete on b1")
	}
}

func TestBeepOmittedIDNonInteractive(t *testing.T) {
	_, cleanup := setupBeepTestEnv(t, func(w http.ResponseWriter, r *http.Request) {
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
		findBeepCmd(t, "show"),
		findBeepCmd(t, "delete"),
		findBeepCmd(t, "pause"),
		findBeepCmd(t, "resume"),
		findBeepCmd(t, "run"),
	}
	for _, c := range commands {
		err := c.RunE(c, nil)
		if err == nil {
			t.Errorf("expected error for command %s without ID in non-interactive mode", c.Name())
		} else if !strings.Contains(err.Error(), "beep ID is required") {
			t.Errorf("expected 'beep ID is required' error for %s, got: %v", c.Name(), err)
		}
	}
}

func TestBeepDualPathExecutionParity(t *testing.T) {
	mockBeep := &client.Beep{
		ID:        "beep_parity_1",
		Title:     "Parity Check",
		Status:    "active",
		Kind:      "recurring",
		Cron:      "0 8 * * *",
		Timezone:  "Asia/Shanghai",
		NextRunAt: "2026-09-17T08:00:00+08:00",
	}

	_, cleanup := setupBeepTestEnv(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/me" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"identity": map[string]any{"id": "id_1", "email": "test@example.com"},
			})
			return
		}
		if r.URL.Path == "/api/v1/beeps" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"beeps": []*client.Beep{mockBeep},
			})
			return
		}
		if r.URL.Path == "/api/v1/beeps/beep_parity_1" {
			_ = json.NewEncoder(w).Encode(mockBeep)
			return
		}
		http.NotFound(w, r)
	})
	defer cleanup()

	// 1. List - Plain table parity
	topList, _, _ := RootCmd.Find([]string{"list"})
	legacyList, _, _ := RootCmd.Find([]string{"beep", "list"})

	outTop, err := captureStdout(func() error { return topList.RunE(topList, nil) })
	if err != nil {
		t.Fatalf("top-level list failed: %v", err)
	}
	outLegacy, err := captureStdout(func() error { return legacyList.RunE(legacyList, nil) })
	if err != nil {
		t.Fatalf("legacy beep list failed: %v", err)
	}
	if outTop != outLegacy {
		t.Errorf("list output mismatch between top-level and legacy:\nTop:\n%s\nLegacy:\n%s", outTop, outLegacy)
	}

	// 2. List - JSON parity
	flagJSON = true
	jsonTop, err := captureStdout(func() error { return topList.RunE(topList, nil) })
	if err != nil {
		t.Fatalf("top-level list --json failed: %v", err)
	}
	jsonLegacy, err := captureStdout(func() error { return legacyList.RunE(legacyList, nil) })
	if err != nil {
		t.Fatalf("legacy beep list --json failed: %v", err)
	}
	flagJSON = false
	if jsonTop != jsonLegacy {
		t.Errorf("list --json output mismatch:\nTop:\n%s\nLegacy:\n%s", jsonTop, jsonLegacy)
	}

	// 3. Show - Parity
	topShow, _, _ := RootCmd.Find([]string{"show"})
	legacyShow, _, _ := RootCmd.Find([]string{"beep", "show"})

	showTop, err := captureStdout(func() error { return topShow.RunE(topShow, []string{"beep_parity_1"}) })
	if err != nil {
		t.Fatalf("top-level show failed: %v", err)
	}
	showLegacy, err := captureStdout(func() error { return legacyShow.RunE(legacyShow, []string{"beep_parity_1"}) })
	if err != nil {
		t.Fatalf("legacy beep show failed: %v", err)
	}
	if showTop != showLegacy {
		t.Errorf("show output mismatch:\nTop:\n%s\nLegacy:\n%s", showTop, showLegacy)
	}
}

func TestNoConflictWithRunnerAndOtherCommands(t *testing.T) {
	// Top-level 'run' resolves to beep run
	runCmd, _, err := RootCmd.Find([]string{"run"})
	if err != nil {
		t.Fatalf("failed to find 'run' command: %v", err)
	}
	if runCmd.Name() != "run" {
		t.Errorf("expected 'run' to resolve to command name 'run', got %q", runCmd.Name())
	}
	if runCmd.GroupID != "beeps" {
		t.Errorf("expected 'run' to have GroupID 'beeps', got %q", runCmd.GroupID)
	}

	// 'runner up' resolves to start
	runnerUpCmd, _, err := RootCmd.Find([]string{"runner", "up"})
	if err != nil {
		t.Fatalf("failed to find 'runner up': %v", err)
	}
	if runnerUpCmd.Name() != "start" {
		t.Errorf("expected 'runner up' to resolve to 'start', got %q", runnerUpCmd.Name())
	}

	// Regression checks for secondary namespaces and subcommands
	beeperListCmd, _, err := RootCmd.Find([]string{"beeper", "list"})
	if err != nil {
		t.Fatalf("failed to find 'beeper list': %v", err)
	}
	if beeperListCmd.Name() != "list" {
		t.Errorf("expected 'beeper list' name to be 'list', got %q", beeperListCmd.Name())
	}

	configShowCmd, _, err := RootCmd.Find([]string{"config", "show"})
	if err != nil {
		t.Fatalf("failed to find 'config show': %v", err)
	}
	if configShowCmd.Name() != "show" {
		t.Errorf("expected 'config show' name to be 'show', got %q", configShowCmd.Name())
	}

	runnerJobListCmd, _, err := RootCmd.Find([]string{"runner", "job", "list"})
	if err != nil {
		t.Fatalf("failed to find 'runner job list': %v", err)
	}
	if runnerJobListCmd.Name() != "list" {
		t.Errorf("expected 'runner job list' name to be 'list', got %q", runnerJobListCmd.Name())
	}
}

func TestNaturalCreateChannelsPriority(t *testing.T) {
	var (
		lastProposalChannels []string
		lastCreatedChannels  []string
	)

	_, cleanup := setupBeepTestEnv(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/me" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"identity": map[string]any{"id": "id_1", "email": "test@example.com"},
			})
			return
		}
		if r.URL.Path == "/api/v1/beep_proposals" && r.Method == http.MethodPost {
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"intent":                "create",
				"title":                 "Test Beep",
				"notification_channels": lastProposalChannels,
				"confirmable":           true,
			})
			return
		}
		if r.URL.Path == "/api/v1/beeps" && r.Method == http.MethodPost {
			var req client.CreateBeepRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			lastCreatedChannels = req.NotificationChannels
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(&client.Beep{
				ID:                   "beep_new",
				Title:                req.Title,
				NotificationChannels: req.NotificationChannels,
			})
			return
		}
		http.NotFound(w, r)
	})
	defer cleanup()

	flagNoInteractive = true
	flagJSON = true
	createCmd := findBeepCmd(t, "create")
	defer func() {
		flagNoInteractive = false
		flagJSON = false
		resetCmdFlags(createCmd)
	}()

	runCreate := func(args ...string) error {
		resetCmdFlags(createCmd)
		if err := createCmd.ParseFlags(args); err != nil {
			return err
		}
		_, err := captureStdout(func() error {
			return createCmd.RunE(createCmd, createCmd.Flags().Args())
		})
		return err
	}

	// 1. Natural create uses proposal channels when no --channels flag
	lastProposalChannels = []string{"web_push"}
	lastCreatedChannels = nil
	if err := runCreate("-n", "只通知给 web push 其他不需要"); err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if len(lastCreatedChannels) != 1 || lastCreatedChannels[0] != "web_push" {
		t.Errorf("expected [web_push], got %v", lastCreatedChannels)
	}

	// 2. --channels flag overrides proposal channels
	lastProposalChannels = []string{"web_push"}
	lastCreatedChannels = nil
	if err := runCreate("-n", "只通知给 web push", "--channels", "email"); err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if len(lastCreatedChannels) != 1 || lastCreatedChannels[0] != "email" {
		t.Errorf("expected [email], got %v", lastCreatedChannels)
	}

	// 3. When proposal has no channels and no flag, channels remain empty in non-interactive mode
	lastProposalChannels = nil
	lastCreatedChannels = nil
	if err := runCreate("-n", "明天打电话给妈"); err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if len(lastCreatedChannels) != 0 {
		t.Errorf("expected empty channels, got %v", lastCreatedChannels)
	}
}
