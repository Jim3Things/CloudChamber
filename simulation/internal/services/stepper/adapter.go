package stepper

// This file contains the grpc stepper service, which converts the calls to
// internal messages, sends them along to the state machine for processing,
// and finally converts the response message back to a grpc result

import (
	"context"
	"io"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"

	"github.com/Jim3Things/CloudChamber/simulation/internal/sm"
	pb "github.com/Jim3Things/CloudChamber/simulation/pkg/protos/services"
)

type server struct {
	pb.UnimplementedStepperServer

	impl *stepper
}

func Register(ctx context.Context, s *grpc.Server, starting pb.StepperPolicy) error {
	svc := &server{
		UnimplementedStepperServer: pb.UnimplementedStepperServer{},
		impl:                       newStepper(convertToInternalPolicy(starting)),
	}

	if err := svc.impl.start(ctx); err != nil {
		return err
	}

	pb.RegisterStepperServer(s, svc)
	return nil
}

func (s *server) SetPolicy(ctx context.Context, in *pb.PolicyRequest) (*pb.StatusResponse, error) {
	if err := in.Validate(); err != nil {
		return nil, err
	}

	rsp, err := s.invoke(ctx, in)
	if err != nil {
		return nil, err
	}

	return toExternalStatusResponse(rsp)
}

func (s *server) Step(ctx context.Context, in *pb.StepRequest) (*pb.StatusResponse, error) {
	if err := in.Validate(); err != nil {
		return nil, err
	}

	rsp, err := s.invoke(ctx, in)
	if err != nil {
		return nil, err
	}

	return toExternalStatusResponse(rsp)
}

func (s *server) Delay(ctx context.Context, in *pb.DelayRequest) (*pb.StatusResponse, error) {
	if err := in.Validate(); err != nil {
		return nil, err
	}

	rsp, err := s.invoke(ctx, in)
	if err != nil {
		return nil, err
	}

	return toExternalStatusResponse(rsp)
}

func (s *server) Reset(ctx context.Context, in *pb.ResetRequest) (*empty.Empty, error) {
	if err := in.Validate(); err != nil {
		return nil, err
	}

	rsp, err := s.invoke(ctx, in)
	if err != nil {
		return nil, err
	}

	return toExternalEmptyResponse(rsp)
}

func (s *server) GetStatus(ctx context.Context, in *pb.GetStatusRequest) (*pb.StatusResponse, error) {
	if err := in.Validate(); err != nil {
		return nil, err
	}

	rsp, err := s.invoke(ctx, in)
	if err != nil {
		return nil, err
	}

	return toExternalStatusResponse(rsp)
}

func (s *server) invoke(ctx context.Context, in interface{}) (*sm.Response, error) {
	ch := make(chan *sm.Response)

	msg, err := toInternal(ctx, in, ch)
	if err != nil {
		return nil, err
	}

	s.impl.Receive(msg)
	rsp, more := <-ch
	if !more {
		return nil, io.ErrClosedPipe
	}

	return rsp, nil
}
