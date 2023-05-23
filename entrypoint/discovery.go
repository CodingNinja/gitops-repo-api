package entrypoint

import (
	"fmt"
	"io/fs"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gosimple/slug"
)

type EntrypointFactory interface {
	MakeEntrypoint(basedir, realpath string, isFile bool) (*Entrypoint, error)
}

type EntrypointDiscoverySpec struct {
	Type    EntrypointType         `json:"type"`
	Regex   regexp.Regexp          `json:"regex"`
	Files   bool                   `json:"files"`
	Context map[string]interface{} `json:"context"`
}

func (epds EntrypointDiscoverySpec) MakeEntrypoint(basedir, repoPath string, isFile bool) (*Entrypoint, error) {
	if !epds.Files && isFile {
		return nil, nil
	}
	if matches, ok := regexNamedMatches(repoPath, epds.Regex); ok {
		epctx := make(map[string]interface{})
		for k, v := range epds.Context {
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
			name = repoPath
		}

		name = slug.Make(name)
		epType := epds.Type
		if ctxType, ok := epctx["type"]; ok && epType == "" {
			if ctxTypeStr, ok := ctxType.(string); ok {
				epType = EntrypointType(ctxTypeStr)
			}
		}

		if !isValidEntrypoint(repoPath, epType) {
			fmt.Printf("%s is not a valid entrypoint\n", repoPath)
			return nil, nil
		}

		fmt.Printf("got entrypoint %q with regex %q\n", name, epds.Regex.String())

		ep := Entrypoint{
			Name:      name,
			Directory: repoPath,
			Type:      epType,
			Context:   epctx,
		}

		return &ep, nil
	}

	return nil, nil
}

var _ EntrypointFactory = EntrypointDiscoverySpec{}

// DiscoverEntrypoints finds all the entrypoints matching the supplied EntrypointDiscoverySpecs
// TODO: Make this more performant, add a flag to only check dir names, include a basedir prop to limit search context
func DiscoverEntrypoints(directory string, specs []EntrypointFactory) ([]Entrypoint, error) {
	directory = path.Clean(directory)
	entrypoints := []Entrypoint{}
	err := filepath.WalkDir(directory, func(path string, d fs.DirEntry, err error) error {
		realpath := strings.TrimLeft(path[len(directory):], "/")
		if len(realpath) >= 4 && realpath[0:4] == ".git" {
			return nil
		}

		for _, s := range specs {

			ep, err := s.MakeEntrypoint(directory, realpath, !d.IsDir())
			if err != nil {
				return err
			}

			if ep != nil {
				entrypoints = append(entrypoints, *ep)
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
