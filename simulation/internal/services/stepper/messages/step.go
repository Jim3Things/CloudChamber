package messages

import (
	"context"

	"github.com/Jim3Things/CloudChamber/simulation/internal/sm"
)

// Step is the message to request that the simulated time move forward by 1
// tick.
type Step struct {
	sm.EnvelopeState
}

func NewStep(ctx context.Context, ch chan *sm.Response) *Step {
	msg := &Step{}
	msg.Initialize(ctx, TagStep, ch)

	return msg
}

// AutoStep is the message that requests the simulated time move forward by
// one step as a result of a wall clock timer expiring.  As this event may
// be in-flight when the policy authorizing it is changed, the message
// includes the policy generation number as a guard.  That number must match
// the current policy generation for it to be processed.
type AutoStep struct {
	sm.EnvelopeState

	Guard int64
}

func NewAutoStep(
	ctx context.Context,
	guard int64,
	ch chan *sm.Response) *AutoStep {

	msg := &AutoStep{}
	msg.Initialize(ctx, TagAutoStep, ch)
	msg.Guard = guard

	return msg
}
