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

// NewCmdResume creates the 'beeper resume' subcommand.
func NewCmdResume() *cobra.Command {
	return &cobra.Command{
		Use:   "resume [id]",
		Short: "Resume a paused monitor beeper",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmdutil.RunWithClient(cmd, func(ctx context.Context, cfg *config.Config, c *client.Client) error {
				id, err := ResolveBeeperID(ctx, c, cmd, args, "resume")
				if err != nil {
					return err
				}

				var b *client.Beeper
				err = ui.WithSpinner("Resuming beeper...", func() error {
					var resumeErr error
					b, resumeErr = c.ResumeBeeper(ctx, id)
					return resumeErr
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

				fmt.Println(ui.Success("Resumed beeper %s (%s)", ui.Bold(b.Title), ui.Dim(b.ID)))
				return nil
			})
		},
	}
}
