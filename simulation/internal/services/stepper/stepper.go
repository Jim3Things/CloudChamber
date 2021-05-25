package stepper

// This file contains the state machine implementation of the stepper service.

import (
	"context"
	"sync"
	"time"

	"github.com/emirpasic/gods/maps/treemap"
	"github.com/emirpasic/gods/utils"

	"github.com/Jim3Things/CloudChamber/simulation/internal/common"
	"github.com/Jim3Things/CloudChamber/simulation/internal/services/stepper/messages"
	"github.com/Jim3Things/CloudChamber/simulation/internal/sm"
	"github.com/Jim3Things/CloudChamber/simulation/internal/tracing"
	"github.com/Jim3Things/CloudChamber/simulation/pkg/errors"
	pb "github.com/Jim3Things/CloudChamber/simulation/pkg/protos/services"
)

const (
	// queueDepth is the number of incoming messages that may be queued up
	// before the sender is forced to wait.
	queueDepth = 100
)

type stepper struct {
	sm *sm.SM

	latest  int64        // current simulate time, in ticks
	waiters *treemap.Map // waiting delay operations
	epoch   int64        // Epoch counter (policy changes)

	firstPolicy int // starting policy

	// The following entries support the Measured policy
	delay      time.Duration // Automatic step delay
	ticker     *time.Ticker  // Recurring timer
	stopTicker chan bool     // Termination marker

	// ch is the channel to send requests along to the stepper's goroutine,
	// which is where the state machine runs.
	ch chan sm.Envelope

	// startLock controls access to start and stop operations, and therefore to
	// the setup and tear down of the stepper's goroutine.
	startLock sync.Mutex
}

func newStepper(startingPolicy int) *stepper {
	s := &stepper{
		sm:          nil,
		latest:      0,
		waiters:     treemap.NewWith(utils.Int64Comparator),
		epoch:       0,
		firstPolicy: startingPolicy,
		delay:       0,
		ticker:      nil,
		ch:          make(chan sm.Envelope, queueDepth),
		startLock:   sync.Mutex{},
	}

	s.sm = sm.NewSM(s,
		"Stepper",
		sm.WithFirstState(
			pb.StepperState_awaiting_start,
			sm.NullEnter,
			[]sm.ActionEntry{
				{sm.TagStartSM, doStart, sm.Stay, pb.StepperState_faulted},
			},
			sm.UnexpectedMessage,
			sm.NullLeave),

		sm.WithState(
			pb.StepperState_invalid,
			invalidStateEnter,
			[]sm.ActionEntry{
				{messages.TagNoWaitPolicy, policy, pb.StepperState_no_wait, sm.Stay},
				{messages.TagMeasuredPolicy, measuredPolicy, pb.StepperState_measured, sm.Stay},
				{messages.TagManualPolicy, policy, pb.StepperState_manual, sm.Stay},
				{messages.TagReset, reset, pb.StepperState_invalid, sm.Stay},
				{messages.TagGetStatus, getStatus, sm.Stay, sm.Stay},
				{messages.TagAutoStep, sm.Ignore, sm.Stay, sm.Stay},
			},
			sm.UnexpectedMessage,
			sm.NullLeave),

		sm.WithState(
			pb.StepperState_no_wait,
			noWaitEnter,
			[]sm.ActionEntry{
				{messages.TagNoWaitPolicy, policy, pb.StepperState_no_wait, sm.Stay},
				{messages.TagMeasuredPolicy, measuredPolicy, pb.StepperState_measured, sm.Stay},
				{messages.TagManualPolicy, policy, pb.StepperState_manual, sm.Stay},
				{messages.TagReset, reset, pb.StepperState_invalid, sm.Stay},
				{messages.TagGetStatus, getStatus, sm.Stay, sm.Stay},
				{messages.TagStep, step, sm.Stay, sm.Stay},
				{messages.TagDelay, nowaitDelay, sm.Stay, sm.Stay},
				{messages.TagAutoStep, sm.Ignore, sm.Stay, sm.Stay},
			},
			sm.UnexpectedMessage,
			sm.NullLeave),

		sm.WithState(
			pb.StepperState_manual,
			sm.NullEnter,
			[]sm.ActionEntry{
				{messages.TagNoWaitPolicy, policy, pb.StepperState_no_wait, sm.Stay},
				{messages.TagMeasuredPolicy, measuredPolicy, pb.StepperState_measured, sm.Stay},
				{messages.TagManualPolicy, policy, pb.StepperState_manual, sm.Stay},
				{messages.TagReset, reset, pb.StepperState_invalid, sm.Stay},
				{messages.TagGetStatus, getStatus, sm.Stay, sm.Stay},
				{messages.TagStep, step, sm.Stay, sm.Stay},
				{messages.TagDelay, delay, sm.Stay, sm.Stay},
				{messages.TagAutoStep, sm.Ignore, sm.Stay, sm.Stay},
			},
			sm.UnexpectedMessage,
			sm.NullLeave),

		sm.WithState(
			pb.StepperState_measured,
			measuredEnter,
			[]sm.ActionEntry{
				{messages.TagNoWaitPolicy, policy, pb.StepperState_no_wait, sm.Stay},
				{messages.TagMeasuredPolicy, measuredPolicy, pb.StepperState_measured, sm.Stay},
				{messages.TagManualPolicy, policy, pb.StepperState_manual, sm.Stay},
				{messages.TagReset, reset, pb.StepperState_invalid, sm.Stay},
				{messages.TagGetStatus, getStatus, sm.Stay, sm.Stay},
				{messages.TagDelay, delay, sm.Stay, sm.Stay},
				{messages.TagAutoStep, autoStep, sm.Stay, sm.Stay},
			},
			sm.UnexpectedMessage,
			measuredLeave),

		sm.WithState(
			pb.StepperState_faulted,
			sm.TerminalEnter,
			[]sm.ActionEntry{},
			sm.UnexpectedMessage,
			sm.NullLeave),
	)

	return s
}

// Receive is the state machine entry function that is called to deliver a new
// message from outside the state machine.
func (s *stepper) Receive(msg sm.Envelope) {
	s.ch <- msg
}

// start initializes the simulated state machine goroutine, and its state
// machine context.
func (s *stepper) start(ctx context.Context) error {
	s.startLock.Lock()
	defer s.startLock.Unlock()

	// Only start the state machine once.  If it has already been started
	// then ignore this call.
	if s.sm.CurrentIndex == pb.StepperState_awaiting_start {
		go s.simulate()

		repl := make(chan *sm.Response)

		msg := sm.NewStartSM(ctx, repl)

		s.ch <- msg

		res := <-repl

		if res != nil {
			return res.Err
		}
	}

	return errors.ErrAlreadyStarted
}

// simulate is the main function for the state machine operation
func (s *stepper) simulate() {
	for !s.sm.Terminated {
		msg := <-s.ch

		ctx, span := tracing.StartSpan(
			context.Background(),
			tracing.WithName("Executing stepper operation"),
			tracing.WithNewRoot(),
			tracing.WithLink(msg.SpanContext(), msg.LinkID()))
		ctx = common.ContextWithTick(ctx, s.latest)

		s.sm.Current.Receive(ctx, s.sm, msg)

		span.End()
	}
}

// +++ State Enter/Leave functions

// invalidStateEnter cancels any outstanding Delay operations as part of
// clearing any prior state.
func invalidStateEnter(ctx context.Context, machine *sm.SM) error {
	s := machine.Parent.(*stepper)

	canceled := 0

	for k, v := s.waiters.Min(); k != nil; k, v = s.waiters.Min() {
		for _, c := range v.([]chan *sm.Response) {
			c <- sm.FailedResponse(s.latest, errors.ErrDelayCanceled)
			close(c)
			canceled++
		}

		s.waiters.Remove(k)
	}

	if canceled != 0 {
		tracing.Debug(ctx, "Canceled %d outstanding Delay operations", canceled)
	} else {
		tracing.Debug(ctx, "No outstanding Delay operations found")
	}

	return nil
}

// noWaitEnter jumps the current time tick ahead, as necessary to trigger the
// Delay operation that is next to expire.
func noWaitEnter(ctx context.Context, machine *sm.SM) error {
	advanceToFirstWaiter(ctx, machine)

	return nil
}

// measuredEnter starts a background recurring timer that will trigger the
// automatic step operations.
func measuredEnter(ctx context.Context, machine *sm.SM) error {
	s := machine.Parent.(*stepper)

	if s.delay == 0 {
		s.delay = time.Second
	}

	tracing.Debug(ctx, "Starting repeating timer with an interval of %v", s.delay)
	s.stopTicker = make(chan bool)

	s.ticker = time.NewTicker(s.delay)
	go func(match int64) {
		for done := false; !done; {
			ctx2, span := tracing.StartSpan(
				context.Background(),
				tracing.WithName("Waiting for ticker"),
				tracing.WithNewRoot(),
				tracing.AsInternal())
			ctx = common.ContextWithTick(ctx, s.latest)

			select {
			case <-s.ticker.C:
				tracing.Debug(ctx2, "Ticker triggered.  Processing.")
				s.ch <- messages.NewAutoStep(ctx2, match, nil)

			case <-s.stopTicker:
				tracing.Debug(ctx2, "Stop ticker received, exiting.")
				done = true
			}

			if !done && match != s.epoch {
				tracing.Debug(ctx2, "Policy has changed.")
				done = true
			}

			span.End()
		}
	}(s.epoch)

	return nil
}

// measuredLeave cancels the recurring timer and the associated background
// goroutine.
func measuredLeave(_ context.Context, machine *sm.SM, _ sm.StateIndex) {
	s := machine.Parent.(*stepper)

	s.stopTicker <- true
	close(s.stopTicker)

	s.ticker.Stop()
}

// --- State Enter/Leave functions

// +++ State actions

// doStart performs the state machine initialization, including an immediate
// transition to the state associated with the configured starting policy.
func doStart(ctx context.Context, machine *sm.SM, msg sm.Envelope) bool {
	tracing.UpdateSpanName(ctx, "Starting the simulated time state machine")

	s := machine.Parent.(*stepper)
	s.epoch++

	var err error

	switch s.firstPolicy {
	case messages.PolicyInvalid:
		err = machine.ChangeState(ctx, pb.StepperState_invalid)

	case messages.PolicyMeasured:
		err = machine.ChangeState(ctx, pb.StepperState_measured)

	case messages.PolicyManual:
		err = machine.ChangeState(ctx, pb.StepperState_manual)

	case messages.PolicyNoWait:
		err = machine.ChangeState(ctx, pb.StepperState_no_wait)

	default:
		err = &errors.ErrInvalidEnum{
			Field:  "StepperPolicy",
			Actual: int64(s.firstPolicy),
		}
	}

	ch := msg.Ch()
	defer close(ch)

	if err != nil {
		ch <- sm.FailedResponse(s.latest, err)
		return false
	}

	ch <- sm.SuccessResponse(s.latest)
	return true
}

// policy sets the simulated time policy for anything but Measured.
func policy(ctx context.Context, machine *sm.SM, msg sm.Envelope) bool {
	s := machine.Parent.(*stepper)
	m := msg.(messages.BasePolicy)

	tracing.UpdateSpanName(ctx, "Setting the simulated time policy to: %v", m)

	guard := m.GetGuard()

	ch := msg.Ch()
	defer close(ch)

	if guard >= 0 && guard < s.epoch {
		ch <- sm.FailedResponse(s.latest, &errors.ErrPolicyTooLate{
			Guard:   guard,
			Current: s.epoch,
		})
		return false
	}

	s.epoch++
	s.delay = 0

	ch <- statusResponse(machine, policyFromTag(m.Tag()))
	return true
}

// measuredPolicy sets the policy to Measured.  It requires an additional
// attribute that the other policies do not need -- the delay between automatic
// simulated time ticks.
func measuredPolicy(ctx context.Context, machine *sm.SM, msg sm.Envelope) bool {
	s := machine.Parent.(*stepper)
	m := msg.(*messages.MeasuredPolicy)

	tracing.UpdateSpanName(ctx, "Setting the simulated time policy to: %v", m)

	guard := m.GetGuard()

	ch := msg.Ch()
	defer close(ch)

	if guard >= 0 && guard < s.epoch {
		ch <- sm.FailedResponse(s.latest, &errors.ErrPolicyTooLate{
			Guard:   guard,
			Current: s.epoch,
		})
		return false
	}

	s.epoch++
	s.delay = m.Delay

	ch <- statusResponse(machine, policyFromTag(m.Tag()))
	return true
}

// reset forcibly sets the time service back to its initial conditions.  Note
// that the work to do so is split between this function and the entry into
// the Invalid state.
func reset(ctx context.Context, machine *sm.SM, msg sm.Envelope) bool {
	tracing.UpdateSpanName(ctx, "Resetting the simulated time service")

	// Reset the time to the start
	s := machine.Parent.(*stepper)
	s.latest = 0

	ch := msg.Ch()
	defer close(ch)

	cancelAllWaiters(ctx, machine)

	ch <- sm.SuccessResponse(s.latest)
	return true
}

// getStatus returns the current runtime status of the simulated time service.
func getStatus(ctx context.Context, machine *sm.SM, msg sm.Envelope) bool {
	tracing.UpdateSpanName(ctx, "Getting the current simulated time service status")

	ch := msg.Ch()
	defer close(ch)

	ch <- statusResponse(machine, policyFromState(machine.CurrentIndex))
	return true
}

// step advances the simulated time by one tick.
func step(ctx context.Context, machine *sm.SM, msg sm.Envelope) bool {
	tracing.UpdateSpanName(ctx, "Advance the simulated time by 1 tick")

	ch := msg.Ch()
	defer close(ch)

	advance(ctx, machine, 1)

	ch <- statusResponse(machine, policyFromState(machine.CurrentIndex))
	return true
}

// nowaitDelay registers a waiter, and then adjusts teh simulated time to
// forcibly expire it.
func nowaitDelay(ctx context.Context, machine *sm.SM, msg sm.Envelope) bool {
	addWaiter(ctx, msg, machine)

	advanceToFirstWaiter(ctx, machine)

	return false
}

// delay registers a waiter, which will return once the simulated time has
// passed its due time.
func delay(ctx context.Context, machine *sm.SM, msg sm.Envelope) bool {
	addWaiter(ctx, msg, machine)

	checkForExpiry(ctx, machine)
	return true
}

// autoStep performs an automatic tick advancement.  Since this is based on a
// wall clock timer, the advancement is guarded by a check to ensure that the
// policy that is in force at this point matches the policy that was in force
// when the wall clock timer was issued.  If it does not, then it is silently
// dropped.
func autoStep(ctx context.Context, machine *sm.SM, msg sm.Envelope) bool {
	m := msg.(*messages.AutoStep)
	s := machine.Parent.(*stepper)

	if m.Guard == s.epoch {
		tracing.UpdateSpanName(ctx, "Automatic simulated time step advance")

		advance(ctx, machine, 1)
		return true
	} else {
		tracing.UpdateSpanName(ctx, "Ignoring stale automatic simulated time step advance")
	}

	return false
}

// --- State actions

// +++ Supporting functions

// advance moves the simulated time tick by the specified number of ticks.
func advance(ctx context.Context, machine *sm.SM, ticks int64) {
	s := machine.Parent.(*stepper)
	s.latest += ticks

	ctx = common.ContextWithTick(ctx, s.latest)

	checkForExpiry(ctx, machine)
}

// advanceToFirstWaiter moves the simulated time forward to the earliest expiry
// time for any outstanding waiters.
func advanceToFirstWaiter(ctx context.Context, machine *sm.SM) {
	s := machine.Parent.(*stepper)

	if k, _ := s.waiters.Min(); k != nil {
		dueTime := k.(int64)
		if dueTime > s.latest {
			s.latest = dueTime
		}

		checkForExpiry(ctx, machine)
	}
}

// addWaiter adds a new waiting notification to the set of known waiters.  Each
// waiting entry consists of a due time and the channel where the success or
// failure response will be sent.  There may be multiple waiters that share a
// due time.
func addWaiter(ctx context.Context, msg sm.Envelope, machine *sm.SM) {
	m := msg.(*messages.Delay)
	s := machine.Parent.(*stepper)

	tracing.UpdateSpanName(ctx, "Wait until simulated time tick %d", m.DueTime)

	value, ok := s.waiters.Get(m.DueTime)
	if !ok {
		slot := []chan *sm.Response{msg.Ch()}
		s.waiters.Put(m.DueTime, slot)
	} else {
		slot := value.([]chan *sm.Response)
		slot = append(slot, msg.Ch())
		s.waiters.Put(m.DueTime, slot)
	}
}

// checkForExpiry tests to see if any waiters have now expired, to signal
// success and remove those that have.
func checkForExpiry(ctx context.Context, machine *sm.SM) {
	s := machine.Parent.(*stepper)

	k, v := s.waiters.Min()
	if k == nil {
		tracing.Debug(ctx, "No waiters found")
		return
	}

	count := 0

	key := k.(int64)
	for key <= s.latest {
		value := v.([]chan *sm.Response)
		for _, ch := range value {
			ch <- statusResponse(machine, policyFromState(machine.CurrentIndex))
			close(ch)
			count++
		}
		s.waiters.Remove(k)

		k, v = s.waiters.Min()
		if k == nil {
			break
		}

		key = k.(int64)
	}

	tracing.Debug(ctx, "Completed %d waiters", count)
}

// cancelAllWaiters forcibly terminates all waiters, issuing a canceled notice
// to each.  This is used when cleaning up the stepper internal state.
func cancelAllWaiters(ctx context.Context, machine *sm.SM) {
	s := machine.Parent.(*stepper)

	count := 0

	for _, k := range s.waiters.Keys() {
		v, ok := s.waiters.Get(k)

		if ok {
			value := v.([]chan *sm.Response)
			for _, ch := range value {
				ch <- sm.FailedResponse(s.latest, errors.ErrTimerCanceled(-1))
				close(ch)

				count++
			}
		}
	}

	tracing.Debug(ctx, "Canceled %d outstanding waiters", count)

	s.waiters = treemap.NewWith(utils.Int64Comparator)
}

// policyFromState translates the current state into a policy enum value that
// can be used in the GetStatus response message.
func policyFromState(state sm.StateIndex) int {
	switch state {
	case pb.StepperState_manual:
		return messages.PolicyManual

	case pb.StepperState_measured:
		return messages.PolicyMeasured

	case pb.StepperState_no_wait:
		return messages.PolicyNoWait

	default:
		return messages.PolicyInvalid
	}
}

// policyFromTag translates the message tag into a policy enum value.  This is
// used when the policy is changing, and the new state transition is committed
// but not yet actioned.
func policyFromTag(tag int) int {
	switch tag {
	case messages.TagManualPolicy:
		return messages.PolicyManual

	case messages.TagNoWaitPolicy:
		return messages.PolicyNoWait

	case messages.TagMeasuredPolicy:
		return messages.PolicyMeasured

	default:
		return messages.PolicyInvalid
	}
}

// statusResponse constructs a response that contains the current stepper
// status.
func statusResponse(machine *sm.SM, state int) *sm.Response {
	s := machine.Parent.(*stepper)

	return &sm.Response{
		Err: nil,
		At:  s.latest,
		Msg: messages.NewStatusResponseBody(
			s.epoch,
			int64(s.waiters.Size()),
			s.delay,
			state),
	}
}

// --- Supporting functions
