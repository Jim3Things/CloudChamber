package server

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"

	"go.opentelemetry.io/otel/api/correlation"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/kv"
	"go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/plugin/grpctrace"
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

// There should be an Xxx and Xxxf method for every severity level, plus some
// specific scenario functions (such as OnEnter to log an information entry
// about arrival at a specific method).
//
// Note: The set of methods that are implemented below are based on what is
// currently needed.  Others will be added as required.

// Post a simple informational trace entry
func Info(ctx context.Context, span trace.Span, tick int64, msg string) {
	span.AddEvent(
		ctx,
		tracing.MethodName(2),
		kv.Int64(tracing.StepperTicksKey, tick),
		kv.Int64(tracing.SeverityKey, int64(log.Severity_Info)),
		kv.String(tracing.StackTraceKey, tracing.StackTrace()),
		kv.String(tracing.MessageTextKey, msg))
}

// Post an informational trace entry with complex formatting
func Infof(ctx context.Context, span trace.Span, tick int64, f string, a ...interface{}) {
	span.AddEvent(
		ctx,
		tracing.MethodName(2),
		kv.Int64(tracing.StepperTicksKey, tick),
		kv.Int64(tracing.SeverityKey, int64(log.Severity_Info)),
		kv.String(tracing.StackTraceKey, tracing.StackTrace()),
		kv.String(tracing.MessageTextKey, fmt.Sprintf(f, a...)))
}

// Post a method arrival informational trace entry
func OnEnter(ctx context.Context, span trace.Span, tick int64, msg string) {
	span.AddEvent(
		ctx,
		fmt.Sprintf("On %q entry", tracing.MethodName(2)),
		kv.Int64(tracing.StepperTicksKey, tick),
		kv.Int64(tracing.SeverityKey, int64(log.Severity_Info)),
		kv.String(tracing.StackTraceKey, tracing.StackTrace()),
		kv.String(tracing.MessageTextKey, msg))
}

// Post a simple error trace
func Error(ctx context.Context, span trace.Span, tick int64, a interface{}) error {
	if msg, ok := a.(string); ok {
		return logError(ctx, span, tick, errors.New(msg))
	}

	if err, ok := a.(error); ok {
		return logError(ctx, span, tick, err)
	}

	panic("Invalid Error call - no valid arguments found")
}

// Post an error trace with a complex string formatting
func Errorf(ctx context.Context, span trace.Span, tick int64, f string, a ...interface{}) error {
	return logError(ctx, span, tick, fmt.Errorf(f, a...))
}

// --- Exported trace invocation methods

// +++ Helper functions

// Write a specific error entry
func logError(ctx context.Context, span trace.Span, tick int64, err error) error {
	span.AddEvent(
		ctx,
		fmt.Sprintf("Error from %q", tracing.MethodName(3)),
		kv.Int64(tracing.StepperTicksKey, tick),
		kv.Int64(tracing.SeverityKey, int64(log.Severity_Error)),
		kv.String(tracing.StackTraceKey, tracing.StackTrace()),
		kv.String(tracing.MessageTextKey, err.Error()))

	return err
}

// Return the caller's fully qualified method name
func MethodName(skip int) string {
	fpcs := make([]uintptr, 1)

	// Get the information up the stack (i.e. the caller of this method, or beyond)
	if runtime.Callers(skip + 1, fpcs) == 0 {
		return "?"
	}

	caller := runtime.FuncForPC(fpcs[0] - 1)
	if caller == nil {
		return "?"
	}

	// ... and return the name
	return simpleName(caller.Name())
}

func simpleName(name string) string {
	idx := strings.LastIndex(name, "/")
	if idx >= 0 {
		name = name[idx + 1:]
	}

	return name
}

// --- Helper functions
