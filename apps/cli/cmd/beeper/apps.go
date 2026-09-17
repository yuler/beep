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

// NewCmdApps creates the 'beeper apps' subcommand.
func NewCmdApps() *cobra.Command {
	return &cobra.Command{
		Use:     "apps",
		Aliases: []string{"templates", "catalog"},
		Short:   "List available beeper probe app templates",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmdutil.RunWithClient(cmd, func(ctx context.Context, cfg *config.Config, c *client.Client) error {
				var apps []*client.BeeperApp
				err := ui.WithSpinner("Fetching app templates...", func() error {
					var listErr error
					apps, listErr = c.ListBeeperApps(ctx)
					return listErr
				})
				if err != nil {
					return err
				}

				if cmdutil.IsJSON(cmd) {
					data, err := json.MarshalIndent(apps, "", "  ")
					if err != nil {
						return err
					}
					fmt.Println(string(data))
					return nil
				}

				fmt.Printf("%s\n\n", ui.Bold(ui.Cyan("Available Beeper App Templates")))

				if len(apps) == 0 {
					fmt.Println(ui.Dim("  No apps available on server."))
					return nil
				}

				tbl := ui.NewTable("SLUG", "NAME", "DEFAULT CRON", "DESCRIPTION")
				tbl.SetIndent("  ")

				for _, a := range apps {
					name := ui.Truncate(a.Name, 28)
					desc := ui.Truncate(a.Description, 50)
					tbl.AddRow(
						ui.Cyan(a.Slug),
						name,
						a.DefaultCron,
						ui.Dim(desc),
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
