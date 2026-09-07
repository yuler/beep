package cmd

import (
	"testing"
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
		result := stripDaemonFlags(tc.input)
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
	}

	for _, tc := range tests {
		result := buildChildDaemonArgs(tc.input)
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
	cmd, _, err := RootCmd.Find([]string{"runner", "up"})
	if err != nil {
		t.Fatalf("failed to find 'runner up' command: %v", err)
	}
	if cmd.Name() != "up" {
		t.Errorf("expected command name 'up', got %s", cmd.Name())
	}

	// 'run' alias should also resolve to 'up'
	aliasCmd, _, err := RootCmd.Find([]string{"runner", "run"})
	if err != nil {
		t.Fatalf("failed to find 'runner run' command: %v", err)
	}
	if aliasCmd.Name() != "up" {
		t.Errorf("expected alias command name 'up', got %s", aliasCmd.Name())
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
}
