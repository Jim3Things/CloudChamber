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

// Interceptor intercepts and extracts incoming trace data
func Interceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
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
	)
	defer span.End()

	return handler(ctx, req)
}

func AddEvent(ctx context.Context, msg string, tick int64, reason string) {
	c := ctx
	if c == nil { c = context.Background() }

	span := trace.SpanFromContext(c)
	ccTickKey := key.New(tracing.StepperTicksKey)
	reasonKey := key.New(tracing.Reason)

	span.AddEvent(c, msg, ccTickKey.Int64(tick), reasonKey.String(reason))
}

func logError(ctx context.Context, tick int64, err error) error {
	c := ctx
	if c == nil { c = context.Background() }

	span := trace.SpanFromContext(c)
	ccTickKey := key.New(tracing.StepperTicksKey)
	span.AddEvent(c, err.Error(), ccTickKey.Int64(tick))

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