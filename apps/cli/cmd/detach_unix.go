//go:build !windows

package cmd

import (
	"os"
	"os/exec"
	"syscall"
)

var shutdownSignals = []os.Signal{os.Interrupt, syscall.SIGTERM}

func setProcessDetach(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}
}

func checkChildExited(pid int) (bool, int) {
	var ws syscall.WaitStatus
	var ru syscall.Rusage
	wpid, waitErr := syscall.Wait4(pid, &ws, syscall.WNOHANG, &ru)
	if waitErr == nil && wpid == pid {
		return true, ws.ExitStatus()
	}
	return false, 0
}
