package messages

import (
	"context"

	"github.com/Jim3Things/CloudChamber/internal/sm"
)

// Reset is the message to force a reset of the simulated time service back to
// its starting point.
type Reset struct {
	sm.EnvelopeState
}

func NewReset(ctx context.Context, ch chan *sm.Response) *Reset {
	msg := &Reset{}
	msg.Initialize(ctx, TagReset, ch)

	return msg
}

