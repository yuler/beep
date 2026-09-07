package cmd

import (
	"fmt"

	"beep/internal/ui"
	"beep/internal/version"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s version %s (%s, %s)\n",
			ui.Bold(ui.Cyan("beep")),
			ui.Bold(version.Version),
			ui.Dim(version.GitCommit),
			ui.Dim(version.BuildDate),
		)
	},
}
