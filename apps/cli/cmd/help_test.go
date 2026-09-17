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

	// Check colon format on command names (including aliases)
	expectedCommandsWithColon := []string{
		"create:",
		"delete, rm:",
		"list, ls:",
		"pause:",
		"resume:",
		"run, trigger:",
		"show, view, info:",
		"beeper:",
		"channel:",
		"runner:",
		"service:",
		"auth:",
		"config:",
		"upgrade, update:",
		"version:",
	}
	for _, c := range expectedCommandsWithColon {
		if !strings.Contains(out, c) {
			t.Errorf("expected command %q in help output", c)
		}
	}

	serviceIdx := strings.Index(out, "service:")
	channelIdx := strings.Index(out, "channel:")
	if serviceIdx == -1 || channelIdx == -1 || serviceIdx > channelIdx {
		t.Errorf("expected 'service:' before 'channel:' in LOCAL SERVICE COMMANDS, got serviceIdx=%d, channelIdx=%d", serviceIdx, channelIdx)
	}

	// Check learn more section
	if !strings.Contains(out, "Use 'beep <command> --help' for more information about a command.") {
		t.Errorf("expected LEARN MORE tip in help output")
	}
	if !strings.Contains(out, "Install shell completion with 'beep completion install'") {
		t.Errorf("expected completion install tip in root help output")
	}
}

func TestCompletionHelpSortsInstallFirst(t *testing.T) {
	completionCmd, _, err := RootCmd.Find([]string{"completion"})
	if err != nil {
		t.Fatalf("failed to find completion command: %v", err)
	}

	var buf bytes.Buffer
	if err := RenderGhHelp(&buf, completionCmd); err != nil {
		t.Fatalf("RenderGhHelp failed: %v", err)
	}

	out := buf.String()
	installIdx := strings.Index(out, "install:")
	bashIdx := strings.Index(out, "bash:")
	if installIdx == -1 || bashIdx == -1 || installIdx > bashIdx {
		t.Errorf("expected 'install:' before 'bash:' in completion help, got:\n%s", out)
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
	if strings.Contains(out, "ALIASES") {
		t.Errorf("expected no ALIASES for beeper command, got:\n%s", out)
	}
	if !strings.Contains(out, "COMMANDS") {
		t.Errorf("expected COMMANDS section for beeper")
	}
	if !strings.Contains(out, "apps, templates, catalog:") || !strings.Contains(out, "create:") {
		t.Errorf("expected subcommands with colon for beeper")
	}
	if strings.Contains(out, "Install shell completion with 'beep completion install'") {
		t.Errorf("subcommand help should not contain root completion install tip")
	}
}

func TestGhHelpWithCustomBinName(t *testing.T) {
	t.Setenv("BEEP_BIN_NAME", "beep-local")

	var buf bytes.Buffer
	if err := RenderGhHelp(&buf, RootCmd); err != nil {
		t.Fatalf("RenderGhHelp failed: %v", err)
	}

	out := buf.String()

	if !strings.Contains(out, "beep-local <command> [flags]") {
		t.Errorf("expected USAGE section with 'beep-local <command> [flags]', got: %s", out)
	}
	if !strings.Contains(out, "Use 'beep-local <command> --help' for more information about a command.") {
		t.Errorf("expected LEARN MORE tip with 'beep-local <command> --help', got: %s", out)
	}
	if !strings.Contains(out, "Install shell completion with 'beep-local completion install'") {
		t.Errorf("expected completion install tip with 'beep-local completion install', got: %s", out)
	}

	beeperCmd, _, err := RootCmd.Find([]string{"beeper"})
	if err != nil {
		t.Fatalf("failed to find beeper command: %v", err)
	}

	var subBuf bytes.Buffer
	if err := RenderGhHelp(&subBuf, beeperCmd); err != nil {
		t.Fatalf("RenderGhHelp on beeper failed: %v", err)
	}

	subOut := subBuf.String()
	if !strings.Contains(subOut, "beep-local beeper <command> [flags]") {
		t.Errorf("expected USAGE with 'beep-local beeper <command> [flags]', got: %s", subOut)
	}
}
