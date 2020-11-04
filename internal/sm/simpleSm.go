package sm

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/api/trace"

	"github.com/Jim3Things/CloudChamber/internal/common"
	"github.com/Jim3Things/CloudChamber/internal/tracing"
)

type Envelope interface {
	Initialize(ctx context.Context, ch chan *Response)
	GetCh() chan *Response

	GetSpanContext() trace.SpanContext
	GetLinkID() string
}

// Response holds the completion response for a processed request, whether it
// was successful or not.
type Response struct {
	// Err holds any completion error code, or nil if the request was
	// successful.
	Err error

	// At contains the simulated time tick when the request completed its
	// processing.
	At int64

	// Msg holds any extended results information, or nil if there either is
	// none, or if an error is returned.
	Msg interface{}
}

// UnexpectedMessage is the standard error when an incoming request arrives in
// a state that is not expecting it.
type UnexpectedMessage struct {
	Msg   string
	State string
}

func (um *UnexpectedMessage) Error() string {
	return fmt.Sprintf("unexpected message %q while in state %q", um.Msg, um.State)
}

// SimpleSMState defines the methods used for state actions and transitions.
type SimpleSMState interface {

	// Enter is called when a state transition moves to this state
	Enter(ctx context.Context, sm *SimpleSM) error

	// Receive is called on the active start implementation when a new
	// incoming message arrives
	Receive(ctx context.Context, machine *SimpleSM, msg Envelope)

	// Leave is called when a state transition moves away from this state
	Leave(ctx context.Context, sm *SimpleSM)

	// Name returns a readable string that names this state
	Name() string
}

// NullState is the default implementation of a simple SM state
type NullState struct{}

// Enter is the default (no-action) implementation.
func (*NullState) Enter(context.Context, *SimpleSM) error { return nil }

// Receive is the default (no-action) implementation.
func (*NullState) Receive(ctx context.Context, machine *SimpleSM, msg Envelope) {
	msg.GetCh() <- &Response{
		Err: &UnexpectedMessage{
			Msg:   fmt.Sprintf("%v", msg),
			State: machine.Current.Name(),
		},
		At:  common.TickFromContext(ctx),
		Msg: nil,
	}
}

// Leave is the default (no-action) implementation.
func (*NullState) Leave(context.Context, *SimpleSM) {}

// Name returns the string identifying the null state
func (*NullState) Name() string { return "NullState" }

// TerminalState is the default final state implementation.
type TerminalState struct {
	NullState
}

// Enter marks this state machine as terminated.
func (*TerminalState) Enter(_ context.Context, sm *SimpleSM) error {
	sm.Terminated = true
	return nil
}

// Name returns the string identifying the terminal state
func (*TerminalState) Name() string { return "Terminated" }

// SimpleSM defines a simplified state machine structure. It assumes that the
// issues of concurrency and lifecycle management are handled by some external
// logic.
type SimpleSM struct {
	common.Guarded

	// CurrentIndex holds the index to the current state
	CurrentIndex int

	// Current is a pointer to the current state
	Current SimpleSMState

	// FirstState is the index to the starting state
	FirstState int

	// States holds the map of known state index values to state implementations
	States map[int]SimpleSMState

	// Parent points to the structure that holds this state machine, and likely
	// holds global context that the state actions need.
	Parent interface{}

	// Terminated is true if the state machine has reached its final state.
	Terminated bool
}

// StateDecl defines the type expected for a state declaration decorator when
// creating a new SimpleSM instance
type StateDecl func() (int, bool, SimpleSMState)

// WithState is a decorator that defines a state in the state machine
func WithState(id int, decl SimpleSMState) StateDecl {
	return func() (int, bool, SimpleSMState) {
		return id, false, decl
	}
}

// WithFirstState is a decorator that defines the starting state for the state
// machine
func WithFirstState(id int, decl SimpleSMState) StateDecl {
	return func() (int, bool, SimpleSMState) {
		return id, true, decl
	}
}

// NewSimpleSM creates a new state machine instance with the associated
// parent instance reference, as well as the state declarations.
func NewSimpleSM(parent interface{}, decls ...StateDecl) *SimpleSM {
	states := make(map[int]SimpleSMState)
	firstState := 0

	for _, decl := range decls {
		stateId, first, instance := decl()
		states[stateId] = instance
		if first {
			firstState = stateId
		}
	}

	return &SimpleSM{
		CurrentIndex: firstState,
		Current:      states[firstState],
		FirstState:   firstState,
		States:       states,
		Parent:       parent,
		Terminated:   false,
	}
}

// ChangeState changes the current state.  Leave the old state, try to
// enter the new state, and declare that state as current if successful.
func (sm *SimpleSM) ChangeState(ctx context.Context, newState int) error {
	tracing.Info(
		ctx,
		"Change state from %q to %q",
		sm.Current.Name(),
		sm.States[newState].Name())

	cur := sm.Current
	cur.Leave(ctx, nil)

	cur = sm.States[newState]
	if err := cur.Enter(ctx, sm); err != nil {
		return tracing.Error(ctx, err)
	}

	sm.CurrentIndex = newState
	sm.Current = cur
	sm.AdvanceGuard(common.TickFromContext(ctx))

	return nil
}

// Receive processes an incoming message by routing it to the active state
// handler.
func (sm *SimpleSM) Receive(ctx context.Context, msg Envelope) {
	sm.Current.Receive(ctx, sm, msg)
}

// Start sets the state machine to its first (starting) state
func (sm *SimpleSM) Start(ctx context.Context) error {
	cur := sm.States[sm.FirstState]
	if err := cur.Enter(ctx, sm); err != nil {
		return tracing.Error(ctx, err)
	}

	sm.CurrentIndex = sm.FirstState
	sm.Current = cur
	return nil
}
