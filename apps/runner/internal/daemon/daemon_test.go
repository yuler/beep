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

	"beep-runner/internal/config"
	"beep-runner/internal/task"
	"beep-runner/internal/workspace"
)

func TestJobEnvOmitsRunnerToken(t *testing.T) {
	d := &Daemon{cfg: &config.Config{
		ServerURL:   "https://core.example.com",
		RunnerToken: "beep_runner_secret",
	}}
	env := d.jobEnv(&task.Task{
		ID:        "run-1",
		JobSlug:   "check",
		LogURL:    "https://core.example.com/api/v1/runner/tasks/run-1/logs",
		ResultURL: "https://core.example.com/api/v1/runner/tasks/run-1/result",
		Config:    map[string]any{"k": "v"},
	})
	for _, item := range env {
		if strings.Contains(item, "beep_runner_secret") {
			t.Fatalf("job env must not include runner token, got %s", item)
		}
	}
}

func TestPollAndExecuteFillsConcurrency(t *testing.T) {
	var polls atomic.Int32
	var tsURL string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/api/v1/runner/tasks/poll":
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
			json.NewEncoder(w).Encode(map[string]any{"status": "acknowledged"})
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
		RunnerToken:  "beep_runner_test",
		Concurrency:  2,
		PollInterval: time.Second,
	}, ws)

	d.pollAndExecute(context.Background())
	if got := polls.Load(); got != 2 {
		t.Fatalf("expected 2 polls to fill concurrency, got %d", got)
	}
}
