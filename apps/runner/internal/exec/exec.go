package exec

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"beep-runner/internal/task"
)

type JobExecutor struct {
	allowExec bool
}

func NewJobExecutor(allowExec bool) *JobExecutor {
	return &JobExecutor{allowExec: allowExec}
}

func (e *JobExecutor) Run(ctx context.Context, argv []string, env []string, timeout time.Duration, onLog func(string)) *task.Result {
	if !e.allowExec {
		return task.Error(
			"Execution disabled",
			"Start the runner with --allow-exec (or BEEP_ALLOW_EXEC=1) to run workspace scripts.",
			map[string]any{"allowed": false},
		)
	}
	if len(argv) == 0 {
		return task.Error("Missing command", "workspace resolved an empty command", nil)
	}

	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(execCtx, argv[0], argv[1:]...)
	cmd.Env = env
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return task.Error("Failed to start job", err.Error(), nil)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return task.Error("Failed to start job", err.Error(), nil)
	}

	start := time.Now()
	if err := cmd.Start(); err != nil {
		return task.Error("Failed to start job", err.Error(), nil)
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go streamLines(stdout, onLog, &wg)
	go streamLines(stderr, onLog, &wg)

	waitErr := cmd.Wait()
	wg.Wait()
	durationMs := time.Since(start).Milliseconds()
	metrics := map[string]any{"duration_ms": durationMs}

	if execCtx.Err() == context.DeadlineExceeded {
		metrics["timed_out"] = true
		return task.Alerting(fmt.Sprintf("Job timed out after %s", timeout), "Execution exceeded the deadline", metrics)
	}

	if waitErr != nil {
		exitCode := 1
		if exitErr, ok := waitErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
		metrics["exit_code"] = exitCode
		return task.Alerting(fmt.Sprintf("Job failed (exit %d)", exitCode), waitErr.Error(), metrics)
	}

	metrics["exit_code"] = 0
	return task.Ok(fmt.Sprintf("Job succeeded (%dms)", durationMs), "Command exited 0", metrics)
}

func streamLines(r io.Reader, onLog func(string), wg *sync.WaitGroup) {
	defer wg.Done()
	scanner := bufio.NewScanner(r)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if onLog != nil {
			onLog(line + "\n")
		}
	}
}

func WithJobEnv(extra []string) []string {
	env := os.Environ()
	if len(extra) == 0 {
		return env
	}
	return append(env, extra...)
}

func ConfigEnv(config map[string]any) []string {
	if config == nil {
		return nil
	}
	var out []string
	for key, val := range config {
		name := strings.ToUpper(strings.ReplaceAll(key, "-", "_"))
		out = append(out, fmt.Sprintf("BEEP_CONFIG_%s=%v", name, val))
	}
	return out
}
