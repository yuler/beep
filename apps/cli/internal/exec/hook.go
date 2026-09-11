package exec

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"beep/internal/client"
	"beep/internal/proc"
)

func FindOnBeepHook(workspaceRoot string) string {
	candidates := []string{
		filepath.Join(workspaceRoot, ".beep", "hooks", "on_beep"),
		filepath.Join(workspaceRoot, "hooks", "on_beep"),
		filepath.Join(workspaceRoot, ".beep", "hooks", "on-beep"),
		filepath.Join(workspaceRoot, "hooks", "on-beep"),
	}

	for _, c := range candidates {
		info, err := os.Stat(c)
		if err == nil && !info.IsDir() {
			return c
		}
	}
	return ""
}

func DispatchOnBeepHook(ctx context.Context, workspaceRoot string, delivery client.CliDelivery) (string, error) {
	hookPath := FindOnBeepHook(workspaceRoot)
	if hookPath == "" {
		return "", nil
	}

	payloadBytes, _ := json.Marshal(delivery.Payload)
	title, _ := delivery.Payload["title"].(string)
	actionHint := ""
	if metadata, ok := delivery.Payload["metadata"].(map[string]any); ok {
		if hint, ok := metadata["action_hint"].(string); ok {
			actionHint = hint
		}
	}

	env := os.Environ()
	env = append(env,
		fmt.Sprintf("BEEP_EVENT_JSON=%s", string(payloadBytes)),
		fmt.Sprintf("BEEP_EVENT_ID=%s", delivery.ID),
		fmt.Sprintf("BEEP_EVENT_TITLE=%s", title),
		fmt.Sprintf("BEEP_EVENT_ACTION_HINT=%s", actionHint),
	)

	execCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(execCtx, hookPath)
	cmd.Dir = workspaceRoot
	cmd.Env = env
	proc.Setpgid(cmd)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("hook %s failed: %w (output: %s)", hookPath, err, string(out))
	}
	return string(out), nil
}
