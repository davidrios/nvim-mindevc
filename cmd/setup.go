package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/davidrios/nvim-mindevc/config"
	"github.com/davidrios/nvim-mindevc/setup"
)

var cmdDevcontainer config.Devcontainer
var skipSelfBinary bool

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

		err = setup.Setup(cmdConfig, cmdDevcontainer, skipSelfBinary)
		if err != nil {
			log.Fatal("Error: ", err)
		}
	},
}

func init() {
	RootCmd.AddCommand(setupCmd)

	setupCmd.Flags().BoolVarP(
		&skipSelfBinary,
		"skip-self", "S",
		false,
		"Don't use self binary on remote container even if os/architecture matches")
}
