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
				beepers, err := c.ListBeepers(ctx)
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
					fmt.Println(ui.Dim("  No beepers found. Create one with 'beep beeper create'."))
					return nil
				}

				// Print table header
				fmt.Printf("  %-10s  %-24s  %-12s  %-10s  %-10s  %-14s  %s\n",
					ui.Dim("ID"),
					ui.Dim("TITLE"),
					ui.Dim("APP"),
					ui.Dim("STATUS"),
					ui.Dim("ALERT"),
					ui.Dim("SCHEDULE"),
					ui.Dim("LAST PING/RUN"),
				)

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
						lastRun = b.LastPingAt
					} else if b.LastRunAt != "" {
						lastRun = b.LastRunAt
					}

					title := ui.Truncate(b.Title, 24)

					fmt.Printf("  %-10s  %-24s  %-12s  %-10s  %-10s  %-14s  %s\n",
						b.ID,
						title,
						appSlug,
						statusStr,
						alertStr,
						scheduleStr,
						lastRun,
					)
				}
				fmt.Println()
				return nil
			})
		},
	}
}
