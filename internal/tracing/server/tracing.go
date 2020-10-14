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

	kind := calculateKind(metadataCopy.Get(tracing.InfraSourceKey), trace.SpanKindServer)

	entries, spanCtx := grpctrace.Extract(ctxIn, &metadataCopy)
	ctx := correlation.ContextWithMap(ctxIn, correlation.NewMap(correlation.MapUpdate { MultiKV: entries }))

	tr := global.TraceProvider().Tracer("")

	ctx, span := tr.Start(
		ctx,
		info.FullMethod,
		trace.WithSpanKind(kind),
		trace.WithNewRoot(),
		trace.LinkedTo(spanCtx),
		trace.WithAttributes(kv.String(tracing.StackTraceKey, tracing.StackTrace())),
	)
	defer span.End()

	return handler(ctx, req)
}

// calculateKind returns the span kind to use: either the default one that the
// caller provided, or SpanKindInternal if the supplied infra key value says
// that the grpc client was called from within an infrastructure span.
func calculateKind(values []string, kind trace.SpanKind) trace.SpanKind {
	if len(values) != 1 || values[0] != tracing.IsInfraSource {
		return kind
	}

	return trace.SpanKindInternal
}
