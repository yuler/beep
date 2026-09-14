package exec

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"beep/internal/client"
	"beep/internal/proc"
)

func FindHook(workspaceRoot string, event string) string {
	candidates := []string{}
	if event == "beep.fired" || event == "beep_fired" || event == "" {
		candidates = append(candidates,
			filepath.Join(workspaceRoot, ".beep", "hooks", "on_beep_fired"),
			filepath.Join(workspaceRoot, "hooks", "on_beep_fired"),
			filepath.Join(workspaceRoot, ".beep", "hooks", "on-beep-fired"),
			filepath.Join(workspaceRoot, "hooks", "on-beep-fired"),
		)
	}

	// Fallback to legacy/generic on_beep hook
	candidates = append(candidates,
		filepath.Join(workspaceRoot, ".beep", "hooks", "on_beep"),
		filepath.Join(workspaceRoot, "hooks", "on_beep"),
		filepath.Join(workspaceRoot, ".beep", "hooks", "on-beep"),
		filepath.Join(workspaceRoot, "hooks", "on-beep"),
	)

	for _, c := range candidates {
		info, err := os.Stat(c)
		if err == nil && !info.IsDir() {
			return c
		}
	}
	return ""
}

func FindOnBeepHook(workspaceRoot string) string {
	return FindHook(workspaceRoot, "beep.fired")
}

func DispatchOnBeepHook(ctx context.Context, workspaceRoot string, delivery client.CliDelivery) (string, error) {
	if strings.TrimSpace(workspaceRoot) == "" {
		return "", nil
	}
	eventName, _ := delivery.Payload["event"].(string)
	if eventName == "" {
		eventName = "beep.fired"
	}

	hookPath := FindHook(workspaceRoot, eventName)
	if hookPath == "" {
		return "", nil
	}

	if !isSafeHook(hookPath) {
		return "", fmt.Errorf("hook %s failed: unsafe permissions or ownership", hookPath)
	}

	payloadBytes, _ := json.Marshal(delivery.Payload)
	title, _ := delivery.Payload["title"].(string)

	source, _ := delivery.Payload["source"].(string)
	if source == "" {
		source = "beep"
	}
	sourceType, _ := delivery.Payload["source_type"].(string)
	sourceID, _ := delivery.Payload["source_id"].(string)
	intent, _ := delivery.Payload["intent"].(string)

	env := os.Environ()
	env = append(env,
		fmt.Sprintf("BEEP_EVENT=%s", eventName),
		fmt.Sprintf("BEEP_EVENT_SOURCE=%s", source),
		fmt.Sprintf("BEEP_EVENT_SOURCE_TYPE=%s", sourceType),
		fmt.Sprintf("BEEP_EVENT_SOURCE_ID=%s", sourceID),
		fmt.Sprintf("BEEP_EVENT_INTENT=%s", intent),
		fmt.Sprintf("BEEP_EVENT_JSON=%s", string(payloadBytes)),
		fmt.Sprintf("BEEP_EVENT_ID=%s", delivery.ID),
		fmt.Sprintf("BEEP_EVENT_TITLE=%s", title),
	)

	execCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(execCtx, hookPath)
	cmd.Dir = workspaceRoot
	cmd.Env = env
	proc.Setpgid(cmd)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("hook %s failed: %w (output: %s)", hookPath, err, truncateHookOutput(string(out)))
	}
	return string(out), nil
}

func truncateHookOutput(s string) string {
	const maxLen = 2048
	if len(s) <= maxLen {
		return s
	}
	return s[len(s)-maxLen:]
}

func isSafeHook(path string) bool {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}
	if info.Mode().Perm()&0o022 != 0 {
		return false
	}
	if info.Mode().Perm()&0o111 == 0 {
		return false
	}
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		if int(stat.Uid) != os.Getuid() {
			return false
		}
	}
	return true
}
