package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"beep/internal/ui"

	"github.com/spf13/cobra"
)

const completionHelpText = `Generate shell completion scripts for Beep CLI commands.

When installing Beep CLI, you can install shell completion automatically
with:

  beep completion install

The generated completions automatically support both 'beep' and 'beep-local'.

If you prefer to configure it manually, follow the instructions below:

### bash

First, ensure that you have bash-completion installed.
Then add this to your ~/.bashrc (or ~/.bash_profile on macOS):

  eval "$(beep completion bash)"
  # or if running beep-local during development:
  eval "$(beep-local completion bash)"

### zsh

Add this to your ~/.zshrc:

  eval "$(beep completion zsh)"
  # or if running beep-local during development:
  eval "$(beep-local completion zsh)"

Or generate a _beep completion script in your $fpath:

  beep completion zsh > /usr/local/share/zsh/site-functions/_beep

### fish

Add this to your ~/.config/fish/config.fish:

  beep completion fish | source
  # or if running beep-local during development:
  beep-local completion fish | source

Or save the completion script directly:

  beep completion fish > ~/.config/fish/completions/beep.fish

### PowerShell

Open your profile script and add:

  Invoke-Expression -Command $(beep completion powershell | Out-String)`

func initCompletionCmd() {
	RootCmd.InitDefaultCompletionCmd()
	for _, c := range RootCmd.Commands() {
		if c.Name() == "completion" {
			c.GroupID = "additional"
			c.Short = "Generate shell completion scripts or install completion to your shell"
			c.Long = completionHelpText
			c.AddCommand(newCmdCompletionInstall())
			wrapShellCompletions(c)
			break
		}
	}
}

func wrapShellCompletions(completionCmd *cobra.Command) {
	for _, sub := range completionCmd.Commands() {
		switch sub.Name() {
		case "bash":
			origRunE := sub.RunE
			sub.RunE = func(cmd *cobra.Command, args []string) error {
				if origRunE != nil {
					if err := origRunE(cmd, args); err != nil {
						return err
					}
				}
				w := cmd.OutOrStdout()
				fmt.Fprintln(w, `
# Register completion for beep-local
if [[ $(type -t compopt) = "builtin" ]]; then
    complete -o default -F __start_beep beep-local 2>/dev/null || true
else
    complete -o default -o nospace -F __start_beep beep-local 2>/dev/null || true
fi`)
				return nil
			}
		case "zsh":
			origRunE := sub.RunE
			sub.RunE = func(cmd *cobra.Command, args []string) error {
				if origRunE != nil {
					if err := origRunE(cmd, args); err != nil {
						return err
					}
				}
				w := cmd.OutOrStdout()
				fmt.Fprintln(w, `
# Register completion for beep-local
compdef _beep beep-local 2>/dev/null || true`)
				return nil
			}
		case "fish":
			origRunE := sub.RunE
			sub.RunE = func(cmd *cobra.Command, args []string) error {
				if origRunE != nil {
					if err := origRunE(cmd, args); err != nil {
						return err
					}
				}
				w := cmd.OutOrStdout()
				fmt.Fprintln(w, `
# Register completion for beep-local
complete -c beep-local -e 2>/dev/null || true
complete -c beep-local -n '__beep_clear_perform_completion_once_result' 2>/dev/null || true
complete -c beep-local -n 'not __beep_requires_order_preservation && __beep_prepare_completions' -f -a '$__beep_comp_results' 2>/dev/null || true
complete -k -c beep-local -n '__beep_requires_order_preservation && __beep_prepare_completions' -f -a '$__beep_comp_results' 2>/dev/null || true`)
				return nil
			}
		case "powershell":
			origRunE := sub.RunE
			sub.RunE = func(cmd *cobra.Command, args []string) error {
				if origRunE != nil {
					if err := origRunE(cmd, args); err != nil {
						return err
					}
				}
				w := cmd.OutOrStdout()
				fmt.Fprintln(w, `
# Register completion for beep-local
Register-ArgumentCompleter -Native -CommandName 'beep-local' -ScriptBlock $scriptblock -ErrorAction SilentlyContinue`)
				return nil
			}
		}
	}
}

func newCmdCompletionInstall() *cobra.Command {
	var flagYes bool
	var flagShell string

	cmd := &cobra.Command{
		Use:   "install [shell]",
		Short: "Install shell completion to your shell configuration file",
		Long: ui.Bold(ui.Cyan("Install Shell Completion")) + ` - Automatically configure autocompletion in your shell configuration file.

Supported shells: bash, zsh, fish.
If no shell is specified, the current shell is detected from $SHELL.
The installed completion supports both 'beep' and 'beep-local'.`,
		Example: `  # Automatically detect shell and install
  $ beep completion install

  # Install for a specific shell
  $ beep completion install zsh
  $ beep completion install bash -y`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			shell := flagShell
			if len(args) > 0 {
				shell = args[0]
			}
			return runCompletionInstall(shell, flagYes)
		},
	}

	cmd.Flags().BoolVarP(&flagYes, "yes", "y", false, "Automatically confirm installation without interactive prompt")
	cmd.Flags().StringVar(&flagShell, "shell", "", "Target shell (bash, zsh, fish; defaults to current shell)")

	return cmd
}

type shellTarget struct {
	Shell       string
	FilePath    string
	DisplayPath string
	Snippet     string
}

func resolveShellTarget(shellName string) (*shellTarget, error) {
	if shellName == "" {
		shellEnv := os.Getenv("SHELL")
		if shellEnv != "" {
			shellName = strings.ToLower(filepath.Base(shellEnv))
		}
	}
	shellName = strings.ToLower(strings.TrimSpace(shellName))
	if shellName == "" {
		return nil, fmt.Errorf("unable to detect shell from $SHELL; please specify a shell explicitly (bash, zsh, fish)")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("could not determine user home directory: %w", err)
	}

	switch shellName {
	case "zsh":
		target := filepath.Join(home, ".zshrc")
		return &shellTarget{
			Shell:       "zsh",
			FilePath:    target,
			DisplayPath: "~/.zshrc",
			Snippet: `# Beep CLI shell completion (supports beep and beep-local)
if command -v beep >/dev/null 2>&1; then
  eval "$(beep completion zsh)"
elif command -v beep-local >/dev/null 2>&1; then
  eval "$(beep-local completion zsh)"
fi`,
		}, nil
	case "bash":
		target := filepath.Join(home, ".bashrc")
		display := "~/.bashrc"
		if runtime.GOOS == "darwin" {
			bashProfile := filepath.Join(home, ".bash_profile")
			if _, err := os.Stat(bashProfile); err == nil {
				target = bashProfile
				display = "~/.bash_profile"
			}
		}
		return &shellTarget{
			Shell:       "bash",
			FilePath:    target,
			DisplayPath: display,
			Snippet: `# Beep CLI shell completion (supports beep and beep-local)
if command -v beep >/dev/null 2>&1; then
  eval "$(beep completion bash)"
elif command -v beep-local >/dev/null 2>&1; then
  eval "$(beep-local completion bash)"
fi`,
		}, nil
	case "fish":
		target := filepath.Join(home, ".config", "fish", "config.fish")
		return &shellTarget{
			Shell:       "fish",
			FilePath:    target,
			DisplayPath: "~/.config/fish/config.fish",
			Snippet: `# Beep CLI shell completion (supports beep and beep-local)
if command -v beep >/dev/null 2>&1
  beep completion fish | source
else if command -v beep-local >/dev/null 2>&1
  beep-local completion fish | source
end`,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported shell %q (supported: bash, zsh, fish)", shellName)
	}
}

func runCompletionInstall(shellName string, autoYes bool) error {
	target, err := resolveShellTarget(shellName)
	if err != nil {
		return err
	}

	// Check if already installed
	if content, err := os.ReadFile(target.FilePath); err == nil {
		str := string(content)
		if strings.Contains(str, "beep completion") || strings.Contains(str, "beep-local completion") {
			fmt.Printf("%s Shell completion is already installed in %s\n", ui.Green("✓"), ui.Bold(target.DisplayPath))
			return nil
		}
	}

	// Display the proposed change
	fmt.Println(ui.Bold("Target file:"), ui.Cyan(target.DisplayPath))
	fmt.Println()
	fmt.Println(ui.Bold("The following configuration will be added:"))
	fmt.Println(ui.Dim("----------------------------------------"))
	for _, line := range strings.Split(target.Snippet, "\n") {
		fmt.Printf("  %s\n", ui.Cyan(line))
	}
	fmt.Println(ui.Dim("----------------------------------------"))
	fmt.Println()

	// Prompt confirmation
	if !autoYes {
		if !ui.IsInteractive() || flagNoInteractive {
			return fmt.Errorf("interactive confirmation required; rerun with -y / --yes to confirm automatically")
		}

		confirm, err := ui.PromptConfirm(fmt.Sprintf("Install Beep completion to %s?", target.DisplayPath), true)
		if err != nil {
			return err
		}
		if !confirm {
			fmt.Println(ui.Dim("Installation canceled."))
			return nil
		}
	}

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(target.FilePath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Append configuration
	var prefix string
	if existing, err := os.ReadFile(target.FilePath); err == nil && len(existing) > 0 {
		if !strings.HasSuffix(string(existing), "\n") {
			prefix = "\n\n"
		} else {
			prefix = "\n"
		}
	}

	f, err := os.OpenFile(target.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open %s for writing: %w", target.DisplayPath, err)
	}
	defer f.Close()

	if _, err := f.WriteString(prefix + target.Snippet + "\n"); err != nil {
		return fmt.Errorf("failed to write completion snippet to %s: %w", target.DisplayPath, err)
	}

	fmt.Printf("%s Successfully installed shell completion to %s\n", ui.Green("✓"), ui.Bold(target.DisplayPath))
	fmt.Println(ui.Dim(fmt.Sprintf("  Restart your shell or run: source %s", target.DisplayPath)))

	return nil
}
