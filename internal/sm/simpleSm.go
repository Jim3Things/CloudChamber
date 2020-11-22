package sm

import (
	"context"

	"github.com/Jim3Things/CloudChamber/internal/common"
	"github.com/Jim3Things/CloudChamber/internal/tracing"
)

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

	// Names holds the map of known state index values to descriptive strings
	Names map[int]string

	// Parent points to the structure that holds this state machine, and likely
	// holds global context that the state actions need.
	Parent interface{}

	// Terminated is true if the state machine has reached its final state.
	Terminated bool
}

// StateDecl defines the type expected for a state declaration decorator when
// creating a new SimpleSM instance
type StateDecl func() (int, bool, string, SimpleSMState)

// WithState is a decorator that defines a state in the state machine
func WithState(
	id int,
	name string,
	onEnter EnterFunc,
	actions []ActionEntry,
	other ActionFunc,
	onLeave LeaveFunc) StateDecl {
	return func() (int, bool, string, SimpleSMState) {
		return id, false, name, NewActionState(actions, other, onEnter, onLeave)
	}
}

// WithFirstState is a decorator that defines the starting state for the state
// machine
func WithFirstState(
	id int,
	name string,
	onEnter EnterFunc,
	actions []ActionEntry,
	other ActionFunc,
	onLeave LeaveFunc) StateDecl {
	return func() (int, bool, string, SimpleSMState) {
		return id, true, name, NewActionState(actions, other, onEnter, onLeave)
	}
}

// NewSimpleSM creates a new state machine instance with the associated
// parent instance reference, as well as the state declarations.
func NewSimpleSM(parent interface{}, decls ...StateDecl) *SimpleSM {
	states := make(map[int]SimpleSMState)
	names := make(map[int]string)

	firstState := 0

	for _, decl := range decls {
		stateId, first, name, instance := decl()
		states[stateId] = instance
		names[stateId] = name

		if first {
			firstState = stateId
		}
	}

	return &SimpleSM{
		CurrentIndex: firstState,
		Current:      states[firstState],
		FirstState:   firstState,
		States:       states,
		Names:        names,
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
		sm.GetCurrentStateName(),
		sm.GetStateName(newState))

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
func (sm *SimpleSM) Receive(ctx context.Context, msg Envelope) {
	sm.Current.Receive(ctx, sm, msg)
}

// Start sets the state machine to its first (starting) state
func (sm *SimpleSM) Start(ctx context.Context) error {
	cur := sm.States[sm.FirstState]

	sm.CurrentIndex = sm.FirstState
	sm.Current = cur

	if err := cur.Enter(ctx, sm); err != nil {
		return tracing.Error(ctx, err)
	}
	return nil
}

// GetStateName returns the text name of the state identified by the supplied
// index.
func (sm *SimpleSM) GetStateName(id int) string {
	return sm.Names[id]
}

// GetCurrentStateName gets the descriptive name for the current state.
func (sm *SimpleSM) GetCurrentStateName() string {
	return sm.Names[sm.CurrentIndex]
}
