package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/davidrios/nvim-mindevc/config"
	"github.com/davidrios/nvim-mindevc/setup"
)

var cmdDevcontainer config.Devcontainer

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup neovim inside devcontainer",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		var devcontainerFileLoc = devcontainerFile
		if devcontainerFileLoc == "" {
			devcontainerFileLoc = cmdConfig.Config.GetDevcontainerFilePath()
		}

		cmdDevcontainer, err = config.LoadDevcontainer(devcontainerFileLoc)
		if err != nil {
			log.Fatal("Error loading dev container: ", err)
		}

		err = setup.Setup(cmdConfig, cmdDevcontainer)
		if err != nil {
			log.Fatal("Error: ", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)
}
