package server

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"

	"go.opentelemetry.io/otel/api/correlation"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/key"
	"go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/plugin/grpctrace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/Jim3Things/CloudChamber/internal/tracing"
)

const (
	StackDepth = 5
)
// Interceptor intercepts and extracts incoming trace data
func Interceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	requestMetadata, _ := metadata.FromIncomingContext(ctx)
	metadataCopy := requestMetadata.Copy()

	entries, spanCtx := grpctrace.Extract(ctx, &metadataCopy)
	ctx = correlation.ContextWithMap(ctx, correlation.NewMap(correlation.MapUpdate{
		MultiKV: entries,
	}))

	stackKey := key.New(tracing.StackTraceKey)

	tr := global.TraceProvider().Tracer("server")

	ctx, span := tr.Start(
		trace.ContextWithRemoteSpanContext(ctx, spanCtx),
		info.FullMethod,
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(stackKey.String(StackTrace())),
	)
	defer span.End()

	return handler(ctx, req)
}

func AddEvent(ctx context.Context, span trace.Span, msg string, tick int64, reason string) {
	ccTickKey := key.New(tracing.StepperTicksKey)
	reasonKey := key.New(tracing.Reason)
	stackKey  := key.New(tracing.StackTraceKey)

	span.AddEvent(
		ctx,
		msg,
		ccTickKey.Int64(tick),
		reasonKey.String(reason),
		stackKey.String(StackTrace()))
}

func logError(ctx context.Context, tick int64, err error) error {
	span := trace.SpanFromContext(ctx)
	ccTickKey := key.New(tracing.StepperTicksKey)
	stackKey  := key.New(tracing.StackTraceKey)

	span.AddEvent(
		ctx,
		err.Error(),
		ccTickKey.Int64(tick),
		stackKey.String(StackTrace()))

	return err
}

func LogError(ctx context.Context, tick int64, a ...interface{}) error {
	switch {
	case len(a) == 1:
		msg, ok := a[0].(string)
		if ok {
			return logError(ctx, tick, errors.New(msg))
		}

		err, ok := a[0].(error)
		if ok {
			return logError(ctx, tick, err)
		}

	case len(a) > 1:
		f, ok := a[0].(string)
		if ok {
			return logError(ctx, tick, fmt.Errorf(f, a[1:]...))
		}
	}

	panic("Invalid LogError call - no valid arguments found")
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

func StackTrace() string {
	res := ""

	fpcs := make([]uintptr, StackDepth)
	runtime.Callers(1, fpcs)
	frames := runtime.CallersFrames(fpcs)

	more := true
	for more {
		var frame runtime.Frame

		frame, more = frames.Next()
		res = fmt.Sprintf("%s\n%s:%d", res, frame.File, frame.Line)
	}

	return res
}