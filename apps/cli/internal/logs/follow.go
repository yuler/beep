package logs

import (
	"bufio"
	"context"
	"io"
	"os"
	"strings"
	"time"

	"beep/internal/daemon"
)

const followPoll = 200 * time.Millisecond

type cursor struct {
	path   string
	source string
	offset int64
	seen   bool
}

// Follow watches today's prefixed log files and writes new matching lines until ctx is done.
// Existing file content is skipped; call History first for the tail.
func Follow(ctx context.Context, workspace string, services, patterns []string, now func() time.Time, w io.Writer, asJSON, color bool) error {
	if now == nil {
		now = time.Now
	}
	cursors := map[string]*cursor{}
	initialTick := true

	tick := func() error {
		day := now().Format("2006-01-02")
		for _, service := range services {
			path := daemon.DailyLogPath(workspace, service, day)
			c, ok := cursors[path]
			if !ok {
				c = &cursor{path: path, source: service}
				cursors[path] = c
			}
			if !c.seen {
				if initialTick {
					if info, err := os.Stat(path); err == nil {
						c.offset = info.Size()
					}
					c.seen = true
					continue
				}
				c.seen = true
			}
			if err := emitNew(c, patterns, w, asJSON, color); err != nil {
				return err
			}
		}
		return nil
	}

	if err := tick(); err != nil {
		return err
	}
	initialTick = false
	timer := time.NewTicker(followPoll)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-timer.C:
			if err := tick(); err != nil {
				return err
			}
		}
	}
}

func emitNew(c *cursor, patterns []string, w io.Writer, asJSON, color bool) error {
	f, err := os.Open(c.path)
	if err != nil {
		if os.IsNotExist(err) {
			c.offset = 0
			return nil
		}
		return err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return err
	}
	if info.Size() < c.offset {
		c.offset = 0
	}
	if _, err := f.Seek(c.offset, io.SeekStart); err != nil {
		return err
	}
	reader := bufio.NewReader(f)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF && (len(line) == 0 || !strings.HasSuffix(line, "\n")) {
			return nil
		}
		if len(line) > 0 {
			c.offset += int64(len(line))
			text := strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
			if MatchGrep(text, patterns) {
				if werr := WriteLines(w, []Line{{Source: c.source, Text: text}}, asJSON, color); werr != nil {
					return werr
				}
			}
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}
