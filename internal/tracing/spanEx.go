package tracing

import (
	"context"
	"errors"
	"fmt"
	log2 "log"

	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/kv"
	"go.opentelemetry.io/otel/api/trace"

	"github.com/Jim3Things/CloudChamber/pkg/protos/log"
)

// spanEx extends the OpenTelemetry Span structure with additional features
// needed by CloudChamber
type spanEx struct {
	// Original OpenTelemetry span
	trace.Span

	// isInternal is true if this span started with SpanKindInternal.
	isInternal bool
}

// Mask applies the policy that a child span of a span that is marked as
// infrastructure (span kind is internal) should also be market as internal.
func (s *spanEx) Mask(kind trace.SpanKind) trace.SpanKind {
	if s.isInternal {
		return trace.SpanKindInternal
	}

	return kind
}

// startSpanConfig defines the attributes used when starting a new span, and
// that can be overridden by the caller.
type startSpanConfig struct {
	name       string
	kind       trace.SpanKind
	stackTrace string
	tick       int64
}

// StartSpanOption denotes optional decoration methods used on StartSpan
type StartSpanOption func(*startSpanConfig)

// WithName adds the supplied value as the name of the span under creation
func WithName(name string) StartSpanOption {
	return func(cfg *startSpanConfig) {
		cfg.name = name
	}
}

// WithKind sets the span kind for the span under creation.  This will be
// overridden if the parent span is of SpanKindInternal.
func WithKind(kind trace.SpanKind) StartSpanOption {
	return func(cfg *startSpanConfig) {
		cfg.kind = kind
	}
}

// WithTick sets the start time for the span to a specific simulated time.
func WithTick(tick int64) StartSpanOption {
	return func(cfg *startSpanConfig) {
		cfg.tick = tick
	}
}

// AsInternal sets the span as an infrastructure (or internal) span kind.
func AsInternal() StartSpanOption {
	return func(cfg *startSpanConfig) {
		cfg.kind = trace.SpanKindInternal
	}
}

// StartSpan creates a new tracing span, with the attributes and linkages
// that the Cloud Chamber logging system expects
func StartSpan(
	ctx context.Context,
	options ...StartSpanOption) (context.Context, trace.Span) {
	cfg := startSpanConfig{
		name:       MethodName(2),
		kind:       trace.SpanKindServer,
		stackTrace: StackTrace(),
		tick:       -1,
	}

	for _, opt := range options {
		opt(&cfg)
	}

	parent := trace.SpanFromContext(ctx)

	if spanEx, ok := parent.(spanEx); ok {
		cfg.kind = spanEx.Mask(cfg.kind)
	}

	tr := global.TraceProvider().Tracer("")

	ctxChild, span := tr.Start(ctx, cfg.name,
		trace.WithSpanKind(cfg.kind),
		trace.WithAttributes(kv.String(StackTraceKey, cfg.stackTrace)))

	if parent.SpanContext().HasSpanID() {
		parent.AddEvent(
			ctxChild,
			cfg.name,
			kv.Int64(StepperTicksKey, cfg.tick),
			kv.Int64(SeverityKey, int64(log.Severity_Info)),
			kv.String(StackTraceKey, StackTrace()),
			kv.String(ChildSpanKey, span.SpanContext().SpanID.String()))
	}

	ccSpan := spanEx{
		Span:       span,
		isInternal: cfg.kind == trace.SpanKindInternal,
	}

	return trace.ContextWithSpan(ctxChild, ccSpan), ccSpan
}

// There should be an Xxx and Xxxf method for every severity level, plus some
// specific scenario functions (such as OnEnter to log an information entry
// about arrival at a specific method).
//
// Note: The set of methods that are implemented below are based on what is
// currently needed.  Others will be added as required.

// Info posts a simple informational trace event
func Info(ctx context.Context, tick int64, msg string) {
	trace.SpanFromContext(ctx).AddEvent(
		ctx,
		MethodName(2),
		kv.Int64(StepperTicksKey, tick),
		kv.Int64(SeverityKey, int64(log.Severity_Info)),
		kv.String(StackTraceKey, StackTrace()),
		kv.String(MessageTextKey, msg))
}

// Infof posts an informational trace event with complex formatting
func Infof(ctx context.Context, tick int64, f string, a ...interface{}) {
	trace.SpanFromContext(ctx).AddEvent(
		ctx,
		MethodName(2),
		kv.Int64(StepperTicksKey, tick),
		kv.Int64(SeverityKey, int64(log.Severity_Info)),
		kv.String(StackTraceKey, StackTrace()),
		kv.String(MessageTextKey, fmt.Sprintf(f, a...)))
}

// OnEnter posts an informational trace event describing the entry into a
// function
func OnEnter(ctx context.Context, tick int64, msg string) {
	trace.SpanFromContext(ctx).AddEvent(
		ctx,
		fmt.Sprintf("On %q entry", MethodName(2)),
		kv.Int64(StepperTicksKey, tick),
		kv.Int64(SeverityKey, int64(log.Severity_Info)),
		kv.String(StackTraceKey, StackTrace()),
		kv.String(MessageTextKey, msg))
}

// Error posts a simple error trace event
func Error(ctx context.Context, tick int64, a interface{}) error {
	if msg, ok := a.(string); ok {
		return logError(ctx, tick, errors.New(msg))
	}

	if err, ok := a.(error); ok {
		return logError(ctx, tick, err)
	}

	panic("Invalid Error call - no valid arguments found")
}

// Errorf posts an error trace event with a complex string formatting
func Errorf(ctx context.Context, tick int64, f string, a ...interface{}) error {
	return logError(ctx, tick, fmt.Errorf(f, a...))
}

// Fatalf traces the error, and then terminates the process.
func Fatalf(ctx context.Context, tick int64, f string, a ...interface{}) {
	log2.Fatal(Errorf(ctx, tick, f, a...))
}

// --- Exported trace invocation methods

// +++ Helper functions

// logError writes a specific error trace event
func logError(ctx context.Context, tick int64, err error) error {
	trace.SpanFromContext(ctx).AddEvent(
		ctx,
		fmt.Sprintf("Error from %q", MethodName(3)),
		kv.Int64(StepperTicksKey, tick),
		kv.Int64(SeverityKey, int64(log.Severity_Error)),
		kv.String(StackTraceKey, StackTrace()),
		kv.String(MessageTextKey, err.Error()))

	return err
}

// --- Helper functions
