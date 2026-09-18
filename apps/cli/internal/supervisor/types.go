package supervisor

import (
	"errors"
)

var (
	// ErrUnsupported is returned when no OS process supervisor is available.
	ErrUnsupported = errors.New("autostart is not supported on this platform (requires systemd on Linux or LaunchAgent on macOS)")
)

// ServiceInfo contains details required to generate and register a supervisor unit.
type ServiceInfo struct {
	Service     string            // "runner" or "channel"
	BinaryName  string            // "beep" or "beep-local"
	Description string            // Human-readable description
	ExecPath    string            // Absolute path to the executable
	Args        []string          // Command line arguments for foreground run (e.g. ["runner", "start"])
	Workspace   string            // Working directory
	Env         map[string]string // Captured environment variables
	LogPath     string            // Unused by supervisors; daemons write daily logs themselves
}

// Status represents the autostart status of a service.
type Status struct {
	Supported    bool
	Platform     string // "systemd", "launchd", or "unsupported"
	Installed    bool
	Active       bool
	UnitName     string
	LingerActive bool   // Linux only: whether loginctl linger is enabled
	Detail       string // Additional status details or error message
}

// Manager talks to the host process supervisor (systemd user units or LaunchAgents).
type Manager interface {
	IsSupported() bool
	PlatformName() string
	Install(info ServiceInfo) error
	Uninstall(service, binaryName string) error
	Start(service, binaryName string) error
	Stop(service, binaryName string) error
	Restart(service, binaryName string) error
	GetStatus(service, binaryName string) Status
	EnsureLinger() (bool, error)
}
