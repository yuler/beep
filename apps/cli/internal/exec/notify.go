package exec

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
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
			script := fmt.Sprintf("display notification %q with title %q", body, title)
			_ = exec.CommandContext(ctx, path, "-e", script).Run()
		}
	}
}
