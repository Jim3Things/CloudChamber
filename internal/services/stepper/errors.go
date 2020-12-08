package stepper

import (
	"errors"
	"fmt"
	"time"
)

var (
	errAlreadyStarted = errors.New("state machine already started")
	errInvalidMessage = errors.New("invalid message encountered")
	errDelayCanceled  = errors.New("the delay operation was canceled")
)

type errDelayMustBeZero struct {
	actual time.Duration
}

func (e *errDelayMustBeZero) Error() string {
	return fmt.Sprintf("the MeasuredDelay must be zero for this policy.  It is %v", e.actual)
}

type errDelayMustBePositive struct {
	actual time.Duration
}

func (e *errDelayMustBePositive) Error() string {
	return fmt.Sprintf("the MeasuredDelay must be positive for this policy.  It is %v", e.actual)
}

type errInvalidPolicy struct {
	policy int
}

func (e *errInvalidPolicy) Error() string {
	return fmt.Sprintf("an invalid policy (%d) encountered", e.policy)
}

type errPolicyTooLate struct {
	guard int64
	current int64
}

func (e *errPolicyTooLate) Error() string {
	return fmt.Sprintf(
		"the SetPolicy operation expects to replace policy version %d, " +
			"but the current policy version is %d",
			e.guard,
			e.current)
}
