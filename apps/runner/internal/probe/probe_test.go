package probe

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPProbeSuccess(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("healthy payload"))
	}))
	defer ts.Close()

	probe := NewHTTPProbe()
	ctx := context.Background()

	sig := probe.Run(ctx, map[string]any{
		"target_url":    ts.URL,
		"body_contains": "healthy",
	})

	if sig.Status != StatusOk {
		t.Fatalf("expected status ok, got %s: %s", sig.Status, sig.Message)
	}
	if sig.Metrics["status_code"] != http.StatusOK {
		t.Errorf("expected status_code 200, got %v", sig.Metrics["status_code"])
	}
}

func TestHTTPProbeStatusMismatch(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	probe := NewHTTPProbe()
	ctx := context.Background()

	sig := probe.Run(ctx, map[string]any{
		"target_url":      ts.URL,
		"expected_status": 200,
	})

	if sig.Status != StatusAlerting {
		t.Fatalf("expected status alerting, got %s", sig.Status)
	}
}

func TestDispatcherRouting(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	d := NewDispatcher(func(ctx context.Context, config map[string]any) *Signal {
		return OkSignal("script ok", "done", nil)
	})

	ctx := context.Background()

	// Test site-uptime routing
	httpTask := &RunnerTask{
		ID:      "task-1",
		AppSlug: "site-uptime",
		Config: map[string]any{
			"target_url": ts.URL,
		},
	}
	sig := d.Dispatch(ctx, httpTask)
	if sig.Status != StatusOk {
		t.Errorf("expected ok for http task, got %s", sig.Status)
	}

	// Test custom-script routing
	execTask := &RunnerTask{
		ID:      "task-2",
		AppSlug: "custom-script",
		Config: map[string]any{
			"command": "echo ok",
		},
	}
	sig = d.Dispatch(ctx, execTask)
	if sig.Status != StatusOk || sig.Title != "script ok" {
		t.Errorf("expected script ok, got %s: %s", sig.Status, sig.Title)
	}
}
