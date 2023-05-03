package resource

import (
	"context"

	"github.com/awslabs/goformation/v7/cloudformation"
	"github.com/codingninja/gitops-repo-api/entrypoint"
	"github.com/codingninja/gitops-repo-api/git"
	"github.com/davecgh/go-spew/spew"
)

type cfnDiffer struct {
}

func (td *cfnDiffer) Diff(ctx context.Context, rs *git.RepoSpec, ep entrypoint.Entrypoint, oldPath, newPath string) ([]ResourceDiff, error) {
	o, n, err := extractConcurrent[*cloudformation.Template](ep, oldPath, newPath, func(dir string, ep entrypoint.Entrypoint) (*cloudformation.Template, error) {
		return RenderCloudformation(dir)
	})
	if err != nil {
		return nil, err
	}
	spew.Dump(o, n)

	panic("here")
	// diff := []ResourceDiff{}
	// for _, rc := range o.Resources {
	// 	rd := ResourceDiff{
	// 		Pre: &CloudformationResource{
	// 			Resource:  rc.Change.Before,
	// 			Sensitive: rc.Change.BeforeSensitive,
	// 			Unknown:   nil,
	// 			Change:    rc,
	// 		},
	// 		Post: &CloudformationResource{
	// 			Resource:  rc.Change.After,
	// 			Sensitive: rc.Change.AfterSensitive,
	// 			Unknown:   rc.Change.AfterUnknown,
	// 			Change:    rc,
	// 		},
	// 	}

	// 	if rc.Change.Actions.Replace() {
	// 		rd.Type = DiffTypeReplace
	// 	} else if rc.Change.Actions.Create() {
	// 		rd.Type = DiffTypeCreate
	// 	} else if rc.Change.Actions.Delete() {
	// 		rd.Type = DiffTypeDelete
	// 	} else if rc.Change.Actions.Update() {
	// 		rd.Type = DiffTypeUpdate
	// 	}

	// 	diff = append(diff, rd)
	// }

	// return diff, nil
}
