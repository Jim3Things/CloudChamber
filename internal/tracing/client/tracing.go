package client

import (
	"context"

	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/kv"
	"go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/instrumentation/grpctrace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/Jim3Things/CloudChamber/internal/tracing"
)

// Interceptor is a function that traces the client side activity for a grpc
// call.
func Interceptor(
	ctx context.Context,
	method string,
	req,
	reply interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption) error {
	return commonInterceptor(
		ctx, method, req, reply, cc, invoker,
		trace.SpanKindClient,
		opts...)
}

// InfraInterceptor is a function that traces the client side activity for a
// grpc infrastructure call.  These traces use an 'internal' span designator,
// so that they can be filtered during display later.
func InfraInterceptor(
	ctx context.Context,
	method string,
	req,
	reply interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption) error {
	return commonInterceptor(
		ctx, method, req, reply, cc, invoker,
		trace.SpanKindInternal,
		opts...)
}

// commonInterceptor performs the logic to encase the grpc call itself in a
// trace span, and to record the final status.
func commonInterceptor(
	ctxIn context.Context,
	method string,
	req,
	reply interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	kind trace.SpanKind,
	opts ...grpc.CallOption) error {
	requestMetadata, _ := metadata.FromOutgoingContext(ctxIn)
	metadataCopy := requestMetadata.Copy()

	parent := trace.SpanFromContext(ctxIn)

	tr := global.TraceProvider().Tracer("client")

	ctx, span := tr.Start(ctxIn, method,
		trace.WithSpanKind(kind),
		trace.WithAttributes(kv.String(tracing.StackTraceKey, tracing.StackTrace())))
	defer span.End()

	if parent.SpanContext().HasSpanID() {
		parent.AddEvent(
			ctx,
			tracing.MethodName(2),
			kv.String(tracing.StackTraceKey, tracing.StackTrace()),
			kv.String(tracing.ChildSpanKey, span.SpanContext().SpanID.String()))
	}

	grpctrace.Inject(ctx, &metadataCopy)
	ctx = metadata.NewOutgoingContext(ctx, metadataCopy)

	err := invoker(ctx, method, req, reply, cc, opts...)
	setTraceStatus(ctx, span, err)

	return err
}

// setTraceStatus records the final status for a trace span.
func setTraceStatus(ctx context.Context, span trace.Span, err error) {
	// Spans assume a status of "OK", so we only need to update the
	// status if it is an error.
	if err != nil {
		s, ok := status.FromError(err)
		code := s.Code()

		if !ok || code == codes.Unknown {
			code = codes.InvalidArgument
		}

		span.RecordError(ctx, err, trace.WithErrorStatus(code))
	} else {
		span.SetStatus(codes.OK, "OK")
	}
}
