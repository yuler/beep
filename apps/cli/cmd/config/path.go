package config

import (
	"fmt"

	"beep/internal/cmdutil"
	"beep/internal/config"

	"github.com/spf13/cobra"
)

// NewCmdPath creates the 'config path' subcommand.
func NewCmdPath() *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Print path to the configuration file",
		Run: func(cmd *cobra.Command, args []string) {
			workspace := cmdutil.GetWorkspace(cmd)
			fmt.Println(config.GetConfigPath(workspace))
		},
	}
}
