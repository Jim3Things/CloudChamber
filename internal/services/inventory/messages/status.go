package messages

import (
	"context"
	"fmt"

	"github.com/Jim3Things/CloudChamber/internal/sm"
)

// GetStatus defines a request for the current execution status of the target
// inventory element.
type GetStatus struct {
	messageBase
}

// NewGetStatus provides a newly created get status request for the target.
func NewGetStatus(
	ctx context.Context,
	target *MessageTarget,
	guard int64,
	ch chan *sm.Response) *GetStatus {
	msg := &GetStatus{}
	msg.Initialize(ctx, TagGetStatus, ch)
	msg.Target = target
	msg.Guard = guard

	return msg
}

// SendVia reflects that the get status request is a message from within the
// simulation, so either goes to the PDU directly (as a simplification of the
// simulated inventory structure), or always passes through the TOR.
func (m *GetStatus) SendVia(ctx context.Context, r viaSender) error {
	if m.Target.IsPdu() {
		return r.ViaPDU(ctx, m)
	}

	return r.ViaTor(ctx, m)
}

func (m *GetStatus) String() string {
	return fmt.Sprintf("Get the status for %s", m.Target.Describe())
}

// CableState contains the operational state for a single cable.
type CableState struct {
	On      bool
	Faulted bool
}

// StatusBody contains the fields common to all inventory elements - the
// state machine state, and the simulated time tick when that state was
// last entered.
type StatusBody struct {
	State     string
	EnteredAt int64
}

// PduStatus contains the operational state for a PDU
type PduStatus struct {
	StatusBody
	Cables map[int64]*CableState
}

// TorStatus contains the operational state for a TOR
type TorStatus struct {
	StatusBody
	Cables map[int64]*CableState
}

// BladeStatus contains the operational state for a blade, including a summary
// of the currently placed workloads.
type BladeStatus struct {
	StatusBody

	Capacity  *Capacity
	Used      *Capacity
	Workloads []string
}

// NewStatusResponse provides a newly created status message.
func NewStatusResponse(occursAt int64, body interface{}) *sm.Response {
	return &sm.Response{
		Err: nil,
		At:  occursAt,
		Msg: body,
	}
}
