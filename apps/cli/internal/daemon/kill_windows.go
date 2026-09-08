//go:build windows

package daemon

import (
	"os"
)

func terminateProcess(proc *os.Process) error {
	return proc.Signal(os.Interrupt)
}

func killProcess(proc *os.Process, _ int) {
	_ = proc.Kill()
}
