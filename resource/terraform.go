package resource

import (
	"context"
	"fmt"
	"os"

	"github.com/codingninja/gitops-repo-api/entrypoint"
	"github.com/codingninja/gitops-repo-api/git"
	"github.com/codingninja/gitops-repo-api/tracing"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
	"go.opentelemetry.io/otel/codes"
)

var tfExecPath = ""
var tfLoaded = make(chan struct{})

func init() {
	tmpDir, err := os.MkdirTemp("", "tfinit-*")
	if err != nil {
		panic("unable to create temp dir for terraform plugins")
	}
	os.Setenv("TF_PLUGIN_CACHE_DIR", tmpDir)
	go func() {
		installer := &releases.ExactVersion{
			Product:    product.Terraform,
			Version:    version.Must(version.NewVersion("1.0.6")),
			InstallDir: "/tmp",
		}

		path, err := installer.Install(context.Background())
		if err != nil {
			panic(fmt.Errorf("error installing Terraform: %s", err))
		}
		tfExecPath = path
		close(tfLoaded)
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

// RenderTerraform renders a Terraform plan for the given working directory.
// It extracts the tracer from the context if available.
func RenderTerraform(ctx context.Context, workingDir string) (*tfjson.Plan, error) {
	tracer := getTracerFromContext(ctx)
	ctx, span := tracer.Start(ctx, "resource.terraform.render")
	defer span.End()

	span.SetAttributes(
		tracing.WorkingDir(workingDir),
		tracing.ResourceType("terraform"),
	)

	<-tfLoaded
	tf, err := tfexec.NewTerraform(workingDir, tfExecPath)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("error running NewTerraform: %s", err)
	}

	fmt.Println("Running init ", tfExecPath)

	err = tf.Init(ctx, tfexec.Upgrade(true))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("error running Init: %s", err)
	}
	fmt.Println("Init completed ", tfExecPath)

	tfpf, err := os.CreateTemp("", "*.tfplan")
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("error creating tfplan file: %s", err)
	}
	changes, err := tf.Plan(ctx, tfexec.Out(tfpf.Name()))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	if !changes {
		return nil, nil
	}

	state, err := tf.ShowPlanFile(ctx, tfpf.Name())
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	if state != nil && state.ResourceChanges != nil {
		span.SetAttributes(tracing.ResourceCount(len(state.ResourceChanges)))
	}

	return state, nil

}

type tfDiffer struct {
}

func (td *tfDiffer) Diff(ctx context.Context, rs *git.RepoSpec, ep entrypoint.Entrypoint, oldDir, newDir string) ([]ResourceDiff, []Resource, []Resource, error) {
	tfplan, err := RenderTerraform(ctx, newDir)
	if err != nil {
		return nil, nil, nil, err
	}
	diff := []ResourceDiff{}
	allResources := []Resource{}
	for _, rc := range tfplan.PlannedValues.RootModule.Resources {
		allResources = append(allResources, &TerraformResource{
			Resource:  rc,
			Sensitive: rc.SensitiveValues,
		})
	}
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

	return diff, allResources, allResources, nil
}
