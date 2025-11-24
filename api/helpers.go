package api

import (
	"github.com/codingninja/gitops-repo-api/diff"
)

func NewDiffResponse(epds []diff.EntrypointDiff) *DiffResponse {
	dr := &DiffResponse{}
	for _, epd := range epds {
		epdRes := &DiffResponse_Diff{
			Entrypoint: &Entrypoint{
				Name:      epd.Entrypoint.Name,
				Directory: epd.Entrypoint.Directory,
				Type:      EntrypointType(EntrypointType_value[string(epd.Entrypoint.Type)]),
			},
		}

		if epd.Error != nil {
			epdRes.Error = epd.Error.Error()
		} else {
			for _, diff := range epd.Diff {
				pre := &Resource{}
				post := &Resource{}
				epdRes.Changes = append(epdRes.Changes, &DiffResponse_Diff_EntrypointDiff{
					Type: DiffType(DiffType_value[string(diff.Type)]),
					Pre:  pre,
					Post: post,
				})
			}
		}
		dr.Diffs = append(dr.Diffs, epdRes)
	}
	// []*DiffResponse_Diff{}
	return dr
}
