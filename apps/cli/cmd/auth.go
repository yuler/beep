package cmd

import (
	cmdauth "beep/cmd/auth"
	"beep/internal/cmdutil"

	"github.com/spf13/cobra"
)

var (
	authCmd       = cmdauth.NewCmdAuth()
	authLogoutCmd = &cobra.Command{
		Use:   "logout",
		Short: "Log out of Beep on this machine",
		RunE: func(cmd *cobra.Command, args []string) error {
			if flagWorkspace != "" {
				cmdutil.SetOverrideWorkspace(flagWorkspace)
				defer cmdutil.SetOverrideWorkspace("")
			}
			return cmdauth.NewCmdLogout().RunE(cmd, args)
		},
	}
)
