package supervisor

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"text/template"
)

const launchdPlistTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>{{.Label}}</string>
	<key>ProgramArguments</key>
	<array>
		<string>{{.ExecPath}}</string>
{{- range .Args}}
		<string>{{.}}</string>
{{- end}}
	</array>
	<key>WorkingDirectory</key>
	<string>{{.Workspace}}</string>
	<key>KeepAlive</key>
	<dict>
		<key>SuccessfulExit</key>
		<false/>
	</dict>
	<key>RunAtLoad</key>
	<true/>
{{- if .Env}}
	<key>EnvironmentVariables</key>
	<dict>
{{- range .Env}}
		<key>{{.Key}}</key>
		<string>{{.Value}}</string>
{{- end}}
	</dict>
{{- end}}
</dict>
</plist>
`

// LaunchdManager manages macOS LaunchAgent services.
type LaunchdManager struct {
	launchctlPath string
	agentDir      string
}

// NewLaunchdManager creates a new LaunchdManager.
func NewLaunchdManager() *LaunchdManager {
	launchctl, _ := exec.LookPath("launchctl")
	var agentDir string
	if home, err := os.UserHomeDir(); err == nil {
		agentDir = filepath.Join(home, "Library", "LaunchAgents")
	}

	return &LaunchdManager{
		launchctlPath: launchctl,
		agentDir:      agentDir,
	}
}

func (m *LaunchdManager) PlatformName() string {
	return "launchd"
}

func (m *LaunchdManager) IsSupported() bool {
	return runtime.GOOS == "darwin" && m.launchctlPath != ""
}

func (m *LaunchdManager) Label(service, binaryName string) string {
	if binaryName == "" {
		binaryName = "beep"
	}
	return fmt.Sprintf("com.%s.%s", binaryName, service)
}

func (m *LaunchdManager) plistPath(service, binaryName string) string {
	return filepath.Join(m.agentDir, fmt.Sprintf("%s.plist", m.Label(service, binaryName)))
}

func (m *LaunchdManager) guiTarget() string {
	uid := os.Getuid()
	return fmt.Sprintf("gui/%d", uid)
}

func (m *LaunchdManager) Install(info ServiceInfo) error {
	if !m.IsSupported() {
		return ErrUnsupported
	}

	if m.agentDir == "" {
		return fmt.Errorf("unable to resolve LaunchAgents directory")
	}

	if err := os.MkdirAll(m.agentDir, 0o755); err != nil {
		return fmt.Errorf("failed to create LaunchAgents directory: %w", err)
	}

	label := m.Label(info.Service, info.BinaryName)
	plistFile := m.plistPath(info.Service, info.BinaryName)

	rendered, err := renderLaunchdPlist(info, label)
	if err != nil {
		return err
	}

	if err := writePrivateFile(plistFile, []byte(rendered)); err != nil {
		return fmt.Errorf("failed to write plist file %s: %w", plistFile, err)
	}

	// Unload first so a rewrite can re-bootstrap (KeepAlive would otherwise race).
	target := m.guiTarget()
	_ = exec.Command(m.launchctlPath, "bootout", fmt.Sprintf("%s/%s", target, label)).Run()
	if out, err := exec.Command(m.launchctlPath, "bootstrap", target, plistFile).CombinedOutput(); err != nil {
		if loadOut, loadErr := exec.Command(m.launchctlPath, "load", "-w", plistFile).CombinedOutput(); loadErr != nil {
			return fmt.Errorf("launchctl load failed: %s (%w); bootstrap output: %s",
				strings.TrimSpace(string(loadOut)), loadErr, strings.TrimSpace(string(out)))
		}
	}

	return nil
}

func renderLaunchdPlist(info ServiceInfo, label string) (string, error) {
	args := make([]string, len(info.Args))
	for i, arg := range info.Args {
		args[i] = xmlEscape(arg)
	}

	type envItem struct {
		Key   string
		Value string
	}
	env := make([]envItem, 0, len(info.Env))
	keys := make([]string, 0, len(info.Env))
	for k := range info.Env {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := info.Env[k]
		if hasXMLUnsafeControl(k) || hasXMLUnsafeControl(v) {
			continue
		}
		env = append(env, envItem{Key: xmlEscape(k), Value: xmlEscape(v)})
	}

	data := struct {
		Label     string
		ExecPath  string
		Args      []string
		Workspace string
		Env       []envItem
	}{
		Label:     xmlEscape(label),
		ExecPath:  xmlEscape(info.ExecPath),
		Args:      args,
		Workspace: xmlEscape(info.Workspace),
		Env:       env,
	}

	tmpl, err := template.New("plist").Parse(launchdPlistTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse plist template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to render plist file: %w", err)
	}
	return buf.String(), nil
}

func (m *LaunchdManager) Uninstall(service, binaryName string) error {
	if !m.IsSupported() {
		return ErrUnsupported
	}

	label := m.Label(service, binaryName)
	plistFile := m.plistPath(service, binaryName)
	target := m.guiTarget()

	// Try bootout gui/<uid>/<label>, fallback to launchctl unload -w
	_ = exec.Command(m.launchctlPath, "bootout", fmt.Sprintf("%s/%s", target, label)).Run()
	_ = exec.Command(m.launchctlPath, "unload", "-w", plistFile).Run()

	if err := os.Remove(plistFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove plist file %s: %w", plistFile, err)
	}

	return nil
}

func (m *LaunchdManager) Start(service, binaryName string) error {
	if !m.IsSupported() {
		return ErrUnsupported
	}
	label := m.Label(service, binaryName)
	plistFile := m.plistPath(service, binaryName)
	target := m.guiTarget()
	domainLabel := fmt.Sprintf("%s/%s", target, label)

	_ = exec.Command(m.launchctlPath, "bootstrap", target, plistFile).Run()
	if err := exec.Command(m.launchctlPath, "kickstart", domainLabel).Run(); err != nil {
		if out, startErr := exec.Command(m.launchctlPath, "start", label).CombinedOutput(); startErr != nil {
			return fmt.Errorf("launchctl start failed: %s (%w)", strings.TrimSpace(string(out)), startErr)
		}
	}
	return nil
}

func (m *LaunchdManager) Stop(service, binaryName string) error {
	if !m.IsSupported() {
		return ErrUnsupported
	}
	label := m.Label(service, binaryName)
	plistFile := m.plistPath(service, binaryName)
	target := m.guiTarget()
	// bootout/unload actually stops KeepAlive jobs; launchctl stop would respawn.
	_ = exec.Command(m.launchctlPath, "bootout", fmt.Sprintf("%s/%s", target, label)).Run()
	_ = exec.Command(m.launchctlPath, "unload", plistFile).Run()
	return nil
}

func (m *LaunchdManager) GetStatus(service, binaryName string) Status {
	label := m.Label(service, binaryName)
	status := Status{
		Supported: m.IsSupported(),
		Platform:  "launchd",
		UnitName:  label,
	}

	if !status.Supported {
		return status
	}

	plistFile := m.plistPath(service, binaryName)
	if _, err := os.Stat(plistFile); err == nil {
		status.Installed = true
	}

	if status.Installed {
		target := m.guiTarget()
		if out, err := exec.Command(m.launchctlPath, "print", fmt.Sprintf("%s/%s", target, label)).CombinedOutput(); err == nil {
			status.Active = strings.Contains(string(out), "state = running")
		} else {
			// Fallback: check launchctl list
			if listOut, listErr := exec.Command(m.launchctlPath, "list").Output(); listErr == nil {
				for _, line := range strings.Split(string(listOut), "\n") {
					fields := strings.Fields(line)
					if len(fields) >= 3 && fields[2] == label {
						if pid, err := strconv.Atoi(fields[0]); err == nil && pid > 0 {
							status.Active = true
						}
						break
					}
				}
			}
		}
	}

	return status
}

func (m *LaunchdManager) EnsureLinger() (bool, error) {
	// macOS launchd agents run in user session; linger does not apply
	return true, nil
}
