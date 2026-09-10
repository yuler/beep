package daemon

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"beep/internal/config"
	"beep/internal/task"
	"beep/internal/workspace"
)

func TestJobEnvOmitsRunnerToken(t *testing.T) {
	t.Setenv("BEEP_RUNNER_TOKEN", "beep_rt_from_environ")
	d := &Daemon{cfg: &config.Config{
		ServerURL:   "https://core.example.com",
		RunnerToken: "beep_rt_secret",
	}}
	env := d.jobEnv(&task.Task{
		ID:        "run-1",
		JobSlug:   "check",
		LogURL:    "https://core.example.com/api/v1/runner/tasks/run-1/logs",
		ResultURL: "https://core.example.com/api/v1/runner/tasks/run-1/result",
		Config:    map[string]any{"k": "v"},
	})
	for _, item := range env {
		if strings.HasPrefix(item, "BEEP_RUNNER_TOKEN=") {
			t.Fatalf("job env must not include BEEP_RUNNER_TOKEN, got %s", item)
		}
		if strings.Contains(item, "beep_rt_secret") || strings.Contains(item, "beep_rt_from_environ") {
			t.Fatalf("job env must not include runner token, got %s", item)
		}
	}
}

func TestJobEnvPrecedence(t *testing.T) {
	root := t.TempDir()
	ws, err := workspace.Open(root)
	if err != nil {
		t.Fatal(err)
	}

	// Base host environment must be overridden by the workspace .env.
	t.Setenv("API_KEY", "from_host_env")

	envFile := filepath.Join(root, ".env")
	envContent := `
API_KEY=from_workspace_env
OVERRIDDEN_BY_LOCAL=from_env
OVERRIDDEN_BY_SERVER=from_workspace_env
BEEP_CONFIG_OVERRIDDEN_BY_SERVER=from_workspace_env
BEEP_RUN_ID=from_workspace_env
`
	if err := os.WriteFile(envFile, []byte(envContent), 0o600); err != nil {
		t.Fatal(err)
	}

	// .env.local overrides .env.
	envLocalContent := `
OVERRIDDEN_BY_LOCAL=from_local
`
	if err := os.WriteFile(filepath.Join(root, ".env.local"), []byte(envLocalContent), 0o600); err != nil {
		t.Fatal(err)
	}

	d := &Daemon{
		cfg: &config.Config{
			ServerURL: "https://core.example.com",
		},
		workspace: ws,
	}

	job := &task.Task{
		ID:        "run-123",
		JobSlug:   "custom-job",
		LogURL:    "https://core.example.com/api/v1/runner/tasks/run-123/logs",
		ResultURL: "https://core.example.com/api/v1/runner/tasks/run-123/result",
		Config: map[string]any{
			"overridden_by_server": "from_server_config",
		},
	}

	env := d.jobEnv(job)
	envMap := make(map[string]string)
	for _, item := range env {
		k, v, _ := strings.Cut(item, "=")
		envMap[k] = v
	}

	tests := []struct {
		key  string
		want string
	}{
		{"API_KEY", "from_workspace_env"},                          // workspace .env overrides host env
		{"OVERRIDDEN_BY_LOCAL", "from_local"},                      // .env.local overrides .env
		{"OVERRIDDEN_BY_SERVER", "from_workspace_env"},             // non-colliding .env var survives
		{"BEEP_CONFIG_OVERRIDDEN_BY_SERVER", "from_server_config"}, // server config overrides .env
		{"BEEP_RUN_ID", "run-123"},                                 // runtime context overrides .env
		{"BEEP_JOB_SLUG", "custom-job"},
	}
	for _, tc := range tests {
		if envMap[tc.key] != tc.want {
			t.Errorf("expected %s=%q, got %q", tc.key, tc.want, envMap[tc.key])
		}
	}
}

func TestPollAndExecuteFillsConcurrency(t *testing.T) {
	var (
		polls atomic.Int32
		pings atomic.Int32
	)
	var tsURL string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/api/v1/runner/ping":
			pings.Add(1)
			json.NewEncoder(w).Encode(map[string]any{
				"status":      "ok",
				"runner_id":   "test-runner",
				"runner_name": "Test Runner",
				"server_time": "2026-09-01T12:00:00Z",
			})
		case r.URL.Path == "/api/v1/runner/tasks":
			n := polls.Add(1)
			json.NewEncoder(w).Encode(map[string]any{
				"task": map[string]any{
					"id":              "task-" + strconv.Itoa(int(n)),
					"job_slug":        "hold",
					"name":            "Hold",
					"timeout_seconds": 30,
					"log_url":         tsURL + "/api/v1/runner/tasks/task-" + strconv.Itoa(int(n)) + "/logs",
					"result_url":      tsURL + "/api/v1/runner/tasks/task-" + strconv.Itoa(int(n)) + "/result",
				},
			})
		default:
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer ts.Close()
	tsURL = ts.URL

	root := t.TempDir()
	jobsDir := filepath.Join(root, "jobs")
	if err := os.MkdirAll(jobsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(jobsDir, "hold")
	if err := os.WriteFile(script, []byte("#!/bin/sh\nsleep 30\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	ws, err := workspace.Open(root)
	if err != nil {
		t.Fatal(err)
	}

	d := New(&config.Config{
		ServerURL:    ts.URL,
		RunnerToken:  "beep_rt_test",
		Concurrency:  2,
		PollInterval: time.Second,
	}, ws)

	d.pollAndExecute(context.Background())
	if got := polls.Load(); got != 2 {
		t.Fatalf("expected 2 polls to fill concurrency, got %d", got)
	}
	if got := pings.Load(); got != 1 {
		t.Fatalf("expected 1 ping when saturated, got %d", got)
	}
}

func TestExecuteReportsLogsAndResult(t *testing.T) {
	var (
		gotLogs   strings.Builder
		gotResult *task.Result
	)
	var tsURL string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/runner/tasks/task-run-1/logs":
			var req struct {
				Chunk string `json:"chunk"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
				gotLogs.WriteString(req.Chunk)
			}
			w.WriteHeader(http.StatusNoContent)
		case "/api/v1/runner/tasks/task-run-1/result":
			var res task.Result
			if err := json.NewDecoder(r.Body).Decode(&res); err == nil {
				gotResult = &res
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()
	tsURL = ts.URL

	root := t.TempDir()
	jobsDir := filepath.Join(root, "jobs")
	if err := os.MkdirAll(jobsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(jobsDir, "slow-check")
	if err := os.WriteFile(script, []byte("#!/bin/sh\necho \"step 1\"\necho \"step 2 completed\"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	ws, err := workspace.Open(root)
	if err != nil {
		t.Fatal(err)
	}

	d := New(&config.Config{
		ServerURL:   tsURL,
		RunnerToken: "beep_rt_test",
	}, ws)

	job := &task.Task{
		ID:             "task-run-1",
		JobSlug:        "slow-check",
		Name:           "Slow Check",
		TimeoutSeconds: 30,
		LogURL:         tsURL + "/api/v1/runner/tasks/task-run-1/logs",
		ResultURL:      tsURL + "/api/v1/runner/tasks/task-run-1/result",
	}

	d.execute(context.Background(), job)

	if !strings.Contains(gotLogs.String(), "step 1") || !strings.Contains(gotLogs.String(), "step 2 completed") {
		t.Fatalf("expected logs to contain step outputs, got: %q", gotLogs.String())
	}
	if gotResult == nil {
		t.Fatal("expected result report to be called, but got nil")
	}
	if gotResult.Status != task.StatusOk {
		t.Fatalf("expected result status 'ok', got %q (%s)", gotResult.Status, gotResult.Title)
	}
}
