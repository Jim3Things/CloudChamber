package inventory

import (
	"context"

	"github.com/Jim3Things/CloudChamber/internal/common"
	"github.com/Jim3Things/CloudChamber/internal/services/inventory/messages"
	"github.com/Jim3Things/CloudChamber/internal/sm"
	"github.com/Jim3Things/CloudChamber/internal/tracing"
	"github.com/Jim3Things/CloudChamber/pkg/errors"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/inventory"
)

// tor defines the state required to simulate a top-of-rack network
// switch.  The simulation is relatively shallow - it is a controller
// with cables that connect to a blade.  Because of this, it is similar to
// the simulation of a PDU, at this time.
type tor struct {
	// cables are the network connections to the Rack's blades.  They are
	// either programmed and working, or un-programmed (black-holed).  They
	// can also be in a faulted state.
	cables map[int64]*cable

	// Rack holds the pointer to the Rack that contains this TOR.
	holder *Rack

	// sm is the state machine for this TOR's simulation
	sm *sm.SM
}

const (
	// torWorkingState is the ID for when the TOR is fully operational.
	torWorkingState = "working"

	// torStuckState is the ID for when the TOR is faulted and unresponsive.
	// Note that programmed cables may or may not continue to be programmed.
	torStuckState = "stuck"
)

// newTor creates a new simulated TOR instance from the definition structure
// and the containing Rack.  Note that it currently does not fill in the cable
// information, as that is missing from the inventory definition.  That is
// done is the fixConnection function below.
func newTor(_ *pb.External_Tor, r *Rack) *tor {
	t := &tor{
		cables: make(map[int64]*cable),
		holder: r,
		sm:     nil,
	}

	t.sm = sm.NewSM(t,
		sm.WithFirstState(
			torWorkingState,
			sm.NullEnter,
			[]sm.ActionEntry{
				{messages.TagGetStatus, torGetStatus, sm.Stay, sm.Stay},
				{messages.TagSetConnection, workingSetConnection, sm.Stay, sm.Stay},
			},
			sm.UnexpectedMessage,
			sm.NullLeave),

		sm.WithState(
			torStuckState,
			sm.NullEnter,
			[]sm.ActionEntry{
				{messages.TagGetStatus, torOnlyGetStatus, sm.Stay, sm.Stay},
			},
			messages.DropMessage,
			sm.NullLeave),
	)

	return t
}

// fixConnection updates the TOR with presumed cable definitions to match up
// with the blades defined for the Rack.  This is a temporary workaround until
// the inventory definition structures include the cable definitions.
func (t *tor) fixConnection(ctx context.Context, id int64) {
	at := common.TickFromContext(ctx)

	t.sm.AdvanceGuard(at)

	t.cables[id] = newCable(false, false, at)
}

// Receive handles incoming messages for the TOR.
func (t *tor) Receive(ctx context.Context, msg sm.Envelope) {
	tracing.Info(ctx, "Processing message %q on TOR", msg)

	t.sm.Receive(ctx, msg)
}

// notifyBladeOfConnectionChange constructs a setConnection message that targets the
// specified blade, and forwards it along.
func (t *tor) notifyBladeOfConnectionChange(ctx context.Context, msg *messages.SetConnection, i int64) {
	fwd := messages.NewSetConnection(
		ctx,
		messages.NewTargetBlade(msg.Target.Rack, i),
		msg.Guard,
		msg.Enabled,
		nil)

	t.holder.forwardToBlade(ctx, i, fwd)

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
		return torOnlyGetStatus(ctx, machine, msg)
	} else if i, isBladeTarget := m.Target.BladeID(); isBladeTarget {
		return torToBladeGetStatus(ctx, t, i, msg)
	}

	processInvalidTarget(ctx, m, m.Target.Describe(), occursAt)
	return true
}

// torToBladeGetStatus processes a get status request that has targeted a
// blade.  It will forward it, if possible, or handle the error, if not.
func torToBladeGetStatus(ctx context.Context, t *tor, i int64, msg sm.Envelope) bool {
	m := msg.(*messages.GetStatus)

	occursAt := common.TickFromContext(ctx)

	c, ok := t.cables[i]

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

	if !t.holder.forwardToBlade(ctx, i, msg) {
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
			State:     t.sm.CurrentIndex,
			EnteredAt: t.sm.EnteredAt,
		},
		Cables: make(map[int64]*messages.CableState),
	}

	for i, c := range t.cables {
		torStatus.Cables[i] = &messages.CableState{
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

	var id int64
	var c *cable
	var ok bool

	id, ok = msg.Target.BladeID()
	if ok {
		c, ok = t.cables[id]
	}

	if c == nil || !ok {
		tracing.Warn(
			ctx,
			"No network connection for %s was found.",
			msg.Target.Describe())

		ch <- messages.InvalidTargetResponse(occursAt)
		return false
	}

	changed, err := t.cables[id].set(msg.Enabled, msg.Guard, occursAt)
	switch err {
	case nil:
		tracing.UpdateSpanName(
			ctx,
			"%s the network connection for %s",
			common.AOrB(msg.Enabled, "Enabling", "Disabling"),
			msg.Target.Describe())

		machine.AdvanceGuard(occursAt)

		if changed {
			t.notifyBladeOfConnectionChange(ctx, msg, id)

			ch <- sm.SuccessResponse(occursAt)
		} else {
			tracing.Info(
				ctx,
				"Network connection for %s has not changed.  It is currently %s.",
				msg.Target.Describe(),
				common.AOrB(c.on, "enabled", "disabled"))

			ch <- sm.FailedResponse(occursAt, errors.ErrNoOperation)
		}
		break

	case errors.ErrCableStuck:
		tracing.Warn(
			ctx,
			"Network connection for %s is stuck.  Unsure if it has been %s.",
			msg.Target.Describe(),
			common.AOrB(c.on, "enabled", "disabled"))

		ch <- sm.FailedResponse(occursAt, err)
		break

	case errors.ErrInventoryChangeTooLate(msg.Guard):
		tracing.Info(
			ctx,
			"Network connection for %s has not changed, as this request arrived "+
				"after other changed occurred.  The blade's network connection "+
				"state remains unchanged.",
			msg.Target.Describe())

		ch <- sm.FailedResponse(occursAt, err)
		break

	default:
		tracing.Warn(ctx, "Unexpected error code: %v", err)

		ch <- sm.FailedResponse(occursAt, err)
	}

	return true
}
