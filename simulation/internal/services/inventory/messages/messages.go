package messages

import (
	"context"

	"github.com/Jim3Things/CloudChamber/simulation/internal/sm"
)

// Define the unique values used to identify the message type during processing.
const (
	// TagInvalid is a reserved, unused message type
	TagInvalid int = iota

	// TagGetStatus identifies a status request message
	TagGetStatus

	// TagSetConnection identifies a network connection change message
	TagSetConnection

	// TagSetPower identifies a power change message
	TagSetPower

	// TagTimerExpiry identifies a timer expiration message.
	TagTimerExpiry
)

// +++ Base message and interfaces

// messageBase is the standard header structure for an inventory repair or
// status message.
type messageBase struct {
	sm.EnvelopeState

	Target *MessageTarget

	Guard int64
}

// messageForwarder denotes the ability to handle targeting of the message to
// the required specific inventory element.
type messageForwarder interface {
	SendVia(ctx context.Context, r viaSender) error
}

// viaSender defines the functions required to send a message to any specific
// element in a rack, by type.
type viaSender interface {
	// ViaTor processes messages that are always routed through or to a TOR.
	// These are simulated 'over the wire' commands, such as status checks or
	// orders to start a workload.
	ViaTor(ctx context.Context, t *MessageTarget, msg sm.Envelope) error

	// ViaPDU processes messages that are always routed through or to a PDU.
	// These are simulated power operations.
	ViaPDU(ctx context.Context, t *MessageTarget, msg sm.Envelope) error

	// ToBlade processes messages that must be directly delivered to a blade,
	// with no intermediate routing.  These are messages that are internal to
	// the operation of hte simulation itself, such as a simulated timeout.
	ToBlade(ctx context.Context, id int64, msg sm.Envelope) error

	// ToTor processes messages that must be directly delivered to a TOR, for
	// reasons that are similar to those in ToBlade.
	ToTor(ctx context.Context, id int64, msg sm.Envelope) error

	// ToPdu processes messages that must be directly delivered to a PDU, for
	// reasons that are similar to those in ToBlade.
	ToPdu(ctx context.Context, id int64, msg sm.Envelope) error
}

// RepairMessage defines the required function signatures for all inventory
// repair messages.
type RepairMessage interface {
	messageForwarder
}

// StatusMessage defines the required function signatures for all inventory
// simulation status request messages.
type StatusMessage interface {
	messageForwarder
	GetStatus(ctx context.Context, sm *sm.SM, s RepairActionState)
}

// --- Base message and interfaces
