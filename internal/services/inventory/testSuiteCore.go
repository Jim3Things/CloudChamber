package inventory

import (
	"context"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/Jim3Things/CloudChamber/internal/common"
	"github.com/Jim3Things/CloudChamber/internal/sm"
	"github.com/Jim3Things/CloudChamber/internal/tracing"
	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters"
	ct "github.com/Jim3Things/CloudChamber/pkg/protos/common"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/inventory"
)

type testSuiteCore struct {
	suite.Suite

	utf *exporters.Exporter
}

func (ts *testSuiteCore) rackName() string { return "rack1" }

func (ts *testSuiteCore) SetupSuite() {
	ts.utf = exporters.NewExporter(exporters.NewUTForwarder())
	exporters.ConnectToProvider(ts.utf)
}

func (ts *testSuiteCore) SetupTest() {
	_ = ts.utf.Open(ts.T())
}

func (ts *testSuiteCore) TearDownTest() {
	ts.utf.Close()
}

func (ts *testSuiteCore) createDummyRack(bladeCount int) *pb.ExternalRack {
	rackDef := &pb.ExternalRack{
		Pdu:    &pb.ExternalPdu{},
		Tor:    &pb.ExternalTor{},
		Blades: make(map[int64]*ct.BladeCapacity),
	}

	for i := 0; i < bladeCount; i++ {
		rackDef.Blades[int64(i)] = &ct.BladeCapacity{}
	}

	return rackDef
}

func (ts *testSuiteCore) completeWithin(ch <-chan *sm.Response, delay time.Duration) *sm.Response {
	select {
	case res := <-ch:
		return res
	case <-time.After(delay):
		return nil
	}
}

func (ts *testSuiteCore) execute(
	ctx context.Context,
	msg sm.Envelope,
	action func(ctx2 context.Context, envelope sm.Envelope)) {
	go func(tick int64, msg sm.Envelope) {
		c2, s := tracing.StartSpan(context.Background(),
			tracing.WithName("Executing simulated inventory operation"),
			tracing.WithNewRoot(),
			tracing.WithLink(msg.GetSpanContext(), msg.GetLinkID()))

		c2 = common.ContextWithTick(c2, tick)

		action(c2, msg)

		s.End()
	}(common.TickFromContext(ctx), msg)
}


