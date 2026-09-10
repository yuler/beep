//go:build windows

// Package proc hides the platform differences of process lifecycle handling:
// detaching daemon children, job process groups, and terminating processes.
package proc

import (
	"os"
	"os/exec"
	"syscall"
)

// ShutdownSignals request a graceful shutdown of the runner daemon.
var ShutdownSignals = []os.Signal{os.Interrupt}

// Detach configures cmd to outlive the parent process (new process group).
func Detach(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
}

// Setpgid is a no-op on Windows: there is no separate job process group to adopt.
func Setpgid(_ *exec.Cmd) {}

// Terminate stops proc. Console interrupt signals fail for processes outside
// the caller's console on Windows, so terminate means kill here.
func Terminate(proc *os.Process) error {
	return proc.Kill()
}

// Kill force-kills proc; process groups do not apply on Windows.
func Kill(proc *os.Process, _ int) error {
	return proc.Kill()
}

// Exited reports whether the child pid has exited, and its exit status.
func Exited(pid int) (bool, int) {
	h, err := syscall.OpenProcess(syscall.PROCESS_QUERY_INFORMATION|syscall.SYNCHRONIZE, false, uint32(pid))
	if err != nil {
		return false, 0
	}
	defer syscall.CloseHandle(h)

	event, err := syscall.WaitForSingleObject(h, 0)
	if err == nil && event == syscall.WAIT_OBJECT_0 {
		var code uint32
		if err := syscall.GetExitCodeProcess(h, &code); err == nil {
			return true, int(code)
		}
		return true, 0
	}
	return false, 0
}
