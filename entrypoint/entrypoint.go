package entrypoint

import "github.com/go-git/go-git/v5/plumbing"

type EntrypointType string

const (
	EntrypointTypeKustomize EntrypointType = "kustomize"
	EntrypointTypeTerraform EntrypointType = "terraform"
	EntrypointTypeHclV1     EntrypointType = "hclv1"
	EntrypointTypeHclV2     EntrypointType = "hclv2"
)

type Entrypoint struct {
	Hash      plumbing.Hash          `json:"hash"`
	Branch    plumbing.ReferenceName `json:"branch"`
	Name      string                 `json:"name"`
	Directory string                 `json:"directory"`
	Type      EntrypointType         `json:"type"`
	Context   map[string]interface{} `json:"context"`
}
