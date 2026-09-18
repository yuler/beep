package logs

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"beep/internal/cmdutil"
	"beep/internal/logview"
	"beep/internal/ui"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

// NewCmdLogs creates the top-level `beep logs` command.
func NewCmdLogs() *cobra.Command {
	var (
		lines    int
		follow   bool
		noFollow bool
		since    string
		until    string
		services []string
		greps    []string
	)

	cmd := &cobra.Command{
		Use:     "logs",
		Aliases: []string{"log"},
		Short:   "View local workspace daemon logs",
		Long:    "View local workspace daemon logs (runner and channel daily files). Tails recent lines and follows when attached to a terminal.",
		Example: `  beep logs
  beep logs --service runner
  beep logs --since 2d -e error
  beep logs --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig(cmd)
			if err != nil {
				return err
			}

			now := timeNow()
			from := logview.StartOfDay(now)
			if since != "" {
				from, err = logview.ParseInstant(since, now)
				if err != nil {
					return err
				}
			}
			to := now
			if until != "" {
				to, err = logview.ParseUntil(until, now)
				if err != nil {
					return err
				}
			}

			svcs, err := logview.ParseServices(services)
			if err != nil {
				return err
			}

			days := logview.DaysCovering(from, to)
			hist, err := logview.History(cfg.Workspace, svcs, days, greps, lines)
			if err != nil {
				return err
			}

			asJSON := cmdutil.IsJSON(cmd)
			color := ui.IsEnabled() && !asJSON
			out := cmd.OutOrStdout()

			if len(hist) == 0 && !logview.ShouldFollow(follow, noFollow, isLogsTTY(), to.Before(now)) {
				fmt.Fprintf(out, "%s\n", ui.Dim(fmt.Sprintf("No daemon logs in %s", logview.LogsDir(cfg.Workspace))))
				return nil
			}

			if err := logview.WriteLines(out, hist, asJSON, color); err != nil {
				return err
			}

			if !logview.ShouldFollow(follow, noFollow, isLogsTTY(), to.Before(now)) {
				return nil
			}

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
			defer stop()
			return logview.Follow(ctx, cfg.Workspace, svcs, greps, timeNow, out, asJSON, color)
		},
	}

	cmd.Flags().IntVarP(&lines, "lines", "n", 100, "Number of matching history lines to show")
	cmd.Flags().BoolVarP(&follow, "follow", "f", false, "Follow log output (default when stdout is a terminal)")
	cmd.Flags().BoolVar(&noFollow, "no-follow", false, "Do not follow; print history and exit")
	cmd.Flags().StringVar(&since, "since", "", "Show logs since a relative time (2d, 12h, 30m) or date (YYYY-MM-DD); default today")
	cmd.Flags().StringVar(&until, "until", "", "Show logs until a relative time or date (YYYY-MM-DD); default now")
	cmd.Flags().StringSliceVar(&services, "service", nil, "Limit to runner and/or channel (comma-separated or repeated)")
	cmd.Flags().StringArrayVarP(&greps, "grep", "e", nil, "Case-insensitive substring filter (repeat for OR)")

	return cmd
}

var timeNow = func() time.Time { return time.Now() }

func isLogsTTY() bool {
	return isatty.IsTerminal(os.Stdout.Fd())
}
