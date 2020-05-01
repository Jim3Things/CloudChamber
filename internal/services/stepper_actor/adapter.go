package stepper

import (
    "context"
    "fmt"
    "time"

    "github.com/AsynkronIT/protoactor-go/actor"
    "github.com/golang/protobuf/ptypes/empty"

    pb "github.com/Jim3Things/CloudChamber/pkg/protos/Stepper"
    "github.com/Jim3Things/CloudChamber/pkg/protos/common"
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

// +++ GRPC Methods

// The following functions are the grpc method overrides. Each takes the
// request argument it receives, sends it to the attached actor and waits
// for the response.  This is itself a message, so it is analyzed and
// converted into an error, if appropriate.  The final result is then
// returned to the grpc caller.

func (s *server) SetPolicy(ctx context.Context, in *pb.PolicyRequest) (res *empty.Empty, err error) {
    c := actor.EmptyRootContext
    _, err = msgToError(c.RequestFuture(s.pid, in, ActorTimeout).Result())
    if err != nil {
        return nil, err
    }

    return &empty.Empty{}, nil
}

func (s *server) Step(ctx context.Context, in *pb.StepRequest) (*empty.Empty, error) {
    c := actor.EmptyRootContext
    _, err := msgToError(c.RequestFuture(s.pid, in, ActorTimeout).Result())
    if err != nil {
        return nil, err
    }

    return &empty.Empty{}, nil
}

func (s *server) Now(ctx context.Context, in *pb.NowRequest) (*common.Timestamp, error) {
    c := actor.EmptyRootContext
    res, err := msgToError(c.RequestFuture(s.pid, in, ActorTimeout).Result())
    if err != nil {
        return nil, err
    }

    return res.(*common.Timestamp), nil
}

func (s *server) Delay(ctx context.Context, in *pb.DelayRequest) (*common.Timestamp, error) {
    c := actor.EmptyRootContext
    res, err := msgToError(c.RequestFuture(s.pid, in, NoTimeout).Result())
    if err != nil {
        return nil, err
    }

    return res.(*common.Timestamp), nil
}

func (s *server) Reset(ctx context.Context, in *pb.ResetRequest) (*empty.Empty, error) {
    c := actor.EmptyRootContext
    res, err := msgToError(c.RequestFuture(s.pid, in, ActorTimeout).Result())
    if err != nil {
        return nil, err
    }

    return res.(*empty.Empty), nil
}

// --- GRPC Methods

// Convert a completion message body into an equivalent error, if needed.
func msgToError(msg interface{}, err error) (interface{}, error) {
    if err == nil {
        v, ok := msg.(*common.Completion)
        if ok && v.IsError {
            return nil, fmt.Errorf("%s", v.Error)
        }
    }

    return msg, err
}
