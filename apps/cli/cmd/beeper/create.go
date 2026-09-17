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
		Long: fmt.Sprintf(`Create a monitor probe beeper.

Interactive mode guides you through selecting an app template and configuring its options.
In non-interactive mode or when flags are provided, options are read from flags.

Examples:
  # Interactive creation wizard
  %s beeper create

  # Heartbeat probe
  %s beeper create --app heartbeat --title "Production API Heartbeat" --cron "*/5 * * * *"

  # HTTP check probe
  %s beeper create --app http --title "Website Health" --cron "*/5 * * * *" --config url=https://example.com`,
			config.BinaryName(), config.BinaryName(), config.BinaryName()),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmdutil.RunWithClient(cmd, func(ctx context.Context, cfg *config.Config, c *client.Client) error {
				appSlug := strings.TrimSpace(flagApp)
				title := strings.TrimSpace(flagTitle)
				cron := strings.TrimSpace(flagCron)
				configMap := make(map[string]any)

				for _, item := range flagConfigs {
					parts := strings.SplitN(item, "=", 2)
					if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" {
						return fmt.Errorf("invalid --config %q: must be in key=value format", item)
					}
					configMap[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
				}

				tz := flagTimezone
				if tz != "" {
					if !workspace.ValidIANATimezone(tz) {
						return fmt.Errorf("invalid --timezone %q: must be a valid IANA timezone (e.g. Asia/Shanghai, UTC, America/New_York)", tz)
					}
				} else {
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
							errList := client.ExtractErrorList(err)
							ui.PrintErrorList("Invalid input", errList)
							retry, promptErr := ui.PromptConfirm("Would you like to adjust your inputs?", true)
							if promptErr != nil || !retry {
								return err
							}
							prompted, pErr := ui.PromptBeeperAdjust(params, apps, defaultChannels, errList)
							if pErr != nil {
								return pErr
							}
							params = *prompted
							continue
						}

						_ = ui.WithSpinner("Creating beeper...", func() error {
							b, err = c.CreateBeeper(ctx, req)
							return err
						})
						if err == nil {
							break
						}

						errList := client.ExtractErrorList(err)
						ui.PrintErrorList("Creation failed", errList)
						retry, promptErr := ui.PromptConfirm("Would you like to adjust your inputs and retry?", true)
						if promptErr != nil || !retry {
							return err
						}
						prompted, pErr := ui.PromptBeeperAdjust(params, apps, defaultChannels, errList)
						if pErr != nil {
							return pErr
						}
						params = *prompted
					}
				} else {
					if params.AppSlug == "" {
						return fmt.Errorf("--app slug is required (view available apps with '%s beeper apps')", config.BinaryName())
					}
					if params.Timezone == "" {
						if detected, ok := workspace.DetectTimezoneOK(); ok {
							params.Timezone = detected
						} else {
							params.Timezone = "UTC"
						}
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
					errCreate = ui.WithSpinner("Creating beeper...", func() error {
						var createErr error
						b, createErr = c.CreateBeeper(ctx, req)
						return createErr
					})
					if errCreate != nil {
						return errCreate
					}
				}

				if cmdutil.IsJSON(cmd) {
					redacted := b.Redacted(true)
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
						url.PathEscape(b.PingToken),
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
