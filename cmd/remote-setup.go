package cmd

import (
	"github.com/spf13/cobra"
)

var remoteSetupCmd = &cobra.Command{
	Use:   "remote-setup",
	Short: "Setup procedure that runs inside the devcontainer",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

func init() {
	rootCmd.AddCommand(remoteSetupCmd)
}
