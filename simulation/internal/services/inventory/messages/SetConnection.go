package messages

import (
	"context"
	"fmt"

	"github.com/Jim3Things/CloudChamber/simulation/internal/common"
	"github.com/Jim3Things/CloudChamber/simulation/internal/sm"
)

// SetConnection is the repair message that directs a change in the simulated
// network connection setting.
type SetConnection struct {
	messageBase

	// Enabled designates whether the simulated network connection is to be
	// Enabled or disabled.
	Enabled bool
}

// NewSetConnection creates a new setConnection message with the values provided.
func NewSetConnection(
	ctx context.Context,
	target *MessageTarget,
	guard int64,
	enabled bool,
	ch chan *sm.Response) *SetConnection {
	msg := &SetConnection{}

	msg.Initialize(ctx, TagSetConnection, ch)
	msg.Target = target
	msg.Guard = guard
	msg.Enabled = enabled

	return msg
}

// SendVia forwards the repair message to the rack's TOR for processing.  This
// is not the final destination for the message.
func (m *SetConnection) SendVia(ctx context.Context, r viaSender) error {
	return r.ViaTor(ctx, m.Target, m)
}

// String provides a formatted description of the message.
func (m *SetConnection) String() string {
	return fmt.Sprintf(
		"%s the network connection to %s",
		common.AOrB(m.Enabled, "Enable", "Disable"),
		m.Target.Describe())
}
