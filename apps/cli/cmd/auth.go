package cmd

import (
	"context"

	cmdauth "beep/cmd/auth"
	"beep/internal/client"
	"beep/internal/cmdutil"
	"beep/internal/config"

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

func ensureLoggedIn(ctx context.Context, cfg *config.Config) (*client.MeResponse, error) {
	return cmdutil.EnsureLoggedIn(ctx, cfg)
}

func resolveAccountSlug(me *client.MeResponse, explicitAccount string, cfgAccount string) (string, error) {
	return cmdutil.ResolveAccountSlug(me, explicitAccount, cfgAccount)
}
