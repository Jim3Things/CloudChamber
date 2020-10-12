package inventory

import (
	"context"

	"github.com/Jim3Things/CloudChamber/internal/common"
	"github.com/Jim3Things/CloudChamber/internal/sm"
	"github.com/Jim3Things/CloudChamber/internal/tracing"
	ct "github.com/Jim3Things/CloudChamber/pkg/protos/common"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/inventory"
	"github.com/Jim3Things/CloudChamber/pkg/protos/services"
)

type tor struct {
	cables map[int64]*cable
	holder *rack

	sm *sm.SimpleSM
}

const (
	torWorkingState int = iota
	torStuckState
)

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

func (t *tor) fixConnection(ctx context.Context, id int64) {
	at := common.TickFromContext(ctx)

	t.sm.AdvanceGuard(at)

	t.cables[id] = newCable(false, false, at)
}

// newStatusReport is a helper function to construct a status response for this
// PDU.
func (t *tor) newStatusReport(
	ctx context.Context,
	target *services.InventoryAddress) *services.InventoryStatusResp {
	return nil
}

type torWorking struct {
	sm.NullState
}

func (s *torWorking) Receive(ctx context.Context, sm *sm.SimpleSM, msg interface{}, ch chan interface{}) {
	t := sm.Parent.(*tor)

	switch msg := msg.(type) {
	case *services.InventoryRepairMsg:
		if connect, ok := msg.GetAction().(*services.InventoryRepairMsg_Connect); ok {
			s.changeConnection(ctx, sm, msg.Target, msg.After, connect, ch)
			return
		}

		// Any other type of repair command, the tor ignores.
		ch <- droppedResponse(msg.Target, common.TickFromContext(ctx))
		return

	case *services.InventoryStatusMsg:
		ch <- t.newStatusReport(ctx, msg.Target)
		return

	default:
		// Invalid message.  This should not happen, and we have no way to
		// send an error back.  Panic.
		tracing.Fatal(ctx, "Invalid message received: %v", msg)
		return
	}
}

func (s *torWorking) changeConnection(
	ctx context.Context,
	sm *sm.SimpleSM,
	target *services.InventoryAddress,
	after *ct.Timestamp,
	connect *services.InventoryRepairMsg_Connect,
	ch chan interface{}) {
	t := sm.Parent.(*tor)

	occursAt := common.TickFromContext(ctx)


	switch elem := target.Element.(type) {
	case *services.InventoryAddress_BladeId:
		id := elem.BladeId

		if _, ok := t.cables[id]; ok {
			if changed, err := t.cables[id].set(connect.Connect, after.Ticks, occursAt); err == nil {
				sm.AdvanceGuard(occursAt)

				if changed {
					t.holder.forwardToBlade(ctx, id, connect, ch)
				}

				ch <- successResponse(target, occursAt)
			} else if err == errStuck {
				ch <- failedResponse(target, occursAt, err.Error())
			} else {
				ch <- droppedResponse(target, occursAt)
			}

			return
		}

	default:
		ch <- failedResponse(
			target, occursAt, "invalid target specified, request ignored")
	}
}

func (s *torWorking) Name() string { return "working" }

// torStuck is the state a TOR is in when it is unresponsive to commands, but
// is still powered on.  By implication, the connection state for each cable is
// also stuck.
type torStuck struct {
	sm.NullState
}

// Receive processes incoming requests for this state.
func (s *torStuck) Receive(ctx context.Context, sm *sm.SimpleSM, msg interface{}, ch chan interface{}) {
	t := sm.Parent.(*tor)

	switch msg := msg.(type) {
	case *services.InventoryRepairMsg:
		// the TOR is not responding to commands, so no repairs can be
		// processed.
		ch <- droppedResponse(msg.Target, common.TickFromContext(ctx))
		return

	case *services.InventoryStatusMsg:
		ch <- t.newStatusReport(ctx, msg.Target)
		return

	default:
		return
	}
}

// Name returns the friendly name for this state.
func (s *torStuck) Name() string { return "stuck" }
