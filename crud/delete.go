package crud

import (
	"github.com/codingninja/gitops-repo-api/entrypoint"
	"github.com/codingninja/gitops-repo-api/resource"
)

func DeleteObject(r resource.KubernetesResource, ep entrypoint.Entrypoint) error {
	r.SetYNode(nil)
	return nil
}
