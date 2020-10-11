package inventory

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/Jim3Things/CloudChamber/internal/common"
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

func (ts *PduTestSuite) TestCreatePdu() {
	require := ts.Require()
	assert := ts.Assert()

	rackDef := createDummyRack(2)

	r := newRack(context.Background(), rackDef)
	require.NotNil(r)
	assert.Equal("working", r.sm.Current.Name())

	p := r.pdu
	require.NotNil(p)

	assert.Equal(2, len(p.cables))

	assert.Equal("working", p.sm.Current.Name())

	for _, cable := range p.cables {
		assert.False(cable.on)
	}
}

func (ts *PduTestSuite) TestBadPowerTarget() {
	require := ts.Require()
	assert := ts.Assert()

	ctx := common.ContextWithTick(context.Background(), 1)

	rackDef := createDummyRack(2)

	r := newRack(ctx, rackDef)
	ctx = common.ContextWithTick(ctx, 2)

	for i := range r.pdu.cables {
		r.pdu.cables[i] = cableState{
			on: true,
			at: 0,
		}
	}

	rsp := make(chan interface{})

	badMsg := &services.InventoryRepairMsg{
		Target: &services.InventoryAddress{
			Rack:    "",
			Element: &services.InventoryAddress_Tor{},
		},
		After:  &ct.Timestamp{Ticks: common.TickFromContext(ctx)},
		Action: &services.InventoryRepairMsg_Power{
			Power: false,
		},
	}

	go func() {
		r.pdu.Receive(ctx, badMsg, rsp)
	}()

	res := common.CompleteWithinInterface(rsp, time.Duration(1) * time.Second)
	assert.NotNil(res)

	repairResp, ok := res.(*services.InventoryRepairResp)
	require.True(ok)

	reason, ok := repairResp.Rsp.(*services.InventoryRepairResp_Failed)
	require.True(ok)

	assert.Equal(badMsg.Target, repairResp.Source)
	assert.Equal(common.TickFromContext(ctx), repairResp.At.Ticks)

	assert.Equal("working", r.pdu.sm.Current.Name())

	for _, cable := range r.pdu.cables {
		assert.True(cable.on)
	}

	assert.Equal("invalid target specified, request ignored", reason.Failed)
}

func (ts *PduTestSuite) TestPowerOffPdu() {
	require := ts.Require()
	assert := ts.Assert()

	ctx := common.ContextWithTick(context.Background(), 1)

	rackDef := createDummyRack(2)

	r := newRack(ctx, rackDef)
	ctx = common.ContextWithTick(ctx, 2)

	rsp := make(chan interface{})

	msg := &services.InventoryRepairMsg{
		Target: &services.InventoryAddress{
			Rack:    "",
			Element: &services.InventoryAddress_Pdu{},
		},
		After:  &ct.Timestamp{Ticks: common.TickFromContext(ctx)},
		Action: &services.InventoryRepairMsg_Power{
			Power: false,
		},
	}

	go func() {
		r.pdu.Receive(ctx, msg, rsp)
	}()

	res := common.CompleteWithinInterface(rsp, time.Duration(1) * time.Second)
	assert.NotNil(res)

	repairResp, ok := res.(*services.InventoryRepairResp)
	require.True(ok)

	_, ok = repairResp.Rsp.(*services.InventoryRepairResp_Dropped)
	require.True(ok)

	assert.Equal(msg.Target, repairResp.Source)
	assert.Equal(common.TickFromContext(ctx), repairResp.At.Ticks)

	for _, cable := range r.pdu.cables {
		assert.False(cable.on)
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

	rsp := make(chan interface{})

	msg := &services.InventoryRepairMsg{
		Target: &services.InventoryAddress{
			Rack:    "",
			Element: &services.InventoryAddress_Pdu{},
		},
		After:  &ct.Timestamp{Ticks: common.TickFromContext(ctx)},
		Action: &services.InventoryRepairMsg_Power{
			Power: true,
		},
	}

	go func() {
		r.pdu.Receive(ctx, msg, rsp)
	}()

	res := common.CompleteWithinInterface(rsp, time.Duration(1) * time.Second)
	assert.NotNil(res)

	repairResp, ok := res.(*services.InventoryRepairResp)
	require.True(ok)

	_, ok = repairResp.Rsp.(*services.InventoryRepairResp_Dropped)
	require.True(ok)

	assert.Equal(msg.Target, repairResp.Source)
	assert.Equal(common.TickFromContext(ctx), repairResp.At.Ticks)

	for _, cable := range r.pdu.cables {
		assert.False(cable.on)
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

	rsp := make(chan interface{})

	msg := &services.InventoryRepairMsg{
		Target: &services.InventoryAddress{
			Rack:    "",
			Element: &services.InventoryAddress_BladeId{
				BladeId: 0,
			},
		},
		After:  &ct.Timestamp{Ticks: common.TickFromContext(ctx)},
		Action: &services.InventoryRepairMsg_Power{
			Power: true,
		},
	}

	go func() {
		r.pdu.Receive(ctx, msg, rsp)
	}()

	res := common.CompleteWithinInterface(rsp, time.Duration(1) * time.Second)
	assert.NotNil(res)

	repairResp, ok := res.(*services.InventoryRepairResp)
	require.True(ok)

	_, ok = repairResp.Rsp.(*services.InventoryRepairResp_Success)
	require.True(ok)

	assert.Equal(msg.Target, repairResp.Source)
	assert.Equal(common.TickFromContext(ctx), repairResp.At.Ticks)
	assert.Equal(common.TickFromContext(ctx), r.pdu.sm.At)
	assert.Equal(common.TickFromContext(ctx), r.pdu.cables[0].at)
	assert.True(r.pdu.cables[0].on)

	assert.Equal("working", r.pdu.sm.Current.Name())
}

func TestPduTestSuite(t *testing.T) {
	suite.Run(t, new(PduTestSuite))
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