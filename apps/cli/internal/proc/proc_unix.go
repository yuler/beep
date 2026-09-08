//go:build !windows

// Package proc hides the platform differences of process lifecycle handling:
// detaching daemon children, job process groups, and terminating processes.
package proc

import (
	"os"
	"os/exec"
	"syscall"
)

// ShutdownSignals request a graceful shutdown of the runner daemon.
var ShutdownSignals = []os.Signal{os.Interrupt, syscall.SIGTERM}

// Detach configures cmd to outlive the parent process (new session).
func Detach(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}

// Setpgid puts cmd in its own process group so Kill can reap the whole tree.
func Setpgid(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// Terminate asks proc to shut down gracefully.
func Terminate(proc *os.Process) error {
	return proc.Signal(syscall.SIGTERM)
}

// Kill force-kills proc and, when pid leads a process group, the whole group.
func Kill(proc *os.Process, pid int) error {
	if pid > 1 {
		_ = syscall.Kill(-pid, syscall.SIGKILL)
	}
	return proc.Signal(syscall.SIGKILL)
}

// Exited reports whether the child pid has exited, and its exit status.
func Exited(pid int) (bool, int) {
	var ws syscall.WaitStatus
	var ru syscall.Rusage
	wpid, waitErr := syscall.Wait4(pid, &ws, syscall.WNOHANG, &ru)
	if waitErr == nil && wpid == pid {
		return true, ws.ExitStatus()
	}
	return false, 0
}
