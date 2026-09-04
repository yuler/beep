package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"beep-runner/internal/config"
	"beep-runner/internal/ui"

	"github.com/spf13/cobra"
)

var (
	flagShowToken bool
	flagSetConfig config.FileConfig
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage local runner configuration (~/.beep-runner/config.json)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigShow(cmd, args)
	},
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Display current configuration parameters",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigShow(cmd, args)
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set configuration parameters (server, token, workspace, concurrency, poll-interval)",
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath := config.GetConfigPath(flagWorkspace)
		fc, err := config.LoadFile(configPath)
		if err != nil {
			fc = &config.FileConfig{}
		}

		updated := false

		if flagSetConfig.ServerURL != "" {
			fc.ServerURL = strings.TrimRight(flagSetConfig.ServerURL, "/")
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

		// Handle positional arguments: e.g. beep-runner config set server https://...
		for i := 0; i < len(args); i++ {
			key := strings.ToLower(args[i])
			if i+1 < len(args) {
				val := args[i+1]
				switch key {
				case "server", "server_url", "server-url", "url":
					fc.ServerURL = strings.TrimRight(val, "/")
					updated = true
					i++
				case "token", "runner_token", "runner-token", "auth":
					fc.RunnerToken = val
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

		// If no parameters provided and interactive, launch Huh prompt wizard!
		if !updated {
			if !flagNoInteractive && ui.IsInteractive() {
				if err := ui.PromptConfigSetWizard(fc); err != nil {
					return err
				}
				updated = true
			} else {
				fmt.Println(ui.Warn("No configuration options provided."))
				fmt.Println()
				fmt.Println(ui.Section("Usage:"))
				fmt.Printf("  %s\n", ui.Cyan("beep-runner config set --server <url> --token <token>"))
				fmt.Printf("  %s\n", ui.Cyan("beep-runner config set server <url>"))
				fmt.Printf("  %s\n", ui.Cyan("beep-runner config set token <token>"))
				return nil
			}
		}

		if err := config.SaveFile(configPath, fc); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Println(ui.Success("Saved configuration to %s", ui.Bold(configPath)))
		fmt.Println()
		return runConfigShow(cmd, nil)
	},
}

var configUnsetCmd = &cobra.Command{
	Use:   "unset [key]",
	Short: "Remove a configuration parameter",
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath := config.GetConfigPath(flagWorkspace)
		fc, err := config.LoadFile(configPath)
		if err != nil {
			return fmt.Errorf("failed to load config file: %w", err)
		}

		var key string
		if len(args) > 0 {
			key = strings.ToLower(args[0])
		} else if !flagNoInteractive && ui.IsInteractive() {
			selectedKey, err := ui.PromptConfigUnsetSelect(fc)
			if err != nil {
				return err
			}
			key = selectedKey
		} else {
			fmt.Println(ui.Section("Usage:"))
			fmt.Printf("  %s\n", ui.Cyan("beep-runner config unset <key>"))
			fmt.Printf("  %s: server, token, workspace, concurrency, poll-interval\n", ui.Dim("Keys"))
			return nil
		}

		switch key {
		case "server", "server_url", "server-url":
			fc.ServerURL = ""
		case "token", "runner_token", "runner-token":
			fc.RunnerToken = ""
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
		return runConfigShow(cmd, nil)
	},
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Print path to the configuration file",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(config.GetConfigPath(flagWorkspace))
	},
}

func init() {
	configShowCmd.Flags().BoolVar(&flagShowToken, "show-token", false, "Display unmasked runner token")
	configCmd.Flags().BoolVar(&flagShowToken, "show-token", false, "Display unmasked runner token")

	configSetCmd.Flags().StringVar(&flagSetConfig.ServerURL, "server", "", "Beep server URL")
	configSetCmd.Flags().StringVar(&flagSetConfig.RunnerToken, "token", "", "Runner token")
	configSetCmd.Flags().StringVar(&flagSetConfig.Workspace, "workspace", "", "Workspace directory")
	configSetCmd.Flags().IntVar(&flagSetConfig.Concurrency, "concurrency", 0, "Max concurrency")
	configSetCmd.Flags().StringVar(&flagSetConfig.PollInterval, "poll-interval", "", "Poll interval (e.g. 3s)")

	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configUnsetCmd)
	configCmd.AddCommand(configPathCmd)
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	tokenStr := config.MaskToken(cfg.RunnerToken)
	if flagShowToken && cfg.RunnerToken != "" {
		tokenStr = cfg.RunnerToken
	}

	fmt.Println(ui.Bold(ui.Cyan("Beep Runner Configuration:")))
	fmt.Println(ui.KeyValue("Config File", ui.Dim(cfg.ConfigFile)))
	fmt.Println(ui.KeyValue("Server URL", ui.Bold(cfg.ServerURL)))
	fmt.Println(ui.KeyValue("Runner Token", ui.Yellow(tokenStr)))
	fmt.Println(ui.KeyValue("Workspace", ui.Dim(cfg.Workspace)))
	fmt.Println(ui.KeyValue("Concurrency", ui.Bold(strconv.Itoa(cfg.Concurrency))))
	fmt.Println(ui.KeyValue("Poll Interval", ui.Bold(cfg.PollInterval.String())))
	fmt.Println(ui.KeyValue("Hostname", ui.Dim(cfg.Hostname)))
	return nil
}
