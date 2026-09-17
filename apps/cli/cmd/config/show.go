package config

import (
	"encoding/json"
	"fmt"
	"strconv"

	"beep/internal/cmdutil"
	"beep/internal/config"
	"beep/internal/ui"

	"github.com/spf13/cobra"
)

// NewCmdShow creates the 'config show' subcommand.
func NewCmdShow() *cobra.Command {
	var flagShowToken bool

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Display current configuration parameters",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunConfigShow(cmd, args, flagShowToken)
		},
	}

	cmd.Flags().BoolVar(&flagShowToken, "show-token", false, "Display unmasked runner and channel tokens")
	return cmd
}

// RunConfigShow outputs current configuration parameters.
func RunConfigShow(cmd *cobra.Command, args []string, showToken bool) error {
	cfg, err := cmdutil.LoadConfig(cmd)
	if err != nil {
		return err
	}

	if cmdutil.IsJSON(cmd) {
		redactedCfg := *cfg
		if !showToken {
			if redactedCfg.RunnerToken != "" {
				redactedCfg.RunnerToken = config.MaskToken(redactedCfg.RunnerToken)
			}
			if redactedCfg.AccessToken != "" {
				redactedCfg.AccessToken = config.MaskToken(redactedCfg.AccessToken)
			}
			if redactedCfg.ChannelToken != "" {
				redactedCfg.ChannelToken = config.MaskToken(redactedCfg.ChannelToken)
			}
		}
		data, err := json.MarshalIndent(redactedCfg, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	tokenStr := config.MaskToken(cfg.RunnerToken)
	if showToken && cfg.RunnerToken != "" {
		tokenStr = cfg.RunnerToken
	}

	fmt.Println(ui.Bold(ui.Cyan("Beep Runner Configuration:")))
	fmt.Println(ui.KeyValue("Config File", ui.Dim(cfg.ConfigFile)))
	fmt.Println(ui.KeyValue("Server URL", ui.Bold(cfg.ServerURL)))
	if cfg.AccountSlug != "" {
		fmt.Println(ui.KeyValue("Account Slug", ui.Bold(cfg.AccountSlug)))
	}
	if cfg.AccessToken != "" {
		accessTokenStr := config.MaskToken(cfg.AccessToken)
		if showToken {
			accessTokenStr = cfg.AccessToken
		}
		displayName := cfg.UserEmail
		if cfg.UserName != "" {
			displayName = fmt.Sprintf("%s (%s)", cfg.UserName, cfg.UserEmail)
		}
		if displayName != "" {
			fmt.Println(ui.KeyValue("Logged In As", ui.Green(displayName)))
		}
		fmt.Println(ui.KeyValue("Access Token", ui.Yellow(accessTokenStr)))
	}
	fmt.Println(ui.KeyValue("Runner Token", ui.Yellow(tokenStr)))
	if cfg.ChannelToken != "" {
		channelTokenStr := config.MaskToken(cfg.ChannelToken)
		if showToken {
			channelTokenStr = cfg.ChannelToken
		}
		fmt.Println(ui.KeyValue("Channel Token", ui.Yellow(channelTokenStr)))
	}
	fmt.Println(ui.KeyValue("Workspace", ui.Dim(cfg.Workspace)))
	fmt.Println(ui.KeyValue("Concurrency", ui.Bold(strconv.Itoa(cfg.Concurrency))))
	fmt.Println(ui.KeyValue("Poll Interval", ui.Bold(cfg.PollInterval.String())))
	fmt.Println(ui.KeyValue("Hostname", ui.Dim(cfg.Hostname)))
	return nil
}
