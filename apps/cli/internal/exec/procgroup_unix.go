//go:build !windows

package exec

import (
	"os/exec"
	"syscall"
)

func prepareProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		if cmd.Process != nil && cmd.Process.Pid > 1 {
			return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		}
		return nil
	}
}

func killProcessGroup(cmd *exec.Cmd) {
	if cmd.Process != nil && cmd.Process.Pid > 1 {
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
}
