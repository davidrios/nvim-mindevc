package cmd

import (
	"log/slog"

	"github.com/davidrios/nvim-mindevc/git"
	"github.com/spf13/cobra"
)

var gitCheckoutOptions git.CheckoutOptions

var gitCheckoutCmd = &cobra.Command{
	Use:  "checkout [<branch> | <commit>]",
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		slog.Debug("args", "a", args)

		gitCheckoutOptions.Branch = args[0]

		return git.Checkout(".", gitCheckoutOptions)
	},
}

func init() {
	gitCmd.AddCommand(gitCheckoutCmd)

	gitCheckoutCmd.Flags().BoolVar(
		&gitCheckoutOptions.RecurseSubmodules,
		"recurse-submodules",
		false,
		"")

	gitCheckoutCmd.Flags().BoolVar(
		&gitCheckoutOptions.Progress,
		"progress",
		false,
		"")
}
