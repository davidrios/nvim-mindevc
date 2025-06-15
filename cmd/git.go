package cmd

import (
	"log/slog"

	"github.com/spf13/cobra"
)

var gitCmd = &cobra.Command{
	Use:   "git",
	Short: "Minimal git implementation",
}

func init() {
	RootCmd.AddCommand(gitCmd)

	gitCmd.PersistentFlags().BoolVarP(
		&verbose,
		"verbose", "v",
		false,
		"show debug messages")

	if verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}
}
