package inventory

import (
	"github.com/Jim3Things/CloudChamber/internal/sm"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/inventory"
)

type tor struct {
	cables map[int64]bool
	holder *rack

	sm *sm.SimpleSM
}

func newTor(t *pb.ExternalTor) *tor {
	return nil
}

func (t *tor) fixConnection(id int64) {

}
