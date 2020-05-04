package stepper

import (
    "context"
    "fmt"
    "math/rand"
    "time"

    "github.com/AsynkronIT/protoactor-go/actor"
    "github.com/AsynkronIT/protoactor-go/scheduler"
    "github.com/golang/protobuf/ptypes"
    "github.com/golang/protobuf/ptypes/empty"
    trc "go.opentelemetry.io/otel/api/trace"

    "github.com/Jim3Things/CloudChamber/internal/sm"
    trace "github.com/Jim3Things/CloudChamber/internal/tracing/server"
    pb "github.com/Jim3Things/CloudChamber/pkg/protos/Stepper"
    "github.com/Jim3Things/CloudChamber/pkg/protos/common"
)

const (
    InvalidState = iota
    NoWaitState
    ManualState
    AutoStepState
)

func (s *Actor) InitializeStates() {
    s.mgr.States = map[int]sm.State{
        InvalidState  : &InvalidStateImpl { Holder: s },
        NoWaitState   : &NoWaitStateImpl { Holder: s },
        ManualState   : &ManualStateImpl { Holder: s },
        AutoStepState : &AutoStepStateImpl { Holder: s },
    }

    s.mgr.StateNames = map[int]string {
        InvalidState  : "Invalid",
        NoWaitState   : "NoWait",
        ManualState   : "Manual",
        AutoStepState : "AutoStep",
    }

    s.policyToIndex = map[pb.StepperPolicy]int{
        pb.StepperPolicy_Invalid  : InvalidState,
        pb.StepperPolicy_NoWait   : NoWaitState,
        pb.StepperPolicy_Measured : AutoStepState,
        pb.StepperPolicy_Manual   : ManualState,
    }
}

// +++ State machine States

// Invalid state. This is the starting state prior to establishing any policy.
// The only allowed operations are to establish an active policy.

type InvalidStateImpl struct {
    sm.EmptyState
    Holder *Actor
}

func (s *InvalidStateImpl) Receive(ctx actor.Context) {
    holder := s.Holder

    c, span := holder.getSpan(ctx)
    defer span.End()

    switch ctx.Message().(type) {
    case *pb.PolicyRequest:
        holder.HandlePolicy(c, span, ctx)

    case *pb.ResetRequest:
        holder.HandleReset(c, span, ctx)

    default:
        if !holder.HandleSystemMessages(c, span, ctx) {
            holder.mgr.RespondWithError(c, span, ctx, ErrInvalidRequest)
        }
    }
}

// NoWait policy state.  This state

type NoWaitStateImpl struct {
    sm.EmptyState
    Holder *Actor
}

func (s *NoWaitStateImpl) Enter(ctx actor.Context, c context.Context, span trc.Span) error {
    holder := s.Holder

    if err := holder.requireZeroDelay(ctx); err != nil {
        return err
    }

    // Moving to this state will automatically bump the time forward
    // to the first waiter, if there are any
    k, _ := holder.waiters.Min()
    if k != nil {
        holder.Advance(c, span, ctx, k.(int64))
    }

    return nil
}

func (s *NoWaitStateImpl) Receive(ctx actor.Context) {
    holder := s.Holder

    c, span := holder.getSpan(ctx)
    defer span.End()


    switch ctx.Message().(type) {
    case *pb.DelayRequest:
        // Set the delay, as normal, then advance time to the earliest delay point,
        // which should be this one
        holder.HandleDelay(c, span, ctx)
        k, _ := holder.waiters.Min()
        if k != nil {
            holder.Advance(c, span, ctx, k.(int64))
        }

    default:
        holder.ApplyDefaultActions(c, span, ctx)
    }
}

// Manual policy state.  In this state time only moves forward as the result of
// an explicit Step request.

type ManualStateImpl struct {
    sm.EmptyState
    Holder *Actor
}

func (s *ManualStateImpl) Enter(ctx actor.Context, c context.Context, span trc.Span) error {
    holder := s.Holder

    if err := holder.requireZeroDelay(ctx); err != nil {
        return err
    }

    return nil
}

func (s *ManualStateImpl) Receive(ctx actor.Context) {
    c, span := s.Holder.getSpan(ctx)
    defer span.End()

    s.Holder.ApplyDefaultActions(c, span, ctx)
}

// Autostep policy state.  In this state time moves forward
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

func (s *AutoStepStateImpl) Enter(ctx actor.Context, c context.Context, span trc.Span) error {
    holder := s.Holder

    // Set up the automatic timer
    measuredDelay := ctx.Message().(*pb.PolicyRequest).MeasuredDelay
    delay, err := ptypes.Duration(measuredDelay)
    if err != nil {
        s.delay = 0
        return err
    }

    if delay <= 0 {
        return trace.LogError(c, holder.latest, "delay must be greater than zero, but was %d", delay)
    }

    s.delay = delay
    s.epoch += 1
    timer := scheduler.NewTimerScheduler()
    s.cancel = timer.SendRepeatedly(s.delay, s.delay, ctx.Self(), &pb.AutoStepRequest{ Epoch: s.epoch })
    return nil
}

func (s *AutoStepStateImpl) Receive(ctx actor.Context) {
    holder := s.Holder

    c, span := holder.getSpan(ctx)
    defer span.End()

    switch ctx.Message().(type) {

    // Timer expired
    case *pb.AutoStepRequest:
        asr := ctx.Message().(*pb.AutoStepRequest)
        if asr.Epoch == s.epoch {
            holder.Advance(c, span, ctx, 1)
        }

    case *pb.StepRequest:
        holder.mgr.RespondWithError(c, span, ctx, ErrInvalidRequest)

    default:
        holder.ApplyDefaultActions(c, span, ctx)
    }
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


func (s *Actor) ApplyDefaultActions(c context.Context, span trc.Span, ctx actor.Context) {
    trace.AddEvent(c, span, "Applying Default Actions", s.latest, "")
    if !s.HandleSystemMessages(c, span, ctx) {
        switch ctx.Message().(type) {
        case *pb.NowRequest:
            ctx.Respond(&common.Timestamp{Ticks: s.latest})

        case *pb.DelayRequest:
            s.HandleDelay(c, span, ctx)

        case *pb.StepRequest:
            s.HandleStep(c, span, ctx)

        case *pb.PolicyRequest:
            s.HandlePolicy(c, span, ctx)

        // Normally we ignore auto-step requests as likely stale
        // notifications from a previously canceled timer.
        case *pb.AutoStepRequest:
            break

        case *pb.ResetRequest:
            s.HandleReset(c, span, ctx)

        default:
            // The message was not one that we had a standard policy for
            s.mgr.RespondWithError(c, span, ctx, ErrInvalidRequest)
        }
    }
}

func (s *Actor) HandleSystemMessages(c context.Context, span trc.Span, ctx actor.Context) bool {
    _, ok := ctx.Message().(actor.SystemMessage)

    return ok
}

func (s *Actor) HandlePolicy(c context.Context, span trc.Span, ctx actor.Context) {
    pr, ok := ctx.Message().(*pb.PolicyRequest)
    if !ok {
        s.mgr.RespondWithError(c, span, ctx, ErrInvalidRequest)
        return
    }

    if err := pr.Validate(); err != nil {
        s.mgr.RespondWithError(c, span, ctx, err)
        return
    }

    index, ok := s.policyToIndex[pr.Policy]
    if !ok {
        s.mgr.RespondWithError(c, span, ctx, ErrInvalidRequest)
        return
    }

    if err := s.mgr.ChangeState(c, span, ctx, s.latest, index); err != nil {
        s.mgr.RespondWithError(c, span, ctx, err)
        return
    }

    ctx.Respond(&empty.Empty{})
}

func (s *Actor) HandleStep(c context.Context, span trc.Span, ctx actor.Context) {
    s.Advance(c, span, ctx, 1)
    ctx.Respond(&common.Completion{
        IsError: false,
        Error:   "",
    })
}

func (s *Actor) Advance(c context.Context, span trc.Span, ctx actor.Context, amount int64) {
    s.latest += amount
    s.checkForExpiry(c, span, ctx)
}

func (s *Actor) HandleDelay(c context.Context, span trc.Span, ctx actor.Context) {
    dr, ok := ctx.Message().(*pb.DelayRequest)
    if !ok {
        s.mgr.RespondWithError(c, span, ctx, ErrInvalidRequest)
        return
    }

    if err := dr.Validate(); err != nil {
        s.mgr.RespondWithError(c, span, ctx, err)
        return
    }

    due := dr.GetAtLeast().Ticks
    if dr.GetJitter() > 0 {
        due += rand.Int63n(dr.GetJitter())
    }

    value, ok := s.waiters.Get(due)
    if !ok {
        // New due time entry
        slot := []*actor.PID { ctx.Sender() }
        s.waiters.Put(due, slot)
    } else {
        // Existing entry, add this sender and update the map
        slot := value.([]*actor.PID)
        slot = append(slot, ctx.Sender())
        s.waiters.Put(due, slot)
    }
}

func (s *Actor) HandleReset(c context.Context, span trc.Span, ctx actor.Context) {
    if err := s.mgr.ChangeState(c, span, ctx, s.latest, InvalidState); err != nil {
        s.mgr.RespondWithError(c, span, ctx, err)
    }

    s.waiters.Clear()
    s.latest = 0

    ctx.Respond(&empty.Empty{})
}

func (s *Actor) checkForExpiry(c context.Context, span trc.Span, ctx actor.Context) {
    k, v := s.waiters.Min()
    if k == nil {
        s.mgr.AddEvent(c, span, s.latest, "No waiters found")
        return
    }

    key := k.(int64)

    for key <= s.latest {
        value := v.([]*actor.PID)
        for _, p := range value {
            ctx.Send(p, &common.Timestamp{Ticks: s.latest })
        }
        s.waiters.Remove(k)

        k, v = s.waiters.Min()
        if k == nil {
            break
        }

        key = k.(int64)
    }
}

func (s *Actor) requireZeroDelay(ctx actor.Context) error {

    // Set up the automatic timer
    measuredDelay := ctx.Message().(*pb.PolicyRequest).MeasuredDelay
    delay, err := ptypes.Duration(measuredDelay)
    if err != nil {
        return err
    }

    if delay != 0 {
        return fmt.Errorf("delay must be zero, but was %d", delay)
    }

    return nil
}

// --- State machine helper functions

