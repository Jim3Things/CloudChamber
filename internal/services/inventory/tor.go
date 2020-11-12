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

// tor defines the state required to simulate a top-of-rack network
// switch.  The simulation is relatively shallow - it is a controller
// with cables that connect to a blade.  Because of this, it is similar to
// the simulation of a PDU, at this time.
type tor struct {
	// cables are the network connections to the rack's blades.  They are
	// either programmed and working, or un-programmed (black-holed).  They
	// can also be in a faulted state.
	cables map[int64]*cable

	// rack holds the pointer to the rack that contains this TOR.
	holder *rack

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
// and the containing rack.  Note that it currently does not fill in the cable
// information, as that is missing from the inventory definition.  That is
// done is the fixConnection function below.
func newTor(_ *pb.ExternalTor, r *rack) *tor {
	t := &tor{
		cables: make(map[int64]*cable),
		holder: r,
		sm:     nil,
	}

	t.sm = sm.NewSimpleSM(t,
		sm.WithFirstState(torWorkingState, &torWorking{}),
		sm.WithState(torStuckState, &torStuck{}))

	return t
}

// fixConnection updates the TOR with presumed cable definitions to match up
// with the blades defined for the rack.  This is a temporary workaround until
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

// sendConnectionToBlade constructs a setConnection message that targets the
// specified blade, and forwards it along.
func (t *tor) sendConnectionToBlade(ctx context.Context, msg *setConnection, i int64) {
	fwd := newSetConnection(
		ctx,
		newTargetBlade(msg.target.rack, i),
		msg.guard,
		msg.enabled,
		nil)

	t.holder.forwardToBlade(ctx, i, fwd)

	tracing.Info(
		ctx,
		"Network connection to %s has changed.  It is now powered %s.",
		fwd.target.describe(),
		aOrB(fwd.enabled, "enabled", "disabled"))
}

// +++ TOR state machine states

// torWorking is the state a TOR is in when it is functioning correctly.
type torWorking struct {
	nullRepairAction
}

// Receive processes incoming requests for this state.
func (s *torWorking) Receive(ctx context.Context, machine *sm.SimpleSM, msg sm.Envelope) {
	s.handleMsg(ctx, machine, s, msg)
}

// connect processes a setConnection request, updating the network connection
// state as required.
func (s *torWorking) connect(ctx context.Context, machine *sm.SimpleSM, msg *setConnection) {
	t := machine.Parent.(*tor)

	occursAt := common.TickFromContext(ctx)

	if id, isBladeTarget := msg.target.bladeID(); isBladeTarget {
		if c, ok := t.cables[id]; ok {
			if changed, err := t.cables[id].set(msg.enabled, msg.guard, occursAt); err == nil {
				tracing.UpdateSpanName(
					ctx,
					"%s the network connection for %s",
					aOrB(msg.enabled, "Enabling", "Disabling"),
					msg.target.describe())

				machine.AdvanceGuard(occursAt)

				if changed {
					t.sendConnectionToBlade(ctx, msg, id)

					msg.GetCh() <- successResponse(occursAt)
				} else {
					tracing.Info(
						ctx,
						"Network connection for %s has not changed.  It is currently %s.",
						msg.target.describe(),
						aOrB(c.on, "enabled", "disabled"))

					msg.GetCh() <- failedResponse(occursAt, ErrNoOperation)
				}
			} else if err == ErrCableStuck {
				tracing.Warn(
					ctx,
					"Network connection for %s is stuck.  Unsure if it has been %s.",
					msg.target.describe(),
					aOrB(c.on, "enabled", "disabled"))

				msg.GetCh() <- failedResponse(occursAt, err)
			} else if err == ErrTooLate {
				tracing.Info(
					ctx,
					"Network connection for %s has not changed, as this request arrived "+
						"after other changed occurred.  The blade's network connection "+
						"state remains unchanged.",
					msg.target.describe())

				msg.GetCh() <- droppedResponse(occursAt)
			} else {
				tracing.Warn(ctx, "Unexpected error code: %v", err)

				msg.GetCh() <- failedResponse(occursAt, err)
			}

			return
		}
	} else {
		tracing.Warn(
			ctx,
			"No network connection for %s was found.",
			msg.target.describe())

		msg.GetCh() <- failedResponse(occursAt, ErrInvalidTarget)
	}
}

// Name returns the friendly name for this state.
func (s *torWorking) Name() string { return "working" }

// torStuck is the state a TOR is in when it is unresponsive to commands, but
// is still powered on.  By implication, the connection state for each cable is
// also stuck.
type torStuck struct {
	dropRepairAction
}

// Receive processes incoming requests for this state.
func (s *torStuck) Receive(ctx context.Context, machine *sm.SimpleSM, msg sm.Envelope) {
	s.handleMsg(ctx, machine, s, msg)
}

// Name returns the friendly name for this state.
func (s *torStuck) Name() string { return "stuck" }

// --- TOR state machine states
