package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"beep-runner/internal/client"
	"beep-runner/internal/schedule"
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

		// If interactive, run inquiry prompt (prefilling slug/name/flags if provided)
		if !flagNoInteractive && ui.IsInteractive() {
			prompted, err := ui.PromptJobCreate(createParams)
			if err != nil {
				return err
			}
			createParams = *prompted
		} else if createParams.Slug == "" {
			return fmt.Errorf("job slug is required (e.g. beep-runner job create intranet-gateway)")
		}

		if err := schedule.Validate(createParams.Cron); err != nil {
			return fmt.Errorf("invalid --cron: %w", err)
		}
		if createParams.Timezone != "" && !workspace.ValidIANATimezone(createParams.Timezone) {
			return fmt.Errorf("invalid --timezone %q: not a valid IANA timezone", createParams.Timezone)
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

		jobID := ""
		if meta, metaErr := workspace.ParseScriptMetadata(filePath); metaErr == nil {
			jobID = meta.ID
			if !created {
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

		if created {
			fmt.Println(ui.Success("Created local script: %s", ui.Cyan(filePath)))
		} else {
			fmt.Println(ui.Info("Local script already exists: %s", ui.Cyan(filePath)))
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
			ID:             jobID,
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
		fmt.Printf("  2. Start runner daemon: %s\n", ui.Green("beep-runner up"))
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
			selected, err := ui.PromptJobRemove(localJobs)
			if err != nil {
				return err
			}
			targetSlugs = selected
		} else {
			return fmt.Errorf("job slug is required (e.g. beep-runner job remove intranet-gateway)")
		}

		c := client.New(cfg)

		for _, slug := range targetSlugs {
			removedFiles, removedFromJSON, err := ws.RemoveJob(slug)
			if err != nil {
				fmt.Println(ui.Info("%v", err))
				continue
			}
			for _, f := range removedFiles {
				fmt.Println(ui.Success("Removed local job script: %s", ui.Cyan(f)))
			}
			if removedFromJSON {
				fmt.Println(ui.Success("Removed %s from jobs.json", ui.Bold(slug)))
			}

			if syncServer && cfg.ServerURL != "" && cfg.RunnerToken != "" {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				if delErr := c.DeleteJob(ctx, slug); delErr != nil {
					fmt.Println(ui.Warn("Failed to delete job on server: %v", delErr))
				} else {
					fmt.Println(ui.Success("Deleted job %s from server", ui.Bold(slug)))
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
					printLocalJobLine(j)
				}
			}
			return nil
		}

		items := ui.PairJobs(localJobs, serverJobs)

		fmt.Printf("%s\n", ui.Bold(ui.Cyan("Workspace & Server Job Status:")))
		fmt.Printf("  %s %s  |  %s %s\n\n", ui.Dim("Workspace:"), ui.Bold(ws.Root), ui.Dim("Server:"), ui.Bold(cfg.ServerURL))

		if len(items) == 0 {
			fmt.Println(ui.Dim("  (No jobs found locally or on server)"))
			fmt.Printf("  Create one with: %s\n", ui.Cyan("beep-runner job create <slug>"))
			return nil
		}

		for _, item := range items {
			slug := item.Slug
			switch item.Status {
			case ui.StatusSynced:
				tz := displayTimezone(item.LocalJob.Timezone)
				id := item.ID
				if id == "" && item.ServerJob != nil {
					id = item.ServerJob.ID
				}
				fmt.Printf("  %s %-18s %s %s\n",
					ui.Green("✓"),
					ui.Bold(ui.Cyan(slug)),
					ui.Green("[synced]"),
					ui.Dim(fmt.Sprintf("(cron: %s, tz: %s, %s)", item.LocalJob.Cron, tz, item.ServerJob.Status)),
				)
				fmt.Printf("    %s %s\n", ui.Dim("File:"), ui.Dim(item.LocalJob.FilePath))
				fmt.Printf("    %s %s\n", ui.Dim("ID:"), ui.Dim(displayJobID(id)))
			case ui.StatusModified:
				id := item.ID
				if id == "" && item.ServerJob != nil {
					id = item.ServerJob.ID
				}
				fmt.Printf("  %s %-18s %s\n",
					ui.Yellow("!"),
					ui.Bold(ui.Cyan(slug)),
					ui.Yellow("[out of sync / modified]"),
				)
				if item.LocalJob != nil && item.LocalJob.FilePath != "" {
					fmt.Printf("    %s %s\n", ui.Dim("File:"), ui.Dim(item.LocalJob.FilePath))
				}
				fmt.Printf("    %s %s\n", ui.Dim("ID:"), ui.Dim(displayJobID(id)))
				fmt.Printf("    %s cron %s  tz %s\n",
					ui.Dim("Local:"),
					ui.Yellow(item.LocalJob.Cron),
					ui.Dim(displayTimezone(item.LocalJob.Timezone)),
				)
				for _, d := range item.Diffs {
					fmt.Printf("    %s %s\n", ui.Yellow("•"), ui.Yellow(d))
				}
				pushSlug := slug
				if item.LocalJob != nil && item.LocalJob.Slug != "" {
					pushSlug = item.LocalJob.Slug
				}
				pullSlug := slug
				if item.ServerJob != nil && item.ServerJob.Slug != "" {
					pullSlug = item.ServerJob.Slug
				}
				fmt.Printf("    %s Run %s to push local, or %s to pull server\n",
					ui.Dim("Tip:"),
					ui.Cyan(fmt.Sprintf("beep-runner job push %s", pushSlug)),
					ui.Cyan(fmt.Sprintf("beep-runner job pull %s", pullSlug)),
				)
			case ui.StatusLocalOnly:
				tz := displayTimezone(item.LocalJob.Timezone)
				fmt.Printf("  %s %-18s %s %s\n",
					ui.Cyan("●"),
					ui.Bold(ui.Cyan(slug)),
					ui.Cyan("[local only]"),
					ui.Dim(fmt.Sprintf("(cron: %s, tz: %s)", item.LocalJob.Cron, tz)),
				)
				if item.LocalJob.FilePath != "" {
					fmt.Printf("    %s %s\n", ui.Dim("File:"), ui.Dim(item.LocalJob.FilePath))
				}
				fmt.Printf("    %s %s\n", ui.Dim("ID:"), ui.Dim(displayJobID(item.LocalJob.ID)))
				fmt.Printf("    %s Run %s to register on server\n", ui.Dim("Tip:"), ui.Cyan(fmt.Sprintf("beep-runner job push %s", slug)))
			case ui.StatusRemoteOnly:
				tz := displayTimezone(item.ServerJob.Timezone)
				fmt.Printf("  %s %-18s %s %s\n",
					ui.Magenta("●"),
					ui.Bold(ui.Cyan(slug)),
					ui.Magenta("[remote only]"),
					ui.Dim(fmt.Sprintf("(cron: %s, tz: %s, %s)", item.ServerJob.Cron, tz, item.ServerJob.Status)),
				)
				fmt.Printf("    %s %s\n", ui.Dim("ID:"), ui.Dim(displayJobID(item.ServerJob.ID)))
				fmt.Printf("    %s Run %s to pull script to workspace\n", ui.Dim("Tip:"), ui.Cyan(fmt.Sprintf("beep-runner job pull %s", slug)))
			}
			fmt.Println()
		}

		return nil
	},
}

func init() {
	jobCreateCmd.Flags().StringVar(&flagJobName, "name", "", "Display name of the job")
	jobCreateCmd.Flags().StringVar(&flagJobCron, "cron", "*/5 * * * *", "Schedule: classic cron or Fugit semantic (e.g. every 5 minutes)")
	jobCreateCmd.Flags().StringVar(&flagJobTimezone, "timezone", "", "Timezone (e.g. Asia/Shanghai, UTC)")
	jobCreateCmd.Flags().StringVar(&flagJobTimezone, "tz", "", "Timezone alias")
	jobCreateCmd.Flags().StringVar(&flagJobDescription, "description", "", "Description of the job")
	jobCreateCmd.Flags().StringVar(&flagJobDescription, "desc", "", "Description alias")
	jobCreateCmd.Flags().StringVar(&flagJobType, "type", "bash", "Runtime / Shebang template: bash, node, bun, python, ruby, or custom shebang")
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

	var serverJobs []*client.ServerJob
	if cfg.ServerURL != "" && cfg.RunnerToken != "" {
		c := client.New(cfg)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		serverJobs, _ = c.ListJobs(ctx)
		cancel()
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
	} else if !flagNoInteractive && ui.IsInteractive() {
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

	var syncReqs []*client.CreateJobRequest
	for _, it := range itemsToPush {
		job := it.LocalJob
		syncReqs = append(syncReqs, &client.CreateJobRequest{
			ID:             job.ID,
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
		if fpath, found := ws.FindScriptFileByID(j.ID); found {
			_ = ws.UpdateScriptID(fpath, j.ID)
		} else if fpath, found := ws.FindScriptFile(j.Slug); found {
			_ = ws.UpdateScriptID(fpath, j.ID)
		}
		fmt.Printf("  %s %-20s %s: %s  %s: %s\n",
			ui.Bullet(), ui.Cyan(j.Slug), ui.Dim("ID"), ui.Bold(j.ID), ui.Dim("cron"), ui.Yellow(j.Cron))
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
	serverJobs, err := c.ListJobs(ctx)
	cancel()
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
	} else if !flagNoInteractive && ui.IsInteractive() {
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
}

func printLocalJobLine(j workspace.LocalJob) {
	tz := displayTimezone(j.Timezone)
	desc := fmt.Sprintf("[%s: %s, %s: %s]",
		ui.Dim("cron"), ui.Yellow(j.Cron),
		ui.Dim("tz"), ui.Dim(tz),
	)
	if j.TimeoutSeconds > 0 && j.TimeoutSeconds != 30 {
		desc += fmt.Sprintf(" (%ds)", j.TimeoutSeconds)
	}
	fmt.Printf("  %s %-18s %s\n", ui.Bullet(), ui.Cyan(j.Slug), desc)
	if j.FilePath != "" {
		fmt.Printf("    %s %s\n", ui.Dim("File:"), ui.Dim(j.FilePath))
	} else if len(j.Command) > 0 {
		fmt.Printf("    %s %s\n", ui.Dim("Cmd:"), ui.Dim(strings.Join(j.Command, " ")))
	}
	fmt.Printf("    %s %s\n", ui.Dim("ID:"), ui.Dim(displayJobID(j.ID)))
}

func displayTimezone(tz string) string {
	tz = strings.TrimSpace(tz)
	if tz == "" {
		return "UTC"
	}
	return tz
}

func displayJobID(id string) string {
	id = strings.TrimSpace(id)
	if id == "" {
		return "(unset)"
	}
	return id
}
