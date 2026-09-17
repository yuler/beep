package job

import (
	"context"
	"fmt"
	"strings"
	"time"

	"beep/internal/client"
	"beep/internal/cmdutil"
	"beep/internal/ui"
	"beep/internal/workspace"

	"github.com/spf13/cobra"
)

// NewCmdPush creates the 'runner job push' subcommand.
func NewCmdPush() *cobra.Command {
	return &cobra.Command{
		Use:   "push [slug]",
		Short: "Push local workspace job(s) to Beep Core (like git push)",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig(cmd)
			if err != nil {
				return err
			}
			if err := cfg.Validate(); err != nil {
				return fmt.Errorf("configuration error: %w", err)
			}

			ws, err := workspace.Open(cfg.Workspace)
			if err != nil {
				return fmt.Errorf("workspace error: %w", err)
			}

			localJobs, err := ws.ListJobs()
			if err != nil {
				return fmt.Errorf("failed to list local jobs: %w", err)
			}

			if len(localJobs) == 0 {
				fmt.Println(ui.Warn("No local jobs found in %s/jobs", ws.Root))
				fmt.Printf("Create one with: %s\n", ui.Cyan("beep runner job create <slug>"))
				return nil
			}

			var serverJobs []*client.ServerJob
			if cfg.ServerURL != "" && cfg.RunnerToken != "" {
				c := client.New(cfg)
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				_ = ui.WithSpinner("Fetching server jobs...", func() error {
					var fetchErr error
					serverJobs, fetchErr = c.ListJobs(ctx)
					return fetchErr
				})
			}

			items := ui.PairJobs(localJobs, serverJobs)

			var itemsToPush []ui.JobCompareItem
			if len(args) > 0 {
				arg := strings.TrimSpace(strings.ToLower(args[0]))
				for _, it := range items {
					if it.LocalJob == nil {
						continue
					}
					match := strings.EqualFold(it.ID, arg) ||
						strings.EqualFold(it.LocalJob.Slug, arg) ||
						(it.ServerJob != nil && strings.EqualFold(it.ServerJob.Slug, arg))
					if match {
						itemsToPush = append(itemsToPush, it)
					}
				}
				if len(itemsToPush) == 0 {
					return fmt.Errorf("no local job matching %q found to push", args[0])
				}
			} else if cmdutil.IsInteractive(cmd) {
				var localItems []ui.JobCompareItem
				for _, it := range items {
					if it.LocalJob != nil {
						localItems = append(localItems, it)
					}
				}
				selectedIDs, err := ui.PromptJobPushSelection(localItems)
				if err != nil {
					return err
				}
				selectedMap := make(map[string]bool)
				for _, s := range selectedIDs {
					selectedMap[strings.ToLower(s)] = true
				}
				for _, it := range localItems {
					key := it.ID
					if key == "" {
						key = it.LocalJob.Slug
					}
					if selectedMap[strings.ToLower(key)] || selectedMap[strings.ToLower(it.LocalJob.Slug)] || (it.ID != "" && selectedMap[strings.ToLower(it.ID)]) {
						itemsToPush = append(itemsToPush, it)
					}
				}
			} else {
				for _, it := range items {
					if it.LocalJob != nil {
						itemsToPush = append(itemsToPush, it)
					}
				}
			}

			var pushReqs []*client.CreateJobRequest
			for _, it := range itemsToPush {
				j := it.LocalJob
				pushReqs = append(pushReqs, &client.CreateJobRequest{
					ID:             j.ID,
					Slug:           j.Slug,
					Name:           j.Name,
					Cron:           j.Cron,
					Timezone:       j.Timezone,
					TimeoutSeconds: j.TimeoutSeconds,
					Description:    j.Description,
				})
			}

			if len(pushReqs) == 0 {
				return fmt.Errorf("no matching local jobs found to push")
			}

			c := client.New(cfg)
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			var pushed []*client.ServerJob
			err = ui.WithSpinner(fmt.Sprintf("Pushing %d job(s) to server...", len(pushReqs)), func() error {
				var pushErr error
				pushed, pushErr = c.PushJobs(ctx, pushReqs)
				return pushErr
			})
			if err != nil {
				return fmt.Errorf("failed to push jobs to server: %w", err)
			}

			fmt.Println(ui.Success("Successfully pushed %d job(s) to server (%s):", len(pushed), ui.Dim(cfg.ServerURL)))
			for _, j := range pushed {
				if fpath, found := ws.FindScriptFileByID(j.ID); found {
					_ = ws.UpdateScriptID(fpath, j.ID)
				} else if fpath, found := ws.FindScriptFile(j.Slug); found {
					_ = ws.UpdateScriptID(fpath, j.ID)
				}
				fmt.Printf("  %s %-20s %s: %s  %s: %s\n",
					ui.Bullet(), ui.Cyan(j.Slug), ui.Dim("ID"), ui.Bold(j.ID), ui.Dim("cron"), ui.Yellow(j.Cron))
			}
			return nil
		},
	}
}
