package resource

import (
	"fmt"

	"github.com/codingninja/gitops-repo-api/entrypoint"
	tfjson "github.com/hashicorp/terraform-json"
)

type TerraformResource struct {
	Resource  interface{}            `json:"resource"`
	Unknown   interface{}            `json:"unknown"`
	Sensitive interface{}            `json:"sensitive"`
	Change    *tfjson.ResourceChange `json:"change"`
}

func (kr *TerraformResource) Type() string {
	return string(entrypoint.EntrypointTypeTerraform)
}
func (kr *TerraformResource) Identifier() string {
	addr := kr.Change.Address

	if after, ok := kr.Change.Change.After.(map[string]interface{}); ok {
		if ns, ok := after["namespace"].(string); ok {
			addr = fmt.Sprintf("%s/%s", addr, ns)
		}
	}

	return fmt.Sprintf("%s[%s]", kr.Change.ProviderName, addr)
}

func (kr *TerraformResource) Name() string {
	return kr.Change.Address
}
