package inventory

// This file contains helper functions that simplify the creation of response
// messages to repair operations.

import (
	"errors"

	"github.com/Jim3Things/CloudChamber/internal/sm"
)

var (
	ErrStuck   = errors.New("cable is faulted")
	ErrTooLate = errors.New("cable modified after the requested time")

	ErrRepairMessageDropped = errors.New("repair message dropped")
	ErrInvalidTarget        = errors.New("invalid target specified, request ignored")
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
