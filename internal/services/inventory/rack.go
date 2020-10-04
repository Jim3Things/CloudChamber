package inventory

import (
	"context"

	"github.com/Jim3Things/CloudChamber/internal/sm"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/inventory"
)

type rack struct {
	ch chan interface{}
	tor *tor
	pdu *pdu
	blades map[int64]*blade

	sm *sm.SimpleSM
}

const (
	rackWorkingState int = iota
	rackFailedState
)

func newRack(ctx context.Context, def *pb.ExternalRack) *rack {
	r := &rack{
		ch:     make(chan interface{}),
		tor:    newTor(def.Tor),
		pdu:    nil,
		blades: make(map[int64]*blade),
		sm:     nil,
	}

	r.sm = sm.NewSimpleSM(r,
		sm.WithFirstState(rackWorkingState, &rackWorking{}),
		sm.WithState(rackFailedState, &rackFailed{}),
	)

	r.pdu = newPdu(def.Pdu, r)

	for i, item := range def.Blades {
		r.blades[i] = newBlade(item)

		// These two calls are temporary fixups until the inventory definition
		// includes the tor and pdu connectors
		r.pdu.fixConnection(ctx, i)
		r.tor.fixConnection(i)
	}

	return r
}

func (r *rack) start(ctx context.Context) error {
	return r.sm.Start(ctx)
}

func (r *rack) findBlade(ctx context.Context, id int64) (*blade, bool) {
	b, ok := r.blades[id]
	return b, ok
}

type rackWorking struct {
	sm.NullState
}

func (s *rackWorking) Name() string { return "working" }

type rackFailed struct {
	sm.NullState
}

func (s *rackFailed) Name() string { return "failed" }
