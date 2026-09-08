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
	"syscall"
	"time"

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
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
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

	streamDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(streamDone)
	}()

	waitDone := make(chan error, 1)
	go func() {
		waitDone <- cmd.Wait()
	}()

	var waitErr error
	select {
	case waitErr = <-waitDone:
		// Main process has exited. Wait for remaining output in pipes with a short safety timeout.
		select {
		case <-streamDone:
		case <-time.After(2 * time.Second):
			_ = stdout.Close()
			_ = stderr.Close()
			<-streamDone
		}
	case <-execCtx.Done():
		// Context timed out or cancelled: close pipes and reap process.
		_ = stdout.Close()
		_ = stderr.Close()
		<-streamDone
		waitErr = <-waitDone
	}

	// Ensure the whole process group is reaped (children of the job script).
	if cmd.Process != nil && cmd.Process.Pid > 1 {
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}

	durationMs := time.Since(start).Milliseconds()
	metrics := map[string]any{"duration_ms": durationMs}

	if execCtx.Err() == context.DeadlineExceeded || ctx.Err() != nil {
		metrics["timed_out"] = execCtx.Err() == context.DeadlineExceeded
		if ctx.Err() != nil && execCtx.Err() != context.DeadlineExceeded {
			return task.Error("Job cancelled", "Runner shut down before the job finished", metrics)
		}
		return task.Error(fmt.Sprintf("Job timed out after %s", timeout), "Execution exceeded the deadline", metrics)
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

// blockedJobEnvKeys must never be inherited by job processes.
var blockedJobEnvKeys = map[string]struct{}{
	"BEEP_RUNNER_TOKEN": {},
}

func WithJobEnv(extra []string) []string {
	env := scrubJobEnv(os.Environ())
	if len(extra) == 0 {
		return env
	}
	return append(env, scrubJobEnv(extra)...)
}

func scrubJobEnv(env []string) []string {
	out := make([]string, 0, len(env))
	for _, item := range env {
		key, _, _ := strings.Cut(item, "=")
		if _, blocked := blockedJobEnvKeys[key]; blocked {
			continue
		}
		out = append(out, item)
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
