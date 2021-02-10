package messages

import (
	"context"
	"time"

	"github.com/Jim3Things/CloudChamber/simulation/internal/sm"
)

// GetStatus is the message to request the simulated time service's current
// execution status.
type GetStatus struct {
	sm.EnvelopeState
}

func NewGetStatus(ctx context.Context, ch chan *sm.Response) *GetStatus {
	msg := &GetStatus{}
	msg.Initialize(ctx, TagGetStatus, ch)

	return msg
}

// StatusResponseBody contains the simulated time service's current execution
// status.
type StatusResponseBody struct {
	Guard int64

	Waiters int64

	MeasuredDelay time.Duration

	Policy int
}

func NewStatusResponseBody(
	guard int64,
	waiters int64,
	measuredDelay time.Duration,
	policy int) *StatusResponseBody {

	return &StatusResponseBody{
		Guard:         guard,
		Waiters:       waiters,
		MeasuredDelay: measuredDelay,
		Policy:        policy,
	}
}
