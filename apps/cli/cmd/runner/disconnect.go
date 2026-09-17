package runner

import (
	"context"
	"fmt"
	"time"

	"beep/internal/client"
	"beep/internal/cmdutil"
	"beep/internal/config"
	"beep/internal/ui"

	"github.com/spf13/cobra"
)

// NewCmdDisconnect creates the 'runner disconnect' subcommand.
func NewCmdDisconnect() *cobra.Command {
	return &cobra.Command{
		Use:   "disconnect",
		Short: "Disconnect CLI runner (removes server-side runner and clears local token)",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig(cmd)
			if err != nil {
				return err
			}

			workspace := cmdutil.GetWorkspace(cmd)
			configPath := config.GetConfigPath(workspace)
			fc, err := config.LoadFile(configPath)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			if fc.RunnerToken == "" {
				fmt.Println(ui.Dim("No CLI runner token configured in " + configPath))
				return nil
			}

			if cfg.ServerURL != "" && cfg.RunnerToken != "" {
				c := client.New(cfg)
				ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
				defer cancel()

				_ = ui.WithSpinner("Disconnecting runner on server...", func() error {
					return c.DisconnectRunner(ctx)
				})
			}

			fc.RunnerToken = ""
			if err := config.SaveFile(configPath, fc); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Printf("%s Disconnected CLI runner from %s\n", ui.Green("✓"), ui.Dim(configPath))
			return nil
		},
	}
}
