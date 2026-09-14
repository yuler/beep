package exec

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"beep/internal/client"
)

func TestDispatchHookRunsOnChannelForBeepFired(t *testing.T) {
	root := t.TempDir()
	writeHook(t, root, "on_channel", `#!/bin/sh
echo "on_channel event=$BEEP_EVENT id=$BEEP_EVENT_ID"
`)

	out, name, err := DispatchHook(context.Background(), root, client.CliDelivery{
		ID: "del_beep",
		Payload: map[string]any{
			"event": "beep.fired",
			"title": "Off work",
		},
	})
	if err != nil {
		t.Fatalf("DispatchHook: %v", err)
	}
	if name != "on_channel" {
		t.Fatalf("hook name = %q, want on_channel", name)
	}
	if !strings.Contains(out, "on_channel event=beep.fired") {
		t.Fatalf("output %q", out)
	}
}

func TestDispatchHookRunsOnChannelForChannelTest(t *testing.T) {
	root := t.TempDir()
	writeHook(t, root, "on_channel", `#!/bin/sh
echo "on_channel event=$BEEP_EVENT"
`)

	out, name, err := DispatchHook(context.Background(), root, client.CliDelivery{
		ID: "del_test",
		Payload: map[string]any{
			"event": "channel.test",
			"title": "Test notification",
		},
	})
	if err != nil {
		t.Fatalf("DispatchHook: %v", err)
	}
	if name != "on_channel" {
		t.Fatalf("hook name = %q, want on_channel", name)
	}
	if !strings.Contains(out, "on_channel event=channel.test") {
		t.Fatalf("output %q", out)
	}
}

func TestDispatchHookPrefersOnChannelOverLegacyOnBeep(t *testing.T) {
	root := t.TempDir()
	writeHook(t, root, "on_beep", `#!/bin/sh
echo "legacy on_beep"
`)
	writeHook(t, root, "on_channel", `#!/bin/sh
echo "on_channel wins"
`)

	out, name, err := DispatchHook(context.Background(), root, client.CliDelivery{
		ID:      "del_both",
		Payload: map[string]any{"event": "beep.fired"},
	})
	if err != nil {
		t.Fatalf("DispatchHook: %v", err)
	}
	if name != "on_channel" {
		t.Fatalf("hook name = %q, want on_channel", name)
	}
	if !strings.Contains(out, "on_channel wins") {
		t.Fatalf("output %q", out)
	}
	if strings.Contains(out, "legacy on_beep") {
		t.Fatalf("legacy on_beep ran: %q", out)
	}
}

func TestDispatchHookFallsBackToLegacyOnBeepFired(t *testing.T) {
	root := t.TempDir()
	writeHook(t, root, "on_beep_fired", `#!/bin/sh
echo "legacy fired"
`)

	out, name, err := DispatchHook(context.Background(), root, client.CliDelivery{
		ID:      "del_legacy",
		Payload: map[string]any{"event": "beep.fired"},
	})
	if err != nil {
		t.Fatalf("DispatchHook: %v", err)
	}
	if name != "on_channel" {
		t.Fatalf("hook name = %q, want on_channel", name)
	}
	if !strings.Contains(out, "legacy fired") {
		t.Fatalf("output %q", out)
	}
}

func TestDispatchHookMissingReturnsEmptyName(t *testing.T) {
	root := t.TempDir()

	out, name, err := DispatchHook(context.Background(), root, client.CliDelivery{
		ID:      "del_none",
		Payload: map[string]any{"event": "beep.fired"},
	})
	if err != nil {
		t.Fatalf("DispatchHook: %v", err)
	}
	if name != "" || out != "" {
		t.Fatalf("name=%q out=%q, want empty", name, out)
	}
}

func writeHook(t *testing.T, workspaceRoot, name, script string) {
	t.Helper()
	dir := filepath.Join(workspaceRoot, "hooks")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
}
