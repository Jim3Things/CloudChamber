package messages

import (
	"context"
	"fmt"

	"github.com/Jim3Things/CloudChamber/simulation/internal/common"
	"github.com/Jim3Things/CloudChamber/simulation/internal/sm"
)

// SetPower is the repair message that directs a change in the simulated power
// setting.
type SetPower struct {
	messageBase

	// On designates whether the simulated power is to be On or off.
	On bool
}

// NewSetPower creates a new SetPower message with the values provided.
func NewSetPower(
	ctx context.Context,
	target *MessageTarget,
	guard int64,
	on bool,
	ch chan *sm.Response) *SetPower {
	msg := &SetPower{}

	msg.Initialize(ctx, TagSetPower, ch)
	msg.Target = target
	msg.Guard = guard
	msg.On = on

	return msg
}

// SendVia forwards the repair message to the rack's PDU for processing.  This
// may or may not be the final destination for the message.
func (m *SetPower) SendVia(ctx context.Context, r viaSender) error {
	return r.ViaPDU(ctx, m.Target, m)
}

// String provides a formatted description of the message.
func (m *SetPower) String() string {
	return fmt.Sprintf("Set the power %s for %s", common.AOrB(m.On, "On", "off"), m.Target.Describe())
}
