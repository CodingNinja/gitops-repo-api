package resource

import (
	"context"
	"encoding/json"
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

type fakeResourceDiff ResourceDiff

func (rd *ResourceDiff) MarshalJSON() ([]byte, error) {
	resourceType := ""
	if rd.Pre != nil {
		resourceType = rd.Pre.Type()
	} else if rd.Post != nil {
		resourceType = rd.Post.Type()
	}

	return json.Marshal(struct {
		fakeResourceDiff
		Identifier   string `json:"identifier"`
		Name         string `json:"name"`
		ResourceType string `json:"resourceType"`
	}{
		fakeResourceDiff: fakeResourceDiff(*rd),
		Identifier:       rd.Identifier(),
		Name:             rd.Name(),
		ResourceType:     resourceType,
	})
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
	Diff(ctx context.Context, rs *git.RepoSpec, ep entrypoint.Entrypoint, oldPath, newPath string) ([]ResourceDiff, []Resource, []Resource, error)
}

func EntrypointDiffer(ep entrypoint.Entrypoint) (ResourceDiffer, error) {
	switch ep.Type {
	case entrypoint.EntrypointTypeKubernetes:
		return &kubeDiffer{}, nil
	case entrypoint.EntrypointTypeKustomize:
		return &kustomizeDiffer{}, nil
	case entrypoint.EntrypointTypeTerraform:
		return &tfDiffer{}, nil
	case entrypoint.EntrypointTypeCloudformation:
		return &cfnDiffer{}, nil
	case entrypoint.EntrypointTypeCdk:
		return &cdkDiffer{}, nil
	default:
		return nil, fmt.Errorf("entrypoint type %q is not supported", ep.Type)
	}
}
