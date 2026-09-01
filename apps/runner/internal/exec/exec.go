package exec

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"beep-runner/internal/probe"
)

type ScriptExecutor struct {
	allowExec bool
}

func NewScriptExecutor(allowExec bool) *ScriptExecutor {
	return &ScriptExecutor{
		allowExec: allowExec,
	}
}

func (e *ScriptExecutor) Run(ctx context.Context, config map[string]any) *probe.Signal {
	if !e.allowExec {
		return probe.ErrorSignal(
			"Command execution disabled",
			"Local command execution is disabled on this runner. Start the runner with --allow-exec (or BEEP_ALLOW_EXEC=1) to enable.",
			map[string]any{"allowed": false},
		)
	}

	command, _ := config["command"].(string)
	if command == "" {
		if script, _ := config["script"].(string); script != "" {
			command = script
		}
	}
	if strings.TrimSpace(command) == "" {
		return probe.ErrorSignal("Missing command", "command or script config parameter is required", nil)
	}

	timeoutSec := 10
	if val, ok := config["timeout_seconds"]; ok {
		if v, ok := val.(float64); ok && v > 0 {
			timeoutSec = int(v)
		}
	}

	execCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSec)*time.Second)
	defer cancel()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(execCtx, "cmd", "/C", command)
	} else {
		cmd = exec.CommandContext(execCtx, "sh", "-c", command)
	}

	cmd.Env = os.Environ()

	var outBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &outBuf

	start := time.Now()
	err := cmd.Run()
	durationMs := time.Since(start).Milliseconds()

	outputBytes := outBuf.Bytes()
	if len(outputBytes) > 8192 {
		outputBytes = outputBytes[:8192]
	}
	outputStr := strings.TrimSpace(string(outputBytes))

	metrics := map[string]any{
		"duration_ms": durationMs,
	}

	if execCtx.Err() == context.DeadlineExceeded {
		metrics["timed_out"] = true
		return probe.AlertingSignal(
			fmt.Sprintf("Command timed out after %ds", timeoutSec),
			fmt.Sprintf("Execution exceeded %d seconds deadline. Output:\n%s", timeoutSec, outputStr),
			metrics,
		)
	}

	if err != nil {
		exitCode := 1
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
		metrics["exit_code"] = exitCode

		title := fmt.Sprintf("Command failed (exit code %d)", exitCode)
		message := outputStr
		if message == "" {
			message = err.Error()
		}

		return probe.AlertingSignal(title, message, metrics)
	}

	metrics["exit_code"] = 0
	title := fmt.Sprintf("Command succeeded (%dms)", durationMs)
	message := outputStr
	if message == "" {
		message = "Command executed successfully with zero output"
	}

	return probe.OkSignal(title, message, metrics)
}
