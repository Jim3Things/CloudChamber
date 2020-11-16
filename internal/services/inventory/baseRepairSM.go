package inventory

import (
	"context"

	"github.com/Jim3Things/CloudChamber/internal/common"
	"github.com/Jim3Things/CloudChamber/internal/services/inventory/messages"
	"github.com/Jim3Things/CloudChamber/internal/sm"
)


// nullRepairAction provides the default implementations for the functions
// that handle the different repair and status messages, as well as the
// common routing logic used whenever a new message is received.
type nullRepairAction struct {
	sm.NullState
}

// Power returns an unexpected message to a setPower message.
func (s *nullRepairAction) Power(ctx context.Context, _ *sm.SimpleSM, msg *messages.SetPower) {
	msg.GetCh() <- messages.UnexpectedMessageResponse(s, common.TickFromContext(ctx), msg)
}

// Connect returns an unexpected message to a setConnection message.
func (s *nullRepairAction) Connect(ctx context.Context, _ *sm.SimpleSM, msg *messages.SetConnection) {
	msg.GetCh() <- messages.UnexpectedMessageResponse(s, common.TickFromContext(ctx), msg)
}

// Timeout ignores any timer expiration notification, as there must not be any
// outstanding timers for it to get to this implementation.
func (s *nullRepairAction) Timeout(_ context.Context, _ *sm.SimpleSM, _ *messages.TimerExpiry) {}

// handleMsg performs the common routing for all incoming messages.  They are
// routed to the known handler functions for messages that are known, and any
// other messages get an unexpected message error.
func (s *nullRepairAction) handleMsg(
	ctx context.Context,
	machine *sm.SimpleSM,
	state messages.RepairActionState,
	msg sm.Envelope) {

	switch body := msg.(type) {
	case messages.RepairMessage:
		body.Do(ctx, machine, state)

	case messages.StatusMessage:
		body.GetStatus(ctx, machine, state)

	default:
		msg.GetCh() <- messages.UnexpectedMessageResponse(state, common.TickFromContext(ctx), body)
	}
}

// dropRepairAction provides the default implementations for the functions
// that handle the different repair and status messages, as well as the
// common routing logic used whenever a new message is received and the
// element is in a state that prevents any processing.
type dropRepairAction struct {
	nullRepairAction
}

// power returns a dropped message
func (s *dropRepairAction) power(ctx context.Context, _ *sm.SimpleSM, msg *messages.SetPower) {
	msg.GetCh() <- messages.DroppedResponse(common.TickFromContext(ctx))
}

// connect returns a dropped message
func (s *dropRepairAction) connect(ctx context.Context, _ *sm.SimpleSM, msg *messages.SetConnection) {
	msg.GetCh() <- messages.DroppedResponse(common.TickFromContext(ctx))
}
