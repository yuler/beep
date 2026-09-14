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
	writeHook(t, root, "on-channel", `#!/bin/sh
echo "on-channel event=$BEEP_EVENT id=$BEEP_EVENT_ID"
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
	if name != "on-channel" {
		t.Fatalf("hook name = %q, want on-channel", name)
	}
	if !strings.Contains(out, "on-channel event=beep.fired") {
		t.Fatalf("output %q", out)
	}
}

func TestDispatchHookRunsOnChannelForChannelTest(t *testing.T) {
	root := t.TempDir()
	writeHook(t, root, "on-channel", `#!/bin/sh
echo "on-channel event=$BEEP_EVENT"
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
	if name != "on-channel" {
		t.Fatalf("hook name = %q, want on-channel", name)
	}
	if !strings.Contains(out, "on-channel event=channel.test") {
		t.Fatalf("output %q", out)
	}
}

func TestDispatchHookIgnoresLegacyOnBeep(t *testing.T) {
	root := t.TempDir()
	writeHook(t, root, "on_beep", `#!/bin/sh
echo "legacy on_beep"
`)

	out, name, err := DispatchHook(context.Background(), root, client.CliDelivery{
		ID:      "del_legacy",
		Payload: map[string]any{"event": "beep.fired"},
	})
	if err != nil {
		t.Fatalf("DispatchHook: %v", err)
	}
	if name != "" || out != "" {
		t.Fatalf("name=%q out=%q, want empty when only legacy on_beep hook exists", name, out)
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

func TestDispatchHookEmptyRootReturnsEmpty(t *testing.T) {
	out, name, err := DispatchHook(context.Background(), "  ", client.CliDelivery{
		ID:      "del_empty",
		Payload: map[string]any{"event": "beep.fired"},
	})
	if err != nil {
		t.Fatalf("DispatchHook: %v", err)
	}
	if name != "" || out != "" {
		t.Fatalf("name=%q out=%q, want empty for blank workspace root", name, out)
	}
}

func TestDispatchHookRejectsGroupWritableHook(t *testing.T) {
	root := t.TempDir()
	path := writeHook(t, root, "on-channel", "#!/bin/sh\necho hi\n")
	if err := os.Chmod(path, 0o775); err != nil {
		t.Fatal(err)
	}

	_, name, err := DispatchHook(context.Background(), root, client.CliDelivery{
		ID:      "del_unsafe",
		Payload: map[string]any{"event": "beep.fired"},
	})
	if err == nil {
		t.Fatal("expected unsafe-permissions error, got nil")
	}
	if name != "on-channel" {
		t.Fatalf("hook name = %q, want on-channel", name)
	}
}

func TestDispatchHookRejectsNonExecutableHook(t *testing.T) {
	root := t.TempDir()
	path := writeHook(t, root, "on-channel", "#!/bin/sh\necho hi\n")
	if err := os.Chmod(path, 0o644); err != nil {
		t.Fatal(err)
	}

	_, _, err := DispatchHook(context.Background(), root, client.CliDelivery{
		ID:      "del_noexec",
		Payload: map[string]any{"event": "beep.fired"},
	})
	if err == nil {
		t.Fatal("expected non-executable error, got nil")
	}
}

func TestTruncateHookOutputKeepsRuneBoundary(t *testing.T) {
	long := strings.Repeat("通知", 1500) // 3000 runes
	got := truncateHookOutput(long)
	if gotLen := len([]rune(got)); gotLen != 2048 {
		t.Fatalf("truncated to %d runes, want 2048", gotLen)
	}
	if !strings.HasSuffix(got, "通知") {
		t.Fatalf("truncation split a multi-byte rune: %q", got[len(got)-10:])
	}
	if short := truncateHookOutput("ok"); short != "ok" {
		t.Fatalf("short output changed: %q", short)
	}
}

func writeHook(t *testing.T, workspaceRoot, name, script string) string {
	t.Helper()
	dir := filepath.Join(workspaceRoot, "hooks")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}
