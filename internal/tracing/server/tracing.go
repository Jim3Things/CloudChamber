package server

import (
	"context"
	"errors"
	"fmt"
	stdLog "log"

	"go.opentelemetry.io/otel/api/correlation"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/kv"
	"go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/instrumentation/grpctrace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/Jim3Things/CloudChamber/internal/tracing"
	"github.com/Jim3Things/CloudChamber/pkg/protos/log"
)

// Interceptor intercepts and extracts incoming trace data
func Interceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (resp interface{}, err error) {
	requestMetadata, _ := metadata.FromIncomingContext(ctx)
	metadataCopy := requestMetadata.Copy()

	entries, spanCtx := grpctrace.Extract(ctx, &metadataCopy)
	ctx = correlation.ContextWithMap(ctx, correlation.NewMap(correlation.MapUpdate{
		MultiKV: entries,
	}))

	tr := global.TraceProvider().Tracer("server")

	ctx, span := tr.Start(
		trace.ContextWithRemoteSpanContext(ctx, spanCtx),
		info.FullMethod,
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(kv.String(tracing.StackTraceKey, tracing.StackTrace())),
	)
	defer span.End()

	return handler(ctx, req)
}

// +++ Exported trace invocation methods

// WithSpan executes the supplied function within a span that conforms to the
// expected tracing pattern
func WithSpan(ctx context.Context, spanName string, fn func(ctx context.Context) error) error {
	tr := global.TraceProvider().Tracer("server")

	return tr.WithSpan(ctx, spanName, fn,
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(kv.String(tracing.StackTraceKey, tracing.StackTrace())))
}

// WithInfraSpan is a variant of WithSpan that marks the span as internal to
// the simulation, and therefore of limited interest outside of development
// or debugging use.
func WithInfraSpan(ctx context.Context, spanName string, fn func(ctx context.Context) error) error {
	tr := global.TraceProvider().Tracer("server")

	return tr.WithSpan(ctx, spanName, fn,
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(kv.String(tracing.StackTraceKey, tracing.StackTrace())))
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
		tracing.MethodName(2),
		kv.Int64(tracing.StepperTicksKey, tick),
		kv.Int64(tracing.SeverityKey, int64(log.Severity_Info)),
		kv.String(tracing.StackTraceKey, tracing.StackTrace()),
		kv.String(tracing.MessageTextKey, msg))
}

// Infof posts an informational trace event with complex formatting
func Infof(ctx context.Context, tick int64, f string, a ...interface{}) {
	trace.SpanFromContext(ctx).AddEvent(
		ctx,
		tracing.MethodName(2),
		kv.Int64(tracing.StepperTicksKey, tick),
		kv.Int64(tracing.SeverityKey, int64(log.Severity_Info)),
		kv.String(tracing.StackTraceKey, tracing.StackTrace()),
		kv.String(tracing.MessageTextKey, fmt.Sprintf(f, a...)))
}

// OnEnter posts an informational trace event describing the entry into a
// function
func OnEnter(ctx context.Context, tick int64, msg string) {
	trace.SpanFromContext(ctx).AddEvent(
		ctx,
		fmt.Sprintf("On %q entry", tracing.MethodName(2)),
		kv.Int64(tracing.StepperTicksKey, tick),
		kv.Int64(tracing.SeverityKey, int64(log.Severity_Info)),
		kv.String(tracing.StackTraceKey, tracing.StackTrace()),
		kv.String(tracing.MessageTextKey, msg))
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
	stdLog.Fatal(Errorf(ctx, tick, f, a...))
}

// --- Exported trace invocation methods

// +++ Helper functions

// logError writes a specific error trace event
func logError(ctx context.Context, tick int64, err error) error {
	trace.SpanFromContext(ctx).AddEvent(
		ctx,
		fmt.Sprintf("Error from %q", tracing.MethodName(3)),
		kv.Int64(tracing.StepperTicksKey, tick),
		kv.Int64(tracing.SeverityKey, int64(log.Severity_Error)),
		kv.String(tracing.StackTraceKey, tracing.StackTrace()),
		kv.String(tracing.MessageTextKey, err.Error()))

	return err
}

// --- Helper functions
