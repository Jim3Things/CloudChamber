package messages

import (
	"context"

	"github.com/Jim3Things/CloudChamber/internal/sm"
)

// GetStatus is the message used to obtain the target element's actual
// simulated operational state.
type GetStatus struct {
	messageBase
}

func (g GetStatus) SendVia(ctx context.Context, r viaSender) error {
	panic("implement me")
}

func (g GetStatus) GetStatus(ctx context.Context, sm *sm.SM, s RepairActionState) {
	panic("implement me")
}

