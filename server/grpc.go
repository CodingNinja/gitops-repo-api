package server

import (
	"context"
	"fmt"

	"github.com/codingninja/gitops-repo-api/api"
	"github.com/codingninja/gitops-repo-api/diff"
	"github.com/codingninja/gitops-repo-api/entrypoint"
	"github.com/codingninja/gitops-repo-api/git"
	"github.com/go-git/go-git/v5/plumbing"
)

func NewGrpc() *diffApiServer {
	return &diffApiServer{}
}

type diffApiServer struct {
	api.DiffApiServer
}

func (das diffApiServer) Diff(ctx context.Context, dr *api.DiffRequest) (*api.DiffResponse, error) {
	from := git.NewRepoSpec(dr.From.Repository.URL, nil)

	var to *git.RepoSpec
	if dr.To.Repository != nil {
		to = git.NewRepoSpec(dr.To.Repository.URL, nil)
	} else {
		to = git.NewRepoSpec(dr.From.Repository.URL, nil)
	}

	epds := []entrypoint.EntrypointFactory{}
	if aep := dr.Filters.GetAutomatedEntrypoint(); aep != nil {
		ctx := map[string]interface{}{}
		for k, v := range aep.Context {
			ctx[k] = v
		}
		epds = append(epds, entrypoint.AutomaticDiscovery(ctx, nil))
	}

	fmt.Printf("will diff from %s to %s\n", dr.From.Target, dr.To.Target)

	differ := diff.NewDiffer(from, to, epds)
	diff, err := differ.Diff(ctx, plumbing.NewBranchReferenceName(dr.From.Target), plumbing.NewBranchReferenceName(dr.To.Target))
	if err != nil {
		return &api.DiffResponse{
			Error: err.Error(),
		}, err
	}
	result := api.NewDiffResponse(diff)
	return result, nil
}
