package inventory

import (
	"errors"
)

var (
	ErrCableStuck = errors.New("cable is faulted")
	ErrTooLate    = errors.New("inventory element modified after the requested time")

	ErrNoOperation = errors.New("repair operation specified the current state, no change occurred")

	ErrRepairMessageDropped = errors.New("repair message dropped")
	ErrInvalidTarget        = errors.New("invalid target specified, request ignored")
)
