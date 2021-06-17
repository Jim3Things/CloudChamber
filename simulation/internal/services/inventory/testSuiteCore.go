package inventory

import (
	"context"
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes/duration"
	"github.com/stretchr/testify/suite"

	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/timestamp"
	"github.com/Jim3Things/CloudChamber/simulation/internal/common"
	"github.com/Jim3Things/CloudChamber/simulation/internal/config"
	"github.com/Jim3Things/CloudChamber/simulation/internal/services/inventory/messages"
	"github.com/Jim3Things/CloudChamber/simulation/internal/sm"
	"github.com/Jim3Things/CloudChamber/simulation/internal/tracing"
	"github.com/Jim3Things/CloudChamber/simulation/internal/tracing/exporters"
	pb "github.com/Jim3Things/CloudChamber/simulation/pkg/protos/inventory"
	"github.com/Jim3Things/CloudChamber/simulation/pkg/protos/services"
	"github.com/Jim3Things/CloudChamber/simulation/test/setup"
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
	require := ts.Require()

	_ = ts.utf.Open(ts.T())

	ctx := context.Background()

	require.NoError(timestamp.Reset(ctx))
	_, err := timestamp.SetPolicy(ctx, services.StepperPolicy_Manual, &duration.Duration{Seconds: 0}, -1)
	require.NoError(err)

	ts.timers = setup.InitTimers()
}

func (ts *testSuiteCore) TearDownTest() {
	ts.utf.Close()
}

func (ts *testSuiteCore) createDummyRack(bladeCount int) *pb.Definition_Rack {
	rackDef := &pb.Definition_Rack{
		Pdus:   make(map[int64]*pb.Definition_Pdu),
		Tors:   make(map[int64]*pb.Definition_Tor),
		Blades: make(map[int64]*pb.Definition_Blade),
	}

	rackDef.Pdus[0] = &pb.Definition_Pdu{
		Details: &pb.PduDetails{},
		Ports:   make(map[int64]*pb.PowerPort),
	}

	rackDef.Tors[0] = &pb.Definition_Tor{
		Details: &pb.TorDetails{},
		Ports:   make(map[int64]*pb.NetworkPort),
	}

	for i := 0; i < bladeCount; i++ {
		hw := &pb.Hardware{
			Type: pb.Hardware_blade,
			Id:   int64(i),
			Port: 0,
		}

		rackDef.Blades[int64(i)] = &pb.Definition_Blade{
			Details:            &pb.BladeDetails{},
			Capacity:           &pb.BladeCapacity{},
			BootInfo:           &pb.BladeBootInfo{},
			BootOnPowerOn: true,
		}

		rackDef.Pdus[0].Ports[int64(i)] = &pb.PowerPort{
			Wired: false,
			Item:  hw,
		}

		rackDef.Tors[0].Ports[int64(i)] = &pb.NetworkPort{
			Wired: false,
			Item:  hw,
		}
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

	r := newRack(
		ctx,
		ts.rackName(),
		rackDef,
		fmt.Sprintf("racks/%s/pdus/", ts.rackName()),
		fmt.Sprintf("racks/%s/tors/", ts.rackName()),
		fmt.Sprintf("racks/%s/blades/", ts.rackName()),
		ts.timers)
	require.NotNil(r)

	tick := common.TickFromContext(ctx)
	ctx = ts.advance(ctx)

	for _, p := range r.pdus {
		for _, c := range p.cables {
			_, _ = c.set(power, tick, tick)
		}
	}

	for _, t := range r.tors {
		for _, c := range t.cables {
			_, _ = c.set(connect, tick, tick)
		}
	}

	require.NoError(r.start(ctx))

	return ctx, r
}

func (ts *testSuiteCore) bootBlade(ctx context.Context, r *Rack, target *messages.MessageTarget) context.Context {
	require := ts.Require()

	rsp := make(chan *sm.Response)

	r.Receive(messages.NewSetPower(
		ctx,
		target,
		common.TickFromContext(ctx),
		true,
		rsp))

	res, ok := ts.completeWithin(rsp, time.Second)
	require.True(ok)
	require.NotNil(res)
	require.NoError(res.Err)

	rsp = make(chan *sm.Response)

	r.Receive(messages.NewSetConnection(
		ctx,
		target,
		common.TickFromContext(ctx),
		true,
		rsp))

	res, ok = ts.completeWithin(rsp, time.Second)
	require.True(ok)
	require.NotNil(res)
	require.NoError(res.Err)

	return ts.advanceToStateChange(ctx, 10, func() bool {
		return pb.BladeSmState_working == r.blades[target.ElementId()].sm.CurrentIndex
	})
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
	compare func() bool) context.Context {

	require := ts.Require()

	for i := 0; i < num; i++ {
		ctx = ts.advance(ctx)
		if compare() {
			return ctx
		}
	}

	require.Eventually(compare, time.Second, 10*time.Millisecond)
	return ctx
}
