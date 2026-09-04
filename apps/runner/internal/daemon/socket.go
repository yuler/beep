package daemon

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"beep-runner/internal/version"
)

// SocketStatus represents the ping/status payload sent over the domain socket.
type SocketStatus struct {
	Status    string `json:"status"`
	PID       int    `json:"pid"`
	Version   string `json:"version"`
	StartTime string `json:"start_time"`
}

// SocketListener manages the Unix domain socket lifecycle to prevent running multiple instances.
type SocketListener struct {
	path      string
	listener  net.Listener
	startTime time.Time
	mu        sync.Mutex
	closed    bool
}

// SocketPath returns the standard socket file path within a workspace directory.
func SocketPath(workspaceDir string) string {
	return filepath.Join(workspaceDir, ".socket")
}

// CheckRunning checks whether a runner daemon is actively listening on the workspace socket.
// Returns true, PID, and nil if running.
func CheckRunning(workspaceDir string) (bool, int, error) {
	socketPath := SocketPath(workspaceDir)
	if _, err := os.Stat(socketPath); os.IsNotExist(err) {
		return false, 0, nil
	}

	conn, err := net.DialTimeout("unix", socketPath, 300*time.Millisecond)
	if err != nil {
		return false, 0, nil
	}
	defer conn.Close()

	_ = conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	var status SocketStatus
	if err := json.NewDecoder(conn).Decode(&status); err != nil {
		// Connected but failed to parse JSON, still considered running
		return true, 0, nil
	}
	return true, status.PID, nil
}

// AcquireSocket attempts to create and listen on the Unix domain socket in workspaceDir.
// If another instance is active, it returns an error. If a stale socket file exists, it cleans it up.
func AcquireSocket(workspaceDir string) (*SocketListener, error) {
	if err := os.MkdirAll(workspaceDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create workspace directory %s: %w", workspaceDir, err)
	}

	socketPath := SocketPath(workspaceDir)

	// Check if socket file exists
	if _, err := os.Stat(socketPath); err == nil {
		conn, dialErr := net.DialTimeout("unix", socketPath, 300*time.Millisecond)
		if dialErr == nil {
			// Another instance is actively listening
			conn.Close()
			return nil, fmt.Errorf("runner daemon is already running (socket: %s)", socketPath)
		}
		// Stale socket file, remove it
		_ = os.Remove(socketPath)
	}

	l, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on socket %s: %w", socketPath, err)
	}

	_ = os.Chmod(socketPath, 0o600)

	sl := &SocketListener{
		path:      socketPath,
		listener:  l,
		startTime: time.Now(),
	}

	go sl.serve()

	return sl, nil
}

func (s *SocketListener) serve() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return
		}
		go s.handleConn(conn)
	}
}

func (s *SocketListener) handleConn(conn net.Conn) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(2 * time.Second))

	status := SocketStatus{
		Status:    "running",
		PID:       os.Getpid(),
		Version:   version.Version,
		StartTime: s.startTime.Format(time.RFC3339),
	}

	_ = json.NewEncoder(conn).Encode(status)
}

// Path returns the path of the socket.
func (s *SocketListener) Path() string {
	return s.path
}

// Close closes the Unix socket and removes the socket file.
func (s *SocketListener) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}
	s.closed = true

	var err error
	if s.listener != nil {
		err = s.listener.Close()
	}
	_ = os.Remove(s.path)
	return err
}
