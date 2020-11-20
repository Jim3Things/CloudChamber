package messages

import (
	"context"

	"github.com/Jim3Things/CloudChamber/internal/sm"
)

type StartSim struct {
	EnvelopeState
}

func NewStartSim(ctx context.Context, ch chan *sm.Response) *StartSim {
	msg := &StartSim{}
	msg.Initialize(ctx, ch)
	msg.Tag = TagStartSim

	return msg
}

type StopSim struct {
	EnvelopeState
}

func NewStopSim(ctx context.Context, ch chan *sm.Response) *StopSim {
	msg := &StopSim{}
	msg.Initialize(ctx, ch)
	msg.Tag = TagStopSim

	return msg
}

