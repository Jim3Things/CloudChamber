package inventory

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	tsc "github.com/Jim3Things/CloudChamber/internal/clients/timestamp"
	"github.com/Jim3Things/CloudChamber/internal/common"
	"github.com/Jim3Things/CloudChamber/internal/services/inventory/messages"
	"github.com/Jim3Things/CloudChamber/internal/sm"
	"github.com/Jim3Things/CloudChamber/internal/tracing"
	"github.com/Jim3Things/CloudChamber/test/utilities"
)

type BladeTestSuite struct {
	testSuiteCore
}

func (ts *BladeTestSuite) issueSetPower(ctx context.Context, r *Rack, id int64, on bool) *sm.Response {
	require := ts.Require()

	rsp := make(chan *sm.Response)

	msg := messages.NewSetPower(
		ctx,
		messages.NewTargetBlade(ts.rackName(), id),
		common.TickFromContext(ctx),
		on,
		rsp)

	r.Receive(msg)

	res, ok := ts.completeWithin(rsp, time.Duration(1)*time.Second)
	require.True(ok)

	return res
}

func (ts *BladeTestSuite) issueSetConnection(ctx context.Context, r *Rack, id int64, on bool) *sm.Response {
	require := ts.Require()

	rsp := make(chan *sm.Response)

	msg := messages.NewSetConnection(
		ctx,
		messages.NewTargetBlade(ts.rackName(), id),
		common.TickFromContext(ctx),
		on,
		rsp)

	r.Receive(msg)

	res, ok := ts.completeWithin(rsp, time.Duration(1)*time.Second)
	require.True(ok)

	return res
}

func (ts *BladeTestSuite) issueGetStatus(ctx context.Context, r *Rack, id int64) *sm.Response {
	require := ts.Require()

	rsp := make(chan *sm.Response)
	msg := messages.NewGetStatus(ctx,
		messages.NewTargetBlade(ts.rackName(), id),
		common.TickFromContext(ctx),
		rsp)

	r.Receive(msg)

	res, ok := ts.completeWithin(rsp, time.Duration(1)*time.Second)
	require.True(ok)

	return res
}

func (ts *BladeTestSuite) TestGetStatus() {
	require := ts.Require()
	assert := ts.Assert()

	rackDef := ts.createDummyRack(2)

	r := newRack(context.Background(), ts.rackName(), rackDef, ts.timers)
	require.NotNil(r)
	ctx := ts.advance(context.Background())

	require.NoError(r.start(context.Background()))

	// Powered off, so this should fail
	sres := ts.issueGetStatus(ctx, r, 0)
	require.Nil(sres)

	p := ts.issueSetPower(ctx, r, 0, true)
	require.NotNil(p)
	require.NoError(p.Err)

	ctx, ok := ts.advanceToStateChange(ctx, 5, func() bool {
		return r.blades[0].sm.CurrentIndex == bladePoweredDiscon
	})

	require.True(ok, "state is %v", r.blades[0].sm.CurrentIndex)

	// Powered on, but disconnected, so this should do nothing
	sres = ts.issueGetStatus(ctx, r, 0)
	require.Nil(sres)

	c := ts.issueSetConnection(ctx, r, 0, true)
	require.NotNil(c)
	require.NoError(c.Err)

	ctx, ok = ts.advanceToStateChange(ctx, 5, func() bool {
		return r.blades[0].sm.CurrentIndex == bladeBooting
	})

	require.True(ok, "state is %v", r.blades[0].sm.CurrentIndex)

	sres = ts.issueGetStatus(ctx, r, 0)
	require.NotNil(sres)
	require.NoError(sres.Err)

	status, ok := sres.Msg.(*messages.BladeStatus)
	require.True(ok)

	assert.Equal(bladeBooting, status.State)

	ctx, ok = ts.advanceToStateChange(ctx, 5, func() bool {
		return r.blades[0].sm.CurrentIndex == bladeWorking
	})

	require.True(ok, "state is %v", r.blades[0].sm.CurrentIndex)

	sres = ts.issueGetStatus(ctx, r, 0)
	require.NotNil(sres)
	require.NoError(sres.Err)

	status, ok = sres.Msg.(*messages.BladeStatus)
	require.True(ok)

	assert.Equal(bladeWorking, status.State)
}

func (ts *BladeTestSuite) TestPowerOn() {
	require := ts.Require()
	assert := ts.Assert()

	rackDef := ts.createDummyRack(2)

	r := newRack(context.Background(), ts.rackName(), rackDef, ts.timers)
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
	assert.Equal(common.TickFromContext(ctx), res.At)
	assert.Nil(res.Msg)

	ctx, ok := ts.advanceToStateChange(ctx, 5, func() bool {
		return r.blades[0].sm.CurrentIndex == bladePoweredDiscon
	})

	require.True(ok, "state is %v", r.blades[0].sm.CurrentIndex)

	res2 := ts.issueGetStatus(ctx, r, 0)
	require.Nil(res2)
}

func (ts *BladeTestSuite) TestPowerOnOffWhileBooting() {
	require := ts.Require()
	assert := ts.Assert()

	ctx, r := ts.createAndStartRack(context.Background(), 2, false, false)

	ctx, span := tracing.StartSpan(
		ctx,
		tracing.WithName("test powering on a blade"),
		tracing.WithContextValue(tsc.EnsureTickInContext))

	res := ts.issueSetPower(ctx, r, 0, true)

	require.NotNil(res)
	require.NoError(res.Err)
	assert.Equal(common.TickFromContext(ctx), res.At)
	assert.Nil(res.Msg)

	ctx, ok := ts.advanceToStateChange(ctx, 2, func() bool {
		return r.blades[0].sm.CurrentIndex == bladePoweredDiscon
	})

	require.True(ok, "state is %v", r.blades[0].sm.CurrentIndex)

	res2 := ts.issueGetStatus(ctx, r, 0)
	require.Nil(res2)

	res = ts.issueSetConnection(ctx, r, 0, true)

	require.NotNil(res)
	require.NoError(res.Err)
	assert.Equal(common.TickFromContext(ctx), res.At)
	assert.Nil(res.Msg)

	ctx, ok = ts.advanceToStateChange(ctx, 2, func() bool {
		return r.blades[0].sm.CurrentIndex == bladeBooting
	})

	require.True(ok, "state is %v", r.blades[0].sm.CurrentIndex)

	res2 = ts.issueGetStatus(ctx, r, 0)
	require.NotNil(res2)
	require.NoError(res2.Err)
	require.NotNil(res2.Msg)
	status := res2.Msg.(*messages.BladeStatus)
	assert.Equal(bladeBooting, status.State)

	span.End()

	ctx, span = tracing.StartSpan(
		ctx,
		tracing.WithName("test powering off a blade"),
		tracing.WithContextValue(tsc.EnsureTickInContext))

	res = ts.issueSetPower(ctx, r, 0, false)

	require.NotNil(res)
	require.NoError(res.Err)

	ok = utilities.WaitForStateChange(1, func() bool {
		return r.blades[0].sm.CurrentIndex == bladeOffConn
	})

	require.True(ok, "state is %v", r.blades[0].sm.CurrentIndex)

	res2 = ts.issueGetStatus(ctx, r, 0)
	require.Nil(res2)

	span.End()
}

func (ts *BladeTestSuite) TestWorkingToIsolatedToWorking() {
	require := ts.Require()

	ctx, r := ts.createAndStartRack(context.Background(), 2, false, false)

	ctx = ts.bootBlade(ctx, r, 0)

	ctx, span := tracing.StartSpan(
		ctx,
		tracing.WithName("Test working to isolated"),
		tracing.WithContextValue(tsc.EnsureTickInContext))

	res := ts.issueSetConnection(ctx, r, 0, false)
	require.NotNil(res)
	require.NoError(res.Err)

	ok := utilities.WaitForStateChange(1, func() bool {
		return r.blades[0].sm.CurrentIndex == bladeIsolated
	})

	require.True(ok, "state is %v", r.blades[0].sm.CurrentIndex)

	span.End()

	ctx, span = tracing.StartSpan(
		ctx,
		tracing.WithName("Test isolated to working"),
		tracing.WithContextValue(tsc.EnsureTickInContext))

	res = ts.issueSetConnection(ctx, r, 0, true)
	require.NotNil(res)
	require.NoError(res.Err)

	ok = utilities.WaitForStateChange(1, func() bool {
		return r.blades[0].sm.CurrentIndex == bladeWorking
	})

	require.True(ok, "state is %v", r.blades[0].sm.CurrentIndex)

	span.End()
}

func (ts *BladeTestSuite) TestWorkingToOffConn() {
	require := ts.Require()

	ctx, r := ts.createAndStartRack(context.Background(), 2, false, false)

	ctx = ts.bootBlade(ctx, r, 0)

	ctx, span := tracing.StartSpan(
		ctx,
		tracing.WithName("Test working to off"),
		tracing.WithContextValue(tsc.EnsureTickInContext))

	res := ts.issueSetPower(ctx, r, 0, false)
	require.NotNil(res)
	require.NoError(res.Err)

	ok := utilities.WaitForStateChange(1, func() bool {
		return r.blades[0].sm.CurrentIndex == bladeOffConn
	})

	require.True(ok, "state is %v", r.blades[0].sm.CurrentIndex)

	span.End()
}

func (ts *BladeTestSuite) TestOffConnToOffDiscon() {
	require := ts.Require()

	ctx, r := ts.createAndStartRack(context.Background(), 2, false, false)

	ctx = ts.bootBlade(ctx, r, 0)

	ctx, span := tracing.StartSpan(
		ctx,
		tracing.WithName("Test working to off-disconnected"),
		tracing.WithContextValue(tsc.EnsureTickInContext))

	res := ts.issueSetPower(ctx, r, 0, false)
	require.NotNil(res)
	require.NoError(res.Err)

	ok := utilities.WaitForStateChange(1, func() bool {
		return r.blades[0].sm.CurrentIndex == bladeOffConn
	})

	require.True(ok, "state is %v", r.blades[0].sm.CurrentIndex)

	res = ts.issueSetConnection(ctx, r, 0, false)
	require.NotNil(res)
	require.NoError(res.Err)

	ok = utilities.WaitForStateChange(1, func() bool {
		return r.blades[0].sm.CurrentIndex == bladeOffDiscon
	})

	require.True(ok, "state is %v", r.blades[0].sm.CurrentIndex)

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
	require.Equal(ErrNoOperation, res.Err)

	ok := utilities.WaitForStateChange(1, func() bool {
		return r.blades[0].sm.CurrentIndex == bladeOffDiscon
	})

	require.True(ok, "state is %v", r.blades[0].sm.CurrentIndex)

	res = ts.issueSetConnection(ctx, r, 0, false)
	require.NotNil(res)
	require.Equal(ErrNoOperation, res.Err)

	ok = utilities.WaitForStateChange(1, func() bool {
		return r.blades[0].sm.CurrentIndex == bladeOffDiscon
	})

	require.True(ok, "state is %v", r.blades[0].sm.CurrentIndex)

	span.End()
}

func TestBladeSuite(t *testing.T) {
	suite.Run(t, new(BladeTestSuite))
}
