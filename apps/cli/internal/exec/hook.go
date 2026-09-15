package exec

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"beep/internal/client"
	"beep/internal/proc"
)

func FindHook(workspaceRoot string) string {
	return FindHookForEvent(workspaceRoot, "")
}

func FindHookForEvent(workspaceRoot string, _ string) string {
	candidates := hookPaths(workspaceRoot, "on-channel")

	for _, c := range candidates {
		info, err := os.Stat(c)
		if err == nil && !info.IsDir() {
			return c
		}
	}
	return ""
}

func hookPaths(workspaceRoot string, names ...string) []string {
	var paths []string
	for _, name := range names {
		paths = append(paths,
			filepath.Join(workspaceRoot, "hooks", name),
			filepath.Join(workspaceRoot, ".beep", "hooks", name),
		)
	}
	return paths
}

func DispatchHook(ctx context.Context, workspaceRoot string, delivery client.CliDelivery) (string, string, error) {
	if strings.TrimSpace(workspaceRoot) == "" {
		return "", "", nil
	}
	eventName, _ := delivery.Payload["event"].(string)
	if eventName == "" {
		eventName = "beep.fired"
	}

	hookPath := FindHookForEvent(workspaceRoot, eventName)
	if hookPath == "" {
		return "", "", nil
	}
	hookName := filepath.Base(hookPath)

	if !isSafeHook(hookPath) {
		return "", hookName, fmt.Errorf("hook %s failed: unsafe permissions or ownership", hookPath)
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
		return string(out), hookName, fmt.Errorf("hook %s failed: %w (output: %s)", hookPath, err, truncateHookOutput(string(out)))
	}
	return string(out), hookName, nil
}

func truncateHookOutput(s string) string {
	const maxLen = 2048
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[len(runes)-maxLen:])
}

func isSafeHook(path string) bool {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}
	if runtime.GOOS == "windows" {
		return true
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
