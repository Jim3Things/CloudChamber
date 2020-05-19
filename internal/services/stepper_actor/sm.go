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

var (
    policyToIndex = map[pb.StepperPolicy]int{
        pb.StepperPolicy_Invalid  : InvalidState,
        pb.StepperPolicy_NoWait   : NoWaitState,
        pb.StepperPolicy_Measured : AutoStepState,
        pb.StepperPolicy_Manual   : ManualState,
    }
)

func (act *Actor) InitializeStates() {
    act.mgr.States = map[int]sm.State{
        InvalidState  : &InvalidStateImpl { Holder: act},
        NoWaitState   : &NoWaitStateImpl { Holder: act},
        ManualState   : &ManualStateImpl { Holder: act},
        AutoStepState : &AutoStepStateImpl { Holder: act},
    }

    act.mgr.StateNames = map[int]string {
        InvalidState  : "Invalid",
        NoWaitState   : "NoWait",
        ManualState   : "Manual",
        AutoStepState : "AutoStep",
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

    switch msg := ctx.Message().(type) {
    case *pb.PolicyRequest:
        holder.HandlePolicy(c, span, ctx, msg)

    case *pb.ResetRequest:
        holder.HandleReset(c, span, ctx)

    default:
        if !isSystemMessage(c, span, ctx) {
            holder.mgr.RespondWithError(c, span, ctx, ErrInvalidRequest)
        }
    }
}

// NoWait policy state.  This state automatically advances time to the earliest
// delay deadline, resulting in no external waiting for time to pass.

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
        holder.advance(c, span, ctx, k.(int64))
    }

    return nil
}

func (s *NoWaitStateImpl) Receive(ctx actor.Context) {
    holder := s.Holder

    c, span := holder.getSpan(ctx)
    defer span.End()

    switch msg := ctx.Message().(type) {
    case *pb.DelayRequest:
        // Set the delay, as normal, then advance time to the earliest delay point,
        // which should be this one
        holder.HandleDelay(c, span, ctx, msg)
        k, _ := holder.waiters.Min()
        if k != nil {
            holder.advance(c, span, ctx, k.(int64))
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

func (s *ManualStateImpl) Enter(ctx actor.Context, _ context.Context, _ trc.Span) error {
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

// Autostep policy state.  In this state time moves forward at a constant rate
// based on the external (actual) clock.  It allows for time expansion or
// compression, and does not require active requests to step time forward.
//
// Note that timers have an associated epoch, which allows detection of late
// timer expiration events in the case where the timer has been superseded.

type AutoStepStateImpl struct {
    sm.EmptyState
    Holder *Actor

    delay  time.Duration            // Automatic step delay
    cancel scheduler.CancelFunc     // Timer state
    epoch  int64                    // Epoch counter for the timer
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
        return trace.Errorf(c, span, holder.latest, "delay must be greater than zero, but was %d", delay)
    }

    s.delay = delay
    s.epoch++
    timer := scheduler.NewTimerScheduler()
    s.cancel = timer.SendRepeatedly(s.delay, s.delay, ctx.Self(), &pb.AutoStepRequest{ Epoch: s.epoch })
    return nil
}

func (s *AutoStepStateImpl) Receive(ctx actor.Context) {
    holder := s.Holder

    c, span := holder.getSpan(ctx)
    defer span.End()

    switch msg := ctx.Message().(type) {

    // Timer expired
    case *pb.AutoStepRequest:
        if msg.Epoch == s.epoch {
            holder.advance(c, span, ctx, 1)
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

// Apply the default actions to the current message.  Most states have special
// processing for a subset of the possible messages.  The pattern is to first
// handle those, and then call this method to handle the rest.
func (act *Actor) ApplyDefaultActions(c context.Context, span trc.Span, ctx actor.Context) {
    trace.OnEnter(c, span, act.latest, "Applying Default Actions")

    if !isSystemMessage(c, span, ctx) {
        switch msg := ctx.Message().(type) {
        case *pb.NowRequest:
            ctx.Respond(&common.Timestamp{Ticks: act.latest})

        case *pb.DelayRequest:
            act.HandleDelay(c, span, ctx, msg)

        case *pb.StepRequest:
            act.HandleStep(c, span, ctx)

        case *pb.PolicyRequest:
            act.HandlePolicy(c, span, ctx, msg)

        // Normally we ignore auto-step requests as likely stale
        // notifications from a previously canceled timer.
        case *pb.AutoStepRequest:
            break

        case *pb.ResetRequest:
            act.HandleReset(c, span, ctx)

        default:
            // The message was not one that we had a standard policy for
            act.mgr.RespondWithError(c, span, ctx, ErrInvalidRequest)
        }
    }
}

// Standard SetPolicy request handler.  Since each state reflects a distinct
// policy option, this method just forces a state change and responds with
// whether or not it worked.
func (act *Actor) HandlePolicy(c context.Context, span trc.Span, ctx actor.Context, pr *pb.PolicyRequest) {
    index, ok := policyToIndex[pr.Policy]
    if !ok {
        act.mgr.RespondWithError(c, span, ctx, ErrInvalidRequest)
        return
    }

    if err := act.mgr.ChangeState(c, span, ctx, act.latest, index); err != nil {
        act.mgr.RespondWithError(c, span, ctx, err)
        return
    }

    ctx.Respond(&empty.Empty{})
}

// Standard single step request handler.  Advance the simulated time by 1 tick.
func (act *Actor) HandleStep(c context.Context, span trc.Span, ctx actor.Context) {
    act.advance(c, span, ctx, 1)
    ctx.Respond(&empty.Empty{})
}

// Standard Delay request handler.
func (act *Actor) HandleDelay(c context.Context, span trc.Span, ctx actor.Context, dr *pb.DelayRequest) {
    due := dr.GetAtLeast().Ticks
    if dr.GetJitter() > 0 {
        due += rand.Int63n(dr.GetJitter())
    }

    // Having established a due time, add this sender to the set of waiters
    value, ok := act.waiters.Get(due)
    if !ok {
        // New due time entry
        slot := []*actor.PID { ctx.Sender() }
        act.waiters.Put(due, slot)
    } else {
        // Existing entry, add this sender and update the map
        slot := value.([]*actor.PID)
        slot = append(slot, ctx.Sender())
        act.waiters.Put(due, slot)
    }

    // Now, check for any waiters that have already expired (this handles the
    // case where the delay due time had already passed)
    act.checkForExpiry(c, span, ctx)
}

// Standard Reset request handler.  Force the state machine back to the
// initial conditions.
func (act *Actor) HandleReset(c context.Context, span trc.Span, ctx actor.Context) {
    if err := act.mgr.ChangeState(c, span, ctx, act.latest, InvalidState); err != nil {
        act.mgr.RespondWithError(c, span, ctx, err)
    }

    // Clear the common fields.
    act.waiters.Clear()
    act.latest = 0

    // Note that the AutoStep timer's epoch counter is specifically not
    // cleared, so that any outstanding timer completions are not processed.

    ctx.Respond(&empty.Empty{})
}

// Determine if the due time for any waiters from delay operations have
// expired.  Respond to all callers that have, waking them.
func (act *Actor) checkForExpiry(c context.Context, span trc.Span, ctx actor.Context) {
    k, v := act.waiters.Min()
    if k == nil {
        trace.Info(c, span, act.latest, "No waiters found")
        return
    }

    key := k.(int64)

    for key <= act.latest {
        value := v.([]*actor.PID)
        for _, p := range value {
            ctx.Send(p, &common.Timestamp{Ticks: act.latest })
        }
        act.waiters.Remove(k)

        k, v = act.waiters.Min()
        if k == nil {
            break
        }

        key = k.(int64)
    }
}

// Helper method that advances the simulated time by the specified amount.
func (act *Actor) advance(c context.Context, span trc.Span, ctx actor.Context, amount int64) {
    act.latest += amount
    act.checkForExpiry(c, span, ctx)
}

// Most policies require that the autostepping delay is zero.  The function
// validates that this is so.
func (act *Actor) requireZeroDelay(ctx actor.Context) error {

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

