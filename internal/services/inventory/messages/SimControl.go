package messages

import (
	"context"

	"github.com/Jim3Things/CloudChamber/internal/sm"
)

// StartSim is the message used to signal the start of a rack-level simulation.
type StartSim struct {
	EnvelopeState
}

func NewStartSim(ctx context.Context, ch chan *sm.Response) *StartSim {
	msg := &StartSim{}
	msg.Initialize(ctx, TagStartSim, ch)

	return msg
}

// StopSim is the message used to signal the termination of a rack-level
// simulation.
type StopSim struct {
	EnvelopeState
}

func NewStopSim(ctx context.Context, ch chan *sm.Response) *StopSim {
	msg := &StopSim{}
	msg.Initialize(ctx, TagStopSim, ch)

	return msg
}
