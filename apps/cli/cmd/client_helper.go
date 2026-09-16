package cmd

import (
	"context"
	"os"
	"os/signal"

	"beep/internal/client"
	"beep/internal/config"
)

// runWithClient encapsulates configuration loading, signal cancellation,
// login verification, and API client initialization for subcommands.
func runWithClient(fn func(ctx context.Context, cfg *config.Config, c *client.Client) error) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if _, err := ensureLoggedIn(ctx, cfg); err != nil {
		return err
	}

	c := client.New(cfg)
	return fn(ctx, cfg, c)
}
