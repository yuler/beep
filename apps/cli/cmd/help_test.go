package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestGhHelpFormatting(t *testing.T) {
	var buf bytes.Buffer
	if err := RenderGhHelp(&buf, RootCmd); err != nil {
		t.Fatalf("RenderGhHelp failed: %v", err)
	}

	out := buf.String()

	// Check that redundant line was removed
	if strings.Contains(out, "Execute 'beep <command> --help' for detailed usage of a specific command.") {
		t.Errorf("output should not contain legacy detailed usage prompt")
	}

	// Check USAGE section
	if !strings.Contains(out, "USAGE") || !strings.Contains(out, "  beep <command> [flags]") {
		t.Errorf("expected USAGE section with 'beep <command> [flags]'")
	}

	// Check group titles
	expectedSections := []string{
		"CORE COMMANDS",
		"BEEP COMMANDS (DEFAULT SCOPE)",
		"LOCAL SERVICE COMMANDS",
		"ADDITIONAL COMMANDS",
		"FLAGS",
		"LEARN MORE",
	}
	for _, sec := range expectedSections {
		if !strings.Contains(out, sec) {
			t.Errorf("expected output to contain section %q", sec)
		}
	}

	if !strings.Contains(out, "Subcommands can be run directly") {
		t.Errorf("expected output to contain note about subcommands being runnable directly")
	}

	// Check colon format on command names
	expectedCommandsWithColon := []string{
		"create:",
		"delete:",
		"list:",
		"pause:",
		"resume:",
		"run:",
		"show:",
		"beeper:",
		"channel:",
		"runner:",
		"status:",
		"stop:",
		"up:",
		"auth:",
		"config:",
		"upgrade:",
		"version:",
	}
	for _, c := range expectedCommandsWithColon {
		if !strings.Contains(out, c) {
			t.Errorf("expected command %q in help output", c)
		}
	}

	// Check learn more section
	if !strings.Contains(out, "Use 'beep <command> --help' for more information about a command.") {
		t.Errorf("expected LEARN MORE tip in help output")
	}
}

func TestSubcommandHelpFormatting(t *testing.T) {
	beeperCmd, _, err := RootCmd.Find([]string{"beeper"})
	if err != nil {
		t.Fatalf("failed to find beeper command: %v", err)
	}

	var buf bytes.Buffer
	if err := RenderGhHelp(&buf, beeperCmd); err != nil {
		t.Fatalf("RenderGhHelp failed: %v", err)
	}

	out := buf.String()

	if !strings.Contains(out, "USAGE") || !strings.Contains(out, "  beep beeper <command> [flags]") {
		t.Errorf("expected USAGE for beeper command")
	}
	if !strings.Contains(out, "ALIASES") || !strings.Contains(out, "  beepers") {
		t.Errorf("expected ALIASES for beeper command")
	}
	if !strings.Contains(out, "COMMANDS") {
		t.Errorf("expected COMMANDS section for beeper")
	}
	if !strings.Contains(out, "apps:") || !strings.Contains(out, "create:") {
		t.Errorf("expected subcommands with colon for beeper")
	}
}
