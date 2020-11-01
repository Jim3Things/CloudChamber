package inventory

import (
	"context"

	"github.com/Jim3Things/CloudChamber/internal/common"
	"github.com/Jim3Things/CloudChamber/internal/sm"
)

// repairActionState is the abstract definition that all inventory state
// machines must implement.
type repairActionState interface {
	sm.SimpleSMState

	// power is the function that responds to a setPower message.
	power(ctx context.Context, sm *sm.SimpleSM, msg *setPower)

	// connect is the function that responds to a setConnection message
	connect(ctx context.Context, sm *sm.SimpleSM, msg *setConnection)
}

// nullRepairAction provides the default implementations for the functions
// that handle the different repair and status messages, as well as the
// common routing logic used whenever a new message is received.
type nullRepairAction struct {
	sm.NullState
}

// power returns an unexpected message to a setPower message.
func (s *nullRepairAction) power(ctx context.Context, _ *sm.SimpleSM, msg *setPower) {
	msg.GetCh() <- unexpectedMessageResponse(s, common.TickFromContext(ctx), msg)
}

// connect returns an unexpected message to a setConnection message.
func (s *nullRepairAction) connect(ctx context.Context, _ *sm.SimpleSM, msg *setConnection) {
	msg.GetCh() <- unexpectedMessageResponse(s, common.TickFromContext(ctx), msg)
}

// handleMsg performs the common routing for all incoming messages.  They are
// routed to the known handler functions for messages that are known, and any
// other messages get an unexpected message error.
func (s *nullRepairAction) handleMsg(
	ctx context.Context,
	machine *sm.SimpleSM,
	state repairActionState,
	msg sm.Envelope) {

	switch body := msg.(type) {
	case repairMessage:
		body.Do(ctx, machine, state)

	case statusMessage:
		body.GetStatus(ctx, machine, state)

	default:
		msg.GetCh() <- unexpectedMessageResponse(state, common.TickFromContext(ctx), body)
	}
}
