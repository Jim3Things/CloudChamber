package messages

// This file contains helper functions that simplify the creation of response
// messages to repair operations.

import (
	"context"

	"github.com/Jim3Things/CloudChamber/simulation/internal/sm"
	"github.com/Jim3Things/CloudChamber/simulation/internal/tracing"
    "github.com/Jim3Things/CloudChamber/simulation/pkg/errors"
)

// InvalidTargetResponse constructs a failure response message with an invalid
// target error code.
func InvalidTargetResponse(occursAt int64) *sm.Response {
	return sm.FailedResponse(occursAt, errors.ErrInvalidTarget)
}

// DropMessage is an action state processor that closes the channel without
// issuing any message.  This indicates that the state machine did not process
// the request, including finding an error to send back.  The closure avoids
// real time delays in waiting for a response, and instead to move any such
// delay into simulated time.
func DropMessage(ctx context.Context, machine *sm.SM, msg sm.Envelope) bool {
	_ = tracing.Error(ctx, "Unexpected message %v arrived in state %q", msg, machine.CurrentIndex)

	ch := msg.Ch()
	if ch != nil {
		close(ch)
	}

	return true
}
