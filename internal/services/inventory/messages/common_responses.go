package messages

// This file contains helper functions that simplify the creation of response
// messages to repair operations.

import (
	"context"
	"errors"

	"github.com/Jim3Things/CloudChamber/internal/common"
	"github.com/Jim3Things/CloudChamber/internal/sm"
	"github.com/Jim3Things/CloudChamber/internal/tracing"
)

var ErrRepairMessageDropped = errors.New("repair message dropped")

// DroppedResponse constructs a dropped response message with the correct time
// and target.
func DroppedResponse(occursAt int64) *sm.Response {
	return &sm.Response{
		Err: ErrRepairMessageDropped,
		At:  occursAt,
		Msg: nil,
	}
}

// FailedResponse constructs a failure response message with the correct time,
// target, and reason.
func FailedResponse(occursAt int64, err error) *sm.Response {
	return &sm.Response{
		Err: err,
		At:  occursAt,
		Msg: nil,
	}
}

// SuccessResponse constructs a success response message with the correct time
// and target.
func SuccessResponse(occursAt int64) *sm.Response {
	return &sm.Response{
		Err: nil,
		At:  occursAt,
		Msg: nil,
	}
}

// InvalidTargetResponse constructs a failure response message with an invalid
// target error code.
func InvalidTargetResponse(occursAt int64) *sm.Response {
	return FailedResponse(occursAt, ErrInvalidTarget)
}

// DropMessage is an action state processor that issues a response message to
// indicate that the state machine did not process the request, nor did it
// logically issues a reply.  This message is used to avoid real time delays
// in waiting for a response, and instead to move any such delay into simulated
// time.
func DropMessage(ctx context.Context, machine *sm.SimpleSM, msg sm.Envelope) bool {
	_ = tracing.Error(ctx, "Unexpected message %v arrived in state %q", msg, machine.GetCurrentStateName())

	ch := msg.GetCh()

	if ch != nil {
		ch <- DroppedResponse(common.TickFromContext(ctx))
	}

	return true
}
