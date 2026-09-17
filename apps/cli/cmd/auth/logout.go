package auth

import (
	"fmt"

	"beep/internal/cmdutil"
	"beep/internal/config"
	"beep/internal/ui"

	"github.com/spf13/cobra"
)

// NewCmdLogout creates the 'auth logout' subcommand.
func NewCmdLogout() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Log out of Beep on this machine",
		RunE: func(cmd *cobra.Command, args []string) error {
			workspace := cmdutil.GetWorkspace(cmd)
			configPath := config.GetConfigPath(workspace)
			fc, err := config.LoadFile(configPath)
			if err != nil || fc.AccessToken == "" {
				fmt.Println(ui.Dim("No active login session found in " + configPath))
				return nil
			}

			userEmail := fc.UserEmail
			fc.AccessToken = ""
			fc.UserEmail = ""
			fc.UserName = ""

			if err := config.SaveFile(configPath, fc); err != nil {
				return fmt.Errorf("failed to update config file: %w", err)
			}

			if userEmail != "" {
				fmt.Printf("%s Logged out of %s\n", ui.Green("✓"), ui.Cyan(userEmail))
			} else {
				fmt.Printf("%s Logged out successfully\n", ui.Green("✓"))
			}
			return nil
		},
	}
}
