package supervisor

import (
	"runtime"
)

// DefaultManager can be replaced in unit tests.
var DefaultManager Manager

// CurrentManager returns the appropriate Manager for the current host environment.
func CurrentManager() Manager {
	if DefaultManager != nil {
		return DefaultManager
	}

	switch runtime.GOOS {
	case "linux":
		sm := NewSystemdManager()
		if sm.IsSupported() {
			return sm
		}
	case "darwin":
		lm := NewLaunchdManager()
		if lm.IsSupported() {
			return lm
		}
	}

	return NewUnsupportedManager()
}
