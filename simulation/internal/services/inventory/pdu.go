package inventory

import (
	"context"

	"github.com/golang/protobuf/proto"

	"github.com/Jim3Things/CloudChamber/simulation/internal/common"
	"github.com/Jim3Things/CloudChamber/simulation/internal/services/inventory/messages"
	"github.com/Jim3Things/CloudChamber/simulation/internal/sm"
	"github.com/Jim3Things/CloudChamber/simulation/internal/tracing"
	"github.com/Jim3Things/CloudChamber/simulation/pkg/errors"
	pb "github.com/Jim3Things/CloudChamber/simulation/pkg/protos/inventory"
)

// pdu defines the state required to simulate a PDU in a Rack.
type pdu struct {
	// cables holds the simulated power cables.  The key is the key-formatted
	// target for the element the cable is attached to.
	cables map[string]*cable

	// Rack holds the pointer to the Rack that contains this PDU.
	holder *Rack

	// id is the index used to identify this PDU within the Rack.
	id int64

	// sm is the state machine for this PDU simulation.
	sm *sm.SM
}

// newPdu creates a new pdu instance from the definition structure and the
// containing Rack.  Note that it currently does not fill in the cable
// information, as that is missing from the inventory definition.  That is
// done is the fixConnection function below.
func newPdu(ctx context.Context, def *pb.Definition_Pdu, name string, r *Rack, id int64) *pdu {
	p := &pdu{
		cables: make(map[string]*cable),
		holder: r,
		id:     id,
		sm:     nil,
	}

	p.sm = sm.NewSM(p,
		name,
		sm.WithFirstState(
			pb.PduState_working,
			sm.NullEnter,
			[]sm.ActionEntry{
				{messages.TagGetStatus, pduGetStatus, sm.Stay, sm.Stay},
				{messages.TagSetPower, workingSetPower, sm.Stay, pb.PduState_off},
			},
			sm.UnexpectedMessage,
			sm.NullLeave),

		sm.WithState(
			pb.PduState_off,
			sm.NullEnter,
			[]sm.ActionEntry{
				{messages.TagGetStatus, pduGetStatus, sm.Stay, sm.Stay},
			},
			messages.DropMessage,
			sm.NullLeave),

		sm.WithState(
			pb.PduState_stuck,
			sm.NullEnter,
			[]sm.ActionEntry{
				{messages.TagGetStatus, pduGetStatus, sm.Stay, sm.Stay},
			},
			messages.DropMessage,
			sm.NullLeave),
	)

	// Wire up all the cables, and also register this instance with the rack for
	// routing to the cable's destination element.
	at := common.TickFromContext(ctx)
	p.sm.AdvanceGuard(at)

	for _, port := range def.GetPorts() {
		target := messages.HardwareToTarget(port.Item)
		key := target.Key()
		p.cables[key] = newCable(target, false, false, at)
		r.AddToPduMap(key, id)
	}

	// Finally, add our self address into the map to ensure that messages that
	// target this PDU directly are routed here.
	r.AddToPduMap(messages.NewTargetPdu(r.sm.Name, id, 0).Key(), id)

	tracing.AddImpact(ctx, tracing.ImpactCreate, name)

	return p
}

// Save returns a protobuf message that contains the data sufficient to persist
// and later restore this state machine to a logically equivalent state.
func (p *pdu) Save() (proto.Message, error) {
	cur, entered, terminal, guard := p.sm.Savable()

	state := &pb.Actual_Pdu{
		Condition: pb.Actual_operational,
		Cables:    make(map[int64]*pb.Actual_Cable),
		SmState:   cur.(pb.PduState_SM),
		Core: &pb.Actual_MachineCore{
			EnteredAt: entered,
			Terminal:  terminal,
			Guard:     guard,
		},
	}

	i := int64(0)
	for _, c := range p.cables {
		state.Cables[i] = c.save()
		i++
	}

	return state, nil
}

// Receive handles incoming messages for the PDU.
func (p *pdu) Receive(ctx context.Context, msg sm.Envelope) {
	tracing.Info(ctx, "Processing message %q on PDU", msg)

	p.sm.Receive(ctx, msg)
}

// notifyBladeOfPowerChange constructs a setPower message that notifies the specified
// blade of the change in power, and sends it along.
func (p *pdu) notifyBladeOfPowerChange(ctx context.Context, msg *messages.SetPower, i int64) {
	fwd := messages.NewSetPower(
		ctx,
		messages.NewTargetBlade(msg.Target.Rack, i, 0),
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

// pduGetStatus returns the current simulated status for the PDU.
func pduGetStatus(ctx context.Context, machine *sm.SM, m sm.Envelope) bool {
	p := machine.Parent.(*pdu)

	ch := m.Ch()
	defer close(ch)

	tracing.AddImpact(ctx, tracing.ImpactRead, machine.Name)

	pduStatus := &messages.PduStatus{
		StatusBody: messages.StatusBody{
			State:     p.sm.CurrentIndex.String(),
			EnteredAt: p.sm.EnteredAt,
		},
		Cables: make(map[string]*messages.CableState),
	}

	for key, c := range p.cables {
		pduStatus.Cables[key] = &messages.CableState{
			On:      c.on,
			Faulted: c.faulted,
		}
	}

	ch <- messages.NewStatusResponse(
		common.TickFromContext(ctx),
		pduStatus)

	return true
}

// workingSetPower processes a set power message for a PDU in the normal
// operational state.  It handles power change messages for either a blade that
// the PDU supports, or for the PDU itself.
func workingSetPower(ctx context.Context, machine *sm.SM, m sm.Envelope) bool {
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

		tracing.AddImpact(ctx, tracing.ImpactModify, machine.Name)
		return setPowerForPdu(ctx, machine, msg, occursAt)
	}

	if msg.Target.IsBlade() {
		// Change the power on/off state for an individual blade
		tracing.UpdateSpanName(
			ctx,
			"Powering %s %s",
			common.AOrB(msg.On, "on", "off"),
			msg.Target.Describe())

		tracing.AddImpact(ctx, tracing.ImpactUse, machine.Name)
		setPowerForBlade(ctx, machine, msg, msg.Target, occursAt)
	} else {
		processInvalidTarget(ctx, msg, msg.Target.Describe(), occursAt)
	}

	return true
}

// setPowerForPdu processes a set power message that targets this PDU.
func setPowerForPdu(
	ctx context.Context,
	machine *sm.SM,
	msg *messages.SetPower,
	occursAt int64) bool {
	p := machine.Parent.(*pdu)

	ch := msg.Ch()
	defer close(ch)

	if machine.Pass(msg.Guard, occursAt) {
		// Change power at the PDU.  This only matters if the command is to
		// turn off the PDU (as this state means that the PDU is on).  And
		// turning off the PDU means turning off all the cables.
		if !msg.On {
			for i, c := range p.cables {
				changed, err := p.cables[i].force(false, msg.Guard, occursAt)

				if changed && err == nil {
					p.notifyBladeOfPowerChange(ctx, msg, c.target.ElementId())
				}
			}

			return false
		}
	} else {
		tracing.Info(ctx, "Request ignored as it has arrived too late")
	}

	return true
}

// setPowerForBlade processes a set power message that targets a blade managed
// by this PDU.
func setPowerForBlade(
	ctx context.Context,
	machine *sm.SM,
	msg *messages.SetPower,
	target *messages.MessageTarget,
	occursAt int64) {
	p := machine.Parent.(*pdu)

	key := target.Key()
	c, ok := p.cables[key]

	if !ok {
		processInvalidTarget(ctx, msg, msg.Target.Describe(), occursAt)
		return
	}

	ch := msg.Ch()
	defer close(ch)

	changed, err := c.set(msg.On, msg.Guard, occursAt)

	switch err {
	case nil:
		// The state machine holds that machine.Guard is always greater than
		// or equal to any cable.at value.  But not all cable.at values
		// are the same.  So even though we're moving this cable.at
		// time forward, it still might be less than some other
		// cable.at time.
		machine.AdvanceGuard(occursAt)

		if changed {
			p.notifyBladeOfPowerChange(ctx, msg, target.ElementId())
			ch <- sm.SuccessResponse(occursAt)
		} else {
			tracing.Info(
				ctx,
				"Power connection to %s has not changed.  It is currently powered %s.",
				msg.Target.Describe(),
				common.AOrB(c.on, "on", "off"))

			ch <- sm.FailedResponse(occursAt, errors.ErrNoOperation)
		}

	case errors.ErrCableStuck:
		tracing.Warn(
			ctx,
			"Power connection to %s is stuck.  Unsure if it has been powered %s.",
			msg.Target.Describe(),
			common.AOrB(msg.On, "on", "off"))

		ch <- sm.FailedResponse(occursAt, err)

	case errors.ErrInventoryChangeTooLate(msg.Guard):
		tracing.Info(
			ctx,
			"Power connection to %s has not changed, as this request arrived "+
				"after other changed occurred.  The blade's power state remains unchanged.",
			msg.Target.Describe())

		ch <- sm.FailedResponse(occursAt, err)

	default:
		tracing.Warn(ctx, "Unexpected error code: %v", err)

		ch <- sm.FailedResponse(occursAt, err)
	}
}
