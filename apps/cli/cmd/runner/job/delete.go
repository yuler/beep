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

// NewCmdDelete creates the 'runner job delete' subcommand.
func NewCmdDelete() *cobra.Command {
	var flagJobNoSync bool

	cmd := &cobra.Command{
		Use:     "delete [slug]",
		Aliases: []string{"remove", "rm", "del"},
		Short:   "Remove local job script and delete from server",
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

			var targetSlugs []string
			syncServer := !flagJobNoSync

			if len(args) > 0 {
				targetSlugs = []string{strings.TrimSpace(strings.ToLower(args[0]))}
			} else if cmdutil.IsInteractive(cmd) {
				selected, err := ui.PromptJobRemove(localJobs)
				if err != nil {
					return err
				}
				targetSlugs = selected
			} else {
				return fmt.Errorf("job slug is required (e.g. beep runner job delete intranet-gateway)")
			}

			c := client.New(cfg)

			for _, slug := range targetSlugs {
				// Prefer @id for server delete so a local rename still removes the paired remote job.
				deleteKey := slug
				for _, lj := range localJobs {
					if strings.EqualFold(lj.Slug, slug) && strings.TrimSpace(lj.ID) != "" {
						deleteKey = strings.TrimSpace(lj.ID)
						break
					}
				}

				removedFiles, err := ws.RemoveJob(slug)
				if err != nil {
					fmt.Println(ui.Info("%v", err))
					continue
				}
				for _, f := range removedFiles {
					fmt.Println(ui.Success("Removed local job script: %s", ui.Cyan(f)))
				}

				if syncServer && cfg.ServerURL != "" && cfg.RunnerToken != "" {
					ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					err := ui.WithSpinner(fmt.Sprintf("Deleting job %s from server...", slug), func() error {
						delErr := c.DeleteJob(ctx, deleteKey)
						if delErr != nil && deleteKey != slug {
							delErr = c.DeleteJob(ctx, slug)
						}
						return delErr
					})
					cancel()
					if err != nil {
						fmt.Println(ui.Warn("Failed to delete job %s on server: %v", ui.Bold(slug), err))
					} else {
						fmt.Println(ui.Success("Deleted job %s from server", ui.Bold(slug)))
					}
				}
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&flagJobNoSync, "no-sync", false, "Remove local script only without deleting from server")
	return cmd
}
