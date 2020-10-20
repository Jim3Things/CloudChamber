package inventory

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel/api/trace"

	"github.com/Jim3Things/CloudChamber/internal/common"
	"github.com/Jim3Things/CloudChamber/internal/sm"
	ct "github.com/Jim3Things/CloudChamber/pkg/protos/common"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/inventory"
	"github.com/Jim3Things/CloudChamber/pkg/protos/services"
)

// pdu defines the state required to simulate a PDU in a rack.
type pdu struct {
	// cables holds the simulated power cables.  The key is the blade id the
	// cable is attached to.
	cables map[int64]*cable

	// rack holds the pointer to the rack that contains this PDU.
	holder *rack

	// sm is the state machine for this PDU's simulation.
	sm *sm.SimpleSM
}

const (
	// pduWorkingState is the state ID for the PDU powered on and working
	// state.
	pduWorkingState int = iota

	// pduOffState is the state ID for the PDU powered off state.
	pduOffState

	// pduStuckState is the state ID for a PDU faulted state where the PDU is
	// unresponsive, but some power may still be on.
	pduStuckState
)

// newPdu creates a new pdu instance from the definition structure and the
// containing rack.  Note that it currently does not fill in the cable
// information, as that is missing from the inventory definition.  That is
// done is the fixConnection function below.
func newPdu(_ *pb.ExternalPdu, r *rack) *pdu {
	p := &pdu{
		cables: make(map[int64]*cable),
		holder: r,
		sm:     nil,
	}

	p.sm = sm.NewSimpleSM(p,
		sm.WithFirstState(pduWorkingState, &pduWorking{}),
		sm.WithState(pduOffState, &pduOff{}),
		sm.WithState(pduStuckState, &pduStuck{}))

	return p
}

// fixConnection updates the PDU with presumed cable definitions to match up
// with the blades defined for the rack.  This is a temporary workaround until
// the inventory definition structures include the cable definitions.
func (p *pdu) fixConnection(ctx context.Context, id int64) {
	at := common.TickFromContext(ctx)

	p.sm.AdvanceGuard(at)

	p.cables[id] = newCable(false, false, at)
}

// Receive handles incoming messages for the PDU.
func (p *pdu) Receive(ctx context.Context, msg *sm.Envelope) {
	p.sm.Receive(ctx, msg)
}

// newStatusReport is a helper function to construct a status response for this
// PDU.
func (p *pdu) newStatusReport(
	ctx context.Context,
	target *services.InventoryAddress) *sm.Response {
	return &sm.Response{
		Err: errors.New("not yet implemented"),
		Msg: nil,
	}
}

// pduWorking is the state a PDU is in when it is turned on and functional.
type pduWorking struct {
	sm.NullState
}

// Receive processes incoming requests for this state.
func (s *pduWorking) Receive(ctx context.Context, sm *sm.SimpleSM, msg *sm.Envelope) {
	p := sm.Parent.(*pdu)

	switch body := msg.Msg.(type) {
	case *services.InventoryRepairMsg:
		if power, ok := body.GetAction().(*services.InventoryRepairMsg_Power); ok {
			s.changePower(ctx, sm, body.Target, body.After, power, msg.Ch)
			return
		}

		// Any other type of repair command, the pdu ignores.
		msg.Ch <- droppedResponse(common.TickFromContext(ctx))

	case *services.InventoryStatusMsg:
		msg.Ch <- p.newStatusReport(ctx, body.Target)

	default:
		// Invalid message.
		msg.Ch <- unexpectedMessageResponse(s, common.TickFromContext(ctx), body)
	}
}

// Name returns the friendly name for this state.
func (s *pduWorking) Name() string { return "working" }

// changePower implements the repair operation to turn either a cable or the
// full PDU on or off.
func (s *pduWorking) changePower(
	ctx context.Context,
	machine *sm.SimpleSM,
	target *services.InventoryAddress,
	after *ct.Timestamp,
	power *services.InventoryRepairMsg_Power,
	ch chan *sm.Response) {
	p := machine.Parent.(*pdu)

	// There are four values that are relevant to how order and time
	// are managed here:
	//
	// - machine.Guard: this is the simulated time tick for the latest time any
	//          operation has executed against this PDU.  It is used as a
	//          pre-condition check for all PDU-wide operations.
	//
	// - cable.at: this is the simulated time tick for the latest time
	//             an operation executed against this cable.  It is never
	//             greater than machine.Guard.  It is used as a pre-condition for any
	//             operation that targets that cable.
	//
	// - after: this parameter specifies the guard test time for an operation.
	//          Any operation is invalid if the relevant test time above is
	//          greater than the after guard value.
	//
	// - occursAt: this is the simulated time tick when the operation executes.
	//             Structurally, it cannot be smaller than the after value.  It
	//             is used to update the machine.Guard and cable.at values, if the
	//             guard test succeeds.
	occursAt := common.TickFromContext(ctx)

	// Process the power command - change state if power command is for
	// the pdu, otherwise, forward along.
	switch elem := target.Element.(type) {

	// Change the power on/off state for the full PDU
	case *services.InventoryAddress_Pdu:
		if machine.Pass(after.Ticks, occursAt) {
			// Change power at the PDU.  This only matters if the command is to
			// turn off the PDU (as this state means that the PDU is on).  And
			// turning off the PDU means turning off all the cables.
			if !power.Power {
				sc := trace.SpanFromContext(ctx).SpanContext()

				for i := range p.cables {

					changed, err := p.cables[i].force(false, after.Ticks, occursAt)

					if changed && err == nil {
						// power is on to this blade.  Turn it off, but tell
						// the blade to not reply, as the blade action is a
						// side effect of the PDU change.
						fwd := &sm.Envelope{
							Ch:   nil,
							Span: sc,
							Msg:  &services.InventoryRepairMsg{
								Target: target,
								After:  after,
								Action: power,
							},
						}

						p.holder.forwardToBlade(ctx, i, fwd)
					}
				}

				_ = machine.ChangeState(ctx, pduOffState)
			}
		}

		ch <- droppedResponse(occursAt)

	// Change the power on/off state for an individual blade
	case *services.InventoryAddress_BladeId:
		id := elem.BladeId

		if _, ok := p.cables[id]; ok {
			if changed, err := p.cables[id].set(power.Power, after.Ticks, occursAt); err == nil {
				// The state machine holds that machine.Guard is always greater than
				// or equal to any cable.at value.  But not all cable.at values
				// are the same.  So even though we're moving this cable.at
				// time forward, it still might be less than some other
				// cable.at time.
				machine.AdvanceGuard(occursAt)

				if changed {
					fwd := &sm.Envelope{
						Ch:   ch,
						Span: trace.SpanFromContext(ctx).SpanContext(),
						Msg:  &services.InventoryRepairMsg{
							Target: target,
							After:  after,
							Action: power,
						},
					}

					p.holder.forwardToBlade(ctx, id, fwd)
				}

				ch <- successResponse(occursAt)
			} else if err == ErrCableStuck {
				ch <- failedResponse(occursAt, err)
			} else {
				ch <- droppedResponse(occursAt)
			}

			return
		}

		ch <- droppedResponse(occursAt)

	default:
		ch <- failedResponse(occursAt, ErrInvalidTarget)
	}
}

// pduOff is the state a PDU is in when it is fully powered off.
type pduOff struct {
	sm.NullState
}

// Receive processes incoming requests for this state.
func (s *pduOff) Receive(ctx context.Context, sm *sm.SimpleSM, msg *sm.Envelope) {
	p := sm.Parent.(*pdu)

	switch body := msg.Msg.(type) {
	case *services.InventoryRepairMsg:
		// Powered off, so no repairs can be processed.
		msg.Ch <- droppedResponse(common.TickFromContext(ctx))

	case *services.InventoryStatusMsg:
		msg.Ch <- p.newStatusReport(ctx, body.Target)

	default:
		msg.Ch <- unexpectedMessageResponse(s, common.TickFromContext(ctx), body)
	}
}

// Name returns the friendly name for this state.
func (s *pduOff) Name() string { return "off" }

// pduStuck is the state a PDU is in when it is unresponsive to commands, but
// is still powered on.  By implication, the powered state for each cable is
// also stuck.
type pduStuck struct {
	sm.NullState
}

// Receive processes incoming requests for this state.
func (s *pduStuck) Receive(ctx context.Context, sm *sm.SimpleSM, msg *sm.Envelope) {
	p := sm.Parent.(*pdu)

	switch body := msg.Msg.(type) {
	case *services.InventoryRepairMsg:
		// the PDU is not responding to commands, so no repairs can be
		// processed.
		msg.Ch <- droppedResponse(common.TickFromContext(ctx))

	case *services.InventoryStatusMsg:
		msg.Ch <- p.newStatusReport(ctx, body.Target)

	default:
		// Invalid message.
		msg.Ch <- unexpectedMessageResponse(s, common.TickFromContext(ctx), body)
	}
}

// Name returns the friendly name for this state.
func (s *pduStuck) Name() string { return "stuck" }
