package inventory

import (
	"context"
	"fmt"

	"github.com/Jim3Things/CloudChamber/internal/sm"
)

// +++ Base message and interfaces

type messageBase struct {
	envelopeState

	target *messageTarget

	guard int64
}

type messageForwarder interface {
	SendVia(ctx context.Context, r *rack) error
}

type messageAction interface {
	Do(ctx context.Context, sm *sm.SimpleSM, s repairActionState)
}

type repairMessage interface {
	messageForwarder
	messageAction
}

type messageStatus interface {
	GetStatus(ctx context.Context, sm *sm.SimpleSM, s repairActionState)
}

type statusMessage interface {
	messageForwarder
	messageStatus
}

// --- Base message and interfaces

// timerExpiry is the message used to notify a simulated inventory element that
// a specific timer, designated by the id field, has expired.
type timerExpiry struct {
	messageBase

	// id is the value used to identify which outstanding timer has expired.
	id int64

	// timer expiration context - what the state machine needs, if anything, to
	// work on the expiration notice.
	body *messageBase
}

func newTimerExpiry(
	ctx context.Context,
	target *messageTarget,
	guard int64,
	id int64,
	body *messageBase,
	ch chan *sm.Response) *timerExpiry {
	msg := &timerExpiry{}

	msg.Initialize(ctx, ch)
	msg.target = target
	msg.guard = guard
	msg.id = id
	msg.body = body

	return msg
}

// SendVia forwards the timer expiration directly to the target element.
func (m *timerExpiry) SendVia(ctx context.Context, r *rack) error {
	if m.target.isPdu() {
		return r.viaPDU(ctx, m)
	}

	if m.target.isTor() {
		return r.viaTor(ctx, m)
	}

	id, ok := m.target.bladeID()
	if !ok {
		return ErrInvalidTarget
	}

	return r.viaBlade(ctx, id, m)
}

// Do executes the action to handle the timer expired notification.
func (m *timerExpiry) Do(ctx context.Context, sm *sm.SimpleSM, s repairActionState) {
	s.timeout(ctx, sm, m)
}

// String provides a formatted description of the message.
func (m *timerExpiry) String() string {
	return fmt.Sprintf("Expiration notice of timer id %d for %q", m.id, m.target.describe())
}

// setPower is the repair message that directs a change in the simulated power
// setting.
type setPower struct {
	messageBase

	// on designates whether the simulated power is to be on or off.
	on bool
}

// newSetPower creates a new setPower message with the values provided.
func newSetPower(
	ctx context.Context,
	target *messageTarget,
	guard int64,
	on bool,
	ch chan *sm.Response) *setPower {
	msg := &setPower{}

	msg.Initialize(ctx, ch)
	msg.target = target
	msg.guard = guard
	msg.on = on

	return msg
}

// SendVia forwards the repair message to the rack's PDU for processing.  This
// may or may not be the final destination for the message.
func (m *setPower) SendVia(ctx context.Context, r *rack) error {
	return r.viaPDU(ctx, m)
}

// Do executes the action to handle the power change request.
func (m *setPower) Do(ctx context.Context, sm *sm.SimpleSM, s repairActionState) {
	s.power(ctx, sm, m)
}

// String provides a formatted description of the message.
func (m *setPower) String() string {
	return fmt.Sprintf("Set the power %s for %s", aOrB(m.on, "on", "off"), m.target.describe())
}

// setConnection is the repair message that directs a change in the simulated
// network connection setting.
type setConnection struct {
	messageBase

	// enabled designates whether the simulated network connection is to be
	// enabled or disabled.
	enabled bool
}

// newSetConnection creates a new setConnection message with the values provided.
func newSetConnection(
	ctx context.Context,
	target *messageTarget,
	guard int64,
	enabled bool,
	ch chan *sm.Response) *setConnection {
	msg := &setConnection{}

	msg.Initialize(ctx, ch)
	msg.target = target
	msg.guard = guard
	msg.enabled = enabled

	return msg
}

// SendVia forwards the repair message to the rack's TOR for processing.  This
// is not the final destination for the message.
func (m *setConnection) SendVia(ctx context.Context, r *rack) error {
	return r.viaTor(ctx, m)
}

// Do executes the action to handle the network connection change request.
func (m *setConnection) Do(ctx context.Context, sm *sm.SimpleSM, s repairActionState) {
	s.connect(ctx, sm, m)
}

// String provides a formatted description of the message.
func (m *setConnection) String() string {
	return fmt.Sprintf(
		"%s the network connection to %s",
		aOrB(m.enabled, "Enable", "Disable"),
		m.target.describe())
}

// getStatus is the message used to obtain the target element's actual
// simulated operational state.
type getStatus struct {
	messageBase
}

func (g getStatus) SendVia(ctx context.Context, r *rack) error {
	panic("implement me")
}

func (g getStatus) GetStatus(ctx context.Context, sm *sm.SimpleSM, s repairActionState) {
	panic("implement me")
}
