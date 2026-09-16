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

// NewCmdRun creates the 'beeper run' subcommand.
func NewCmdRun() *cobra.Command {
	return &cobra.Command{
		Use:     "run [id]",
		Aliases: []string{"trigger"},
		Short:   "Immediately trigger a monitor beeper run",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmdutil.RunWithClient(cmd, func(ctx context.Context, cfg *config.Config, c *client.Client) error {
				id, err := ResolveBeeperID(ctx, c, cmd, args, "trigger")
				if err != nil {
					return err
				}

				run, err := c.RunBeeper(ctx, id)
				if err != nil {
					return err
				}

				if cmdutil.IsJSON(cmd) {
					data, err := json.MarshalIndent(run, "", "  ")
					if err != nil {
						return err
					}
					fmt.Println(string(data))
					return nil
				}

				fmt.Println(ui.Success("Triggered beeper %s (run id: %s, status: %s)",
					ui.Bold(id),
					ui.Dim(run.ID),
					FormatRunStatus(run.Status),
				))
				return nil
			})
		},
	}
}
