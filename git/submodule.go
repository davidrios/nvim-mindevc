package git

import (
	"fmt"
	"log/slog"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
)

func SubmoduleStatusRecursive(r *git.Repository, subpath string, depth int) error {
	depth -= 1
	if depth < 0 {
		slog.Debug("maximum submodule recursion depth reached")
		return nil
	}

	worktree, err := r.Worktree()
	if err != nil {
		return err
	}

	submodules, err := worktree.Submodules()
	if err != nil {
		return err
	}

	for _, submodule := range submodules {
		status, err := submodule.Status()
		if err != nil {
			return err
		}

		statusLine := status.String()
		fmt.Printf("%s%s%s\n", statusLine[:42], subpath, statusLine[42:])

		sr, err := submodule.Repository()
		if err != nil {
			return err
		}

		subpath = subpath + submodule.Config().Path + "/"

		err = SubmoduleStatusRecursive(sr, subpath, depth)
		if err != nil {
			return err
		}
	}

	return nil
}

type SubmoduleStatusOptions struct {
	Recursive bool
}

func SubmoduleStatus(repoDir string, options SubmoduleStatusOptions) error {
	abs, err := filepath.Abs(repoDir)
	if err != nil {
		return err
	}

	r, err := git.PlainOpen(abs)
	if err != nil {
		return err
	}

	depth := 1
	if options.Recursive {
		depth = int(git.DefaultSubmoduleRecursionDepth)
	}

	return SubmoduleStatusRecursive(r, "", depth)
}

func SubmoduleInit(repoDir string) error {
	abs, err := filepath.Abs(repoDir)
	if err != nil {
		return err
	}

	r, err := git.PlainOpen(abs)
	if err != nil {
		return err
	}

	_, err = SubmoduleInitRepo(r)
	if err != nil {
		return err
	}

	return nil
}

func SubmoduleInitRepo(r *git.Repository) (git.Submodules, error) {
	worktree, err := r.Worktree()
	if err != nil {
		return nil, err
	}

	submodules, err := worktree.Submodules()
	if err != nil {
		return nil, err
	}

	if len(submodules) == 0 {
		return nil, nil
	}

	remotes, err := r.Remotes()
	if err != err {
		return nil, err
	}
	if len(remotes) == 0 {
		return nil, fmt.Errorf("no remotes found in repo")
	}

	remote := remotes[0]

	rconfig := remote.Config()
	if len(rconfig.URLs) == 0 {
		return nil, fmt.Errorf("no urls found")
	}

	rootUrl := rconfig.URLs[0]

	rurl, err := url.Parse(rootUrl)
	if err != nil {
		return nil, err
	}

	if rurl.Scheme != "http" && rurl.Scheme != "https" {
		return nil, fmt.Errorf("unsupported scheme for repo url %s", rurl.Scheme)
	}

	adjustUrls := map[string]string{}

	for _, submodule := range submodules {
		err := submodule.Init()
		if err != nil {
			if err == git.ErrSubmoduleAlreadyInitialized {
				continue
			}

			return nil, err
		}
		smurl := submodule.Config().URL
		if strings.Index(smurl, "../") == 0 {
			// go-git doesn't support this kind of url natively, so hack it together
			smurl = rurl.JoinPath(smurl).String()
			adjustUrls[submodule.Config().Name] = smurl
		}
		fmt.Printf("Submodule '%s' (%s) registered for path '%s'\n", submodule.Config().Name, smurl, submodule.Config().Path)
	}

	if len(adjustUrls) > 0 {
		cfg, err := r.Config()
		if err != nil {
			return nil, err
		}

		for mname, murl := range adjustUrls {
			subm := cfg.Submodules[mname]
			subm.URL = murl
		}

		err = r.SetConfig(cfg)
		if err != nil {
			return nil, err
		}
	}

	return submodules, nil
}

func SubmoduleUpdate(repoDir string) error {
	abs, err := filepath.Abs(repoDir)
	if err != nil {
		return err
	}
	r, err := git.PlainOpen(abs)
	if err != nil {
		return err
	}

	return SubmoduleUpdateRepo(r)
}

func SubmoduleUpdateRepo(r *git.Repository) error {
	worktree, err := r.Worktree()
	if err != nil {
		return err
	}

	submodules, err := worktree.Submodules()
	if err != nil {
		return err
	}

	return SubmoduleUpdateModules(submodules)
}

func SubmoduleUpdateModules(submodules git.Submodules) error {
	for _, submodule := range submodules {
		err := submodule.Update(&git.SubmoduleUpdateOptions{})
		if err != nil {
			return err
		}
		status, err := submodule.Status()
		if err != nil {
			return err
		}
		fmt.Printf("Submodule path '%s': checked out '%s'\n", submodule.Config().Path, status.Current.String())
	}

	return nil
}

func SubmoduleInitAndUpdateRecursive(r *git.Repository, depth int) error {
	depth -= 1
	if depth < 0 {
		slog.Debug("maximum submodule recursion depth reached")
		return nil
	}

	submodules, err := SubmoduleInitRepo(r)
	if err != nil {
		return err
	}

	err = SubmoduleUpdateRepo(r)
	if err != nil {
		return err
	}

	for _, submodule := range submodules {
		sr, err := submodule.Repository()
		if err != nil {
			return err
		}

		err = SubmoduleInitAndUpdateRecursive(sr, depth)
		if err != nil {
			return err
		}
	}

	return nil
}
