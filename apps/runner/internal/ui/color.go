package ui

import (
	"fmt"
	"os"
	"strings"
)

var enabled = true

func init() {
	if os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb" {
		enabled = false
		return
	}
	// Check if stdout is a character device
	if fileInfo, err := os.Stdout.Stat(); err == nil {
		if (fileInfo.Mode() & os.ModeCharDevice) == 0 {
			// Redirected or piped
			if os.Getenv("CLICOLOR_FORCE") != "1" && os.Getenv("FORCE_COLOR") != "1" {
				enabled = false
			}
		}
	}
}

// SetEnabled overrides color output enablement.
func SetEnabled(val bool) {
	enabled = val
}

// IsEnabled returns true if colored output is active.
func IsEnabled() bool {
	return enabled
}

const (
	resetCode   = "\033[0m"
	boldCode    = "\033[1m"
	dimCode     = "\033[2m"
	redCode     = "\033[31m"
	greenCode   = "\033[32m"
	yellowCode  = "\033[33m"
	blueCode    = "\033[34m"
	magentaCode = "\033[35m"
	cyanCode    = "\033[36m"
	whiteCode   = "\033[37m"
	grayCode    = "\033[90m"
)

func colorize(code, s string) string {
	if !enabled || s == "" {
		return s
	}
	return code + s + resetCode
}

// Formatting helpers
func Bold(s string) string    { return colorize(boldCode, s) }
func Dim(s string) string     { return colorize(dimCode, s) }
func Red(s string) string     { return colorize(redCode, s) }
func Green(s string) string   { return colorize(greenCode, s) }
func Yellow(s string) string  { return colorize(yellowCode, s) }
func Blue(s string) string    { return colorize(blueCode, s) }
func Magenta(s string) string { return colorize(magentaCode, s) }
func Cyan(s string) string    { return colorize(cyanCode, s) }
func White(s string) string   { return colorize(whiteCode, s) }
func Gray(s string) string    { return colorize(grayCode, s) }

// High-level UI helpers
func Success(format string, a ...any) string {
	msg := fmt.Sprintf(format, a...)
	return Green("✓ ") + msg
}

func Info(format string, a ...any) string {
	msg := fmt.Sprintf(format, a...)
	return Cyan("ℹ ") + msg
}

func Warn(format string, a ...any) string {
	msg := fmt.Sprintf(format, a...)
	return Yellow("⚠ ") + msg
}

func Error(format string, a ...any) string {
	msg := fmt.Sprintf(format, a...)
	return Red("✗ ") + msg
}

func Bullet() string {
	return Green("•")
}

func Section(title string) string {
	return Bold(Yellow(title))
}

func KeyValue(key, val string) string {
	return fmt.Sprintf("  %-16s %s", Cyan(key+":"), val)
}

func StatusBadge(status string) string {
	switch strings.ToLower(status) {
	case "active", "online", "ok", "healthy", "succeeded":
		return Green(fmt.Sprintf("[%s]", status))
	case "paused", "firing", "running", "pending":
		return Yellow(fmt.Sprintf("[%s]", status))
	case "offline", "error", "failed", "alerting":
		return Red(fmt.Sprintf("[%s]", status))
	default:
		return Gray(fmt.Sprintf("[%s]", status))
	}
}
