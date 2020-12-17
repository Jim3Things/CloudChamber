package inventory

import (
	"errors"
)

var (
	ErrCableStuck = errors.New("cable is faulted")

	ErrTooLate = errors.New("inventory element modified after the requested time")

	ErrNoOperation = errors.New("repair operation specified the current state, no change occurred")

	ErrAlreadyStarted = errors.New("rack simulation has already started")
)
