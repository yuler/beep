//go:build !windows

package daemon

import (
	"os"
	"syscall"
)

func terminateProcess(proc *os.Process) error {
	return proc.Signal(syscall.SIGTERM)
}

func killProcess(proc *os.Process, pid int) {
	if pid > 1 {
		_ = syscall.Kill(-pid, syscall.SIGKILL)
	}
	_ = proc.Signal(syscall.SIGKILL)
}
