package inventory

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/Jim3Things/CloudChamber/internal/common"
	"github.com/Jim3Things/CloudChamber/internal/sm"
)

type TorTestSuite struct {
	testSuiteCore
}

func (ts *TorTestSuite) TestCreateTor() {
	require := ts.Require()
	assert := ts.Assert()

	rackDef := ts.createDummyRack(2)

	r := newRack(context.Background(), ts.rackName(), rackDef)
	require.NotNil(r)

	t := r.tor
	require.NotNil(t)

	assert.Equal(2, len(t.cables))

	assert.Equal("working", t.sm.Current.Name())

	for _, c := range t.cables {
		assert.False(c.on)
		assert.False(c.faulted)
	}
}

func (ts *TorTestSuite) TestBadConnectionTarget() {
	require := ts.Require()
	assert := ts.Assert()

	ctx := common.ContextWithTick(context.Background(), 1)

	rackDef := ts.createDummyRack(2)

	r := newRack(ctx, ts.rackName(), rackDef)
	ctx = common.ContextWithTick(ctx, 2)

	t := r.tor
	require.NotNil(t)

	for i := range t.cables {
		t.cables[i] = newCable(true, false, 0)
	}

	rsp := make(chan *sm.Response)

	badMsg := newSetConnection(ctx, newTargetTor(ts.rackName()), common.TickFromContext(ctx), false, rsp)

	ts.execute(ctx, badMsg, r.tor.Receive)

	res := ts.completeWithin(rsp, time.Duration(1)*time.Second)
	require.NotNil(res)
	assert.Error(res.Err)
	assert.Equal(ErrInvalidTarget, res.Err)
	assert.Equal(common.TickFromContext(ctx), res.At)
	assert.Nil(res.Msg)

	assert.Equal("working", t.sm.Current.Name())

	for _, c := range t.cables {
		assert.True(c.on)
	}
}

func (ts *TorTestSuite) TestConnectBlade() {
	require := ts.Require()
	assert := ts.Assert()

	ctx := common.ContextWithTick(context.Background(), 1)

	rackDef := ts.createDummyRack(2)

	r := newRack(ctx, ts.rackName(), rackDef)
	require.NotNil(r)
	t := r.tor

	ctx = common.ContextWithTick(ctx, 2)

	rsp := make(chan *sm.Response)

	msg := newSetConnection(ctx, newTargetBlade(ts.rackName(), 0), common.TickFromContext(ctx), true, rsp)

	ts.execute(ctx, msg, r.tor.Receive)

	res := ts.completeWithin(rsp, time.Duration(1)*time.Second)
	require.NotNil(res)

	assert.NoError(res.Err)

	assert.Equal(common.TickFromContext(ctx), res.At)
	assert.Nil(res.Msg)

	assert.Equal(common.TickFromContext(ctx), t.sm.Guard)
	assert.Equal(common.TickFromContext(ctx), t.cables[0].Guard)
	assert.True(t.cables[0].on)
	assert.False(t.cables[0].faulted)

	assert.Equal("working", t.sm.Current.Name())
}

func (ts *TorTestSuite) TestConnectTooLate() {
	require := ts.Require()
	assert := ts.Assert()

	startTime := int64(1)
	ctx := common.ContextWithTick(context.Background(), startTime)

	rackDef := ts.createDummyRack(2)

	r := newRack(ctx, ts.rackName(), rackDef)
	t := r.tor
	ctx = common.ContextWithTick(ctx, 2)

	rsp := make(chan *sm.Response)

	msg := newSetConnection(ctx, newTargetBlade(ts.rackName(), 0), startTime - 1, true, rsp)

	ts.execute(ctx, msg, r.tor.Receive)

	res := ts.completeWithin(rsp, time.Duration(1)*time.Second)
	require.NotNil(res)

	require.Error(res.Err)
	assert.Equal(ErrRepairMessageDropped, res.Err)

	assert.Equal(common.TickFromContext(ctx), res.At)
	assert.Nil(res.Msg)

	assert.Equal(startTime, t.sm.Guard)
	assert.Equal(startTime, t.cables[0].Guard)
	assert.False(t.cables[0].on)

	assert.Equal("working", t.sm.Current.Name())
}

func (ts *TorTestSuite) TestStuckCable() {
	require := ts.Require()
	assert := ts.Assert()

	ctx := common.ContextWithTick(context.Background(), 1)

	rackDef := ts.createDummyRack(2)

	r := newRack(ctx, ts.rackName(), rackDef)
	t := r.tor

	startTime := common.TickFromContext(ctx)
	require.Nil(t.cables[0].fault(false, startTime, startTime))
	ctx = common.ContextWithTick(ctx, 2)

	rsp := make(chan *sm.Response)

	msg := newSetConnection(ctx, newTargetBlade(ts.rackName(), 0), common.TickFromContext(ctx), true, rsp)

	ts.execute(ctx, msg, r.tor.Receive)

	res := ts.completeWithin(rsp, time.Duration(1)*time.Second)
	require.NotNil(res)

	assert.Error(res.Err)
	assert.Equal(ErrCableStuck, res.Err)

	assert.Equal(common.TickFromContext(ctx), res.At)
	assert.Nil(res.Msg)

	assert.Equal(startTime, t.sm.Guard)
	assert.Equal(startTime, t.cables[0].Guard)
	assert.False(t.cables[0].on)
	assert.Equal(true, t.cables[0].faulted)

	assert.Equal("working", t.sm.Current.Name())
}

func TestTorTestSuite(t *testing.T) {
	suite.Run(t, new(TorTestSuite))
}
