package git

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/go-git/go-git/v5"
)

func cloneRepo(ctx context.Context, directory string, isBare bool, opts git.CloneOptions) (*git.Repository, error) {
	err := os.MkdirAll(path.Dir(directory), 0700)
	if err != nil {
		return nil, fmt.Errorf("unable to create temp dir for repo - %w", err)
	}
	var r *git.Repository
	if file, err := os.Stat(directory); err == nil && file.IsDir() {
		var err error
		r, err = git.PlainOpenWithOptions(directory, &git.PlainOpenOptions{})
		if err != nil {
			if err := os.RemoveAll(directory); err != nil {
				return nil, fmt.Errorf("unable to cleanup bad cache dir - %w", err)
			}
		} else {
			err = r.FetchContext(ctx, &git.FetchOptions{
				Auth:     opts.Auth,
				Progress: opts.Progress,
				Force:    true,
			})
			if err != nil && err != git.NoErrAlreadyUpToDate {
				return nil, fmt.Errorf("unable to fetch latest changes for %q from %q - %w", directory, opts.URL, err)
			}
		}
	}

	if r == nil {
		var err error
		r, err = git.PlainCloneContext(ctx, directory, isBare, &opts)
		if err != nil {
			return nil, fmt.Errorf("unable to clone repo %q to %q - %w", opts.URL, directory, err)
		}
	}

	return r, nil
}
