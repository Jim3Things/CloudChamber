package messages

import (
	"context"

	"github.com/Jim3Things/CloudChamber/internal/sm"
)

// Now is the message that requests the current simulated time.
type Now struct {
	sm.EnvelopeState
}

func NewNow(ctx context.Context, ch chan *sm.Response) *Now {
	msg := &Now{}
	msg.Initialize(ctx, TagNow, ch)

	return msg
}
