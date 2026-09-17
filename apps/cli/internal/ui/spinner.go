package ui

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/mattn/go-isatty"
)

var (
	spinnerEnabled = true
	spinnerMu      sync.Mutex
)

// SetSpinnerEnabled sets whether the spinner TUI animation can be shown.
func SetSpinnerEnabled(v bool) {
	spinnerMu.Lock()
	defer spinnerMu.Unlock()
	spinnerEnabled = v
}

// IsSpinnerEnabled returns whether the spinner TUI animation is enabled.
func IsSpinnerEnabled() bool {
	spinnerMu.Lock()
	defer spinnerMu.Unlock()
	if !spinnerEnabled || !enabled {
		return false
	}
	return isatty.IsTerminal(os.Stderr.Fd()) && os.Getenv("TERM") != "dumb"
}

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

const minSpinnerDuration = 250 * time.Millisecond

// WithSpinner executes fn while displaying a gh-style Braille loading spinner on os.Stderr.
// If the terminal is non-interactive, disabled, or redirected, fn is executed directly.
func WithSpinner(msg string, fn func() error) error {
	if !IsSpinnerEnabled() {
		return fn()
	}

	start := time.Now()
	stopCh := make(chan struct{})
	doneCh := make(chan struct{})

	go func() {
		defer close(doneCh)
		ticker := time.NewTicker(80 * time.Millisecond)
		defer ticker.Stop()

		i := 0
		// Render initial frame immediately
		fmt.Fprintf(os.Stderr, "\r\033[K%s %s", Cyan(spinnerFrames[0]), Dim(msg))
		i++

		for {
			select {
			case <-stopCh:
				clearLine(os.Stderr)
				return
			case <-ticker.C:
				frame := spinnerFrames[i%len(spinnerFrames)]
				fmt.Fprintf(os.Stderr, "\r\033[K%s %s", Cyan(frame), Dim(msg))
				i++
			}
		}
	}()

	err := fn()

	if elapsed := time.Since(start); elapsed < minSpinnerDuration {
		time.Sleep(minSpinnerDuration - elapsed)
	}

	close(stopCh)
	<-doneCh

	return err
}

// WithSpinnerResult executes fn while displaying a loading spinner and returns its result.
func WithSpinnerResult[T any](msg string, fn func() (T, error)) (T, error) {
	if !IsSpinnerEnabled() {
		return fn()
	}

	start := time.Now()
	stopCh := make(chan struct{})
	doneCh := make(chan struct{})

	go func() {
		defer close(doneCh)
		ticker := time.NewTicker(80 * time.Millisecond)
		defer ticker.Stop()

		i := 0
		// Render initial frame immediately
		fmt.Fprintf(os.Stderr, "\r\033[K%s %s", Cyan(spinnerFrames[0]), Dim(msg))
		i++

		for {
			select {
			case <-stopCh:
				clearLine(os.Stderr)
				return
			case <-ticker.C:
				frame := spinnerFrames[i%len(spinnerFrames)]
				fmt.Fprintf(os.Stderr, "\r\033[K%s %s", Cyan(frame), Dim(msg))
				i++
			}
		}
	}()

	res, err := fn()

	if elapsed := time.Since(start); elapsed < minSpinnerDuration {
		time.Sleep(minSpinnerDuration - elapsed)
	}

	close(stopCh)
	<-doneCh

	return res, err
}

func clearLine(w io.Writer) {
	fmt.Fprint(w, "\r\033[K")
}
