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

// NewCmdRuns creates the 'beeper runs' subcommand.
func NewCmdRuns() *cobra.Command {
	return &cobra.Command{
		Use:     "runs [id]",
		Aliases: []string{"history", "logs"},
		Short:   "View recent execution runs for a monitor beeper",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmdutil.RunWithClient(cmd, func(ctx context.Context, cfg *config.Config, c *client.Client) error {
				id, err := ResolveBeeperID(ctx, c, cmd, args, "view runs for")
				if err != nil {
					return err
				}

				runs, err := c.ListBeeperRuns(ctx, id)
				if err != nil {
					return err
				}

				if cmdutil.IsJSON(cmd) {
					data, err := json.MarshalIndent(runs, "", "  ")
					if err != nil {
						return err
					}
					fmt.Println(string(data))
					return nil
				}

				fmt.Printf("%s %s\n\n", ui.Bold(ui.Cyan("Beeper Runs")), ui.Dim(fmt.Sprintf("(beeper: %s)", id)))

				if len(runs) == 0 {
					fmt.Println(ui.Dim("  No runs found for this beeper."))
					return nil
				}

				fmt.Printf("  %-10s  %-24s  %-10s  %-14s  %s\n",
					ui.Dim("ID"),
					ui.Dim("SCHEDULED FOR"),
					ui.Dim("STATUS"),
					ui.Dim("SIGNAL STATUS"),
					ui.Dim("CREATED AT"),
				)

				for _, r := range runs {
					sigStatus := "-"
					if r.SignalStatus != "" {
						sigStatus = r.SignalStatus
					}
					fmt.Printf("  %-10s  %-24s  %-10s  %-14s  %s\n",
						r.ID,
						r.ScheduledFor,
						FormatRunStatus(r.Status),
						sigStatus,
						r.CreatedAt,
					)
				}
				fmt.Println()
				return nil
			})
		},
	}
}
