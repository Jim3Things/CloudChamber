package sm

import (
	"context"

	"github.com/Jim3Things/CloudChamber/internal/common"
)

// SimpleSMState defines the methods used for state actions and transitions.
type SimpleSMState interface {

	// Enter is called when a state transition moves to this state
	Enter(ctx context.Context, sm *SimpleSM) error

	// Receive is called on the active start implementation when a new
	// incoming message arrives
	Receive(ctx context.Context, machine *SimpleSM, msg Envelope)

	// Leave is called when a state transition moves away from this state
	Leave(ctx context.Context, sm *SimpleSM, nextState int)
}

// NullState is the default implementation of a simple SM state
type NullState struct{}

// Enter is the default (no-action) implementation.
func (*NullState) Enter(context.Context, *SimpleSM) error { return nil }

// Receive is the default (no-action) implementation.
func (*NullState) Receive(ctx context.Context, machine *SimpleSM, msg Envelope) {
	msg.GetCh() <- UnexpectedMessageResponse(machine, common.TickFromContext(ctx), msg)
}

// Leave is the default (no-action) implementation.
func (*NullState) Leave(context.Context, *SimpleSM, int) {}

// TerminalEnter is the standard terminal state Enter handler, which marks the
// state machine as terminated.
func TerminalEnter(_ context.Context, machine *SimpleSM) error {
	machine.Terminated = true
	return nil
}
