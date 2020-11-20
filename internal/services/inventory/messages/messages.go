package messages

import (
	"context"
	"errors"

	"github.com/Jim3Things/CloudChamber/internal/sm"
)

var ErrInvalidTarget = errors.New("invalid target specified, request ignored")

const (
	TagInvalid int = iota
	TagGetStatus
	TagSetConnection
	TagSetPower
	TagStartSim
	TagStopSim
	TagTimerExpiry
)

// +++ Base message and interfaces

type messageBase struct {
	EnvelopeState

	Target *MessageTarget

	Guard int64
}

type messageForwarder interface {
	SendVia(ctx context.Context, r viaSender) error
}

type viaSender interface {
	ViaTor(ctx context.Context, msg sm.Envelope) error
	ViaPDU(ctx context.Context, msg sm.Envelope) error
	ViaBlade(ctx context.Context, id int64, msg sm.Envelope) error
}

type messageAction interface {
	Do(ctx context.Context, sm *sm.SimpleSM, s RepairActionState)
}

type RepairMessage interface {
	messageForwarder
	messageAction
}

type messageStatus interface {
	GetStatus(ctx context.Context, sm *sm.SimpleSM, s RepairActionState)
}

type StatusMessage interface {
	messageForwarder
	messageStatus
}

// --- Base message and interfaces
