package inventory

import (
	"context"
	"time"

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
)

type testSuiteCore struct {
	suite.Suite

	utf *exporters.Exporter

	cfg *config.GlobalConfig

	timers *timestamp.Timers
}

func (ts *testSuiteCore) rackName() string { return "rack1" }

func (ts *testSuiteCore) SetupSuite() {
	ts.utf = exporters.NewExporter(exporters.NewUTForwarder())
	exporters.ConnectToProvider(ts.utf)

	cfg, err := utilities.StartSimSupportServices()
	ts.Require().NoError(err)

	ts.cfg = cfg
}

func (ts *testSuiteCore) SetupTest() {
	_ = ts.utf.Open(ts.T())

	err := timestamp.Reset(context.Background())
	ts.Require().Nilf(err, "Reset failed")

	ctx := context.Background()

	ts.Require().Nil(timestamp.SetPolicy(ctx, services.StepperPolicy_Manual, &duration.Duration{Seconds: 0}, -1))
	ts.timers = utilities.InitTimers()
}

func (ts *testSuiteCore) TearDownTest() {
	ts.utf.Close()
}

func (ts *testSuiteCore) createDummyRack(bladeCount int) *pb.ExternalRack {
	rackDef := &pb.ExternalRack{
		Pdu:    &pb.ExternalPdu{},
		Tor:    &pb.ExternalTor{},
		Blades: make(map[int64]*pbc.BladeCapacity),
	}

	for i := 0; i < bladeCount; i++ {
		rackDef.Blades[int64(i)] = &pbc.BladeCapacity{}
	}

	return rackDef
}

func (ts *testSuiteCore) createAndStartRack(
	ctx context.Context,
	bladeCount int,
	power bool,
	connect bool) (context.Context, *rack) {

	require := ts.Require()

	ctx = ts.advance(ctx)
	rackDef := ts.createDummyRack(bladeCount)

	ctx, span := tracing.StartSpan(
		ctx,
		tracing.WithName("starting rack (test infra)"),
		tracing.WithContextValue(timestamp.EnsureTickInContext))
	defer span.End()

	r := newRack(ctx, ts.rackName(), rackDef, ts.timers)
	require.NotNil(r)

	tick := common.TickFromContext(ctx)
	ctx = ts.advance(ctx)

	for i := range r.pdu.cables {
		r.pdu.cables[i] = newCable(power, false, tick)
	}

	for i := range r.tor.cables {
		r.tor.cables[i] = newCable(connect, false, tick)
	}

	require.NoError(r.start(ctx))

	return ctx, r
}

func (ts *testSuiteCore) bootBlade(ctx context.Context, r *rack, id int64) context.Context {
	require := ts.Require()

	rsp := make(chan *sm.Response)

	msg := newSetPower(ctx, newTargetBlade(r.name, id), common.TickFromContext(ctx), true, rsp)

	r.Receive(msg)

	res := ts.completeWithin(rsp, time.Duration(1)*time.Second)
	require.NotNil(res)
	require.NoError(res.Err)

	ctx, ok := ts.advanceToStateChange(ctx, 10, func() bool {
		return bladeWorkingState == r.blades[id].sm.CurrentIndex
	})

	require.True(ok)

	return ctx
}

func (ts *testSuiteCore) completeWithin(ch <-chan *sm.Response, delay time.Duration) *sm.Response {
	select {
	case res := <-ch:
		return res
	case <-time.After(delay):
		return nil
	}
}

func (ts *testSuiteCore) advance(ctx context.Context) context.Context {
	require := ts.Require()

	require.NoError(timestamp.Advance(ctx))
	return common.ContextWithTick(ctx, timestamp.Tick(ctx))
}

func (ts *testSuiteCore) advanceToStateChange(ctx context.Context, num int, compare func() bool) (context.Context, bool) {
	for i := 0; i < num; i++ {
		ctx = ts.advance(ctx)
		if compare() {
			return ctx, true
		}
	}

	return ctx, ts.waitForStateChange(compare)
}

func (ts *testSuiteCore) waitForStateChange(compare func() bool) bool {
	for i := 0; i < 10; i++ {
		time.Sleep(time.Duration(100) * time.Millisecond)
		if compare() {
			return true
		}
	}

	return false
}
