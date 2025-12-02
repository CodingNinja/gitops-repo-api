package resource

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/codingninja/gitops-repo-api/entrypoint"
	"github.com/codingninja/gitops-repo-api/git"
	"github.com/codingninja/gitops-repo-api/tracing"
	"go.opentelemetry.io/otel/codes"
	"gopkg.in/yaml.v3"
)

var (
	NpxExecutable     = "npx"
	NpmExecutable     = "npm"
	NpxExecutablePath = ""
	NpmExecutablePath = ""
)

func init() {
	npxPath, err := exec.LookPath(NpxExecutable)
	if err == nil {
		NpxExecutablePath = npxPath
	}

	npmPathd, err := exec.LookPath(NpmExecutable)
	if err == nil {
		NpmExecutablePath = npmPathd
	}
}

// RenderCdk renders a CDK application to a CloudFormation template.
// It extracts the tracer from the context if available.
func RenderCdk(ctx context.Context, cdkDir string) (*CloudformationTemplate, error) {
	if NpmExecutablePath == "" || NpxExecutablePath == "" {
		return nil, fmt.Errorf("missing NPM Executable Path or NPX Executable path")
	}
	tracer := getTracerFromContext(ctx)
	ctx, span := tracer.Start(ctx, "resource.cdk.render")
	defer span.End()

	span.SetAttributes(
		tracing.WorkingDir(cdkDir),
		tracing.ResourceType("cdk"),
	)

	cdkDir = strings.TrimSuffix(cdkDir, "cdk.json")
	// Open a template from file (can be JSON or YAML)
	ciCmd := exec.Command(NpmExecutablePath, "ci")
	ciCmd.Dir = cdkDir
	npmCiRes, err := ciCmd.CombinedOutput()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("unable to run `%s` - %w - %s", ciCmd.String(), err, npmCiRes)
	}
	synthCmd := exec.Command(NpxExecutablePath, "aws-cdk", "synth")
	synthCmd.Env = append(os.Environ(), "JSII_SILENCE_WARNING_DEPRECATED_NODE_VERSION=1")
	synthCmd.Dir = cdkDir
	tpl, err := synthCmd.CombinedOutput()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("unable to run `%s` - %w - %s", synthCmd.String(), err, tpl)
	}
	cft := &CloudformationTemplate{}
	if err := yaml.Unmarshal([]byte(tpl), &cft); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	if cft.Resources != nil {
		span.SetAttributes(tracing.ResourceCount(len(cft.Resources)))
	}

	return cft, nil
}

type cdkDiffer struct{}

func (td *cdkDiffer) Diff(ctx context.Context, rs *git.RepoSpec, ep entrypoint.Entrypoint, oldPath, newPath string) ([]ResourceDiff, []Resource, []Resource, error) {
	// Won't actually run concurrently because we block during CFN builds currently due to a concurrent map read/write related to intrinsic funcs in cfn library
	old, new, err := extractConcurrent(ep, oldPath, newPath, func(dir string, ep entrypoint.Entrypoint) (*CloudformationTemplate, error) {
		return RenderCdk(ctx, dir)
	})
	if err != nil && old == nil && new == nil {
		return nil, nil, nil, fmt.Errorf("error extracting cloudformation from CDK - %w", err)
	}

	return doCfnDiff(ctx, old, new)
}
