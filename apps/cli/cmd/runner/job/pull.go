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

// NewCmdPull creates the 'runner job pull' subcommand.
func NewCmdPull() *cobra.Command {
	var flagJobForce bool

	cmd := &cobra.Command{
		Use:   "pull [slug]",
		Short: "Pull server job(s) to local workspace (like git pull)",
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

			c := client.New(cfg)
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			var serverJobs []*client.ServerJob
			err = ui.WithSpinner("Fetching server jobs...", func() error {
				var listErr error
				serverJobs, listErr = c.ListJobs(ctx)
				return listErr
			})
			if err != nil {
				return fmt.Errorf("failed to fetch server jobs: %w", err)
			}

			if len(serverJobs) == 0 {
				fmt.Println(ui.Info("No jobs registered on server (%s)", cfg.ServerURL))
				return nil
			}

			localJobs, _ := ws.ListJobs()
			force := flagJobForce
			items := ui.PairJobs(localJobs, serverJobs)

			var itemsToPull []ui.JobCompareItem
			if len(args) > 0 {
				arg := strings.TrimSpace(strings.ToLower(args[0]))
				for _, it := range items {
					if it.ServerJob == nil {
						continue
					}
					match := strings.EqualFold(it.ID, arg) ||
						strings.EqualFold(it.ServerJob.Slug, arg) ||
						(it.LocalJob != nil && strings.EqualFold(it.LocalJob.Slug, arg))
					if match {
						itemsToPull = append(itemsToPull, it)
					}
				}
				if len(itemsToPull) == 0 {
					return fmt.Errorf("no server job matching %q found to pull", args[0])
				}
			} else if cmdutil.IsInteractive(cmd) {
				var remoteItems []ui.JobCompareItem
				for _, it := range items {
					if it.ServerJob != nil {
						remoteItems = append(remoteItems, it)
					}
				}
				selectedIDs, err := ui.PromptJobPullSelection(remoteItems)
				if err != nil {
					return err
				}
				selectedMap := make(map[string]bool)
				for _, s := range selectedIDs {
					selectedMap[strings.ToLower(s)] = true
				}
				for _, it := range remoteItems {
					key := it.ID
					if key == "" {
						key = it.ServerJob.Slug
					}
					if selectedMap[strings.ToLower(key)] || selectedMap[strings.ToLower(it.ServerJob.Slug)] || (it.LocalJob != nil && selectedMap[strings.ToLower(it.LocalJob.Slug)]) {
						itemsToPull = append(itemsToPull, it)
					}
				}
			} else {
				for _, it := range items {
					if it.ServerJob != nil {
						itemsToPull = append(itemsToPull, it)
					}
				}
			}

			pulledCount := 0

			for _, it := range itemsToPull {
				sj := it.ServerJob
				desc := ""
				if sj.Config != nil {
					if d, ok := sj.Config["description"].(string); ok {
						desc = d
					}
				}

				filePath, created, pullErr := ws.PullJob(sj.Slug, "sh", sj.ID, sj.Name, sj.Cron, sj.Timezone, desc, sj.TimeoutSeconds, force)
				if pullErr != nil {
					fmt.Printf("  %s %s: %v\n", ui.Error("Failed to pull"), ui.Cyan(sj.Slug), pullErr)
					continue
				}

				pulledCount++
				if created {
					fmt.Println(ui.Success("Created local script: %s (%s: %s)", ui.Cyan(filePath), ui.Dim("ID"), ui.Dim(sj.ID)))
				} else {
					fmt.Println(ui.Success("Updated local script: %s (%s: %s)", ui.Cyan(filePath), ui.Dim("ID"), ui.Dim(sj.ID)))
				}
			}

			fmt.Println()
			if pulledCount > 0 {
				fmt.Println(ui.Success("Successfully pulled %d job(s) from server (%s)", pulledCount, ui.Dim(cfg.ServerURL)))
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&flagJobForce, "force", false, "Overwrite existing local scripts with server definition")
	return cmd
}
