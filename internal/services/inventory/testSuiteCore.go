package inventory

import (
	"context"
	"time"

<<<<<<< HEAD
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/stretchr/testify/suite"

	"github.com/Jim3Things/CloudChamber/internal/clients/timestamp"
	"github.com/Jim3Things/CloudChamber/internal/common"
	"github.com/Jim3Things/CloudChamber/internal/config"
	"github.com/Jim3Things/CloudChamber/internal/sm"
	"github.com/Jim3Things/CloudChamber/internal/tracing"
	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters"
	pbc "github.com/Jim3Things/CloudChamber/pkg/protos/common"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/inventory"
	"github.com/Jim3Things/CloudChamber/pkg/protos/services"
	"github.com/Jim3Things/CloudChamber/test/utilities"
=======
	"github.com/stretchr/testify/suite"

	"github.com/Jim3Things/CloudChamber/internal/common"
	"github.com/Jim3Things/CloudChamber/internal/sm"
	"github.com/Jim3Things/CloudChamber/internal/tracing"
	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters"
	ct "github.com/Jim3Things/CloudChamber/pkg/protos/common"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/inventory"
>>>>>>> master
)

type testSuiteCore struct {
	suite.Suite

	utf *exporters.Exporter
<<<<<<< HEAD

	cfg *config.GlobalConfig
=======
>>>>>>> master
}

func (ts *testSuiteCore) rackName() string { return "rack1" }

func (ts *testSuiteCore) SetupSuite() {
	ts.utf = exporters.NewExporter(exporters.NewUTForwarder())
	exporters.ConnectToProvider(ts.utf)
<<<<<<< HEAD

	cfg, err := utilities.StartSimSupportServices()
	ts.Require().NoError(err)

	ts.cfg = cfg
=======
>>>>>>> master
}

func (ts *testSuiteCore) SetupTest() {
	_ = ts.utf.Open(ts.T())
<<<<<<< HEAD

	err := timestamp.Reset(context.Background())
	ts.Require().Nilf(err, "Reset failed")

	ctx := context.Background()

	ts.Require().Nil(timestamp.SetPolicy(ctx, services.StepperPolicy_Manual, &duration.Duration{Seconds: 0}, -1))
=======
>>>>>>> master
}

func (ts *testSuiteCore) TearDownTest() {
	ts.utf.Close()
}

func (ts *testSuiteCore) createDummyRack(bladeCount int) *pb.ExternalRack {
	rackDef := &pb.ExternalRack{
		Pdu:    &pb.ExternalPdu{},
		Tor:    &pb.ExternalTor{},
<<<<<<< HEAD
		Blades: make(map[int64]*pbc.BladeCapacity),
	}

	for i := 0; i < bladeCount; i++ {
		rackDef.Blades[int64(i)] = &pbc.BladeCapacity{}
=======
		Blades: make(map[int64]*ct.BladeCapacity),
	}

	for i := 0; i < bladeCount; i++ {
		rackDef.Blades[int64(i)] = &ct.BladeCapacity{}
>>>>>>> master
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

<<<<<<< HEAD
func (ts *testSuiteCore) execute(msg sm.Envelope, action func(ctx2 context.Context, envelope sm.Envelope)) {
	go func(msg sm.Envelope) {
		ctx, span := tracing.StartSpan(context.Background(),
			tracing.WithName("Executing simulated inventory operation"),
			tracing.WithNewRoot(),
			tracing.WithLink(msg.GetSpanContext(), msg.GetLinkID()),
			tracing.WithContextValue(timestamp.EnsureTickInContext))

		action(ctx, msg)

		span.End()
	}(msg)
}

func (ts *testSuiteCore) powerOnBlade(ctx context.Context, r *rack, id int64) {
	require := ts.Require()

	ctx, span := tracing.StartSpan(
		ctx,
		tracing.WithName("power on blade (test infra)"),
		tracing.WithContextValue(timestamp.EnsureTickInContext))

	rsp := make(chan *sm.Response)

	msg := newSetPower(
		ctx,
		newTargetBlade(ts.rackName(), id),
		common.TickFromContext(ctx),
		true,
		rsp)

	span.End()

	ts.execute(msg, r.pdu.Receive)

	res := ts.completeWithin(rsp, time.Duration(1)*time.Second)
	require.NotNil(res)
	require.NoError(res.Err)
}

func (ts *testSuiteCore) advance(ctx context.Context) context.Context {
	require := ts.Require()

	require.NoError(timestamp.Advance(ctx))
	return common.ContextWithTick(ctx, timestamp.Tick(ctx))
}
=======
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


>>>>>>> master
