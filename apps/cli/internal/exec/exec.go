package exec

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"beep/internal/proc"
	"beep/internal/task"
)

type JobExecutor struct{}

func NewJobExecutor() *JobExecutor {
	return &JobExecutor{}
}

func (e *JobExecutor) Run(ctx context.Context, argv []string, env []string, timeout time.Duration, onLog func(string)) *task.Result {
	if len(argv) == 0 {
		return task.Error("Missing command", "workspace resolved an empty command", nil)
	}

	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(execCtx, argv[0], argv[1:]...)
	cmd.Env = env
	proc.Setpgid(cmd)
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return nil
		}
		return proc.Kill(cmd.Process, cmd.Process.Pid)
	}
	// Bound how long Wait may block on I/O pipes after the process exits
	// (orphaned children keeping them open). Mirror of the old 2s grace.
	cmd.WaitDelay = 2 * time.Second

	// Stdout/Stderr as writers make Wait wait for all output to be copied
	// before returning, so no log lines can be lost to a pipe-close race.
	stdout := &lineWriter{onLog: onLog}
	stderr := &lineWriter{onLog: onLog}
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	start := time.Now()
	if err := cmd.Start(); err != nil {
		return task.Error("Failed to start job", err.Error(), nil)
	}

	waitErr := cmd.Wait()

	// Ensure the whole process group is reaped (children of the job script).
	if cmd.Process != nil {
		_ = proc.Kill(cmd.Process, cmd.Process.Pid)
	}

	stdout.flush()
	stderr.flush()

	durationMs := time.Since(start).Milliseconds()
	metrics := map[string]any{"duration_ms": durationMs}

	if execCtx.Err() == context.DeadlineExceeded || ctx.Err() != nil {
		metrics["timed_out"] = execCtx.Err() == context.DeadlineExceeded
		if ctx.Err() != nil && execCtx.Err() != context.DeadlineExceeded {
			return task.Error("Job cancelled", "Runner shut down before the job finished", metrics)
		}
		return task.Error(fmt.Sprintf("Job timed out after %s", timeout), "Execution exceeded the deadline", metrics)
	}

	// Wait returns ErrWaitDelay when it force-closed the I/O pipes of an
	// otherwise successful process (e.g. a daemonized grandchild holding them).
	if errors.Is(waitErr, exec.ErrWaitDelay) {
		waitErr = nil
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

// lineWriter splits process output into lines and forwards each to onLog.
type lineWriter struct {
	mu    sync.Mutex
	buf   []byte
	onLog func(string)
}

func (w *lineWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	w.buf = append(w.buf, p...)
	for {
		i := bytes.IndexByte(w.buf, '\n')
		if i < 0 {
			break
		}
		w.emit(w.buf[:i+1])
		w.buf = w.buf[i+1:]
	}
	w.mu.Unlock()
	return len(p), nil
}

// flush emits any trailing partial line left after the process exited.
func (w *lineWriter) flush() {
	w.mu.Lock()
	if len(w.buf) > 0 {
		w.emit(append(w.buf, '\n'))
		w.buf = nil
	}
	w.mu.Unlock()
}

func (w *lineWriter) emit(line []byte) {
	if w.onLog != nil {
		w.onLog(string(line))
	}
}

// blockedJobEnvKeys must never be inherited by job processes.
var blockedJobEnvKeys = map[string]struct{}{
	"BEEP_RUNNER_TOKEN": {},
}

func WithJobEnv(extra []string) []string {
	merged := make(map[string]string)
	var order []string

	set := func(item string) {
		key, val, found := strings.Cut(item, "=")
		if !found {
			return
		}
		if _, blocked := blockedJobEnvKeys[key]; blocked {
			return
		}
		if _, exists := merged[key]; !exists {
			order = append(order, key)
		}
		merged[key] = val
	}

	for _, item := range os.Environ() {
		set(item)
	}
	for _, item := range extra {
		set(item)
	}

	out := make([]string, 0, len(order))
	for _, k := range order {
		out = append(out, fmt.Sprintf("%s=%s", k, merged[k]))
	}
	return out
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
