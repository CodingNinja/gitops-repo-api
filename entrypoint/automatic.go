package entrypoint

import (
	"path"

	"github.com/gosimple/slug"
)

func AutomaticDiscovery(ctx map[string]interface{}) EntrypointAutomaticDiscovery {
	return EntrypointAutomaticDiscovery{
		SupportedTypes: DefaultSupportedTypes,
		Context:        ctx,
	}
}

type EntrypointAutomaticDiscovery struct {
	SupportedTypes map[EntrypointType]bool
	Context        map[string]interface{}
}

var DefaultSupportedTypes = map[EntrypointType]bool{
	EntrypointTypeCdk:            true,
	EntrypointTypeHclV1:          false,
	EntrypointTypeHclV2:          false,
	EntrypointTypeCloudformation: true,
	EntrypointTypeKubernetes:     true,
	EntrypointTypeKustomize:      true,
	EntrypointTypeTerraform:      true,
}

func (epds EntrypointAutomaticDiscovery) MakeEntrypoint(basedir, repoPath string, isFile bool) (*Entrypoint, error) {
	abs := path.Join(basedir, repoPath)
	if epds.SupportedTypes[EntrypointTypeCdk] && !isFile {
		if isValidCdkEntrypoint(abs) {
			return &Entrypoint{
				Type:      EntrypointTypeCdk,
				Name:      slug.Make(repoPath),
				Directory: repoPath,
				Context:   copyMap(epds.Context),
			}, nil
		}
	}
	if epds.SupportedTypes[EntrypointTypeCloudformation] && isFile {
		if isValidCloudformationEntrypoint(abs) {
			return &Entrypoint{
				Type:      EntrypointTypeCloudformation,
				Name:      slug.Make(repoPath),
				Directory: repoPath,
				Context:   copyMap(epds.Context),
			}, nil
		}
	}
	if epds.SupportedTypes[EntrypointTypeKubernetes] && !isFile {
		if isValidKubernetesEntrypoint(abs) {
			return &Entrypoint{
				Type:      EntrypointTypeKubernetes,
				Name:      slug.Make(repoPath),
				Directory: repoPath,
				Context:   copyMap(epds.Context),
			}, nil
		}
	}
	if epds.SupportedTypes[EntrypointTypeKustomize] && !isFile {
		if isValidKustomizeEntrypoint(abs) {
			return &Entrypoint{
				Type:      EntrypointTypeKustomize,
				Name:      slug.Make(repoPath),
				Directory: repoPath,
				Context:   copyMap(epds.Context),
			}, nil
		}
	}
	if epds.SupportedTypes[EntrypointTypeTerraform] && !isFile {
		if isValidTerraformEntrypoint(abs) {
			return &Entrypoint{
				Type:      EntrypointTypeTerraform,
				Name:      slug.Make(repoPath),
				Directory: repoPath,
				Context:   copyMap(epds.Context),
			}, nil
		}
	}

	return nil, nil
}

func copyMap[K string, V any](o map[K]V) map[K]V {
	c := map[K]V{}
	for k, v := range o {
		c[k] = v
	}
	return c
}
