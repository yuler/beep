package beep

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

// NewCmdResume creates the 'beep resume' subcommand.
func NewCmdResume() *cobra.Command {
	return &cobra.Command{
		Use:   "resume [id]",
		Short: "Resume a paused beep",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmdutil.RunWithClient(cmd, func(ctx context.Context, cfg *config.Config, c *client.Client) error {
				id, err := ResolveBeepID(ctx, c, cmd, args, "resume")
				if err != nil {
					return err
				}

				var b *client.Beep
				err = ui.WithSpinner("Resuming beep...", func() error {
					var resumeErr error
					b, resumeErr = c.ResumeBeep(ctx, id)
					return resumeErr
				})
				if err != nil {
					return err
				}

				if cmdutil.IsJSON(cmd) {
					data, err := json.MarshalIndent(b, "", "  ")
					if err != nil {
						return err
					}
					fmt.Println(string(data))
					return nil
				}

				fmt.Println(ui.Success("Resumed beep %s (%s)", ui.Bold(b.Title), ui.Dim(b.ID)))
				return nil
			})
		},
	}
}
