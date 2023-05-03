package diff

import (
	"context"
	"errors"
	"fmt"
	"path"
	"sync"

	"github.com/codingninja/gitops-repo-api/entrypoint"
	"github.com/codingninja/gitops-repo-api/git"
	"github.com/codingninja/gitops-repo-api/resource"
	"github.com/go-git/go-git/v5/plumbing"
)

func NewDiffer(preRs *git.RepoSpec, postRs *git.RepoSpec, epds []entrypoint.EntrypointDiscoverySpec) *repoDiffer {
	return &repoDiffer{
		preRs:  preRs,
		postRs: postRs,
		epds:   epds,
	}
}

type repoDiffer struct {
	preRs  *git.RepoSpec
	postRs *git.RepoSpec
	epds   []entrypoint.EntrypointDiscoverySpec
}

type EntrypointDiff struct {
	Entrypoint entrypoint.Entrypoint   `json:"entrypoint"`
	Error      error                   `json:"error"`
	Diff       []resource.ResourceDiff `json:"diff"`
}

// Diff will return either an EntrypointDiff, or an Error for every Entrypoint that is discovered in the
// pre
func (rd *repoDiffer) Diff(ctx context.Context, pre, post plumbing.ReferenceName) ([]EntrypointDiff, error) {
	_, preDir, err := rd.preRs.Checkout(ctx, pre)
	if err != nil {
		return nil, fmt.Errorf("unable to pre change dir - %w", err)
	}

	_, postDir, err := rd.postRs.Checkout(ctx, post)
	if err != nil {
		return nil, fmt.Errorf("unable to checkout post change dir - %w", err)
	}

	defer func() {
		// var errs error
		// if err := os.RemoveAll(preDir); err != nil {
		// 	errs = errors.Join(errs, err)
		// }
		// if err := os.RemoveAll(postDir); err != nil {
		// 	errs = errors.Join(errs, err)
		// }
		// if errs != nil {
		// 	panic(errs.Error())
		// }
	}()

	eps, err := discoverEntrypoints(ctx, preDir, postDir, rd.epds)
	if err != nil {
		return nil, err
	}

	var errs error
	allDiff := []EntrypointDiff{}
	wg := sync.WaitGroup{}
	for _, ep := range eps {
		ep := ep
		wg.Add(1)
		go func() {
			defer wg.Done()

			diff, err := rd.diffEntrypoint(ctx, ep.ep, preDir, postDir)
			if err != nil {
				errs = errors.Join(errs, err)
			}

			allDiff = append(allDiff, EntrypointDiff{
				Entrypoint: ep.ep,
				Diff:       diff,
				Error:      err,
			})
		}()
	}

	wg.Wait()

	return allDiff, errs
}

type internalentrypoint struct {
	t      string
	ep     entrypoint.Entrypoint
	hash   plumbing.Hash
	branch plumbing.ReferenceName
}

func discoverEntrypoints(ctx context.Context, preDir, postDir string, epds []entrypoint.EntrypointDiscoverySpec) ([]internalentrypoint, error) {
	// This should be re-implemented to use channels
	preEps, err := entrypoint.DiscoverEntrypoints(preDir, epds)

	if err != nil {
		return nil, err
	}
	postEps, err := entrypoint.DiscoverEntrypoints(postDir, epds)

	if err != nil {
		return nil, err
	}
	eps := map[string]bool{}
	eplist := []internalentrypoint{}
	for _, ep := range preEps {
		eps[ep.Directory] = true
		eplist = append(eplist, internalentrypoint{t: "existing", ep: ep})
	}
	for _, ep := range postEps {
		if _, ok := eps[ep.Directory]; ok {
			continue
		}
		eps[ep.Directory] = true
		eplist = append(eplist, internalentrypoint{t: "new", ep: ep})
	}

	return eplist, nil
}

func (rd *repoDiffer) diffEntrypoint(ctx context.Context, ep entrypoint.Entrypoint, preDir, postDir string) ([]resource.ResourceDiff, error) {
	differ, err := resource.EntrypointDiffer(ep)
	if err != nil {
		return nil, fmt.Errorf("unable to get differ for entrypoint - %w", err)
	}

	diff, err := differ.Diff(ctx, rd.preRs, ep, path.Join(preDir, ep.Directory), path.Join(postDir, ep.Directory))
	if err != nil {
		return nil, fmt.Errorf("unable to extract entrypoint diff - %w", err)
	}

	return diff, nil
}
