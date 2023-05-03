package resource

import (
	"context"
	"fmt"

	"encoding/json"
	"os"
	"strings"

	"github.com/awslabs/goformation/v7/cloudformation"
	"github.com/codingninja/gitops-repo-api/entrypoint"
	"github.com/codingninja/gitops-repo-api/git"
	r3diff "github.com/r3labs/diff/v3"

	"gopkg.in/yaml.v3"
)

type cfnResource struct {
	Type       string                 `json:"Type" yaml:"Type"`
	Properties map[string]interface{} `json:"Properties" yaml:"Properties"`
	Metadata   map[string]interface{} `json:"Metadata" yaml:"Metadata"`
}

type CloudformationResource struct {
	ResName  string      `json:"string"`
	Resource cfnResource `json:"resource"`
}

func (kr *CloudformationResource) Type() string {
	return string(entrypoint.EntrypointTypeCloudformation)
}

func (kr *CloudformationResource) Identifier() string {

	return fmt.Sprintf("%s[%s]", kr.Resource, kr.Name())
}

func (kr *CloudformationResource) Name() string {
	return kr.ResName
}

// Unfortunately, the CFN golang library can't be used because we don't want to bork on version mismatch
type CloudformationTemplate struct {
	AWSTemplateFormatVersion string                    `json:"AWSTemplateFormatVersion,omitempty" yaml:"AWSTemplateFormatVersion"`
	Transform                *cloudformation.Transform `json:"Transform,omitempty"`
	Description              string                    `json:"Description,omitempty" yaml:"Description"`
	Metadata                 map[string]interface{}    `json:"Metadata,omitempty" yaml:"Metadata"`
	Parameters               map[string]interface{}    `json:"Parameters,omitempty" yaml:"Parameters"`
	Mappings                 map[string]interface{}    `json:"Mappings,omitempty" yaml:"Mappings"`
	Conditions               map[string]interface{}    `json:"Conditions,omitempty" yaml:"Conditions"`
	Resources                map[string]cfnResource    `json:"Resources,omitempty" yaml:"Resources"`
	Outputs                  cloudformation.Outputs    `json:"Outputs,omitempty" yaml:"Outputs"`
	Globals                  map[string]interface{}    `json:"Globals,omitempty" yaml:"Globals"`
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

type cfnDiffer struct {
}

func (td *cfnDiffer) Diff(ctx context.Context, rs *git.RepoSpec, ep entrypoint.Entrypoint, oldPath, newPath string) ([]ResourceDiff, error) {
	// Won't actually run concurrently because we block during CFN builds currently due to a concurrent map read/write related to intrinsic funcs in cfn library
	old, new, err := extractConcurrent(ep, oldPath, newPath, func(dir string, ep entrypoint.Entrypoint) (*CloudformationTemplate, error) {
		return RenderCloudformation(dir)
	})

	if err != nil {
		return nil, fmt.Errorf("unable to concurrently render cfn resources - %w", err)
	}

	diff := []ResourceDiff{}
	for name, res := range new.Resources {
		if oldRes, ok := old.Resources[name]; !ok {

			rd := ResourceDiff{
				Type: DiffTypeCreate,
				Pre:  nil,
				Post: &CloudformationResource{
					ResName:  name,
					Resource: res,
				},
				Diff: r3diff.Changelog{},
			}
			diff = append(diff, rd)
		} else {
			rDiff, err := cfnDiffResource(res, oldRes)
			if err != nil {
				return nil, fmt.Errorf("unable to diff resources - %w", err)
			}
			if len(rDiff) > 0 {
				rd := ResourceDiff{
					Type: DiffTypeUpdate,
					Pre: &CloudformationResource{
						ResName:  name,
						Resource: oldRes,
					},
					Post: &CloudformationResource{
						ResName:  name,
						Resource: res,
					},
					Diff: rDiff,
				}
				diff = append(diff, rd)
			}
		}
	}

	for name, res := range old.Resources {
		if _, ok := new.Resources[name]; !ok {

			rd := ResourceDiff{
				Type: DiffTypeDelete,
				Pre: &CloudformationResource{
					ResName:  name,
					Resource: res,
				},
				Diff: r3diff.Changelog{},
			}
			diff = append(diff, rd)
		}
	}

	return diff, nil
}

func cfnDiffResource(aRes, bRes interface{}) (r3diff.Changelog, error) {
	aYaml, err := yaml.Marshal(aRes)
	if err != nil {
		return nil, fmt.Errorf("unable to render original resource as yaml - %w", err)
	}
	bYaml, err := yaml.Marshal(bRes)
	if err != nil {
		return nil, fmt.Errorf("unabel to render new resource as yaml - %w", err)
	}
	aObj := make(map[string]interface{})
	bObj := make(map[string]interface{})
	if err := yaml.Unmarshal([]byte(aYaml), &aObj); err != nil {
		return nil, fmt.Errorf("unable to parse yaml for item a - %w", err)
	}
	if err := yaml.Unmarshal([]byte(bYaml), &bObj); err != nil {
		return nil, fmt.Errorf("unable to parse yaml for item b - %w", err)
	}

	return r3diff.Diff(aObj, bObj)
}
