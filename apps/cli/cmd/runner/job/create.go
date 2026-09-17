package job

import (
	"context"
	"fmt"
	"strings"
	"time"

	"beep/internal/client"
	"beep/internal/cmdutil"
	"beep/internal/config"
	"beep/internal/schedule"
	"beep/internal/ui"
	"beep/internal/workspace"

	"github.com/spf13/cobra"
)

// NewCmdCreate creates the 'runner job create' subcommand.
func NewCmdCreate() *cobra.Command {
	var (
		flagJobName        string
		flagJobCron        string
		flagJobTimezone    string
		flagJobDescription string
		flagJobType        string
		flagJobTimeout     int
		flagJobNoSync      bool
	)

	cmd := &cobra.Command{
		Use:   "create [slug]",
		Short: "Create a local job script scaffold (interactive if slug is omitted)",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig(cmd)
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

			if cmdutil.IsInteractive(cmd) {
				prompted, err := ui.PromptJobCreate(createParams)
				if err != nil {
					return err
				}
				createParams = *prompted
			} else if createParams.Slug == "" {
				return fmt.Errorf("job slug is required (e.g. beep runner job create intranet-gateway)")
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
				fmt.Printf("  %s %s\n", ui.Dim("Tip: Configure once with"), ui.Cyan(fmt.Sprintf("%s runner config set --server <url> --token <token>", config.BinaryName())))
				return nil
			}

			c := client.New(cfg)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			var serverJob *client.ServerJob
			err = ui.WithSpinner("Syncing job to server...", func() error {
				var createErr error
				serverJob, createErr = c.CreateJob(ctx, &client.CreateJobRequest{
					ID:             jobID,
					Slug:           createParams.Slug,
					Name:           createParams.Name,
					Cron:           createParams.Cron,
					Timezone:       createParams.Timezone,
					TimeoutSeconds: createParams.TimeoutSeconds,
					Description:    createParams.Description,
				})
				return createErr
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
			fmt.Printf("  2. Start runner daemon: %s\n", ui.Green(fmt.Sprintf("%s runner up", config.BinaryName())))
			return nil
		},
	}

	cmd.Flags().StringVar(&flagJobName, "name", "", "Display name of the job")
	cmd.Flags().StringVar(&flagJobCron, "cron", "*/5 * * * *", "Schedule: classic cron or Fugit semantic (e.g. every 5 minutes)")
	cmd.Flags().StringVar(&flagJobTimezone, "timezone", "", "Timezone (e.g. Asia/Shanghai, UTC)")
	cmd.Flags().StringVar(&flagJobTimezone, "tz", "", "Timezone alias")
	cmd.Flags().StringVar(&flagJobDescription, "description", "", "Description of the job")
	cmd.Flags().StringVar(&flagJobDescription, "desc", "", "Description alias")
	cmd.Flags().StringVar(&flagJobType, "type", "bash", "Runtime / Shebang template: bash, node, bun, python, ruby, or custom shebang")
	cmd.Flags().IntVar(&flagJobTimeout, "timeout", 30, "Timeout in seconds")
	cmd.Flags().BoolVar(&flagJobNoSync, "no-sync", false, "Do not sync to server")

	return cmd
}
