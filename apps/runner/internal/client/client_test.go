package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"beep-runner/internal/config"
	"beep-runner/internal/task"
)

func TestClientPing(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/runner/ping" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"status":      "ok",
			"runner_id":   "run_123",
			"runner_name": "My-Runner",
			"server_time": "2026-09-01T12:00:00Z",
		})
	}))
	defer ts.Close()

	c := New(&config.Config{ServerURL: ts.URL, RunnerToken: "beep_runner_test"})
	res, err := c.Ping(context.Background())
	if err != nil {
		t.Fatalf("unexpected ping error: %v", err)
	}
	if res.RunnerID != "run_123" {
		t.Errorf("expected runner_id run_123, got %s", res.RunnerID)
	}
}

func TestClientPollLogAndResult(t *testing.T) {
	var gotLog, gotResult bool
	var tsURL string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/runner/tasks/poll":
			json.NewEncoder(w).Encode(map[string]any{
				"task": map[string]any{
					"id":              "task-1",
					"job_slug":        "intranet-http",
					"name":            "Intranet HTTP",
					"timeout_seconds": 15,
					"log_url":         tsURL + "/api/v1/runner/tasks/task-1/logs",
					"result_url":      tsURL + "/api/v1/runner/tasks/task-1/result",
					"config":          map[string]any{"target_url": "http://10.0.0.5"},
				},
			})
		case "/api/v1/runner/tasks/task-1/logs":
			gotLog = true
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]any{"status": "acknowledged"})
		case "/api/v1/runner/tasks/task-1/result":
			gotResult = true
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]any{"status": "acknowledged"})
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer ts.Close()
	tsURL = ts.URL

	c := New(&config.Config{ServerURL: ts.URL, RunnerToken: "beep_runner_test"})
	job, err := c.Poll(context.Background())
	if err != nil || job == nil || job.JobSlug != "intranet-http" {
		t.Fatalf("poll: %v %#v", err, job)
	}

	if err := c.ReportLog(context.Background(), job.LogURL, "hello\n"); err != nil {
		t.Fatal(err)
	}
	if err := c.ReportResult(context.Background(), job.ResultURL, task.Ok("ok", "", nil)); err != nil {
		t.Fatal(err)
	}
	if !gotLog || !gotResult {
		t.Fatalf("expected log and result posts, got log=%v result=%v", gotLog, gotResult)
	}
}

func TestReportRejectsForeignCallbackURL(t *testing.T) {
	c := New(&config.Config{ServerURL: "https://core.example.com", RunnerToken: "beep_runner_test"})
	if err := c.ReportLog(context.Background(), "https://evil.example/api/v1/runner/tasks/x/logs", "leak\n"); err == nil {
		t.Fatal("expected foreign log URL to be rejected")
	}
	if err := c.ReportResult(context.Background(), "https://core.example.com/not-a-task", task.Ok("ok", "", nil)); err == nil {
		t.Fatal("expected non-task path to be rejected")
	}
}
