package cmd

import (
	"bytes"
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

	for _, shell := range []string{"bash", "zsh", "fish"} {
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
