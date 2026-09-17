package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestCompletionCommandRegistered(t *testing.T) {
	cmd, _, err := RootCmd.Find([]string{"completion"})
	if err != nil {
		t.Fatalf("failed to find completion command: %v", err)
	}
	if cmd.Name() != "completion" {
		t.Errorf("expected command name %q, got %q", "completion", cmd.Name())
	}

	for _, shell := range []string{"bash", "zsh", "fish", "install"} {
		sub, _, err := RootCmd.Find([]string{"completion", shell})
		if err != nil {
			t.Fatalf("failed to find completion %s: %v", shell, err)
		}
		if sub.Name() != shell {
			t.Errorf("expected completion subcommand %q, got %q", shell, sub.Name())
		}
	}
}

func TestCompletionScripts(t *testing.T) {
	cases := []struct {
		shell    string
		generate func(*bytes.Buffer) error
		contains []string
	}{
		{
			shell: "bash",
			generate: func(buf *bytes.Buffer) error {
				return RootCmd.GenBashCompletionV2(buf, true)
			},
			contains: []string{
				"__beep_get_completion_results",
				"complete",
			},
		},
		{
			shell: "zsh",
			generate: func(buf *bytes.Buffer) error {
				return RootCmd.GenZshCompletion(buf)
			},
			contains: []string{
				"#compdef beep",
			},
		},
		{
			shell: "fish",
			generate: func(buf *bytes.Buffer) error {
				return RootCmd.GenFishCompletion(buf, true)
			},
			contains: []string{
				"complete -c beep",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.shell, func(t *testing.T) {
			var buf bytes.Buffer
			if err := tc.generate(&buf); err != nil {
				t.Fatalf("completion %s: %v", tc.shell, err)
			}
			script := buf.String()
			if script == "" {
				t.Fatalf("completion %s produced empty script", tc.shell)
			}
			for _, needle := range tc.contains {
				if !strings.Contains(script, needle) {
					t.Errorf("completion %s script missing %q", tc.shell, needle)
				}
			}
		})
	}
}

func TestCompletionSubcommandsIncludeBeepLocal(t *testing.T) {
	shells := []string{"bash", "zsh", "fish", "powershell"}
	for _, shell := range shells {
		t.Run(shell, func(t *testing.T) {
			sub, _, err := RootCmd.Find([]string{"completion", shell})
			if err != nil {
				t.Fatalf("failed to find completion %s: %v", shell, err)
			}
			var buf bytes.Buffer
			sub.SetOut(&buf)
			if err := sub.RunE(sub, nil); err != nil {
				t.Fatalf("completion %s failed: %v", shell, err)
			}
			out := buf.String()
			if !strings.Contains(out, "beep-local") {
				t.Errorf("expected completion %s script to include 'beep-local' registration", shell)
			}
		})
	}
}

func TestCompletionHelpOutput(t *testing.T) {
	completionCmd, _, err := RootCmd.Find([]string{"completion"})
	if err != nil {
		t.Fatalf("failed to find completion command: %v", err)
	}

	var buf bytes.Buffer
	if err := RenderGhHelp(&buf, completionCmd); err != nil {
		t.Fatalf("RenderGhHelp failed: %v", err)
	}

	out := buf.String()
	requiredSnippets := []string{
		"beep completion install",
		"### bash",
		"eval \"$(beep completion bash)\"",
		"### zsh",
		"eval \"$(beep completion zsh)\"",
		"### fish",
		"beep completion fish | source",
		"### PowerShell",
		"install:",
	}
	for _, snippet := range requiredSnippets {
		if !strings.Contains(out, snippet) {
			t.Errorf("expected completion help to contain %q, but got:\n%s", snippet, out)
		}
	}

	installIdx := strings.Index(out, "install:")
	bashIdx := strings.Index(out, "bash:")
	if installIdx == -1 || bashIdx == -1 || installIdx > bashIdx {
		t.Errorf("expected 'install:' to appear before 'bash:' in subcommands list, got installIdx=%d, bashIdx=%d", installIdx, bashIdx)
	}
}

func TestCompletionInstall(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// 1. Install zsh with autoYes = true
	if err := runCompletionInstall("zsh", true); err != nil {
		t.Fatalf("runCompletionInstall(zsh) failed: %v", err)
	}
	zshrc := tmpHome + "/.zshrc"
	content, err := os.ReadFile(zshrc)
	if err != nil {
		t.Fatalf("failed to read created .zshrc: %v", err)
	}
	if !strings.Contains(string(content), "eval \"$(beep completion zsh)\"") {
		t.Errorf("expected .zshrc to contain zsh completion snippet, got:\n%s", string(content))
	}

	// 2. Re-install should detect already installed without error
	if err := runCompletionInstall("zsh", true); err != nil {
		t.Fatalf("runCompletionInstall(zsh) repeat failed: %v", err)
	}

	// 3. Install bash
	if err := runCompletionInstall("bash", true); err != nil {
		t.Fatalf("runCompletionInstall(bash) failed: %v", err)
	}
	bashrc := tmpHome + "/.bashrc"
	contentBash, err := os.ReadFile(bashrc)
	if err != nil {
		t.Fatalf("failed to read created .bashrc: %v", err)
	}
	if !strings.Contains(string(contentBash), "eval \"$(beep completion bash)\"") {
		t.Errorf("expected .bashrc to contain bash completion snippet, got:\n%s", string(contentBash))
	}

	// 4. Install fish
	if err := runCompletionInstall("fish", true); err != nil {
		t.Fatalf("runCompletionInstall(fish) failed: %v", err)
	}
	fishConfig := tmpHome + "/.config/fish/config.fish"
	contentFish, err := os.ReadFile(fishConfig)
	if err != nil {
		t.Fatalf("failed to read created config.fish: %v", err)
	}
	if !strings.Contains(string(contentFish), "beep completion fish | source") {
		t.Errorf("expected config.fish to contain fish completion snippet, got:\n%s", string(contentFish))
	}

	// 5. Non-interactive without -y should fail
	flagNoInteractive = true
	defer func() { flagNoInteractive = false }()

	tmpHome2 := t.TempDir()
	t.Setenv("HOME", tmpHome2)
	err = runCompletionInstall("zsh", false)
	if err == nil {
		t.Errorf("expected error when running non-interactively without auto-yes, got nil")
	} else if !strings.Contains(err.Error(), "confirmation required") {
		t.Errorf("unexpected error message: %v", err)
	}

	// 6. Unsupported shell
	err = runCompletionInstall("elvish", true)
	if err == nil {
		t.Errorf("expected error for unsupported shell, got nil")
	}
}
