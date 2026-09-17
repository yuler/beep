package service

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"beep/internal/cliservice"
	"beep/internal/cmdutil"
	"beep/internal/config"
	"beep/internal/daemon"

	"github.com/charmbracelet/huh"
)

func TestServiceCommandStructure(t *testing.T) {
	cmd := NewCmdService()
	if cmd.Name() != "service" {
		t.Fatalf("expected command name 'service', got %q", cmd.Name())
	}

	expectedSubcommands := map[string][]string{
		"start":   {"up"},
		"stop":    {"down"},
		"restart": nil,
		"status":  nil,
	}

	for name, expectedAliases := range expectedSubcommands {
		sub, _, err := cmd.Find([]string{name})
		if err != nil {
			t.Fatalf("failed to find subcommand %q: %v", name, err)
		}
		if sub.Name() != name {
			t.Errorf("expected subcommand name %q, got %q", name, sub.Name())
		}
		for _, alias := range expectedAliases {
			foundAlias := false
			for _, a := range sub.Aliases {
				if a == alias {
					foundAlias = true
					break
				}
			}
			if !foundAlias {
				t.Errorf("expected subcommand %q to have alias %q, got aliases %v", name, alias, sub.Aliases)
			}
		}
	}
}

func TestServiceStartUnknownTarget(t *testing.T) {
	cmd := NewCmdStart()
	err := cmd.RunE(cmd, []string{"unknown"})
	if err == nil {
		t.Fatal("expected error for unknown target, got nil")
	}
	if !strings.Contains(err.Error(), "unknown service \"unknown\"") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestServiceStopUnknownTarget(t *testing.T) {
	cmd := NewCmdStop()
	err := cmd.RunE(cmd, []string{"unknown"})
	if err == nil {
		t.Fatal("expected error for unknown target, got nil")
	}
	if !strings.Contains(err.Error(), "unknown service \"unknown\"") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestServiceRestartUnknownTarget(t *testing.T) {
	cmd := NewCmdRestart()
	err := cmd.RunE(cmd, []string{"unknown"})
	if err == nil {
		t.Fatal("expected error for unknown target, got nil")
	}
	if !strings.Contains(err.Error(), "unknown service \"unknown\"") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestServiceStatusUnknownTarget(t *testing.T) {
	cmd := NewCmdStatus()
	err := cmd.RunE(cmd, []string{"unknown"})
	if err == nil {
		t.Fatal("expected error for unknown target, got nil")
	}
	if !strings.Contains(err.Error(), "unknown service \"unknown\"") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestServiceRestartWithMock(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		Workspace:    tmpDir,
		ChannelToken: "beep_ct_test",
		RunnerToken:  "beep_rt_test",
	}

	var calledServices []string
	origFn := cliservice.StartServiceDaemonFn
	defer func() { cliservice.StartServiceDaemonFn = origFn }()

	cliservice.StartServiceDaemonFn = func(service string, childSubcommand []string, rawArgs []string, c *config.Config) error {
		calledServices = append(calledServices, service)
		return nil
	}

	// Restart runner only
	err := restartService(daemon.ServiceRunner, cfg, 1*time.Second, true)
	if err != nil {
		t.Fatalf("restartService runner failed: %v", err)
	}
	if len(calledServices) != 1 || calledServices[0] != daemon.ServiceRunner {
		t.Errorf("expected called services [%s], got %v", daemon.ServiceRunner, calledServices)
	}

	// Restart all
	calledServices = nil
	err = restartAll(cfg, 1*time.Second, true)
	if err != nil {
		t.Fatalf("restartAll failed: %v", err)
	}
	if len(calledServices) != 2 {
		t.Fatalf("expected 2 services restarted, got %d (%v)", len(calledServices), calledServices)
	}
}

func TestServiceHelpOutput(t *testing.T) {
	cmd := NewCmdService()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	_ = cmd.Help()
	out := buf.String()

	if !strings.Contains(out, "start") || !strings.Contains(out, "stop") || !strings.Contains(out, "restart") || !strings.Contains(out, "status") {
		t.Errorf("expected help output to contain start, stop, restart, status; got:\n%s", out)
	}
}

func TestServiceRestartInteractive(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		Workspace:    tmpDir,
		ChannelToken: "beep_ct_test",
		RunnerToken:  "beep_rt_test",
	}

	var calledServices []string
	origStartFn := cliservice.StartServiceDaemonFn
	defer func() { cliservice.StartServiceDaemonFn = origStartFn }()

	cliservice.StartServiceDaemonFn = func(service string, childSubcommand []string, rawArgs []string, c *config.Config) error {
		calledServices = append(calledServices, service)
		return nil
	}

	origPromptFn := promptRestartServicesFn
	defer func() { promptRestartServicesFn = origPromptFn }()

	// Test case 1: Select runner only
	promptRestartServicesFn = func(options []huh.Option[string], defaultSelected []string) ([]string, error) {
		return []string{daemon.ServiceRunner}, nil
	}
	calledServices = nil
	err := restartInteractive(cfg, 1*time.Second, true)
	if err != nil {
		t.Fatalf("restartInteractive failed: %v", err)
	}
	if len(calledServices) != 1 || calledServices[0] != daemon.ServiceRunner {
		t.Errorf("expected [runner], got %v", calledServices)
	}

	// Test case 2: Select channel only
	promptRestartServicesFn = func(options []huh.Option[string], defaultSelected []string) ([]string, error) {
		return []string{daemon.ServiceChannel}, nil
	}
	calledServices = nil
	err = restartInteractive(cfg, 1*time.Second, true)
	if err != nil {
		t.Fatalf("restartInteractive failed: %v", err)
	}
	if len(calledServices) != 1 || calledServices[0] != daemon.ServiceChannel {
		t.Errorf("expected [channel], got %v", calledServices)
	}

	// Test case 3: Select both
	promptRestartServicesFn = func(options []huh.Option[string], defaultSelected []string) ([]string, error) {
		return []string{daemon.ServiceRunner, daemon.ServiceChannel}, nil
	}
	calledServices = nil
	err = restartInteractive(cfg, 1*time.Second, true)
	if err != nil {
		t.Fatalf("restartInteractive failed: %v", err)
	}
	if len(calledServices) != 2 {
		t.Errorf("expected 2 services restarted, got %d (%v)", len(calledServices), calledServices)
	}

	// Test case 4: Deselect all (empty selection)
	promptRestartServicesFn = func(options []huh.Option[string], defaultSelected []string) ([]string, error) {
		return []string{}, nil
	}
	calledServices = nil
	err = restartInteractive(cfg, 1*time.Second, true)
	if err != nil {
		t.Fatalf("restartInteractive failed: %v", err)
	}
	if len(calledServices) != 0 {
		t.Errorf("expected 0 services restarted, got %v", calledServices)
	}
}

func TestServiceRestartNonInteractiveFallback(t *testing.T) {
	tmpDir := t.TempDir()
	cmdutil.SetOverrideWorkspace(tmpDir)
	defer cmdutil.SetOverrideWorkspace("")

	fc := &config.FileConfig{
		ServerURL:    "http://localhost:3000",
		RunnerToken:  "beep_rt_test",
		ChannelToken: "beep_ct_test",
	}
	if err := config.SaveFile(filepath.Join(tmpDir, "config.json"), fc); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	var calledServices []string
	origStartFn := cliservice.StartServiceDaemonFn
	defer func() { cliservice.StartServiceDaemonFn = origStartFn }()

	cliservice.StartServiceDaemonFn = func(service string, childSubcommand []string, rawArgs []string, c *config.Config) error {
		calledServices = append(calledServices, service)
		return nil
	}

	cmd := NewCmdRestart()
	err := cmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("restart command failed in non-interactive mode: %v", err)
	}
	if len(calledServices) != 2 {
		t.Errorf("expected all 2 configured services restarted in non-interactive mode, got %d (%v)", len(calledServices), calledServices)
	}
}
