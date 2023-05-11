package entrypoint

import (
	"encoding/json"
	"fmt"
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
		if len(realpath) >= 4 && realpath[0:4] == ".git" {
			return nil
		}
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

				if name == "" {
					name = realpath
				}

				name = slug.Make(name)
				epType := s.Type
				if ctxType, ok := epctx["type"]; ok && epType == "" {
					if ctxTypeStr, ok := ctxType.(string); ok {
						epType = EntrypointType(ctxTypeStr)
					}
				}

				if !isValidEntrypoint(path, epType) {
					fmt.Printf("%s is not a valid entrypoint\n", path)
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

func isValidEntrypoint(epPath string, epType EntrypointType) bool {
	if epType == EntrypointTypeCloudformation {
		content, err := os.ReadFile(epPath)
		tpl := &cfnMinimalTemplate{}
		if err == nil {
			if strings.Contains(epPath, ".json") {
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

	if epType == EntrypointTypeCdk {
		if stat, err := os.Stat(path.Join(epPath, "cdk.json")); err == nil && stat != nil {
			return true
		}
	}

	return epType == EntrypointTypeTerraform || epType == EntrypointTypeKustomize
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
