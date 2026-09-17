package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"beep/internal/cmdutil"
	"beep/internal/config"
	"beep/internal/daemon"
	"beep/internal/ui"
	"beep/internal/updater"
	"beep/internal/version"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var (
	flagUpgradeCheck   bool
	flagUpgradeVersion string
	flagUpgradeForce   bool
	flagUpgradeYes     bool
)

func newUpgradeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "upgrade",
		Aliases: []string{"update"},
		Short:   fmt.Sprintf("Upgrade %s CLI to the latest version or a specific version", config.BinaryName()),
		Long: ui.Bold(ui.Cyan(fmt.Sprintf("%s upgrade", config.BinaryName()))) + fmt.Sprintf(` - Self-update the %s CLI binary in-place.

Checks GitHub Releases for new versions, downloads the platform-specific
binary archive, verifies the SHA-256 checksum, and replaces the running executable.`, config.BinaryName()),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpgrade(cmd, args)
		},
	}

	cmd.Flags().BoolVarP(&flagUpgradeCheck, "check", "c", false, "Check for available updates without installing")
	cmd.Flags().StringVar(&flagUpgradeVersion, "version", "", "Target release version or tag to install (e.g. v0.2.2)")
	cmd.Flags().BoolVarP(&flagUpgradeForce, "force", "f", false, "Force reinstall even if already on the target version")
	cmd.Flags().BoolVarP(&flagUpgradeYes, "yes", "y", false, "Automatically confirm upgrade without interactive prompt")

	return cmd
}

var upgradeCmd = newUpgradeCmd()

func runUpgrade(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	repo := updater.GetRepo()
	targetVer := strings.TrimSpace(flagUpgradeVersion)

	// --check mode
	if flagUpgradeCheck {
		fmt.Printf("%s Checking for updates (%s)...\n", ui.Cyan("●"), repo)
		var rel *updater.ReleaseInfo
		var err error
		if targetVer == "" || strings.EqualFold(targetVer, "latest") {
			rel, err = updater.FetchLatestRelease(ctx, repo)
		} else {
			rel, err = updater.FetchReleaseByTag(ctx, repo, targetVer)
		}
		if err != nil {
			return fmt.Errorf("failed to check release: %w", err)
		}

		cmp := updater.CompareVersions(rel.Tag, version.Version)
		if cmp > 0 {
			fmt.Println()
			fmt.Printf("%s %s\n", ui.Yellow("!"), ui.Bold(fmt.Sprintf("A new release of %s is available: %s → %s", config.BinaryName(), ui.Dim(version.Version), ui.Green(rel.Tag))))
			if rel.HTMLURL != "" {
				fmt.Printf("  %s %s\n", ui.Dim("Release notes:"), ui.Cyan(rel.HTMLURL))
			}
			fmt.Printf("  %s Run %s to update.\n", ui.Dim("Action:       "), ui.Bold(ui.Cyan(fmt.Sprintf("%s upgrade", config.BinaryName()))))
		} else if cmp == 0 {
			fmt.Printf("%s %s is already up to date (%s)\n", ui.Green("✓"), config.BinaryName(), ui.Bold(version.Version))
		} else {
			fmt.Printf("%s Current version (%s) is newer than release (%s)\n", ui.Cyan("ℹ"), version.Version, rel.Tag)
		}
		return nil
	}

	// Upgrade mode: first query release info
	fmt.Printf("%s Resolving release version (%s)...\n", ui.Cyan("●"), repo)
	var rel *updater.ReleaseInfo
	var err error
	if targetVer == "" || strings.EqualFold(targetVer, "latest") {
		rel, err = updater.FetchLatestRelease(ctx, repo)
	} else {
		rel, err = updater.FetchReleaseByTag(ctx, repo, targetVer)
	}
	if err != nil {
		return fmt.Errorf("failed to query release: %w", err)
	}

	cmp := updater.CompareVersions(rel.Tag, version.Version)
	if cmp == 0 && !flagUpgradeForce {
		fmt.Printf("%s %s is already at the target version %s (%s)\n", ui.Green("✓"), config.BinaryName(), ui.Bold(rel.Tag), ui.Dim("use --force to reinstall"))
		return nil
	}

	// Interactive confirmation if running in a TTY and not auto-approved
	if ui.IsInteractive() && !flagUpgradeYes && !flagNoInteractive {
		var confirm bool
		actionDesc := fmt.Sprintf("Upgrade %s from %s to %s?", config.BinaryName(), version.Version, rel.Tag)
		if cmp < 0 {
			actionDesc = fmt.Sprintf("Downgrade %s from %s to %s?", config.BinaryName(), version.Version, rel.Tag)
		} else if cmp == 0 {
			actionDesc = fmt.Sprintf("Reinstall %s version %s?", config.BinaryName(), rel.Tag)
		}

		form := huh.NewConfirm().
			Title(actionDesc).
			Description(fmt.Sprintf("Release: %s", rel.HTMLURL)).
			Value(&confirm)

		if err := form.Run(); err != nil {
			return err
		}
		if !confirm {
			fmt.Println(ui.Dim("Upgrade canceled."))
			return nil
		}
	}

	// Load config to check daemon and workspace
	cfg, _ := loadUpgradeConfig()
	ws := ""
	if cfg != nil {
		ws = cfg.Workspace
	}

	// Execute upgrade
	res, err := updater.Upgrade(ctx, updater.UpgradeOptions{
		TargetVersion: rel.Tag,
		Force:         flagUpgradeForce,
		Repo:          repo,
		Workspace:     ws,
		Release:       rel,
		OnProgress: func(stage string) {
			fmt.Printf("%s %s\n", ui.Cyan("●"), stage)
		},
	})
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Printf("%s %s (%s)\n",
		ui.Green("✓"),
		ui.Bold(fmt.Sprintf("Successfully upgraded %s to %s", config.BinaryName(), res.NewVersion)),
		ui.Dim(res.ExecutablePath),
	)

	// Check if any daemon is running in this workspace
	if ws != "" {
		runnerRun, runnerPid, _ := daemon.CheckRunning(ws, daemon.ServiceRunner)
		channelRun, channelPid, _ := daemon.CheckRunning(ws, daemon.ServiceChannel)

		if runnerRun || channelRun {
			pidList := []string{}
			if runnerRun {
				pidList = append(pidList, fmt.Sprintf("runner PID %d", runnerPid))
			}
			if channelRun {
				pidList = append(pidList, fmt.Sprintf("channel PID %d", channelPid))
			}
			fmt.Println()
			fmt.Printf("%s %s (%s)\n",
				ui.Yellow("⚠"),
				ui.Bold("A background daemon is currently running:"),
				strings.Join(pidList, ", "),
			)
			fmt.Printf("  Run %s or restart the process to apply the new version.\n", ui.Cyan(fmt.Sprintf("%s up -d", config.BinaryName())))
		}
	}

	return nil
}

// loadUpgradeConfig loads config for the upgrade pre-check only (daemon/workspace).
func loadUpgradeConfig() (*config.Config, error) {
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
