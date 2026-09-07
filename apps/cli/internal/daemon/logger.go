package daemon

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"
)

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

// DailyLogWriter is a thread-safe io.WriteCloser that writes to a log file
// named with the current date (YYYY-MM-DD) and rotates automatically when the date changes.
type DailyLogWriter struct {
	mu          sync.Mutex
	dir         string
	prefix      string
	currentDay  string
	currentFile *os.File
}

// NewDailyLogWriter creates a new DailyLogWriter for the given directory and filename prefix.
func NewDailyLogWriter(dir, prefix string) (*DailyLogWriter, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create log directory %s: %w", dir, err)
	}
	w := &DailyLogWriter{
		dir:    dir,
		prefix: prefix,
	}
	if err := w.rotateIfNeeded(time.Now()); err != nil {
		return nil, err
	}
	return w, nil
}

// CurrentLogFilePath returns the absolute path to the current log file.
func (w *DailyLogWriter) CurrentLogFilePath() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.currentFile != nil {
		return w.currentFile.Name()
	}
	day := time.Now().Format("2006-01-02")
	return filepath.Join(w.dir, fmt.Sprintf("%s-%s.log", w.prefix, day))
}

func (w *DailyLogWriter) rotateIfNeeded(now time.Time) error {
	day := now.Format("2006-01-02")
	if w.currentDay == day && w.currentFile != nil {
		return nil
	}
	if w.currentFile != nil {
		_ = w.currentFile.Close()
		w.currentFile = nil
	}
	filename := filepath.Join(w.dir, fmt.Sprintf("%s-%s.log", w.prefix, day))
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("failed to open log file %s: %w", filename, err)
	}
	w.currentDay = day
	w.currentFile = f
	return nil
}

// Write writes bytes to the active daily log file, stripping ANSI color escape codes.
func (w *DailyLogWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	now := time.Now()
	if err := w.rotateIfNeeded(now); err != nil {
		return 0, err
	}

	clean := ansiRegex.ReplaceAll(p, nil)
	_, err = w.currentFile.Write(clean)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

// Close closes the currently open log file.
func (w *DailyLogWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.currentFile != nil {
		err := w.currentFile.Close()
		w.currentFile = nil
		return err
	}
	return nil
}

// SetupLogger sets up the global log package to write to the daily log writer,
// and optionally also to os.Stdout if toStdout is true.
func SetupLogger(workspaceDir string, toStdout bool) (*DailyLogWriter, string, error) {
	logsDir := filepath.Join(workspaceDir, "logs")
	writer, err := NewDailyLogWriter(logsDir, "beep-runner")
	if err != nil {
		return nil, "", err
	}

	if toStdout {
		log.SetOutput(io.MultiWriter(os.Stdout, writer))
	} else {
		log.SetOutput(writer)
	}

	log.SetFlags(log.Ldate | log.Ltime)
	return writer, writer.CurrentLogFilePath(), nil
}
