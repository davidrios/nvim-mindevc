package cmd

import (
	"log"
	"log/slog"

	"github.com/davidrios/nvim-mindevc/git"
	"github.com/spf13/cobra"
)

var gitSubmoduleStatusOptions git.SubmoduleStatusOptions

var gitSubmoduleCmd = &cobra.Command{
	Use: "submodule [command]",
	Run: func(cmd *cobra.Command, args []string) {
		slog.Debug("args", "a", args)

		if err := git.SubmoduleStatus(".", gitSubmoduleStatusOptions); err != nil {
			log.Fatal("error: ", err)
		}
	},
}

var gitSubmoduleStatusCmd = &cobra.Command{
	Use: "status",
	Run: func(cmd *cobra.Command, args []string) {
		slog.Debug("args", "a", args)

		if err := git.SubmoduleStatus(".", gitSubmoduleStatusOptions); err != nil {
			log.Fatal("error: ", err)
		}
	},
}

var gitSubmoduleInitCmd = &cobra.Command{
	Use: "init",
	Run: func(cmd *cobra.Command, args []string) {
		slog.Debug("args", "a", args)

		if err := git.SubmoduleInit("."); err != nil {
			log.Fatal("error: ", err)
		}
	},
}

var gitSubmoduleUpdateCmd = &cobra.Command{
	Use: "update",
	Run: func(cmd *cobra.Command, args []string) {
		slog.Debug("args", "a", args)

		if err := git.SubmoduleUpdate("."); err != nil {
			log.Fatal("error: ", err)
		}
	},
}

func init() {
	gitCmd.AddCommand(gitSubmoduleCmd)

	gitSubmoduleCmd.AddCommand(gitSubmoduleStatusCmd)
	gitSubmoduleStatusCmd.Flags().BoolVar(
		&gitSubmoduleStatusOptions.Recursive,
		"recursive",
		false,
		"")

	gitSubmoduleCmd.AddCommand(gitSubmoduleInitCmd)
	gitSubmoduleCmd.AddCommand(gitSubmoduleUpdateCmd)
}
