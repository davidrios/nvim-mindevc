package cmd

import (
	"log"
	"log/slog"

	"github.com/davidrios/nvim-mindevc/git"
	"github.com/spf13/cobra"
)

var gitLsFilesOptions git.LsFilesOptions

var gitLsFilesCmd = &cobra.Command{
	Use: "ls-files",
	Run: func(cmd *cobra.Command, args []string) {
		slog.Debug("args", "a", args)
		if err := git.LsFiles(".", gitLsFilesOptions); err != nil {
			log.Fatal("error: ", err)
		}
	},
}

func init() {
	gitCmd.AddCommand(gitLsFilesCmd)

	gitLsFilesCmd.Flags().BoolVarP(
		&gitLsFilesOptions.Deleted,
		"deleted", "d",
		false,
		"")

	gitLsFilesCmd.Flags().BoolVarP(
		&gitLsFilesOptions.Modified,
		"modified", "m",
		false,
		"")
}
