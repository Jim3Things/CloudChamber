package inventory

import (
	"context"
	"errors"

	"github.com/Jim3Things/CloudChamber/internal/common"
	"github.com/Jim3Things/CloudChamber/internal/sm"
	"github.com/Jim3Things/CloudChamber/internal/tracing"
	ct "github.com/Jim3Things/CloudChamber/pkg/protos/common"
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

	// sm is the state machine fro this TOR's simulation
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
	t := &tor {
		cables: make(map[int64]*cable),
		holder: r,
		sm: nil,
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
func (t *tor) Receive(ctx context.Context, msg *sm.Envelope) {
	t.sm.Receive(ctx, msg)
}

// newStatusReport is a helper function to construct a status response for this
// TOR.
func (t *tor) newStatusReport(
	ctx context.Context,
	target *services.InventoryAddress) *sm.Response {
	return &sm.Response{
		Err: errors.New("not yet implemented"),
		Msg: nil,
	}
}

// torWorking is the state a TOR is in when it is functioning correctly.
type torWorking struct {
	sm.NullState
}

// Receive processes incoming requests for this state.
func (s *torWorking) Receive(ctx context.Context, sm *sm.SimpleSM, msg *sm.Envelope) {
	t := sm.Parent.(*tor)

	switch body := msg.Msg.(type) {
	case *services.InventoryRepairMsg:
		if connect, ok := body.GetAction().(*services.InventoryRepairMsg_Connect); ok {
			s.changeConnection(ctx, sm, body.Target, body.After, connect, msg.Ch)
			return
		}

		// Any other type of repair command, the tor ignores.
		msg.Ch <- droppedResponse(common.TickFromContext(ctx))

	case *services.InventoryStatusMsg:
		msg.Ch <- t.newStatusReport(ctx, body.Target)

	default:
		// Invalid message.
		msg.Ch <- unexpectedMessageResponse(s, common.TickFromContext(ctx), body)
	}
}

// Name returns the friendly name for this state.
func (s *torWorking) Name() string { return "working" }

// changeConnection implements the repair operation to program or deprogram a
// network cable in the TOR.
func (s *torWorking) changeConnection(
	ctx context.Context,
	machine *sm.SimpleSM,
	target *services.InventoryAddress,
	after *ct.Timestamp,
	connect *services.InventoryRepairMsg_Connect,
	ch chan *sm.Response) {
	t := machine.Parent.(*tor)

	occursAt := common.TickFromContext(ctx)

	switch elem := target.Element.(type) {
	case *services.InventoryAddress_BladeId:
		id := elem.BladeId

		if _, ok := t.cables[id]; ok {
			if changed, err := t.cables[id].set(connect.Connect, after.Ticks, occursAt); err == nil {
				tracing.UpdateSpanName(
					ctx,
					"%s the network connection for %s",
					aOrB(connect.Connect, "Enabling", "Disabling"),
					target.Describe())

				machine.AdvanceGuard(occursAt)

				if changed {
					fwd := sm.NewEnvelope(
						ctx,
						&services.InventoryRepairMsg{
							Target: target,
							After:  after,
							Action: connect,
						},
						ch)

					t.holder.forwardToBlade(ctx, id, fwd)
				}

				ch <- successResponse(occursAt)
			} else if err == ErrCableStuck {
				ch <- failedResponse(occursAt, err)
			} else {
				ch <- droppedResponse(occursAt)
			}

			return
		}

	default:
		ch <- failedResponse(occursAt, ErrInvalidTarget)
	}
}

// torStuck is the state a TOR is in when it is unresponsive to commands, but
// is still powered on.  By implication, the connection state for each cable is
// also stuck.
type torStuck struct {
	sm.NullState
}

// Receive processes incoming requests for this state.
func (s *torStuck) Receive(ctx context.Context, sm *sm.SimpleSM, msg *sm.Envelope) {
	t := sm.Parent.(*tor)

	switch body := msg.Msg.(type) {
	case *services.InventoryRepairMsg:
		// the TOR is not responding to commands, so no repairs can be
		// processed.
		msg.Ch <- droppedResponse(common.TickFromContext(ctx))

	case *services.InventoryStatusMsg:
		msg.Ch <- t.newStatusReport(ctx, body.Target)

	default:
		// Invalid message.
		msg.Ch <- unexpectedMessageResponse(s, common.TickFromContext(ctx), body)
	}
}

// Name returns the friendly name for this state.
func (s *torStuck) Name() string { return "stuck" }
