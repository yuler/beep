package daemon

import (
	"os"
	"testing"
)

func TestAcquireSocket(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "beep-sock-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// First acquire should succeed
	sl, err := AcquireSocket(tempDir)
	if err != nil {
		t.Fatalf("expected AcquireSocket to succeed, got error: %v", err)
	}

	// Second acquire should fail with already running error
	_, err2 := AcquireSocket(tempDir)
	if err2 == nil {
		t.Fatalf("expected second AcquireSocket to fail")
	}

	running, pid, err := CheckRunning(tempDir)
	if err != nil || !running {
		t.Fatalf("expected CheckRunning to return true, got running=%v, pid=%d, err=%v", running, pid, err)
	}
	if pid != os.Getpid() {
		t.Errorf("expected PID %d, got %d", os.Getpid(), pid)
	}

	// Close the first listener
	if err := sl.Close(); err != nil {
		t.Fatalf("failed to close socket listener: %v", err)
	}

	// After closing, CheckRunning should be false
	runningAfter, _, _ := CheckRunning(tempDir)
	if runningAfter {
		t.Fatalf("expected CheckRunning to return false after close")
	}

	// Now third acquire should succeed again
	sl2, err := AcquireSocket(tempDir)
	if err != nil {
		t.Fatalf("expected third AcquireSocket to succeed, got: %v", err)
	}
	_ = sl2.Close()
}

func TestGetDaemonStatusAndStopDaemon(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "beep-sock-status-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// When not running, GetDaemonStatus returns nil, nil
	status, err := GetDaemonStatus(tempDir)
	if err != nil || status != nil {
		t.Fatalf("expected nil status when daemon not running, got %v, err: %v", status, err)
	}

	sl, err := AcquireSocket(tempDir)
	if err != nil {
		t.Fatalf("failed to acquire socket: %v", err)
	}
	defer sl.Close()

	status, err = GetDaemonStatus(tempDir)
	if err != nil {
		t.Fatalf("failed to get status: %v", err)
	}
	if status == nil || status.PID != os.Getpid() {
		t.Fatalf("expected status PID %d, got %v", os.Getpid(), status)
	}
}
