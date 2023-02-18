package entrypoint

import (
	"fmt"
	"path"

	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

func RenderKustomize(directory string) (resmap.ResMap, error) {
	opts := krusty.MakeDefaultOptions()
	pc := types.EnabledPluginConfig(types.BploLoadFromFileSys)
	pc.HelmConfig.Command = "helm"
	opts.PluginConfig = pc
	k := krusty.MakeKustomizer(opts)

	resmap, err := k.Run(filesys.MakeFsOnDisk(), path.Dir(directory))

	if err != nil {
		return nil, fmt.Errorf("unable to build entrypoint with  kustomize - %w", err)
	}
	r, _ := resmap.AsYaml()
	fmt.Print(string(r))

	return resmap, nil
}
