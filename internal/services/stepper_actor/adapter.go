package stepper

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/empty"
	"go.opentelemetry.io/otel/api/trace"

	"github.com/Jim3Things/CloudChamber/internal/tracing"
	"github.com/Jim3Things/CloudChamber/pkg/protos/common"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/services"
)

const (
	// This is the standard timeout given for the synchronous processing of the
	// local actor.  This is used for almost all calls.
	ActorTimeout = 1 * time.Second

	// This is the value used to indicate no timeout.  This is used for Delay
	// processing, as that call will wait for an indefinite time (until enough
	// simulated time passes)
	NoTimeout = -1 * time.Second
)

// Define the grpc server that is used solely as an adapter to an attached
// stepper service actor.
type server struct {
	pb.UnimplementedStepperServer

	// Attached actor's ID
	pid *actor.PID
}

// Attach an actor to this grpc adapter
func (s *server) Attach(pid *actor.PID) {
	s.pid = pid
}

// Set the default policy.  Note that this is a direct call, not one that
// passes through the grpc listener.
func (s *server) SetDefaultPolicy(p pb.StepperPolicy) error {
	delay := &duration.Duration{Seconds: 1}
	if p != pb.StepperPolicy_Measured {
		delay = &duration.Duration{Seconds: 0}
	}

	in := &pb.PolicyRequest{
		Policy:        p,
		MeasuredDelay: delay,
	}

	_, err := msgToError(actorContext(context.Background()).RequestFuture(s.pid, in, ActorTimeout).Result())

	return err
}

// +++ GRPC Methods

// The following functions are the grpc method overrides. Each takes the
// request argument it receives, sends it to the attached actor and waits
// for the response.  This is itself a message, so it is analyzed and
// converted into an error, if appropriate.  The final result is then
// returned to the grpc caller.

func (s *server) SetPolicy(ctx context.Context, in *pb.PolicyRequest) (*empty.Empty, error) {
	if err := in.Validate(); err != nil {
		return nil, err
	}

	return asEmpty(msgToError(actorContext(ctx).RequestFuture(s.pid, in, ActorTimeout).Result()))
}

func (s *server) Step(ctx context.Context, in *pb.StepRequest) (*empty.Empty, error) {
	if err := in.Validate(); err != nil {
		return nil, err
	}

	return asEmpty(msgToError(actorContext(ctx).RequestFuture(s.pid, in, ActorTimeout).Result()))
}

func (s *server) Now(ctx context.Context, in *pb.NowRequest) (*common.Timestamp, error) {
	if err := in.Validate(); err != nil {
		return nil, err
	}

	return asTimestamp(msgToError(actorContext(ctx).RequestFuture(s.pid, in, ActorTimeout).Result()))
}

func (s *server) Delay(ctx context.Context, in *pb.DelayRequest) (*common.Timestamp, error) {
	if err := in.Validate(); err != nil {
		return nil, err
	}

	return asTimestamp(msgToError(actorContext(ctx).RequestFuture(s.pid, in, NoTimeout).Result()))
}

func (s *server) Reset(ctx context.Context, in *pb.ResetRequest) (*empty.Empty, error) {
	if err := in.Validate(); err != nil {
		return nil, err
	}

	return asEmpty(msgToError(actorContext(ctx).RequestFuture(s.pid, in, ActorTimeout).Result()))
}

func (s *server) GetStatus(ctx context.Context, in *pb.GetStatusRequest) (*pb.StatusResponse, error) {
	if err := in.Validate(); err != nil {
		return nil, err
	}

	return asStatusResponse(msgToError(actorContext(ctx).RequestFuture(s.pid, in, ActorTimeout).Result()))
}

// --- GRPC Methods

// +++ Helper functions

// Convert the return pair into (Timestamp, error) types
func asTimestamp(res interface{}, err error) (*common.Timestamp, error) {
	if err == nil {
		return res.(*common.Timestamp), err
	}

	return nil, err
}

// Convert the return pair into (Empty, error) types
func asEmpty(res interface{}, err error) (*empty.Empty, error) {
	if err == nil {
		return res.(*empty.Empty), err
	}

	return nil, err
}

// Convert the return pair into (StatusResposne, error) types
func asStatusResponse(res interface{}, err error) (*pb.StatusResponse, error) {
	if err == nil {
		return res.(*pb.StatusResponse), err
	}

	return nil, err
}

// Convert a completion message body into an equivalent error, if needed.
func msgToError(msg interface{}, err error) (interface{}, error) {
	if err == nil {
		v, ok := msg.(*common.Completion)
		if ok {
			return nil, fmt.Errorf("%s", v.Error)
		}
	}

	return msg, err
}

// Get the current span context and encode it into the outgoing message header
func actorContext(ctx context.Context) *actor.RootContext {
	spanContext := trace.SpanFromContext(ctx).SpanContext()
	flags := strconv.Itoa(int(spanContext.TraceFlags))

	return actor.NewRootContext(nil).
		WithSenderMiddleware(
			func(next actor.SenderFunc) actor.SenderFunc {
				return func(ctx actor.SenderContext, target *actor.PID, envelope *actor.MessageEnvelope) {
					envelope.SetHeader(tracing.SourceTraceID, spanContext.TraceID.String())
					envelope.SetHeader(tracing.SourceSpanID, spanContext.SpanID.String())
					envelope.SetHeader(tracing.SourceTraceFlgs, flags)

					next(ctx, target, envelope)
				}
			})

}
