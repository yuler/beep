package cmd

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"beep/internal/ui"

	"github.com/spf13/cobra"
)

// SetupHelp configures the gh-style help function and template for a Cobra command and all its children.
func SetupHelp(cmd *cobra.Command) {
	cmd.SetHelpFunc(GhHelpFunc)
}

// GhHelpFunc formats command help information in GitHub CLI (gh) style.
func GhHelpFunc(cmd *cobra.Command, args []string) {
	out := cmd.OutOrStdout()
	if out == nil {
		out = os.Stdout
	}
	_ = RenderGhHelp(out, cmd)
}

// RenderGhHelp renders the gh-style help to the provided writer.
func RenderGhHelp(w io.Writer, cmd *cobra.Command) error {
	// 1. Long or Short Description
	desc := strings.TrimSpace(cmd.Long)
	if desc == "" {
		desc = strings.TrimSpace(cmd.Short)
	}
	if desc != "" {
		if _, err := fmt.Fprintf(w, "%s\n\n", desc); err != nil {
			return err
		}
	}

	// 2. USAGE
	fmt.Fprintln(w, ui.Bold("USAGE"))
	if cmd.HasAvailableSubCommands() {
		if cmd.Parent() == nil {
			fmt.Fprintln(w, "  beep <command> [flags]")
		} else {
			fmt.Fprintf(w, "  %s <command> [flags]\n", cmd.CommandPath())
		}
	} else {
		fmt.Fprintf(w, "  %s\n", cmd.UseLine())
	}
	fmt.Fprintln(w)

	// 3. ALIASES (for leaf commands)
	if len(cmd.Aliases) > 0 {
		fmt.Fprintln(w, ui.Bold("ALIASES"))
		fmt.Fprintf(w, "  %s\n\n", strings.Join(cmd.Aliases, ", "))
	}

	// 4. SUBCOMMANDS grouped in gh style
	if cmd.HasAvailableSubCommands() {
		renderSubcommands(w, cmd)
	}

	// 5. FLAGS (Local non-inherited flags)
	localFlags := cmd.NonInheritedFlags()
	if localFlags.HasAvailableFlags() {
		fmt.Fprintln(w, ui.Bold("FLAGS"))
		fmt.Fprint(w, localFlags.FlagUsages())
		fmt.Fprintln(w)
	}

	// 6. INHERITED FLAGS
	if cmd.Parent() != nil {
		inheritedFlags := cmd.InheritedFlags()
		if inheritedFlags.HasAvailableFlags() {
			fmt.Fprintln(w, ui.Bold("INHERITED FLAGS"))
			fmt.Fprint(w, inheritedFlags.FlagUsages())
			fmt.Fprintln(w)
		}
	}

	// 7. EXAMPLES
	if cmd.Example != "" {
		fmt.Fprintln(w, ui.Bold("EXAMPLES"))
		lines := strings.Split(strings.TrimSpace(cmd.Example), "\n")
		for _, line := range lines {
			fmt.Fprintf(w, "  %s\n", line)
		}
		fmt.Fprintln(w)
	}

	// 8. LEARN MORE
	fmt.Fprintln(w, ui.Bold("LEARN MORE"))
	if cmd.HasAvailableSubCommands() {
		if cmd.Parent() == nil {
			fmt.Fprintln(w, "  Use 'beep <command> --help' for more information about a command.")
		} else {
			fmt.Fprintf(w, "  Use '%s <command> --help' for more information about a command.\n", cmd.CommandPath())
		}
	}
	if cmd.Parent() == nil {
		fmt.Fprintln(w, "  Install shell completion with 'beep completion install' (or 'beep completion --help')")
	}
	fmt.Fprintln(w, "  Read the documentation at https://github.com/yuler/beep")

	return nil
}

func renderSubcommands(w io.Writer, cmd *cobra.Command) {
	type cmdGroup struct {
		id       string
		title    string
		commands []*cobra.Command
	}

	// Build group index
	groupMap := make(map[string]*cmdGroup)
	var orderedGroups []*cmdGroup

	for _, g := range cmd.Groups() {
		cg := &cmdGroup{
			id:       g.ID,
			title:    g.Title,
			commands: nil,
		}
		groupMap[g.ID] = cg
		orderedGroups = append(orderedGroups, cg)
	}

	var ungrouped []*cobra.Command
	for _, child := range cmd.Commands() {
		if !child.IsAvailableCommand() || child.Name() == "help" {
			continue
		}
		if child.GroupID != "" && groupMap[child.GroupID] != nil {
			groupMap[child.GroupID].commands = append(groupMap[child.GroupID].commands, child)
		} else {
			ungrouped = append(ungrouped, child)
		}
	}

	// Calculate maximum command name width across all available commands for clean alignment
	maxNameWidth := 0
	for _, child := range cmd.Commands() {
		if !child.IsAvailableCommand() || child.Name() == "help" {
			continue
		}
		nameWithColon := child.Name() + ":"
		if len(nameWithColon) > maxNameWidth {
			maxNameWidth = len(nameWithColon)
		}
	}
	if maxNameWidth < 12 {
		maxNameWidth = 12
	}

	var groupDescriptions = map[string]string{
		"beeps": "Subcommands can be run directly (e.g. 'beep list') without repeating 'beep beep'",
	}

	printGroup := func(id, title string, commands []*cobra.Command) {
		if len(commands) == 0 {
			return
		}
		sort.Slice(commands, func(i, j int) bool {
			if cmd.Name() == "completion" {
				if commands[i].Name() == "install" {
					return true
				}
				if commands[j].Name() == "install" {
					return false
				}
			}
			return commands[i].Name() < commands[j].Name()
		})

		fmt.Fprintln(w, ui.Bold(strings.ToUpper(title)))
		if desc, ok := groupDescriptions[id]; ok && desc != "" {
			fmt.Fprintf(w, "  %s\n", ui.Dim(desc))
		}
		for _, c := range commands {
			nameWithColon := c.Name() + ":"
			pad := maxNameWidth - len(nameWithColon)
			if pad < 0 {
				pad = 0
			}
			fmt.Fprintf(w, "  %s%s  %s\n", ui.Bold(nameWithColon), strings.Repeat(" ", pad), c.Short)
		}
		fmt.Fprintln(w)
	}

	for _, g := range orderedGroups {
		printGroup(g.id, g.title, g.commands)
	}

	if len(ungrouped) > 0 {
		title := "COMMANDS"
		if len(orderedGroups) > 0 {
			title = "ADDITIONAL COMMANDS"
		}
		printGroup("", title, ungrouped)
	}
}
