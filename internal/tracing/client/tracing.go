package client

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/api/kv"
	"go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/instrumentation/grpctrace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/Jim3Things/CloudChamber/internal/common"
	"github.com/Jim3Things/CloudChamber/internal/tracing"
	"github.com/Jim3Things/CloudChamber/pkg/protos/log"
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
	requestMetadata, _ := metadata.FromOutgoingContext(ctx)

	metadataCopy := requestMetadata.Copy()
	metadataCopy.Set(tracing.InfraSourceKey, parentIsInfra(ctx))

	grpctrace.Inject(ctx, &metadataCopy)
	ctx = metadata.NewOutgoingContext(ctx, metadataCopy)

	err := invoker(ctx, method, req, reply, cc, opts...)

	sev, resultMsg := decodeGrpcErr(err)

	trace.SpanFromContext(ctx).AddEvent(
		ctx,
		method,
		kv.Int64(tracing.SeverityKey, int64(sev)),
		kv.String(tracing.StackTraceKey, tracing.StackTrace()),
		kv.Int64(tracing.StepperTicksKey, common.TickFromContext(ctx)),
		kv.String(tracing.MessageTextKey,
			fmt.Sprintf("Called %q, GRPC completion status: %s", method, resultMsg)))

	return err
}

func decodeGrpcErr(err error) (log.Severity, string) {
	// Assume success
	sev := log.Severity_Info
	msg := "Success"

	// We have an error, so evaluate what it should be
	if err != nil {
		s, _ := status.FromError(err)
		code := s.Code()

		msg = s.Message()

		if code != codes.OK {
			sev = log.Severity_Error
		}
	}

	return sev, msg
}

// parentIsInfra returns the string value that denotes whether or not the
// current active span is an infrastructure span.
func parentIsInfra(ctx context.Context) string {

	parent := trace.SpanFromContext(ctx)

	if s, ok := parent.(tracing.SpanEx); ok {
		if s.Mask(trace.SpanKindServer) == trace.SpanKindInternal {
			return tracing.IsInfraSource
		}
	}

	return tracing.IsNotInfraSource
}
