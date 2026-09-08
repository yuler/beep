//go:build windows

package exec

import (
	"os/exec"
)

func prepareProcessGroup(cmd *exec.Cmd) {
	cmd.Cancel = func() error {
		if cmd.Process != nil {
			return cmd.Process.Kill()
		}
		return nil
	}
}

func killProcessGroup(cmd *exec.Cmd) {
	if cmd.Process != nil {
		_ = cmd.Process.Kill()
	}
}
