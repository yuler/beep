package cliservice

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"

	"beep/internal/config"
	"beep/internal/daemon"
	"beep/internal/ui"
)

func TestCopyPrefixedLines(t *testing.T) {
	ui.SetEnabled(false)
	var buf bytes.Buffer
	if err := CopyPrefixedLines("runner", strings.NewReader("hello\nworld\n"), &buf); err != nil {
		t.Fatalf("CopyPrefixedLines: %v", err)
	}
	got := buf.String()
	want := "runner | hello\nrunner | world\n"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestForegroundCommandIsAttachedChild(t *testing.T) {
	cfg := &config.Config{
		Workspace:   t.TempDir(),
		RunnerToken: "beep_rt_test",
	}
	cmd, err := ForegroundCommand(daemon.ServiceRunner, []string{"service", "start", "-w", cfg.Workspace}, cfg)
	if err != nil {
		t.Fatalf("ForegroundCommand: %v", err)
	}
	for _, e := range cmd.Env {
		if e == "BEEP_DAEMON_CHILD=1" {
			t.Fatal("foreground child must not set BEEP_DAEMON_CHILD (needs stdout+file logs)")
		}
		if e == "BEEP_RUNNER_TOKEN=beep_rt_test" {
			goto haveToken
		}
	}
	t.Fatal("expected BEEP_RUNNER_TOKEN in child env")
haveToken:
	if cmd.SysProcAttr != nil && cmd.SysProcAttr.Setsid {
		t.Fatal("foreground child must not Detach/Setsid")
	}
	args := cmd.Args[1:]
	if len(args) < 2 || args[0] != "runner" || args[1] != "up" {
		t.Fatalf("expected child args to start with runner up, got %v", args)
	}
}

func TestRunForegroundServicesPrefixesOutput(t *testing.T) {
	ui.SetEnabled(false)
	orig := SpawnForegroundFn
	t.Cleanup(func() { SpawnForegroundFn = orig })

	SpawnForegroundFn = func(service string, rawArgs []string, cfg *config.Config) (*exec.Cmd, error) {
		return exec.Command("echo", "ready-"+service), nil
	}

	var buf bytes.Buffer
	cfg := &config.Config{Workspace: t.TempDir()}
	err := runForegroundServices(cfg, []string{daemon.ServiceRunner, daemon.ServiceChannel}, nil, &buf)
	if err != nil {
		t.Fatalf("runForegroundServices: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "runner | ready-runner") {
		t.Errorf("missing runner line in %q", out)
	}
	if !strings.Contains(out, "channel | ready-channel") {
		t.Errorf("missing channel line in %q", out)
	}
}

func TestRunForegroundServicesKillsStartedOnSpawnError(t *testing.T) {
	orig := SpawnForegroundFn
	t.Cleanup(func() { SpawnForegroundFn = orig })

	cmd := exec.Command("sleep", "30")
	SpawnForegroundFn = func(service string, rawArgs []string, cfg *config.Config) (*exec.Cmd, error) {
		if service == daemon.ServiceChannel {
			return nil, fmt.Errorf("channel spawn failed")
		}
		return cmd, nil
	}

	cfg := &config.Config{Workspace: t.TempDir()}
	err := runForegroundServices(cfg, []string{daemon.ServiceRunner, daemon.ServiceChannel}, nil, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "channel spawn failed") {
		t.Fatalf("expected channel spawn error, got %v", err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if cmd.Process == nil {
			return
		}
		if err := cmd.Process.Signal(syscall.Signal(0)); err != nil {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("expected first child to be killed after later spawn failure")
}

func TestRunForegroundServicesKillsOthersWhenOneExits(t *testing.T) {
	orig := SpawnForegroundFn
	t.Cleanup(func() { SpawnForegroundFn = orig })

	SpawnForegroundFn = func(service string, rawArgs []string, cfg *config.Config) (*exec.Cmd, error) {
		if service == daemon.ServiceRunner {
			return exec.Command("sh", "-c", "exit 7"), nil
		}
		return exec.Command("sleep", "30"), nil
	}

	cfg := &config.Config{Workspace: t.TempDir()}
	err := runForegroundServices(cfg, []string{daemon.ServiceRunner, daemon.ServiceChannel}, nil, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error when a child exits non-zero")
	}
}

func TestForegroundCommandMissingExecutable(t *testing.T) {
	orig := osExecutable
	t.Cleanup(func() { osExecutable = orig })
	osExecutable = func() (string, error) { return "", os.ErrNotExist }

	_, err := ForegroundCommand(daemon.ServiceRunner, nil, &config.Config{Workspace: t.TempDir()})
	if err == nil {
		t.Fatal("expected error when executable cannot be resolved")
	}
}
