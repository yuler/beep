package exec

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"beep/internal/client"
)

func TestDispatchOnBeepHook(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "beep-hook-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	hookDir := filepath.Join(tmpDir, ".beep", "hooks")
	if err := os.MkdirAll(hookDir, 0o755); err != nil {
		t.Fatalf("failed to create hook dir: %v", err)
	}

	hookPath := filepath.Join(hookDir, "on_beep_fired")
	hookScript := `#!/bin/sh
echo "Hook received event: $BEEP_EVENT source: $BEEP_EVENT_SOURCE intent: $BEEP_EVENT_INTENT id: $BEEP_EVENT_ID title: $BEEP_EVENT_TITLE"
`
	if err := os.WriteFile(hookPath, []byte(hookScript), 0o755); err != nil {
		t.Fatalf("failed to write hook script: %v", err)
	}

	delivery := client.CliDelivery{
		ID: "del_12345",
		Payload: map[string]any{
			"event":  "beep.fired",
			"source": "runner_job",
			"intent": "get_off_work",
			"title":  "Off work notification",
			"metadata": map[string]any{
				"action_hint": "get_off_work",
			},
		},
	}

	out, err := DispatchOnBeepHook(context.Background(), tmpDir, delivery)
	if err != nil {
		t.Fatalf("DispatchOnBeepHook returned unexpected error: %v", err)
	}

	if !strings.Contains(out, "event: beep.fired") {
		t.Errorf("expected hook output to contain event name, got: %s", out)
	}
	if !strings.Contains(out, "source: runner_job") {
		t.Errorf("expected hook output to contain source, got: %s", out)
	}
	if !strings.Contains(out, "intent: get_off_work") {
		t.Errorf("expected hook output to contain intent, got: %s", out)
	}
	if !strings.Contains(out, "id: del_12345") {
		t.Errorf("expected hook output to contain delivery ID, got: %s", out)
	}
	if !strings.Contains(out, "title: Off work notification") {
		t.Errorf("expected hook output to contain title, got: %s", out)
	}
}
