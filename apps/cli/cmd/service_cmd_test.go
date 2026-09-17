package cmd

import (
	"strings"
	"testing"

	"beep/internal/cliservice"
	"beep/internal/config"
	"beep/internal/daemon"
)

func TestStripDaemonFlags(t *testing.T) {
	tests := []struct {
		input    []string
		expected []string
	}{
		{
			input:    []string{"up", "-d", "--server", "http://localhost:3000"},
			expected: []string{"up", "--server", "http://localhost:3000"},
		},
		{
			input:    []string{"-d", "--daemon", "--workspace", "/tmp/ws"},
			expected: []string{"--workspace", "/tmp/ws"},
		},
		{
			input:    []string{"up", "--daemon=true", "-c", "10"},
			expected: []string{"up", "-c", "10"},
		},
	}

	for _, tc := range tests {
		result := cliservice.StripDaemonFlags(tc.input)
		if len(result) != len(tc.expected) {
			t.Fatalf("expected len %d, got %d (result: %v)", len(tc.expected), len(result), result)
		}
		for i := range result {
			if result[i] != tc.expected[i] {
				t.Errorf("at index %d: expected %s, got %s", i, tc.expected[i], result[i])
			}
		}
	}
}

func TestBuildChildDaemonArgs(t *testing.T) {
	tests := []struct {
		input    []string
		expected []string
	}{
		{
			input:    []string{"up", "-d", "--workspace", "/var/run", "-c", "10"},
			expected: []string{"runner", "up", "--workspace", "/var/run", "-c", "10"},
		},
		{
			input:    []string{"runner", "up", "--workspace", "runner", "-d"},
			expected: []string{"runner", "up", "--workspace", "runner"},
		},
		{
			input:    []string{"up", "--daemon=true"},
			expected: []string{"runner", "up"},
		},
		{
			input:    []string{"up", "-d", "-t", "secret_token_val", "--workspace", "/tmp/ws"},
			expected: []string{"runner", "up", "--workspace", "/tmp/ws"},
		},
		{
			input:    []string{"channel", "up", "-d", "--token=secret_val", "--server", "http://localhost:3000"},
			expected: []string{"runner", "up", "--server", "http://localhost:3000"},
		},
		{
			// Flag values that look like subcommand verbs must survive.
			input:    []string{"up", "--server", "up"},
			expected: []string{"runner", "up", "--server", "up"},
		},
	}

	for _, tc := range tests {
		result := cliservice.BuildServiceChildArgs("runner", tc.input)
		if len(result) != len(tc.expected) {
			t.Fatalf("expected len %d, got %d (result: %v)", len(tc.expected), len(result), result)
		}
		for i := range result {
			if result[i] != tc.expected[i] {
				t.Errorf("at index %d: expected %s, got %s", i, tc.expected[i], result[i])
			}
		}
	}
}

func TestUpCommandRegistration(t *testing.T) {
	cmd, _, err := RootCmd.Find([]string{"runner", "start"})
	if err != nil {
		t.Fatalf("failed to find 'runner start' command: %v", err)
	}
	if cmd.Name() != "start" {
		t.Errorf("expected command name 'start', got %s", cmd.Name())
	}

	// 'up' alias should also resolve to 'start'
	aliasCmd, _, err := RootCmd.Find([]string{"runner", "up"})
	if err != nil {
		t.Fatalf("failed to find 'runner up' command: %v", err)
	}
	if aliasCmd.Name() != "start" {
		t.Errorf("expected alias command name 'start', got %s", aliasCmd.Name())
	}
}

func TestStatusAndStopCommandRegistration(t *testing.T) {
	statusCmd, _, err := RootCmd.Find([]string{"runner", "status"})
	if err != nil {
		t.Fatalf("failed to find 'runner status' command: %v", err)
	}
	if statusCmd.Name() != "status" {
		t.Errorf("expected command name 'status', got %s", statusCmd.Name())
	}

	stopCmd, _, err := RootCmd.Find([]string{"runner", "stop"})
	if err != nil {
		t.Fatalf("failed to find 'runner stop' command: %v", err)
	}
	if stopCmd.Name() != "stop" {
		t.Errorf("expected command name 'stop', got %s", stopCmd.Name())
	}

	downCmd, _, err := RootCmd.Find([]string{"runner", "down"})
	if err != nil {
		t.Fatalf("failed to find 'runner down' command: %v", err)
	}
	if downCmd.Name() != "stop" {
		t.Errorf("expected alias command name 'stop', got %s", downCmd.Name())
	}
}

func TestServiceCommandRegistration(t *testing.T) {
	// Root level commands must NOT have up, stop, status directly
	for _, name := range []string{"up", "stop", "status"} {
		cmd, _, _ := RootCmd.Find([]string{name})
		if cmd != nil && cmd.Name() == name && cmd.Parent() == RootCmd {
			t.Errorf("expected %q not to be a direct root command", name)
		}
	}

	// service should be registered at root
	serviceCmd, _, err := RootCmd.Find([]string{"service"})
	if err != nil || serviceCmd == nil {
		t.Fatalf("expected to find 'service' command at root: %v", err)
	}
	if serviceCmd.Name() != "service" {
		t.Errorf("expected command name 'service', got %q", serviceCmd.Name())
	}

	// Subcommands under service
	for _, sub := range []string{"start", "stop", "restart", "status"} {
		subCmd, _, err := RootCmd.Find([]string{"service", sub})
		if err != nil || subCmd == nil {
			t.Fatalf("expected to find 'service %s' command: %v", sub, err)
		}
		if subCmd.Name() != sub {
			t.Errorf("expected command name %q, got %q", sub, subCmd.Name())
		}
	}
}

func TestBuildServiceChildArgs(t *testing.T) {
	runnerArgs := cliservice.BuildServiceChildArgs("runner", []string{"up", "-d", "--workspace", "/var/run"})
	if strings.Join(runnerArgs, " ") != "runner up --workspace /var/run" {
		t.Errorf("unexpected runner args: %v", runnerArgs)
	}

	channelArgs := cliservice.BuildServiceChildArgs("channel", []string{"up", "-d", "--workspace", "/var/run"})
	if strings.Join(channelArgs, " ") != "channel up --workspace /var/run" {
		t.Errorf("unexpected channel args: %v", channelArgs)
	}

	precedingFlagsArgs := cliservice.BuildServiceChildArgs("runner", []string{"--workspace", "/var/run", "up", "-d"})
	if strings.Join(precedingFlagsArgs, " ") != "runner up --workspace /var/run" {
		t.Errorf("unexpected runner args with preceding flags: %v", precedingFlagsArgs)
	}
}

func TestBuildServiceChildArgsIgnoresConnect(t *testing.T) {
	args := cliservice.BuildServiceChildArgs("runner", []string{"runner", "connect", "--workspace", "/var/run"})
	if strings.Join(args, " ") != "runner up --workspace /var/run" {
		t.Errorf("unexpected runner args: %v", args)
	}

	channelArgs := cliservice.BuildServiceChildArgs("channel", []string{"channel", "connect", "--workspace", "/var/run"})
	if strings.Join(channelArgs, " ") != "channel up --workspace /var/run" {
		t.Errorf("unexpected channel args: %v", channelArgs)
	}
}

func TestAutoStartServiceDaemon(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		Workspace:    tmpDir,
		ChannelToken: "beep_ct_test",
	}

	origStart := cliservice.StartServiceDaemonFn
	defer func() { cliservice.StartServiceDaemonFn = origStart }()

	var calledService string
	var calledRawArgs []string
	cliservice.StartServiceDaemonFn = func(service string, childSubcommand []string, rawArgs []string, c *config.Config) error {
		calledService = service
		calledRawArgs = rawArgs
		return nil
	}

	if err := cliservice.AutoStartServiceDaemon(daemon.ServiceChannel, cfg, []string{"--workspace", tmpDir}); err != nil {
		t.Fatalf("AutoStartServiceDaemon failed: %v", err)
	}

	if calledService != daemon.ServiceChannel {
		t.Errorf("expected service %q, got %q", daemon.ServiceChannel, calledService)
	}
	if len(calledRawArgs) < 2 || calledRawArgs[0] != "--workspace" || calledRawArgs[1] != tmpDir {
		t.Errorf("expected rawArgs to contain --workspace %s, got %v", tmpDir, calledRawArgs)
	}
}
