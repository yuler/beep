package service

import (
	"errors"
	"strings"
	"testing"
	"time"

	"beep/internal/config"
	"beep/internal/daemon"
	"beep/internal/supervisor"
)

type stubSupervisor struct {
	installed    map[string]bool
	uninstalled  []string
	uninstallErr error
}

func (s *stubSupervisor) IsSupported() bool                         { return true }
func (s *stubSupervisor) PlatformName() string                      { return "systemd" }
func (s *stubSupervisor) Install(info supervisor.ServiceInfo) error { return nil }
func (s *stubSupervisor) Uninstall(service, binaryName string) error {
	if s.uninstallErr != nil {
		return s.uninstallErr
	}
	s.uninstalled = append(s.uninstalled, service)
	if s.installed != nil {
		s.installed[service] = false
	}
	return nil
}
func (s *stubSupervisor) Start(service, binaryName string) error { return nil }
func (s *stubSupervisor) Stop(service, binaryName string) error  { return nil }
func (s *stubSupervisor) GetStatus(service, binaryName string) supervisor.Status {
	return supervisor.Status{
		Supported: true,
		Platform:  "systemd",
		Installed: s.installed[service],
		UnitName:  service,
	}
}
func (s *stubSupervisor) EnsureLinger() (bool, error) { return true, nil }

func TestSupervisorEnvOmitsTokens(t *testing.T) {
	t.Setenv("BEEP_RUNNER_TOKEN", "from-shell")
	t.Setenv("PATH", "/bin")
	env := supervisorEnv(&config.Config{
		Workspace:    "/ws",
		ServerURL:    "https://example.test",
		RunnerToken:  "cfg-runner",
		ChannelToken: "cfg-channel",
	})
	if env["PATH"] != "/bin" {
		t.Fatalf("PATH=%q", env["PATH"])
	}
	if env["BEEP_WORKSPACE"] != "/ws" || env["BEEP_SERVER"] != "https://example.test" {
		t.Fatalf("expected workspace/server in env, got %#v", env)
	}
	for _, key := range []string{"BEEP_RUNNER_TOKEN", "BEEP_CHANNEL_TOKEN"} {
		if _, ok := env[key]; ok {
			t.Fatalf("expected %s omitted from supervisor unit env", key)
		}
	}
}

func TestStopSingleServiceReturnsUninstallError(t *testing.T) {
	orig := supervisor.DefaultManager
	supervisor.DefaultManager = &stubSupervisor{
		installed:    map[string]bool{daemon.ServiceRunner: true},
		uninstallErr: errors.New("disable failed"),
	}
	t.Cleanup(func() { supervisor.DefaultManager = orig })

	err := StopSingleService(daemon.ServiceRunner, t.TempDir(), time.Second, true)
	if err == nil || !strings.Contains(err.Error(), "unregister supervisor") {
		t.Fatalf("expected unregister error, got %v", err)
	}
}
