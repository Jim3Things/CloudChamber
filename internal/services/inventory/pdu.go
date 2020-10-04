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

type cableState struct {
	on bool
	at int64
}

// pdu defines the state required to simulate a PDU in a rack.
type pdu struct {
	cables map[int64]cableState
	holder *rack

	sm *sm.SimpleSM
}

func (p *pdu) Receive(ctx context.Context, msg interface{}, ch chan interface{}) {
	p.sm.Receive(ctx, msg, ch)
}

const (
	pduWorkingState int = iota
	pduOffState
	pduStuckState
)

func newPdu(_ *pb.ExternalPdu, r *rack) *pdu {
	p := &pdu{
		cables: make(map[int64]cableState),
		holder: r,
		sm:     nil,
	}

	p.sm = sm.NewSimpleSM(p,
		sm.WithFirstState(pduWorkingState, &pduWorking{}),
		sm.WithState(pduOffState, &pduOff{}),
		sm.WithState(pduStuckState, &pduStuck{}))

	return p
}

func (p *pdu) fixConnection(ctx context.Context, id int64) {
	at := common.TickFromContext(ctx)

	p.sm.At = common.MaxInt64(p.sm.At, at)

	p.cables[id] = cableState{
		on: false,
		at: at,
	}
}

func (p *pdu) forwardToBlade(ctx context.Context, id int64, msg interface{}, ch chan interface{}) {
	if b, ok := p.holder.blades[id]; ok {
		b.Receive(ctx, msg, ch)
	}
}

func (p *pdu) newStatusReport(ctx context.Context) *services.InventoryStatusResp {
	return nil
}

type pduWorking struct {
	sm.NullState
}

func (s *pduWorking) Receive(ctx context.Context, sm *sm.SimpleSM, msg interface{}, ch chan interface{}) {
	p := sm.Parent.(*pdu)

	switch msg := msg.(type) {
	case *services.InventoryRepairMsg:
		if power, ok := msg.GetAction().(*services.InventoryRepairMsg_Power); ok {
			s.changePower(ctx, sm, msg.Target, msg.After, power, ch)
			return
		}

		// Any other type of command, the pdu ignores.
		ch <- &services.InventoryRepairResp{
			Source: msg.Target,
			At:     &ct.Timestamp{Ticks: common.TickFromContext(ctx)},
			Rsp:    &services.InventoryRepairResp_Dropped{},
		}
		return

	case *services.InventoryStatusMsg:
		ch <- p.newStatusReport(ctx)
		return

	default:
		// Invalid message.  This should not happen, and we have no way to
		// send an error back.  Panic.
		tracing.Fatalf(ctx, "Invalid message received: %v", msg)
		return
	}
}

func (s *pduWorking) Name() string { return "working" }

func (s *pduWorking) changePower(
	ctx context.Context,
	sm *sm.SimpleSM,
	target *services.InventoryAddress,
	after *ct.Timestamp,
	power *services.InventoryRepairMsg_Power,
	ch chan interface{}) {
	p := sm.Parent.(*pdu)

	// There are four values that are relevant to how order and time
	// are managed here:
	//
	// - sm.At: this is the simulated time tick for the latest time any
	//          operation has executed against this PDU.  It is used as a
	//          pre-condition check for all PDU-wide operations.
	//
	// - cable.at: this is the simulated time tick for the latest time
	//             an operation executed against this cable.  It is never
	//             greater than sm.At.  It is used as a pre-condition for any
	//             operation that targets that cable.
	//
	// - after: this parameter specifies the guard test time for an operation.
	//          Any operation is invalid if the relevant test time above is
	//          greater than the after guard value.
	//
	// - occursAt: this is the simulated time tick when the operation executes.
	//             Structurally, it cannot be smaller than the after value.  It
	//             is used to update the sm.At and cable.at values, if the
	//             guard test succeeds.
	occursAt := common.TickFromContext(ctx)

	// Process the power command - change state if power command is for
	// the pdu, otherwise, forward along.
	switch elem := target.Element.(type) {

	case *services.InventoryAddress_Pdu:
		if sm.At < after.Ticks {

			// This command is newer than the last one that the PDU received
			// so it will be executed.  Record the updated last time of
			// operation.
			sm.At = occursAt

			// Change power at the PDU.  This only matters if the command is to
			// turn off the PDU (as this state means that the PDU is on).  And
			// turning off the PDU means turning off all the cables.
			if !power.Power {
				for i, cable := range p.cables {
					on := cable.on

					p.cables[i] = cableState{
						on: false,
						at: occursAt,
					}

					if on {
						// power is on to this blade.  Turn it off, but tell
						// the blade to not reply, as this is a side effect.
						p.forwardToBlade(ctx, i, power, nil)
					}
				}

				_ = sm.ChangeState(ctx, pduOffState)
			}
		}

		ch <- &services.InventoryRepairResp{
			Source: target,
			At:     &ct.Timestamp{Ticks: occursAt},
			Rsp:    &services.InventoryRepairResp_Dropped{},
		}

	case *services.InventoryAddress_BladeId:
		id := elem.BladeId

		if _, ok := p.cables[id]; ok {
			cable := p.cables[id]

			if cable.at < after.Ticks {
				// The state machine holds that sm.At is always greater than
				// or equal to any cable.at value.  But not all cable.at values
				// are the same.  So even though we're moving this cable.at
				// time forward, it still might be less than some other
				// cable.at time.  Hence the MaxInt64 call.
				sm.At = common.MaxInt64(sm.At, occursAt)

				on := cable.on

				p.cables[id] = cableState{
					on: power.Power,
					at: occursAt,
				}

				if on != power.Power {
					p.forwardToBlade(ctx, id, power, ch)
				}

				ch <- &services.InventoryRepairResp{
					Source: target,
					At:     &ct.Timestamp{Ticks: occursAt},
					Rsp:    &services.InventoryRepairResp_Success{},
				}

				return
			}
		}

		ch <- &services.InventoryRepairResp{
			Source: target,
			At:     &ct.Timestamp{Ticks: occursAt},
			Rsp:    &services.InventoryRepairResp_Dropped{},
		}

	default:
		ch <- &services.InventoryRepairResp{
			Source: target,
			At:     &ct.Timestamp{Ticks: occursAt},
			Rsp: &services.InventoryRepairResp_Failed{
				Failed: "invalid target specified, request ignored",
			},
		}
	}
}

type pduOff struct {
	sm.NullState
}

func (s *pduOff) Receive(ctx context.Context, sm *sm.SimpleSM, msg interface{}, ch chan interface{}) {
	p := sm.Parent.(*pdu)

	switch msg := msg.(type) {
	case *services.InventoryRepairMsg:
		ch <- &services.InventoryRepairResp{
			Source: msg.Target,
			At:     &ct.Timestamp{Ticks: common.TickFromContext(ctx)},
			Rsp:    &services.InventoryRepairResp_Dropped{},
		}
		return

	case *services.InventoryStatusMsg:
		ch <- p.newStatusReport(ctx)
		return

	default:
		return
	}
}

func (s *pduOff) Name() string { return "off" }

type pduStuck struct {
	sm.NullState
}

func (s *pduStuck) Receive(ctx context.Context, sm *sm.SimpleSM, msg interface{}, ch chan interface{}) {
	p := sm.Parent.(*pdu)

	switch msg.(type) {
	case *services.InventoryRepairMsg:
		return

	case *services.InventoryStatusMsg:
		ch <- p.newStatusReport(ctx)
		return

	default:
		return
	}
}

func (s *pduStuck) Name() string { return "stuck" }
