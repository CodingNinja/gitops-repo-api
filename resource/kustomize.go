package resource

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/codingninja/gitops-repo-api/entrypoint"
	"github.com/codingninja/gitops-repo-api/git"
	"gopkg.in/yaml.v3"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

const KustomizationFileSuffix = "kustomization.yaml"

func RenderKustomize(kustomizeDir string) (resmap.ResMap, error) {
	opts := krusty.MakeDefaultOptions()
	pc := types.EnabledPluginConfig(types.BploLoadFromFileSys)
	pc.HelmConfig.Command = "helm"
	kustfile := kustomizeDir
	if !strings.HasSuffix(kustomizeDir, KustomizationFileSuffix) {
		kustfile = path.Join(kustfile, KustomizationFileSuffix)
	}

	opts.PluginConfig = pc
	k := krusty.MakeKustomizer(opts)

	kustomization, err := os.ReadFile(kustfile)
	if err != nil {
		return nil, nil
	}
	kust := &types.Kustomization{}
	if err := yaml.Unmarshal(kustomization, kust); err != nil {
		return nil, fmt.Errorf("unable to parse kustomization - %w", err)
	}
	kust.BuildMetadata = append(kust.BuildMetadata, "originAnnotations")
	kustomization, err = yaml.Marshal(kust)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal - %w", err)
	}
	if err := os.WriteFile(kustfile, kustomization, 0777); err != nil {
		return nil, fmt.Errorf("unable to write new kustomization - %w", err)
	}

	resmap, err := k.Run(filesys.MakeFsOnDisk(), filepath.Dir(kustfile))

	if err != nil {
		return nil, fmt.Errorf("unable to build entrypoint with  kustomize - %w", err)
	}

	return resmap, nil
}

type kustomizeDiffer struct {
}

func (kd *kustomizeDiffer) Diff(ctx context.Context, rs *git.RepoSpec, ep entrypoint.Entrypoint, oldPath, newPath string) ([]ResourceDiff, []Resource, []Resource, error) {
	old, new, err := extractConcurrent(ep, oldPath, newPath, func(dir string, ep entrypoint.Entrypoint) (resmap.ResMap, error) {
		return RenderKustomize(dir)
	})

	if err != nil {
		return nil, nil, nil, err
	}

	return doResmapDiff(ctx, rs, ep, old, new)
}
