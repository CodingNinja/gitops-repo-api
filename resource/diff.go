package resource

import (
	"context"
	"errors"
	"fmt"

	r3diff "github.com/r3labs/diff/v3"
	"gopkg.in/yaml.v2"
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

func diffResources(aRes, bRes *resource.Resource) (r3diff.Changelog, error) {
	aYaml, err := aRes.AsYAML()
	if err != nil {
		return nil, err
	}
	bYaml, err := bRes.AsYAML()
	if err != nil {
		return nil, err
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

func Diff(ctx context.Context, old resmap.ResMap, new resmap.ResMap) ([]ResourceDiff, error) {
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

		changelog, err := diffResources(matching, r)

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
