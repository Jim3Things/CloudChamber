package inventory

import (
	"context"
	"errors"

	"github.com/Jim3Things/CloudChamber/internal/common"
	"github.com/Jim3Things/CloudChamber/internal/sm"
	"github.com/Jim3Things/CloudChamber/internal/tracing"
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

	// sm is the state machine for this PDU simulation.
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
func (p *pdu) Receive(ctx context.Context, msg sm.Envelope) {
	tracing.Info(ctx, "Processing message %q on PDU", msg)

	p.sm.Receive(ctx, msg)
}

// newStatusReport is a helper function to construct a status response for this
// PDU.
func (p *pdu) newStatusReport(
	_ context.Context,
	_ *services.InventoryAddress) *sm.Response {
	return &sm.Response{
		Err: errors.New("not yet implemented"),
		Msg: nil,
	}
}

// sendPowerToBlade constructs a setPower message that targets the specified
// blade, and forwards it along.
func (p *pdu) sendPowerToBlade(ctx context.Context, msg *setPower, i int64, rsp chan *sm.Response) {
	fwd := newSetPower(
		ctx,
		newTargetBlade(msg.target.rack, i),
		msg.guard,
		msg.on,
		rsp)

	p.holder.forwardToBlade(ctx, i, fwd)

	tracing.Info(
		ctx,
		"Power connection to %s has changed.  It is now powered %s.",
		fwd.target.describe(),
		aOrB(fwd.on, "on", "off"))
}

// pduWorking is the state a PDU is in when it is turned on and functional.
type pduWorking struct {
	nullRepairAction
}

// Receive processes incoming requests for this state.
func (s *pduWorking) Receive(ctx context.Context, machine *sm.SimpleSM, msg sm.Envelope) {
	s.handleMsg(ctx, machine, s, msg)
}

func (s *pduWorking) power(ctx context.Context, machine *sm.SimpleSM, msg *setPower) {
	tracing.UpdateSpanName(ctx, msg.String())

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
	// - msg.guard: this parameter specifies the guard test time for an operation.
	//              Any operation is invalid if the relevant test time above is
	//              greater than the guard value.
	//
	// - occursAt: this is the simulated time tick when the operation executes.
	//             Structurally, it cannot be smaller than the guard value.  It
	//             is used to update the machine.Guard and cable.at values, if the
	//             guard test succeeds.
	occursAt := common.TickFromContext(ctx)

	// Process the power command - change state if power command is for
	// the pdu, otherwise, forward along.
	if msg.target.isPdu() {
		// Change the power on/off state for the full PDU
		tracing.UpdateSpanName(
			ctx,
			"Powering %s %s",
			aOrB(msg.on, "on", "off"),
			msg.target.describe())

		if machine.Pass(msg.guard, occursAt) {
			// Change power at the PDU.  This only matters if the command is to
			// turn off the PDU (as this state means that the PDU is on).  And
			// turning off the PDU means turning off all the cables.
			if !msg.on {
				for i := range p.cables {
					changed, err := p.cables[i].force(false, msg.guard, occursAt)

					if changed && err == nil {
						p.sendPowerToBlade(ctx, msg, i, nil)
					}
				}

				_ = machine.ChangeState(ctx, pduOffState)
			}
		} else {
			tracing.Info(ctx, "Request ignored as it has arrived too late")
		}

		msg.GetCh() <- droppedResponse(occursAt)
	} else if id, isBladeTarget := msg.target.bladeID(); isBladeTarget {
		// Change the power on/off state for an individual blade
		tracing.UpdateSpanName(
			ctx,
			"Powering %s %s",
			aOrB(msg.on, "on", "off"),
			msg.target.describe())

		if c, ok := p.cables[id]; ok {
			if changed, err := p.cables[id].set(msg.on, msg.guard, occursAt); err == nil {
				// The state machine holds that machine.Guard is always greater than
				// or equal to any cable.at value.  But not all cable.at values
				// are the same.  So even though we're moving this cable.at
				// time forward, it still might be less than some other
				// cable.at time.
				machine.AdvanceGuard(occursAt)

				if changed {
					p.sendPowerToBlade(ctx, msg, id, msg.GetCh())
				} else {
					tracing.Info(
						ctx,
						"Power connection to %s has not changed.  It is currently powered %s.",
						msg.target.describe(),
						aOrB(c.on, "on", "off"))

					msg.GetCh() <- failedResponse(occursAt, ErrNoOperation)
				}
			} else if err == ErrCableStuck {
				tracing.Warn(
					ctx,
					"Power connection to %s is stuck.  Unsure if it has been powered %s.",
					msg.target.describe(),
					aOrB(msg.on, "on", "off"))

				msg.GetCh() <- failedResponse(occursAt, err)
			} else if err == ErrTooLate {
				tracing.Info(
					ctx,
					"Power connection to %s has not changed, as this request arrived "+
						"after other changed occurred.  The blade's power state remains unchanged.",
					msg.target.describe())

				msg.GetCh() <- droppedResponse(occursAt)
			} else {
				tracing.Warn(ctx, "Unexpected error code: %v", err)

				msg.GetCh() <- failedResponse(occursAt, err)
			}

			return
		}

		tracing.Warn(
			ctx,
			"No power connection for blade %d was found.",
			id)

		msg.GetCh() <- failedResponse(occursAt, ErrInvalidTarget)

	} else {
		msg.GetCh() <- failedResponse(occursAt, ErrInvalidTarget)
	}
}

// Name returns the friendly name for this state.
func (s *pduWorking) Name() string { return "working" }

// pduOff is the state a PDU is in when it is fully powered off.
type pduOff struct {
	dropRepairAction
}

// Receive processes incoming requests for this state.
func (s *pduOff) Receive(ctx context.Context, machine *sm.SimpleSM, msg sm.Envelope) {
	s.handleMsg(ctx, machine, s, msg)
}

// Name returns the friendly name for this state.
func (s *pduOff) Name() string { return "off" }

// pduStuck is the state a PDU is in when it is unresponsive to commands, but
// is still powered on.  By implication, the powered state for each cable is
// also stuck.
type pduStuck struct {
	dropRepairAction
}

// Receive processes incoming requests for this state.
func (s *pduStuck) Receive(ctx context.Context, machine *sm.SimpleSM, msg sm.Envelope) {
	s.handleMsg(ctx, machine, s, msg)
}

// Name returns the friendly name for this state.
func (s *pduStuck) Name() string { return "stuck" }
