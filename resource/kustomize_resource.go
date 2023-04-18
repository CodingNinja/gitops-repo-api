package resource

import (
	"fmt"

	"github.com/codingninja/gitops-repo-api/entrypoint"
	"sigs.k8s.io/kustomize/api/resource"
)

type KustomizeResource struct {
	Resource *resource.Resource `json:"resource"`
	Origin   resource.Origin    `json:"origin"`
}

func (kr *KustomizeResource) Type() string {
	return string(entrypoint.EntrypointTypeKustomize)
}
func (kr *KustomizeResource) Identifier() string {
	return fmt.Sprintf("%s[%s]", kr.Resource.GetGvk().String(), kr.Name())
}

func (kr *KustomizeResource) Name() string {
	ns := kr.Resource.GetNamespace()
	name := kr.Resource.GetName()
	if ns == "" {
		return name
	}

	return fmt.Sprintf("%s/%s", ns, name)
}
