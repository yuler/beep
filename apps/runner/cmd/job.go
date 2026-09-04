package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"beep-runner/internal/client"
	"beep-runner/internal/ui"
	"beep-runner/internal/workspace"

	"github.com/spf13/cobra"
)

var (
	flagJobName        string
	flagJobCron        string
	flagJobTimezone    string
	flagJobDescription string
	flagJobType        string
	flagJobTimeout     int
	flagJobNoSync      bool
	flagJobForce       bool
)

var jobCmd = &cobra.Command{
	Use:   "job",
	Short: "Manage workspace jobs and sync with Beep Core",
	Long: ui.Bold(ui.Cyan("Job Workspace Management")) + `

Like git, you can create local scripts, push definitions to Beep Core, pull remote jobs, and remove checks.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var jobCreateCmd = &cobra.Command{
	Use:   "create [slug]",
	Short: "Create a local job script scaffold (interactive if slug is omitted)",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		ws, err := workspace.Open(cfg.Workspace)
		if err != nil {
			return fmt.Errorf("workspace error: %w", err)
		}

		var slug string
		if len(args) > 0 {
			slug = strings.TrimSpace(strings.ToLower(args[0]))
		}

		createParams := ui.JobCreateParams{
			Slug:           slug,
			Name:           flagJobName,
			ScriptType:     flagJobType,
			Cron:           flagJobCron,
			Timezone:       flagJobTimezone,
			Description:    flagJobDescription,
			TimeoutSeconds: flagJobTimeout,
			SyncToServer:   !flagJobNoSync,
		}

		// If slug is omitted and interactive, run Huh inquiry prompt!
		if createParams.Slug == "" {
			if !flagNoInteractive && ui.IsInteractive() {
				prompted, err := ui.PromptJobCreate(createParams)
				if err != nil {
					return err
				}
				createParams = *prompted
			} else {
				return fmt.Errorf("job slug is required (e.g. beep-runner job create intranet-gateway)")
			}
		}

		filePath, created, err := ws.CreateScript(
			createParams.Slug,
			createParams.ScriptType,
			"",
			createParams.Name,
			createParams.Cron,
			createParams.Timezone,
			createParams.Description,
			createParams.TimeoutSeconds,
		)
		if err != nil {
			return fmt.Errorf("failed to create script: %w", err)
		}

		if created {
			fmt.Println(ui.Success("Created local script: %s", ui.Cyan(filePath)))
		} else {
			fmt.Println(ui.Info("Local script already exists: %s", ui.Cyan(filePath)))
			if meta, metaErr := workspace.ParseScriptMetadata(filePath); metaErr == nil {
				if createParams.Name == "" && meta.Name != "" {
					createParams.Name = meta.Name
				}
				if createParams.Cron == "*/5 * * * *" && meta.Cron != "" {
					createParams.Cron = meta.Cron
				}
				if createParams.Timezone == "" && meta.Timezone != "" {
					createParams.Timezone = meta.Timezone
				}
				if createParams.Description == "" && meta.Description != "" {
					createParams.Description = meta.Description
				}
				if createParams.TimeoutSeconds == 30 && meta.TimeoutSeconds > 0 {
					createParams.TimeoutSeconds = meta.TimeoutSeconds
				}
			}
		}

		if !createParams.SyncToServer {
			fmt.Println(ui.Dim("Skipping server sync (--no-sync specified)."))
			return nil
		}

		if cfg.ServerURL == "" || cfg.RunnerToken == "" {
			fmt.Println(ui.Info("Server URL or Runner Token not configured; skipping server sync."))
			fmt.Printf("  %s %s\n", ui.Dim("Tip: Configure once with"), ui.Cyan("beep-runner config set --server <url> --token <token>"))
			return nil
		}

		c := client.New(cfg)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		serverJob, err := c.CreateJob(ctx, &client.CreateJobRequest{
			Slug:           createParams.Slug,
			Name:           createParams.Name,
			Cron:           createParams.Cron,
			Timezone:       createParams.Timezone,
			TimeoutSeconds: createParams.TimeoutSeconds,
			Description:    createParams.Description,
		})
		if err != nil {
			return fmt.Errorf("failed to sync job to server: %w", err)
		}

		_ = ws.UpdateScriptID(filePath, serverJob.ID)

		fmt.Println(ui.Success("Successfully registered job on server: %s (%s: %s, %s: %s)",
			ui.Bold(serverJob.Name), ui.Dim("ID"), ui.Dim(serverJob.ID), ui.Dim("Cron"), ui.Yellow(serverJob.Cron)))
		fmt.Println()
		fmt.Println(ui.Section("Next steps:"))
		fmt.Printf("  1. Edit your script: %s\n", ui.Cyan(filePath))
		fmt.Printf("  2. Start runner daemon: %s\n", ui.Green("beep-runner run"))
		return nil
	},
}

var jobPushCmd = &cobra.Command{
	Use:   "push [slug]",
	Short: "Push local workspace job(s) to Beep Core (like git push)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runJobPushExec(cmd, args)
	},
}

var jobPullCmd = &cobra.Command{
	Use:   "pull [slug]",
	Short: "Pull server job(s) to local workspace (like git pull)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runJobPullExec(cmd, args)
	},
}

var jobRemoveCmd = &cobra.Command{
	Use:     "remove [slug]",
	Aliases: []string{"rm", "delete", "del"},
	Short:   "Remove local job script and delete from server",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
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
		} else if !flagNoInteractive && ui.IsInteractive() {
			selected, serverConfirm, err := ui.PromptJobRemove(localJobs)
			if err != nil {
				return err
			}
			targetSlugs = selected
			syncServer = serverConfirm
		} else {
			return fmt.Errorf("job slug is required (e.g. beep-runner job remove intranet-gateway)")
		}

		c := client.New(cfg)

		for _, slug := range targetSlugs {
			removedFiles, removedFromJSON, err := ws.RemoveJob(slug)
			if err != nil {
				fmt.Println(ui.Info("%v", err))
			} else {
				for _, f := range removedFiles {
					fmt.Println(ui.Success("Removed local job script: %s", ui.Cyan(f)))
				}
				if removedFromJSON {
					fmt.Println(ui.Success("Removed %q from jobs.json", ui.Bold(slug)))
				}
			}

			if syncServer && cfg.ServerURL != "" && cfg.RunnerToken != "" {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				if delErr := c.DeleteJob(ctx, slug); delErr != nil {
					fmt.Println(ui.Warn("Failed to delete job on server: %v", delErr))
				} else {
					fmt.Println(ui.Success("Deleted job %q from server", ui.Bold(slug)))
				}
				cancel()
			}
		}
		return nil
	},
}

var jobListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List workspace jobs and compare with server",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
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

		localMap := make(map[string]workspace.LocalJob)
		for _, lj := range localJobs {
			localMap[strings.ToLower(lj.Slug)] = lj
		}

		var serverJobs []*client.ServerJob
		var serverErr error
		if cfg.ServerURL != "" && cfg.RunnerToken != "" {
			c := client.New(cfg)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			serverJobs, serverErr = c.ListJobs(ctx)
			cancel()
		}

		if serverErr != nil {
			fmt.Println(ui.Warn("Warning: Could not fetch server jobs: %v", serverErr))
		}

		// If no server configured or server call failed, just show local jobs
		if cfg.ServerURL == "" || cfg.RunnerToken == "" || serverErr != nil {
			fmt.Printf("%s (%s):\n", ui.Bold(ui.Cyan("Local Workspace Jobs")), ui.Dim(ws.Root))
			if len(localJobs) == 0 {
				fmt.Println(ui.Dim("  (No local jobs found in jobs/)"))
			} else {
				for _, j := range localJobs {
					desc := fmt.Sprintf("[%s, %s: %s]", ui.Bold(j.Name), ui.Dim("cron"), ui.Yellow(j.Cron))
					if j.TimeoutSeconds > 0 && j.TimeoutSeconds != 30 {
						desc += fmt.Sprintf(" (%ds)", j.TimeoutSeconds)
					}
					idStr := ""
					if j.ID != "" {
						idStr = fmt.Sprintf(" (ID: %s)", ui.Dim(j.ID))
					}
					if j.FilePath != "" {
						fmt.Printf("  %s %-18s %-36s%s -> %s\n", ui.Bullet(), ui.Cyan(j.Slug), desc, idStr, ui.Dim(j.FilePath))
					} else {
						fmt.Printf("  %s %-18s %-36s%s -> %s (jobs.json)\n", ui.Bullet(), ui.Cyan(j.Slug), desc, idStr, ui.Dim(strings.Join(j.Command, " ")))
					}
				}
			}
			return nil
		}

		serverMap := make(map[string]*client.ServerJob)
		for _, sj := range serverJobs {
			serverMap[strings.ToLower(sj.Slug)] = sj
		}

		var allSlugs []string
		seen := make(map[string]bool)
		for _, lj := range localJobs {
			slug := strings.ToLower(lj.Slug)
			if !seen[slug] {
				seen[slug] = true
				allSlugs = append(allSlugs, slug)
			}
		}
		for _, sj := range serverJobs {
			slug := strings.ToLower(sj.Slug)
			if !seen[slug] {
				seen[slug] = true
				allSlugs = append(allSlugs, slug)
			}
		}

		fmt.Printf("%s\n", ui.Bold(ui.Cyan("Workspace & Server Job Status:")))
		fmt.Printf("  %s %s  |  %s %s\n\n", ui.Dim("Workspace:"), ui.Bold(ws.Root), ui.Dim("Server:"), ui.Bold(cfg.ServerURL))

		if len(allSlugs) == 0 {
			fmt.Println(ui.Dim("  (No jobs found locally or on server)"))
			fmt.Printf("  Create one with: %s\n", ui.Cyan("beep-runner job create <slug>"))
			return nil
		}

		for _, slug := range allSlugs {
			lj, hasLocal := localMap[slug]
			sj, hasServer := serverMap[slug]

			if hasLocal && hasServer {
				serverDesc := ""
				if sj.Config != nil {
					if d, ok := sj.Config["description"].(string); ok {
						serverDesc = d
					}
				}

				var diffs []string
				if lj.Name != "" && sj.Name != "" && lj.Name != sj.Name {
					diffs = append(diffs, fmt.Sprintf("name: local %q != server %q", lj.Name, sj.Name))
				}
				if lj.Cron != "" && sj.Cron != "" && lj.Cron != sj.Cron {
					diffs = append(diffs, fmt.Sprintf("schedule: local %q != server %q", lj.Cron, sj.Cron))
				}
				serverTz := sj.Timezone
				if serverTz == "" {
					serverTz = "UTC"
				}
				localTz := lj.Timezone
				if localTz == "" {
					localTz = "UTC"
				}
				if !strings.EqualFold(localTz, serverTz) {
					diffs = append(diffs, fmt.Sprintf("timezone: local %q != server %q", localTz, serverTz))
				}
				serverTimeout := sj.TimeoutSeconds
				if serverTimeout <= 0 {
					serverTimeout = 30
				}
				localTimeout := lj.TimeoutSeconds
				if localTimeout <= 0 {
					localTimeout = 30
				}
				if localTimeout != serverTimeout {
					diffs = append(diffs, fmt.Sprintf("timeout: local %ds != server %ds", localTimeout, serverTimeout))
				}
				if lj.Description != "" && serverDesc != "" && lj.Description != serverDesc {
					diffs = append(diffs, fmt.Sprintf("description: local %q != server %q", lj.Description, serverDesc))
				}

				if len(diffs) == 0 {
					fmt.Printf("  %s %-18s %s %s\n",
						ui.Green("✓"),
						ui.Bold(ui.Cyan(slug)),
						ui.Green("[synced]"),
						ui.Dim(fmt.Sprintf("(%s, %s, %s)", lj.Name, lj.Cron, sj.Status)),
					)
					fmt.Printf("    %s %s (%s: %s)\n", ui.Dim("File:"), ui.Dim(lj.FilePath), ui.Dim("ID"), ui.Dim(sj.ID))
				} else {
					fmt.Printf("  %s %-18s %s %s\n",
						ui.Yellow("!"),
						ui.Bold(ui.Cyan(slug)),
						ui.Yellow("[out of sync / modified]"),
						ui.Dim(fmt.Sprintf("(%s)", lj.Name)),
					)
					fmt.Printf("    %s %s (%s: %s)\n", ui.Dim("File:"), ui.Dim(lj.FilePath), ui.Dim("ID"), ui.Dim(sj.ID))
					for _, d := range diffs {
						fmt.Printf("    %s %s\n", ui.Yellow("•"), ui.Yellow(d))
					}
					fmt.Printf("    %s Run %s to push local, or %s to pull server\n",
						ui.Dim("Tip:"),
						ui.Cyan(fmt.Sprintf("beep-runner job push %s", slug)),
						ui.Cyan(fmt.Sprintf("beep-runner job pull %s", slug)),
					)
				}
			} else if hasLocal && !hasServer {
				fmt.Printf("  %s %-18s %s %s\n",
					ui.Cyan("●"),
					ui.Bold(ui.Cyan(slug)),
					ui.Cyan("[local only]"),
					ui.Dim(fmt.Sprintf("(%s, %s)", lj.Name, lj.Cron)),
				)
				fmt.Printf("    %s %s\n", ui.Dim("File:"), ui.Dim(lj.FilePath))
				fmt.Printf("    %s Run %s to register on server\n", ui.Dim("Tip:"), ui.Cyan(fmt.Sprintf("beep-runner job push %s", slug)))
			} else if !hasLocal && hasServer {
				fmt.Printf("  %s %-18s %s %s\n",
					ui.Magenta("●"),
					ui.Bold(ui.Cyan(slug)),
					ui.Magenta("[remote only]"),
					ui.Dim(fmt.Sprintf("(%s, %s, %s)", sj.Name, sj.Cron, sj.Status)),
				)
				fmt.Printf("    %s %s\n", ui.Dim("ID:"), ui.Dim(sj.ID))
				fmt.Printf("    %s Run %s to pull script to workspace\n", ui.Dim("Tip:"), ui.Cyan(fmt.Sprintf("beep-runner job pull %s", slug)))
			}
			fmt.Println()
		}

		return nil
	},
}

func init() {
	jobCreateCmd.Flags().StringVar(&flagJobName, "name", "", "Display name of the job")
	jobCreateCmd.Flags().StringVar(&flagJobCron, "cron", "*/5 * * * *", "Cron schedule expression")
	jobCreateCmd.Flags().StringVar(&flagJobTimezone, "timezone", "", "Timezone (e.g. Asia/Shanghai, UTC)")
	jobCreateCmd.Flags().StringVar(&flagJobTimezone, "tz", "", "Timezone alias")
	jobCreateCmd.Flags().StringVar(&flagJobDescription, "description", "", "Description of the job")
	jobCreateCmd.Flags().StringVar(&flagJobDescription, "desc", "", "Description alias")
	jobCreateCmd.Flags().StringVar(&flagJobType, "type", "sh", "Script type: sh, py, js, rb")
	jobCreateCmd.Flags().IntVar(&flagJobTimeout, "timeout", 30, "Timeout in seconds")
	jobCreateCmd.Flags().BoolVar(&flagJobNoSync, "no-sync", false, "Do not sync to server")

	jobPullCmd.Flags().BoolVar(&flagJobForce, "force", false, "Overwrite existing local scripts with server definition")

	jobRemoveCmd.Flags().BoolVar(&flagJobNoSync, "no-sync", false, "Remove local script only without deleting from server")

	jobCmd.AddCommand(jobCreateCmd)
	jobCmd.AddCommand(jobPushCmd)
	jobCmd.AddCommand(jobPullCmd)
	jobCmd.AddCommand(jobRemoveCmd)
	jobCmd.AddCommand(jobListCmd)
}

func runJobPushExec(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
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
		fmt.Printf("Create one with: %s\n", ui.Cyan("beep-runner job create <slug>"))
		return nil
	}

	var targetSlugs []string
	if len(args) > 0 {
		targetSlugs = []string{strings.TrimSpace(strings.ToLower(args[0]))}
	} else if !flagNoInteractive && ui.IsInteractive() && len(localJobs) > 1 {
		selected, err := ui.PromptJobPushSelection(localJobs)
		if err != nil {
			return err
		}
		targetSlugs = selected
	} else {
		for _, j := range localJobs {
			targetSlugs = append(targetSlugs, j.Slug)
		}
	}

	targetMap := make(map[string]bool)
	for _, s := range targetSlugs {
		targetMap[strings.ToLower(s)] = true
	}

	var syncReqs []*client.CreateJobRequest
	for _, job := range localJobs {
		if !targetMap[strings.ToLower(job.Slug)] {
			continue
		}
		syncReqs = append(syncReqs, &client.CreateJobRequest{
			Slug:           job.Slug,
			Name:           job.Name,
			Cron:           job.Cron,
			Timezone:       job.Timezone,
			TimeoutSeconds: job.TimeoutSeconds,
			Description:    job.Description,
		})
	}

	if len(syncReqs) == 0 {
		return fmt.Errorf("no matching local jobs found to push")
	}

	c := client.New(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	synced, err := c.SyncJobs(ctx, syncReqs)
	if err != nil {
		return fmt.Errorf("failed to push jobs to server: %w", err)
	}

	fmt.Println(ui.Success("Successfully pushed %d job(s) to server (%s):", len(synced), ui.Dim(cfg.ServerURL)))
	for _, j := range synced {
		tz := j.Timezone
		if tz == "" {
			tz = "UTC"
		}
		// Update local script ID
		if fpath, found := ws.FindScriptFile(j.Slug); found {
			_ = ws.UpdateScriptID(fpath, j.ID)
		}
		fmt.Printf("  %s %s [%s, %s: %s] (%s, %s: %s)\n",
			ui.Bullet(), ui.Cyan(j.Slug), ui.Bold(j.Name), ui.Dim("cron"), ui.Yellow(j.Cron), ui.Dim(tz), ui.Dim("ID"), ui.Dim(j.ID))
	}
	return nil
}

func runJobPullExec(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
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

	serverJobs, err := c.ListJobs(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch server jobs: %w", err)
	}

	if len(serverJobs) == 0 {
		fmt.Println(ui.Info("No jobs registered on server (%s)", cfg.ServerURL))
		return nil
	}

	var targetSlugs []string
	force := flagJobForce

	if len(args) > 0 {
		targetSlugs = []string{strings.TrimSpace(strings.ToLower(args[0]))}
	} else if !flagNoInteractive && ui.IsInteractive() && len(serverJobs) > 1 {
		selected, f, err := ui.PromptJobPullSelection(serverJobs)
		if err != nil {
			return err
		}
		targetSlugs = selected
		force = f || force
	} else {
		for _, j := range serverJobs {
			targetSlugs = append(targetSlugs, j.Slug)
		}
	}

	targetMap := make(map[string]bool)
	for _, s := range targetSlugs {
		targetMap[strings.ToLower(s)] = true
	}

	pulledCount := 0

	for _, sj := range serverJobs {
		if !targetMap[strings.ToLower(sj.Slug)] {
			continue
		}

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
			fmt.Println(ui.Success("Updated local script header: %s (%s: %s)", ui.Cyan(filePath), ui.Dim("ID"), ui.Dim(sj.ID)))
		}
	}

	fmt.Println()
	if pulledCount > 0 {
		fmt.Println(ui.Success("Successfully pulled %d job(s) from server (%s)", pulledCount, ui.Dim(cfg.ServerURL)))
	}
	return nil
}
