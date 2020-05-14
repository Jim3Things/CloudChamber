package client

import (
	"context"

	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/kv"
	"go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/plugin/grpctrace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/Jim3Things/CloudChamber/internal/tracing"
)

func Interceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	requestMetadata, _ := metadata.FromOutgoingContext(ctx)
	metadataCopy := requestMetadata.Copy()

	tr := global.TraceProvider().Tracer("client")
	err := tr.WithSpan(ctx, method,
		func(ctx context.Context) error {
			grpctrace.Inject(ctx, &metadataCopy)
			ctx = metadata.NewOutgoingContext(ctx, metadataCopy)

			err := invoker(ctx, method, req, reply, cc, opts...)
			setTraceStatus(ctx, err)
			return err
		},
		trace.WithAttributes(kv.String(tracing.StackTraceKey, tracing.StackTrace())))
	return err
}

func setTraceStatus(ctx context.Context, err error) {
	// Spans assume a status of "OK", so we only need to update the
	// status if it is an error.
	if err != nil {
		s, _ := status.FromError(err)
		trace.SpanFromContext(ctx).SetStatus(s.Code(), "Setting completion status")
	}
}
