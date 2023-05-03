package resource

import (
	"context"
	"fmt"
	"os"

	"github.com/codingninja/gitops-repo-api/entrypoint"
	"github.com/codingninja/gitops-repo-api/git"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
)

func init() {
	go func() {
		installer := &releases.ExactVersion{
			Product:    product.Terraform,
			Version:    version.Must(version.NewVersion("1.0.6")),
			InstallDir: "/tmp",
		}

		fmt.Println("installing terraform")
		path, err := installer.Install(context.Background())
		if err != nil {
			panic(fmt.Errorf("error installing Terraform: %s", err))
		}
		tfExecPath = path
		fmt.Println("installed terraform to ", tfExecPath)
	}()
}

type TerraformResource struct {
	Resource  interface{}            `json:"resource"`
	Unknown   interface{}            `json:"unknown"`
	Sensitive interface{}            `json:"sensitive"`
	Change    *tfjson.ResourceChange `json:"change"`
}

func (kr *TerraformResource) Type() string {
	return string(entrypoint.EntrypointTypeTerraform)
}

func (kr *TerraformResource) Identifier() string {
	addr := kr.Change.Address

	if after, ok := kr.Change.Change.After.(map[string]interface{}); ok {
		if ns, ok := after["namespace"].(string); ok {
			addr = fmt.Sprintf("%s/%s", addr, ns)
		}
	}

	return fmt.Sprintf("%s[%s]", kr.Change.ProviderName, addr)
}

func (kr *TerraformResource) Name() string {
	return kr.Change.Address
}

var tfExecPath = ""

func RenderTerraform(workingDir string) (*tfjson.Plan, error) {
	tf, err := tfexec.NewTerraform(workingDir, tfExecPath)
	if err != nil {
		return nil, fmt.Errorf("error running NewTerraform: %s", err)
	}

	fmt.Println("Running init ", tfExecPath)
	err = tf.Init(context.Background(), tfexec.Upgrade(true))
	if err != nil {
		return nil, fmt.Errorf("error running Init: %s", err)
	}
	fmt.Println("Init completed ", tfExecPath)

	tfpf, err := os.CreateTemp("", "*.tfplan")
	if err != nil {
		return nil, fmt.Errorf("error creating tfplan file: %s", err)
	}
	changes, err := tf.Plan(context.Background(), tfexec.Out(tfpf.Name()))
	if err != nil {
		return nil, err
	}

	if !changes {
		return nil, nil
	}

	state, err := tf.ShowPlanFile(context.Background(), tfpf.Name())
	if err != nil {
		return nil, err
	}

	return state, nil

}

type tfDiffer struct {
}

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
