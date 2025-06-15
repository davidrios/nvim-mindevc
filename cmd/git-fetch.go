package cmd

import (
	"log/slog"

	"github.com/davidrios/nvim-mindevc/git"
	"github.com/spf13/cobra"
)

var gitFetchOptions git.FetchOptions

var gitFetchCmd = &cobra.Command{
	Use: "fetch",
	RunE: func(cmd *cobra.Command, args []string) error {
		slog.Debug("args", "a", args)
		return git.Fetch(".", gitFetchOptions)
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
