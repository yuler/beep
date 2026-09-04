package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"beep-runner/internal/daemon"
	"beep-runner/internal/ui"
	"beep-runner/internal/workspace"

	"github.com/spf13/cobra"
)

var (
	flagConcurrency  int
	flagPollInterval time.Duration
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Start the runner daemon to poll and execute scheduled tasks",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDaemon(cmd, args)
	},
}

func init() {
	runCmd.Flags().IntVarP(&flagConcurrency, "concurrency", "c", 0, "Max concurrent jobs (default 5)")
	runCmd.Flags().DurationVarP(&flagPollInterval, "poll-interval", "i", 0, "Poll interval (default 3s)")
}

func runDaemon(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	if flagConcurrency > 0 {
		cfg.Concurrency = flagConcurrency
	}
	if flagPollInterval > 0 {
		cfg.PollInterval = flagPollInterval
	}
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	ws, err := workspace.Open(cfg.Workspace)
	if err != nil {
		return fmt.Errorf("workspace error: %w", err)
	}

	d := daemon.New(cfg, ws)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println(ui.Dim("[beep-runner] Received termination signal..."))
		cancel()
	}()

	return d.Start(ctx)
}
