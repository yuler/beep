package cmd

import (
	"strings"

	cmdrunner "beep/cmd/runner"
	"beep/internal/cliservice"
	"beep/internal/cmdutil"
	"beep/internal/config"

	"github.com/spf13/cobra"
)

var (
	flagWorkspace string
	flagServer    string
	flagToken     string
)

var runnerCmd = cmdrunner.NewCmdRunner()

func newRunnerConnectCmd() *cobra.Command {
	cmd := cmdrunner.NewCmdConnect()
	origRunE := cmd.RunE
	cmd.RunE = func(c *cobra.Command, args []string) error {
		if flagWorkspace != "" {
			cmdutil.SetOverrideWorkspace(flagWorkspace)
			defer cmdutil.SetOverrideWorkspace("")
		}
		if startServiceDaemonFn != nil {
			orig := cliservice.StartServiceDaemonFn
			cliservice.StartServiceDaemonFn = startServiceDaemonFn
			defer func() { cliservice.StartServiceDaemonFn = orig }()
		}
		return origRunE(c, args)
	}
	return cmd
}

func newRunnerDisconnectCmd() *cobra.Command {
	cmd := cmdrunner.NewCmdDisconnect()
	origRunE := cmd.RunE
	cmd.RunE = func(c *cobra.Command, args []string) error {
		if flagWorkspace != "" {
			cmdutil.SetOverrideWorkspace(flagWorkspace)
			defer cmdutil.SetOverrideWorkspace("")
		}
		return origRunE(c, args)
	}
	return cmd
}

func newRunnerUpCmd() *cobra.Command {
	return cmdrunner.NewCmdUp()
}

func newRunnerStopCmd() *cobra.Command {
	return cmdrunner.NewCmdStop()
}

func newRunnerStatusCmd() *cobra.Command {
	return cmdrunner.NewCmdStatus()
}

func loadConfig() (*config.Config, error) {
	workspace := flagWorkspace
	if workspace == "" {
		workspace = cmdutil.GetWorkspace(nil)
	}
	cfg, err := config.Load(workspace)
	if err != nil {
		return nil, err
	}
	if flagServer != "" {
		cfg.ServerURL = flagServer
	}
	if flagToken != "" {
		cfg.RunnerToken = flagToken
	}
	if flagWorkspace != "" {
		cfg.Workspace = flagWorkspace
	}
	if flagAccount != "" {
		cfg.AccountSlug = strings.TrimSpace(flagAccount)
	}
	return cfg, nil
}
