package server

import (
	"context"

	"go.opentelemetry.io/otel/api/correlation"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/kv"
	"go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/instrumentation/grpctrace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/Jim3Things/CloudChamber/internal/tracing"
)

// Interceptor intercepts and extracts incoming trace data
func Interceptor(
	ctxIn context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (resp interface{}, err error) {
	requestMetadata, _ := metadata.FromIncomingContext(ctxIn)
	metadataCopy := requestMetadata.Copy()

	entries, spanCtx := grpctrace.Extract(ctxIn, &metadataCopy)
	ctx := correlation.ContextWithMap(ctxIn, correlation.NewMap(correlation.MapUpdate{
		MultiKV: entries,
	}))

	tr := global.TraceProvider().Tracer("")

	ctx, span := tr.Start(
		ctxIn,
		info.FullMethod,
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithNewRoot(),
		trace.LinkedTo(spanCtx),
		trace.WithAttributes(kv.String(tracing.StackTraceKey, tracing.StackTrace())),
	)
	defer span.End()

	return handler(ctx, req)
}

// +++ Exported trace invocation methods

// WithSpan executes the supplied function within a span that conforms to the
// expected tracing pattern.
func WithSpan(ctx context.Context, fn func(ctx context.Context) error) error {
	return doSpan(ctx, tracing.MethodName(2), trace.SpanKindServer, fn)
}

// WithNamedSpan executes the supplied function within a span that conforms to
// the expected tracing pattern, and with a custom span name.
func WithNamedSpan(ctx context.Context, spanName string, fn func(ctx context.Context) error) error {
	return doSpan(ctx, spanName, trace.SpanKindServer, fn)
}

func doSpan(ctxIn context.Context, spanName string, spanKind trace.SpanKind, fn func(ctx context.Context) error) error {
	parent := trace.SpanFromContext(ctxIn)

	tr := global.TraceProvider().Tracer("")

	ctx, span := tr.Start(ctxIn, spanName,
		trace.WithSpanKind(spanKind),
		trace.WithAttributes(kv.String(tracing.StackTraceKey, tracing.StackTrace())))
	defer span.End()

	if parent.SpanContext().HasSpanID() {
		parent.AddEvent(
			ctx,
			tracing.MethodName(2),
			kv.String(tracing.StackTraceKey, tracing.StackTrace()),
			kv.String(tracing.ChildSpanKey, span.SpanContext().SpanID.String()))
	}

	return fn(ctx)
}

