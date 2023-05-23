package util

import (
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

type basicKubeResource struct {
	APIVersion string `json:"apiVersion" yaml:"apiVersion"`
	Kind       string `json:"kind" yaml:"kind"`
}

func IsValidKubeFile(file string) bool {
	if strings.HasSuffix(file, ".yaml") || strings.HasSuffix(file, ".yml") {
		content, err := os.ReadFile(file)
		if err != nil {
			return false
		}
		kr := &basicKubeResource{}
		if err := yaml.Unmarshal(content, kr); err != nil {
			return false
		}
		if kr.APIVersion != "" && kr.Kind != "" {
			return true
		}
	}
	return false
}
