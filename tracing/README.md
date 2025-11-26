# Tracing Package

OpenTelemetry tracing instrumentation for gitops-repo-api using the functional options pattern.

## Overview

This package provides a lightweight wrapper around OpenTelemetry tracing that integrates seamlessly with the gitops-repo-api codebase. It supports both active tracing (when configured) and no-op tracing (when not configured), ensuring zero performance impact when tracing is disabled.

## Features

- **Functional Options Pattern**: Clean, extensible API using functional options
- **Zero Dependencies When Disabled**: No-op tracer has no performance overhead
- **Context Propagation**: Proper trace context propagation through concurrent operations
- **Standard Attributes**: Pre-defined attribute helpers for consistent naming
- **Backward Compatible**: All existing code continues to work without changes

## Quick Start

### Basic Usage (No Tracing)

By default, all components use a no-op tracer:

```go
import (
    "github.com/codingninja/gitops-repo-api/diff"
    "github.com/codingninja/gitops-repo-api/git"
)

// Works exactly as before - no tracing overhead
preRepo := git.NewRepoSpec("https://github.com/example/repo", nil)
postRepo := git.NewRepoSpec("https://github.com/example/repo", nil)
differ := diff.NewDiffer(preRepo, postRepo, epds)
```

### Enable Tracing

To enable tracing, configure OpenTelemetry and pass a tracer using functional options:

```go
import (
    "go.opentelemetry.io/otel"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    "github.com/codingninja/gitops-repo-api/tracing"
    "github.com/codingninja/gitops-repo-api/diff"
    "github.com/codingninja/gitops-repo-api/git"
)

// Setup OpenTelemetry
tp := sdktrace.NewTracerProvider(
    sdktrace.WithSampler(sdktrace.AlwaysSample()),
    // Add exporters here (e.g., Jaeger, Zipkin, etc.)
)
otel.SetTracerProvider(tp)

// Create wrapped tracer
otelTracer := tp.Tracer("gitops-repo-api")
tracer := tracing.NewTracer(otelTracer)

// Use functional options to enable tracing
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

differ := diff.NewDiffer(
    preRepo,
    postRepo,
    epds,
    diff.WithTracer(tracer),
)
```

## Instrumented Operations

### Git Operations

The `git` package instruments:

- `git.open` - Opening/cloning repositories
- `git.checkout` - Checking out branches
- `git.clone` - Cloning operations

**Attributes**: `git.url`, `git.ref`, `git.branch`, `git.hash`, `git.directory`

### Diff Operations

The `diff` package instruments:

- `gitops.diff` - Root diff operation comparing two refs
- `gitops.extract` - Extracting resources from a single ref
- `gitops.diff.entrypoint` - Diffing a specific entrypoint
- `gitops.diff.discover` - Discovering entrypoints

**Attributes**: `git.pre_ref`, `git.post_ref`, `entrypoint.type`, `entrypoint.directory`, `entrypoint.count`, `diff.count`

### Resource Rendering

The `resource` package instruments:

- `resource.terraform.render` - Terraform plan generation
- `resource.cdk.render` - CDK synthesis
- `resource.kubernetes.render` - Kubernetes manifest rendering

**Attributes**: `resource.type`, `resource.count`, `dir.working`

## Context-Based Tracing

For the resource package, tracing uses context propagation:

```go
import (
    "context"
    "github.com/codingninja/gitops-repo-api/resource"
    "github.com/codingninja/gitops-repo-api/tracing"
)

// Add tracer to context
ctx := resource.ContextWithTracer(context.Background(), tracer)

// Functions automatically use tracer from context
resMap, err := resource.RenderKubernetes(ctx, "/path/to/manifests")
plan, err := resource.RenderTerraform(ctx, "/path/to/terraform")
template, err := resource.RenderCdk(ctx, "/path/to/cdk")
```

## Span Attributes

Standard attributes are provided for consistent naming:

### Git Attributes
```go
tracing.GitURL(url)           // git.url
tracing.GitRef(ref)           // git.ref
tracing.GitBranch(branch)     // git.branch
tracing.GitHash(hash)         // git.hash
tracing.GitDirectory(dir)     // git.directory
tracing.GitPreRef(ref)        // git.pre_ref
tracing.GitPostRef(ref)       // git.post_ref
```

### Entrypoint Attributes
```go
tracing.EntrypointType(type)      // entrypoint.type
tracing.EntrypointDirectory(dir)  // entrypoint.directory
tracing.EntrypointCount(count)    // entrypoint.count
```

### Resource Attributes
```go
tracing.ResourceType(type)      // resource.type
tracing.ResourceCount(count)    // resource.count
tracing.ResourceName(name)      // resource.name
```

### Diff Attributes
```go
tracing.DiffCount(count)        // diff.count
tracing.DiffCreated(count)      // diff.created
tracing.DiffUpdated(count)      // diff.updated
tracing.DiffDeleted(count)      // diff.deleted
tracing.DiffReplaced(count)     // diff.replaced
```

### Directory Attributes
```go
tracing.WorkingDir(dir)     // dir.working
tracing.PreDir(dir)         // dir.pre
tracing.PostDir(dir)        // dir.post
```

## Error Handling

Errors are automatically recorded on spans with proper status codes:

```go
ctx, span := tracer.Start(ctx, "operation")
defer span.End()

result, err := doSomething()
if err != nil {
    span.RecordError(err)
    span.SetStatus(codes.Error, err.Error())
    return err
}
```

## Concurrent Operations

The implementation properly propagates trace context through goroutines:

```go
// Parent span
ctx, span := tracer.Start(ctx, "parent.operation")
defer span.End()

var wg sync.WaitGroup
for _, item := range items {
    wg.Add(1)
    go func(ctx context.Context, item Item) {
        defer wg.Done()
        // Child span - properly linked to parent
        ctx, childSpan := tracer.Start(ctx, "child.operation")
        defer childSpan.End()

        processItem(ctx, item)
    }(ctx, item) // Pass context to preserve trace
}
wg.Wait()
```

## Performance Considerations

- **No-Op Tracer**: Zero overhead when tracing is disabled
- **Lazy Attribute Setting**: Attributes are only computed when spans are active
- **Minimal Allocations**: Efficient attribute creation and span management
- **Async Export**: OpenTelemetry batches span exports asynchronously

## Integration with Observability Platforms

The implementation works with any OpenTelemetry-compatible backend:

### Jaeger
```go
import "go.opentelemetry.io/otel/exporters/jaeger"

exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint("http://localhost:14268/api/traces")))
tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exporter))
```

### Zipkin
```go
import "go.opentelemetry.io/otel/exporters/zipkin"

exporter, err := zipkin.New("http://localhost:9411/api/v2/spans")
tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exporter))
```

### Cloud Providers
- AWS X-Ray
- Google Cloud Trace
- Azure Monitor
- Datadog
- New Relic
- Honeycomb

## Best Practices

1. **Always defer span.End()**: Ensures spans are closed even on panics
2. **Set meaningful attributes**: Help with debugging and filtering traces
3. **Record errors on spans**: Provides visibility into failures
4. **Propagate context**: Pass `ctx` to child operations for proper trace linking
5. **Use standard attributes**: Maintains consistency across the codebase
6. **Sample appropriately**: Configure sampling rates based on traffic volume

## Migration Guide

Existing code requires no changes. To add tracing:

1. Configure OpenTelemetry in your application entry point
2. Create a `tracing.Tracer` instance
3. Pass it via functional options when creating components
4. For resource functions, add tracer to context

Example:

```go
// Before
preRepo := git.NewRepoSpec(url, auth)

// After (with tracing)
preRepo := git.NewRepoSpec(url, auth, git.WithTracer(tracer))

// Before
differ := diff.NewDiffer(preRepo, postRepo, epds)

// After (with tracing)
differ := diff.NewDiffer(preRepo, postRepo, epds, diff.WithTracer(tracer))
```

## Architecture

```
┌─────────────────────────────────────────┐
│   Application (main.go)                 │
│   - Configures OpenTelemetry            │
│   - Creates tracer instances            │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│   tracing.Tracer (wrapper)              │
│   - Wraps OTel trace.Tracer             │
│   - Provides no-op fallback             │
│   - Helper methods                      │
└──────────────┬──────────────────────────┘
               │
    ┌──────────┼──────────┐
    ▼          ▼          ▼
┌────────┐ ┌──────┐  ┌──────────┐
│  git   │ │ diff │  │ resource │
│package │ │package│ │ package  │
└────────┘ └──────┘  └──────────┘
     │         │           │
     └─────────┴───────────┘
               │
               ▼
    ┌────────────────────┐
    │ OpenTelemetry SDK  │
    │ - Sampling         │
    │ - Batching         │
    │ - Export           │
    └────────┬───────────┘
             │
             ▼
    ┌────────────────┐
    │   Backend      │
    │ (Jaeger, etc.) │
    └────────────────┘
```

## License

Same as gitops-repo-api project.
