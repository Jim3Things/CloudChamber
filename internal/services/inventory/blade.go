package inventory

import (
	"context"

	"github.com/Jim3Things/CloudChamber/internal/sm"
	"github.com/Jim3Things/CloudChamber/pkg/protos/common"
)

type blade struct {

}

func newBlade(b *common.BladeCapacity) *blade {
	return nil
}

func (b *blade) Receive(ctx context.Context, msg *sm.Envelope) {

}
