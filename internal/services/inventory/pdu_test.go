package inventory

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/Jim3Things/CloudChamber/internal/common"
	"github.com/Jim3Things/CloudChamber/internal/sm"
	"github.com/Jim3Things/CloudChamber/internal/tracing"
	ct "github.com/Jim3Things/CloudChamber/pkg/protos/common"
	"github.com/Jim3Things/CloudChamber/pkg/protos/services"
)

type PduTestSuite struct {
	testSuiteCore
}

func (ts *PduTestSuite) TestCreatePdu() {
	require := ts.Require()
	assert := ts.Assert()

	rackDef := ts.createDummyRack(2)

	r := newRack(context.Background(), ts.rackName(), rackDef)
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

	ctx := common.ContextWithTick(context.Background(), 1)

	rackDef := ts.createDummyRack(2)

	r := newRack(ctx, ts.rackName(), rackDef)
	ctx = common.ContextWithTick(ctx, 2)

	ctx, span := tracing.StartSpan(
		ctx,
		tracing.WithName("test bad power target"))

	for i := range r.pdu.cables {
		r.pdu.cables[i] = newCable(true, false, 0)
	}

	rsp := make(chan *sm.Response)

	badMsg := sm.NewEnvelope(
		ctx,
		&services.InventoryRepairMsg{
			Target: ts.torTarget(),
			After: &ct.Timestamp{Ticks: common.TickFromContext(ctx)},
			Action: &services.InventoryRepairMsg_Power{
				Power: false,
			},
		},
		rsp)

	span.End()

	ts.execute(ctx, badMsg, r.pdu.Receive)

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

	ctx := common.ContextWithTick(context.Background(), 1)

	rackDef := ts.createDummyRack(2)

	r := newRack(ctx, ts.rackName(), rackDef)
	ctx = common.ContextWithTick(ctx, 2)

	ctx, span := tracing.StartSpan(
		ctx,
		tracing.WithName("test powering off a PDU"))

	rsp := make(chan *sm.Response)

	msg := sm.NewEnvelope(
		ctx,
		&services.InventoryRepairMsg{
			Target: ts.pduTarget(),
			After: &ct.Timestamp{Ticks: common.TickFromContext(ctx)},
			Action: &services.InventoryRepairMsg_Power{
				Power: false,
			},
		},
		rsp)

	span.End()

	ts.execute(ctx, msg, r.pdu.Receive)

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

func (ts *PduTestSuite) TestPowerOnPdu() {
	require := ts.Require()
	assert := ts.Assert()

	ctx := common.ContextWithTick(context.Background(), 1)

	rackDef := ts.createDummyRack(2)

	r := newRack(ctx, ts.rackName(), rackDef)
	ctx = common.ContextWithTick(ctx, 2)

	ctx, span := tracing.StartSpan(
		ctx,
		tracing.WithName("test powering on a PDU"))

	rsp := make(chan *sm.Response)

	msg := sm.NewEnvelope(
		ctx,
		&services.InventoryRepairMsg{
			Target: ts.pduTarget(),
			After: &ct.Timestamp{Ticks: common.TickFromContext(ctx)},
			Action: &services.InventoryRepairMsg_Power{
				Power: true,
			},
		},
		rsp)

	span.End()

	ts.execute(ctx, msg, r.pdu.Receive)

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

	ctx := common.ContextWithTick(context.Background(), 1)

	rackDef := ts.createDummyRack(2)

	r := newRack(ctx, ts.rackName(), rackDef)
	ctx = common.ContextWithTick(ctx, 2)

	ctx, span := tracing.StartSpan(
		ctx,
		tracing.WithName("test powering on a blade"))

	rsp := make(chan *sm.Response)

	msg := sm.NewEnvelope(
		ctx,
		&services.InventoryRepairMsg{
			Target: ts.bladeTarget(0),
			After: &ct.Timestamp{Ticks: common.TickFromContext(ctx)},
			Action: &services.InventoryRepairMsg_Power{
				Power: true,
			},
		},
		rsp)

	span.End()

	ts.execute(ctx, msg, r.pdu.Receive)

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

func (ts *PduTestSuite) TestPowerOnBladeTooLate() {
	require := ts.Require()
	assert := ts.Assert()

	startTime := int64(1)
	ctx := common.ContextWithTick(context.Background(), startTime)

	rackDef := ts.createDummyRack(2)

	r := newRack(ctx, ts.rackName(), rackDef)
	ctx = common.ContextWithTick(ctx, 2)

	ctx, span := tracing.StartSpan(
		ctx,
		tracing.WithName("test powering on a blade (too late)"))

	rsp := make(chan *sm.Response)

	msg := sm.NewEnvelope(
		ctx,
		&services.InventoryRepairMsg{
			Target: ts.bladeTarget(0),
			After: &ct.Timestamp{Ticks: startTime - 1},
			Action: &services.InventoryRepairMsg_Power{
				Power: true,
			},
		},
		rsp)

	span.End()

	ts.execute(ctx, msg, r.pdu.Receive)

	res := ts.completeWithin(rsp, time.Duration(1)*time.Second)
	require.NotNil(res)
	require.Error(res.Err)
	assert.Equal(ErrRepairMessageDropped, res.Err)
	assert.Equal(common.TickFromContext(ctx), res.At)
	assert.Nil(res.Msg)

	assert.Equal(startTime, r.pdu.sm.Guard)
	assert.Equal(startTime, r.pdu.cables[0].Guard)
	assert.False(r.pdu.cables[0].on)

	assert.Equal("working", r.pdu.sm.Current.Name())
}

func (ts *PduTestSuite) TestStuckCable() {
	require := ts.Require()
	assert := ts.Assert()

	ctx := common.ContextWithTick(context.Background(), 1)

	rackDef := ts.createDummyRack(2)

	r := newRack(ctx, ts.rackName(), rackDef)

	startTime := common.TickFromContext(ctx)
	require.Nil(r.pdu.cables[0].fault(false, startTime, startTime))
	ctx = common.ContextWithTick(ctx, 2)

	ctx, span := tracing.StartSpan(
		ctx,
		tracing.WithName("test powering on a blade (stuck cable)"))

	rsp := make(chan *sm.Response)

	msg := sm.NewEnvelope(
		ctx,
		&services.InventoryRepairMsg{
			Target: ts.bladeTarget(0),
			After: &ct.Timestamp{Ticks: common.TickFromContext(ctx)},
			Action: &services.InventoryRepairMsg_Power{
				Power: true,
			},
		},
		rsp)

	span.End()

	ts.execute(ctx, msg, r.pdu.Receive)

	res := ts.completeWithin(rsp, time.Duration(1)*time.Second)
	require.NotNil(res)
	assert.Error(res.Err)
	assert.Equal(ErrCableStuck, res.Err)
	assert.Equal(common.TickFromContext(ctx), res.At)
	assert.Nil(res.Msg)

	assert.Equal(startTime, r.pdu.sm.Guard)
	assert.Equal(startTime, r.pdu.cables[0].Guard)
	assert.False(r.pdu.cables[0].on)
	assert.Equal(true, r.pdu.cables[0].faulted)

	assert.Equal("working", r.pdu.sm.Current.Name())
}

func (ts *PduTestSuite) TestStuckCablePduOff() {
	require := ts.Require()
	assert := ts.Assert()

	ctx := common.ContextWithTick(context.Background(), 1)

	rackDef := ts.createDummyRack(2)
	r := newRack(ctx, ts.rackName(), rackDef)

	startTime := common.TickFromContext(ctx)
	require.Nil(r.pdu.cables[0].fault(true, startTime, startTime))
	ctx = common.ContextWithTick(ctx, 2)

	ctx, span := tracing.StartSpan(
		ctx,
		tracing.WithName("test powering off a pdu (stuck cable)"))

	rsp := make(chan *sm.Response)

	msg := sm.NewEnvelope(
		ctx,
		&services.InventoryRepairMsg{
			Target: ts.pduTarget(),
			After: &ct.Timestamp{Ticks: common.TickFromContext(ctx)},
			Action: &services.InventoryRepairMsg_Power{
				Power: false,
			},
		},
		rsp)

	span.End()

	ts.execute(ctx, msg, r.pdu.Receive)

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
