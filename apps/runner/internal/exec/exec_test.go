package exec

import (
	"context"
	"testing"

	"beep-runner/internal/probe"
)

func TestScriptExecutorDisabled(t *testing.T) {
	executor := NewScriptExecutor(false)
	ctx := context.Background()

	sig := executor.Run(ctx, map[string]any{
		"command": "echo hello",
	})

	if sig.Status != probe.StatusError {
		t.Fatalf("expected error status when allow_exec is false, got %s", sig.Status)
	}
}

func TestScriptExecutorSuccess(t *testing.T) {
	executor := NewScriptExecutor(true)
	ctx := context.Background()

	sig := executor.Run(ctx, map[string]any{
		"command": "echo 'test execution success'",
	})

	if sig.Status != probe.StatusOk {
		t.Fatalf("expected ok status, got %s: %s", sig.Status, sig.Message)
	}
	if sig.Metrics["exit_code"] != 0 {
		t.Errorf("expected exit code 0, got %v", sig.Metrics["exit_code"])
	}
	if sig.Message != "'test execution success'" && sig.Message != "test execution success" {
		t.Errorf("unexpected output: %s", sig.Message)
	}
}

func TestScriptExecutorNonZeroExit(t *testing.T) {
	executor := NewScriptExecutor(true)
	ctx := context.Background()

	sig := executor.Run(ctx, map[string]any{
		"command": "exit 2",
	})

	if sig.Status != probe.StatusAlerting {
		t.Fatalf("expected alerting status on non-zero exit, got %s", sig.Status)
	}
	if sig.Metrics["exit_code"] != 2 {
		t.Errorf("expected exit code 2, got %v", sig.Metrics["exit_code"])
	}
}
