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

// NewCmdList creates the 'config list' subcommand.
func NewCmdList() *cobra.Command {
	var flagShowToken bool

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List configuration parameters in table format",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig(cmd)
			if err != nil {
				return err
			}

			type configEntry struct {
				Key   string `json:"key"`
				Value string `json:"value"`
			}

			tokenVal := config.MaskToken(cfg.RunnerToken)
			if flagShowToken && cfg.RunnerToken != "" {
				tokenVal = cfg.RunnerToken
			}
			accessVal := config.MaskToken(cfg.AccessToken)
			if flagShowToken && cfg.AccessToken != "" {
				accessVal = cfg.AccessToken
			}
			chanVal := config.MaskToken(cfg.ChannelToken)
			if flagShowToken && cfg.ChannelToken != "" {
				chanVal = cfg.ChannelToken
			}

			entries := []configEntry{
				{"server", cfg.ServerURL},
				{"account", cfg.AccountSlug},
				{"token", tokenVal},
				{"access_token", accessVal},
				{"channel_token", chanVal},
				{"workspace", cfg.Workspace},
				{"concurrency", strconv.Itoa(cfg.Concurrency)},
				{"poll_interval", cfg.PollInterval.String()},
				{"config_file", cfg.ConfigFile},
			}

			if cmdutil.IsJSON(cmd) {
				data, err := json.MarshalIndent(entries, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(data))
				return nil
			}

			tbl := ui.NewTable("KEY", "VALUE")
			tbl.SetIndent("  ")
			for _, e := range entries {
				val := e.Value
				if val == "" {
					val = ui.Dim("-")
				}
				tbl.AddRow(ui.Bold(e.Key), val)
			}

			fmt.Println(ui.Bold(ui.Cyan("Configuration:")))
			fmt.Println()
			if err := tbl.Print(); err != nil {
				return err
			}
			fmt.Println()
			return nil
		},
	}

	cmd.Flags().BoolVar(&flagShowToken, "show-token", false, "Display unmasked runner and channel tokens")
	return cmd
}
