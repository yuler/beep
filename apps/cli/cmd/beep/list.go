package beep

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"beep/internal/client"
	"beep/internal/cmdutil"
	"beep/internal/config"
	"beep/internal/ui"

	"github.com/spf13/cobra"
)

// NewCmdList creates the 'beep list' / top-level 'list' subcommand.
func NewCmdList() *cobra.Command {
	var flagStatus string
	var flagKind string

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List beeps (default: active)",
		Long: `List beeps for the current account.

Defaults to active beeps to match the web UI.
Use --status all to include every status, or pass a specific status
(active, paused, completed, cancelled, firing).
Optional --kind once|recurring mirrors the API/web kind filter.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmdutil.RunWithClient(cmd, func(ctx context.Context, cfg *config.Config, c *client.Client) error {
				params := client.ListBeepsParams{
					Status: flagStatus,
					Kind:   flagKind,
				}

				var beeps []*client.Beep
				err := ui.WithSpinner("Fetching beeps...", func() error {
					var listErr error
					beeps, listErr = c.ListBeeps(ctx, params)
					return listErr
				})
				if err != nil {
					return err
				}

				if cmdutil.IsJSON(cmd) {
					data, err := json.MarshalIndent(beeps, "", "  ")
					if err != nil {
						return err
					}
					fmt.Println(string(data))
					return nil
				}

				accountDisplay := cfg.AccountSlug
				if accountDisplay == "" {
					accountDisplay = "personal"
				}
				filterNote := formatListFilterNote(params)
				fmt.Printf("%s %s\n\n", ui.Bold(ui.Cyan("Beeps")), ui.Dim(fmt.Sprintf("(account: %s%s)", accountDisplay, filterNote)))

				if len(beeps) == 0 {
					hint := fmt.Sprintf("  No beeps found. Create one with '%s create'.", config.BinaryName())
					if !strings.EqualFold(strings.TrimSpace(params.Status), "all") && params.Kind == "" {
						hint = fmt.Sprintf("  No beeps found. Try '%s list --status all', or create one with '%s create'.", config.BinaryName(), config.BinaryName())
					}
					fmt.Println(ui.Dim(hint))
					return nil
				}

				tbl := ui.NewTable("ID", "TITLE", "STATUS", "KIND", "SCHEDULE", "CHANNELS")
				tbl.SetIndent("  ")

				for _, b := range beeps {
					statusStr := FormatBeepStatus(b.Status)
					scheduleStr := FormatBeepSchedule(b)
					title := ui.Truncate(b.Title, 28)

					chans := "-"
					if len(b.NotificationChannels) > 0 {
						chans = strings.Join(b.NotificationChannels, ", ")
					}

					tbl.AddRow(
						b.ID,
						title,
						statusStr,
						b.Kind,
						scheduleStr,
						chans,
					)
				}

				if err := tbl.Print(); err != nil {
					return err
				}
				fmt.Println()
				return nil
			})
		},
	}

	cmd.Flags().StringVar(&flagStatus, "status", "active", "Filter by status: active, paused, completed, cancelled, firing, or all")
	cmd.Flags().StringVar(&flagKind, "kind", "", "Filter by kind: once or recurring")

	return cmd
}

func formatListFilterNote(params client.ListBeepsParams) string {
	parts := make([]string, 0, 2)
	status := strings.TrimSpace(params.Status)
	if status == "" {
		status = "all"
	}
	parts = append(parts, "status: "+status)
	if kind := strings.TrimSpace(params.Kind); kind != "" {
		parts = append(parts, "kind: "+kind)
	}
	return ", " + strings.Join(parts, ", ")
}
