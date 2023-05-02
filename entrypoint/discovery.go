package entrypoint

import (
	"io/fs"
	"path"
	"path/filepath"
	"regexp"

	"github.com/gosimple/slug"
)

type EntrypointDiscoverySpec struct {
	Type    EntrypointType         `json:"type"`
	Regex   regexp.Regexp          `json:"regex"`
	Files   bool                   `json:"files"`
	Context map[string]interface{} `json:"context"`
}

// DiscoverEntrypoints finds all the entrypoints matching the supplied EntrypointDiscoverySpecs
// TODO: Make this more performant, add a flag to only check dir names, include a basedir prop to limit search context
func DiscoverEntrypoints(directory string, specs []EntrypointDiscoverySpec) ([]Entrypoint, error) {
	directory = path.Clean(directory)
	entrypoints := []Entrypoint{}
	err := filepath.WalkDir(directory, func(path string, d fs.DirEntry, err error) error {
		for _, s := range specs {
			if !d.IsDir() && !s.Files {
				continue
			}
			if matches, ok := regexNamedMatches(path, s.Regex); ok {
				epctx := make(map[string]interface{})
				for k, v := range s.Context {
					epctx[k] = v
				}
				for k, v := range matches {
					epctx[k] = v
				}

				name := ""
				if n, ok := epctx["name"]; ok {
					if ns, ok := n.(string); ok {
						name = ns
					}
				}

				if !ok {
					name = slug.Make(path[len(directory)+1:])
				}

				ep := Entrypoint{
					Name:      name,
					Directory: path[len(filepath.Clean(directory))+1:],
					Type:      s.Type,
					Context:   epctx,
				}

				entrypoints = append(entrypoints, ep)

				return nil
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return entrypoints, nil
}

// regexNamedMatches Returns a map of named capture => value
func regexNamedMatches(str string, regex regexp.Regexp) (map[string]string, bool) {
	match := regex.FindStringSubmatch(str)

	if match == nil {
		return nil, false
	}

	result := make(map[string]string)
	for i, name := range regex.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}

	return result, true
}
