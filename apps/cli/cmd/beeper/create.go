package beeper

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"beep/internal/client"
	"beep/internal/cmdutil"
	"beep/internal/config"
	"beep/internal/ui"
	"beep/internal/workspace"

	"github.com/spf13/cobra"
)

// NewCmdCreate creates the 'beeper create' subcommand.
func NewCmdCreate() *cobra.Command {
	var (
		flagApp      string
		flagTitle    string
		flagBody     string
		flagCron     string
		flagTimezone string
		flagChannels string
		flagConfigs  []string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a monitor probe beeper",
		Long: `Create a monitor probe beeper.

Interactive mode guides you through selecting an app template and configuring its options.
In non-interactive mode or when flags are provided, options are read from flags.

Examples:
  # Interactive creation wizard
  beep beeper create

  # Heartbeat probe
  beep beeper create --app heartbeat --title "Production API Heartbeat" --cron "*/5 * * * *"

  # HTTP check probe
  beep beeper create --app http --title "Website Health" --cron "*/5 * * * *" --config url=https://example.com`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmdutil.RunWithClient(cmd, func(ctx context.Context, cfg *config.Config, c *client.Client) error {
				appSlug := strings.TrimSpace(flagApp)
				title := strings.TrimSpace(flagTitle)
				cron := strings.TrimSpace(flagCron)
				configMap := make(map[string]any)

				for _, item := range flagConfigs {
					parts := strings.SplitN(item, "=", 2)
					if len(parts) == 2 {
						configMap[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
					}
				}

				tz := flagTimezone
				if tz == "" {
					if detected, ok := workspace.DetectTimezoneOK(); ok {
						tz = detected
					} else {
						tz = "UTC"
					}
				}

				params := client.CreateBeeperParams{
					AppSlug:  appSlug,
					Title:    title,
					Body:     strings.TrimSpace(flagBody),
					Cron:     cron,
					Timezone: tz,
					Config:   configMap,
					Channels: flagChannels,
				}

				var b *client.Beeper
				if cmdutil.IsInteractive(cmd) {
					apps, err := c.ListBeeperApps(ctx)
					if err != nil {
						return fmt.Errorf("failed to list beeper apps: %w", err)
					}

					// Default channel selection comes from account settings.
					defaultChannels := client.DefaultNotificationChannels
					if s, err := c.GetSettings(ctx); err == nil && s != nil {
						defaultChannels = client.SanitizeChannelDefaults(s.NotificationChannels)
					}

					if params.AppSlug == "" || params.Title == "" {
						prompted, err := ui.PromptBeeperCreate(params, apps, defaultChannels)
						if err != nil {
							return err
						}
						params = *prompted
					}

					for {
						req, err := params.ToRequest()
						if err != nil {
							ui.PrintErrorList("Invalid input", client.ExtractErrorList(err))
							retry, promptErr := ui.PromptConfirm("Would you like to adjust your inputs?", true)
							if promptErr != nil || !retry {
								return err
							}
							prompted, pErr := ui.PromptBeeperAdjust(params, apps, defaultChannels)
							if pErr != nil {
								return pErr
							}
							params = *prompted
							continue
						}

						b, err = c.CreateBeeper(ctx, req)
						if err == nil {
							break
						}

						ui.PrintErrorList("Creation failed", client.ExtractErrorList(err))
						retry, promptErr := ui.PromptConfirm("Would you like to adjust your inputs and retry?", true)
						if promptErr != nil || !retry {
							return err
						}
						prompted, pErr := ui.PromptBeeperAdjust(params, apps, defaultChannels)
						if pErr != nil {
							return pErr
						}
						params = *prompted
					}
				} else {
					if params.AppSlug == "" {
						return errors.New("--app slug is required (view available apps with 'beep beeper apps')")
					}
					if params.Title == "" {
						selectedApp, err := c.GetBeeperApp(ctx, params.AppSlug)
						if err != nil {
							return fmt.Errorf("failed to fetch beeper app %s: %w", params.AppSlug, err)
						}
						params.Title = selectedApp.Name
						if params.Cron == "" {
							params.Cron = selectedApp.DefaultCron
						}
					}

					req, err := params.ToRequest()
					if err != nil {
						return err
					}

					var errCreate error
					b, errCreate = c.CreateBeeper(ctx, req)
					if errCreate != nil {
						return errCreate
					}
				}

				if cmdutil.IsJSON(cmd) {
					redacted := b.Redacted(false)
					data, err := json.MarshalIndent(redacted, "", "  ")
					if err != nil {
						return err
					}
					fmt.Println(string(data))
					return nil
				}

				fmt.Println(ui.Success("Created beeper %s (%s)", ui.Bold(b.Title), ui.Dim(b.ID)))
				fmt.Printf("  %s %s\n", ui.Dim("App:"), params.AppSlug)
				if b.Cron != "" {
					fmt.Printf("  %s %s\n", ui.Dim("Cron:"), b.Cron)
				}
				if b.PingToken != "" {
					fmt.Printf("  %s %s\n", ui.Dim("Ping Token:"), ui.Yellow(b.PingToken))
					fmt.Printf("  %s %s/api/v1/beeper_apps/heartbeat/pings/%s\n",
						ui.Dim("Ping URL:"),
						cfg.ServerURL,
						b.PingToken,
					)
				}
				return nil
			})
		},
	}

	cmd.Flags().StringVar(&flagApp, "app", "", "Beeper app slug (e.g. 'heartbeat-ping')")
	cmd.Flags().StringVar(&flagTitle, "title", "", "Beeper probe title")
	cmd.Flags().StringVarP(&flagBody, "body", "b", "", "Description / body for this monitor")
	cmd.Flags().StringVarP(&flagCron, "cron", "c", "", "Recurring cron schedule (e.g. '*/5 * * * *')")
	cmd.Flags().StringVarP(&flagTimezone, "timezone", "z", "", "Timezone (defaults to local timezone)")
	cmd.Flags().StringVar(&flagChannels, "channels", "", "Comma-separated notification channel names or IDs")
	cmd.Flags().StringSliceVar(&flagConfigs, "config", nil, "App config key=value pair (can be specified multiple times)")

	return cmd
}
