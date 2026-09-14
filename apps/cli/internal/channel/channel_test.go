package channel

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"beep/internal/config"
	"beep/internal/workspace"
)

func TestChannelPollInbox(t *testing.T) {
	var (
		inboxCalls atomic.Int32
		ackCalls   atomic.Int32
	)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/api/v1/channels/cli/inbox":
			inboxCalls.Add(1)
			deliveries := []map[string]any{
				{
					"id": "del-1",
					"payload": map[string]any{
						"title": "Alert 1",
						"body":  "Test body",
					},
				},
			}
			json.NewEncoder(w).Encode(map[string]any{
				"deliveries": deliveries,
			})
		case r.URL.Path == "/api/v1/channels/cli/deliveries/del-1/ack":
			ackCalls.Add(1)
			json.NewEncoder(w).Encode(map[string]any{
				"status": "ok",
			})
		default:
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer ts.Close()

	root := t.TempDir()
	ws, err := workspace.Open(root)
	if err != nil {
		t.Fatal(err)
	}

	ch := New(&config.Config{
		ServerURL:    ts.URL,
		ChannelToken: "beep_ct_test",
		PollInterval: time.Second,
	}, ws)

	ch.PollInbox(context.Background())

	if inboxCalls.Load() != 1 {
		t.Fatalf("expected 1 inbox call, got %d", inboxCalls.Load())
	}
	if ackCalls.Load() != 1 {
		t.Fatalf("expected 1 ack call, got %d", ackCalls.Load())
	}
}

func TestChannelTokenFallback(t *testing.T) {
	ch1 := New(&config.Config{ChannelToken: "token1"}, nil)
	if ch1.Token() != "token1" {
		t.Errorf("expected token1, got %s", ch1.Token())
	}

	ch2 := New(&config.Config{CliToken: "token2"}, nil)
	if ch2.Token() != "token2" {
		t.Errorf("expected token2, got %s", ch2.Token())
	}

	ch3 := New(&config.Config{DeviceToken: "token3"}, nil)
	if ch3.Token() != "token3" {
		t.Errorf("expected token3, got %s", ch3.Token())
	}
}
