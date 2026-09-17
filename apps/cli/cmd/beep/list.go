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

// NewCmdList creates the 'beep list' subcommand.
func NewCmdList() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List beeps",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmdutil.RunWithClient(cmd, func(ctx context.Context, cfg *config.Config, c *client.Client) error {
				var beeps []*client.Beep
				err := ui.WithSpinner("Fetching beeps...", func() error {
					var listErr error
					beeps, listErr = c.ListBeeps(ctx)
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
				fmt.Printf("%s %s\n\n", ui.Bold(ui.Cyan("Beeps")), ui.Dim(fmt.Sprintf("(account: %s)", accountDisplay)))

				if len(beeps) == 0 {
					fmt.Println(ui.Dim("  No beeps found. Create one with 'beep create'."))
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
}
