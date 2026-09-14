package exec

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// NotifyDesktop sends an OS-native desktop notification if a supported tool is available.
func NotifyDesktop(title, body string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	switch runtime.GOOS {
	case "linux":
		if path, err := exec.LookPath("notify-send"); err == nil {
			_ = exec.CommandContext(ctx, path, "-a", "Beep", title, body).Run()
		}
	case "darwin":
		if path, err := exec.LookPath("osascript"); err == nil {
			script := fmt.Sprintf("display notification %s with title %s",
				appleScriptString(body), appleScriptString(title))
			_ = exec.CommandContext(ctx, path, "-e", script).Run()
		}
	}
}

func appleScriptString(s string) string {
	if len(s) > 500 {
		s = s[:500]
	}
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '\\', '"':
			b.WriteByte('\\')
			b.WriteRune(r)
		case '\n', '\r':
			b.WriteByte(' ')
		default:
			if r < 0x20 {
				b.WriteByte(' ')
			} else {
				b.WriteRune(r)
			}
		}
	}
	b.WriteByte('"')
	return b.String()
}
