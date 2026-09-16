package beep

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"beep/internal/client"
	"beep/internal/cmdutil"
	"beep/internal/config"
	"beep/internal/ui"

	"github.com/spf13/cobra"
)

// NewCmdShow creates the 'beep show' subcommand.
func NewCmdShow() *cobra.Command {
	return &cobra.Command{
		Use:     "show [id]",
		Aliases: []string{"view", "info"},
		Short:   "Show details and recent run history of a beep",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmdutil.RunWithClient(cmd, func(ctx context.Context, cfg *config.Config, c *client.Client) error {
				id, err := ResolveBeepID(ctx, c, cmd, args, "view")
				if err != nil {
					return err
				}

				b, err := c.GetBeep(ctx, id)
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

				fmt.Println()
				fmt.Printf("  %s %s\n", ui.Bold("Beep:"), ui.Cyan(b.Title))
				fmt.Println(ui.KeyValue("ID", b.ID))
				fmt.Println(ui.KeyValue("Status", FormatBeepStatus(b.Status)))
				fmt.Println(ui.KeyValue("Kind", b.Kind))
				if b.Cron != "" {
					fmt.Println(ui.KeyValue("Cron", b.Cron))
				}
				if b.RunAt != "" {
					fmt.Println(ui.KeyValue("Run At", b.RunAt))
				}
				if b.NextRunAt != "" {
					fmt.Println(ui.KeyValue("Next Run", b.NextRunAt))
				}
				if b.LastRunAt != "" {
					fmt.Println(ui.KeyValue("Last Run", b.LastRunAt))
				}
				if b.Timezone != "" {
					fmt.Println(ui.KeyValue("Timezone", b.Timezone))
				}
				if len(b.NotificationChannels) > 0 {
					fmt.Println(ui.KeyValue("Channels", strings.Join(b.NotificationChannels, ", ")))
				}
				if b.Body != "" {
					fmt.Println()
					fmt.Println(ui.Dim("  Message Body:"))
					for _, line := range strings.Split(b.Body, "\n") {
						fmt.Printf("    %s\n", line)
					}
				}

				if len(b.Runs) > 0 {
					fmt.Println()
					fmt.Println(ui.Section("  Recent Runs:"))
					fmt.Printf("    %-10s  %-24s  %-10s\n",
						ui.Dim("RUN ID"),
						ui.Dim("SCHEDULED FOR"),
						ui.Dim("STATUS"),
					)
					for _, r := range b.Runs {
						fmt.Printf("    %-10s  %-24s  %-10s\n",
							r.ID,
							r.ScheduledFor,
							FormatRunStatus(r.Status),
						)
					}
				}
				fmt.Println()
				return nil
			})
		},
	}
}
