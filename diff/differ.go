package diff

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"

	"git.dmann.xyz/davidmann/gitops-repo-api/entrypoint"
	"git.dmann.xyz/davidmann/gitops-repo-api/git"
	"git.dmann.xyz/davidmann/gitops-repo-api/resource"
	"github.com/go-git/go-git/v5/plumbing"
	"sigs.k8s.io/kustomize/api/resmap"
)

type EntrypointDiff struct {
	Entrypoint entrypoint.Entrypoint
	Error      error
	Diff       []resource.ResourceDiff
}

// Diff will return either an EntrypointDiff, or an Error for every Entrypoint that is discovered in the
// pre
func Diff(ctx context.Context, rs *git.RepoSpec, epds []entrypoint.EntrypointDiscoverySpec, pre, post *plumbing.Reference) ([]EntrypointDiff, error) {
	cleanup := true

	_, preDir, err := rs.Checkout(ctx, pre)
	if err != nil {
		return nil, fmt.Errorf("unable to pre change dir - %w", err)
	}

	_, postDir, err := rs.Checkout(ctx, post)
	if err != nil {
		return nil, fmt.Errorf("unable to checkout post change dir - %w", err)
	}

	defer func() {
		if !cleanup {
			fmt.Printf("\n\n=========================\n\nNot cleaning up\n%s\n%s\n\n=========================\n\n", preDir, postDir)
			return
		}

		var errs error
		if err := os.RemoveAll(preDir); err != nil {
			errs = errors.Join(errs, err)
		}
		if err := os.RemoveAll(postDir); err != nil {
			errs = errors.Join(errs, err)
		}
		if errs != nil {
			panic(errs.Error())
		}
	}()

	eps, err := discoverEntrypoints(ctx, preDir, postDir, epds)
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

			diff, err := diffEntrypoint(ctx, ep.ep, preDir, postDir)
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
	t  string
	ep entrypoint.Entrypoint
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

func diffEntrypoint(ctx context.Context, ep entrypoint.Entrypoint, preDir, postDir string) ([]resource.ResourceDiff, error) {
	ewg := sync.WaitGroup{}
	ewg.Add(1)
	var preResources *resmap.ResMap
	var postResources *resmap.ResMap
	var buildErrs error
	go func() {
		defer ewg.Done()
		pr, err := entrypoint.ExtractResources(preDir, ep)
		if err != nil {
			buildErrs = errors.Join(buildErrs, fmt.Errorf("unable to build pre-entrypoint - %w", err))
			return
		}
		preResources = &pr

	}()

	ewg.Add(1)
	go func() {
		defer ewg.Done()
		pr, err := entrypoint.ExtractResources(postDir, ep)
		if err != nil {
			buildErrs = errors.Join(buildErrs, fmt.Errorf("unable to build pre-entrypoint - %w", err))
			return
		}
		postResources = &pr
	}()

	ewg.Wait()
	if buildErrs != nil {
		return nil, buildErrs
	}

	diff, err := resource.Diff(ctx, *preResources, *postResources)
	if err != nil {
		return nil, err
	}

	return diff, nil
}
