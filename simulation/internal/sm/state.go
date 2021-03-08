package sm

import (
	"context"

	"github.com/Jim3Things/CloudChamber/simulation/internal/common"
)

// State defines the methods used for state actions and transitions.
type State interface {

	// Enter is called when a state transition moves to this state
	Enter(ctx context.Context, sm *SM) error

	// Receive is called on the active start implementation when a new
	// incoming message arrives
	Receive(ctx context.Context, machine *SM, msg Envelope)

	// Leave is called when a state transition moves away from this state
	Leave(ctx context.Context, sm *SM, nextState StateIndex)
}

// NullState is the default implementation of an SM state
type NullState struct{}

// Enter is the default (no-action) implementation.
func (*NullState) Enter(context.Context, *SM) error { return nil }

// Receive is the default (no-action) implementation.
func (*NullState) Receive(ctx context.Context, machine *SM, msg Envelope) {
	msg.Ch() <- UnexpectedMessageResponse(machine, common.TickFromContext(ctx), msg)
}

// Leave is the default (no-action) implementation.
func (*NullState) Leave(context.Context, *SM, string) {}

// TerminalEnter is the standard terminal state Enter handler, which marks the
// state machine as terminated.
func TerminalEnter(_ context.Context, machine *SM) error {
	machine.Terminated = true
	return nil
}
