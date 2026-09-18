package supervisor

import (
	"os"
)

// CaptureEnv snapshots PATH, HOME, USER, LANG, and SHELL for supervisor units.
// Tokens and other BEEP_* values are not copied; daemons load those from config.json.
func CaptureEnv() map[string]string {
	env := make(map[string]string)
	for _, key := range []string{"PATH", "HOME", "USER", "LANG", "SHELL"} {
		if val, ok := os.LookupEnv(key); ok && val != "" {
			env[key] = val
		}
	}
	return env
}
