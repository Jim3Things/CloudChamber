package inventory

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/stretchr/testify/suite"

	"github.com/Jim3Things/CloudChamber/simulation/internal/common"
	"github.com/Jim3Things/CloudChamber/simulation/internal/services/inventory/messages"
	"github.com/Jim3Things/CloudChamber/simulation/internal/sm"
	"github.com/Jim3Things/CloudChamber/simulation/pkg/errors"
	pb "github.com/Jim3Things/CloudChamber/simulation/pkg/protos/inventory"
)

type PduTestSuite struct {
	testSuiteCore
}

func (ts *PduTestSuite) TestCreatePdu() {
	require := ts.Require()
	assert := ts.Assert()

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
	assert.Equal(pb.Actual_Rack_awaiting_start, r.sm.CurrentIndex)

	require.Len(r.pdus, 1)
	p := r.pdus[0]
	require.NotNil(p)

	assert.Equal(2, len(p.cables))

	assert.Equal(pb.PduState_working, p.sm.CurrentIndex)

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
		messages.NewTargetTor(ts.rackName(), 0, 0),
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

	p := r.pdus[0]

	require.Eventually(func() bool {
		return p.sm.CurrentIndex == pb.PduState_working
	}, time.Second, 10*time.Millisecond, "state is %v", p.sm.CurrentIndex)

	for _, c := range p.cables {
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
		messages.NewTargetPdu(ts.rackName(), 0, 0),
		common.TickFromContext(ctx),
		false,
		rsp)

	r.Receive(msg)

	res, ok := ts.completeWithin(rsp, time.Second)
	require.True(ok)
	require.Nil(res)

	p := r.pdus[0]

	for _, c := range p.cables {
		assert.False(c.on)
	}

	require.Eventually(func() bool {
		return p.sm.CurrentIndex == pb.PduState_off
	}, time.Second, 10*time.Millisecond, "state is %v", p.sm.CurrentIndex)
}

func (ts *PduTestSuite) TestOffGetStatus() {
	require := ts.Require()
	assert := ts.Assert()

	ctx, r := ts.createAndStartRack(context.Background(), 2, true, true)

	rsp := make(chan *sm.Response)
	target := messages.NewTargetPdu(ts.rackName(), 0, 0)

	tick := common.TickFromContext(ctx)

	msg := messages.NewSetPower(
		ctx,
		target,
		tick,
		false,
		rsp)

	r.Receive(msg)

	res, ok := ts.completeWithin(rsp, time.Second)
	require.True(ok)
	require.Nil(res)

	p := r.pdus[0]

	require.Eventually(func() bool {
		return p.sm.CurrentIndex == pb.PduState_off
	}, time.Second, 10*time.Millisecond, "state is %v", p.sm.CurrentIndex)

	for _, c := range p.cables {
		assert.False(c.on)
	}

	rsp2 := make(chan *sm.Response)

	msg2 := messages.NewGetStatus(
		ctx,
		messages.NewTargetPdu(ts.rackName(), 0, 0),
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

	assert.Equal(pb.PduState_off.String(), status.State)
	assert.Equal(tick, status.EnteredAt)

	for _, c := range status.Cables {
		assert.False(c.On)
		assert.False(c.Faulted)
	}

	require.True(ok, "state is %v", r.pdus[0].sm.CurrentIndex)
}

func (ts *PduTestSuite) TestPowerOffPduTooLate() {
	require := ts.Require()
	assert := ts.Assert()
	startTime := int64(1)

	ctx, r := ts.createAndStartRack(context.Background(), 2, true, true)

	p := r.pdus[0]

	require.Eventually(func() bool {
		return p.sm.CurrentIndex == pb.PduState_working
	}, time.Second, 10*time.Millisecond, "state is %v", p.sm.CurrentIndex)

	rsp := make(chan *sm.Response)

	msg := messages.NewSetPower(
		ctx,
		messages.NewTargetPdu(ts.rackName(), 0, 0),
		startTime-1,
		false,
		rsp)

	r.Receive(msg)

	res, ok := ts.completeWithin(rsp, time.Second)
	require.True(ok)
	require.Nil(res)

	for _, c := range p.cables {
		assert.True(c.on)
	}

	// Verify that it did not change - should never need to wait for this.
	assert.Equal(pb.PduState_working, r.pdus[0].sm.CurrentIndex)
}

func (ts *PduTestSuite) TestPowerOnPdu() {
	require := ts.Require()
	assert := ts.Assert()

	ctx, r := ts.createAndStartRack(context.Background(), 2, false, true)

	p := r.pdus[0]

	require.Eventually(func() bool {
		return p.sm.CurrentIndex == pb.PduState_working
	}, time.Second, 10*time.Millisecond, "state is %v", p.sm.CurrentIndex)

	rsp := make(chan *sm.Response)

	msg := messages.NewSetPower(
		ctx,
		messages.NewTargetPdu(ts.rackName(), 0, 0),
		common.TickFromContext(ctx),
		true,
		rsp)

	r.Receive(msg)

	res, ok := ts.completeWithin(rsp, time.Second)
	require.True(ok)
	require.Nil(res)

	for _, c := range p.cables {
		assert.False(c.on)
	}

	// Verify that it did not change - should never need to wait for this.
	assert.Equal(pb.PduState_working, p.sm.CurrentIndex)
}

func (ts *PduTestSuite) TestPowerOnPduPersistence() {
	require := ts.Require()
	assert := ts.Assert()

	ctx, r := ts.createAndStartRack(context.Background(), 2, false, true)

	p := r.pdus[0]

	require.Eventually(func() bool {
		return p.sm.CurrentIndex == pb.PduState_working
	}, time.Second, 10*time.Millisecond, "state is %v", p.sm.CurrentIndex)

	rsp := make(chan *sm.Response)

	msg := messages.NewSetPower(
		ctx,
		messages.NewTargetPdu(ts.rackName(), 0, 0),
		common.TickFromContext(ctx),
		true,
		rsp)

	r.Receive(msg)

	res, ok := ts.completeWithin(rsp, time.Second)
	require.True(ok)
	require.Nil(res)

	for _, c := range p.cables {
		assert.False(c.on)
	}

	// Verify that it did not change - should never need to wait for this.
	assert.Equal(pb.PduState_working, p.sm.CurrentIndex)

	saved, err := p.Save()
	require.NoError(err)

	m := jsonpb.Marshaler{}
	_, err = m.MarshalToString(saved)
	require.NoError(err)
}

func (ts *PduTestSuite) TestWorkingGetStatus() {
	require := ts.Require()
	assert := ts.Assert()

	ctx, r := ts.createAndStartRack(context.Background(), 2, false, true)

	p := r.pdus[0]

	require.Eventually(func() bool {
		return p.sm.CurrentIndex == pb.PduState_working
	}, time.Second, 10*time.Millisecond, "state is %v", p.sm.CurrentIndex)

	rsp := make(chan *sm.Response)

	msg := messages.NewGetStatus(
		ctx,
		messages.NewTargetPdu(ts.rackName(), 0, 0),
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

	assert.Equal(pb.PduState_working.String(), status.State)
	assert.Equal(int64(0), status.EnteredAt)

	for _, c := range status.Cables {
		assert.False(c.On)
		assert.False(c.Faulted)
	}

	// Verify that it did not change - should never need to wait for this.
	assert.Equal(pb.PduState_working, p.sm.CurrentIndex)
}

func (ts *PduTestSuite) TestPowerOnBlade() {
	require := ts.Require()
	assert := ts.Assert()

	ctx, r := ts.createAndStartRack(context.Background(), 2, false, true)

	p := r.pdus[0]

	rsp := make(chan *sm.Response)
	target := messages.NewTargetBlade(ts.rackName(), 0, 0)

	msg := messages.NewSetPower(
		ctx,
		target,
		common.TickFromContext(ctx),
		true,
		rsp)

	r.Receive(msg)

	res, ok := ts.completeWithin(rsp, time.Second)
	require.True(ok)
	require.NotNil(res)
	assert.NoError(res.Err)

	assert.Equal(common.TickFromContext(ctx), res.At)
	assert.Equal(common.TickFromContext(ctx), p.sm.Guard)

	c := p.cables[target.Key()]

	assert.Equal(common.TickFromContext(ctx), c.Guard)
	assert.True(c.on)
	assert.False(c.faulted)

	// SetPower above will have synchronized enough that the state should be
	// correct without any waiting
	assert.Equal(pb.PduState_working, p.sm.CurrentIndex)
}

func (ts *PduTestSuite) TestPowerOnBladeBadID() {
	require := ts.Require()
	assert := ts.Assert()

	ctx, r := ts.createAndStartRack(context.Background(), 2, true, true)

	p := r.pdus[0]

	rsp := make(chan *sm.Response)

	msg := messages.NewSetPower(
		ctx,
		messages.NewTargetBlade(ts.rackName(), 9, 0),
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
	assert.Less(p.sm.Guard, msg.Guard)

	// SetPower above will have synchronized enough that the state should be
	// correct without any waiting
	assert.Equal(pb.PduState_working, p.sm.CurrentIndex)
}

func (ts *PduTestSuite) TestPowerOnBladeWhileOn() {
	require := ts.Require()
	assert := ts.Assert()

	ctx, r := ts.createAndStartRack(context.Background(), 2, true, true)
	p := r.pdus[0]
	target := messages.NewTargetBlade(ts.rackName(), 0, 0)

	rsp := make(chan *sm.Response)

	msg := messages.NewSetPower(
		ctx,
		target,
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
	assert.Equal(common.TickFromContext(ctx), p.sm.Guard)

	c := p.cables[target.Key()]

	assert.Equal(common.TickFromContext(ctx), c.Guard)
	assert.True(c.on)
	assert.False(c.faulted)

	// SetPower above will have synchronized enough that the state should be
	// correct without any waiting
	assert.Equal(pb.PduState_working, p.sm.CurrentIndex)
}

func (ts *PduTestSuite) TestPowerOnBladeTooLate() {
	require := ts.Require()
	assert := ts.Assert()

	ctx := ts.advance(context.Background())
	commandTime := common.TickFromContext(ctx)

	ctx, r := ts.createAndStartRack(ctx, 2, false, true)

	p := r.pdus[0]
	target := messages.NewTargetBlade(ts.rackName(), 0, 0)

	rsp := make(chan *sm.Response)

	msg := messages.NewSetPower(
		ctx,
		target,
		commandTime,
		true,
		rsp)

	r.Receive(msg)

	res, ok := ts.completeWithin(rsp, time.Second)
	require.True(ok)
	require.NotNil(res)
	require.Error(errors.ErrInventoryChangeTooLate(commandTime), res.Err)

	assert.Less(p.sm.Guard, common.TickFromContext(ctx))

	c := p.cables[target.Key()]

	assert.Less(commandTime, c.Guard)
	assert.False(c.on)

	// SetPower above will have synchronized enough that the state should be
	// correct without any waiting
	assert.Equal(pb.PduState_working, p.sm.CurrentIndex)
}

func (ts *PduTestSuite) TestStuckCable() {
	require := ts.Require()
	assert := ts.Assert()

	ctx, r := ts.createAndStartRack(context.Background(), 2, false, true)
	p := r.pdus[0]
	target := messages.NewTargetBlade(ts.rackName(), 0, 0)

	c := p.cables[target.Key()]

	c.faulted = true

	rsp := make(chan *sm.Response)

	commandTime := common.TickFromContext(ctx)
	msg := messages.NewSetPower(
		ctx,
		target,
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

	assert.Less(p.sm.Guard, commandTime)

	c = p.cables[target.Key()]
	assert.Less(c.Guard, commandTime)
	assert.False(c.on)
	assert.True(c.faulted)

	// SetPower above will have synchronized enough that the state should be
	// correct without any waiting
	assert.Equal(pb.PduState_working, p.sm.CurrentIndex)
}

func (ts *PduTestSuite) TestStuckCablePduOff() {
	require := ts.Require()
	assert := ts.Assert()

	ctx, r := ts.createAndStartRack(context.Background(), 2, true, true)

	p := r.pdus[0]

	rsp := make(chan *sm.Response)

	msg := messages.NewSetPower(
		ctx,
		messages.NewTargetPdu(ts.rackName(), 0, 0),
		common.TickFromContext(ctx),
		false,
		rsp)

	r.Receive(msg)

	res, ok := ts.completeWithin(rsp, time.Second)
	require.True(ok)
	require.Nil(res)

	for _, c := range p.cables {
		assert.False(c.on)
	}

	require.Eventually(func() bool {
		return p.sm.CurrentIndex == pb.PduState_off
	}, time.Second, 10*time.Millisecond, "state is %v", p.sm.CurrentIndex)
}

func TestPduTestSuite(t *testing.T) {
	suite.Run(t, new(PduTestSuite))
}
