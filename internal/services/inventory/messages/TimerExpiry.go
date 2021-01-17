package messages

import (
	"context"
	"fmt"

	"github.com/Jim3Things/CloudChamber/internal/sm"
    "github.com/Jim3Things/CloudChamber/pkg/errors"
)

// TimerExpiry is the message used to notify a simulated inventory element that
// a specific timer, designated by the Id field, has expired.
type TimerExpiry struct {
	messageBase

	// Id is the value used to identify which outstanding timer has expired.
	Id int64

	// timer expiration context - what the state machine needs, if anything, to
	// work On the expiration notice.
	Body *messageBase
}

// NewTimerExpiry creates a new TimerExpiry message.
func NewTimerExpiry(
	target *MessageTarget,
	guard int64,
	id int64,
	body *messageBase,
	ch chan *sm.Response) *TimerExpiry {
	msg := &TimerExpiry{}

	msg.InitializeNoLink(TagTimerExpiry, ch)
	msg.Target = target
	msg.Guard = guard
	msg.Id = id
	msg.Body = body

	return msg
}

// SendVia forwards the timer expiration directly to the target element.
func (m *TimerExpiry) SendVia(ctx context.Context, r viaSender) error {
	if m.Target.IsPdu() {
		return r.ViaPDU(ctx, m)
	}

	if m.Target.IsTor() {
		return r.ViaTor(ctx, m)
	}

	id, ok := m.Target.BladeID()
	if !ok {
		return errors.ErrInvalidTarget
	}

	return r.ViaBlade(ctx, id, m)
}

// String provides a formatted description of the message.
func (m *TimerExpiry) String() string {
	return fmt.Sprintf("Expiration notice of timer id %d for %q", m.Id, m.Target.Describe())
}
