package cmd

import (
	"context"
	"fmt"
	"time"

	"beep-runner/internal/client"
	"beep-runner/internal/ui"

	"github.com/spf13/cobra"
)

var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Test connectivity and authentication with Beep Core",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}
		if err := cfg.Validate(); err != nil {
			return err
		}

		c := client.New(cfg)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		res, err := c.Ping(ctx)
		if err != nil {
			return fmt.Errorf("ping failed: %w", err)
		}

		fmt.Println(ui.Success("Ping successful!"))
		fmt.Println(ui.KeyValue("Runner ID", ui.Bold(res.RunnerID)))
		fmt.Println(ui.KeyValue("Runner Name", ui.Bold(res.RunnerName)))
		fmt.Println(ui.KeyValue("Server Time", ui.Dim(res.ServerTime)))
		return nil
	},
}
