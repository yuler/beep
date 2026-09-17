package cmd

import (
	cmdchannel "beep/cmd/channel"
	"beep/internal/cmdutil"

	"github.com/spf13/cobra"
)

var (
	channelCmd        = cmdchannel.NewCmdChannel()
	channelConnectCmd = &cobra.Command{
		Use:   "connect",
		Short: "Connect this CLI as a notification channel via Web browser (RFC 8628)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if flagWorkspace != "" {
				cmdutil.SetOverrideWorkspace(flagWorkspace)
				defer cmdutil.SetOverrideWorkspace("")
			}
			return cmdchannel.NewCmdConnect().RunE(cmd, args)
		},
	}
	channelDisconnectCmd = cmdchannel.NewCmdDisconnect()
)

func newChannelUpCmd() *cobra.Command {
	return cmdchannel.NewCmdUp()
}

func newChannelStopCmd() *cobra.Command {
	return cmdchannel.NewCmdStop()
}

func newChannelStatusCmd() *cobra.Command {
	return cmdchannel.NewCmdStatus()
}
