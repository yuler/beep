package cmd

import (
	cmdrunner "beep/cmd/runner"
)

var (
	flagWorkspace string
	flagServer    string
	flagToken     string
)

var runnerCmd = cmdrunner.NewCmdRunner()
