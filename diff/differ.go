package diff

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"sync"

	"git.dmann.xyz/davidmann/gitops-repo-api/entrypoint"
	"git.dmann.xyz/davidmann/gitops-repo-api/git"
	"git.dmann.xyz/davidmann/gitops-repo-api/resource"
	"github.com/go-git/go-git/v5/plumbing"
	"sigs.k8s.io/kustomize/api/resmap"
)

type EntrypointDiff struct {
	Entrypoint entrypoint.Entrypoint
	Diff       []resource.ResourceDiff
}

func Diff(ctx context.Context, rs *git.RepoSpec, pre, post *plumbing.Reference) ([]EntrypointDiff, error) {
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

	eps, err := entrypoint.DiscoverEntrypoints(preDir, []entrypoint.EntrypointDiscoverySpec{
		{
			Type:  "kustomization",
			Regex: *regexp.MustCompile(`/(?P<name>[^/]+)/overlays/(?P<overlay>[^/]+)/kustomization.yaml`),
		},
	})

	if err != nil {
		return nil, err
	}

	var errs error
	allDiff := []EntrypointDiff{}
	wg := sync.WaitGroup{}
	for _, ep := range eps {
		if ep.Name != "k8-workshop" {
			continue
		}
		ep := ep
		wg.Add(1)
		go func() {
			defer wg.Done()
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
				errs = errors.Join(errs, buildErrs)
				return
			}

			if preResources == nil || postResources == nil {
				errs = errors.Join(errs, fmt.Errorf("unknown error fetching pre/post resources"))
				return
			}

			diff, err := resource.Diff(ctx, *preResources, *postResources)
			if err != nil {
				errs = errors.Join(errs, err)
				return
			}

			allDiff = append(allDiff, EntrypointDiff{
				Entrypoint: ep,
				Diff:       diff,
			})
		}()
	}

	wg.Wait()

	return allDiff, errs
}
