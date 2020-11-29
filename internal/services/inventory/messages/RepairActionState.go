package messages

import (
	"context"

	"github.com/Jim3Things/CloudChamber/internal/sm"
)

// RepairActionState is the abstract definition that all inventory state
// machines must implement.
type RepairActionState interface {
	sm.SimpleSMState

	// Power is the function that responds to a setPower message.
	Power(ctx context.Context, sm *sm.SimpleSM, msg *SetPower)

	// Connect is the function that responds to a setConnection message.
	Connect(ctx context.Context, sm *sm.SimpleSM, msg *SetConnection)

	// Timeout is the function that processes a timer expiration message.
	Timeout(ctx context.Context, sm *sm.SimpleSM, msg *TimerExpiry)
}
