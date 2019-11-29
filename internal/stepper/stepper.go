package stepper

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/empty"

	pb "../../pkg/protos/Stepper"

	"google.golang.org/grpc"
)

var policy = pb.StepperPolicy_Invalid
var delay duration.Duration

var syncLock sync.Mutex
var broadcast *sync.Cond

var latest int64 = 0

type server struct {
	pb.UnimplementedStepperServer
}

func waitUntil(atLeast int64) {
	for atLeast > latest {
		func() {
			syncLock.Lock()
			defer syncLock.Unlock()

			if atLeast > latest {
				if policy == pb.StepperPolicy_NoWait {
					latest = atLeast
				} else {
					broadcast.Wait()
				}
			}
		}()
	}
}

func autoStep() {
	time.Sleep(time.Duration(delay.Seconds) * time.Second)
	for policy == pb.StepperPolicy_Measured {
		func() {
			syncLock.Lock()
			defer syncLock.Unlock()
			latest++

			broadcast.Broadcast()
		}()

		time.Sleep(time.Duration(delay.Seconds) * time.Second)
	}
}

// Forcibly reset the stepper to its initial state.  This is used by the unit
// tests to ensure a known starting point.
func Reset() {
	policy = pb.StepperPolicy_Invalid
	delay.Seconds = 0
	latest = 0
}

func (s *server) SetPolicy(ctx context.Context, in *pb.PolicyRequest) (*empty.Empty, error) {
	if policy != pb.StepperPolicy_Invalid {
		// A policy has been set already.  If there is no change, then we can silently
		// ignore this call.  Otherwise, this is an error
		if (policy != in.Policy) || (delay.GetSeconds() != in.MeasuredDelay.GetSeconds()) {
			return nil, fmt.Errorf(
				"stepper already initialized, cannot change setting from %v: %d to %v: %d",
				policy,
				delay.Seconds,
				in.Policy,
				in.MeasuredDelay.GetSeconds())
		}

		return &empty.Empty{}, nil
	}

	// This is an initial policy setup, so make the appropriate change after validating
	// the input.
	switch in.Policy {
	case pb.StepperPolicy_Invalid:
		return nil, fmt.Errorf("stepper policy may not be set to %v", pb.StepperPolicy_Invalid)

	case pb.StepperPolicy_Measured:
		if in.MeasuredDelay.Seconds <= 0 {
			return nil, fmt.Errorf("delay must be greater than zero, but was %d", in.MeasuredDelay.Seconds)
		}

	case pb.StepperPolicy_NoWait, pb.StepperPolicy_Manual:
		if in.MeasuredDelay.Seconds != 0 {
			return nil, fmt.Errorf(
				"delay must be zero when the policy is not %v, but was specified as %v: %d",
				pb.StepperPolicy_Measured,
				in.Policy,
				in.MeasuredDelay.Seconds)
		}

	default:
		return nil, fmt.Errorf("unknown policy specified: %v", in.Policy)
	}

	syncLock.Lock()
	defer syncLock.Unlock()

	broadcast = sync.NewCond(&syncLock)

	policy = in.Policy
	delay = *in.MeasuredDelay

	// If the policy is 'measured', start the recurring auto-stepper go routine
	if policy == pb.StepperPolicy_Measured {
		go autoStep()
	}

	return &empty.Empty{}, nil
}

func (s *server) Step(ctx context.Context, in *empty.Empty) (*empty.Empty, error) {
	if policy == pb.StepperPolicy_Invalid {
		return nil, errors.New("stepper not initialized: no stepper policy has been set")
	}

	if policy != pb.StepperPolicy_Manual {
		return nil, fmt.Errorf(
			"stepper must be using the %v policy.  Currently using %v",
			pb.StepperPolicy_Manual,
			policy)
	}

	syncLock.Lock()
	defer syncLock.Unlock()
	latest++

	broadcast.Broadcast()

	return &empty.Empty{}, nil
}

func (s *server) Now(ctx context.Context, in *empty.Empty) (*pb.TimeResponse, error) {
	if policy == pb.StepperPolicy_Invalid {
		return nil, errors.New("stepper not initialized: no stepper policy has been set")
	}

	syncLock.Lock()
	defer syncLock.Unlock()

	return &pb.TimeResponse{Current: latest}, nil
}

func (s *server) Delay(ctx context.Context, in *pb.DelayRequest) (*pb.TimeResponse, error) {
	if policy == pb.StepperPolicy_Invalid {
		return nil, errors.New("stepper not initialized: no stepper policy has been set")
	}

	if in.AtLeast < 0 {
		return nil, fmt.Errorf("base delay time must be non-negative, was specified as %d", in.AtLeast)
	}

	if in.Jitter < 0 {
		return nil, fmt.Errorf("delay jitter must be non-negative, was specified as %d", in.Jitter)
	}

	var adjust int64 = 0
	if in.Jitter > 0 {
		adjust = rand.Int63n(in.Jitter)
	}

	waitUntil(in.AtLeast + adjust)
	return &pb.TimeResponse{Current: latest}, nil
}

func (s *server) SetToLatest(ctx context.Context, in *pb.SetToLatestRequest) (*pb.TimeResponse, error) {
	if policy == pb.StepperPolicy_Invalid {
		return nil, errors.New("stepper not initialized: no stepper policy has been set")
	}

	if (in.FirstTicks < 0) || (in.SecondTicks < 0) {
		return nil, fmt.Errorf("delay times must be non-negative, were specified as %d, %d", in.FirstTicks, in.SecondTicks)
	}

	waitUntil(in.FirstTicks)
	waitUntil(in.SecondTicks)
	return &pb.TimeResponse{Current: latest}, nil
}

func (s *server) WaitForSync(ctx context.Context, in *pb.WaitForSyncRequest) (*pb.TimeResponse, error) {
	if policy == pb.StepperPolicy_Invalid {
		return nil, errors.New("stepper not initialized: no stepper policy has been set")
	}

	if in.AtLeast < 0 {
		return nil, fmt.Errorf("delay time must be non-negative, was specified as %d", in.AtLeast)
	}

	waitUntil(in.AtLeast)
	return &pb.TimeResponse{Current: latest}, nil
}

func Register(s *grpc.Server) {
	pb.RegisterStepperServer(s, &server{})
}
