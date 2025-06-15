package git

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/go-git/go-git/v5"
)

type LsFilesOptions struct {
	Deleted  bool
	Modified bool
}

func LsFiles(repoDir string, options LsFilesOptions) error {
	abs, err := filepath.Abs(repoDir)
	if err != nil {
		return err
	}

	r, err := git.PlainOpen(abs)
	if err != nil {
		return err
	}

	tree, err := r.Worktree()
	if err != nil {
		return fmt.Errorf("worktree error %w", err)
	}

	status, err := tree.StatusWithOptions(git.StatusOptions{
		Strategy: git.Preload,
	})
	if err != nil {
		return fmt.Errorf("status error %w", err)
	}
	for file, fileStatus := range status {
		slog.Debug("file", "n", file, "s", rune(fileStatus.Worktree))
		if fileStatus.Worktree == git.Untracked {
			continue
		}
		if !options.Deleted && !options.Modified {
			fmt.Println(file)
		} else if (fileStatus.Worktree == git.Deleted && options.Deleted) || (fileStatus.Worktree == git.Modified && options.Modified) {
			fmt.Println(file)
		}
	}

	return nil
}
