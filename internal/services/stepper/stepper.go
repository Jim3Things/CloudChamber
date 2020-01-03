// This module contains the implementation for the simulated time management
// features embodied in the stepper service.

package stepper

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"

	trace "github.com/Jim3Things/CloudChamber/internal/tracing/server"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/Stepper"
)

var policy = pb.StepperPolicy_Invalid // Active stepper policy
var delay duration.Duration           // Time between ticks, iff Measured policy

var syncLock sync.Mutex  // Access lock for current simulated time
var broadcast *sync.Cond // Broadcast channel for time change notification

var latest int64 = 0 // Current simulated time

// Define the skeleton grpc server
type server struct {
	pb.UnimplementedStepperServer
}

// Register this with grpc as the stepper service.
func Register(s *grpc.Server) {
	pb.RegisterStepperServer(s, &server{})
}

// Wait until the local simulated time is at least as late as the supplied
// target time.  This routine handles all the waiting variances from the
// different policies
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

// When in the 'Measured' policy, this routine waits the required
// amount of time and then automatically executes a 'step'.
//
// Note that it checks the current policy in case it changed.  That is not
// expected to happen in normal use, but does in several unit tests.  This
// check allows those tests to run in a single combined test suite.
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

// The remaining methods are the implementations for the stepper protocol grpc
// class.  See ../../../pkg/stepper.proto for the interface details.

// Set the stepper's policy governing the rate and conditions for the simulated
// time to move forward.
func (s *server) SetPolicy(ctx context.Context, in *pb.PolicyRequest) (*empty.Empty, error) {
	trace.AddEvent(ctx, in.String(), latest, "Setting the policy")

	if policy != pb.StepperPolicy_Invalid {
		// A policy has been set already.  If there is no change, then we can silently
		// ignore this call.  Otherwise, this is an error
		if (policy != in.Policy) || (delay.GetSeconds() != in.MeasuredDelay.GetSeconds()) {
			return nil, trace.LogError(ctx, latest,
				"stepper already initialized, cannot change setting from %v: %d to %v: %d",
				policy,
				delay.Seconds,
				in.Policy,
				in.MeasuredDelay.GetSeconds())
		}

		// The current policy is exactly the same as the new one - so silently ignore.
		return &empty.Empty{}, nil
	}

	// This is an initial policy setup, so make the appropriate change after validating
	// the input.
	switch in.Policy {
	case pb.StepperPolicy_Invalid:
		return nil, trace.LogError(ctx, latest, "stepper policy may not be set to %v", pb.StepperPolicy_Invalid)

	case pb.StepperPolicy_Measured:
		if in.MeasuredDelay.Seconds <= 0 {
			return nil, trace.LogError(ctx, latest, "delay must be greater than zero, but was %d", in.MeasuredDelay.Seconds)
		}

	case pb.StepperPolicy_NoWait, pb.StepperPolicy_Manual:
		if in.MeasuredDelay.Seconds != 0 {
			return nil, trace.LogError(ctx, latest,
				"delay must be zero when the policy is not %v, but was specified as %v: %d",
				pb.StepperPolicy_Measured,
				in.Policy,
				in.MeasuredDelay.Seconds)
		}

	default:
		return nil, trace.LogError(ctx, latest, "unknown policy specified: %v", in.Policy)
	}

	// We have a new, valid policy.  Set it up.
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

// When the stepper policy is for manual single-stepping, this function forces
// a single step forward in simulated time.
func (s *server) Step(ctx context.Context, _ *empty.Empty) (*empty.Empty, error) {
	trace.AddEvent(ctx, "", latest, "Single stepping time")

	if policy == pb.StepperPolicy_Invalid {
		return nil, trace.LogError(ctx, latest, "stepper not initialized: no stepper policy has been set")
	}

	if policy != pb.StepperPolicy_Manual {
		return nil, trace.LogError(ctx, latest,
			"stepper must be using the %v policy.  Currently using %v",
			pb.StepperPolicy_Manual,
			policy)
	}

	syncLock.Lock()
	defer syncLock.Unlock()
	latest++

	broadcast.Broadcast()

	trace.AddEvent(ctx, "Stepped", latest, "Step completed")
	return &empty.Empty{}, nil
}

// Get the current simulated time.
func (s *server) Now(ctx context.Context, in *empty.Empty) (*pb.TimeResponse, error) {
	trace.AddEvent(ctx, in.String(), latest, "Get the time")

	if policy == pb.StepperPolicy_Invalid {
		return nil, trace.LogError(ctx, latest, "stepper not initialized: no stepper policy has been set")
	}

	syncLock.Lock()
	defer syncLock.Unlock()

	return &pb.TimeResponse{Current: latest}, nil
}

// Delay the simulated time by a specified amount +/- an allowed variance.  Do
// not return until that new time is current.
func (s *server) Delay(ctx context.Context, in *pb.DelayRequest) (*pb.TimeResponse, error) {
	trace.AddEvent(ctx, in.String(), latest, "Wait for the target time")

	if policy == pb.StepperPolicy_Invalid {
		return nil, trace.LogError(ctx, latest, "stepper not initialized: no stepper policy has been set")
	}

	if in.AtLeast < 0 {
		return nil, trace.LogError(ctx, latest, "base delay time must be non-negative, was specified as %d", in.AtLeast)
	}

	if in.Jitter < 0 {
		return nil, trace.LogError(ctx, latest, "delay jitter must be non-negative, was specified as %d", in.Jitter)
	}

	var adjust int64 = 0
	if in.Jitter > 0 {
		adjust = rand.Int63n(in.Jitter)
	}

	waitUntil(in.AtLeast + adjust)
	resp := pb.TimeResponse{Current: latest}
	trace.AddEvent(ctx, resp.String(), latest, "Delay completed")
	return &resp, nil
}
