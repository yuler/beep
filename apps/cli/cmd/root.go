package cmd

import (
	"fmt"
	"os"
	"strings"

	"beep/cmd/beep"
	"beep/cmd/beeper"
	"beep/cmd/service"
	"beep/internal/cmdutil"
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
func skipUpdateHooks(cmd *cobra.Command) bool {
	for c := cmd; c != nil; c = c.Parent() {
		name := c.Name()
		switch name {
		case "upgrade", "update", "version", "completion", "help":
			return true
		}
		if strings.HasPrefix(name, "__") {
			return true
		}
	}
	return false
}

var RootCmd = &cobra.Command{
	Use:   "beep",
	Short: "Beep command-line interface",
	Long:  ui.Bold(ui.Cyan("Beep CLI")) + " - Command-line interface for the Beep platform.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if flagNoColor {
			ui.SetEnabled(false)
		}
		if flagNoInteractive || flagJSON || flagNoColor {
			ui.SetSpinnerEnabled(false)
		}
		if !skipUpdateHooks(cmd) {
			updater.TriggerBackgroundCheck(flagWorkspace)
		}
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if skipUpdateHooks(cmd) {
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
	RootCmd.Use = config.BinaryName()
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, ui.Error("%v", err))
		os.Exit(1)
	}
}

func init() {
	binName := config.BinaryName()
	RootCmd.Use = binName
	RootCmd.Long = ui.Bold(ui.Cyan(fmt.Sprintf("%s CLI", strings.ToUpper(binName[:1])+binName[1:]))) + " - Command-line interface for the Beep platform."
	SetupHelp(RootCmd)

	cmdutil.WorkspaceHook = func() string {
		return flagWorkspace
	}

	RootCmd.PersistentFlags().BoolVar(&flagNoColor, "no-color", false, "Disable colored output")
	RootCmd.PersistentFlags().BoolVar(&flagNoInteractive, "no-interactive", false, "Disable interactive prompts")
	RootCmd.PersistentFlags().StringVarP(&flagWorkspace, "workspace", "w", "", fmt.Sprintf("Local job workspace directory (default %s, env: BEEP_WORKSPACE)", config.DefaultWorkspaceDisplay()))
	RootCmd.PersistentFlags().StringVarP(&flagServer, "server", "s", "", "Beep server URL (env: BEEP_SERVER)")
	RootCmd.PersistentFlags().StringVarP(&flagToken, "token", "t", "", "Runner authentication token (env: BEEP_RUNNER_TOKEN)")
	RootCmd.PersistentFlags().StringVarP(&flagAccount, "account", "a", "", "Account slug to operate on (defaults to config account_slug or personal account; env: BEEP_ACCOUNT)")
	RootCmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "Output results in JSON format")

	// Command groups
	RootCmd.AddGroup(
		&cobra.Group{ID: "core", Title: "Core Commands"},
		&cobra.Group{ID: "beeps", Title: "Beep Commands (Default Scope)"},
		&cobra.Group{ID: "service", Title: "Local Service Commands"},
		&cobra.Group{ID: "additional", Title: "Additional Commands"},
	)

	// 1. Core commands (auth, beep, beeper, config)
	authCmd.GroupID = "core"
	configCmd.GroupID = "core"
	beeperCmd := beeper.NewCmdBeeper()
	beeperCmd.GroupID = "core"
	legacyBeepCmd := beep.NewCmdBeep()
	legacyBeepCmd.GroupID = "core"
	legacyBeepCmd.Hidden = false

	RootCmd.AddCommand(authCmd)
	RootCmd.AddCommand(legacyBeepCmd)
	RootCmd.AddCommand(beeperCmd)
	RootCmd.AddCommand(configCmd)

	// 2. Beep subcommands (default scope - can be used directly without 'beep beep')
	beepCommands := []*cobra.Command{
		beep.NewCmdCreate(),
		beep.NewCmdDelete(),
		beep.NewCmdList(),
		beep.NewCmdPause(),
		beep.NewCmdResume(),
		beep.NewCmdRun(),
		beep.NewCmdShow(),
	}
	for _, cmd := range beepCommands {
		cmd.GroupID = "beeps"
		RootCmd.AddCommand(cmd)
	}

	// 3. Local service commands
	channelCmd.GroupID = "service"
	runnerCmd.GroupID = "service"
	serviceCmd := service.NewCmdService()
	serviceCmd.GroupID = "service"

	RootCmd.AddCommand(channelCmd)
	RootCmd.AddCommand(runnerCmd)
	RootCmd.AddCommand(serviceCmd)

	// 4. Additional commands
	upgradeCmd.GroupID = "additional"
	versionCmd.GroupID = "additional"

	RootCmd.AddCommand(upgradeCmd)
	RootCmd.AddCommand(versionCmd)

	initCompletionCmd()
}
