package resource

import (
	"fmt"

	"github.com/codingninja/gitops-repo-api/entrypoint"
)

type CloudformationResource struct {
	ResName  string      `json:"string"`
	Resource interface{} `json:"resource"`
}

func (kr *CloudformationResource) Type() string {
	return string(entrypoint.EntrypointTypeCloudformation)
}

func (kr *CloudformationResource) Identifier() string {
	return fmt.Sprintf("%s[%s]", kr.Resource, kr.Name())
}

func (kr *CloudformationResource) Name() string {
	return kr.ResName
}
