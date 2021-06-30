package inventory

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/stretchr/testify/suite"

	tsc "github.com/Jim3Things/CloudChamber/simulation/internal/clients/timestamp"
	"github.com/Jim3Things/CloudChamber/simulation/internal/common"
	"github.com/Jim3Things/CloudChamber/simulation/internal/services/inventory/messages"
	"github.com/Jim3Things/CloudChamber/simulation/internal/sm"
	"github.com/Jim3Things/CloudChamber/simulation/internal/tracing"
	"github.com/Jim3Things/CloudChamber/simulation/pkg/errors"
	pb "github.com/Jim3Things/CloudChamber/simulation/pkg/protos/inventory"
)

type BladeTestSuite struct {
	testSuiteCore
}

func (ts *BladeTestSuite) issueSetPower(ctx context.Context, r *Rack, id int64, on bool) *sm.Response {
	require := ts.Require()

	rsp := make(chan *sm.Response)

	msg := messages.NewSetPower(
		ctx,
		messages.NewTargetBlade(ts.rackName(), id, 0),
		common.TickFromContext(ctx),
		on,
		rsp)

	r.Receive(msg)

	res, ok := ts.completeWithin(rsp, time.Second)
	require.True(ok)

	return res
}

func (ts *BladeTestSuite) issueSetConnection(ctx context.Context, r *Rack, id int64, on bool) *sm.Response {
	require := ts.Require()

	rsp := make(chan *sm.Response)

	msg := messages.NewSetConnection(
		ctx,
		messages.NewTargetBlade(ts.rackName(), id, 0),
		common.TickFromContext(ctx),
		on,
		rsp)

	r.Receive(msg)

	res, ok := ts.completeWithin(rsp, time.Second)
	require.True(ok)

	return res
}

func (ts *BladeTestSuite) issueGetStatus(ctx context.Context, r *Rack, id int64) *sm.Response {
	require := ts.Require()

	rsp := make(chan *sm.Response)
	msg := messages.NewGetStatus(ctx,
		messages.NewTargetBlade(ts.rackName(), id, 0),
		common.TickFromContext(ctx),
		rsp)

	r.Receive(msg)

	res, ok := ts.completeWithin(rsp, time.Second)
	require.True(ok)

	return res
}

func (ts *BladeTestSuite) TestGetStatus() {
	require := ts.Require()

	rackDef := ts.createDummyRack(2)

	r := newRack(
		context.Background(),
		ts.rackName(),
		rackDef,
		ts.cfg,
		fmt.Sprintf("racks/%s/pdus/", ts.rackName()),
		fmt.Sprintf("racks/%s/tors/", ts.rackName()),
		fmt.Sprintf("racks/%s/blades/", ts.rackName()),
		ts.timers)
	require.NotNil(r)
	ctx := ts.advance(context.Background())

	require.NoError(r.start(context.Background()))

	// Powered off, so this should fail
	require.Nil(ts.issueGetStatus(ctx, r, 0))

	p := ts.issueSetPower(ctx, r, 0, true)
	require.NotNil(p)
	require.NoError(p.Err)

	ctx = ts.advanceToStateChange(ctx, 5, func() bool {
		return r.blades[0].sm.CurrentIndex == pb.BladeState_powered_disconnected
	})

	// Powered on, but disconnected, so this should do nothing
	require.Nil(ts.issueGetStatus(ctx, r, 0))

	c := ts.issueSetConnection(ctx, r, 0, true)
	require.NotNil(c)
	require.NoError(c.Err)

	ctx = ts.advanceToStateChange(ctx, 8, func() bool {
		return r.blades[0].sm.CurrentIndex == pb.BladeState_booting
	})

	sres := ts.issueGetStatus(ctx, r, 0)
	require.NotNil(sres)
	require.NoError(sres.Err)

	status, ok := sres.Msg.(*messages.BladeStatus)
	require.True(ok)

	require.Equal(pb.BladeState_booting.String(), status.State)

	ctx = ts.advanceToStateChange(ctx, 8, func() bool {
		return r.blades[0].sm.CurrentIndex == pb.BladeState_working
	})

	sres = ts.issueGetStatus(ctx, r, 0)
	require.NotNil(sres)
	require.NoError(sres.Err)

	status, ok = sres.Msg.(*messages.BladeStatus)
	require.True(ok)

	require.Equal(pb.BladeState_working.String(), status.State)
}

func (ts *BladeTestSuite) TestPowerOn() {
	require := ts.Require()

	rackDef := ts.createDummyRack(2)

	r := newRack(
		context.Background(),
		ts.rackName(),
		rackDef,
		ts.cfg,
		fmt.Sprintf("racks/%s/pdus/", ts.rackName()),
		fmt.Sprintf("racks/%s/tors/", ts.rackName()),
		fmt.Sprintf("racks/%s/blades/", ts.rackName()),
		ts.timers)
	require.NotNil(r)
	ctx := ts.advance(context.Background())

	ctx, span := tracing.StartSpan(
		ctx,
		tracing.WithName("test powering on a blade"),
		tracing.WithContextValue(tsc.EnsureTickInContext))

	require.NoError(r.start(context.Background()))

	res := ts.issueSetPower(ctx, r, 0, true)
	span.End()

	require.NotNil(res)
	require.NoError(res.Err)
	require.Equal(common.TickFromContext(ctx), res.At)
	require.Nil(res.Msg)

	ctx = ts.advanceToStateChange(ctx, 5, func() bool {
		return r.blades[0].sm.CurrentIndex == pb.BladeState_powered_disconnected
	})

	res2 := ts.issueGetStatus(ctx, r, 0)
	require.Nil(res2)
}

func (ts *BladeTestSuite) TestPowerOnPersistence() {
	require := ts.Require()

	rackDef := ts.createDummyRack(2)

	r := newRack(
		context.Background(),
		ts.rackName(),
		rackDef,
		ts.cfg,
		fmt.Sprintf("racks/%s/pdus/", ts.rackName()),
		fmt.Sprintf("racks/%s/tors/", ts.rackName()),
		fmt.Sprintf("racks/%s/blades/", ts.rackName()),
		ts.timers)
	require.NotNil(r)
	ctx := ts.advance(context.Background())

	ctx, span := tracing.StartSpan(
		ctx,
		tracing.WithName("test powering on a blade"),
		tracing.WithContextValue(tsc.EnsureTickInContext))

	require.NoError(r.start(context.Background()))

	res := ts.issueSetPower(ctx, r, 0, true)
	span.End()

	require.NotNil(res)
	require.NoError(res.Err)
	require.Equal(common.TickFromContext(ctx), res.At)
	require.Nil(res.Msg)

	ctx = ts.advanceToStateChange(ctx, 5, func() bool {
		return r.blades[0].sm.CurrentIndex == pb.BladeState_powered_disconnected
	})

	res2 := ts.issueGetStatus(ctx, r, 0)
	require.Nil(res2)

	saved, err := r.blades[0].Save()
	require.NoError(err)

	m := jsonpb.Marshaler{}
	_, err = m.MarshalToString(saved)
	require.NoError(err)
}

func (ts *BladeTestSuite) TestPowerOnOffWhileBooting() {
	require := ts.Require()

	ctx, r := ts.createAndStartRack(context.Background(), 2, false, false)

	ctx, span := tracing.StartSpan(
		ctx,
		tracing.WithName("test powering on a blade"),
		tracing.WithContextValue(tsc.EnsureTickInContext))

	res := ts.issueSetPower(ctx, r, 0, true)

	require.NotNil(res)
	require.NoError(res.Err)
	require.Equal(common.TickFromContext(ctx), res.At)
	require.Nil(res.Msg)

	ctx = ts.advanceToStateChange(ctx, 2, func() bool {
		return r.blades[0].sm.CurrentIndex == pb.BladeState_powered_disconnected
	})

	res2 := ts.issueGetStatus(ctx, r, 0)
	require.Nil(res2)

	res = ts.issueSetConnection(ctx, r, 0, true)

	require.NotNil(res)
	require.NoError(res.Err)
	require.Equal(common.TickFromContext(ctx), res.At)
	require.Nil(res.Msg)

	ctx = ts.advanceToStateChange(ctx, 2, func() bool {
		return r.blades[0].sm.CurrentIndex == pb.BladeState_booting
	})

	res2 = ts.issueGetStatus(ctx, r, 0)
	require.NotNil(res2)
	require.NoError(res2.Err)
	require.NotNil(res2.Msg)
	status := res2.Msg.(*messages.BladeStatus)
	require.Equal(pb.BladeState_booting.String(), status.State)

	span.End()

	ctx, span = tracing.StartSpan(
		ctx,
		tracing.WithName("test powering off a blade"),
		tracing.WithContextValue(tsc.EnsureTickInContext))

	res = ts.issueSetPower(ctx, r, 0, false)

	require.NotNil(res)
	require.NoError(res.Err)

	ctx = ts.advance(ctx)
	ctx = ts.advance(ctx)

	require.Eventually(func() bool {
		return r.blades[0].sm.CurrentIndex == pb.BladeState_off_connected
	}, time.Second, 10*time.Millisecond,
		"state is %v", r.blades[0].sm.CurrentIndex)

	res2 = ts.issueGetStatus(ctx, r, 0)
	require.Nil(res2)

	span.End()
}

func (ts *BladeTestSuite) TestWorkingToIsolatedToWorking() {
	require := ts.Require()

	ctx, r := ts.createAndStartRack(context.Background(), 2, false, false)
	target := messages.NewTargetBlade(ts.rackName(), 0, 0)
	ctx = ts.bootBlade(ctx, r, target)

	ctx, span := tracing.StartSpan(
		ctx,
		tracing.WithName("Test working to isolated"),
		tracing.WithContextValue(tsc.EnsureTickInContext))

	res := ts.issueSetConnection(ctx, r, 0, false)
	require.NotNil(res)
	require.NoError(res.Err)

	ctx = ts.advance(ctx)
	ctx = ts.advance(ctx)

	require.Eventually(func() bool {
		return r.blades[0].sm.CurrentIndex == pb.BladeState_isolated
	}, time.Second, 10*time.Millisecond,
		"state is %v", r.blades[0].sm.CurrentIndex)

	span.End()

	ctx, span = tracing.StartSpan(
		ctx,
		tracing.WithName("Test isolated to working"),
		tracing.WithContextValue(tsc.EnsureTickInContext))

	res = ts.issueSetConnection(ctx, r, 0, true)
	require.NotNil(res)
	require.NoError(res.Err)

	ctx = ts.advance(ctx)
	ctx = ts.advance(ctx)

	require.Eventually(func() bool {
		return r.blades[0].sm.CurrentIndex == pb.BladeState_working
	}, time.Second, 10*time.Millisecond,
		"state is %v", r.blades[0].sm.CurrentIndex)

	span.End()
}

func (ts *BladeTestSuite) TestWorkingToOffConn() {
	require := ts.Require()

	ctx, r := ts.createAndStartRack(context.Background(), 2, false, false)
	target := messages.NewTargetBlade(ts.rackName(), 0, 0)
	ctx = ts.bootBlade(ctx, r, target)

	ctx, span := tracing.StartSpan(
		ctx,
		tracing.WithName("Test working to off"),
		tracing.WithContextValue(tsc.EnsureTickInContext))

	res := ts.issueSetPower(ctx, r, 0, false)
	require.NotNil(res)
	require.NoError(res.Err)

	ctx = ts.advance(ctx)
	ctx = ts.advance(ctx)

	require.Eventually(func() bool {
		return r.blades[0].sm.CurrentIndex == pb.BladeState_off_connected
	}, time.Second, 10*time.Millisecond,
		"state is %v", r.blades[0].sm.CurrentIndex)

	span.End()
}

func (ts *BladeTestSuite) TestOffConnToOffDiscon() {
	require := ts.Require()

	ctx, r := ts.createAndStartRack(context.Background(), 2, false, false)

	target := messages.NewTargetBlade(ts.rackName(), 0, 0)
	ctx = ts.bootBlade(ctx, r, target)

	ctx, span := tracing.StartSpan(
		ctx,
		tracing.WithName("Test working to off-disconnected"),
		tracing.WithContextValue(tsc.EnsureTickInContext))

	res := ts.issueSetPower(ctx, r, 0, false)
	require.NotNil(res)
	require.NoError(res.Err)

	ctx = ts.advance(ctx)
	ctx = ts.advance(ctx)

	doneConnTest := func() bool {
		return r.blades[0].sm.CurrentIndex == pb.BladeState_off_connected
	}

	require.Eventually(doneConnTest, time.Second, 10*time.Millisecond,
		"state is %v", r.blades[0].sm.CurrentIndex)

	res = ts.issueSetConnection(ctx, r, 0, false)
	require.NotNil(res)
	require.NoError(res.Err)

	ctx = ts.advance(ctx)
	ctx = ts.advance(ctx)

	doneTest := func() bool {
		return r.blades[0].sm.CurrentIndex == pb.BladeState_off_disconnected
	}

	require.Eventually(doneTest, time.Second, 10*time.Millisecond,
		"state is %v", r.blades[0].sm.CurrentIndex)

	span.End()
}

func (ts *BladeTestSuite) TestDuplicateOffDiscon() {
	require := ts.Require()

	ctx, r := ts.createAndStartRack(context.Background(), 2, false, false)

	ctx, span := tracing.StartSpan(
		ctx,
		tracing.WithName("Test duplicate off-disconnected"),
		tracing.WithContextValue(tsc.EnsureTickInContext))

	res := ts.issueSetPower(ctx, r, 0, false)
	require.NotNil(res)
	require.Equal(errors.ErrNoOperation, res.Err)

	doneTest := func() bool {
		return r.blades[0].sm.CurrentIndex == pb.BladeState_off_disconnected
	}

	require.Eventually(doneTest, time.Second, 10*time.Millisecond,
		"state is %v", r.blades[0].sm.CurrentIndex)

	res = ts.issueSetConnection(ctx, r, 0, false)
	require.NotNil(res)
	require.Equal(errors.ErrNoOperation, res.Err)

	require.Eventually(doneTest, time.Second, 10*time.Millisecond,
		"state is %v", r.blades[0].sm.CurrentIndex)

	span.End()
}

func TestBladeSuite(t *testing.T) {
	suite.Run(t, new(BladeTestSuite))
}
