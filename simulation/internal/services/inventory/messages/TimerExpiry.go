package messages

import (
	"context"
	"fmt"

	"github.com/Jim3Things/CloudChamber/simulation/internal/sm"
	"github.com/Jim3Things/CloudChamber/simulation/pkg/errors"
)

// TimerExpiry is the message used to notify a simulated inventory element that
// a specific timer, designated by the Id field, has expired.
type TimerExpiry struct {
	messageBase

	// Id is the value used to identify which outstanding timer has expired.
	Id int64

	// timer expiration context - what the state machine needs, if anything, to
	// work On the expiration notice.
	Body sm.Envelope
}

// NewTimerExpiry creates a new TimerExpiry message.
func NewTimerExpiry(
	target *MessageTarget,
	guard int64,
	id int64,
	body sm.Envelope,
	ch chan *sm.Response) *TimerExpiry {
	msg := &TimerExpiry{}

	msg.InitializeNoLink(TagTimerExpiry, ch)
	msg.Target = target
	msg.Guard = guard
	msg.Id = id
	msg.Body = body

	return msg
}

// SendVia forwards the timer expiration directly to the target element.  This
// either sends the enclosed body, or the outer TimerExpiry message itself, if
// no body is present.
func (m *TimerExpiry) SendVia(ctx context.Context, r viaSender) error {
	var msg sm.Envelope = m
	if m.Body != nil {
		msg = m.Body
	}

	t := m.Target
	id := t.ElementId()

	switch {
	case t.IsPdu():
		return r.ToPdu(ctx, id, msg)

	case t.IsTor():
		return r.ToTor(ctx, id, msg)

	case t.IsBlade():
		return r.ToBlade(ctx, id, msg)

	default:
		return errors.ErrInvalidTarget
	}
}

// String provides a formatted description of the message.
func (m *TimerExpiry) String() string {
	return fmt.Sprintf("Expiration notice of timer id %d for %q", m.Id, m.Target.Describe())
}
