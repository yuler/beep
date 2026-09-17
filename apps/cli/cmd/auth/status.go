package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"beep/internal/client"
	"beep/internal/cmdutil"
	"beep/internal/config"
	"beep/internal/ui"

	"github.com/spf13/cobra"
)

// NewCmdStatus creates the 'auth status' subcommand.
func NewCmdStatus() *cobra.Command {
	return &cobra.Command{
		Use:     "status",
		Aliases: []string{"whoami"},
		Short:   "View current authentication status",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig(cmd)
			if err != nil {
				return err
			}

			if !cfg.IsLoggedIn() {
				fmt.Println(ui.Dim("Not logged in to Beep."))
				if cfg.ServerURL != "" {
					fmt.Printf("Server: %s\n", ui.Cyan(cfg.ServerURL))
				}
				fmt.Printf("\n%s Run %s to log in.\n", ui.Dim("Tip:"), ui.Cyan(config.BinaryName()+" auth login"))
				return nil
			}

			c := client.New(cfg)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			var me *client.MeResponse
			err = ui.WithSpinner("Verifying authentication with server...", func() error {
				var getErr error
				me, getErr = c.GetMe(ctx)
				return getErr
			})

			if err != nil {
				if strings.Contains(err.Error(), "401") || strings.Contains(err.Error(), "invalid or expired") {
					fmt.Println(ui.Error("Stored login session is expired or invalid."))
					fmt.Printf("Run %s to re-authenticate.\n", ui.Cyan(config.BinaryName()+" auth login"))
					return nil
				}
				// Offline or network error
				fmt.Println(ui.Yellow("⚠ Could not verify authentication with server (network error): ") + err.Error())
				fmt.Printf("Stored token: %s\n", ui.Dim(config.MaskToken(cfg.AccessToken)))
				if cfg.UserEmail != "" {
					fmt.Printf("Logged in as: %s\n", ui.Cyan(cfg.UserEmail))
				}
				return nil
			}

			displayName := me.Identity.Name
			if displayName == "" {
				displayName = me.Identity.Email
			} else {
				displayName = fmt.Sprintf("%s (%s)", displayName, me.Identity.Email)
			}

			fmt.Printf("%s Logged in to %s\n", ui.Green("✓"), ui.Bold(cfg.ServerURL))
			fmt.Printf("  %s %s\n", ui.Dim("User:"), ui.Cyan(displayName))
			if cfg.AccountSlug != "" {
				fmt.Printf("  %s %s\n", ui.Dim("Active account:"), ui.Yellow(cfg.AccountSlug))
			} else if me.LastAccountSlug != "" {
				fmt.Printf("  %s %s\n", ui.Dim("Active account:"), ui.Yellow(me.LastAccountSlug))
			}
			fmt.Printf("  %s %s\n", ui.Dim("Token:"), ui.Dim(config.MaskToken(cfg.AccessToken)))

			if len(me.Accounts) > 1 {
				fmt.Println()
				fmt.Println(ui.Dim("Available accounts:"))
				for _, acc := range me.Accounts {
					prefix := "  • "
					if acc.Slug == cfg.AccountSlug || (cfg.AccountSlug == "" && acc.Slug == me.LastAccountSlug) {
						prefix = "  * "
					}
					fmt.Printf("%s%s (%s)\n", prefix, ui.Bold(acc.Name), ui.Dim(acc.Slug))
				}
			}

			return nil
		},
	}
}
