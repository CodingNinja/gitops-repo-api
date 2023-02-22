package entrypoint

import "github.com/go-git/go-git/v5/plumbing"

type EntrypointType string

const (
	EntrypointTypeKustomize EntrypointType = "kustomization"
)

type Entrypoint struct {
	Hash      *plumbing.Hash
	Branch    *plumbing.ReferenceName
	Name      string
	Directory string
	Type      EntrypointType
	Context   map[string]string
}
