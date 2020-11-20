package inventory

import (
	"context"

	"github.com/Jim3Things/CloudChamber/internal/common"
	"github.com/Jim3Things/CloudChamber/internal/services/inventory/messages"
	"github.com/Jim3Things/CloudChamber/internal/sm"
	"github.com/Jim3Things/CloudChamber/internal/tracing"
)

func UnexpectedMessage(ctx context.Context, machine *sm.SimpleSM, msg sm.Envelope) bool {
	_ = tracing.Error(ctx, "Unexpected message %v arrived in state %q", msg, machine.GetCurrentStateName())

	ch := msg.GetCh()

	if ch != nil {
		ch <- messages.UnexpectedMessageResponse(machine, common.TickFromContext(ctx), msg)
	}

	return true
}

func DropMessage(ctx context.Context, machine *sm.SimpleSM, msg sm.Envelope) bool {
	_ = tracing.Error(ctx, "Unexpected message %v arrived in state %q", msg, machine.GetCurrentStateName())

	ch := msg.GetCh()

	if ch != nil {
		ch <- messages.DroppedResponse(common.TickFromContext(ctx))
	}

	return true
}
