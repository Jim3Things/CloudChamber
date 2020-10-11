package inventory

import (
	"context"

	"github.com/Jim3Things/CloudChamber/internal/common"
	"github.com/Jim3Things/CloudChamber/internal/sm"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/inventory"
)

type tor struct {
	cables map[int64]*cable
	holder *rack

	sm *sm.SimpleSM
}

func newTor(ext *pb.ExternalTor, r *rack) *tor {
	t := &tor {
		cables: make(map[int64]*cable),
		holder: r,
		sm: nil,
	}

	t.sm = sm.NewSimpleSM(t)

	return t
}

func (t *tor) fixConnection(ctx context.Context, id int64) {
	at := common.TickFromContext(ctx)

	t.sm.AdvanceGuard(at)

	t.cables[id] = newCable(false, false, at)
}

type torWorking struct {
	sm.NullState
}
