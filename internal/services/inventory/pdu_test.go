package inventory

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/Jim3Things/CloudChamber/internal/common"
	"github.com/Jim3Things/CloudChamber/internal/sm"
	"github.com/Jim3Things/CloudChamber/internal/tracing"
	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters"
	ct "github.com/Jim3Things/CloudChamber/pkg/protos/common"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/inventory"
	"github.com/Jim3Things/CloudChamber/pkg/protos/services"
)

type PduTestSuite struct {
	suite.Suite

	utf *exporters.Exporter
}

func (ts *PduTestSuite) SetupSuite() {
	ts.utf = exporters.NewExporter(exporters.NewUTForwarder())
	exporters.ConnectToProvider(ts.utf)
}

func (ts *PduTestSuite) SetupTest() {
	_ = ts.utf.Open(ts.T())
}

func (ts *PduTestSuite) TearDownTest() {
	ts.utf.Close()
}

func createDummyRack(bladeCount int) *pb.ExternalRack {
	rackDef := &pb.ExternalRack{
		Pdu:    &pb.ExternalPdu{},
		Tor:    &pb.ExternalTor{},
		Blades: make(map[int64]*ct.BladeCapacity),
	}

	for i := 0; i < bladeCount; i++ {
		rackDef.Blades[int64(i)] = &ct.BladeCapacity{}
	}

	return rackDef
}

func (ts *PduTestSuite) completeWithin(ch <-chan *sm.Response, delay time.Duration) *sm.Response {
	select {
	case res := <-ch:
		return res
	case <-time.After(delay):
		return nil
	}
}

func execute(ctx context.Context, msg *sm.Envelope, action func(ctx2 context.Context, envelope *sm.Envelope)) {
	go func(tick int64, msg *sm.Envelope) {
		c2, s := tracing.StartSpan(context.Background(),
			tracing.WithName("Executing simulated inventory operation"),
			tracing.WithNewRoot(),
			tracing.WithLink(msg.Span, msg.Link))

		c2 = common.ContextWithTick(c2, tick)

		action(c2, msg)

		s.End()
	}(common.TickFromContext(ctx), msg)
}

func (ts *PduTestSuite) TestCreatePdu() {
	require := ts.Require()
	assert := ts.Assert()

	rackDef := createDummyRack(2)

	r := newRack(context.Background(), rackDef)
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

	rackDef := createDummyRack(2)

	r := newRack(ctx, rackDef)
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
			Target: &services.InventoryAddress{
				Rack:    "",
				Element: &services.InventoryAddress_Tor{},
			},
			After: &ct.Timestamp{Ticks: common.TickFromContext(ctx)},
			Action: &services.InventoryRepairMsg_Power{
				Power: false,
			},
		},
		rsp)

	span.End()

	execute(ctx, badMsg, r.pdu.Receive)

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

	rackDef := createDummyRack(2)

	r := newRack(ctx, rackDef)
	ctx = common.ContextWithTick(ctx, 2)

	ctx, span := tracing.StartSpan(
		ctx,
		tracing.WithName("test powering off a PDU"))

	rsp := make(chan *sm.Response)

	msg := sm.NewEnvelope(
		ctx,
		&services.InventoryRepairMsg{
			Target: &services.InventoryAddress{
				Rack:    "",
				Element: &services.InventoryAddress_Pdu{},
			},
			After: &ct.Timestamp{Ticks: common.TickFromContext(ctx)},
			Action: &services.InventoryRepairMsg_Power{
				Power: false,
			},
		},
		rsp)

	span.End()

	execute(ctx, msg, r.pdu.Receive)

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

	rackDef := createDummyRack(2)

	r := newRack(ctx, rackDef)
	ctx = common.ContextWithTick(ctx, 2)

	ctx, span := tracing.StartSpan(
		ctx,
		tracing.WithName("test powering on a PDU"))

	rsp := make(chan *sm.Response)

	msg := sm.NewEnvelope(
		ctx,
		&services.InventoryRepairMsg{
			Target: &services.InventoryAddress{
				Rack:    "",
				Element: &services.InventoryAddress_Pdu{},
			},
			After: &ct.Timestamp{Ticks: common.TickFromContext(ctx)},
			Action: &services.InventoryRepairMsg_Power{
				Power: true,
			},
		},
		rsp)

	span.End()

	execute(ctx, msg, r.pdu.Receive)

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

	rackDef := createDummyRack(2)

	r := newRack(ctx, rackDef)
	ctx = common.ContextWithTick(ctx, 2)

	ctx, span := tracing.StartSpan(
		ctx,
		tracing.WithName("test powering on a blade"))

	rsp := make(chan *sm.Response)

	msg := sm.NewEnvelope(
		ctx,
		&services.InventoryRepairMsg{
			Target: &services.InventoryAddress{
				Rack: "",
				Element: &services.InventoryAddress_BladeId{
					BladeId: 0,
				},
			},
			After: &ct.Timestamp{Ticks: common.TickFromContext(ctx)},
			Action: &services.InventoryRepairMsg_Power{
				Power: true,
			},
		},
		rsp)

	span.End()

	execute(ctx, msg, r.pdu.Receive)

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

	rackDef := createDummyRack(2)

	r := newRack(ctx, rackDef)
	ctx = common.ContextWithTick(ctx, 2)

	ctx, span := tracing.StartSpan(
		ctx,
		tracing.WithName("test powering on a blade (too late)"))

	rsp := make(chan *sm.Response)

	msg := sm.NewEnvelope(
		ctx,
		&services.InventoryRepairMsg{
			Target: &services.InventoryAddress{
				Rack: "",
				Element: &services.InventoryAddress_BladeId{
					BladeId: 0,
				},
			},
			After: &ct.Timestamp{Ticks: startTime - 1},
			Action: &services.InventoryRepairMsg_Power{
				Power: true,
			},
		},
		rsp)

	span.End()

	execute(ctx, msg, r.pdu.Receive)

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

	rackDef := createDummyRack(2)

	r := newRack(ctx, rackDef)

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
			Target: &services.InventoryAddress{
				Rack: "",
				Element: &services.InventoryAddress_BladeId{
					BladeId: 0,
				},
			},
			After: &ct.Timestamp{Ticks: common.TickFromContext(ctx)},
			Action: &services.InventoryRepairMsg_Power{
				Power: true,
			},
		},
		rsp)

	span.End()

	execute(ctx, msg, r.pdu.Receive)

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

	rackDef := createDummyRack(2)
	r := newRack(ctx, rackDef)

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
			Target: &services.InventoryAddress{
				Rack:    "",
				Element: &services.InventoryAddress_Pdu{},
			},
			After: &ct.Timestamp{Ticks: common.TickFromContext(ctx)},
			Action: &services.InventoryRepairMsg_Power{
				Power: false,
			},
		},
		rsp)

	span.End()

	execute(ctx, msg, r.pdu.Receive)

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
