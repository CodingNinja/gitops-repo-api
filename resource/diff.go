package resource

import (
	"errors"
	"fmt"

	r3diff "github.com/r3labs/diff/v3"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
)

type DiffType string

const (
	DiffTypeCreate DiffType = "create"
	DiffTypeUpdate DiffType = "update"
	DiffTypeDelete DiffType = "delete"
)

type ResourceDiff struct {
	Type DiffType
	Pre  *resource.Resource
	Post *resource.Resource
	Diff r3diff.Changelog
}

func Diff(old resmap.ResMap, new resmap.ResMap) ([]ResourceDiff, error) {
	diff := []ResourceDiff{}

	var errs error
	for _, r := range new.Resources() {
		matching, err := old.GetByCurrentId(r.CurId())
		if err != nil {
			diff = append(diff, ResourceDiff{
				Type: DiffTypeCreate,
				Pre:  nil,
				Post: r,
				Diff: r3diff.Changelog{},
			})
			continue
		}
		fmt.Printf("Comparing new object %q with old %q\n", r.CurId(), matching.CurId())
		changelog, err := r3diff.Diff(matching, r)

		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}

		if len(changelog) > 0 {
			diff = append(diff, ResourceDiff{
				Type: DiffTypeUpdate,
				Pre:  r,
				Post: matching,
				Diff: changelog,
			})
		}
	}
	for _, r := range old.Resources() {
		if _, err := new.GetByCurrentId(r.CurId()); err != nil {
			diff = append(diff, ResourceDiff{
				Type: DiffTypeUpdate,
				Pre:  r,
				Post: nil,
				Diff: r3diff.Changelog{},
			})
		}
	}

	return diff, errs
}
