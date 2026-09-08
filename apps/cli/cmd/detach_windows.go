//go:build windows

package cmd

import (
	"os"
	"os/exec"
)

var shutdownSignals = []os.Signal{os.Interrupt}

func setProcessDetach(cmd *exec.Cmd) {
	// Child process detach logic for Windows
}

func checkChildExited(pid int) (bool, int) {
	return false, 0
}
