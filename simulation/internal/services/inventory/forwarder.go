package inventory

// This file contains common support functions for the inventory elements that
// forward along cable lines to blades.

import (
	"context"

	"github.com/Jim3Things/CloudChamber/simulation/internal/services/inventory/messages"
	"github.com/Jim3Things/CloudChamber/simulation/internal/sm"
	"github.com/Jim3Things/CloudChamber/simulation/internal/tracing"
)

// processInvalidTarget issues an invalid target error response and closes the
// response channel.
func processInvalidTarget(ctx context.Context, msg sm.Envelope, target string, occursAt int64) {
	ch := msg.Ch()
	defer close(ch)

	tracing.Warn(ctx, "No connection for %s was found.", target)

	ch <- messages.InvalidTargetResponse(occursAt)
}
