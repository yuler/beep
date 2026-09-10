package cmd

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"beep/internal/client"
	"beep/internal/config"
	"beep/internal/ui"

	"github.com/spf13/cobra"
)

var (
	flagChannelKind string
	flagAccountSlug string
	flagAuthToken   string
)

var channelCmd = &cobra.Command{
	Use:     "channel",
	Aliases: []string{"channels"},
	Short:   "Manage notification channels and device tokens",
}

var channelListCmd = &cobra.Command{
	Use:   "list",
	Short: "List registered notification channels",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		accountSlug := flagAccountSlug
		if accountSlug == "" {
			accountSlug = os.Getenv("BEEP_ACCOUNT")
		}
		if accountSlug == "" {
			return fmt.Errorf("account slug is required (set via --account or BEEP_ACCOUNT)")
		}

		authToken := flagAuthToken
		if authToken == "" {
			authToken = os.Getenv("BEEP_AUTH_TOKEN")
		}

		c := client.New(cfg)
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		channels, err := c.ListChannels(ctx, accountSlug, authToken)
		if err != nil {
			return err
		}

		if len(channels) == 0 {
			fmt.Println(ui.Dim("No notification channels found."))
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			ui.Bold("NAME"),
			ui.Bold("KIND"),
			ui.Bold("STATUS"),
			ui.Bold("TOKEN"),
			ui.Bold("LAST SEEN"),
		)

		for _, ch := range channels {
			lastSeen := "(never)"
			if ch.LastSeenAt != nil {
				lastSeen = ch.LastSeenAt.Format("2006-01-02 15:04:05")
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				ui.Bold(ch.Name),
				ch.Kind,
				ch.Status,
				ui.Dim(ch.MaskedToken),
				ui.Dim(lastSeen),
			)
		}
		return w.Flush()
	},
}

var channelCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new notification channel (default: device)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		accountSlug := flagAccountSlug
		if accountSlug == "" {
			accountSlug = os.Getenv("BEEP_ACCOUNT")
		}
		if accountSlug == "" {
			return fmt.Errorf("account slug is required (set via --account or BEEP_ACCOUNT)")
		}

		authToken := flagAuthToken
		if authToken == "" {
			authToken = os.Getenv("BEEP_AUTH_TOKEN")
		}

		c := client.New(cfg)
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		ch, err := c.CreateChannel(ctx, accountSlug, authToken, name, flagChannelKind)
		if err != nil {
			return err
		}

		fmt.Printf("%s Created channel %s (%s)\n",
			ui.Green("✓"),
			ui.Bold(ch.Name),
			ui.Cyan(ch.Kind),
		)
		if ch.Token != "" {
			fmt.Printf("  %s %s\n", ui.Dim("Device Token:"), ui.Yellow(ch.Token))
			fmt.Printf("  %s %s\n", ui.Dim("To configure on this machine:"), ui.Cyan(fmt.Sprintf("beep config set device_token %s", ch.Token)))
		}
		return nil
	},
}

var channelSetTokenCmd = &cobra.Command{
	Use:   "set-token <token>",
	Short: "Configure device token for local machine in ~/.beep/config.json",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		token := args[0]
		configPath := config.GetConfigPath(flagWorkspace)
		fc, err := config.LoadFile(configPath)
		if err != nil {
			fc = &config.FileConfig{}
		}

		fc.DeviceToken = token
		if err := config.SaveFile(configPath, fc); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Println(ui.Success("Configured local device token in %s", ui.Dim(configPath)))
		return nil
	},
}

func init() {
	channelCreateCmd.Flags().StringVar(&flagChannelKind, "kind", "device", "Channel kind (device, webhook, email)")
	channelCmd.PersistentFlags().StringVar(&flagAccountSlug, "account", "", "Account slug")
	channelCmd.PersistentFlags().StringVar(&flagAuthToken, "auth-token", "", "User authentication token")

	channelCmd.AddCommand(channelListCmd)
	channelCmd.AddCommand(channelCreateCmd)
	channelCmd.AddCommand(channelSetTokenCmd)
}
