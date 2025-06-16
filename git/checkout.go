package git

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

type CheckoutOptions struct {
	Branch            string
	Progress          bool
	RecurseSubmodules bool
}

func Checkout(repoDir string, options CheckoutOptions) error {
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

	refIter, err := r.References()
	if err != nil {
		return err
	}

	var foundRef *plumbing.Reference
	err = refIter.ForEach(func(ref *plumbing.Reference) error {
		refName := ref.Name().String()
		lastIdx := strings.LastIndex(refName, options.Branch)
		if lastIdx == -1 {
			return nil
		}
		if lastIdx == len(refName)-len(options.Branch) {
			foundRef = ref
			return fmt.Errorf("stop iteration")
		}
		return nil
	})
	if err == nil || foundRef == nil {
		if foundRev, err := r.ResolveRevision(plumbing.Revision(options.Branch)); err == nil {
			err = tree.Checkout(&git.CheckoutOptions{Hash: *foundRev, Keep: true})
			if err == nil {
				return nil
			}
		}

		return fmt.Errorf("branch or commit not found: %s", options.Branch)
	}

	if strings.Index(options.Branch, "refs/") != 0 &&
		strings.Index(options.Branch, "tags/") != 0 &&
		strings.Index(foundRef.Name().String(), "refs/heads/") != 0 {
		err = tree.Checkout(&git.CheckoutOptions{
			Branch: plumbing.ReferenceName("refs/heads/" + options.Branch),
			Hash:   foundRef.Hash(),
			Create: true,
			Keep:   true,
		})
	} else {
		err = tree.Checkout(&git.CheckoutOptions{
			Branch: foundRef.Name(),
			Keep:   true,
		})
	}
	if err != nil {
		return fmt.Errorf("error checking out %w", err)
	}

	return nil
}
