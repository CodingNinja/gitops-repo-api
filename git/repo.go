package git

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"sync"

	"github.com/codingninja/gitops-repo-api/tracing"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// RepoSpecOption is a functional option for configuring a RepoSpec.
type RepoSpecOption func(*RepoSpec)

// WithTracer configures the RepoSpec to use the provided tracer for instrumentation.
func WithTracer(tracer *tracing.Tracer) RepoSpecOption {
	return func(rs *RepoSpec) {
		rs.tracer = tracer
	}
}

// WithOtelTracer configures the RepoSpec to use the provided OpenTelemetry tracer for instrumentation.
func WithOtelTracer(tracer trace.Tracer) RepoSpecOption {
	return func(rs *RepoSpec) {
		rs.tracer = tracing.NewTracer(tracer)
	}
}

// NewRepoSpec creates a new RepoSpec instance with the given URL and credentials.
// Optional functional options can be provided to configure the RepoSpec's behavior,
// such as adding tracing support.
func NewRepoSpec(url string, credentials transport.AuthMethod, opts ...RepoSpecOption) *RepoSpec {
	rs := &RepoSpec{
		URL:         url,
		Credentials: credentials,
		tracer:      tracing.NoOpTracer(),
	}

	for _, opt := range opts {
		opt(rs)
	}

	return rs
}

type RepoSpec struct {
	URL         string
	Credentials transport.AuthMethod
	Progress    io.Writer
	repo        *git.Repository
	l           sync.Mutex
	tracer      *tracing.Tracer
}

func (rs *RepoSpec) Name() string {
	return slug.Make(rs.URL)
}

func (rs *RepoSpec) CloneDirectory(branch string) string {
	return path.Join(os.TempDir(), rs.Name(), branch)
}

func (rs *RepoSpec) ListBranches(ctx context.Context) ([]*plumbing.Reference, error) {
	repo, err := rs.Open(ctx)
	if err != nil {
		return nil, fmt.Errorf("error opening repo %q - %w", rs.URL, err)
	}
	refs, err := repo.Branches()
	if err != nil {
		return nil, fmt.Errorf("error getting branches for repo %q - %w", rs.URL, err)
	}
	var branches []*plumbing.Reference
	err = refs.ForEach(func(ref *plumbing.Reference) error {
		branches = append(branches, ref)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error getting branches for repo %q - %w", rs.URL, err)
	}
	return branches, nil
}

func (rs *RepoSpec) ResolveRevion(ctx context.Context, ref string) (*plumbing.Hash, error) {
	repo, err := rs.Open(ctx)
	if err != nil {
		return nil, fmt.Errorf("error opening repo %q - %w", rs.URL, err)
	}
	hash, err := repo.ResolveRevision(plumbing.Revision(ref))
	if err != nil {
		return nil, fmt.Errorf("error resolving revision %q for repo %q - %w", ref, rs.URL, err)
	}
	return hash, nil
}

// Open opens the repository, cloning it if necessary.
// The repository is cached for subsequent calls.
func (rs *RepoSpec) Open(ctx context.Context) (*git.Repository, error) {
	ctx, span := rs.tracer.Start(ctx, "git.open")
	defer span.End()

	span.SetAttributes(
		tracing.GitURL(rs.URL),
	)

	rs.l.Lock()
	defer rs.l.Unlock()
	if rs.repo != nil {
		return rs.repo, nil
	}
	directory := rs.CloneDirectory(".root")

	span.SetAttributes(tracing.GitDirectory(directory))

	r, err := cloneRepo(ctx, directory, true, git.CloneOptions{
		URL:      rs.URL,
		Auth:     rs.Credentials,
		Progress: rs.Progress,
	}, rs.tracer)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("unable to clone the main repo - %w", err)
	}

	rs.repo = r

	return rs.repo, nil
}

// Checkout checks out a specific reference and returns the repository and directory.
// It creates a shallow clone of the reference to enable concurrent operations.
func (rs *RepoSpec) Checkout(ctx context.Context, reference plumbing.ReferenceName) (*git.Repository, string, error) {
	ctx, span := rs.tracer.Start(ctx, "git.checkout")
	defer span.End()

	span.SetAttributes(
		tracing.GitURL(rs.URL),
		tracing.GitRef(reference.String()),
	)

	repo, err := rs.Open(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, "", fmt.Errorf("error opening repo %q - %w", rs.URL, err)
	}
	// If the thing we want to checkout is a ref, first update that ref
	// in the local base repo to match the latest fetched origin hash
	cur, err := repo.ResolveRevision(plumbing.Revision(plumbing.NewRemoteReferenceName("origin", reference.Short())))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, "", fmt.Errorf("error resolving repo reference %q %q - %w", rs.URL, reference, err)
	}

	span.SetAttributes(tracing.GitHash(cur.String()))

	if err := repo.Storer.SetReference(plumbing.NewHashReference(reference, *cur)); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, "", err
	}
	hashRef := plumbing.NewHashReference(reference, *cur)

	rootDirectory := rs.CloneDirectory(".root")
	// We clone the root directory to enable multiple concurrent bulids of the same
	// repo without killing the upstream git repo
	dirName := hashRef.Target().Short()
	if dirName == "" {
		dirName = hashRef.Hash().String()
	}
	branchDirectory := path.Join(rs.CloneDirectory(dirName), uuid.New().String())

	span.SetAttributes(tracing.GitDirectory(branchDirectory))

	branchRepo, err := cloneRepo(ctx, branchDirectory, false, git.CloneOptions{
		URL:               rootDirectory,
		Depth:             1,
		ReferenceName:     hashRef.Name(),
		SingleBranch:      true,
		Progress:          rs.Progress,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	}, rs.tracer)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, "", fmt.Errorf("unable to create branch repo - %w", err)
	}

	wt, err := branchRepo.Worktree()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, "", fmt.Errorf("unable to get worktree for branch repo - %w", err)
	}

	if err := wt.Checkout(&git.CheckoutOptions{
		Hash: hashRef.Hash(),
	}); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, "", fmt.Errorf("unable to checkout reference %q for hash ref - %w", reference.String(), err)
	}

	return branchRepo, branchDirectory, nil
}
