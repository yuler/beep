package config

import (
	"fmt"
	"strconv"
	"strings"

	"beep/internal/cmdutil"
	"beep/internal/config"
	"beep/internal/ui"

	"github.com/spf13/cobra"
)

// NewCmdSet creates the 'config set' subcommand.
func NewCmdSet() *cobra.Command {
	var flagSetConfig config.FileConfig

	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set configuration parameters (server, token, workspace, concurrency, poll-interval)",
		RunE: func(cmd *cobra.Command, args []string) error {
			workspace := cmdutil.GetWorkspace(cmd)
			configPath := config.GetConfigPath(workspace)
			fc, err := config.LoadFile(configPath)
			if err != nil {
				fc = &config.FileConfig{}
			}

			updated := false

			if flagSetConfig.ServerURL != "" {
				fc.ServerURL = strings.TrimRight(flagSetConfig.ServerURL, "/")
				updated = true
			}
			accountVal := flagSetConfig.AccountSlug
			if accountVal == "" && cmd.Root() != nil && cmd.Root().PersistentFlags().Lookup("account") != nil {
				accountVal, _ = cmd.Root().PersistentFlags().GetString("account")
			}
			if accountVal != "" {
				fc.AccountSlug = strings.TrimSpace(accountVal)
				updated = true
			}
			if flagSetConfig.RunnerToken != "" {
				fc.RunnerToken = flagSetConfig.RunnerToken
				updated = true
			}
			if flagSetConfig.Workspace != "" {
				fc.Workspace = flagSetConfig.Workspace
				updated = true
			}
			if flagSetConfig.Concurrency > 0 {
				fc.Concurrency = flagSetConfig.Concurrency
				updated = true
			}
			if flagSetConfig.PollInterval != "" {
				fc.PollInterval = flagSetConfig.PollInterval
				updated = true
			}

			// Handle positional arguments: e.g. beep config set server https://...
			for i := 0; i < len(args); i++ {
				key := strings.ToLower(args[i])
				if i+1 < len(args) {
					val := args[i+1]
					switch key {
					case "server", "server_url", "server-url", "url":
						fc.ServerURL = strings.TrimRight(val, "/")
						updated = true
						i++
					case "account", "account_slug", "account-slug", "slug":
						fc.AccountSlug = strings.TrimSpace(val)
						updated = true
						i++
					case "access_token", "access-token", "user_token", "user-token":
						fc.AccessToken = val
						updated = true
						i++
					case "token", "runner_token", "runner-token", "auth":
						fc.RunnerToken = val
						updated = true
						i++
					case "channel_token", "channel-token", "channel", "cli_token", "device_token":
						fc.ChannelToken = val
						fc.CliToken = ""
						fc.DeviceToken = ""
						updated = true
						i++
					case "workspace", "dir", "workdir":
						fc.Workspace = val
						updated = true
						i++
					case "concurrency":
						if n, err := strconv.Atoi(val); err == nil && n > 0 {
							fc.Concurrency = n
							updated = true
							i++
						}
					case "poll_interval", "poll-interval", "interval":
						fc.PollInterval = val
						updated = true
						i++
					}
				}
			}

			// If no parameters provided and interactive, launch prompt wizard
			if !updated {
				if cmdutil.IsInteractive(cmd) {
					if err := ui.PromptConfigSetWizard(fc); err != nil {
						return err
					}
					updated = true
				} else {
					fmt.Println(ui.Warn("No configuration options provided."))
					fmt.Println()
					bin := config.BinaryName()
					fmt.Printf("  %s\n", ui.Cyan(fmt.Sprintf("%s config set --server <url> --token <token>", bin)))
					fmt.Printf("  %s\n", ui.Cyan(fmt.Sprintf("%s config set server <url>", bin)))
					fmt.Printf("  %s\n", ui.Cyan(fmt.Sprintf("%s config set token <token>", bin)))
					return nil
				}
			}

			if err := config.SaveFile(configPath, fc); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Println(ui.Success("Saved configuration to %s", ui.Bold(configPath)))
			fmt.Println()
			return RunConfigShow(cmd, nil, false)
		},
	}

	cmd.Flags().StringVar(&flagSetConfig.ServerURL, "server", "", "Beep server URL")
	cmd.Flags().StringVar(&flagSetConfig.RunnerToken, "token", "", "Runner token")
	cmd.Flags().StringVar(&flagSetConfig.Workspace, "workspace", "", "Workspace directory")
	cmd.Flags().IntVar(&flagSetConfig.Concurrency, "concurrency", 0, "Max concurrency")
	cmd.Flags().StringVar(&flagSetConfig.PollInterval, "poll-interval", "", "Poll interval (e.g. 3s)")

	return cmd
}
