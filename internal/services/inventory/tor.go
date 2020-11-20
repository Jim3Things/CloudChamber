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
	sm *sm.SimpleSM
}

const (
	// torWorkingState is the ID for when the TOR is fully operational.
	torWorkingState int = iota

	// torStuckState is the ID for when the TOR is faulted and unresponsive.
	// Note that programmed cables may or may not continue to be programmed.
	torStuckState
)

// newTor creates a new simulated TOR instance from the definition structure
// and the containing Rack.  Note that it currently does not fill in the cable
// information, as that is missing from the inventory definition.  That is
// done is the fixConnection function below.
func newTor(_ *pb.ExternalTor, r *Rack) *tor {
	t := &tor{
		cables: make(map[int64]*cable),
		holder: r,
		sm:     nil,
	}

	t.sm = sm.NewSimpleSM(t,
		sm.WithFirstState(
			torWorkingState,
			"working",
			sm.NullEnter,
			[]sm.ActionEntry{
				{messages.TagSetConnection, workingSetConnection, sm.Stay, sm.Stay},
			},
			UnexpectedMessage,
			sm.NullLeave),

		sm.WithState(
			torStuckState,
			"stuck",
			sm.NullEnter,
			[]sm.ActionEntry{},
			DropMessage,
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

// newStatusReport is a helper function to construct a status response for this
// TOR.
func (t *tor) newStatusReport(
	_ context.Context,
	_ *services.InventoryAddress) *sm.Response {
	return &sm.Response{
		Err: errors.New("not yet implemented"),
		Msg: nil,
	}
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

// workingSetConnection processes a setConnection request, updating the network
// connection state as required.
func workingSetConnection(ctx context.Context, machine *sm.SimpleSM, m sm.Envelope) bool {
	msg := m.(*messages.SetConnection)
	t := machine.Parent.(*tor)

	occursAt := common.TickFromContext(ctx)

	var id int64
	var c *cable
	var ok bool

	id, ok = msg.Target.BladeID()
	if ok {
		c, ok = t.cables[id]
	}

	if !ok {
		tracing.Warn(
			ctx,
			"No network connection for %s was found.",
			msg.Target.Describe())

		msg.GetCh() <- messages.InvalidTargetResponse(occursAt)
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

			msg.GetCh() <- messages.SuccessResponse(occursAt)
		} else {
			tracing.Info(
			ctx,
			"Network connection for %s has not changed.  It is currently %s.",
			msg.Target.Describe(),
			common.AOrB(c.on, "enabled", "disabled"))

			msg.GetCh() <- messages.FailedResponse(occursAt, ErrNoOperation)
		}
		break
	case ErrCableStuck:
		tracing.Warn(
			ctx,
			"Network connection for %s is stuck.  Unsure if it has been %s.",
			msg.Target.Describe(),
			common.AOrB(c.on, "enabled", "disabled"))

		msg.GetCh() <- messages.FailedResponse(occursAt, err)
		break

	case ErrTooLate:
		tracing.Info(
			ctx,
			"Network connection for %s has not changed, as this request arrived "+
				"after other changed occurred.  The blade's network connection "+
				"state remains unchanged.",
			msg.Target.Describe())

		msg.GetCh() <- messages.DroppedResponse(occursAt)
		break

	default:
		tracing.Warn(ctx, "Unexpected error code: %v", err)

		msg.GetCh() <- messages.FailedResponse(occursAt, err)
	}

	return true
}

// --- TOR state machine states
