package inventory

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	tsc "github.com/Jim3Things/CloudChamber/simulation/internal/clients/timestamp"
	"github.com/Jim3Things/CloudChamber/simulation/internal/common"
	"github.com/Jim3Things/CloudChamber/simulation/internal/services/inventory/messages"
	"github.com/Jim3Things/CloudChamber/simulation/internal/sm"
	"github.com/Jim3Things/CloudChamber/simulation/internal/tracing"
	"github.com/Jim3Things/CloudChamber/simulation/pkg/errors"
	"github.com/Jim3Things/CloudChamber/simulation/test/utilities"
)

type TorTestSuite struct {
	testSuiteCore
}

func (ts *TorTestSuite) TestCreateTor() {
	require := ts.Require()
	assert := ts.Assert()

	rackDef := ts.createDummyRack(2)

	r := newRack(context.Background(), ts.rackName(), rackDef, ts.timers)
	require.NotNil(r)

	t := r.tor
	require.NotNil(t)

	assert.Equal(2, len(t.cables))

	assert.Equal(torWorkingState, t.sm.CurrentIndex)

	for _, c := range t.cables {
		assert.False(c.on)
		assert.False(c.faulted)
	}
}

func (ts *TorTestSuite) TestBadConnectionTarget() {
	require := ts.Require()
	assert := ts.Assert()

	ctx, r := ts.createAndStartRack(context.Background(), 2, true, false)
	t := r.tor

	require.NotNil(t)

	for i := range t.cables {
		t.cables[i] = newCable(true, false, 0)
	}

	rsp := make(chan *sm.Response)

	r.Receive(
		messages.NewSetConnection(
			ctx,
			messages.NewTargetTor(ts.rackName()),
			common.TickFromContext(ctx),
			false,
			rsp))

	res, ok := ts.completeWithin(rsp, time.Duration(1)*time.Second)
	require.True(ok)
	require.NotNil(res)
	assert.Error(res.Err)
	assert.Equal(errors.ErrInvalidTarget, res.Err)
	assert.Equal(common.TickFromContext(ctx), res.At)
	assert.Nil(res.Msg)

	assert.Equal(torWorkingState, t.sm.CurrentIndex)

	for _, c := range t.cables {
		assert.True(c.on)
	}
}

func (ts *TorTestSuite) TestConnectTooLate() {
	require := ts.Require()
	assert := ts.Assert()

	ctx := ts.advance(context.Background())
	commandTime := common.TickFromContext(ctx)

	ctx, r := ts.createAndStartRack(ctx, 2, true, false)
	t := r.tor
	require.NotNil(t)

	ctx = ts.advance(ctx)
	checkTime := common.TickFromContext(ctx)

	rsp := make(chan *sm.Response)

	msg := messages.NewSetConnection(
		ctx,
		messages.NewTargetBlade(ts.rackName(), 0),
		commandTime,
		true,
		rsp)

	r.Receive(msg)

	res, ok := ts.completeWithin(rsp, time.Duration(1)*time.Second)
	require.True(ok)
	require.NotNil(res)
	require.Error(errors.ErrInventoryChangeTooLate(commandTime), res.Err)

	assert.Less(t.sm.Guard, checkTime)
	assert.Less(t.cables[0].Guard, checkTime)
	assert.False(t.cables[0].on)

	assert.Equal(torWorkingState, t.sm.CurrentIndex)
}

func (ts *TorTestSuite) TestConnectBlade() {
	require := ts.Require()
	assert := ts.Assert()

	ctx, r := ts.createAndStartRack(context.Background(), 2, false, false)

	ok := utilities.WaitForStateChange(1, func() bool {
		return r.tor.sm.CurrentIndex == torWorkingState
	})

	require.True(ok, "state is %v", r.tor.sm.CurrentIndex)

	ok = utilities.WaitForStateChange(1, func() bool {
		return r.blades[0].sm.CurrentIndex == bladeOffDiscon
	})

	require.True(ok, "state is %v", r.blades[0].sm.CurrentIndex)

	t := r.tor
	require.NotNil(t)

	ctx, span := tracing.StartSpan(
		ctx,
		tracing.WithName("test connecting up a blade"),
		tracing.WithContextValue(tsc.EnsureTickInContext))

	rsp := make(chan *sm.Response)

	r.Receive(
		messages.NewSetConnection(
			ctx,
			messages.NewTargetBlade(ts.rackName(), 0),
			common.TickFromContext(ctx),
			true,
			rsp))

	span.End()

	res, ok := ts.completeWithin(rsp, time.Duration(1)*time.Second)
	require.True(ok)
	require.NotNil(res)

	assert.NoError(res.Err)

	assert.Equal(common.TickFromContext(ctx), res.At)
	assert.Nil(res.Msg)

	assert.Equal(common.TickFromContext(ctx), t.sm.Guard)
	assert.Equal(common.TickFromContext(ctx), t.cables[0].Guard)
	assert.True(t.cables[0].on)
	assert.False(t.cables[0].faulted)
	ok = utilities.WaitForStateChange(1, func() bool {
		return r.blades[0].sm.CurrentIndex == bladeOffConn
	})

	require.True(ok, "state is %v", r.blades[0].sm.CurrentIndex)

	assert.Equal(torWorkingState, t.sm.CurrentIndex)
}

func (ts *TorTestSuite) TestConnectBladeWhileWorking() {
	require := ts.Require()
	assert := ts.Assert()

	ctx, r := ts.createAndStartRack(context.Background(), 2, false, false)

	ctx = ts.bootBlade(ctx, r, 0)

	t := r.tor
	require.NotNil(t)

	require.True(t.cables[0].on)
	require.False(t.cables[0].faulted)
	require.Equal(bladeWorking, r.blades[0].sm.CurrentIndex)

	ctx, span := tracing.StartSpan(
		ctx,
		tracing.WithName("test connecting up a blade"),
		tracing.WithContextValue(tsc.EnsureTickInContext))

	rsp := make(chan *sm.Response)

	r.Receive(
		messages.NewSetConnection(
			ctx,
			messages.NewTargetBlade(ts.rackName(), 0),
			common.TickFromContext(ctx),
			false,
			rsp))

	span.End()

	res, ok := ts.completeWithin(rsp, time.Duration(1)*time.Second)
	require.True(ok)
	require.NotNil(res)

	assert.NoError(res.Err)

	assert.Equal(common.TickFromContext(ctx), res.At)
	assert.Nil(res.Msg)

	assert.Equal(common.TickFromContext(ctx), t.sm.Guard)
	assert.Equal(common.TickFromContext(ctx), t.cables[0].Guard)
	assert.False(t.cables[0].on)
	assert.False(t.cables[0].faulted)
	ok = utilities.WaitForStateChange(1, func() bool {
		return r.blades[0].sm.CurrentIndex == bladeIsolated
	})

	require.True(ok, "state is %v", r.blades[0].sm.CurrentIndex)

	assert.Equal(torWorkingState, t.sm.CurrentIndex)
}

func (ts *TorTestSuite) TestStuckCable() {
	require := ts.Require()
	assert := ts.Assert()

	ctx, r := ts.createAndStartRack(context.Background(), 2, true, false)
	t := r.tor

	t.cables[0].faulted = true

	rsp := make(chan *sm.Response)

	commandTime := common.TickFromContext(ctx)
	msg := messages.NewSetConnection(
		ctx,
		messages.NewTargetBlade(ts.rackName(), 0),
		commandTime,
		true,
		rsp)

	r.Receive(msg)

	res, ok := ts.completeWithin(rsp, time.Duration(1)*time.Second)
	require.True(ok)
	require.NotNil(res)

	assert.Error(res.Err)
	assert.Equal(errors.ErrCableStuck, res.Err)

	assert.Equal(common.TickFromContext(ctx), res.At)
	assert.Nil(res.Msg)

	assert.Less(t.sm.Guard, commandTime)
	assert.Less(t.cables[0].Guard, commandTime)
	assert.False(t.cables[0].on)
	assert.Equal(true, t.cables[0].faulted)

	assert.Equal(torWorkingState, t.sm.CurrentIndex)
}

func (ts *TorTestSuite) TestWorkingGetStatus() {
	require := ts.Require()
	assert := ts.Assert()

	ctx, r := ts.createAndStartRack(context.Background(), 2, true, false)

	rsp := make(chan *sm.Response)
	msg := messages.NewGetStatus(
		ctx,
		messages.NewTargetTor(ts.rackName()),
		common.TickFromContext(ctx),
		rsp)

	r.Receive(msg)

	res, ok := ts.completeWithin(rsp, time.Duration(1)*time.Second)
	require.True(ok)
	require.NotNil(res)

	assert.NoError(res.Err)
	assert.Equal(common.TickFromContext(ctx), res.At)
	require.NotNil(res.Msg)

	status := res.Msg.(*messages.TorStatus)

	assert.Equal(torWorkingState, status.State)
	assert.Equal(int64(0), status.EnteredAt)

	for i, c := range status.Cables {
		assert.False(c.On, "Cable.On[%d]", i)
		assert.False(c.Faulted, "Cable.Faulted[%d]", i)
	}
}

func TestTorTestSuite(t *testing.T) {
	suite.Run(t, new(TorTestSuite))
}
