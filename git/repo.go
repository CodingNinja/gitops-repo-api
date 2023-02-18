package git

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"sync"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/google/uuid"
	"github.com/gosimple/slug"
)

func NewRepoSpec(url string, credentials transport.AuthMethod) *repoSpec {
	return &repoSpec{
		URL:         url,
		Credentials: credentials,
		Progress:    os.Stdout,
	}
}

type repoSpec struct {
	URL         string
	Credentials transport.AuthMethod
	Progress    io.Writer
	repo        *git.Repository
	l           sync.Mutex
}

func (rs *repoSpec) Name() string {
	return slug.Make(rs.URL)
}

func (rs *repoSpec) CloneDirectory(branch string) string {
	return path.Join(os.TempDir(), rs.Name(), branch)
}

func (rs *repoSpec) Open(ctx context.Context) (*git.Repository, error) {
	rs.l.Lock()
	defer rs.l.Unlock()
	if rs.repo != nil {
		return rs.repo, nil
	}
	directory := rs.CloneDirectory(".root")

	r, err := cloneRepo(ctx, directory, true, git.CloneOptions{
		URL:               rs.URL,
		Auth:              rs.Credentials,
		NoCheckout:        true,
		Progress:          rs.Progress,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	})

	if err != nil {
		return nil, fmt.Errorf("unable to clone repo - %w", err)
	}

	rs.repo = r

	return rs.repo, nil
}

func (rs *repoSpec) Checkout(ctx context.Context, reference *plumbing.Reference) (*git.Repository, string, error) {
	_, err := rs.Open(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("error opening repo %q - %w", rs.URL, err)
	}

	rootDirectory := rs.CloneDirectory(".root")
	dirName := reference.Target().Short()
	if dirName == "" {
		dirName = reference.Hash().String()
	}
	branchDirectory := path.Join(rs.CloneDirectory(dirName), uuid.New().String())

	branchRepo, err := cloneRepo(ctx, branchDirectory, false, git.CloneOptions{
		URL:               rootDirectory,
		NoCheckout:        false,
		Progress:          rs.Progress,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	})

	if err != nil {
		return nil, "", fmt.Errorf("unable to create branch repo - %w", err)
	}

	wt, err := branchRepo.Worktree()
	if err != nil {
		return nil, "", fmt.Errorf("unable to get worktree for branch repo - %w", err)
	}

	if reference.Type() == plumbing.HashReference {
		err := wt.Checkout(&git.CheckoutOptions{
			Hash: reference.Hash(),
		})
		if err != nil {
			return nil, "", fmt.Errorf("unable to checkout reference %q for branch repo - %w", reference.String(), err)
		}
	} else {
		err := wt.Checkout(&git.CheckoutOptions{
			Branch: reference.Name(),
			Force:  true,
		})
		if err != nil {
			return nil, "", fmt.Errorf("unable to checkout reference %q for branch repo - %w", reference.String(), err)
		}
	}

	return branchRepo, branchDirectory, nil
}
