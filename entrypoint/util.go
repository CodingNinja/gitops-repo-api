package entrypoint

import (
	"encoding/json"
	"os"
	"path"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

func isValidCloudformationEntrypoint(epPath string) bool {
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
	return false
}

func isValidCdkEntrypoint(epPath string) bool {
	if stat, err := os.Stat(path.Join(epPath, "cdk.json")); err == nil && stat != nil {
		return true
	}
	return false
}

func isValidKustomizeEntrypoint(epPath string) bool {
	if stat, err := os.Stat(path.Join(epPath, "kustomization.yaml")); err == nil && stat != nil {
		return true
	}
	return false
}

func isValidKubernetesEntrypoint(epPath string) bool {
	files, err := os.ReadDir(epPath)
	if err != nil {
		return false
	}

	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".yaml") || strings.HasSuffix(f.Name(), ".yml") {
			return true
		}
	}

	return false
}
func isValidTerraformEntrypoint(epPath string) bool {
	files, err := os.ReadDir(epPath)
	if err != nil {
		return false
	}

	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".tf") {
			return true
		}
	}

	return false
}
func isValidEntrypoint(epPath string, epType EntrypointType) bool {
	switch epType {
	case EntrypointTypeCloudformation:
		return isValidCloudformationEntrypoint(epPath)
	case EntrypointTypeCdk:
		return isValidCdkEntrypoint(epPath)
	case EntrypointTypeKubernetes:
		return isValidKubernetesEntrypoint(epPath)
	case EntrypointTypeKustomize:
		return isValidKustomizeEntrypoint(epPath)
	case EntrypointTypeTerraform:
		return isValidTerraformEntrypoint(epPath)
	case EntrypointTypeHclV1:
		return isValidCdkEntrypoint(epPath)
	}
	return false
}

// regexNamedMatches returns a map of any named capture => value in regex and a boolean indicating if a match was made at all
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
