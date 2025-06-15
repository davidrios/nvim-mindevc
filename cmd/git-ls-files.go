package cmd

import (
	"log/slog"

	"github.com/davidrios/nvim-mindevc/git"
	"github.com/spf13/cobra"
)

var gitLsFilesOptions git.LsFilesOptions

var gitLsFilesCmd = &cobra.Command{
	Use: "ls-files",
	RunE: func(cmd *cobra.Command, args []string) error {
		slog.Debug("args", "a", args)
		return git.LsFiles(".", gitLsFilesOptions)
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
