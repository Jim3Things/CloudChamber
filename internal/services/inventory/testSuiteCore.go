package inventory

import (
	"context"
	"time"

	"github.com/golang/protobuf/ptypes/duration"
	"github.com/stretchr/testify/suite"

	"github.com/Jim3Things/CloudChamber/internal/clients/timestamp"
	"github.com/Jim3Things/CloudChamber/internal/common"
	"github.com/Jim3Things/CloudChamber/internal/config"
	"github.com/Jim3Things/CloudChamber/internal/services/inventory/messages"
	"github.com/Jim3Things/CloudChamber/internal/sm"
	"github.com/Jim3Things/CloudChamber/internal/tracing"
	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/inventory"
	"github.com/Jim3Things/CloudChamber/pkg/protos/services"
	"github.com/Jim3Things/CloudChamber/test/setup"
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

	cfg, err := setup.StartSimSupportServices()
	ts.Require().NoError(err)

	ts.cfg = cfg
}

func (ts *testSuiteCore) SetupTest() {
	_ = ts.utf.Open(ts.T())

	err := timestamp.Reset(context.Background())
	ts.Require().Nilf(err, "Reset failed")

	ctx := context.Background()

	ts.Require().Nil(timestamp.SetPolicy(ctx, services.StepperPolicy_Manual, &duration.Duration{Seconds: 0}, -1))
	ts.timers = setup.InitTimers()
}

func (ts *testSuiteCore) TearDownTest() {
	ts.utf.Close()
}

func (ts *testSuiteCore) createDummyRack(bladeCount int) *pb.External_Rack {
	rackDef := &pb.External_Rack{
		Pdu:    &pb.External_Pdu{},
		Tor:    &pb.External_Tor{},
		Blades: make(map[int64]*pb.BladeCapacity),
	}

	for i := 0; i < bladeCount; i++ {
		rackDef.Blades[int64(i)] = &pb.BladeCapacity{}
	}

	return rackDef
}

func (ts *testSuiteCore) createAndStartRack(
	ctx context.Context,
	bladeCount int,
	power bool,
	connect bool) (context.Context, *Rack) {

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

func (ts *testSuiteCore) bootBlade(ctx context.Context, r *Rack, id int64) context.Context {
	require := ts.Require()

	rsp := make(chan *sm.Response)

	r.Receive(messages.NewSetPower(
		ctx,
		messages.NewTargetBlade(r.name, id),
		common.TickFromContext(ctx),
		true,
		rsp))

	res, ok := ts.completeWithin(rsp, time.Duration(1)*time.Second)
	require.True(ok)
	require.NotNil(res)
	require.NoError(res.Err)

	rsp = make(chan *sm.Response)

	r.Receive(messages.NewSetConnection(
		ctx,
		messages.NewTargetBlade(r.name, id),
		common.TickFromContext(ctx),
		true,
		rsp))

	res, ok = ts.completeWithin(rsp, time.Duration(1)*time.Second)
	require.True(ok)
	require.NotNil(res)
	require.NoError(res.Err)

	ctx, ok = ts.advanceToStateChange(ctx, 10, func() bool {
		return bladeWorking == r.blades[id].sm.CurrentIndex
	})

	require.True(ok)

	return ctx
}

func (ts *testSuiteCore) completeWithin(ch <-chan *sm.Response, delay time.Duration) (*sm.Response, bool) {
	select {
	case res, ok := <-ch:
		if !ok {
			return nil, true
		}

		return res, true

	case <-time.After(delay):
		return nil, false
	}
}

func (ts *testSuiteCore) advance(ctx context.Context) context.Context {
	require := ts.Require()

	require.NoError(timestamp.Advance(ctx))
	return common.ContextWithTick(ctx, timestamp.Tick(ctx))
}

func (ts *testSuiteCore) advanceToStateChange(
	ctx context.Context,
	num int,
	compare func() bool) (context.Context, bool) {
	for i := 0; i < num; i++ {
		ctx = ts.advance(ctx)
		if compare() {
			return ctx, true
		}
	}

	return ctx, utilities.WaitForStateChange(1, compare)
}
