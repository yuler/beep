package beeper

import (
	"context"
	"encoding/json"
	"fmt"

	"beep/internal/client"
	"beep/internal/cmdutil"
	"beep/internal/config"
	"beep/internal/ui"

	"github.com/spf13/cobra"
)

// NewCmdList creates the 'beeper list' subcommand.
func NewCmdList() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List monitor beepers",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmdutil.RunWithClient(cmd, func(ctx context.Context, cfg *config.Config, c *client.Client) error {
				var beepers []*client.Beeper
				err := ui.WithSpinner("Fetching beepers...", func() error {
					var listErr error
					beepers, listErr = c.ListBeepers(ctx)
					return listErr
				})
				if err != nil {
					return err
				}

				if cmdutil.IsJSON(cmd) {
					redacted := client.RedactBeepers(beepers, false)
					data, err := json.MarshalIndent(redacted, "", "  ")
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
				fmt.Printf("%s %s\n\n", ui.Bold(ui.Cyan("Beepers")), ui.Dim(fmt.Sprintf("(account: %s)", accountDisplay)))

				if len(beepers) == 0 {
					fmt.Println(ui.Dim(fmt.Sprintf("  No beepers found. Create one with '%s beeper create'.", config.BinaryName())))
					return nil
				}

				tbl := ui.NewTable("ID", "TITLE", "APP", "STATUS", "ALERT", "SCHEDULE", "LAST PING/RUN")
				tbl.SetIndent("  ")

				for _, b := range beepers {
					appSlug := "-"
					if b.BeeperApp != nil {
						appSlug = b.BeeperApp.Slug
					}
					statusStr := FormatBeeperStatus(b.Status)
					alertStr := FormatBeeperAlert(b.AlertState)
					scheduleStr := FormatBeeperSchedule(b.Cron)
					lastRun := "-"
					if b.LastPingAt != "" {
						lastRun = ui.FormatTimestamp(b.LastPingAt, b.Timezone)
					} else if b.LastRunAt != "" {
						lastRun = ui.FormatTimestamp(b.LastRunAt, b.Timezone)
					}

					title := ui.Truncate(b.Title, 28)

					tbl.AddRow(
						b.ID,
						title,
						appSlug,
						statusStr,
						alertStr,
						scheduleStr,
						lastRun,
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
