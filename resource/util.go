package resource

import (
	"errors"
	"fmt"
	"sync"

	"github.com/codingninja/gitops-repo-api/entrypoint"
)

type ResourceExtractor[T any] func(dir string, ep entrypoint.Entrypoint) (T, error)

func extractConcurrent[T any](ep entrypoint.Entrypoint, preDir string, postDir string, extract ResourceExtractor[T]) (T, T, error) {

	ewg := sync.WaitGroup{}
	ewg.Add(1)
	var preResources T
	var postResources T
	var buildErrs error
	go func() {
		defer ewg.Done()
		if preDir != "" {

			pr, err := extract(preDir, ep)
			if err != nil {
				buildErrs = errors.Join(buildErrs, fmt.Errorf("unable to build pre-entrypoint %q - %w", preDir, err))
				return
			}
			preResources = pr
		}

	}()

	ewg.Add(1)
	go func() {
		defer ewg.Done()
		if postDir != "" {
			pr, err := extract(postDir, ep)
			if err != nil {
				buildErrs = errors.Join(buildErrs, fmt.Errorf("unable to build post-entrypoint %q - %w", postDir, err))
				return
			}
			postResources = pr
		}
	}()

	ewg.Wait()

	return preResources, postResources, buildErrs
}
