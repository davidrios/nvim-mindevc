package cmd

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
)

var gitShowVersion bool

var gitCmd = &cobra.Command{
	Use:   "git",
	Short: "Minimal git implementation",
	RunE: func(cmd *cobra.Command, args []string) error {
		if gitShowVersion {
			fmt.Println("git version 2.38.5")
			return nil
		}

		return fmt.Errorf("no command provided")
	},
}

func init() {
	RootCmd.AddCommand(gitCmd)

	gitCmd.PersistentFlags().BoolVarP(
		&verbose,
		"verbose", "v",
		false,
		"show debug messages")

	gitCmd.Flags().BoolVar(
		&gitShowVersion,
		"version",
		false,
		"")

	if verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}
}
