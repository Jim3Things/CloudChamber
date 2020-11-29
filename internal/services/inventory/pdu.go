package inventory

import (
	"context"
	"errors"

	"github.com/Jim3Things/CloudChamber/internal/common"
	"github.com/Jim3Things/CloudChamber/internal/services/inventory/messages"
	"github.com/Jim3Things/CloudChamber/internal/sm"
	"github.com/Jim3Things/CloudChamber/internal/tracing"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/inventory"
	"github.com/Jim3Things/CloudChamber/pkg/protos/services"
)

// pdu defines the state required to simulate a PDU in a Rack.
type pdu struct {
	// cables holds the simulated power cables.  The key is the blade id the
	// cable is attached to.
	cables map[int64]*cable

	// Rack holds the pointer to the Rack that contains this PDU.
	holder *Rack

	// sm is the state machine for this PDU simulation.
	sm *sm.SimpleSM
}

const (
	// pduWorkingState is the state ID for the PDU powered on and working
	// state.
	pduWorkingState = "working"

	// pduOffState is the state ID for the PDU powered off state.
	pduOffState = "off"

	// pduStuckState is the state ID for a PDU faulted state where the PDU is
	// unresponsive, but some power may still be on.
	pduStuckState = "stuck"
)

// newPdu creates a new pdu instance from the definition structure and the
// containing Rack.  Note that it currently does not fill in the cable
// information, as that is missing from the inventory definition.  That is
// done is the fixConnection function below.
func newPdu(_ *pb.ExternalPdu, r *Rack) *pdu {
	p := &pdu{
		cables: make(map[int64]*cable),
		holder: r,
		sm:     nil,
	}

	p.sm = sm.NewSimpleSM(p,
		sm.WithFirstState(
			pduWorkingState,
			sm.NullEnter,
			[]sm.ActionEntry{
				{messages.TagSetPower, workingSetPower, sm.Stay, pduOffState},
			},
			sm.UnexpectedMessage,
			sm.NullLeave),

		sm.WithState(
			pduOffState,
			sm.NullEnter,
			[]sm.ActionEntry{},
			messages.DropMessage,
			sm.NullLeave),

		sm.WithState(
			pduStuckState,
			sm.NullEnter,
			[]sm.ActionEntry{},
			messages.DropMessage,
			sm.NullLeave),
	)

	return p
}

// fixConnection updates the PDU with presumed cable definitions to match up
// with the blades defined for the Rack.  This is a temporary workaround until
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

// notifyBladeOfPowerChange constructs a setPower message that notifies the specified
// blade of the change in power, and sends it along.
func (p *pdu) notifyBladeOfPowerChange(ctx context.Context, msg *messages.SetPower, i int64) {
	fwd := messages.NewSetPower(
		ctx,
		messages.NewTargetBlade(msg.Target.Rack, i),
		msg.Guard,
		msg.On,
		nil)

	p.holder.forwardToBlade(ctx, i, fwd)

	tracing.Info(
		ctx,
		"Power connection to %s has changed.  It is now powered %s.",
		fwd.Target.Describe(),
		common.AOrB(fwd.On, "on", "off"))
}

// workingSetPower processes a set power message for a PDU in the normal
// operational state.  It handles power change messages for either a blade that
// the PDU supports, or for the PDU itself.
func workingSetPower(ctx context.Context, machine *sm.SimpleSM, m sm.Envelope) bool {
	msg := m.(*messages.SetPower)

	tracing.UpdateSpanName(ctx, msg.String())

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
	if msg.Target.IsPdu() {
		// Change the power on/off state for the full PDU
		tracing.UpdateSpanName(
			ctx,
			"Powering %s %s",
			common.AOrB(msg.On, "on", "off"),
			msg.Target.Describe())

		return setPowerForPdu(ctx, machine, msg, occursAt)
	}

	if id, isBladeTarget := msg.Target.BladeID(); isBladeTarget {
		// Change the power on/off state for an individual blade
		tracing.UpdateSpanName(
			ctx,
			"Powering %s %s",
			common.AOrB(msg.On, "on", "off"),
			msg.Target.Describe())

		setPowerForBlade(ctx, machine, msg, id, occursAt)
	} else {
		msg.GetCh() <- messages.InvalidTargetResponse(occursAt)
	}

	return true
}

// setPowerForPdu processes a set power message that targets this PDU.
func setPowerForPdu(
	ctx context.Context,
	machine *sm.SimpleSM,
	msg *messages.SetPower,
	occursAt int64) bool {
	p := machine.Parent.(*pdu)

	if machine.Pass(msg.Guard, occursAt) {
		// Change power at the PDU.  This only matters if the command is to
		// turn off the PDU (as this state means that the PDU is on).  And
		// turning off the PDU means turning off all the cables.
		if !msg.On {
			for i := range p.cables {
				changed, err := p.cables[i].force(false, msg.Guard, occursAt)

				if changed && err == nil {
					p.notifyBladeOfPowerChange(ctx, msg, i)
				}
			}

			msg.GetCh() <- messages.DroppedResponse(occursAt)
			return false
		}
	} else {
		tracing.Info(ctx, "Request ignored as it has arrived too late")
	}

	msg.GetCh() <- messages.DroppedResponse(occursAt)
	return true
}

// setPowerForBlade processes a set power message that targets a blade managed
// by this PDU.
func setPowerForBlade(
	ctx context.Context,
	machine *sm.SimpleSM,
	msg *messages.SetPower,
	id int64,
	occursAt int64) {
	p := machine.Parent.(*pdu)

	c, ok := p.cables[id]

	if !ok {
		tracing.Warn(ctx, "No power connection for blade %d was found.", id)

		msg.GetCh() <- messages.InvalidTargetResponse(occursAt)
		return
	}

	changed, err := p.cables[id].set(msg.On, msg.Guard, occursAt)

	switch err {
	case nil:
		// The state machine holds that machine.Guard is always greater than
		// or equal to any cable.at value.  But not all cable.at values
		// are the same.  So even though we're moving this cable.at
		// time forward, it still might be less than some other
		// cable.at time.
		machine.AdvanceGuard(occursAt)

		if changed {
			p.notifyBladeOfPowerChange(ctx, msg, id)
			msg.GetCh() <- messages.SuccessResponse(occursAt)
		} else {
			tracing.Info(
				ctx,
				"Power connection to %s has not changed.  It is currently powered %s.",
				msg.Target.Describe(),
				common.AOrB(c.on, "on", "off"))

			msg.GetCh() <- messages.FailedResponse(occursAt, ErrNoOperation)
		}
		break

	case ErrCableStuck:
		tracing.Warn(
			ctx,
			"Power connection to %s is stuck.  Unsure if it has been powered %s.",
			msg.Target.Describe(),
			common.AOrB(msg.On, "on", "off"))

		msg.GetCh() <- messages.FailedResponse(occursAt, err)
		break

	case ErrTooLate:
		tracing.Info(
			ctx,
			"Power connection to %s has not changed, as this request arrived "+
				"after other changed occurred.  The blade's power state remains unchanged.",
			msg.Target.Describe())

		msg.GetCh() <- messages.DroppedResponse(occursAt)
		break

	default:
		tracing.Warn(ctx, "Unexpected error code: %v", err)

		msg.GetCh() <- messages.FailedResponse(occursAt, err)
	}
}
