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

// tor defines the state required to simulate a top-of-rack network
// switch.  The simulation is relatively shallow - it is a controller
// with cables that connect to a blade.  Because of this, it is similar to
// the simulation of a PDU, at this time.
type tor struct {
	// cables are the network connections to the Rack's blades.  They are
	// either programmed and working, or un-programmed (black-holed).  They
	// can also be in a faulted state.
	cables map[string]*cable

	// Rack holds the pointer to the Rack that contains this TOR.
	holder *Rack

	// id is the index used to identify this TOR within the Rack.
	id int64

	// propDelay is the propagation delay in network connections, stored as a
	// range of ticks that the network operation is delayed from taking effect.
	propDelay common.Range

	// sm is the state machine for this TOR's simulation
	sm *sm.SM
}

// newTor creates a new simulated TOR instance from the definition structure
// and the containing Rack.  Note that it currently does not fill in the cable
// information, as that is missing from the inventory definition.  That is
// done is the fixConnection function below.
func newTor(
	ctx context.Context,
	def *pb.Definition_Tor,
	propDelay common.Range,
	name string,
	r *Rack,
	id int64) *tor {
	t := &tor{
		cables:    make(map[string]*cable),
		holder:    r,
		id:        id,
		propDelay: propDelay,
		sm:        nil,
	}

	t.sm = sm.NewSM(t,
		name,
		sm.WithFirstState(
			pb.TorState_working,
			sm.NullEnter,
			[]sm.ActionEntry{
				{messages.TagGetStatus, torGetStatus, sm.Stay, sm.Stay},
				{messages.TagSetConnection, workingSetConnection, sm.Stay, sm.Stay},
			},
			sm.UnexpectedMessage,
			sm.NullLeave),

		sm.WithState(
			pb.TorState_stuck,
			sm.NullEnter,
			[]sm.ActionEntry{
				{messages.TagGetStatus, torOnlyGetStatus, sm.Stay, sm.Stay},
			},
			messages.DropMessage,
			sm.NullLeave),
	)

	// Wire up all the cables, and also register this instance with the rack for
	// routing to the cable's destination element.
	at := common.TickFromContext(ctx)
	t.sm.AdvanceGuard(at)

	for _, port := range def.GetPorts() {
		target := messages.HardwareToTarget(port.Item)
		key := target.Key()
		t.cables[key] = newCable(target, false, false, at)
		r.AddToTorMap(key, id)
	}

	// Finally, add our self address into the map to ensure that messages that
	// target this TOR directly are routed here.
	r.AddToTorMap(messages.NewTargetTor(r.sm.Name, id, 0).Key(), id)

	tracing.AddImpact(ctx, tracing.ImpactCreate, name)

	return t
}

// Save returns a protobuf message that contains the data sufficient to persist
// and later restore this state machine to a logically equivalent state.
func (t *tor) Save() (proto.Message, error) {
	cur, entered, terminal, guard := t.sm.Savable()

	state := &pb.Actual_Tor{
		Condition: pb.Actual_operational,
		Cables:    make(map[int64]*pb.Actual_Cable),
		SmState:   cur.(pb.TorState_SM),
		Core: &pb.Actual_MachineCore{
			EnteredAt: entered,
			Terminal:  terminal,
			Guard:     guard,
		},
	}

	i := int64(0)
	for _, c := range t.cables {
		state.Cables[i] = c.save()
		i++
	}

	return state, nil
}

// Receive handles incoming messages for the TOR.
func (t *tor) Receive(ctx context.Context, msg sm.Envelope) {
	tracing.Info(ctx, "Processing message %q on TOR", msg)

	t.sm.Receive(ctx, msg)
}

// notifyBladeOfConnectionChange constructs a setConnection message that targets the
// specified blade, and forwards it along.
func (t *tor) notifyBladeOfConnectionChange(ctx context.Context, msg *messages.SetConnection) {
	fwd := messages.NewSetConnection(
		ctx,
		messages.NewTargetBlade(msg.Target.Rack, msg.Target.ElementId(), msg.Target.Port()),
		msg.Guard,
		msg.Enabled,
		nil)

	t.holder.forwardToBlade(ctx, t.propDelay.Pick(), msg.Target, fwd)

	tracing.Info(
		ctx,
		"Network connection to %s has changed.  It is now powered %s.",
		fwd.Target.Describe(),
		common.AOrB(fwd.Enabled, "enabled", "disabled"))
}

// torGetStatus returns the current simulated status for the TOR, or passes
// through to the blade, if that is the target.
func torGetStatus(ctx context.Context, machine *sm.SM, msg sm.Envelope) bool {
	t := machine.Parent.(*tor)
	m := msg.(*messages.GetStatus)

	occursAt := common.TickFromContext(ctx)

	if m.Target.IsTor() {
		tracing.AddImpact(ctx, tracing.ImpactRead, machine.Name)
		return torOnlyGetStatus(ctx, machine, msg)
	} else if m.Target.IsBlade() {
		tracing.AddImpact(ctx, tracing.ImpactUse, machine.Name)
		return torToBladeGetStatus(ctx, t, msg)
	}

	processInvalidTarget(ctx, m, m.Target.Describe(), occursAt)
	return true
}

// torToBladeGetStatus processes a get status request that has targeted a
// blade.  It will forward it, if possible, or handle the error, if not.
func torToBladeGetStatus(ctx context.Context, t *tor, msg sm.Envelope) bool {
	m := msg.(*messages.GetStatus)
	key := m.Target.Key()

	occursAt := common.TickFromContext(ctx)

	c, ok := t.cables[key]

	if !ok {
		processInvalidTarget(ctx, msg, m.Target.Describe(), occursAt)
		return false
	}

	if !c.on || c.faulted {
		ch := msg.Ch()
		if ch != nil {
			close(ch)
		}

		return false
	}

	if !t.holder.forwardToBlade(ctx, 0, m.Target, msg) {
		ch := msg.Ch()
		if ch != nil {
			close(ch)
		}

		return false
	}

	return true
}

// torOnlyGetStatus returns the current simulated status for the TOR.
func torOnlyGetStatus(ctx context.Context, machine *sm.SM, msg sm.Envelope) bool {
	t := machine.Parent.(*tor)
	m := msg.(*messages.GetStatus)

	ch := msg.Ch()
	defer close(ch)

	tracing.UpdateSpanName(
		ctx,
		"Getting the current status for %s",
		m.Target.Describe())

	torStatus := &messages.TorStatus{
		StatusBody: messages.StatusBody{
			State:     t.sm.CurrentIndex.String(),
			EnteredAt: t.sm.EnteredAt,
		},
		Cables: make(map[string]*messages.CableState),
	}

	for key, c := range t.cables {
		torStatus.Cables[key] = &messages.CableState{
			On:      c.on,
			Faulted: c.faulted,
		}
	}

	ch <- messages.NewStatusResponse(
		common.TickFromContext(ctx),
		torStatus)

	return true
}

// workingSetConnection processes a setConnection request, updating the network
// connection state as required.
func workingSetConnection(ctx context.Context, machine *sm.SM, m sm.Envelope) bool {
	msg := m.(*messages.SetConnection)

	ch := msg.Ch()
	defer close(ch)

	t := machine.Parent.(*tor)

	occursAt := common.TickFromContext(ctx)

	var c *cable

	key := msg.Target.Key()
	ok := msg.Target.IsBlade()
	if ok {
		c, ok = t.cables[key]
	}

	if c == nil || !ok {
		tracing.Warn(
			ctx,
			"No network connection for %s was found.",
			msg.Target.Describe())

		ch <- messages.InvalidTargetResponse(occursAt)
		return false
	}

	changed, err := c.set(msg.Enabled, msg.Guard, occursAt)
	switch err {
	case nil:
		tracing.AddImpact(ctx, tracing.ImpactModify, machine.Name)
		tracing.UpdateSpanName(
			ctx,
			"%s the network connection for %s",
			common.AOrB(msg.Enabled, "Enabling", "Disabling"),
			msg.Target.Describe())

		machine.AdvanceGuard(occursAt)

		if changed {
			t.notifyBladeOfConnectionChange(ctx, msg)

			ch <- sm.SuccessResponse(occursAt)
		} else {
			tracing.AddImpact(ctx, tracing.ImpactRead, machine.Name)
			tracing.Info(
				ctx,
				"Network connection for %s has not changed.  It is currently %s.",
				msg.Target.Describe(),
				common.AOrB(c.on, "enabled", "disabled"))

			ch <- sm.FailedResponse(occursAt, errors.ErrNoOperation)
		}

	case errors.ErrCableStuck:
		tracing.AddImpact(ctx, tracing.ImpactRead, machine.Name)
		tracing.Warn(
			ctx,
			"Network connection for %s is stuck.  Unsure if it has been %s.",
			msg.Target.Describe(),
			common.AOrB(c.on, "enabled", "disabled"))

		ch <- sm.FailedResponse(occursAt, err)

	case errors.ErrInventoryChangeTooLate(msg.Guard):
		tracing.Info(
			ctx,
			"Network connection for %s has not changed, as this request arrived "+
				"after other changed occurred.  The blade's network connection "+
				"state remains unchanged.",
			msg.Target.Describe())

		ch <- sm.FailedResponse(occursAt, err)

	default:
		tracing.Warn(ctx, "Unexpected error code: %v", err)

		ch <- sm.FailedResponse(occursAt, err)
	}

	return true
}
