package resource

import (
	"context"
	"errors"
	"fmt"
	"path"
	"path/filepath"

	"git.dmann.xyz/davidmann/gitops-repo-api/entrypoint"
	"git.dmann.xyz/davidmann/gitops-repo-api/git"
	r3diff "github.com/r3labs/diff/v3"
	"gopkg.in/yaml.v2"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/kyaml/resid"
)

type DiffType string

const (
	DiffTypeCreate DiffType = r3diff.CREATE
	DiffTypeUpdate DiffType = r3diff.UPDATE
	DiffTypeDelete DiffType = r3diff.DELETE
)

type Resource struct {
	*resource.Resource
	resource.Origin
}
type ResourceDiff struct {
	Type DiffType
	Pre  *Resource
	Post *Resource
	Diff r3diff.Changelog
}

func (rd *ResourceDiff) Gvk() resid.Gvk {
	if rd.Post != nil {
		return rd.Post.Resource.GetGvk()
	}
	if rd.Pre != nil {
		return rd.Pre.Resource.GetGvk()
	}

	return resid.Gvk{}
}
func (rd *ResourceDiff) Name() string {
	ns := ""
	name := ""
	if rd.Post != nil {
		ns = rd.Post.Resource.GetNamespace()
		name = rd.Post.Resource.GetName()
	}
	if rd.Pre != nil {
		ns = rd.Pre.Resource.GetNamespace()
		name = rd.Pre.Resource.GetName()
	}
	if ns == "" {
		return name
	}
	return fmt.Sprintf("%s/%s", ns, name)
}
func (rd *ResourceDiff) String() string {
	return fmt.Sprintf("%s[%s]", rd.Gvk().String(), rd.Name())
}

func diffResources(aRes, bRes *resource.Resource) (r3diff.Changelog, error) {
	aYaml, err := aRes.AsYAML()
	if err != nil {
		return nil, err
	}
	bYaml, err := bRes.AsYAML()
	if err != nil {
		return nil, err
	}
	aObj := make(map[string]interface{})
	bObj := make(map[string]interface{})
	if err := yaml.Unmarshal([]byte(aYaml), &aObj); err != nil {
		return nil, fmt.Errorf("unable to parse yaml for item a - %w", err)
	}
	if err := yaml.Unmarshal([]byte(bYaml), &bObj); err != nil {
		return nil, fmt.Errorf("unable to parse yaml for item b - %w", err)
	}

	return r3diff.Diff(aObj, bObj)
}

func cleanResmap(rm resmap.ResMap) (resmap.ResMap, error) {
	crm := rm.DeepCopy()

	crm.RemoveBuildAnnotations()

	if err := crm.RemoveTransformerAnnotations(); err != nil {
		return nil, fmt.Errorf("unable to remove origin annotations - %w", err)
	}

	if err := crm.RemoveOriginAnnotations(); err != nil {
		return nil, fmt.Errorf("unable to remove origin annotations - %w", err)
	}

	return crm, nil
}

func Diff(ctx context.Context, rs *git.RepoSpec, ep entrypoint.Entrypoint, old resmap.ResMap, new resmap.ResMap) ([]ResourceDiff, error) {
	diff := []ResourceDiff{}

	// We clean the resmap for all the times we return a resource out of the diff
	// because we don't really want to expose the internal kustomize representation
	cleanedOld, err := cleanResmap(old)
	if err != nil {
		return nil, fmt.Errorf("unable to clean old resmap - %w", err)
	}

	cleanedNew, err := cleanResmap(new)
	if err != nil {
		return nil, fmt.Errorf("unable to clean old resmap - %w", err)
	}

	var errs error
	for _, newRes := range new.Resources() {
		// Match objects which we have no "old" version of
		// which indicates they are being created
		origRes, err := old.GetByCurrentId(newRes.CurId())
		if err != nil {
			cleanedPost, err := cleanedNew.GetByCurrentId(newRes.CurId())
			if err != nil {
				return nil, fmt.Errorf("unable to get cleaned copy of new resource - %W", err)
			}
			// We explicitly ignore errors here as they are only returned when there is a YAML
			// decoding error parsing the origin field
			postOrigin, _ := newRes.GetOrigin()

			diff = append(diff, ResourceDiff{
				Type: DiffTypeCreate,
				Pre:  nil,
				Post: &Resource{
					Resource: cleanedPost,
					Origin:   epOrigin(rs, ep, postOrigin),
				},
				Diff: r3diff.Changelog{},
			})
			continue
		}

		changelog, err := diffResources(origRes, newRes)

		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}

		// Catch objects which have been modified
		if len(changelog) > 0 {
			cleanedPre, err := cleanedOld.GetByCurrentId(origRes.CurId())
			if err != nil {
				return nil, fmt.Errorf("unable to get cleaned copy of resource - %W", err)
			}
			preOrigin, _ := origRes.GetOrigin()

			cleanedPost, err := cleanedNew.GetByCurrentId(newRes.CurId())
			if err != nil {
				return nil, fmt.Errorf("unable to get cleaned copy of new resource - %W", err)
			}
			postOrigin, _ := newRes.GetOrigin()

			diff = append(diff, ResourceDiff{
				Type: DiffTypeUpdate,
				Pre: &Resource{
					Resource: cleanedPre,
					Origin:   epOrigin(rs, ep, preOrigin),
				},
				Post: &Resource{
					Resource: cleanedPost,
					Origin:   epOrigin(rs, ep, postOrigin),
				},
				Diff: changelog,
			})
		}
	}

	// Finally we collect a list of all the deleted resources
	for _, r := range old.Resources() {
		if _, err := new.GetByCurrentId(r.CurId()); err != nil {
			cleanedPre, err := cleanedOld.GetByCurrentId(r.CurId())
			if err != nil {
				return nil, fmt.Errorf("unable to get cleaned copy of resource - %W", err)
			}
			origin, _ := r.GetOrigin()
			diff = append(diff, ResourceDiff{
				Type: DiffTypeUpdate,
				Pre: &Resource{
					Resource: cleanedPre,
					Origin:   epOrigin(rs, ep, origin),
				},
				Post: nil,
				Diff: r3diff.Changelog{},
			})
		}
	}

	return diff, errs
}

func epOrigin(rs *git.RepoSpec, ep entrypoint.Entrypoint, o *resource.Origin) resource.Origin {
	if o == nil {
		return resource.Origin{
			Path: ep.Directory,
			Ref:  "Unknown",
		}
	}

	origin := *o

	if origin.Path != "" {

		abs, err := filepath.Abs(path.Join("/", ep.Directory, origin.Path))
		if err == nil {
			// Strip off the leading /
			origin.Path = abs[1:]
		}
	}

	if origin.Repo == "" {
		ref := ""
		if ep.Hash != nil {
			ref = ep.Hash.String()
		} else if ep.Branch != nil {
			ref = ep.Branch.String()
		} else {
			ref = "unknown"
		}

		// todo: bug we don't actually pass the ref through properly
		// because currently the entrypoint is shared between pre and post
		// maybe extract to entrypointpath and entrypoint ?
		origin.Repo = rs.URL
		origin.Ref = ref
	}
	return origin
}
