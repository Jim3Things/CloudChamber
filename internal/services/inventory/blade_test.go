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
)

type BladeTestSuite struct {
	testSuiteCore
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

	rsp := make(chan *sm.Response)

	msg := messages.NewSetPower(
		ctx,
		messages.NewTargetBlade(ts.rackName(), 0),
		common.TickFromContext(ctx),
		true,
		rsp)
	r.Receive(msg)

	span.End()

	res := ts.completeWithin(rsp, time.Duration(1)*time.Second)
	require.NotNil(res)
	require.NoError(res.Err)
	assert.Equal(common.TickFromContext(ctx), res.At)
	assert.Nil(res.Msg)

	ts.advanceToStateChange(ctx, 5, func() bool {
		return r.blades[0].sm.CurrentIndex == bladeWorkingState
	})
}

func (ts *BladeTestSuite) TestPowerOnOffWhileBooting() {
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

	rsp := make(chan *sm.Response)

	r.Receive(
		messages.NewSetPower(
			ctx,
			messages.NewTargetBlade(ts.rackName(), 0),
			common.TickFromContext(ctx),
			true,
			rsp))

	res := ts.completeWithin(rsp, time.Duration(1)*time.Second)
	require.NotNil(res)
	require.NoError(res.Err)
	assert.Equal(common.TickFromContext(ctx), res.At)
	assert.Nil(res.Msg)

	ts.advanceToStateChange(ctx, 2, func() bool {
		return r.blades[0].sm.CurrentIndex == bladeBootingState
	})

	span.End()

	ctx, span = tracing.StartSpan(
		ctx,
		tracing.WithName("test powering off a blade"),
		tracing.WithContextValue(tsc.EnsureTickInContext))

	r.Receive(
		messages.NewSetPower(
			ctx,
			messages.NewTargetBlade(ts.rackName(), 0),
			common.TickFromContext(ctx),
			false,
			rsp))

	span.End()

	ctx, span = tracing.StartSpan(
		ctx,
		tracing.WithName("test powering off a blade"),
		tracing.WithContextValue(tsc.EnsureTickInContext))

	res = ts.completeWithin(rsp, time.Duration(1)*time.Second)
	require.NotNil(res)
	require.NoError(res.Err)
	span.End()

	ts.waitForStateChange(func() bool {
		return r.blades[0].sm.CurrentIndex == bladeOffState
	})
}

func TestBladeSuite(t *testing.T) {
	suite.Run(t, new(BladeTestSuite))
}
