package stepper

import (
    "context"
    "errors"
    "fmt"
    "math/rand"
    "time"

    "github.com/AsynkronIT/protoactor-go/actor"
    "github.com/AsynkronIT/protoactor-go/scheduler"
    "github.com/golang/protobuf/ptypes"
    "github.com/golang/protobuf/ptypes/duration"
    "github.com/golang/protobuf/ptypes/empty"

    "github.com/Jim3Things/CloudChamber/internal/sm"
    "github.com/Jim3Things/CloudChamber/internal/tracing"
    "github.com/Jim3Things/CloudChamber/pkg/protos/common"
    pb "github.com/Jim3Things/CloudChamber/pkg/protos/services"
)

const (
	InvalidState = iota
	NoWaitState
	ManualState
	AutoStepState
)

var (
	policyToIndex = map[pb.StepperPolicy]int{
		pb.StepperPolicy_Invalid:  InvalidState,
		pb.StepperPolicy_NoWait:   NoWaitState,
		pb.StepperPolicy_Measured: AutoStepState,
		pb.StepperPolicy_Manual:   ManualState,
	}

	indexToPolicy = map[int]pb.StepperPolicy{
		InvalidState:  pb.StepperPolicy_Invalid,
		NoWaitState:   pb.StepperPolicy_NoWait,
		AutoStepState: pb.StepperPolicy_Measured,
		ManualState:   pb.StepperPolicy_Manual,
	}

	ErrStepperStaleVersion = errors.New("CloudChamber: simulated time state has a newer version than expected")
)

func (act *Actor) InitializeStates() {
	act.mgr.States = map[int]sm.State{
		InvalidState:  &InvalidStateImpl{Holder: act},
		NoWaitState:   &NoWaitStateImpl{Holder: act},
		ManualState:   &ManualStateImpl{Holder: act},
		AutoStepState: &AutoStepStateImpl{Holder: act},
	}

	act.mgr.StateNames = map[int]string{
		InvalidState:  "Invalid",
		NoWaitState:   "NoWait",
		ManualState:   "Manual",
		AutoStepState: "AutoStep",
	}
}

// +++ State machine States

// Invalid state. This is the starting state prior to establishing any policy.
// The only allowed operations are to establish an active policy.

type InvalidStateImpl struct {
	sm.EmptyState
	Holder *Actor
}

func (s *InvalidStateImpl) Receive(ca actor.Context) {
	holder := s.Holder
	ctx := sm.DecorateContext(ca)
	holder.TraceOnReceive(ctx)

	switch msg := ca.Message().(type) {
	case *pb.PolicyRequest:
		holder.HandlePolicy(ctx, msg)

	case *pb.ResetRequest:
		holder.HandleReset(ctx)

	case *pb.GetStatusRequest:
		holder.HandleGetStatus(ctx)

	default:
		if !isSystemMessage(ctx) {
			holder.mgr.RespondWithError(ctx, ErrInvalidRequest)
		}
	}
}

// NoWait policy state.  This state automatically advances time to the earliest
// delay deadline, resulting in no external waiting for time to pass.

type NoWaitStateImpl struct {
	sm.EmptyState
	Holder *Actor
}

func (s *NoWaitStateImpl) Enter(ctx context.Context) error {
	holder := s.Holder

	if err := holder.requireZeroDelay(ctx); err != nil {
		return err
	}

	// Moving to this state will automatically bump the time forward
	// to the first waiter, if there are any
	k, _ := holder.waiters.Min()
	if k != nil {
		holder.advance(ctx, k.(int64))
	}

	return nil
}

func (s *NoWaitStateImpl) Receive(ca actor.Context) {
	holder := s.Holder
	ctx := sm.DecorateContext(ca)
	holder.TraceOnReceive(ctx)

	switch msg := ca.Message().(type) {
	case *pb.DelayRequest:
		// Set the delay, as normal, then advance time to the earliest delay point,
		// which should be this one
		holder.HandleDelay(ctx, msg)
		k, _ := holder.waiters.Min()
		if k != nil {
			holder.advance(ctx, k.(int64))
		}

	default:
		holder.ApplyDefaultActions(ctx)
	}
}

// Manual policy state.  In this state time only moves forward as the result of
// an explicit Step request.

type ManualStateImpl struct {
	sm.EmptyState
	Holder *Actor
}

func (s *ManualStateImpl) Enter(ctx context.Context) error {
	holder := s.Holder

	if err := holder.requireZeroDelay(ctx); err != nil {
		return err
	}

	return nil
}

func (s *ManualStateImpl) Receive(ca actor.Context) {
	holder := s.Holder
	ctx := sm.DecorateContext(ca)
	holder.TraceOnReceive(ctx)

	s.Holder.ApplyDefaultActions(ctx)
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

	delay  time.Duration        // Automatic step delay
	cancel scheduler.CancelFunc // Timer state
}

func (s *AutoStepStateImpl) Enter(ctx context.Context) error {
	holder := s.Holder

	ca := sm.ActorContext(ctx)

	// Set up the automatic timer
	measuredDelay := ca.Message().(*pb.PolicyRequest).MeasuredDelay
	delay, err := ptypes.Duration(measuredDelay)
	if err != nil {
		s.delay = 0
		return err
	}

	if delay <= 0 {
		return tracing.Errorf(ctx, holder.latest, "delay must be greater than zero, but was %d", delay)
	}

	s.delay = delay
	timer := scheduler.NewTimerScheduler()
	s.cancel = timer.SendRepeatedly(s.delay, s.delay, ca.Self(), &pb.AutoStepRequest{Epoch: holder.epoch + 1})
	return nil
}

func (s *AutoStepStateImpl) Receive(ca actor.Context) {
	holder := s.Holder
	ctx := sm.DecorateContext(ca)
	holder.TraceOnReceive(ctx)

	switch msg := ca.Message().(type) {

	// Timer expired
	case *pb.AutoStepRequest:
		if msg.Epoch == holder.epoch {
			holder.advance(ctx, 1)
		}

	case *pb.StepRequest:
		holder.mgr.RespondWithError(ctx, ErrInvalidRequest)

	// Return the current status, including the measured delay
	case *pb.GetStatusRequest:
		rsp := &pb.StatusResponse{
			Policy:        indexToPolicy[holder.mgr.Current],
			MeasuredDelay: ptypes.DurationProto(s.delay),
			Now:           &common.Timestamp{Ticks: holder.latest},
			WaiterCount:   int64(holder.waiters.Size()),
			Epoch:         holder.epoch,
		}

		sm.ActorContext(ctx).Respond(rsp)

	default:
		holder.ApplyDefaultActions(ctx)
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
func (act *Actor) ApplyDefaultActions(ctx context.Context) {
	tracing.OnEnter(ctx, act.latest, "Applying Default Actions")
	ca := sm.ActorContext(ctx)

	if !isSystemMessage(ctx) {
		switch msg := ca.Message().(type) {
		case *pb.NowRequest:
			ca.Respond(&common.Timestamp{Ticks: act.latest})

		case *pb.DelayRequest:
			act.HandleDelay(ctx, msg)

		case *pb.StepRequest:
			act.HandleStep(ctx)

		case *pb.PolicyRequest:
			act.HandlePolicy(ctx, msg)

		// Normally we ignore auto-step requests as likely stale
		// notifications from a previously canceled timer.
		case *pb.AutoStepRequest:
			break

		case *pb.ResetRequest:
			act.HandleReset(ctx)

		case *pb.GetStatusRequest:
			act.HandleGetStatus(ctx)

		default:
			// The message was not one that we had a standard policy for
			act.mgr.RespondWithError(ctx, ErrInvalidRequest)
		}
	}
}

// Standard SetPolicy request handler.  Since each state reflects a distinct
// policy option, this method just forces a state change and responds with
// whether or not it worked.
func (act *Actor) HandlePolicy(ctx context.Context, pr *pb.PolicyRequest) {
	ca := sm.ActorContext(ctx)

	index, ok := policyToIndex[pr.Policy]
	if !ok {
		act.mgr.RespondWithError(ctx, ErrInvalidRequest)
		return
	}

	if pr.MatchEpoch >= 0 && pr.MatchEpoch != act.epoch {
		act.mgr.RespondWithError(ctx, ErrStepperStaleVersion)
		return
	}

	if err := act.mgr.ChangeState(ctx, act.latest, index); err != nil {
		act.mgr.RespondWithError(ctx, err)
		return
	}

	act.epoch++

	ca.Respond(&empty.Empty{})
}

// Standard single step request handler.  Advance the simulated time by 1 tick.
func (act *Actor) HandleStep(ctx context.Context) {
	act.advance(ctx, 1)
	sm.ActorContext(ctx).Respond(&empty.Empty{})
}

// Standard Delay request handler.
func (act *Actor) HandleDelay(ctx context.Context, dr *pb.DelayRequest) {
	ca := sm.ActorContext(ctx)

	due := dr.GetAtLeast().Ticks
	if dr.GetJitter() > 0 {
		due += rand.Int63n(dr.GetJitter())
	}

	// Having established a due time, add this sender to the set of waiters
	value, ok := act.waiters.Get(due)
	if !ok {
		// New due time entry
		slot := []*actor.PID{ca.Sender()}
		act.waiters.Put(due, slot)
	} else {
		// Existing entry, add this sender and update the map
		slot := value.([]*actor.PID)
		slot = append(slot, ca.Sender())
		act.waiters.Put(due, slot)
	}

	// Now, check for any waiters that have already expired (this handles the
	// case where the delay due time had already passed)
	act.checkForExpiry(ctx)
}

// Standard Reset request handler.  Force the state machine back to the
// initial conditions.
func (act *Actor) HandleReset(ctx context.Context) {
	if err := act.mgr.ChangeState(ctx, act.latest, InvalidState); err != nil {
		act.mgr.RespondWithError(ctx, err)
	}

	// Clear the common fields.
	act.waiters.Clear()
	act.latest = 0

	// Note that the AutoStep timer's epoch counter is specifically not
	// cleared, so that any outstanding timer completions are not processed.

	sm.ActorContext(ctx).Respond(&empty.Empty{})
}

// Standard Get Status request handler.  Returns the normal information, and
// assumes that the measured delay is zero.
func (act *Actor) HandleGetStatus(ctx context.Context) {
	rsp := &pb.StatusResponse{
		Policy:        indexToPolicy[act.mgr.Current],
		MeasuredDelay: &duration.Duration{Seconds: 0},
		Now:           &common.Timestamp{Ticks: act.latest},
		WaiterCount:   int64(act.waiters.Size()),
		Epoch:         act.epoch,
	}

	sm.ActorContext(ctx).Respond(rsp)
}

// Determine if the due time for any waiters from delay operations have
// expired.  Respond to all callers that have, waking them.
func (act *Actor) checkForExpiry(ctx context.Context) {
	ca := sm.ActorContext(ctx)

	k, v := act.waiters.Min()
	if k == nil {
		tracing.Info(ctx, act.latest, "No waiters found")
		return
	}

	key := k.(int64)

	for key <= act.latest {
		value := v.([]*actor.PID)
		for _, p := range value {
			ca.Send(p, &common.Timestamp{Ticks: act.latest})
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
func (act *Actor) advance(ctx context.Context, amount int64) {
	act.latest += amount
	act.checkForExpiry(ctx)
}

// Most policies require that the autostepping delay is zero.  The function
// validates that this is so.
func (act *Actor) requireZeroDelay(ctx context.Context) error {

	// Set up the automatic timer
	measuredDelay := sm.ActorContext(ctx).Message().(*pb.PolicyRequest).MeasuredDelay
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
