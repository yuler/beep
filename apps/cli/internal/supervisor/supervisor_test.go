package supervisor

import (
	"os"
	"path/filepath"
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
	if env["BEEP_RUNNER_TOKEN"] != "" {
		t.Errorf("expected BEEP_RUNNER_TOKEN not to be captured into supervisor env")
	}
	if _, ok := env["BEEP_DAEMON_CHILD"]; ok {
		t.Errorf("expected BEEP_DAEMON_CHILD to be excluded from captured env")
	}
	if _, ok := env["OTHER_VAR"]; ok {
		t.Errorf("expected OTHER_VAR to be excluded from captured env")
	}
}

func TestSystemdUnitGeneration(t *testing.T) {
	mgr := &SystemdManager{}
	info := ServiceInfo{
		Service:     "runner",
		BinaryName:  "beep",
		Description: "Beep Runner Daemon",
		ExecPath:    "/usr/local/bin/beep",
		Args:        []string{"runner", "start", "--workspace", "/path with spaces"},
		Workspace:   "/home/user/My Workspace",
		Env: map[string]string{
			"PATH":         "/usr/local/bin:/bin",
			"HOME":         "/home/user",
			"BEEP_QUOTE":   `say "hi"`,
			"BEEP_DOLLAR":  "$HOME/token",
			"BEEP_NEWLINE": "a\nb",
			"BEEP_BAD=KEY": "nope",
		},
	}

	unitName := mgr.UnitName(info.Service, info.BinaryName)
	if unitName != "beep-runner.service" {
		t.Fatalf("expected unit name beep-runner.service, got %q", unitName)
	}

	unit, err := renderSystemdUnit(info)
	if err != nil {
		t.Fatalf("renderSystemdUnit: %v", err)
	}
	if !strings.Contains(unit, `ExecStart=/usr/local/bin/beep runner start --workspace "/path with spaces"`) {
		t.Errorf("expected quoted spaces in ExecStart, got:\n%s", unit)
	}
	if !strings.Contains(unit, `WorkingDirectory="/home/user/My Workspace"`) {
		t.Errorf("expected quoted WorkingDirectory, got:\n%s", unit)
	}
	if !strings.Contains(unit, `Environment="BEEP_QUOTE=say \"hi\""`) {
		t.Errorf("expected escaped quotes in Environment, got:\n%s", unit)
	}
	if !strings.Contains(unit, `Environment="BEEP_DOLLAR=$$HOME/token"`) {
		t.Errorf("expected escaped $ in Environment, got:\n%s", unit)
	}
	if strings.Contains(unit, "BEEP_NEWLINE") {
		t.Errorf("expected newline env values to be omitted, got:\n%s", unit)
	}
	if strings.Contains(unit, "BEEP_BAD=KEY") {
		t.Errorf("expected invalid env keys to be omitted, got:\n%s", unit)
	}
	if !strings.Contains(unit, "Restart=always") {
		t.Errorf("expected Restart=always, got:\n%s", unit)
	}
	if strings.Contains(unit, "Restart=on-failure") {
		t.Errorf("did not expect Restart=on-failure, got:\n%s", unit)
	}
	if !strings.Contains(unit, "StartLimitBurst=5") {
		t.Errorf("expected StartLimitBurst=5, got:\n%s", unit)
	}
	if !strings.Contains(unit, "StartLimitIntervalSec=60") {
		t.Errorf("expected StartLimitIntervalSec=60, got:\n%s", unit)
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

	plist, err := renderLaunchdPlist(ServiceInfo{
		Service:    "channel",
		BinaryName: "beep",
		ExecPath:   `/usr/local/bin/beep & "helper"`,
		Args:       []string{"channel", "up", "<workspace>"},
		Workspace:  `/tmp/a&b`,
		Env: map[string]string{
			"BEEP_TOKEN":   "a&b<c>",
			"BEEP_NEWLINE": "a\nb",
			"BEEP_CTRL":    "ok\x00bad",
		},
	}, label)
	if err != nil {
		t.Fatalf("renderLaunchdPlist: %v", err)
	}
	if strings.Contains(plist, "StandardOutPath") || strings.Contains(plist, "/tmp/should-not-appear.log") {
		t.Errorf("expected no supervisor log redirect, got:\n%s", plist)
	}
	if !strings.Contains(plist, `<string>/usr/local/bin/beep &amp; &#34;helper&#34;</string>`) {
		t.Errorf("expected XML-escaped ExecPath, got:\n%s", plist)
	}
	if !strings.Contains(plist, `<string>&lt;workspace&gt;</string>`) {
		t.Errorf("expected XML-escaped args, got:\n%s", plist)
	}
	if !strings.Contains(plist, `<string>a&amp;b&lt;c&gt;</string>`) {
		t.Errorf("expected XML-escaped env value, got:\n%s", plist)
	}
	if strings.Contains(plist, "BEEP_NEWLINE") || strings.Contains(plist, "a\nb") {
		t.Errorf("expected newline env values to be omitted, got:\n%s", plist)
	}
	if strings.Contains(plist, "BEEP_CTRL") {
		t.Errorf("expected control-char env values to be omitted, got:\n%s", plist)
	}
	if !strings.Contains(plist, "<key>KeepAlive</key>\n\t<true/>") {
		t.Errorf("expected KeepAlive true, got:\n%s", plist)
	}
	if strings.Contains(plist, "<key>SuccessfulExit</key>") {
		t.Errorf("did not expect KeepAlive SuccessfulExit, got:\n%s", plist)
	}
	if !strings.Contains(plist, "<key>ThrottleInterval</key>\n\t<integer>12</integer>") {
		t.Errorf("expected ThrottleInterval 12, got:\n%s", plist)
	}
}

func TestWritePrivateFileMode(t *testing.T) {
	path := filepath.Join(t.TempDir(), "beep-runner.service")
	if err := os.WriteFile(path, []byte("old\n"), 0o644); err != nil {
		t.Fatalf("seed world-readable file: %v", err)
	}
	if err := os.Chmod(path, 0o644); err != nil {
		t.Fatalf("chmod 0644: %v", err)
	}
	if err := writePrivateFile(path, []byte("[Unit]\n")); err != nil {
		t.Fatalf("writePrivateFile: %v", err)
	}
	st, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if got := st.Mode().Perm(); got != 0o600 {
		t.Fatalf("expected mode 0600 after overwrite, got %o", got)
	}
}

func TestServiceTitle(t *testing.T) {
	if got := ServiceTitle("runner"); got != "Runner" {
		t.Fatalf("ServiceTitle(runner)=%q", got)
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
