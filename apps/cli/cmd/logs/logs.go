package logs

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"beep/internal/cmdutil"
	daemonlogs "beep/internal/logs"
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
		Long: `View local workspace daemon logs from prefixed daily files (runner-YYYY-MM-DD.log and channel-YYYY-MM-DD.log). Tails recent lines and follows when attached to a terminal.

--since / --until select whole calendar-day files (Nd or YYYY-MM-DD), not per-line timestamps. Hours and minutes (12h, 30m) are rejected.

Unprefixed YYYY-MM-DD.log files from older combined foreground start are not read.`,
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
			from := daemonlogs.StartOfDay(now)
			if since != "" {
				from, err = daemonlogs.ParseInstant(since, now)
				if err != nil {
					return err
				}
			}
			to := now
			if until != "" {
				to, err = daemonlogs.ParseUntil(until, now)
				if err != nil {
					return err
				}
			}

			svcs, err := daemonlogs.ParseServices(services)
			if err != nil {
				return err
			}

			days := daemonlogs.DaysCovering(from, to)
			hist, err := daemonlogs.History(cfg.Workspace, svcs, days, greps, lines)
			if err != nil {
				return err
			}

			asJSON := cmdutil.IsJSON(cmd)
			color := ui.IsEnabled() && !asJSON
			out := cmd.OutOrStdout()

			if len(hist) == 0 && !daemonlogs.ShouldFollow(follow, noFollow, isLogsTTY(), to.Before(now)) {
				fmt.Fprintf(out, "%s\n", ui.Dim(fmt.Sprintf("No daemon logs in %s", daemonlogs.LogsDir(cfg.Workspace))))
				return nil
			}

			if err := daemonlogs.WriteLines(out, hist, asJSON, color); err != nil {
				return err
			}

			if !daemonlogs.ShouldFollow(follow, noFollow, isLogsTTY(), to.Before(now)) {
				return nil
			}

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
			defer stop()
			return daemonlogs.Follow(ctx, cfg.Workspace, svcs, greps, timeNow, out, asJSON, color)
		},
	}

	cmd.Flags().IntVarP(&lines, "lines", "n", 100, "Number of matching history lines to show")
	cmd.Flags().BoolVarP(&follow, "follow", "f", false, "Follow log output (default when stdout is a terminal)")
	cmd.Flags().BoolVar(&noFollow, "no-follow", false, "Do not follow; print history and exit")
	cmd.Flags().StringVar(&since, "since", "", "Show logs since a relative number of days (e.g. 2d) or date (YYYY-MM-DD); default today")
	cmd.Flags().StringVar(&until, "until", "", "Show logs until a relative number of days or date (YYYY-MM-DD); default now")
	cmd.Flags().StringSliceVar(&services, "service", nil, "Limit to runner and/or channel (comma-separated or repeated)")
	cmd.Flags().StringArrayVarP(&greps, "grep", "e", nil, "Case-insensitive substring filter (repeat for OR)")

	return cmd
}

var timeNow = func() time.Time { return time.Now() }

func isLogsTTY() bool {
	return isatty.IsTerminal(os.Stdout.Fd())
}
