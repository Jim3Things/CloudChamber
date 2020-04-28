package stepper

import (
    "context"
    "fmt"
    "time"

    "github.com/AsynkronIT/protoactor-go/actor"
    "github.com/AsynkronIT/protoactor-go/scheduler"
    "github.com/golang/protobuf/ptypes"

    "github.com/Jim3Things/CloudChamber/internal/sm"
    trace "github.com/Jim3Things/CloudChamber/internal/tracing/server"
    pb "github.com/Jim3Things/CloudChamber/pkg/protos/Stepper"
    "github.com/Jim3Things/CloudChamber/pkg/protos/common"
)

func (s *Actor) InitializeStates() {
    s.mgr.States[InvalidState] = &InvalidStateImpl { Holder: s }
    s.mgr.States[NoWaitState] = &NoWaitStateImpl { Holder: s }
    s.mgr.States[ManualState] = &ManualStateImpl { Holder: s }
    s.mgr.States[AutoStepState] = &AutoStepStateImpl { Holder: s }
}

// +++ State machine States

const (
    InvalidState = iota
    NoWaitState
    ManualState
    AutoStepState
)

// Invalid state. This is the starting state prior to establishing any policy.
// The only allowed operations are to establish an active policy.

type InvalidStateImpl struct {
    sm.EmptyState
    Holder *Actor
}

func (s *InvalidStateImpl) Receive(ctx actor.Context) {
    spa := s.Holder

    if !spa.HandleSystemMessages(ctx) {
        switch ctx.Message().(type) {
        case *pb.PolicyRequest:
            spa.HandlePolicy(ctx)

        default:
            spa.mgr.RespondWithError(ctx, ErrInvalidRequest)
        }
    }
}

// NoWait policy state.  This state

type NoWaitStateImpl struct {
    sm.EmptyState
    Holder *Actor
}

func (s *NoWaitStateImpl) Receive(ctx actor.Context) {
    spa := s.Holder

    if !spa.HandleSystemMessages(ctx) {
        switch ctx.Message().(type) {
        case *pb.PolicyRequest:
            spa.HandlePolicy(ctx)

        default:
            if !spa.HandleSimple(ctx) {
                spa.mgr.RespondWithError(ctx, ErrInvalidRequest)
            }
        }
    }
}

// Manual policy state.  In this state time only moves forward as the result of
// an explicit Step request.

type ManualStateImpl struct {
    sm.EmptyState
    Holder *Actor
}

func (s *ManualStateImpl) Receive(ctx actor.Context) {
    spa := s.Holder

    if !spa.HandleSimple(ctx) {

    }
}

type AutoStepStateImpl struct {
    sm.EmptyState
    Holder *Actor

    // Automatic step delay
    delay time.Duration

    // Timer state
    cancel scheduler.CancelFunc

    // Epoch counter for the timer
    epoch int64
}

func (s *AutoStepStateImpl) Receive(ctx actor.Context) {
    spa := s.Holder

    if !spa.HandleSystemMessages(ctx) {
        switch ctx.Message().(type) {

        // Timer expired
        case *pb.AutoStepRequest:
            asr := ctx.Message().(*pb.AutoStepRequest)
            if asr.Epoch == s.epoch {
                spa.HandleStep(ctx)
            }

        case *pb.StepRequest:
            spa.mgr.RespondWithError(ctx, ErrInvalidRequest)

        default:
            if !spa.HandleSimple(ctx) {
                // Unknown message
                spa.mgr.RespondWithError(ctx, ErrInvalidRequest)
            }
        }
    }
}

func (s *AutoStepStateImpl) Enter(ctx actor.Context) error {
    spa := s.Holder

    // Set up the automatic timer
    measuredDelay := ctx.Message().(pb.PolicyRequest).MeasuredDelay
    delay, err := ptypes.Duration(measuredDelay)
    if err != nil {
        s.delay = 0
        return err
    }

    if delay <= 0 {
        c := context.Background()
        return trace.LogError(c, spa.latest, "delay must be greater than zero, but was %d", delay)
    }

    s.delay = delay
    s.epoch += 1
    timer := scheduler.NewTimerScheduler()
    s.cancel = timer.SendRepeatedly(s.delay, s.delay, ctx.Self(), &pb.AutoStepRequest{Epoch: s.epoch})
    return nil
}

func (s *AutoStepStateImpl) Leave() {
    // Cancel the automatic timer
    if s.cancel != nil {
        s.cancel()
        s.cancel = nil
    }
}

// --- State machine States

// +++ State machine helper functions

func (s *Actor) HandleSimple(ctx actor.Context) bool {
    switch ctx.Message().(type) {
    case *pb.NowRequest:
        ctx.Respond(&common.Timestamp{Ticks: s.latest })
        return true

    // Normally we ignore auto-step requests as likely stale
    // notifications from a previously canceled timer.
    case *pb.AutoStepRequest:
        return true
    }

    // The message was not one that we had a simple handler for
    return false
}

func (s *Actor) HandleSystemMessages(ctx actor.Context) bool {
    m, ok := ctx.Message().(actor.SystemMessage)

    fmt.Printf("Message arrived.  system: %v, msg: %v]\n", ok, m)
    return ok
}

func (s *Actor) HandlePolicy(ctx actor.Context) {

}

func (s *Actor) HandleStep(ctx actor.Context) {
}

// --- State machine helper functions

