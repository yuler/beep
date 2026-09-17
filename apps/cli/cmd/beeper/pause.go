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

// NewCmdPause creates the 'beeper pause' subcommand.
func NewCmdPause() *cobra.Command {
	return &cobra.Command{
		Use:   "pause [id]",
		Short: "Pause a monitor beeper",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmdutil.RunWithClient(cmd, func(ctx context.Context, cfg *config.Config, c *client.Client) error {
				id, err := ResolveBeeperID(ctx, c, cmd, args, "pause")
				if err != nil {
					return err
				}

				var b *client.Beeper
				err = ui.WithSpinner("Pausing beeper...", func() error {
					var pauseErr error
					b, pauseErr = c.PauseBeeper(ctx, id)
					return pauseErr
				})
				if err != nil {
					return err
				}

				if cmdutil.IsJSON(cmd) {
					redacted := b.Redacted(false)
					data, err := json.MarshalIndent(redacted, "", "  ")
					if err != nil {
						return err
					}
					fmt.Println(string(data))
					return nil
				}

				fmt.Println(ui.Success("Paused beeper %s (%s)", ui.Bold(b.Title), ui.Dim(b.ID)))
				return nil
			})
		},
	}
}
