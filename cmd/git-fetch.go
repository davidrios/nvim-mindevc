package cmd

import (
	"log"
	"log/slog"

	"github.com/davidrios/nvim-mindevc/git"
	"github.com/spf13/cobra"
)

var gitFetchOptions git.FetchOptions

var gitFetchCmd = &cobra.Command{
	Use: "fetch",
	Run: func(cmd *cobra.Command, args []string) {
		slog.Debug("args", "a", args)
		if err := git.Fetch(".", gitFetchOptions); err != nil {
			log.Fatal("error: ", err)
		}
	},
}

func init() {
	gitCmd.AddCommand(gitFetchCmd)

	gitFetchCmd.Flags().BoolVar(
		&gitFetchOptions.Force,
		"force",
		false,
		"")

	gitFetchCmd.Flags().BoolVar(
		&gitFetchOptions.Progress,
		"progress",
		false,
		"")

	gitFetchCmd.Flags().BoolVar(
		&gitFetchOptions.Tags,
		"tags",
		false,
		"")

	gitFetchCmd.Flags().BoolVar(
		&gitFetchOptions.RecurseSubmodules,
		"recurse-submodules",
		false,
		"")
}
