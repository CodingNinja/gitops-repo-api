package crud

import (
	"github.com/codingninja/gitops-repo-api/entrypoint"
	"github.com/codingninja/gitops-repo-api/resource"
)

func UpdateObject(r resource.KubernetesResource, patch resource.KubernetesResource, ep entrypoint.Entrypoint) error {
	return r.ApplySmPatch(patch.Resource)
}
