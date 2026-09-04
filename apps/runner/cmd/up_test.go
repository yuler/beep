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

func TestUpCommandRegistration(t *testing.T) {
	cmd, _, err := RootCmd.Find([]string{"up"})
	if err != nil {
		t.Fatalf("failed to find 'up' command: %v", err)
	}
	if cmd.Name() != "up" {
		t.Errorf("expected command name 'up', got %s", cmd.Name())
	}

	// 'run' alias should also resolve to 'up'
	aliasCmd, _, err := RootCmd.Find([]string{"run"})
	if err != nil {
		t.Fatalf("failed to find 'run' command: %v", err)
	}
	if aliasCmd.Name() != "up" {
		t.Errorf("expected alias command name 'up', got %s", aliasCmd.Name())
	}
}

func TestStatusAndStopCommandRegistration(t *testing.T) {
	statusCmd, _, err := RootCmd.Find([]string{"status"})
	if err != nil {
		t.Fatalf("failed to find 'status' command: %v", err)
	}
	if statusCmd.Name() != "status" {
		t.Errorf("expected command name 'status', got %s", statusCmd.Name())
	}

	stopCmd, _, err := RootCmd.Find([]string{"stop"})
	if err != nil {
		t.Fatalf("failed to find 'stop' command: %v", err)
	}
	if stopCmd.Name() != "stop" {
		t.Errorf("expected command name 'stop', got %s", stopCmd.Name())
	}
}
