package beep

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"beep/internal/client"
	"beep/internal/cmdutil"
	"beep/internal/config"
	"beep/internal/ui"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

// NewCmdBeep creates and returns the parent 'beep' command.
func NewCmdBeep() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "beep",
		Short: "Manage beeps and notifications",
		Long:  ui.Bold(ui.Cyan("Beep Management")) + ` - Create, list, trigger, pause, and delete beeps.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(NewCmdList())
	cmd.AddCommand(NewCmdShow())
	cmd.AddCommand(NewCmdCreate())
	cmd.AddCommand(NewCmdDelete())
	cmd.AddCommand(NewCmdPause())
	cmd.AddCommand(NewCmdResume())
	cmd.AddCommand(NewCmdRun())

	return cmd
}

// ResolveBeepID resolves a beep ID from args, or prompts interactively.
func ResolveBeepID(ctx context.Context, c *client.Client, cmd *cobra.Command, args []string, action string) (string, error) {
	if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
		return strings.TrimSpace(args[0]), nil
	}

	if !cmdutil.IsInteractive(cmd) {
		return "", fmt.Errorf("beep ID is required (e.g. %s %s <id>)", config.BinaryName(), action)
	}

	beeps, err := c.ListBeeps(ctx, client.ListBeepsParams{Status: "all"})
	if err != nil {
		return "", fmt.Errorf("failed to list beeps: %w", err)
	}
	if len(beeps) == 0 {
		return "", errors.New("no beeps found in this account")
	}

	options := make([]huh.Option[string], 0, len(beeps))
	for _, b := range beeps {
		title := ui.Truncate(b.Title, 30)
		label := fmt.Sprintf("%-30s (%s - %s)", title, b.ID, b.Status)
		options = append(options, huh.NewOption(label, b.ID))
	}

	return ui.PromptSelectResource(fmt.Sprintf("Select beep to %s", action), options)
}

func FormatBeepStatus(s string) string {
	switch strings.ToLower(s) {
	case "active":
		return ui.Green("active")
	case "paused":
		return ui.Yellow("paused")
	case "completed":
		return ui.Dim("completed")
	case "cancelled":
		return ui.Dim("cancelled")
	default:
		return s
	}
}

func FormatRunStatus(s string) string {
	switch strings.ToLower(s) {
	case "succeeded", "ok":
		return ui.Green(s)
	case "failed", "error":
		return ui.Red(s)
	case "firing", "running":
		return ui.Yellow(s)
	default:
		return s
	}
}

func FormatBeepSchedule(b *client.Beep) string {
	if b.Kind == "recurring" && b.Cron != "" {
		return "cron: " + b.Cron
	}
	if b.NextRunAt != "" {
		return b.NextRunAt
	}
	if b.RunAt != "" {
		return b.RunAt
	}
	return ui.Dim("instant")
}
