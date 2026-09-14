package browser

import (
	"fmt"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
)

func Open(rawURL string) error {
	rawURL = strings.TrimSpace(rawURL)
	u, err := url.Parse(rawURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("refusing to open invalid URL")
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme != "https" {
		if scheme != "http" || !isLocalhost(u.Hostname()) {
			return fmt.Errorf("refusing to open non-HTTPS URL")
		}
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", rawURL)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", rawURL)
	default:
		cmd = exec.Command("xdg-open", rawURL)
	}
	return cmd.Start()
}

func isLocalhost(host string) bool {
	h := strings.ToLower(strings.TrimSpace(host))
	if h == "localhost" || h == "127.0.0.1" || h == "::1" {
		return true
	}
	return strings.HasSuffix(h, ".localhost")
}
