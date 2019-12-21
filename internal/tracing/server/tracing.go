package server

import (
    "context"

    "go.opentelemetry.io/otel/api/distributedcontext"
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
    ctx = distributedcontext.WithMap(ctx, distributedcontext.NewMap(distributedcontext.MapUpdate{
        MultiKV: entries,
    }))

    tr := global.TraceProvider().Tracer("server")

    ctx, span := tr.Start(
        ctx,
        info.FullMethod,
        trace.ChildOf(spanCtx),
        trace.WithSpanKind(trace.SpanKindServer),
    )
    defer span.End()

    return handler(ctx, req)
}

func AddEvent(ctx context.Context, msg string, tick int64, reason string) {
    span := trace.SpanFromContext(ctx)
    ccTickKey := key.New(tracing.StepperTicksKey)
    reasonKey := key.New(tracing.Reason)

    span.AddEvent(ctx, msg, ccTickKey.Int64(tick), reasonKey.String(reason))
}

func LogError(ctx context.Context, tick int64, err error) error {
    span := trace.SpanFromContext(ctx)
    ccTickKey := key.New(tracing.StepperTicksKey)
    span.AddEvent(ctx, err.Error(), ccTickKey.Int64(tick))

    return err
}
