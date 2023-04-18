package resource

import (
	"context"

	"github.com/codingninja/gitops-repo-api/entrypoint"
	"github.com/codingninja/gitops-repo-api/git"
)

type tfDiffer struct {
}

/*
 ResourceChanges: ([]*tfjson.ResourceChange) (len=104 cap=141) {
  (*tfjson.ResourceChange)(0x14005857710)({
   Address: (string) (len=24) "random_password.password",
   ModuleAddress: (string) "",
   Mode: (tfjson.ResourceMode) (len=7) "managed",
   Type: (string) (len=15) "random_password",
   Name: (string) (len=8) "password",
   Index: (interface {}) <nil>,
   ProviderName: (string) (len=38) "registry.terraform.io/hashicorp/random",
   DeposedKey: (string) "",
   Change: (*tfjson.Change)(0x14000270230)({
    Actions: (tfjson.Actions) (len=1 cap=4) {
     (tfjson.Action) (len=6) "create"
    },
    Before: (interface {}) <nil>,
    After: (map[string]interface {}) (len=12) {
     (string) (len=5) "lower": (bool) true,
     (string) (len=9) "min_lower": (float64) 0,
     (string) (len=11) "min_numeric": (float64) 0,
     (string) (len=9) "min_upper": (float64) 0,
     (string) (len=6) "number": (bool) true,
     (string) (len=7) "numeric": (bool) true,
     (string) (len=16) "override_special": (string) (len=20) "!#$%&*()-_=+[]{}<>:?",
     (string) (len=7) "special": (bool) true,
     (string) (len=7) "keepers": (interface {}) <nil>,
     (string) (len=6) "length": (float64) 16,
     (string) (len=11) "min_special": (float64) 0,
     (string) (len=5) "upper": (bool) true
    },
    AfterUnknown: (map[string]interface {}) (len=3) {
     (string) (len=11) "bcrypt_hash": (bool) true,
     (string) (len=2) "id": (bool) true,
     (string) (len=6) "result": (bool) true
    },
    BeforeSensitive: (bool) false,
    AfterSensitive: (map[string]interface {}) (len=2) {
     (string) (len=11) "bcrypt_hash": (bool) true,
     (string) (len=6) "result": (bool) true
    }
   })
  }),

*/

func (td *tfDiffer) Diff(ctx context.Context, rs *git.RepoSpec, ep entrypoint.Entrypoint, oldDir, newDir string) ([]ResourceDiff, error) {
	tfplan, err := RenderTerraform(newDir)
	if err != nil {
		return nil, err
	}
	diff := []ResourceDiff{}
	for _, rc := range tfplan.ResourceChanges {
		rd := ResourceDiff{
			Pre: &TerraformResource{
				Resource:  rc.Change.Before,
				Sensitive: rc.Change.BeforeSensitive,
				Unknown:   nil,
				Change:    rc,
			},
			Post: &TerraformResource{
				Resource:  rc.Change.After,
				Sensitive: rc.Change.AfterSensitive,
				Unknown:   rc.Change.AfterUnknown,
				Change:    rc,
			},
		}

		if rc.Change.Actions.Replace() {
			rd.Type = DiffTypeReplace
		} else if rc.Change.Actions.Create() {
			rd.Type = DiffTypeCreate
		} else if rc.Change.Actions.Delete() {
			rd.Type = DiffTypeDelete
		} else if rc.Change.Actions.Update() {
			rd.Type = DiffTypeUpdate
		}

		diff = append(diff, rd)
	}

	return diff, nil
}
