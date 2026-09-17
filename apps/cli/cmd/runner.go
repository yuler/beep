package cmd

import (
	"strings"

	cmdrunner "beep/cmd/runner"
	"beep/internal/cmdutil"
	"beep/internal/config"
)

var (
	flagWorkspace string
	flagServer    string
	flagToken     string
)

var runnerCmd = cmdrunner.NewCmdRunner()

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
