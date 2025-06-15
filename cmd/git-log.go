package cmd

import (
	"log/slog"

	"github.com/davidrios/nvim-mindevc/git"
	"github.com/spf13/cobra"
)

var gitLogOptions git.GitLogOptions

var gitLogCmd = &cobra.Command{
	Use: "log [<revision-range>]",
	RunE: func(cmd *cobra.Command, args []string) error {
		slog.Debug("args", "a", args)
		var revRange string
		if len(args) > 0 {
			revRange = args[0]
		}
		return git.PrintLog(".", revRange, gitLogOptions)
	},
}

func init() {
	gitCmd.AddCommand(gitLogCmd)

	gitLogCmd.Flags().StringVar(
		&gitLogOptions.Pretty,
		"pretty",
		"",
		"")

	gitLogCmd.Flags().StringVar(
		&gitLogOptions.Date,
		"date",
		"",
		"")

	gitLogCmd.Flags().StringVar(
		&gitLogOptions.Color,
		"color",
		"",
		"")

	gitLogCmd.Flags().BoolVar(
		&gitLogOptions.AbbrevCommit,
		"abbrev-commit",
		false,
		"")

	gitLogCmd.Flags().BoolVar(
		&gitLogOptions.Decorate,
		"decorate",
		false,
		"")

	gitLogCmd.Flags().BoolVar(
		&gitLogOptions.NoShowSignature,
		"no-show-signature",
		false,
		"")
}
