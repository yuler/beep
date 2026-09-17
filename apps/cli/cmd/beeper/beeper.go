package beeper

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

// NewCmdBeeper creates and returns the parent 'beeper' command.
func NewCmdBeeper() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "beeper",
		Short: "Manage monitor probe beepers",
		Long:    ui.Bold(ui.Cyan("Beeper Management")) + ` - Manage monitoring probes, view catalog apps, and check probe runs.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(NewCmdList())
	cmd.AddCommand(NewCmdShow())
	cmd.AddCommand(NewCmdApps())
	cmd.AddCommand(NewCmdCreate())
	cmd.AddCommand(NewCmdDelete())
	cmd.AddCommand(NewCmdPause())
	cmd.AddCommand(NewCmdResume())
	cmd.AddCommand(NewCmdRun())
	cmd.AddCommand(NewCmdRuns())

	return cmd
}

// ResolveBeeperID resolves a beeper ID from args, or prompts interactively.
func ResolveBeeperID(ctx context.Context, c *client.Client, cmd *cobra.Command, args []string, action string) (string, error) {
	if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
		return strings.TrimSpace(args[0]), nil
	}

	if !cmdutil.IsInteractive(cmd) {
		return "", fmt.Errorf("beeper ID is required (e.g. %s beeper %s <id>)", config.BinaryName(), action)
	}

	beepers, err := c.ListBeepers(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list beepers: %w", err)
	}
	if len(beepers) == 0 {
		return "", errors.New("no beepers found in this account")
	}

	options := make([]huh.Option[string], 0, len(beepers))
	for _, b := range beepers {
		title := ui.Truncate(b.Title, 30)
		label := fmt.Sprintf("%-30s (%s - %s)", title, b.ID, b.Status)
		options = append(options, huh.NewOption(label, b.ID))
	}

	return ui.PromptSelectResource(fmt.Sprintf("Select beeper to %s", action), options)
}

func FormatBeeperStatus(s string) string {
	switch strings.ToLower(s) {
	case "active":
		return ui.Green("active")
	case "paused":
		return ui.Yellow("paused")
	case "disabled":
		return ui.Dim("disabled")
	default:
		return s
	}
}

func FormatBeeperAlert(s string) string {
	switch strings.ToLower(s) {
	case "ok", "clear", "healthy":
		return ui.Green(s)
	case "firing", "alerting", "critical":
		return ui.Red(s)
	case "pending":
		return ui.Yellow(s)
	default:
		return ui.Dim(s)
	}
}

func FormatBeeperSchedule(cron string) string {
	if cron == "" {
		return ui.Dim("webhook only")
	}
	return cron
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
