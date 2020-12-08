package sm

import (
	"context"
)


const (
	// TagStartSM identifies the message used to start a state machine
	// goroutine
	TagStartSM = -1

	// TagStopSM identifies the message used to terminate a state machine
	// goroutine
	TagStopSM = -2
)

// StartSM is the message used to signal the start of a state machine.
type StartSM struct {
	EnvelopeState
}

func NewStartSM(ctx context.Context, ch chan *Response) *StartSM {
	msg := &StartSM{}
	msg.Initialize(ctx, TagStartSM, ch)

	return msg
}

// StopSM is the message used to signal the termination of a stat machine.
type StopSM struct {
	EnvelopeState
}

func NewStopSM(ctx context.Context, ch chan *Response) *StopSM {
	msg := &StopSM{}
	msg.Initialize(ctx, TagStopSM, ch)

	return msg
}
