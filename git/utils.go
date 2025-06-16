package git

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// Create HEADs for remotes. This is an ugly hack to emulate something that the
// git client does and the go-git library doesn't.
//
// This is necessary because lazy.nvim depends on it
func CreateRemoteHeads(repoDir string, r *git.Repository) error {
	remotes, err := r.Remotes()
	if err != nil {
		return err
	}

	for _, remote := range remotes {
		refs, err := remote.List(&git.ListOptions{})
		if err != nil {
			return err
		}

		for _, ref := range refs {
			if ref.Name() == plumbing.HEAD {
				strs := ref.Strings()

				fp, err := os.Create(filepath.Join(repoDir, ".git", "refs", "remotes", remote.Config().Name, "HEAD"))
				if err != nil {
					return err
				}

				newRef := strings.ReplaceAll(strs[1], "heads", "remotes/"+remote.Config().Name) + "\n"
				if len(newRef) > 0 {
					_, err = fp.WriteString(newRef)
					if err != nil {
						return err
					}

					return nil
				}
			}
		}
	}

	return nil
}
