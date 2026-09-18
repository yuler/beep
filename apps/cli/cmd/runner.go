package cmd

import (
	cmdrunner "beep/cmd/runner"
)

var (
	flagWorkspace string
	flagServer    string
)

var runnerCmd = cmdrunner.NewCmdRunner()
