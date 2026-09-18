package supervisor

import (
	"os"
	"strings"
)

// transientEnvVars are environment variables that should NOT be captured into supervisor configs.
var transientEnvVars = map[string]bool{
	"BEEP_DAEMON_CHILD": true,
}

// CaptureEnv snapshots critical environment variables (PATH, HOME, USER, and BEEP_*)
// so supervisor-started processes keep a usable runtime environment.
func CaptureEnv() map[string]string {
	env := make(map[string]string)

	// Always capture PATH, HOME, and USER if available
	for _, key := range []string{"PATH", "HOME", "USER", "LANG", "SHELL"} {
		if val, ok := os.LookupEnv(key); ok && val != "" {
			env[key] = val
		}
	}

	// Capture all relevant BEEP_* environment variables
	for _, e := range os.Environ() {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			key := parts[0]
			val := parts[1]
			if strings.HasPrefix(key, "BEEP_") && !transientEnvVars[key] && val != "" {
				env[key] = val
			}
		}
	}

	return env
}
