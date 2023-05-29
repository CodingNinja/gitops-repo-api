package entrypoint

// EntrypointType represents the tool which is used to manage the IaC resources in an Entrypoint
type EntrypointType string

const (
	EntrypointTypeKubernetes     EntrypointType = "kubernetes"
	EntrypointTypeKustomize      EntrypointType = "kustomize"
	EntrypointTypeCloudformation EntrypointType = "cloudformation"
	EntrypointTypeCdk            EntrypointType = "cdk"
	EntrypointTypeTerraform      EntrypointType = "terraform"
	EntrypointTypeHclV1          EntrypointType = "hclv1"
	EntrypointTypeHclV2          EntrypointType = "hclv2"
)

// Entrypoint represents a path in a repository which contains IaC resources which can be loaded by Sancire
type Entrypoint struct {
	Name      string                 `json:"name"`
	Directory string                 `json:"directory"`
	Type      EntrypointType         `json:"type"`
	Context   map[string]interface{} `json:"context"`
}
