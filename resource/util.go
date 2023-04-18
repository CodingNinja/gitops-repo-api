package resource

import (
	"errors"
	"fmt"
	"sync"

	"github.com/codingninja/gitops-repo-api/entrypoint"
)

type ResourceExtractor func(dir string, ep entrypoint.Entrypoint) (interface{}, error)

func extractConcurrent(ep entrypoint.Entrypoint, preDir string, postDir string, extract ResourceExtractor) (interface{}, interface{}, error) {

	ewg := sync.WaitGroup{}
	ewg.Add(1)
	var preResources interface{}
	var postResources interface{}
	var buildErrs error
	go func() {
		defer ewg.Done()
		pr, err := extract(preDir, ep)
		if err != nil {
			buildErrs = errors.Join(buildErrs, fmt.Errorf("unable to build pre-entrypoint - %w", err))
			return
		}
		preResources = pr

	}()

	ewg.Add(1)
	go func() {
		defer ewg.Done()
		pr, err := extract(postDir, ep)
		if err != nil {
			buildErrs = errors.Join(buildErrs, fmt.Errorf("unable to build post-entrypoint - %w", err))
			return
		}
		postResources = pr
	}()

	ewg.Wait()
	if buildErrs != nil && preResources == nil && postResources == nil {
		return nil, nil, buildErrs
	}

	return preResources, postResources, nil
}
