package inventory

// This file contains helper functions that simplify the creation of response
// messages to repair operations.

import (
	"fmt"

	"github.com/Jim3Things/CloudChamber/internal/sm"
)

// droppedResponse constructs a dropped response message with the correct time
// and target.
func droppedResponse(occursAt int64) *sm.Response {
	return &sm.Response{
		Err: ErrRepairMessageDropped,
		At:  occursAt,
		Msg: nil,
	}
}

// failedResponse constructs a failure response message with the correct time,
// target, and reason.
func failedResponse(occursAt int64, err error) *sm.Response {
	return &sm.Response{
		Err: err,
		At:  occursAt,
		Msg: nil,
	}
}

// successResponse constructs a success response message with the correct time
// and target.
func successResponse(occursAt int64) *sm.Response {
	return &sm.Response{
		Err: nil,
		At:  occursAt,
		Msg: nil,
	}
}

// unexpectedMessageResponse constructs a failure response for the case where
// the incoming request arrives when it is unexpected by the state machine.
func unexpectedMessageResponse(s sm.SimpleSMState, occursAt int64, body interface{}) *sm.Response {
	return &sm.Response{
		Err: &sm.UnexpectedMessage{
			Msg:   fmt.Sprintf("%v", body),
			State: s.Name(),
		},
		At:  occursAt,
		Msg: nil,
	}
}
