package git

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/codingninja/gitops-repo-api/tracing"
	"github.com/go-git/go-git/v5"
	"go.opentelemetry.io/otel/codes"
)

// cloneRepo clones a git repository to the specified directory with tracing support.
// If the directory already exists, it attempts to fetch the latest changes instead.
func cloneRepo(ctx context.Context, directory string, isBare bool, opts git.CloneOptions, tracer *tracing.Tracer) (*git.Repository, error) {
	ctx, span := tracer.Start(ctx, "git.clone")
	defer span.End()

	span.SetAttributes(
		tracing.GitURL(opts.URL),
		tracing.GitDirectory(directory),
		tracing.GitRef(opts.ReferenceName.String()),
	)

	err := os.MkdirAll(path.Dir(directory), 0700)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("unable to create temp dir for repo - %w", err)
	}
	var r *git.Repository
	if file, err := os.Stat(directory); err == nil && file.IsDir() {
		var err error
		r, err = git.PlainOpenWithOptions(directory, &git.PlainOpenOptions{})
		if err != nil {
			if err := os.RemoveAll(directory); err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
				return nil, fmt.Errorf("unable to cleanup bad cache dir - %w", err)
			}
		} else {
			err = r.FetchContext(ctx, &git.FetchOptions{
				Auth:     opts.Auth,
				Progress: opts.Progress,
				Force:    true,
			})
			if err != nil && err != git.NoErrAlreadyUpToDate {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
				return nil, fmt.Errorf("unable to fetch latest changes for %q from %q - %w", directory, opts.URL, err)
			}
		}
	}

	if r == nil {
		var err error
		r, err = git.PlainCloneContext(ctx, directory, isBare, &opts)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, fmt.Errorf("unable to clone repo %q @ %q to %q - %w", opts.URL, opts.ReferenceName.String(), directory, err)
		}
	}

	return r, nil
}
