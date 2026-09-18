package supervisor

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"text/template"
)

const systemdUnitTemplate = `[Unit]
Description={{.Description}}
After=network.target

[Service]
Type=simple
ExecStart={{.ExecCmd}}
WorkingDirectory={{.Workspace}}
Restart=on-failure
RestartSec=5s
{{- range .EnvLines}}
Environment={{.}}
{{- end}}

[Install]
WantedBy=default.target
`

// SystemdManager manages systemd user-level services.
type SystemdManager struct {
	// systemctlPath and loginctlPath allow mocking in tests
	systemctlPath string
	loginctlPath  string
	userDir       string
}

// NewSystemdManager creates a new SystemdManager.
func NewSystemdManager() *SystemdManager {
	sysPath, _ := exec.LookPath("systemctl")
	loginPath, _ := exec.LookPath("loginctl")

	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		if home, err := os.UserHomeDir(); err == nil {
			configHome = filepath.Join(home, ".config")
		}
	}

	var userDir string
	if configHome != "" {
		userDir = filepath.Join(configHome, "systemd", "user")
	}

	return &SystemdManager{
		systemctlPath: sysPath,
		loginctlPath:  loginPath,
		userDir:       userDir,
	}
}

func (m *SystemdManager) PlatformName() string {
	return "systemd"
}

func (m *SystemdManager) IsSupported() bool {
	if runtime.GOOS != "linux" || m.systemctlPath == "" {
		return false
	}
	// Check if systemd user manager is accessible
	cmd := exec.Command(m.systemctlPath, "--user", "is-system-running")
	out, err := cmd.CombinedOutput()
	if err != nil {
		// systemctl is-system-running can return non-zero for "degraded" or "starting",
		// which still means systemd is functional. Check the output string.
		state := strings.TrimSpace(string(out))
		if state == "degraded" || state == "running" || state == "starting" {
			return true
		}
		// Try a fallback query
		checkCmd := exec.Command(m.systemctlPath, "--user", "status")
		return checkCmd.Run() == nil
	}
	return true
}

func (m *SystemdManager) UnitName(service, binaryName string) string {
	if binaryName == "" {
		binaryName = "beep"
	}
	return fmt.Sprintf("%s-%s.service", binaryName, service)
}

func (m *SystemdManager) unitFilePath(service, binaryName string) string {
	return filepath.Join(m.userDir, m.UnitName(service, binaryName))
}

func (m *SystemdManager) Install(info ServiceInfo) error {
	if !m.IsSupported() {
		return ErrUnsupported
	}

	if m.userDir == "" {
		return fmt.Errorf("unable to resolve user configuration directory")
	}

	if err := os.MkdirAll(m.userDir, 0o755); err != nil {
		return fmt.Errorf("failed to create systemd user directory: %w", err)
	}

	unitName := m.UnitName(info.Service, info.BinaryName)
	unitPath := filepath.Join(m.userDir, unitName)

	rendered, err := renderSystemdUnit(info)
	if err != nil {
		return err
	}

	if err := writePrivateFile(unitPath, []byte(rendered)); err != nil {
		return fmt.Errorf("failed to write unit file %s: %w", unitPath, err)
	}

	// Reload systemd daemon
	if out, err := exec.Command(m.systemctlPath, "--user", "daemon-reload").CombinedOutput(); err != nil {
		return fmt.Errorf("systemctl --user daemon-reload failed: %s (%w)", strings.TrimSpace(string(out)), err)
	}

	// Enable unit
	if out, err := exec.Command(m.systemctlPath, "--user", "enable", unitName).CombinedOutput(); err != nil {
		return fmt.Errorf("systemctl --user enable %s failed: %s (%w)", unitName, strings.TrimSpace(string(out)), err)
	}

	// Try to enable user lingering so the user service survives reboot without active login
	_, _ = m.EnsureLinger()

	return nil
}

func (m *SystemdManager) Uninstall(service, binaryName string) error {
	if !m.IsSupported() {
		return ErrUnsupported
	}

	unitName := m.UnitName(service, binaryName)
	unitPath := m.unitFilePath(service, binaryName)

	// Stop unit (ignore errors if not running)
	_ = exec.Command(m.systemctlPath, "--user", "stop", unitName).Run()

	// Disable unit (ignore errors if not enabled)
	_ = exec.Command(m.systemctlPath, "--user", "disable", unitName).Run()

	// Remove unit file
	if err := os.Remove(unitPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove unit file %s: %w", unitPath, err)
	}

	// Reload daemon
	_ = exec.Command(m.systemctlPath, "--user", "daemon-reload").Run()

	return nil
}

func (m *SystemdManager) Start(service, binaryName string) error {
	if !m.IsSupported() {
		return ErrUnsupported
	}
	unitName := m.UnitName(service, binaryName)
	out, err := exec.Command(m.systemctlPath, "--user", "start", unitName).CombinedOutput()
	if err != nil {
		return fmt.Errorf("systemctl --user start %s failed: %s (%w)", unitName, strings.TrimSpace(string(out)), err)
	}
	return nil
}

func (m *SystemdManager) Stop(service, binaryName string) error {
	if !m.IsSupported() {
		return ErrUnsupported
	}
	unitName := m.UnitName(service, binaryName)
	out, err := exec.Command(m.systemctlPath, "--user", "stop", unitName).CombinedOutput()
	if err != nil {
		return fmt.Errorf("systemctl --user stop %s failed: %s (%w)", unitName, strings.TrimSpace(string(out)), err)
	}
	return nil
}

func (m *SystemdManager) GetStatus(service, binaryName string) Status {
	unitName := m.UnitName(service, binaryName)
	status := Status{
		Supported: m.IsSupported(),
		Platform:  "systemd",
		UnitName:  unitName,
	}

	if !status.Supported {
		return status
	}

	unitPath := m.unitFilePath(service, binaryName)
	if _, err := os.Stat(unitPath); err == nil {
		status.Installed = true
	}

	if status.Installed {
		cmd := exec.Command(m.systemctlPath, "--user", "is-active", "--quiet", unitName)
		status.Active = (cmd.Run() == nil)
	}

	status.LingerActive, _ = m.checkLinger()
	return status
}

func (m *SystemdManager) checkLinger() (bool, error) {
	if m.loginctlPath == "" {
		return false, fmt.Errorf("loginctl not found")
	}
	user := os.Getenv("USER")
	if user == "" {
		return false, fmt.Errorf("USER environment variable not set")
	}

	out, err := exec.Command(m.loginctlPath, "show-user", user, "--property=Linger").Output()
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(out)) == "Linger=yes", nil
}

func (m *SystemdManager) EnsureLinger() (bool, error) {
	active, err := m.checkLinger()
	if err == nil && active {
		return true, nil
	}

	if m.loginctlPath == "" {
		return false, fmt.Errorf("loginctl not found")
	}

	user := os.Getenv("USER")
	cmd := exec.Command(m.loginctlPath, "enable-linger")
	if user != "" {
		cmd = exec.Command(m.loginctlPath, "enable-linger", user)
	}

	if out, err := cmd.CombinedOutput(); err != nil {
		return false, fmt.Errorf("enable-linger failed: %s (%w)", strings.TrimSpace(string(out)), err)
	}

	return m.checkLinger()
}

func renderSystemdUnit(info ServiceInfo) (string, error) {
	var cmdParts []string
	cmdParts = append(cmdParts, systemdQuote(info.ExecPath))
	for _, arg := range info.Args {
		cmdParts = append(cmdParts, systemdQuote(arg))
	}

	desc := info.Description
	if desc == "" {
		desc = fmt.Sprintf("Beep %s Daemon", ServiceTitle(info.Service))
	}

	envLines := make([]string, 0, len(info.Env))
	keys := make([]string, 0, len(info.Env))
	for k := range info.Env {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		line, ok := systemdEnvironmentLine(k, info.Env[k])
		if ok {
			envLines = append(envLines, line)
		}
	}

	data := struct {
		Description string
		ExecCmd     string
		Workspace   string
		EnvLines    []string
	}{
		Description: desc,
		ExecCmd:     strings.Join(cmdParts, " "),
		Workspace:   systemdQuote(info.Workspace),
		EnvLines:    envLines,
	}

	tmpl, err := template.New("unit").Parse(systemdUnitTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse unit template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to render unit file: %w", err)
	}
	return buf.String(), nil
}
