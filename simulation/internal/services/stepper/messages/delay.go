package messages

import (
	"context"

	"github.com/Jim3Things/CloudChamber/simulation/internal/sm"
)

// Delay is the message to request that the a response be sent only when the
// simulated time passes the value stored as DueTime.
type Delay struct {
	sm.EnvelopeState

	DueTime int64
}

func NewDelay(ctx context.Context, dueTime int64, ch chan *sm.Response) *Delay {
	msg := &Delay{}
	msg.Initialize(ctx, TagDelay, ch)
	msg.DueTime = dueTime

	return msg
}

