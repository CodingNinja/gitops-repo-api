package resource

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/codingninja/gitops-repo-api/entrypoint"
	"github.com/codingninja/gitops-repo-api/tracing"
)

// contextKey is a custom type for context keys to avoid collisions.
type contextKey int

const (
	tracerContextKey contextKey = iota
)

// ContextWithTracer returns a new context with the provided tracer attached.
func ContextWithTracer(ctx context.Context, tracer *tracing.Tracer) context.Context {
	return context.WithValue(ctx, tracerContextKey, tracer)
}

// getTracerFromContext extracts a tracer from the context.
// If no tracer is found, it returns a no-op tracer.
func getTracerFromContext(ctx context.Context) *tracing.Tracer {
	if tracer, ok := ctx.Value(tracerContextKey).(*tracing.Tracer); ok {
		return tracer
	}
	return tracing.NoOpTracer()
}

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
