package resource

import (
	"context"
	"fmt"

	"github.com/codingninja/gitops-repo-api/entrypoint"
	"github.com/codingninja/gitops-repo-api/git"
	r3diff "github.com/r3labs/diff/v3"
)

type DiffType string

const (
	DiffTypeCreate  DiffType = r3diff.CREATE
	DiffTypeReplace DiffType = "replace"
	DiffTypeUpdate  DiffType = r3diff.UPDATE
	DiffTypeDelete  DiffType = r3diff.DELETE
)

type Resource interface {
	Type() string
	Identifier() string
	Name() string
}

type ResourceDiff struct {
	Type DiffType         `json:"type"`
	Pre  Resource         `json:"pre"`
	Post Resource         `json:"post"`
	Diff r3diff.Changelog `json:"diff"`
}

func (rd *ResourceDiff) Identifier() string {
	if rd.Post != nil {
		return rd.Post.Identifier()
	}
	if rd.Pre != nil {
		return rd.Pre.Identifier()
	}

	return "unknown"
}

func (rd *ResourceDiff) Name() string {
	if rd.Post != nil {
		return rd.Post.Name()
	}
	if rd.Pre != nil {
		return rd.Pre.Name()
	}

	return "unknown"
}

func (rd *ResourceDiff) String() string {
	return fmt.Sprintf("%s[%s]", rd.Identifier(), rd.Name())
}

type ResourceDiffer interface {
	Diff(ctx context.Context, rs *git.RepoSpec, ep entrypoint.Entrypoint, oldPath, newPath string) ([]ResourceDiff, error)
	// TODO: Implement this so that we can iterate all objects, not just all changes
	// Diff(ctx context.Context, rs *git.RepoSpec, ep entrypoint.Entrypoint, oldPath, newPath string) ([]ResourceDiff, []Resource, []Resource, error)
}

func EntrypointDiffer(ep entrypoint.Entrypoint) (ResourceDiffer, error) {
	if ep.Type == entrypoint.EntrypointTypeKustomize {
		return &kubeDiffer{}, nil
	}
	if ep.Type == entrypoint.EntrypointTypeTerraform {
		return &tfDiffer{}, nil
	}
	if ep.Type == entrypoint.EntrypointTypeCloudformation {
		return &cfnDiffer{}, nil
	}
	return nil, fmt.Errorf("entrypoint type %q is not supported", ep.Type)
}
