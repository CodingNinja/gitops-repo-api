package resource

import (
	"github.com/awslabs/goformation/v7"
	"github.com/awslabs/goformation/v7/cloudformation"
)

func RenderCloudformation(cfnFile string) (*cloudformation.Template, error) {
	// Open a template from file (can be JSON or YAML)
	template, err := goformation.Open(cfnFile)
	if err != nil {
		return nil, err
	}

	return template, nil
}
