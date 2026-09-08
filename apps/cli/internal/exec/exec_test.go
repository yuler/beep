package exec

import (
	"context"
	"sync"
	"testing"
	"time"

	"beep/internal/task"
)

func TestJobExecutorMissingCommand(t *testing.T) {
	result := NewJobExecutor().Run(context.Background(), []string{}, nil, time.Second, nil)
	if result.Status != task.StatusError {
		t.Fatalf("expected error, got %s", result.Status)
	}
}

func TestJobExecutorSuccess(t *testing.T) {
	var mu sync.Mutex
	var logs string
	result := NewJobExecutor().Run(context.Background(), []string{"echo", "hello-workspace"}, nil, 5*time.Second, func(line string) {
		mu.Lock()
		defer mu.Unlock()
		logs += line
	})
	if result.Status != task.StatusOk {
		t.Fatalf("expected ok, got %s %s", result.Status, result.Message)
	}
	mu.Lock()
	captured := logs
	mu.Unlock()
	if captured == "" {
		t.Fatalf("expected captured stdout")
	}
}

func TestJobExecutorCapturesPartialLine(t *testing.T) {
	var mu sync.Mutex
	var logs string
	result := NewJobExecutor().Run(context.Background(), []string{"/bin/sh", "-c", "printf 'no-newline-tail'"}, nil, 5*time.Second, func(line string) {
		mu.Lock()
		defer mu.Unlock()
		logs += line
	})
	if result.Status != task.StatusOk {
		t.Fatalf("expected ok, got %s %s", result.Status, result.Message)
	}
	mu.Lock()
	captured := logs
	mu.Unlock()
	if captured != "no-newline-tail\n" {
		t.Fatalf("expected trailing partial line, got %q", captured)
	}
}

func TestJobExecutorCapturesLargeOutput(t *testing.T) {
	var mu sync.Mutex
	var lines int
	result := NewJobExecutor().Run(context.Background(), []string{"/bin/sh", "-c", "seq 1 20000"}, nil, 5*time.Second, func(line string) {
		mu.Lock()
		defer mu.Unlock()
		lines++
	})
	if result.Status != task.StatusOk {
		t.Fatalf("expected ok, got %s %s", result.Status, result.Message)
	}
	mu.Lock()
	captured := lines
	mu.Unlock()
	if captured != 20000 {
		t.Fatalf("expected 20000 lines, got %d", captured)
	}
}

func TestJobExecutorNonZeroExit(t *testing.T) {
	result := NewJobExecutor().Run(context.Background(), []string{"/bin/sh", "-c", "exit 2"}, nil, 5*time.Second, nil)
	if result.Status != task.StatusAlerting {
		t.Fatalf("expected alerting, got %s", result.Status)
	}
	if result.Metrics["exit_code"] != 2 {
		t.Errorf("expected exit 2, got %v", result.Metrics["exit_code"])
	}
}
