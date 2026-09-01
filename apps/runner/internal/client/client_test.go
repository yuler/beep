package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"beep-runner/internal/config"
	"beep-runner/internal/probe"
)

func TestClientPing(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/runner/ping" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("X-Runner-Token") != "beep_rt_test" {
			t.Errorf("unexpected token header: %s", r.Header.Get("X-Runner-Token"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"status":      "ok",
			"runner_id":   "run_123",
			"runner_name": "My-Runner",
			"server_time": "2026-09-01T12:00:00Z",
		})
	}))
	defer ts.Close()

	cfg := &config.Config{
		ServerURL:   ts.URL,
		RunnerToken: "beep_rt_test",
	}

	c := New(cfg)
	res, err := c.Ping(context.Background())
	if err != nil {
		t.Fatalf("unexpected ping error: %v", err)
	}

	if res.RunnerID != "run_123" {
		t.Errorf("expected runner_id run_123, got %s", res.RunnerID)
	}
}

func TestClientPollAndReportResult(t *testing.T) {
	polled := false
	resultReported := false

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/api/v1/runner/tasks/poll" {
			polled = true
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]any{
				"task": map[string]any{
					"id":              "task-uuid-1",
					"beeper_id":       "beeper-uuid-1",
					"title":           "Site Check",
					"app_slug":        "site-uptime",
					"config":          map[string]any{"target_url": "https://example.com"},
					"timeout_seconds": 15,
				},
			})
			return
		}

		if r.URL.Path == "/api/v1/runner/tasks/task-uuid-1/result" {
			resultReported = true
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]any{
				"status": "acknowledged",
			})
			return
		}

		t.Errorf("unexpected path: %s", r.URL.Path)
	}))
	defer ts.Close()

	cfg := &config.Config{
		ServerURL:   ts.URL,
		RunnerToken: "beep_rt_test",
	}

	c := New(cfg)
	task, err := c.Poll(context.Background())
	if err != nil {
		t.Fatalf("unexpected poll error: %v", err)
	}
	if task == nil || task.ID != "task-uuid-1" {
		t.Fatalf("expected task id task-uuid-1, got %v", task)
	}

	sig := probe.OkSignal("HTTP 200 OK", "all good", map[string]any{"latency_ms": 25})
	if err := c.ReportResult(context.Background(), task.ID, sig); err != nil {
		t.Fatalf("unexpected report result error: %v", err)
	}

	if !polled || !resultReported {
		t.Errorf("expected both poll and report result to be executed")
	}
}
