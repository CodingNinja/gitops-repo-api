package resource

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Unfortunately, the CFN golang library can't be used because we don't want to bork on version mismatch
type CloudformationTemplate struct {
	AWSTemplateFormatVersion string                 `json:"AWSTemplateFormatVersion,omitempty" yaml:"AWSTemplateFormatVersion"`
	Description              string                 `json:"Description,omitempty" yaml:"Description"`
	Metadata                 map[string]interface{} `json:"Metadata,omitempty" yaml:"Metadata"`
	Parameters               map[string]interface{} `json:"Parameters,omitempty" yaml:"Parameters"`
	Mappings                 map[string]interface{} `json:"Mappings,omitempty" yaml:"Mappings"`
	Conditions               map[string]interface{} `json:"Conditions,omitempty" yaml:"Conditions"`
	Resources                map[string]interface{} `json:"Resources,omitempty" yaml:"Resources"`
	Outputs                  map[string]interface{} `json:"Outputs,omitempty" yaml:"Outputs"`
	Globals                  map[string]interface{} `json:"Globals,omitempty" yaml:"Globals"`
}

func RenderCloudformation(cfnFile string) (*CloudformationTemplate, error) {
	// Open a template from file (can be JSON or YAML)
	template, err := os.ReadFile(cfnFile)
	if err != nil {
		return nil, fmt.Errorf("error loading cloudformation %q - %w", cfnFile, err)
	}

	if strings.Contains(cfnFile, ".json") {
		tpl := CloudformationTemplate{}
		if err := json.Unmarshal(template, &tpl); err != nil {
			return nil, err
		}
		return &tpl, nil
	}
	tpl := CloudformationTemplate{}
	if err := yaml.Unmarshal(template, &tpl); err != nil {
		return nil, err
	}
	return &tpl, nil
}
