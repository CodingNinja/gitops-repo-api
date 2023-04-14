package server

// import (
// 	"context"
// 	"fmt"
// 	"os"
// 	"regexp"

// 	"github.com/codingninja/gitops-repo-api/api"
// 	"github.com/codingninja/gitops-repo-api/diff"
// 	"github.com/codingninja/gitops-repo-api/entrypoint"
// 	"github.com/codingninja/gitops-repo-api/git"
// 	"github.com/go-git/go-git/v5/plumbing"
// )

func NewGrpc() *diffApiServer {
	return &diffApiServer{}
}

type diffApiServer struct {
	// api.DiffApiServer
}

// func (das diffApiServer) Diff(ctx context.Context, dr *api.DiffRequest) (*api.DiffResponse, error) {
// 	debug := true
// 	rs := git.NewRepoSpec(dr.RepoUrl, nil)

// 	if debug {
// 		rs.Progress = os.Stdout
// 	}

// 	branchName := dr.To
// 	preRef := plumbing.NewHashReference(plumbing.NewBranchReferenceName(branchName), plumbing.NewHash(dr.From))
// 	postRef := plumbing.NewSymbolicReference(preRef.Name(), preRef.Name())
// 	epds := []entrypoint.EntrypointDiscoverySpec{
// 		{
// 			Type: "kustomization",
// 			// Regex: *regexp.MustCompile(`/(?P<name>[^/]+)/overlays/(?P<overlay>[^/]+)/kustomization.yaml`),
// 			Regex: *regexp.MustCompile(`/k8-workshop/overlays/(?P<overlay>[^/]+)/kustomization.yaml`),
// 			Context: map[string]string{
// 				"name": "k8-workshop",
// 			},
// 		},
// 	}
// 	differ := diff.NewDiffer(rs, epds)
// 	diff, err := differ.Diff(ctx, preRef, postRef)
// 	if err != nil {
// 		fmt.Printf("Got errors diffing resources:\n\n%s\n", err.Error())
// 	}
// 	result := api.DiffResponse(api.DiffResponse{
// 		Diffs: []*api.DiffResponse_Diff{
// 			&api.DiffResponse_Diff{
// 				Entrypoint: &api.Entrypoint{},
// 			},
// 		},
// 	})
// 	return &result, nil
// }
