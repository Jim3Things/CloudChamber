// This module contains the common support methods and base structures for
// actor state machine handling.

package sm

import (
    "context"

    "github.com/AsynkronIT/protoactor-go/actor"
    trc "go.opentelemetry.io/otel/api/trace"

    trace "github.com/Jim3Things/CloudChamber/internal/tracing/server"
    "github.com/Jim3Things/CloudChamber/pkg/protos/common"
)

// Define the common interface for a state in the state machine
type State interface {
    // Define 'Receive' for event notification
    actor.Actor

    // Function to handle transition into this state
    Enter(ctx actor.Context, c context.Context, span trc.Span) error

    // Function to handle transition out of this state
    Leave()
}

// Define the default (null) state, and the state handlers.  Provides a base
// implementation when a particular state does not require some aspect.
type EmptyState struct {
}

func (*EmptyState) Enter(_ actor.Context, _ context.Context, _ trc.Span) error { return nil }
func (*EmptyState) Receive(_ actor.Context) {}
func (*EmptyState) Leave()                  {}

// Define the common state machine structure
type SM struct {
    Current     int             // Index to the current state
    Behavior    actor.Behavior  // Current behavior
    States      map[int]State   // Set of states, indexed by state number
    StateNames  map[int]string  // Names of the states, indexed by state number
}

// Common method to change the current state.  Leave the old state, try to
// enter the new state, and declare that state as current if successful.
func (sm *SM) ChangeState(
        c context.Context,
        span trc.Span,
        ctx actor.Context,
        latest int64,
        newState int) error {
    trace.Infof(c, span, latest, "Change state to %q", sm.StateNames[newState])
    cur := sm.States[sm.Current]
    cur.Leave()

    cur = sm.States[newState]
    if err := cur.Enter(ctx, c, span); err != nil {
        return trace.Error(c, span, latest, err)
    }

    sm.Current = newState
    sm.Behavior.Become(cur.Receive)
    return nil
}

// Set the state machine to its first, and starting, state
func (sm *SM) Initialize(c context.Context, span trc.Span, firstState int) error {
    cur := sm.States[firstState]
    if err := cur.Enter(nil, c, span); err != nil {
        return trace.Error(c, span, 0, err)
    }

    sm.Current = firstState
    sm.Behavior.Become(cur.Receive)
    return nil
}

// Helper method that responds to the sender with an error message
func (sm *SM) RespondWithError(_ context.Context, _ trc.Span, ctx actor.Context, err error) {
    ctx.Respond(&common.Completion{
        Error: err.Error(),
    })
}

// Helper method that gets the textual name of the current state
func (sm *SM) GetStateName() string {
    n, ok := sm.StateNames[sm.Current]
    if !ok { n = "<unknown>"}
    return n
}

