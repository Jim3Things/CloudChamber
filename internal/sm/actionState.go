package sm

import (
	"context"

	"github.com/Jim3Things/CloudChamber/internal/common"
	"github.com/Jim3Things/CloudChamber/internal/tracing"
)

const (
	// Stay is used in an ActionEntry as shorthand to indicate that there is
	// no state transition to process.
	Stay = ""
)

var (
	// NullEnter defines a state entry function that does nothing.
	NullEnter EnterFunc = nil

	// NullLeave defines a state exit function that does nothing.
	NullLeave LeaveFunc = nil
)

// ActionState is the common definition for how to process incoming requests
// for a particular state.
type ActionState struct {
	NullState

	// Entries is the set of actions that can be taken in this state, with the
	// require incoming message match criteria, and state transitions.
	Entries []ActionEntry

	// Default is the action to take if no item in the Entries array matches
	// the incoming message.
	Default ActionFunc

	// OnEnter is the optional function to call when entering this state
	// (including transition from this state back to this state).
	OnEnter EnterFunc

	// OnLeave is the optional function to call when exiting this state
	// (including transition from this state back to this state).
	OnLeave LeaveFunc
}

// EnterFunc is the signature definition for OnEnter functions.
type EnterFunc func(ctx context.Context, machine *SM) error

// LeaveFunc is the signature definition for OnLeave functions.  The nextState
// parameter contains the state ID for the state being transitioned to.  This
// allows for special processing when, for instance, the transition is from
// the current state back into the current state.
type LeaveFunc func(ctx context.Context, machine *SM, nextState string)

// ActionFunc is the signature definition for a message processing function
// listed in an ActionEntry.
type ActionFunc func(ctx context.Context, machine *SM, msg Envelope) bool

// ActionEntry defines a single match and process rule for a state.
type ActionEntry struct {
	// Match is the message Tag value which causes this rule to trigger.
	Match int

	// Action is the processing function to call.
	Action ActionFunc

	// TrueState is the state ID to change to if Action returns true.
	TrueState string

	// FalseState is the state ID to change to if Action returns false.
	FalseState string
}

// NewActionState creates a new ActionState instance with the supplied match
// and processing rules.
func NewActionState(
	actions []ActionEntry,
	other ActionFunc,
	onEnter EnterFunc,
	onLeave LeaveFunc) *ActionState {
	return &ActionState{
		NullState: NullState{},
		Entries:   actions,
		Default:   other,
		OnEnter:   onEnter,
		OnLeave:   onLeave,
	}
}
func (s *ActionState) Enter(ctx context.Context, machine *SM) error {
	if s.OnEnter != nil {
		return s.OnEnter(ctx, machine)
	}

	return nil
}

func (s *ActionState) Receive(ctx context.Context, machine *SM, msg Envelope) {
	if err := s.Process(ctx, machine, msg); err != nil {
		_ = tracing.Error(ctx, err)
	}
}

func (s *ActionState) Leave(ctx context.Context, sm *SM, nextState string) {
	if s.OnLeave != nil {
		s.OnLeave(ctx, sm, nextState)
	}
}

// Process is the standard processing step that finds the appropriate rule and
// executes it.
func (s *ActionState) Process(
	ctx context.Context,
	machine *SM,
	msg Envelope) error {

	for _, entry := range s.Entries {
		if entry.Match == msg.Tag() {
			if entry.Action != nil {
				nextState := entry.FalseState

				if entry.Action(ctx, machine, msg) {
					nextState = entry.TrueState
				}

				if nextState != Stay {
					return machine.ChangeState(ctx, nextState)
				}
			}

			return nil
		}
	}

	s.Default(ctx, machine, msg)
	return nil
}

// Ignore is a standard processing action that ignores the arriving message and
// produces no reply message at all.
func Ignore(ctx context.Context, machine *SM, msg Envelope) bool {
	tracing.Debug(
		ctx,
		"ignoring message %v while in state %q",
		msg,
		machine.CurrentIndex)
	return true
}

// UnexpectedMessage is an action function that signals that the incoming
// message was not expected, and its presence suggests a model consistency
// failure.
func UnexpectedMessage(ctx context.Context, machine *SM, msg Envelope) bool {
	_ = tracing.Error(ctx, "Unexpected message %v arrived in state %q", msg, machine.CurrentIndex)

	ch := msg.Ch()

	if ch != nil {
		ch <- UnexpectedMessageResponse(machine, common.TickFromContext(ctx), msg)
	}

	return true
}
