package beeper

import (
	"context"
	"fmt"

	"beep/internal/client"
	"beep/internal/cmdutil"
	"beep/internal/config"
	"beep/internal/ui"

	"github.com/spf13/cobra"
)

// NewCmdDelete creates the 'beeper delete' subcommand.
func NewCmdDelete() *cobra.Command {
	return &cobra.Command{
		Use:     "delete [id]",
		Aliases: []string{"rm"},
		Short:   "Delete a monitor beeper",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmdutil.RunWithClient(cmd, func(ctx context.Context, cfg *config.Config, c *client.Client) error {
				id, err := ResolveBeeperID(ctx, c, cmd, args, "delete")
				if err != nil {
					return err
				}

				if cmdutil.IsInteractive(cmd) && !cmdutil.IsJSON(cmd) {
					confirm, err := ui.PromptConfirm(fmt.Sprintf("Are you sure you want to delete beeper %s?", id), false)
					if err != nil {
						return err
					}
					if !confirm {
						fmt.Println(ui.Dim("Cancelled."))
						return nil
					}
				}

				if err := c.DeleteBeeper(ctx, id); err != nil {
					return err
				}

				if cmdutil.IsJSON(cmd) {
					fmt.Printf("{\"id\":%q,\"deleted\":true}\n", id)
					return nil
				}

				fmt.Println(ui.Success("Deleted beeper %s", ui.Bold(id)))
				return nil
			})
		},
	}
}
