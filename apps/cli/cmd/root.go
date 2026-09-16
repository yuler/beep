package cmd

import (
	"fmt"
	"os"
	"strings"

	"beep/cmd/beep"
	"beep/cmd/beeper"
	"beep/internal/config"
	"beep/internal/ui"
	"beep/internal/updater"

	"github.com/spf13/cobra"
)

var (
	flagNoColor       bool
	flagNoInteractive bool
	flagAccount       string
	flagJSON          bool
)

// skipUpdateHooks returns true for commands that should not trigger background
// update checks or print update notices (upgrade/update/version/completion/help
// and cobra internal __* commands).
func skipUpdateHooks(name string) bool {
	switch name {
	case "upgrade", "update", "version", "completion", "help":
		return true
	}
	return strings.HasPrefix(name, "__")
}

var RootCmd = &cobra.Command{
	Use:   "beep",
	Short: "Beep command-line interface",
	Long: ui.Bold(ui.Cyan("Beep CLI")) + ` - Command-line interface for the Beep platform.

Execute 'beep <command> --help' for detailed usage of a specific command.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if flagNoColor {
			ui.SetEnabled(false)
		}
		if !skipUpdateHooks(cmd.Name()) {
			updater.TriggerBackgroundCheck(flagWorkspace)
		}
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if skipUpdateHooks(cmd.Name()) {
			return
		}
		if notice := updater.CheckNotice(flagWorkspace); notice != "" {
			if _, err := fmt.Fprint(os.Stderr, notice); err == nil {
				_ = updater.MarkNotified(flagWorkspace)
			}
		}
	},
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, ui.Error("%v", err))
		os.Exit(1)
	}
}

func init() {
	RootCmd.PersistentFlags().BoolVar(&flagNoColor, "no-color", false, "Disable colored output")
	RootCmd.PersistentFlags().BoolVar(&flagNoInteractive, "no-interactive", false, "Disable interactive prompts")
	RootCmd.PersistentFlags().StringVarP(&flagWorkspace, "workspace", "w", "", fmt.Sprintf("Local job workspace directory (default %s, env: BEEP_WORKSPACE)", config.DefaultWorkspaceDisplay()))
	RootCmd.PersistentFlags().StringVarP(&flagServer, "server", "s", "", "Beep server URL (env: BEEP_SERVER)")
	RootCmd.PersistentFlags().StringVarP(&flagToken, "token", "t", "", "Runner authentication token (env: BEEP_RUNNER_TOKEN)")
	RootCmd.PersistentFlags().StringVarP(&flagAccount, "account", "a", "", "Account slug to operate on (defaults to config account_slug or personal account; env: BEEP_ACCOUNT)")
	RootCmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "Output results in JSON format")

	// Command groups
	RootCmd.AddGroup(
		&cobra.Group{ID: "beeps", Title: "Beep Commands:"},
	)

	// Flattened beep commands (canonical)
	beepCommands := []*cobra.Command{
		beep.NewCmdList(),
		beep.NewCmdShow(),
		beep.NewCmdCreate(),
		beep.NewCmdDelete(),
		beep.NewCmdPause(),
		beep.NewCmdResume(),
		beep.NewCmdRun(),
	}
	for _, cmd := range beepCommands {
		cmd.GroupID = "beeps"
		RootCmd.AddCommand(cmd)
	}

	// Legacy beep namespace (hidden compatibility alias)
	legacyBeepCmd := beep.NewCmdBeep()
	legacyBeepCmd.Hidden = true
	RootCmd.AddCommand(legacyBeepCmd)

	// Beeper commands
	RootCmd.AddCommand(beeper.NewCmdBeeper())

	// Daemon commands
	RootCmd.AddCommand(newUpCmd())
	RootCmd.AddCommand(newStopCmd())
	RootCmd.AddCommand(newStatusCmd())

	// System / management subcommands
	RootCmd.AddCommand(authCmd)
	RootCmd.AddCommand(runnerCmd)
	RootCmd.AddCommand(channelCmd)
	RootCmd.AddCommand(configCmd)
	RootCmd.AddCommand(versionCmd)
	RootCmd.AddCommand(upgradeCmd)
}
