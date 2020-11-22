package messages

import (
	"context"
	"errors"

	"github.com/Jim3Things/CloudChamber/internal/sm"
)

// ErrInvalidTarget is an error used to indicate that the incoming message had
// a target element that either was not valid for the message, or an element
// that could not be found.
var ErrInvalidTarget = errors.New("invalid target specified, request ignored")

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

	// TagStartSim identifies the message used to start a rack-level
	// simulation
	TagStartSim

	// TagStopSim identifies the message used to terminate a rack-level
	// simulation
	TagStopSim

	// TagTimerExpiry identifies a timer expiration message.
	TagTimerExpiry
)

// +++ Base message and interfaces

// messageBase is the standard header structure for an inventory repair or
// status message.
type messageBase struct {
	EnvelopeState

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
	ViaTor(ctx context.Context, msg sm.Envelope) error
	ViaPDU(ctx context.Context, msg sm.Envelope) error
	ViaBlade(ctx context.Context, id int64, msg sm.Envelope) error
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
	GetStatus(ctx context.Context, sm *sm.SimpleSM, s RepairActionState)
}

// --- Base message and interfaces
