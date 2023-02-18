package entrypoint

import (
	"fmt"
	"path/filepath"

	"sigs.k8s.io/kustomize/api/resmap"
)

func ExtractResources(root string, entrypoint Entrypoint) (resmap.ResMap, error) {
	if entrypoint.Type == EntrypointTypeKustomize {
		return RenderKustomize(filepath.Join(root, entrypoint.Directory))
	}

	return nil, fmt.Errorf("unknown entrypoint %+v", entrypoint)

}
