package cmd

import (
	"log"
	"log/slog"

	"github.com/davidrios/nvim-mindevc/git"
	"github.com/spf13/cobra"
)

var gitCloneOptions git.CloneOptions

var gitCloneCmd = &cobra.Command{
	Use:  "clone <repository> [<directory>]",
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		slog.Debug("args", "a", args)

		gitCloneOptions.Url = args[0]
		if len(args) > 1 {
			gitCloneOptions.Directory = args[1]
		}

		if err := git.Clone(gitCloneOptions); err != nil {
			log.Fatal("error: ", err)
		}
	},
}

func init() {
	gitCmd.AddCommand(gitCloneCmd)

	gitCloneCmd.Flags().StringVar(
		&gitCloneOptions.Filter,
		"filter",
		"",
		"")

	gitCloneCmd.Flags().StringVarP(
		&gitCloneOptions.Branch,
		"branch", "b",
		"",
		"")

	gitCloneCmd.Flags().StringVar(
		&gitCloneOptions.Origin,
		"origin",
		"",
		"")

	gitCloneCmd.Flags().StringVarP(
		&gitCloneOptions.Config,
		"config", "c",
		"",
		"")

	gitCloneCmd.Flags().BoolVar(
		&gitCloneOptions.Progress,
		"progress",
		false,
		"")

	gitCloneCmd.Flags().BoolVar(
		&gitCloneOptions.RecurseSubmodules,
		"recurse-submodules",
		false,
		"")
}
