package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/davidrios/nvim-mindevc/setup"
)

var remoteSetupCmd = &cobra.Command{
	Use:   "remote-setup",
	Short: "Setup procedure that runs inside the devcontainer",
	Run: func(cmd *cobra.Command, args []string) {
		err := setup.RemoteSetup(cmdConfig)
		if err != nil {
			log.Fatal("Error: ", err)
		}
	},
}

func init() {
	RootCmd.AddCommand(remoteSetupCmd)
}
