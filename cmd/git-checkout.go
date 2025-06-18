package cmd

import (
	"log"
	"log/slog"

	"github.com/davidrios/nvim-mindevc/git"
	"github.com/spf13/cobra"
)

var gitCheckoutOptions git.CheckoutOptions

var gitCheckoutCmd = &cobra.Command{
	Use:  "checkout [<branch> | <commit>]",
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		slog.Debug("args", "a", args)

		gitCheckoutOptions.Branch = args[0]

		if err := git.Checkout(".", gitCheckoutOptions); err != nil {
			log.Fatal("error: ", err)
		}
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
