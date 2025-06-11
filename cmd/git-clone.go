package cmd

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
)

var cloneFilter string

var gitCloneCmd = &cobra.Command{
	Use:  "clone <repository> [<directory>]",
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		slog.Debug("args", "a", args)

		var directory string
		if len(args) == 1 {
			directory = strings.Replace(filepath.Base(args[0]), ".git", "", -1)
		} else {
			directory = args[1]
		}

		options := git.CloneOptions{
			URL:      args[0],
			Progress: os.Stdout,
		}
		// if cloneFilter != "" {
		// 	options.Filter = cloneFilter
		// }

		slog.Info("cloning repository, please wait...", "target_dir", directory)
		_, err := git.PlainClone(directory, false, &options)
		if err != nil {
			return err
		}

		slog.Info("done cloning")
		return nil
	},
}

func init() {
	gitCmd.AddCommand(gitCloneCmd)

	gitCloneCmd.Flags().StringVar(
		&cloneFilter,
		"filter",
		"",
		"")
}
