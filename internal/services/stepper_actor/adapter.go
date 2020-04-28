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
    NoTimeout = -1 * time.Second
)

// Define the skeleton grpc server
type server struct {
    pb.UnimplementedStepperServer
}

func (s *Actor) SetPolicy(ctx context.Context, in *pb.PolicyRequest) (res *empty.Empty, err error) {
    c := actor.EmptyRootContext
    _, err = c.RequestFuture(s.pid, in, ActorTimeout).Result()
    if err != nil {
        return nil, err
    }

    return &empty.Empty{}, nil
}

func (s *Actor) Step(ctx context.Context, in *pb.StepRequest) (*empty.Empty, error) {
    c := actor.EmptyRootContext
    _, err := msgToError(c.RequestFuture(s.pid, in, ActorTimeout).Result())
    if err != nil {
        return nil, err
    }

    return &empty.Empty{}, nil
}

func (s *Actor) Now(ctx context.Context, in *pb.NowRequest) (*common.Timestamp, error) {
    c := actor.EmptyRootContext
    res, err := msgToError(c.RequestFuture(s.pid, in, ActorTimeout).Result())
    if err != nil {
        return nil, err
    }

    return res.(*common.Timestamp), nil
}

func (s *Actor) Delay(ctx context.Context, in *pb.DelayRequest) (*common.Timestamp, error) {
    c := actor.EmptyRootContext
    res, err := msgToError(c.RequestFuture(s.pid, in, NoTimeout).Result())
    if err != nil {
        return nil, err
    }

    return res.(*common.Timestamp), nil
}

func msgToError(msg interface{}, err error) (interface{}, error) {
    if err == nil {
        v, ok := msg.(common.Completion)
        if ok && v.IsError {
            return nil, fmt.Errorf("%s", v.Error)
        }
    }

    return msg, err
}
