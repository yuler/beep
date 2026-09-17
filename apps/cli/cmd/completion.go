package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"beep/internal/config"
	"beep/internal/ui"

	"github.com/spf13/cobra"
)

func getCompletionHelpText(binName string) string {
	return fmt.Sprintf(`Generate shell completion scripts for %s CLI commands.

When installing %s CLI, you can install shell completion automatically
with:

  %s completion install

If you prefer to configure it manually, follow the instructions below:

### bash

First, ensure that you have bash-completion installed.
Then add this to your ~/.bashrc (or ~/.bash_profile on macOS):

  eval "$(%s completion bash)"

### zsh

Add this to your ~/.zshrc:

  eval "$(%s completion zsh)"

Or generate a _%s completion script in your $fpath:

  %s completion zsh > /usr/local/share/zsh/site-functions/_%s

### fish

Add this to your ~/.config/fish/config.fish:

  %s completion fish | source

Or save the completion script directly:

  %s completion fish > ~/.config/fish/completions/%s.fish

### PowerShell

Open your profile script and add:

  Invoke-Expression -Command $(%s completion powershell | Out-String)`,
		binName, binName, binName, binName, binName, binName, binName, binName, binName, binName, binName, binName)
}

func initCompletionCmd() {
	binName := config.BinaryName()
	RootCmd.InitDefaultCompletionCmd()
	for _, c := range RootCmd.Commands() {
		if c.Name() == "completion" {
			c.GroupID = "additional"
			c.Short = "Generate shell completion scripts or install completion to your shell"
			c.Long = getCompletionHelpText(binName)
			c.AddCommand(newCmdCompletionInstall())
			break
		}
	}
}

func newCmdCompletionInstall() *cobra.Command {
	var flagYes bool
	var flagShell string
	binName := config.BinaryName()

	cmd := &cobra.Command{
		Use:   "install [shell]",
		Short: "Install shell completion to your shell configuration file",
		Long: ui.Bold(ui.Cyan("Install Shell Completion")) + fmt.Sprintf(` - Automatically configure autocompletion in your shell configuration file.

Supported shells: bash, zsh, fish.
If no shell is specified, the current shell is detected from $SHELL.
The installed completion automatically supports '%s'.`, binName),
		Example: fmt.Sprintf(`  # Automatically detect shell and install
  $ %s completion install

  # Install for a specific shell
  $ %s completion install zsh
  $ %s completion install bash -y`, binName, binName, binName),
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

	binName := config.BinaryName()

	switch shellName {
	case "zsh":
		target := filepath.Join(home, ".zshrc")
		return &shellTarget{
			Shell:       "zsh",
			FilePath:    target,
			DisplayPath: "~/.zshrc",
			Snippet:     fmt.Sprintf("# Beep CLI shell completion\neval \"$(%s completion zsh)\"", binName),
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
			Snippet:     fmt.Sprintf("# Beep CLI shell completion\neval \"$(%s completion bash)\"", binName),
		}, nil
	case "fish":
		target := filepath.Join(home, ".config", "fish", "config.fish")
		return &shellTarget{
			Shell:       "fish",
			FilePath:    target,
			DisplayPath: "~/.config/fish/config.fish",
			Snippet:     fmt.Sprintf("# Beep CLI shell completion\n%s completion fish | source", binName),
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

	binName := config.BinaryName()

	// Check if already installed
	if content, err := os.ReadFile(target.FilePath); err == nil {
		str := string(content)
		if strings.Contains(str, binName+" completion") {
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

		confirm, err := ui.PromptConfirm(fmt.Sprintf("Install %s completion to %s?", binName, target.DisplayPath), true)
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
