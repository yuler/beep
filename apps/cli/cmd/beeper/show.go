package beeper

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"beep/internal/client"
	"beep/internal/cmdutil"
	"beep/internal/config"
	"beep/internal/ui"

	"github.com/spf13/cobra"
)

// NewCmdShow creates the 'beeper show' subcommand.
func NewCmdShow() *cobra.Command {
	var flagShowToken bool

	cmd := &cobra.Command{
		Use:     "show [id]",
		Aliases: []string{"view", "info"},
		Short:   "Show details and health status of a monitor beeper",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmdutil.RunWithClient(cmd, func(ctx context.Context, cfg *config.Config, c *client.Client) error {
				id, err := ResolveBeeperID(ctx, c, cmd, args, "view")
				if err != nil {
					return err
				}

				var b *client.Beeper
				err = ui.WithSpinner("Fetching beeper details...", func() error {
					var getErr error
					b, getErr = c.GetBeeper(ctx, id)
					return getErr
				})
				if err != nil {
					return err
				}

				if cmdutil.IsJSON(cmd) {
					redacted := b.Redacted(flagShowToken)
					data, err := json.MarshalIndent(redacted, "", "  ")
					if err != nil {
						return err
					}
					fmt.Println(string(data))
					return nil
				}

				tokenDisplay := client.MaskToken(b.PingToken)
				if flagShowToken {
					tokenDisplay = b.PingToken
				}

				fmt.Println()
				fmt.Printf("  %s %s\n", ui.Bold("Beeper:"), ui.Cyan(b.Title))
				fmt.Println(ui.KeyValue("ID", b.ID))
				if b.BeeperApp != nil {
					fmt.Println(ui.KeyValue("App", fmt.Sprintf("%s (%s)", b.BeeperApp.Name, b.BeeperApp.Slug)))
				}
				fmt.Println(ui.KeyValue("Status", FormatBeeperStatus(b.Status)))
				fmt.Println(ui.KeyValue("Alert State", FormatBeeperAlert(b.AlertState)))
				if b.ConsecutiveFailures > 0 {
					fmt.Println(ui.KeyValue("Failures", ui.Red(fmt.Sprintf("%d consecutive", b.ConsecutiveFailures))))
				}
				if b.Cron != "" {
					fmt.Println(ui.KeyValue("Cron", b.Cron))
				}
				if b.Timezone != "" {
					fmt.Println(ui.KeyValue("Timezone", b.Timezone))
				}
				if tokenDisplay != "" {
					fmt.Println(ui.KeyValue("Ping Token", ui.Yellow(tokenDisplay)))
					if flagShowToken {
						fmt.Println(ui.KeyValue("Ping URL", fmt.Sprintf("%s/api/v1/beeper_apps/heartbeat/pings/%s", cfg.ServerURL, url.PathEscape(b.PingToken))))
					} else {
						fmt.Println(ui.KeyValue("Ping URL", fmt.Sprintf("%s/api/v1/beeper_apps/heartbeat/pings/%s", cfg.ServerURL, url.PathEscape(client.MaskToken(b.PingToken)))))
					}
				}
				if b.LastPingAt != "" {
					fmt.Println(ui.KeyValue("Last Ping", ui.FormatTimestamp(b.LastPingAt, b.Timezone)))
				}
				if b.NextRunAt != "" {
					fmt.Println(ui.KeyValue("Next Run", ui.FormatTimestamp(b.NextRunAt, b.Timezone)))
				}
				if b.LastRunAt != "" {
					fmt.Println(ui.KeyValue("Last Run", ui.FormatTimestamp(b.LastRunAt, b.Timezone)))
				}
				if len(b.NotificationChannels) > 0 {
					fmt.Println(ui.KeyValue("Channels", strings.Join(b.NotificationChannels, ", ")))
				}
				if b.Body != "" {
					fmt.Println()
					fmt.Println(ui.Dim("  Description:"))
					for _, line := range strings.Split(b.Body, "\n") {
						fmt.Printf("    %s\n", line)
					}
				}

				if b.RunStats != nil {
					fmt.Println()
					fmt.Println(ui.Section("  Run Statistics:"))
					fmt.Printf("    Total runs: %d | Succeeded: %d\n", b.RunStats.Total, b.RunStats.Succeeded)
				}

				if len(b.Runs) > 0 {
					fmt.Println()
					fmt.Println(ui.Section("  Recent Runs:"))
					tbl := ui.NewTable("RUN ID", "SCHEDULED FOR", "STATUS", "SIGNAL STATUS")
					tbl.SetIndent("    ")
					for _, r := range b.Runs {
						sigStatus := "-"
						if r.SignalStatus != "" {
							sigStatus = r.SignalStatus
						}
						tbl.AddRow(
							r.ID,
							ui.FormatTimestamp(r.ScheduledFor, b.Timezone),
							FormatRunStatus(r.Status),
							sigStatus,
						)
					}
					if err := tbl.Print(); err != nil {
						return err
					}
				}
				fmt.Println()
				return nil
			})
		},
	}

	cmd.Flags().BoolVar(&flagShowToken, "show-token", false, "Display the full unredacted ping token")
	return cmd
}
