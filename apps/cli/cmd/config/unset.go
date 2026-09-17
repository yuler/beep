package config

import (
	"fmt"
	"strings"

	"beep/internal/cmdutil"
	"beep/internal/config"
	"beep/internal/ui"

	"github.com/spf13/cobra"
)

// NewCmdUnset creates the 'config unset' subcommand.
func NewCmdUnset() *cobra.Command {
	return &cobra.Command{
		Use:   "unset [key]",
		Short: "Remove a configuration parameter",
		RunE: func(cmd *cobra.Command, args []string) error {
			workspace := cmdutil.GetWorkspace(cmd)
			configPath := config.GetConfigPath(workspace)
			fc, err := config.LoadFile(configPath)
			if err != nil {
				return fmt.Errorf("failed to load config file: %w", err)
			}

			var key string
			if len(args) > 0 {
				key = strings.ToLower(args[0])
			} else if cmdutil.IsInteractive(cmd) {
				selectedKey, err := ui.PromptConfigUnsetSelect(fc)
				if err != nil {
					return err
				}
				key = selectedKey
			} else {
				fmt.Println(ui.Section("Usage:"))
				fmt.Printf("  %s\n", ui.Cyan("beep config unset <key>"))
				fmt.Printf("  %s: server, account, token, workspace, concurrency, poll-interval\n", ui.Dim("Keys"))
				return nil
			}

			switch key {
			case "server", "server_url", "server-url":
				fc.ServerURL = ""
			case "account", "account_slug", "account-slug", "slug":
				fc.AccountSlug = ""
			case "access_token", "access-token", "user_token", "user-token":
				fc.AccessToken = ""
				fc.UserEmail = ""
				fc.UserName = ""
			case "token", "runner_token", "runner-token":
				fc.RunnerToken = ""
			case "channel_token", "channel-token", "channel", "cli_token", "device_token":
				fc.ChannelToken = ""
				fc.CliToken = ""
				fc.DeviceToken = ""
			case "workspace", "dir":
				fc.Workspace = ""
			case "concurrency":
				fc.Concurrency = 0
			case "poll_interval", "poll-interval":
				fc.PollInterval = ""
			default:
				return fmt.Errorf("unknown config key %q", key)
			}

			if err := config.SaveFile(configPath, fc); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Println(ui.Success("Unset %s in %s", ui.Bold(key), ui.Dim(configPath)))
			fmt.Println()
			return RunConfigShow(cmd, nil, false)
		},
	}
}
