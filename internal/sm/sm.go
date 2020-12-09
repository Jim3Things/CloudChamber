package sm

import (
	"context"

	"github.com/Jim3Things/CloudChamber/internal/common"
	"github.com/Jim3Things/CloudChamber/internal/tracing"
)

const (
	invalidState = "Invalid"
)

// SM defines a simplified state machine structure. It assumes that the issues
// of concurrency and lifecycle management are handled by some external logic.
type SM struct {
	common.Guarded

	// CurrentIndex holds the index to the current state
	CurrentIndex string

	// Current is a pointer to the current state
	Current SmState

	// FirstState is the index to the starting state
	FirstState string

	// States holds the map of known state index values to state implementations
	States map[string]SmState

	// Parent points to the structure that holds this state machine, and likely
	// holds global context that the state actions need.
	Parent interface{}

	// Terminated is true if the state machine has reached its final state.
	Terminated bool
}

// StateDecl defines the type expected for a state declaration decorator when
// creating a new SM instance
type StateDecl func() (bool, string, SmState)

// WithState is a decorator that defines a state in the state machine
func WithState(
	name string,
	onEnter EnterFunc,
	actions []ActionEntry,
	other ActionFunc,
	onLeave LeaveFunc) StateDecl {
	return func() (bool, string, SmState) {
		return false, name, NewActionState(actions, other, onEnter, onLeave)
	}
}

// WithFirstState is a decorator that defines the starting state for the state
// machine
func WithFirstState(
	name string,
	onEnter EnterFunc,
	actions []ActionEntry,
	other ActionFunc,
	onLeave LeaveFunc) StateDecl {
	return func() (bool, string, SmState) {
		return true, name, NewActionState(actions, other, onEnter, onLeave)
	}
}

// NewSM creates a new state machine instance with the associated
// parent instance reference, as well as the state declarations.
func NewSM(parent interface{}, decls ...StateDecl) *SM {
	states := make(map[string]SmState)

	firstState := invalidState

	for _, decl := range decls {
		first, name, instance := decl()
		states[name] = instance

		if first {
			firstState = name
		}
	}

	return &SM{
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
func (sm *SM) ChangeState(ctx context.Context, newState string) error {
	tracing.Info(
		ctx,
		"Change state from %q to %q",
		sm.CurrentIndex,
		newState)

	cur := sm.Current
	cur.Leave(ctx, sm, newState)

	cur = sm.States[newState]
	sm.CurrentIndex = newState
	sm.Current = cur
	sm.AdvanceGuard(common.TickFromContext(ctx))

	if err := cur.Enter(ctx, sm); err != nil {
		return tracing.Error(ctx, err)
	}

	return nil
}

// Receive processes an incoming message by routing it to the active state
// handler.
func (sm *SM) Receive(ctx context.Context, msg Envelope) {
	sm.Current.Receive(ctx, sm, msg)
}

// Start sets the state machine to its first (starting) state
func (sm *SM) Start(ctx context.Context) error {
	cur := sm.States[sm.FirstState]

	sm.CurrentIndex = sm.FirstState
	sm.Current = cur

	if err := cur.Enter(ctx, sm); err != nil {
		return tracing.Error(ctx, err)
	}
	return nil
}
