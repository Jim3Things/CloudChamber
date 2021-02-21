package inventory

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/Jim3Things/CloudChamber/simulation/internal/common"
	"github.com/Jim3Things/CloudChamber/simulation/internal/services/inventory/messages"
	"github.com/Jim3Things/CloudChamber/simulation/internal/sm"
	"github.com/Jim3Things/CloudChamber/simulation/pkg/errors"
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
	assert.Equal(rackAwaitingStartState, r.sm.CurrentIndex)

	p := r.pdu
	require.NotNil(p)

	assert.Equal(2, len(p.cables))

	assert.Equal(pduWorkingState, p.sm.CurrentIndex)

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

	badMsg := messages.NewSetPower(
		ctx,
		messages.NewTargetTor(ts.rackName()),
		common.TickFromContext(ctx),
		false,
		rsp)

	r.Receive(badMsg)

	res, ok := ts.completeWithin(rsp, time.Second)
	require.True(ok)
	require.NotNil(res)

	assert.Error(res.Err)
	assert.Equal(errors.ErrInvalidTarget, res.Err)
	assert.Equal(common.TickFromContext(ctx), res.At)
	assert.Nil(res.Msg)

	require.Eventually(func() bool {
		return r.pdu.sm.CurrentIndex == pduWorkingState
	}, time.Second, 10*time.Millisecond,
	"state is %v", r.pdu.sm.CurrentIndex)

	for _, c := range r.pdu.cables {
		assert.True(c.on)
	}
}

func (ts *PduTestSuite) TestPowerOffPdu() {
	require := ts.Require()
	assert := ts.Assert()

	ctx, r := ts.createAndStartRack(context.Background(), 2, true, true)

	rsp := make(chan *sm.Response)

	msg := messages.NewSetPower(
		ctx,
		messages.NewTargetPdu(ts.rackName()),
		common.TickFromContext(ctx),
		false,
		rsp)

	r.Receive(msg)

	res, ok := ts.completeWithin(rsp, time.Second)
	require.True(ok)
	require.Nil(res)

	for _, c := range r.pdu.cables {
		assert.False(c.on)
	}

	require.Eventually(func() bool {
		return r.pdu.sm.CurrentIndex == pduOffState
	}, time.Second, 10*time.Millisecond,
	"state is %v", r.pdu.sm.CurrentIndex)
}

func (ts *PduTestSuite) TestOffGetStatus() {
	require := ts.Require()
	assert := ts.Assert()

	ctx, r := ts.createAndStartRack(context.Background(), 2, true, true)

	rsp := make(chan *sm.Response)

	tick := common.TickFromContext(ctx)

	msg := messages.NewSetPower(
		ctx,
		messages.NewTargetPdu(ts.rackName()),
		tick,
		false,
		rsp)

	r.Receive(msg)

	res, ok := ts.completeWithin(rsp, time.Second)
	require.True(ok)
	require.Nil(res)

	require.Eventually(func() bool {
		return r.pdu.sm.CurrentIndex == pduOffState
	}, time.Second, 10*time.Millisecond)

	for _, c := range r.pdu.cables {
		assert.False(c.on)
	}

	rsp2 := make(chan *sm.Response)

	msg2 := messages.NewGetStatus(
		ctx,
		messages.NewTargetPdu(ts.rackName()),
		common.TickFromContext(ctx),
		rsp2)

	r.Receive(msg2)

	res2, ok := ts.completeWithin(rsp2, time.Second)
	require.True(ok)
	require.NotNil(res2)
	require.NoError(res2.Err)
	assert.Equal(common.TickFromContext(ctx), res2.At)
	require.NotNil(res2.Msg)

	status, ok := res2.Msg.(*messages.PduStatus)
	require.True(ok)

	assert.Equal(pduOffState, status.State)
	assert.Equal(tick, status.EnteredAt)

	for _, c := range status.Cables {
		assert.False(c.On)
		assert.False(c.Faulted)
	}

	require.True(ok, "state is %v", r.pdu.sm.CurrentIndex)
}

func (ts *PduTestSuite) TestPowerOffPduTooLate() {
	require := ts.Require()
	assert := ts.Assert()
	startTime := int64(1)

	ctx, r := ts.createAndStartRack(context.Background(), 2, true, true)
	require.Eventually(func() bool {
		return r.pdu.sm.CurrentIndex == pduWorkingState
	}, time.Second, 10*time.Millisecond,
	"state is %v", r.pdu.sm.CurrentIndex)

	rsp := make(chan *sm.Response)

	msg := messages.NewSetPower(
		ctx,
		messages.NewTargetPdu(ts.rackName()),
		startTime-1,
		false,
		rsp)

	r.Receive(msg)

	res, ok := ts.completeWithin(rsp, time.Second)
	require.True(ok)
	require.Nil(res)

	for _, c := range r.pdu.cables {
		assert.True(c.on)
	}

	// Verify that it did not change - should never need to wait for this.
	assert.Equal(pduWorkingState, r.pdu.sm.CurrentIndex)
}

func (ts *PduTestSuite) TestPowerOnPdu() {
	require := ts.Require()
	assert := ts.Assert()

	ctx, r := ts.createAndStartRack(context.Background(), 2, false, true)
	require.Eventually(func() bool {
		return r.pdu.sm.CurrentIndex == pduWorkingState
	}, time.Second, 10*time.Millisecond, "state is %v", r.pdu.sm.CurrentIndex)

	rsp := make(chan *sm.Response)

	msg := messages.NewSetPower(
		ctx,
		messages.NewTargetPdu(ts.rackName()),
		common.TickFromContext(ctx),
		true,
		rsp)

	r.Receive(msg)

	res, ok := ts.completeWithin(rsp, time.Second)
	require.True(ok)
	require.Nil(res)

	for _, c := range r.pdu.cables {
		assert.False(c.on)
	}

	// Verify that it did not change - should never need to wait for this.
	assert.Equal(pduWorkingState, r.pdu.sm.CurrentIndex)
}

func (ts *PduTestSuite) TestWorkingGetStatus() {
	require := ts.Require()
	assert := ts.Assert()

	ctx, r := ts.createAndStartRack(context.Background(), 2, false, true)
	require.Eventually(func() bool {
		return r.pdu.sm.CurrentIndex == pduWorkingState
	}, time.Second, 10*time.Millisecond,
	"state is %v", r.pdu.sm.CurrentIndex)

	rsp := make(chan *sm.Response)

	msg := messages.NewGetStatus(
		ctx,
		messages.NewTargetPdu(ts.rackName()),
		common.TickFromContext(ctx),
		rsp)

	r.Receive(msg)

	res, ok := ts.completeWithin(rsp, time.Second)
	require.True(ok)
	require.NotNil(res)
	require.NoError(res.Err)
	assert.Equal(common.TickFromContext(ctx), res.At)
	require.NotNil(res.Msg)

	status, ok := res.Msg.(*messages.PduStatus)
	require.True(ok)

	assert.Equal(pduWorkingState, status.State)
	assert.Equal(int64(0), status.EnteredAt)

	for _, c := range status.Cables {
		assert.False(c.On)
		assert.False(c.Faulted)
	}

	// Verify that it did not change - should never need to wait for this.
	assert.Equal(pduWorkingState, r.pdu.sm.CurrentIndex)
}

func (ts *PduTestSuite) TestPowerOnBlade() {
	require := ts.Require()
	assert := ts.Assert()

	ctx, r := ts.createAndStartRack(context.Background(), 2, false, true)

	rsp := make(chan *sm.Response)

	msg := messages.NewSetPower(
		ctx,
		messages.NewTargetBlade(ts.rackName(), 0),
		common.TickFromContext(ctx),
		true,
		rsp)

	r.Receive(msg)

	res, ok := ts.completeWithin(rsp, time.Second)
	require.True(ok)
	require.NotNil(res)
	assert.NoError(res.Err)

	assert.Equal(common.TickFromContext(ctx), res.At)
	assert.Equal(common.TickFromContext(ctx), r.pdu.sm.Guard)
	assert.Equal(common.TickFromContext(ctx), r.pdu.cables[0].Guard)
	assert.True(r.pdu.cables[0].on)
	assert.False(r.pdu.cables[0].faulted)

	// SetPower above will have synchronized enough that the state should be
	// correct without any waiting
	assert.Equal(pduWorkingState, r.pdu.sm.CurrentIndex)
}

func (ts *PduTestSuite) TestPowerOnBladeBadID() {
	require := ts.Require()
	assert := ts.Assert()

	ctx, r := ts.createAndStartRack(context.Background(), 2, true, true)

	rsp := make(chan *sm.Response)

	msg := messages.NewSetPower(
		ctx,
		messages.NewTargetBlade(ts.rackName(), 9),
		common.TickFromContext(ctx),
		true,
		rsp)

	r.Receive(msg)

	res, ok := ts.completeWithin(rsp, time.Second)
	require.True(ok)
	require.NotNil(res)
	assert.Error(res.Err)
	assert.Equal(errors.ErrInvalidTarget, res.Err)

	assert.Equal(common.TickFromContext(ctx), res.At)
	assert.Less(r.pdu.sm.Guard, msg.Guard)

	// SetPower above will have synchronized enough that the state should be
	// correct without any waiting
	assert.Equal(pduWorkingState, r.pdu.sm.CurrentIndex)
}

func (ts *PduTestSuite) TestPowerOnBladeWhileOn() {
	require := ts.Require()
	assert := ts.Assert()

	ctx, r := ts.createAndStartRack(context.Background(), 2, true, true)

	rsp := make(chan *sm.Response)

	msg := messages.NewSetPower(
		ctx,
		messages.NewTargetBlade(ts.rackName(), 0),
		common.TickFromContext(ctx),
		true,
		rsp)

	r.Receive(msg)

	res, ok := ts.completeWithin(rsp, time.Second)
	require.True(ok)
	require.NotNil(res)
	require.Error(res.Err)
	assert.Equal(errors.ErrNoOperation, res.Err)

	assert.Equal(common.TickFromContext(ctx), res.At)
	assert.Equal(common.TickFromContext(ctx), r.pdu.sm.Guard)
	assert.Equal(common.TickFromContext(ctx), r.pdu.cables[0].Guard)

	assert.True(r.pdu.cables[0].on)
	assert.False(r.pdu.cables[0].faulted)

	// SetPower above will have synchronized enough that the state should be
	// correct without any waiting
	assert.Equal(pduWorkingState, r.pdu.sm.CurrentIndex)
}

func (ts *PduTestSuite) TestPowerOnBladeTooLate() {
	require := ts.Require()
	assert := ts.Assert()

	ctx := ts.advance(context.Background())
	commandTime := common.TickFromContext(ctx)

	ctx, r := ts.createAndStartRack(ctx, 2, false, true)

	rsp := make(chan *sm.Response)

	msg := messages.NewSetPower(
		ctx,
		messages.NewTargetBlade(ts.rackName(), 0),
		commandTime,
		true,
		rsp)

	r.Receive(msg)

	res, ok := ts.completeWithin(rsp, time.Second)
	require.True(ok)
	require.NotNil(res)
	require.Error(errors.ErrInventoryChangeTooLate(commandTime), res.Err)

	assert.Less(r.pdu.sm.Guard, common.TickFromContext(ctx))
	assert.Less(commandTime, r.pdu.cables[0].Guard)
	assert.False(r.pdu.cables[0].on)

	// SetPower above will have synchronized enough that the state should be
	// correct without any waiting
	assert.Equal(pduWorkingState, r.pdu.sm.CurrentIndex)
}

func (ts *PduTestSuite) TestStuckCable() {
	require := ts.Require()
	assert := ts.Assert()

	ctx, r := ts.createAndStartRack(context.Background(), 2, false, true)
	r.pdu.cables[0].faulted = true

	rsp := make(chan *sm.Response)

	commandTime := common.TickFromContext(ctx)
	msg := messages.NewSetPower(
		ctx,
		messages.NewTargetBlade(ts.rackName(), 0),
		commandTime,
		true,
		rsp)

	r.Receive(msg)

	res, ok := ts.completeWithin(rsp, time.Second)
	require.True(ok)
	require.NotNil(res)
	assert.Error(res.Err)
	assert.Equal(errors.ErrCableStuck, res.Err)
	assert.Equal(common.TickFromContext(ctx), res.At)
	assert.Nil(res.Msg)

	assert.Less(r.pdu.sm.Guard, commandTime)
	assert.Less(r.pdu.cables[0].Guard, commandTime)
	assert.False(r.pdu.cables[0].on)
	assert.Equal(true, r.pdu.cables[0].faulted)

	// SetPower above will have synchronized enough that the state should be
	// correct without any waiting
	assert.Equal(pduWorkingState, r.pdu.sm.CurrentIndex)
}

func (ts *PduTestSuite) TestStuckCablePduOff() {
	require := ts.Require()
	assert := ts.Assert()

	ctx, r := ts.createAndStartRack(context.Background(), 2, true, true)

	rsp := make(chan *sm.Response)

	msg := messages.NewSetPower(
		ctx,
		messages.NewTargetPdu(ts.rackName()),
		common.TickFromContext(ctx),
		false,
		rsp)

	r.Receive(msg)

	res, ok := ts.completeWithin(rsp, time.Second)
	require.True(ok)
	require.Nil(res)

	for _, c := range r.pdu.cables {
		assert.False(c.on)
	}

	require.Eventually(func() bool {
		return r.pdu.sm.CurrentIndex == pduOffState
	}, time.Second, 10*time.Millisecond,
	"state is %v", r.pdu.sm.CurrentIndex)
}

func TestPduTestSuite(t *testing.T) {
	suite.Run(t, new(PduTestSuite))
}
