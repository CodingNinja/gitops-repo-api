package entrypoint

type EntrypointType string

const (
	EntrypointTypeKustomize EntrypointType = "kustomization"
)

type Entrypoint struct {
	Name      string
	Directory string
	Type      EntrypointType
	Context   map[string]string
}
