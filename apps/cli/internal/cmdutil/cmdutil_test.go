package cmdutil

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestIsInteractiveWithJSON(t *testing.T) {
	rootCmd := &cobra.Command{Use: "beep"}
	rootCmd.PersistentFlags().Bool("json", false, "")
	rootCmd.PersistentFlags().Bool("no-interactive", false, "")

	subCmd := &cobra.Command{Use: "sub"}
	rootCmd.AddCommand(subCmd)

	// When --json is true, IsInteractive must be false
	if err := rootCmd.PersistentFlags().Set("json", "true"); err != nil {
		t.Fatalf("failed to set json flag: %v", err)
	}

	if IsInteractive(subCmd) {
		t.Errorf("expected IsInteractive to be false when --json is active")
	}

	// Reset --json
	_ = rootCmd.PersistentFlags().Set("json", "false")
	_ = rootCmd.PersistentFlags().Set("no-interactive", "true")
	if IsInteractive(subCmd) {
		t.Errorf("expected IsInteractive to be false when --no-interactive is active")
	}
}
