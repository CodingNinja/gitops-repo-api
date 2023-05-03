package entrypoint

import (
	"encoding/json"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gosimple/slug"
	"gopkg.in/yaml.v3"
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
		realpath := strings.TrimLeft(path[len(directory):], "/")
		for _, s := range specs {
			if !d.IsDir() && !s.Files {
				continue
			}
			if matches, ok := regexNamedMatches(realpath, s.Regex); ok {
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
					name = slug.Make(realpath)
				}
				epType := s.Type
				if ctxType, ok := epctx["type"]; ok && epType == "" {
					if ctxTypeStr, ok := ctxType.(string); ok {
						epType = EntrypointType(ctxTypeStr)
					}
				}

				if !isValidEntrypoint(path, epType) {
					return nil
				}

				ep := Entrypoint{
					Name:      name,
					Directory: realpath,
					Type:      epType,
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

type cfnMinimalTemplate struct {
	AWSTemplateFormatVersion string `json:"AWSTemplateFormatVersion" yaml:"AWSTemplateFormatVersion"`
}

func isValidEntrypoint(path string, epType EntrypointType) bool {
	if epType == EntrypointTypeCloudformation {
		content, err := os.ReadFile(path)
		tpl := &cfnMinimalTemplate{}
		if err != nil {
			if strings.Contains(path, ".json") {
				if err := json.Unmarshal(content, tpl); err != nil {
					return false
				}
			} else {
				if err := yaml.Unmarshal(content, tpl); err != nil {
					return false
				}
			}

			return tpl.AWSTemplateFormatVersion != ""
		}

	}
	return true
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
