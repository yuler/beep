//go:build windows

package cmd

import (
	"os"
	"os/exec"
	"syscall"
)

var shutdownSignals = []os.Signal{os.Interrupt}

func setProcessDetach(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
}

func checkChildExited(pid int) (bool, int) {
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
