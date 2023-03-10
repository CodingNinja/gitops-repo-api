package entrypoint

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

func RenderKustomize(kustfile string) (resmap.ResMap, error) {
	opts := krusty.MakeDefaultOptions()
	pc := types.EnabledPluginConfig(types.BploLoadFromFileSys)
	pc.HelmConfig.Command = "helm"

	opts.PluginConfig = pc
	k := krusty.MakeKustomizer(opts)

	kustomization, err := os.ReadFile(kustfile)
	if err != nil {
		return nil, fmt.Errorf("unable to read kustomization - %w", err)
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
