package service

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// NewCmdService creates the 'beep service' parent command.
func NewCmdService() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "service",
		Short: "Manage local background daemon services (runner and channel)",
		Long:  "Manage local background daemon services for executing jobs (runner) and receiving notifications (channel).",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(NewCmdStart())
	cmd.AddCommand(NewCmdStop())
	cmd.AddCommand(NewCmdRestart())
	cmd.AddCommand(NewCmdStatus())

	return cmd
}

func parseServiceArg(args []string) (string, error) {
	if len(args) == 0 {
		return "", nil
	}
	target := strings.ToLower(args[0])
	if target != "runner" && target != "channel" {
		return "", fmt.Errorf("unknown service %q (expected 'runner' or 'channel')", target)
	}
	return target, nil
}
