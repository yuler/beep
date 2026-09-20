package supervisor

// UnsupportedManager is used when no system supervisor is available.
type UnsupportedManager struct{}

func NewUnsupportedManager() *UnsupportedManager {
	return &UnsupportedManager{}
}

func (m *UnsupportedManager) PlatformName() string {
	return "unsupported"
}

func (m *UnsupportedManager) IsSupported() bool {
	return false
}

func (m *UnsupportedManager) Install(info ServiceInfo) error {
	return ErrUnsupported
}

func (m *UnsupportedManager) Uninstall(service, binaryName string) error {
	return ErrUnsupported
}

func (m *UnsupportedManager) Start(service, binaryName string) error {
	return ErrUnsupported
}

func (m *UnsupportedManager) Stop(service, binaryName string) error {
	return ErrUnsupported
}

func (m *UnsupportedManager) GetStatus(service, binaryName string) Status {
	return Status{
		Supported: false,
		Platform:  "unsupported",
		Detail:    "process supervisor requires systemd (Linux) or launchd (macOS)",
	}
}

func (m *UnsupportedManager) EnsureLinger() (bool, error) {
	return false, ErrUnsupported
}
