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
	"github.com/codingninja/gitops-repo-api/tracing"
	"github.com/go-git/go-git/v5/plumbing"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type Differ interface {
	Extract(context.Context, plumbing.ReferenceName) ([]EntrypointDiff, error)
	Diff(context.Context, plumbing.ReferenceName, plumbing.ReferenceName) ([]EntrypointDiff, error)
	DiffEntrypoint(context.Context, entrypoint.Entrypoint, string, string) ([]resource.ResourceDiff, []resource.Resource, []resource.Resource, error)
}

// DifferOption is a functional option for configuring a Differ.
type DifferOption func(*repoDiffer)

// WithTracer configures the Differ to use the provided tracer for instrumentation.
func WithTracer(tracer *tracing.Tracer) DifferOption {
	return func(rd *repoDiffer) {
		rd.tracer = tracer
	}
}

// WithOtelTracer configures the Differ to use the provided OpenTelemetry tracer for instrumentation.
func WithOtelTracer(tracer trace.Tracer) DifferOption {
	return func(rd *repoDiffer) {
		rd.tracer = tracing.NewTracer(tracer)
	}
}

// NewDiffer creates a new Differ instance with the given repository specifications
// and entrypoint factories. Optional functional options can be provided to configure
// the Differ's behavior, such as adding tracing support.
func NewDiffer(preRs *git.RepoSpec, postRs *git.RepoSpec, epds []entrypoint.EntrypointFactory, opts ...DifferOption) *repoDiffer {
	rd := &repoDiffer{
		preRs:  preRs,
		postRs: postRs,
		epds:   epds,
		tracer: tracing.NoOpTracer(),
	}

	for _, opt := range opts {
		opt(rd)
	}

	return rd
}

type repoDiffer struct {
	preRs  *git.RepoSpec
	postRs *git.RepoSpec
	epds   []entrypoint.EntrypointFactory
	tracer *tracing.Tracer
}

type EntrypointDiff struct {
	Entrypoint entrypoint.Entrypoint   `json:"entrypoint"`
	Error      error                   `json:"error"`
	Diff       []resource.ResourceDiff `json:"diff"`
	All        []resource.Resource     `json:"all"`
}

// Extract extracts all resources from a single git reference.
// It discovers entrypoints and returns their current state.
func (rd *repoDiffer) Extract(ctx context.Context, ref plumbing.ReferenceName) ([]EntrypointDiff, error) {
	ctx, span := rd.tracer.Start(ctx, "gitops.extract")
	defer span.End()

	span.SetAttributes(
		tracing.GitRef(ref.String()),
	)

	_, dir, err := rd.preRs.Checkout(ctx, ref)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("unable to pre change dir - %w", err)
	}

	span.SetAttributes(tracing.WorkingDir(dir))

	eps, err := discoverEntrypoints(ctx, "", dir, rd.epds, rd.tracer)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	span.SetAttributes(tracing.EntrypointCount(len(eps)))

	var errs error
	allDiff := []EntrypointDiff{}
	wg := sync.WaitGroup{}
	for _, ep := range eps {
		ep := ep
		wg.Add(1)
		go func() {
			defer wg.Done()

			diff, all, _, err := rd.DiffEntrypoint(ctx, ep.ep, "", dir)
			if err != nil {
				errs = errors.Join(errs, err)
			}

			allDiff = append(allDiff, EntrypointDiff{
				Entrypoint: ep.ep,
				Diff:       diff,
				Error:      err,
				All:        all,
			})
		}()
	}

	wg.Wait()

	if errs != nil {
		span.RecordError(errs)
		span.SetStatus(codes.Error, errs.Error())
	}

	return allDiff, errs
}

// Diff compares two git references and returns the differences in resources.
// It discovers entrypoints in both references and computes resource changes.
func (rd *repoDiffer) Diff(ctx context.Context, pre, post plumbing.ReferenceName) ([]EntrypointDiff, error) {
	ctx, span := rd.tracer.Start(ctx, "gitops.diff")
	defer span.End()

	span.SetAttributes(
		tracing.GitPreRef(pre.String()),
		tracing.GitPostRef(post.String()),
	)

	_, preDir, err := rd.preRs.Checkout(ctx, pre)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("unable to pre change dir - %w", err)
	}

	span.SetAttributes(tracing.PreDir(preDir))

	_, postDir, err := rd.postRs.Checkout(ctx, post)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("unable to checkout post change dir - %w", err)
	}

	span.SetAttributes(tracing.PostDir(postDir))

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

	eps, err := discoverEntrypoints(ctx, preDir, postDir, rd.epds, rd.tracer)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	span.SetAttributes(tracing.EntrypointCount(len(eps)))

	var errs error
	allDiff := []EntrypointDiff{}
	wg := sync.WaitGroup{}
	for _, ep := range eps {
		ep := ep
		wg.Add(1)
		go func() {
			defer wg.Done()

			diff, _, post, err := rd.DiffEntrypoint(ctx, ep.ep, preDir, postDir)
			if err != nil {
				errs = errors.Join(errs, err)
			}

			allDiff = append(allDiff, EntrypointDiff{
				Entrypoint: ep.ep,
				Diff:       diff,
				Error:      err,
				All:        post,
			})
		}()
	}

	wg.Wait()

	if errs != nil {
		span.RecordError(errs)
		span.SetStatus(codes.Error, errs.Error())
	}

	return allDiff, errs
}

type internalentrypoint struct {
	t      string
	ep     entrypoint.Entrypoint
	hash   plumbing.Hash
	branch plumbing.ReferenceName
}

func discoverEntrypoints(ctx context.Context, preDir, postDir string, epds []entrypoint.EntrypointFactory, tracer *tracing.Tracer) ([]internalentrypoint, error) {
	ctx, span := tracer.Start(ctx, "gitops.diff.discover")
	defer span.End()

	span.SetAttributes(
		tracing.PreDir(preDir),
		tracing.PostDir(postDir),
	)

	// This should be re-implemented to use channels
	var preEps []entrypoint.Entrypoint
	if preDir != "" {
		preEpss, err := entrypoint.DiscoverEntrypoints(preDir, epds)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, err
		}

		preEps = preEpss
	}
	var postEps []entrypoint.Entrypoint
	if postDir != "" {
		postEpss, err := entrypoint.DiscoverEntrypoints(postDir, epds)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, err
		}

		postEps = postEpss
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

	span.SetAttributes(
		tracing.EntrypointCount(len(eplist)),
	)

	return eplist, nil
}

// DiffEntrypoint computes the resource differences for a specific entrypoint
// between pre and post directories.
func (rd *repoDiffer) DiffEntrypoint(ctx context.Context, ep entrypoint.Entrypoint, preDir, postDir string) ([]resource.ResourceDiff, []resource.Resource, []resource.Resource, error) {
	ctx, span := rd.tracer.Start(ctx, "gitops.diff.entrypoint")
	defer span.End()

	span.SetAttributes(
		tracing.EntrypointType(string(ep.Type)),
		tracing.EntrypointDirectory(ep.Directory),
		tracing.PreDir(preDir),
		tracing.PostDir(postDir),
	)

	differ, err := resource.EntrypointDiffer(ep)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, nil, nil, fmt.Errorf("unable to get differ for entrypoint - %w", err)
	}
	if preDir != "" {
		preDir = path.Join(preDir, ep.Directory)
	}
	if postDir != "" {
		postDir = path.Join(postDir, ep.Directory)
	}

	diff, pre, post, err := differ.Diff(ctx, rd.preRs, ep, preDir, postDir)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, nil, nil, fmt.Errorf("unable to extract entrypoint diff - %w", err)
	}

	span.SetAttributes(
		tracing.DiffCount(len(diff)),
		tracing.ResourceCount(len(post)),
	)

	return diff, pre, post, nil
}
