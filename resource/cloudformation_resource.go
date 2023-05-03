package resource

import (
	"fmt"

	"github.com/awslabs/goformation/v7/cloudformation"
	"github.com/codingninja/gitops-repo-api/entrypoint"
)

type CloudformationResource struct {
	ResName  string                  `json:"string"`
	Resource cloudformation.Resource `json:"resource"`
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
