package tracing_test

import (
	"context"
	"fmt"

	"github.com/codingninja/gitops-repo-api/diff"
	"github.com/codingninja/gitops-repo-api/entrypoint"
	"github.com/codingninja/gitops-repo-api/git"
	"github.com/codingninja/gitops-repo-api/resource"
	"github.com/codingninja/gitops-repo-api/tracing"
	"github.com/go-git/go-git/v5/plumbing"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// ExampleTracer demonstrates basic usage of the tracing package.
func ExampleTracer() {
	// Create a no-op tracer (tracing disabled)
	tracer := tracing.NoOpTracer()

	ctx := context.Background()
	ctx, span := tracer.Start(ctx, "example.operation")
	defer span.End()

	// Add attributes to the span
	span.SetAttributes(
		tracing.GitURL("https://github.com/example/repo"),
		tracing.GitRef("main"),
	)

	fmt.Println("Operation traced")
	// Output: Operation traced
}

// ExampleNewTracer shows how to create a tracer with OpenTelemetry.
func ExampleNewTracer() {
	// Create an OpenTelemetry tracer provider
	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		panic(err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
	)
	defer tp.Shutdown(context.Background())

	// Set as global tracer provider
	otel.SetTracerProvider(tp)

	// Create a tracer
	otelTracer := tp.Tracer("gitops-repo-api")
	tracer := tracing.NewTracer(otelTracer)

	// Use the tracer
	ctx := context.Background()
	ctx, span := tracer.Start(ctx, "example.operation")
	defer span.End()

	fmt.Println("Traced with OpenTelemetry")
	// Output: Traced with OpenTelemetry
}

// ExampleWithTracer_diff shows how to use tracing with the diff package.
func ExampleWithTracer_diff() {
	// Setup OpenTelemetry (simplified for example)
	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	otelTracer := tp.Tracer("gitops-repo-api")
	tracer := tracing.NewTracer(otelTracer)

	// Create RepoSpec with tracing
	preRepo := git.NewRepoSpec(
		"https://github.com/example/repo",
		nil,
		git.WithTracer(tracer),
	)

	postRepo := git.NewRepoSpec(
		"https://github.com/example/repo",
		nil,
		git.WithTracer(tracer),
	)

	// Create Differ with tracing
	epds := []entrypoint.EntrypointFactory{}
	differ := diff.NewDiffer(
		preRepo,
		postRepo,
		epds,
		diff.WithTracer(tracer),
	)

	// Use the differ (spans will be created automatically)
	ctx := context.Background()
	_, _ = differ.Diff(ctx, plumbing.NewBranchReferenceName("main"), plumbing.NewBranchReferenceName("feature"))

	fmt.Println("Diff operation traced")
	// Output: Diff operation traced
}

// ExampleContextWithTracer shows how to use context-based tracing for resource rendering.
func ExampleContextWithTracer() {
	// Setup tracer
	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	otelTracer := tp.Tracer("gitops-repo-api")
	tracer := tracing.NewTracer(otelTracer)

	// Add tracer to context for resource package functions
	ctx := context.Background()
	ctx = resource.ContextWithTracer(ctx, tracer)

	// Resource rendering functions will automatically use the tracer from context
	// Example: RenderKubernetes(ctx, "/path/to/manifests")
	// Example: RenderTerraform(ctx, "/path/to/terraform")
	// Example: RenderCdk(ctx, "/path/to/cdk")

	fmt.Println("Context-based tracing configured")
	// Output: Context-based tracing configured
}

// ExampleAttributes demonstrates using standard attributes.
func ExampleAttributes() {
	tracer := tracing.NoOpTracer()
	ctx, span := tracer.Start(context.Background(), "git.operation")
	defer span.End()

	// Set multiple attributes at once
	span.SetAttributes(
		tracing.GitURL("https://github.com/example/repo"),
		tracing.GitBranch("main"),
		tracing.GitHash("abc123"),
		tracing.GitDirectory("/tmp/repo"),
		tracing.EntrypointType("kubernetes"),
		tracing.EntrypointDirectory("manifests/"),
		tracing.ResourceCount(42),
		tracing.DiffCount(5),
		tracing.DiffCreated(2),
		tracing.DiffUpdated(2),
		tracing.DiffDeleted(1),
	)

	fmt.Println("Attributes added to span")
	// Output: Attributes added to span
}
