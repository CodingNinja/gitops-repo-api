package tracing

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

// Tracer is a wrapper around OpenTelemetry's trace.Tracer interface
// that provides a simplified API for creating and managing spans.
type Tracer struct {
	tracer trace.Tracer
}

// NewTracer creates a new Tracer instance wrapping the provided trace.Tracer.
// If tracer is nil, a no-op tracer is used, making tracing optional.
func NewTracer(tracer trace.Tracer) *Tracer {
	if tracer == nil {
		tracer = noop.NewTracerProvider().Tracer("")
	}
	return &Tracer{tracer: tracer}
}

// Start creates a new span and returns both the updated context and the span.
// The span must be ended with span.End() when the operation completes.
// It's idiomatic to defer span.End() immediately after calling Start.
//
// Example:
//
//	ctx, span := tracer.Start(ctx, "operation.name")
//	defer span.End()
func (t *Tracer) Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, spanName, opts...)
}

// Span is a convenience wrapper around trace.Span that provides helper methods
// for common tracing operations.
type Span struct {
	trace.Span
}

// RecordError records an error on the span and sets the span status to error.
// This is a convenience method that combines RecordError and SetStatus.
func (s Span) RecordError(err error) {
	if err == nil {
		return
	}
	s.Span.RecordError(err)
	s.Span.SetStatus(codes.Error, err.Error())
}

// SetAttributes sets multiple attributes on the span at once.
func (s Span) SetAttributes(attrs ...attribute.KeyValue) {
	s.Span.SetAttributes(attrs...)
}

// SetStatus sets the status of the span.
func (s Span) SetStatus(code codes.Code, description string) {
	s.Span.SetStatus(code, description)
}

// StartSpan is a helper function that creates a span and returns a wrapped Span.
// This provides access to the helper methods defined on Span.
func (t *Tracer) StartSpan(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, Span) {
	ctx, span := t.tracer.Start(ctx, spanName, opts...)
	return ctx, Span{Span: span}
}

// NoOpTracer returns a tracer that performs no operations.
// This is useful for testing or when tracing is disabled.
func NoOpTracer() *Tracer {
	return NewTracer(nil)
}
