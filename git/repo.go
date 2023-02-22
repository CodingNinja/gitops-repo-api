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

func NewRepoSpec(url string, credentials transport.AuthMethod) *RepoSpec {
	return &RepoSpec{
		URL:         url,
		Credentials: credentials,
	}
}

type RepoSpec struct {
	URL         string
	Credentials transport.AuthMethod
	Progress    io.Writer
	repo        *git.Repository
	l           sync.Mutex
}

func (rs *RepoSpec) Name() string {
	return slug.Make(rs.URL)
}

func (rs *RepoSpec) CloneDirectory(branch string) string {
	return path.Join(os.TempDir(), rs.Name(), branch)
}

func (rs *RepoSpec) Open(ctx context.Context) (*git.Repository, error) {
	rs.l.Lock()
	defer rs.l.Unlock()
	if rs.repo != nil {
		return rs.repo, nil
	}
	directory := rs.CloneDirectory(".root")

	r, err := cloneRepo(ctx, directory, true, git.CloneOptions{
		URL:      rs.URL,
		Auth:     rs.Credentials,
		Progress: rs.Progress,
	})

	if err != nil {
		return nil, fmt.Errorf("unable to clone repo - %w", err)
	}

	rs.repo = r

	return rs.repo, nil
}

func (rs *RepoSpec) Checkout(ctx context.Context, reference *plumbing.Reference) (*git.Repository, string, error) {
	repo, err := rs.Open(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("error opening repo %q - %w", rs.URL, err)
	}
	// If the thing we want to checkout is a ref, first update that ref
	// in the local base repo to match the latest fetched origin hash
	if reference.Type() == plumbing.SymbolicReference {
		cur, err := repo.ResolveRevision(plumbing.Revision(plumbing.NewRemoteReferenceName("origin", reference.Name().Short())))
		if err != nil {
			return nil, "", fmt.Errorf("error opening repo %q - %w", rs.URL, err)
		}

		if err := repo.Storer.SetReference(plumbing.NewHashReference(reference.Name(), *cur)); err != nil {
			return nil, "", err
		}
	}

	rootDirectory := rs.CloneDirectory(".root")
	// We clone the root directory to enable multiple concurrent bulids of the same
	// repo without killing the upstream git repo
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
