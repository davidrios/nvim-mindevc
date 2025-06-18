package git

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

type CloneOptions struct {
	Directory         string
	Url               string
	Branch            string
	Origin            string
	RecurseSubmodules bool
	Config            string
	Filter            string
	Progress          bool
}

func Clone(options CloneOptions) error {
	var directory string
	if options.Directory == "" {
		directory = strings.Replace(filepath.Base(options.Url), ".git", "", -1)
	} else {
		directory = options.Directory
	}

	cloneOptions := git.CloneOptions{
		URL:      options.Url,
		Progress: os.Stderr,
	}
	if options.Branch != "" {
		cloneOptions.ReferenceName = plumbing.ReferenceName(options.Branch)
	}
	if options.Origin != "" {
		cloneOptions.RemoteName = options.Origin
	}
	if options.Filter != "" {
		slog.Debug("filter option not working yet")
		// cloneOptions.Filter = packp.Filter(options.Filter)
		// not working yet
	}

	slog.Info("cloning repository, please wait...", "target_dir", directory)
	r, err := git.PlainClone(directory, false, &cloneOptions)
	if err != nil {
		return err
	}

	err = CreateRemoteHeads(directory, r)
	if err != nil {
		return err
	}

	if options.RecurseSubmodules {
		err = SubmoduleInitAndUpdateRecursive(r, int(git.DefaultSubmoduleRecursionDepth))
		if err != nil {
			return err
		}
	}

	slog.Info("done cloning")

	return nil
}
