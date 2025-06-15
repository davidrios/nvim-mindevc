package git

import (
	"path/filepath"

	"github.com/go-git/go-git/v5"
)

type FetchOptions struct {
	RecurseSubmodules bool
	Tags              bool
	Force             bool
	Progress          bool
}

func Fetch(repoDir string, options FetchOptions) error {
	abs, err := filepath.Abs(repoDir)
	if err != nil {
		return err
	}

	r, err := git.PlainOpen(abs)
	if err != nil {
		return err
	}

	var fetchOptions git.FetchOptions
	if options.Tags {
		fetchOptions.Tags = git.AllTags
	}
	fetchOptions.Force = options.Force

	if options.RecurseSubmodules {
	}

	err = r.Fetch(&fetchOptions)
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return err
	}

	return nil
}
