package resource

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"

	"github.com/codingninja/gitops-repo-api/entrypoint"
	"github.com/codingninja/gitops-repo-api/git"
	"github.com/codingninja/gitops-repo-api/util"
	r3diff "github.com/r3labs/diff/v3"
	"gopkg.in/yaml.v3"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

func RenderKubernetes(manifestDir string) (resmap.ResMap, error) {
	opts := krusty.MakeDefaultOptions()
	pc := types.EnabledPluginConfig(types.BploLoadFromFileSys)
	pc.HelmConfig.Command = "helm"

	opts.PluginConfig = pc
	k := krusty.MakeKustomizer(opts)
	recursive := false

	resources := []string{}
	if recursive {
		filepath.WalkDir(manifestDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if util.IsValidKubeFile(path) {
				resources = append(resources, path[len(manifestDir)+1:])
			}
			return nil
		})
	} else {
		entries, err := os.ReadDir(manifestDir)
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			manifestAbsPath := path.Join(manifestDir, entry.Name())
			if util.IsValidKubeFile(manifestAbsPath) {
				resources = append(resources, entry.Name())
			}else{
				fmt.Printf("File %q is not a valid kubernetes manifest\n", manifestAbsPath)
			}
		}
	}

	kust := &types.Kustomization{
		Resources: resources,
	}
	kust.BuildMetadata = append(kust.BuildMetadata, "originAnnotations")
	kustomization, err := yaml.Marshal(kust)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal - %w", err)
	}
	kustfile := path.Join(manifestDir, KustomizationFileSuffix)
	if err := os.WriteFile(kustfile, kustomization, 0o777); err != nil {
		return nil, fmt.Errorf("unable to write new kustomization - %w", err)
	}

	resmap, err := k.Run(filesys.MakeFsOnDisk(), filepath.Dir(kustfile))
	if err != nil {
		return nil, fmt.Errorf("unable to build entrypoint with  kustomize - %w", err)
	}

	return resmap, nil
}

type KubernetesResource struct {
	Resource *resource.Resource `json:"resource"`
	Origin   resource.Origin    `json:"origin"`
}

func (kr *KubernetesResource) Type() string {
	return string(entrypoint.EntrypointTypeKubernetes)
}

func (kr *KubernetesResource) Identifier() string {
	return fmt.Sprintf("%s[%s]", kr.Resource.GetGvk().String(), kr.Name())
}

func (kr *KubernetesResource) Name() string {
	ns := kr.Resource.GetNamespace()
	name := kr.Resource.GetName()
	if ns == "" {
		return name
	}

	return fmt.Sprintf("%s/%s", ns, name)
}

type kubeDiffer struct{}

func (kd *kubeDiffer) Diff(ctx context.Context, rs *git.RepoSpec, ep entrypoint.Entrypoint, oldPath, newPath string) ([]ResourceDiff, []Resource, []Resource, error) {
	old, new, err := extractConcurrent(ep, oldPath, newPath, func(dir string, ep entrypoint.Entrypoint) (resmap.ResMap, error) {
		return RenderKubernetes(dir)
	})
	if err != nil {
		return nil, nil, nil, err
	}

	return doResmapDiff(ctx, rs, ep, old, new)
}

func doResmapDiff(ctx context.Context, rs *git.RepoSpec, ep entrypoint.Entrypoint, old resmap.ResMap, new resmap.ResMap) ([]ResourceDiff, []Resource, []Resource, error) {
	if old == nil {
		old = resmap.New()
	}

	if new == nil {
		new = resmap.New()
	}

	diff := []ResourceDiff{}

	// We clean the resmap for all the times we return a resource out of the diff
	// because we don't really want to expose the internal kustomize representation
	cleanedOld, err := kubeCleanResmap(old)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("unable to clean old resmap - %w", err)
	}

	cleanedNew, err := kubeCleanResmap(new)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("unable to clean old resmap - %w", err)
	}

	var errs error
	allNew := []Resource{}
	for _, newRes := range new.Resources() {
		// We explicitly ignore errors here as they are only returned when there is a YAML
		// decoding error parsing the origin field
		newResOrigin, _ := newRes.GetOrigin()
		allNew = append(allNew, &KubernetesResource{
			Resource: newRes,
			Origin:   kubeEntrypointOrigin(rs, ep, newResOrigin),
		})
		// Match objects which we have no "old" version of
		// which indicates they are being created
		origRes, err := old.GetByCurrentId(newRes.CurId())
		if err != nil {
			cleanedPost, err := cleanedNew.GetByCurrentId(newRes.CurId())
			if err != nil {
				return nil, nil, nil, fmt.Errorf("unable to get cleaned copy of new resource - %W", err)
			}

			rd := ResourceDiff{
				Type: DiffTypeCreate,
				Pre:  nil,
				Post: &KubernetesResource{
					Resource: cleanedPost,
					Origin:   kubeEntrypointOrigin(rs, ep, newResOrigin),
				},
				Diff: r3diff.Changelog{},
			}
			diff = append(diff, rd)
			continue
		}

		changelog, err := kubeDiffResmap(origRes, newRes)
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}

		// Catch objects which have been modified
		if len(changelog) > 0 {
			cleanedPre, err := cleanedOld.GetByCurrentId(origRes.CurId())
			if err != nil {
				return nil, nil, nil, fmt.Errorf("unable to get cleaned copy of resource - %W", err)
			}
			preOrigin, _ := origRes.GetOrigin()

			cleanedPost, err := cleanedNew.GetByCurrentId(newRes.CurId())
			if err != nil {
				return nil, nil, nil, fmt.Errorf("unable to get cleaned copy of new resource - %W", err)
			}
			postOrigin, _ := newRes.GetOrigin()

			diff = append(diff, ResourceDiff{
				Type: DiffTypeUpdate,
				Pre: &KubernetesResource{
					Resource: cleanedPre,
					Origin:   kubeEntrypointOrigin(rs, ep, preOrigin),
				},
				Post: &KubernetesResource{
					Resource: cleanedPost,
					Origin:   kubeEntrypointOrigin(rs, ep, postOrigin),
				},
				Diff: changelog,
			})
		}
	}

	allOld := []Resource{}
	// Finally we collect a list of all the deleted resources
	for _, r := range old.Resources() {
		newResOrigin, _ := r.GetOrigin()
		allOld = append(allOld, &KubernetesResource{
			Resource: r,
			Origin:   kubeEntrypointOrigin(rs, ep, newResOrigin),
		})
		if _, err := new.GetByCurrentId(r.CurId()); err != nil {
			cleanedPre, err := cleanedOld.GetByCurrentId(r.CurId())
			if err != nil {
				return nil, nil, nil, fmt.Errorf("unable to get cleaned copy of resource - %W", err)
			}
			origin, _ := r.GetOrigin()
			diff = append(diff, ResourceDiff{
				Type: DiffTypeUpdate,
				Pre: &KubernetesResource{
					Resource: cleanedPre,
					Origin:   kubeEntrypointOrigin(rs, ep, origin),
				},
				Post: nil,
				Diff: r3diff.Changelog{},
			})
		}
	}

	return diff, allOld, allNew, errs
}

func kubeCleanResmap(rm resmap.ResMap) (resmap.ResMap, error) {
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

func kubeEntrypointOrigin(rs *git.RepoSpec, ep entrypoint.Entrypoint, o *resource.Origin) resource.Origin {
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
		// todo: bug we don't actually pass the ref through properly
		// because currently the entrypoint is shared between pre and post
		// maybe extract to entrypointpath and entrypoint ?
		origin.Repo = rs.URL
		origin.Ref = "unknown"
	}
	return origin
}

func kubeDiffResmap(aRes, bRes *resource.Resource) (r3diff.Changelog, error) {
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
