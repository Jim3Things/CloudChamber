package sm

import (
	"context"

	"github.com/Jim3Things/CloudChamber/internal/common"
	"github.com/Jim3Things/CloudChamber/internal/tracing"
)

// SimpleSMState defines the methods used for state actions and transitions.
type SimpleSMState interface {

	// Enter is called when a state transition moves to this state
	Enter(ctx context.Context, sm *SimpleSM) error

	// Receive is called on the active start implementation when a new
	// incoming message arrives
	Receive(ctx context.Context, sm *SimpleSM, msg interface{}, ch chan interface{})

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
func (*NullState) Receive(context.Context, *SimpleSM, interface{}, chan interface{}) {}

// Leave is the default (no-action) implementation.
func (*NullState) Leave(context.Context, *SimpleSM) {}

// Name returns the string identifying the null state
func (x *NullState) Name() string { return "NullState" }

// SimpleSM defines a simplified state machine structure. It assumes that the
// issues of concurrency and lifecycle management are handled by some external
// logic.
type SimpleSM struct {

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

	// At is the simulated time tick when the current state was entered.
	At int64
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
	}
}

// ChangeState changes the current state.  Leave the old state, try to
// enter the new state, and declare that state as current if successful.
func (sm *SimpleSM) ChangeState(ctx context.Context, newState int) error {
	tracing.Infof(ctx, "Change state to %q", sm.States[newState].Name())
	cur := sm.Current
	cur.Leave(ctx, nil)

	cur = sm.States[newState]
	if err := cur.Enter(ctx, nil); err != nil {
		return tracing.Error(ctx, err)
	}

	sm.CurrentIndex = newState
	sm.Current = cur
	sm.At = common.TickFromContext(ctx)
	return nil
}

// Receive processes an incoming message by routing it to the active state
// handler.
func (sm *SimpleSM) Receive(ctx context.Context, msg interface{}, ch chan interface{}) {
	sm.Current.Receive(ctx, sm, msg, ch)
}

// Start sets the state machine to its first (starting) state
func (sm *SimpleSM) Start(ctx context.Context) error {
	cur := sm.States[sm.FirstState]
	if err := cur.Enter(ctx, nil); err != nil {
		return tracing.Error(ctx, err)
	}

	sm.CurrentIndex = sm.FirstState
	sm.Current = cur
	return nil
}
