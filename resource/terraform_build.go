package resource

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
)

var execPath = ""

func init() {

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
	execPath = path
	fmt.Println("installed terraform to ", execPath)
}

func RenderTerraform(workingDir string) (*tfjson.Plan, error) {
	tf, err := tfexec.NewTerraform(workingDir, execPath)
	if err != nil {
		return nil, fmt.Errorf("error running NewTerraform: %s", err)
	}

	fmt.Println("Running init ", execPath)
	err = tf.Init(context.Background(), tfexec.Upgrade(true))
	if err != nil {
		return nil, fmt.Errorf("error running Init: %s", err)
	}
	fmt.Println("Init completed ", execPath)

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
