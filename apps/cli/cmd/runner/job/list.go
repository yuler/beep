package job

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"beep/internal/client"
	"beep/internal/cmdutil"
	"beep/internal/config"
	"beep/internal/ui"
	"beep/internal/workspace"

	"github.com/spf13/cobra"
)

// NewCmdList creates the 'runner job list' subcommand.
func NewCmdList() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List workspace jobs and compare with server",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig(cmd)
			if err != nil {
				return err
			}

			ws, err := workspace.Open(cfg.Workspace)
			if err != nil {
				return fmt.Errorf("workspace error: %w", err)
			}

			localJobs, err := ws.ListJobs()
			if err != nil {
				return fmt.Errorf("failed to list local jobs: %w", err)
			}

			var serverJobs []*client.ServerJob
			var serverErr error
			if cfg.ServerURL != "" && cfg.RunnerToken != "" {
				c := client.New(cfg)
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				_ = ui.WithSpinner("Fetching server jobs...", func() error {
					serverJobs, serverErr = c.ListJobs(ctx)
					return serverErr
				})
			}

			if serverErr != nil {
				fmt.Println(ui.Warn("Warning: Could not fetch server jobs: %v", serverErr))
			}

			// JSON output
			if cmdutil.IsJSON(cmd) {
				items := ui.PairJobs(localJobs, serverJobs)
				data, err := json.MarshalIndent(items, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(data))
				return nil
			}

			// If no server configured or server call failed, just show local jobs
			if cfg.ServerURL == "" || cfg.RunnerToken == "" || serverErr != nil {
				fmt.Printf("%s %s\n\n", ui.Bold(ui.Cyan("Local Workspace Jobs")), ui.Dim(fmt.Sprintf("(workspace: %s)", ws.Root)))
				if len(localJobs) == 0 {
					fmt.Println(ui.Dim(fmt.Sprintf("  No local jobs found. Create one with '%s runner job create <slug>'.", config.BinaryName())))
					return nil
				}

				tbl := ui.NewTable("SLUG", "SCHEDULE", "TIMEZONE", "ID", "FILE")
				tbl.SetIndent("  ")
				for _, j := range localJobs {
					tz := displayTimezone(j.Timezone)
					cronStr := j.Cron
					if j.TimeoutSeconds > 0 && j.TimeoutSeconds != 30 {
						cronStr += fmt.Sprintf(" (%ds)", j.TimeoutSeconds)
					}
					idStr := j.ID
					if idStr == "" {
						idStr = ui.Dim("-")
					}
					tbl.AddRow(
						ui.Cyan(j.Slug),
						cronStr,
						tz,
						idStr,
						ui.Dim(j.FilePath),
					)
				}
				if err := tbl.Print(); err != nil {
					return err
				}
				fmt.Println()
				return nil
			}

			items := ui.PairJobs(localJobs, serverJobs)

			fmt.Printf("%s %s\n\n", ui.Bold(ui.Cyan("Workspace & Server Jobs")), ui.Dim(fmt.Sprintf("(workspace: %s)", ws.Root)))

			if len(items) == 0 {
				fmt.Println(ui.Dim(fmt.Sprintf("  No jobs found locally or on server. Create one with '%s runner job create <slug>'.", config.BinaryName())))
				return nil
			}

			tbl := ui.NewTable("STATUS", "SLUG", "SCHEDULE", "TIMEZONE", "ID", "DETAILS")
			tbl.SetIndent("  ")

			for _, item := range items {
				statusStr := ""
				cronStr := "-"
				tzStr := "-"
				idStr := item.ID
				if idStr == "" {
					idStr = "-"
				}
				details := ""

				switch item.Status {
				case ui.StatusSynced:
					statusStr = ui.Green("synced")
					if item.LocalJob != nil {
						cronStr = item.LocalJob.Cron
						tzStr = displayTimezone(item.LocalJob.Timezone)
					}
					if item.ServerJob != nil && item.ServerJob.Status != "" {
						details = ui.Dim(item.ServerJob.Status)
					}
				case ui.StatusModified:
					statusStr = ui.Yellow("modified")
					if item.LocalJob != nil {
						cronStr = item.LocalJob.Cron
						tzStr = displayTimezone(item.LocalJob.Timezone)
					}
					if len(item.Diffs) > 0 {
						details = ui.Yellow(strings.Join(item.Diffs, "; "))
					}
				case ui.StatusLocalOnly:
					statusStr = ui.Cyan("local only")
					if item.LocalJob != nil {
						cronStr = item.LocalJob.Cron
						tzStr = displayTimezone(item.LocalJob.Timezone)
					}
					details = ui.Dim("push to sync")
				case ui.StatusRemoteOnly:
					statusStr = ui.Magenta("remote only")
					if item.ServerJob != nil {
						cronStr = item.ServerJob.Cron
						tzStr = displayTimezone(item.ServerJob.Timezone)
						details = ui.Dim(item.ServerJob.Status)
					}
				}

				tbl.AddRow(
					statusStr,
					ui.Bold(item.Slug),
					cronStr,
					tzStr,
					ui.Dim(idStr),
					details,
				)
			}

			if err := tbl.Print(); err != nil {
				return err
			}
			fmt.Println()
			return nil
		},
	}
}

func displayTimezone(tz string) string {
	tz = strings.TrimSpace(tz)
	if tz == "" {
		return "UTC"
	}
	return tz
}
