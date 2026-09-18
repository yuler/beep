package autostart

import (
	"strings"
	"testing"
)

func TestCaptureEnv(t *testing.T) {
	t.Setenv("PATH", "/usr/local/bin:/usr/bin")
	t.Setenv("HOME", "/home/testuser")
	t.Setenv("BEEP_RUNNER_TOKEN", "test_token_123")
	t.Setenv("BEEP_DAEMON_CHILD", "1") // should be filtered out
	t.Setenv("OTHER_VAR", "ignored")

	env := CaptureEnv()
	if env["PATH"] != "/usr/local/bin:/usr/bin" {
		t.Errorf("expected PATH in env, got %q", env["PATH"])
	}
	if env["HOME"] != "/home/testuser" {
		t.Errorf("expected HOME in env, got %q", env["HOME"])
	}
	if env["BEEP_RUNNER_TOKEN"] != "test_token_123" {
		t.Errorf("expected BEEP_RUNNER_TOKEN in env, got %q", env["BEEP_RUNNER_TOKEN"])
	}
	if _, ok := env["BEEP_DAEMON_CHILD"]; ok {
		t.Errorf("expected BEEP_DAEMON_CHILD to be excluded from captured env")
	}
	if _, ok := env["OTHER_VAR"]; ok {
		t.Errorf("expected OTHER_VAR to be excluded from captured env")
	}
}

func TestSystemdUnitGeneration(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := &SystemdManager{
		userDir: tmpDir,
	}

	info := ServiceInfo{
		Service:     "runner",
		BinaryName:  "beep",
		Description: "Beep Runner Daemon",
		ExecPath:    "/usr/local/bin/beep",
		Args:        []string{"runner", "start", "--workspace", "/path with spaces"},
		Workspace:   "/home/user/.beep",
		Env: map[string]string{
			"PATH": "/usr/local/bin:/bin",
			"HOME": "/home/user",
		},
	}

	unitName := mgr.UnitName(info.Service, info.BinaryName)
	if unitName != "beep-runner.service" {
		t.Fatalf("expected unit name beep-runner.service, got %q", unitName)
	}

	// Test template rendering logic directly
	var cmdParts []string
	cmdParts = append(cmdParts, systemdQuote(info.ExecPath))
	for _, arg := range info.Args {
		cmdParts = append(cmdParts, systemdQuote(arg))
	}
	execCmd := strings.Join(cmdParts, " ")

	if !strings.Contains(execCmd, `--workspace "/path with spaces"`) {
		t.Errorf("expected quoted spaces in execCmd, got: %s", execCmd)
	}
}

func TestLaunchdPlistGeneration(t *testing.T) {
	mgr := &LaunchdManager{}
	label := mgr.Label("channel", "beep")
	if label != "com.beep.channel" {
		t.Fatalf("expected com.beep.channel, got %q", label)
	}

	plistPath := mgr.plistPath("channel", "beep")
	if !strings.HasSuffix(plistPath, "com.beep.channel.plist") {
		t.Fatalf("expected plist path to end with com.beep.channel.plist, got %q", plistPath)
	}
}

func TestUnsupportedManager(t *testing.T) {
	mgr := NewUnsupportedManager()
	if mgr.IsSupported() {
		t.Fatal("expected unsupported manager to return false for IsSupported")
	}
	err := mgr.Install(ServiceInfo{})
	if err != ErrUnsupported {
		t.Fatalf("expected ErrUnsupported, got %v", err)
	}
	st := mgr.GetStatus("runner", "beep")
	if st.Supported {
		t.Fatal("expected status Supported to be false")
	}
}

func TestMockManagerSelection(t *testing.T) {
	orig := DefaultManager
	defer func() { DefaultManager = orig }()

	mock := NewUnsupportedManager()
	DefaultManager = mock

	cur := CurrentManager()
	if cur != mock {
		t.Fatalf("expected CurrentManager to return DefaultManager when set")
	}
}
