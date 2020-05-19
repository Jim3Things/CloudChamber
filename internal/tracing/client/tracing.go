package client

import (
	"context"

	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/plugin/grpctrace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func Interceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	requestMetadata, _ := metadata.FromOutgoingContext(ctx)
	metadataCopy := requestMetadata.Copy()

	tr := global.TraceProvider().Tracer("client")

	ctx, span := tr.Start(ctx, method)
	defer span.End()

	grpctrace.Inject(ctx, &metadataCopy)
	ctx = metadata.NewOutgoingContext(ctx, metadataCopy)

	err := invoker(ctx, method, req, reply, cc, opts...)
	setTraceStatus(span, err)

	return err
}

func setTraceStatus(span trace.Span, err error) {
	// Spans assume a status of "OK", so we only need to update the
	// status if it is an error.
	if err != nil {
		s, ok := status.FromError(err)
		code := s.Code()

		if !ok || code == codes.Unknown {
			code = codes.InvalidArgument
		}

		span.SetStatus(code, "returned error")
	} else {
		span.SetStatus(codes.OK, "OK")
	}
}
