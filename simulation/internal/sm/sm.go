package sm

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/proto"

	"github.com/Jim3Things/CloudChamber/simulation/internal/common"
	"github.com/Jim3Things/CloudChamber/simulation/internal/tracing"
)

// StateIndex denotes that the value is used as an index into the state machine
// action states.
type StateIndex interface {
	fmt.Stringer
}

type invalidIndexType int32

func (s invalidIndexType) String() string { return "invalid" }

const invalidIndex invalidIndexType = 0

// SM defines a simplified state machine structure. It assumes that the issues
// of concurrency and lifecycle management are handled by some external logic.
type SM struct {
	common.Guarded

	// CurrentIndex holds the index to the current state
	CurrentIndex StateIndex

	// Current is a pointer to the current state
	Current State

	// FirstState is the index to the starting state
	FirstState StateIndex

	// States holds the map of known state index values to state implementations
	States map[StateIndex]State

	// Parent points to the structure that holds this state machine, and likely
	// holds global context that the state actions need.
	Parent interface{}

	// Terminated is true if the state machine has reached its final state.
	Terminated bool

	// EnteredAt is the simulated time tick when the current state was entered.
	EnteredAt int64

	// name is the string that can be used to identify this particular instance.
	Name string
}

type Persistable interface {
	Save() (proto.Message, error)
}

// StateDecl defines the type expected for a state declaration decorator when
// creating a new SM instance
type StateDecl func() (bool, StateIndex, State)

// WithState is a decorator that defines a state in the state machine
func WithState(
	name StateIndex,
	onEnter EnterFunc,
	actions []ActionEntry,
	other ActionFunc,
	onLeave LeaveFunc) StateDecl {
	return func() (bool, StateIndex, State) {
		return false, name, NewActionState(actions, other, onEnter, onLeave)
	}
}

// WithFirstState is a decorator that defines the starting state for the state
// machine
func WithFirstState(
	name StateIndex,
	onEnter EnterFunc,
	actions []ActionEntry,
	other ActionFunc,
	onLeave LeaveFunc) StateDecl {
	return func() (bool, StateIndex, State) {
		return true, name, NewActionState(actions, other, onEnter, onLeave)
	}
}

// NewSM creates a new state machine instance with the associated
// parent instance reference, as well as the state declarations.
func NewSM(parent interface{}, name string, decls ...StateDecl) *SM {
	states := make(map[StateIndex]State)

	var firstState StateIndex = invalidIndex

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
		EnteredAt:    0,
		Name:         name,
	}
}

// ChangeState changes the current state.  Leave the old state, try to
// enter the new state, and declare that state as current if successful.
func (sm *SM) ChangeState(ctx context.Context, newState StateIndex) error {
	tracing.Info(
		ctx,
		"Change state from %q to %q in state machine %q",
		sm.CurrentIndex,
		newState,
		sm.Name)

	cur := sm.Current
	cur.Leave(ctx, sm, newState)

	cur = sm.States[newState]
	sm.CurrentIndex = newState
	sm.Current = cur

	tick := common.TickFromContext(ctx)
	sm.EnteredAt = tick
	sm.AdvanceGuard(tick)

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

// Savable returns the SM state that can be usefully saved off and later
// restored as part of implementing a persistent state machine.
func (sm *SM) Savable() (StateIndex, int64, bool, int64) {
	return sm.CurrentIndex, sm.EnteredAt, sm.Terminated, sm.Guarded.Guard
}
