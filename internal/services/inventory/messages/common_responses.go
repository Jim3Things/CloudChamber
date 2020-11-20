package messages

// This file contains helper functions that simplify the creation of response
// messages to repair operations.

import (
	"errors"
	"fmt"

	"github.com/Jim3Things/CloudChamber/internal/sm"
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

func InvalidTargetResponse(occursAt int64) *sm.Response {
	return FailedResponse(occursAt, ErrInvalidTarget)
}

// UnexpectedMessageResponse constructs a failure response for the case where
// the incoming request arrives when it is unexpected by the state machine.
func UnexpectedMessageResponse(machine *sm.SimpleSM, occursAt int64, body interface{}) *sm.Response {
	return &sm.Response{
		Err: &sm.UnexpectedMessage{
			Msg:   fmt.Sprintf("%v", body),
			State: machine.GetCurrentStateName(),
		},
		At:  occursAt,
		Msg: nil,
	}
}
