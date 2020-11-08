package inventory

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/Jim3Things/CloudChamber/internal/common"
	"github.com/Jim3Things/CloudChamber/internal/sm"
)

type PduTestSuite struct {
	testSuiteCore
}

func (ts *PduTestSuite) TestCreatePdu() {
	require := ts.Require()
	assert := ts.Assert()

	rackDef := ts.createDummyRack(2)

	r := newRack(context.Background(), ts.rackName(), rackDef, ts.timers)
	require.NotNil(r)
	assert.Equal("AwaitingStart", r.sm.Current.Name())

	p := r.pdu
	require.NotNil(p)

	assert.Equal(2, len(p.cables))

	assert.Equal("working", p.sm.Current.Name())

	for _, c := range p.cables {
		assert.False(c.on)
		assert.False(c.faulted)
	}
}

func (ts *PduTestSuite) TestBadPowerTarget() {
	require := ts.Require()
	assert := ts.Assert()

	ctx, r := ts.createAndStartRack(context.Background(), 2, true, true)

	rsp := make(chan *sm.Response)

	badMsg := newSetPower(ctx, newTargetTor(ts.rackName()), common.TickFromContext(ctx), false, rsp)

	r.Receive(badMsg)

	res := ts.completeWithin(rsp, time.Duration(1)*time.Second)
	require.NotNil(res)

	assert.Error(res.Err)
	assert.Equal(ErrInvalidTarget, res.Err)
	assert.Equal(common.TickFromContext(ctx), res.At)
	assert.Nil(res.Msg)

	assert.Equal("working", r.pdu.sm.Current.Name())

	for _, c := range r.pdu.cables {
		assert.True(c.on)
	}
}

func (ts *PduTestSuite) TestPowerOffPdu() {
	require := ts.Require()
	assert := ts.Assert()

	ctx, r := ts.createAndStartRack(context.Background(), 2, true, true)

	rsp := make(chan *sm.Response)

	msg := newSetPower(ctx, newTargetPdu(ts.rackName()), common.TickFromContext(ctx), false, rsp)

	r.Receive(msg)

	res := ts.completeWithin(rsp, time.Duration(1)*time.Second)
	require.NotNil(res)
	require.Error(res.Err)
	assert.Equal(ErrRepairMessageDropped, res.Err)
	assert.Equal(common.TickFromContext(ctx), res.At)
	assert.Nil(res.Msg)

	for _, c := range r.pdu.cables {
		assert.False(c.on)
	}

	assert.Equal("off", r.pdu.sm.Current.Name())
}

func (ts *PduTestSuite) TestPowerOffPduTooLate() {
	require := ts.Require()
	assert := ts.Assert()
	startTime := int64(1)

	ctx, r := ts.createAndStartRack(context.Background(), 2, true, true)

	rsp := make(chan *sm.Response)

	msg := newSetPower(ctx, newTargetPdu(ts.rackName()), startTime-1, false, rsp)

	r.Receive(msg)

	res := ts.completeWithin(rsp, time.Duration(1)*time.Second)
	require.NotNil(res)
	require.Error(res.Err)
	assert.Equal(ErrRepairMessageDropped, res.Err)

	assert.Equal(common.TickFromContext(ctx), res.At)
	assert.Nil(res.Msg)

	for _, c := range r.pdu.cables {
		assert.True(c.on)
	}

	assert.Equal("working", r.pdu.sm.Current.Name())
}

func (ts *PduTestSuite) TestPowerOnPdu() {
	require := ts.Require()
	assert := ts.Assert()

	ctx, r := ts.createAndStartRack(context.Background(), 2, false, true)

	rsp := make(chan *sm.Response)

	msg := newSetPower(ctx, newTargetPdu(ts.rackName()), common.TickFromContext(ctx), true, rsp)

	r.Receive(msg)

	res := ts.completeWithin(rsp, time.Duration(1)*time.Second)
	require.NotNil(res)
	require.Error(res.Err)
	assert.Equal(ErrRepairMessageDropped, res.Err)
	assert.Equal(common.TickFromContext(ctx), res.At)
	assert.Nil(res.Msg)

	for _, c := range r.pdu.cables {
		assert.False(c.on)
	}

	assert.Equal("working", r.pdu.sm.Current.Name())
}

func (ts *PduTestSuite) TestPowerOnBlade() {
	require := ts.Require()
	assert := ts.Assert()

	ctx, r := ts.createAndStartRack(context.Background(), 2, false, true)

	rsp := make(chan *sm.Response)

	msg := newSetPower(ctx, newTargetBlade(ts.rackName(), 0), common.TickFromContext(ctx), true, rsp)

	r.Receive(msg)

	res := ts.completeWithin(rsp, time.Duration(1)*time.Second)
	require.NotNil(res)
	assert.NoError(res.Err)

	assert.Equal(common.TickFromContext(ctx), res.At)
	assert.Equal(common.TickFromContext(ctx), r.pdu.sm.Guard)
	assert.Equal(common.TickFromContext(ctx), r.pdu.cables[0].Guard)
	assert.True(r.pdu.cables[0].on)
	assert.False(r.pdu.cables[0].faulted)

	assert.Equal("working", r.pdu.sm.Current.Name())
}

func (ts *PduTestSuite) TestPowerOnBladeBadID() {
	require := ts.Require()
	assert := ts.Assert()

	ctx, r := ts.createAndStartRack(context.Background(), 2, true, true)

	rsp := make(chan *sm.Response)

	msg := newSetPower(ctx, newTargetBlade(ts.rackName(), 9), common.TickFromContext(ctx), true, rsp)

	r.Receive(msg)

	res := ts.completeWithin(rsp, time.Duration(1)*time.Second)
	require.NotNil(res)
	assert.Error(res.Err)
	assert.Equal(ErrInvalidTarget, res.Err)

	assert.Equal(common.TickFromContext(ctx), res.At)
	assert.Less(r.pdu.sm.Guard, msg.guard)

	assert.Equal("working", r.pdu.sm.Current.Name())
}

func (ts *PduTestSuite) TestPowerOnBladeWhileOn() {
	require := ts.Require()
	assert := ts.Assert()

	ctx, r := ts.createAndStartRack(context.Background(), 2, true, true)

	rsp := make(chan *sm.Response)

	msg := newSetPower(ctx, newTargetBlade(ts.rackName(), 0), common.TickFromContext(ctx), true, rsp)

	r.Receive(msg)

	res := ts.completeWithin(rsp, time.Duration(1)*time.Second)

	require.NotNil(res)
	require.Error(res.Err)
	assert.Equal(ErrNoOperation, res.Err)

	assert.Equal(common.TickFromContext(ctx), res.At)
	assert.Equal(common.TickFromContext(ctx), r.pdu.sm.Guard)
	assert.Equal(common.TickFromContext(ctx), r.pdu.cables[0].Guard)

	assert.True(r.pdu.cables[0].on)
	assert.False(r.pdu.cables[0].faulted)

	assert.Equal("working", r.pdu.sm.Current.Name())
}

func (ts *PduTestSuite) TestPowerOnBladeTooLate() {
	require := ts.Require()
	assert := ts.Assert()

	ctx := ts.advance(context.Background())
	commandTime := common.TickFromContext(ctx)

	ctx, r := ts.createAndStartRack(ctx, 2, false, true)

	rsp := make(chan *sm.Response)

	msg := newSetPower(ctx, newTargetBlade(ts.rackName(), 0), commandTime, true, rsp)

	r.Receive(msg)

	res := ts.completeWithin(rsp, time.Duration(1)*time.Second)
	require.NotNil(res)
	require.Error(res.Err)
	assert.Equal(ErrRepairMessageDropped, res.Err)
	assert.Equal(common.TickFromContext(ctx), res.At)
	assert.Nil(res.Msg)

	assert.Less(r.pdu.sm.Guard, common.TickFromContext(ctx))
	assert.Less(commandTime, r.pdu.cables[0].Guard)
	assert.False(r.pdu.cables[0].on)

	assert.Equal("working", r.pdu.sm.Current.Name())
}

func (ts *PduTestSuite) TestStuckCable() {
	require := ts.Require()
	assert := ts.Assert()

	ctx, r := ts.createAndStartRack(context.Background(), 2, false, true)
	r.pdu.cables[0].faulted = true

	rsp := make(chan *sm.Response)

	commandTime := common.TickFromContext(ctx)
	msg := newSetPower(ctx, newTargetBlade(ts.rackName(), 0), commandTime, true, rsp)

	r.Receive(msg)

	res := ts.completeWithin(rsp, time.Duration(1)*time.Second)
	require.NotNil(res)
	assert.Error(res.Err)
	assert.Equal(ErrCableStuck, res.Err)
	assert.Equal(common.TickFromContext(ctx), res.At)
	assert.Nil(res.Msg)

	assert.Less(r.pdu.sm.Guard, commandTime)
	assert.Less(r.pdu.cables[0].Guard, commandTime)
	assert.False(r.pdu.cables[0].on)
	assert.Equal(true, r.pdu.cables[0].faulted)

	assert.Equal("working", r.pdu.sm.Current.Name())
}

func (ts *PduTestSuite) TestStuckCablePduOff() {
	require := ts.Require()
	assert := ts.Assert()

	ctx, r := ts.createAndStartRack(context.Background(), 2, true, true)

	rsp := make(chan *sm.Response)

	msg := newSetPower(ctx, newTargetPdu(ts.rackName()), common.TickFromContext(ctx), false, rsp)

	r.Receive(msg)

	res := ts.completeWithin(rsp, time.Duration(1)*time.Second)
	require.NotNil(res)
	require.Error(res.Err)
	assert.Equal(ErrRepairMessageDropped, res.Err)
	assert.Equal(common.TickFromContext(ctx), res.At)
	assert.Nil(res.Msg)

	for _, c := range r.pdu.cables {
		assert.False(c.on)
	}

	assert.Equal("off", r.pdu.sm.Current.Name())
}

func TestPduTestSuite(t *testing.T) {
	suite.Run(t, new(PduTestSuite))
}
