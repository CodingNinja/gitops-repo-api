package entrypoint

import (
	"io/fs"
	"path/filepath"
	"regexp"

	"github.com/gosimple/slug"
)

type EntrypointDiscoverySpec struct {
	Type    EntrypointType
	Regex   regexp.Regexp
	Context map[string]string
}

func DiscoverEntrypoints(directory string, specs []EntrypointDiscoverySpec) ([]Entrypoint, error) {
	entrypoints := []Entrypoint{}
	err := filepath.WalkDir(directory, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		for _, s := range specs {
			if matches, ok := RegexMatches(path, s.Regex); ok {
				name, ok := matches["name"]
				if !ok {
					name = slug.Make(path)
				}
				epctx := s.Context
				for k, v := range matches {
					epctx[k] = v
				}

				entrypoints = append(entrypoints, Entrypoint{
					Name:      name,
					Directory: path[len(filepath.Clean(directory))+1:],
					Type:      s.Type,
					Context:   epctx,
				})

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

func RegexMatches(str string, regex regexp.Regexp) (map[string]string, bool) {
	match := regex.FindStringSubmatch(str)
	result := make(map[string]string)

	if len(match) == 0 {
		return nil, false
	}

	for i, name := range regex.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}

	return result, len(result) > 0
}
