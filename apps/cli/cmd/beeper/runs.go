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

				var runs []*client.BeeperRun
				err = ui.WithSpinner("Fetching beeper runs...", func() error {
					var listErr error
					runs, listErr = c.ListBeeperRuns(ctx, id)
					return listErr
				})
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

				tbl := ui.NewTable("ID", "SCHEDULED FOR", "STATUS", "SIGNAL STATUS", "CREATED AT")
				tbl.SetIndent("  ")

				for _, r := range runs {
					sigStatus := "-"
					if r.SignalStatus != "" {
						sigStatus = r.SignalStatus
					}
					tbl.AddRow(
						r.ID,
						r.ScheduledFor,
						FormatRunStatus(r.Status),
						sigStatus,
						r.CreatedAt,
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
