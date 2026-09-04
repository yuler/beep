package exec

import (
	"context"
	"testing"
	"time"

	"beep-runner/internal/task"
)

func TestJobExecutorDisabled(t *testing.T) {
	result := NewJobExecutor(false).Run(context.Background(), []string{"true"}, nil, time.Second, nil)
	if result.Status != task.StatusError {
		t.Fatalf("expected error, got %s", result.Status)
	}
}

func TestJobExecutorSuccess(t *testing.T) {
	var logs string
	result := NewJobExecutor(true).Run(context.Background(), []string{"echo", "hello-workspace"}, nil, 5*time.Second, func(line string) {
		logs += line
	})
	if result.Status != task.StatusOk {
		t.Fatalf("expected ok, got %s %s", result.Status, result.Message)
	}
	if logs == "" {
		t.Fatalf("expected captured stdout")
	}
}

func TestJobExecutorNonZeroExit(t *testing.T) {
	result := NewJobExecutor(true).Run(context.Background(), []string{"/bin/sh", "-c", "exit 2"}, nil, 5*time.Second, nil)
	if result.Status != task.StatusAlerting {
		t.Fatalf("expected alerting, got %s", result.Status)
	}
	if result.Metrics["exit_code"] != 2 {
		t.Errorf("expected exit 2, got %v", result.Metrics["exit_code"])
	}
}
