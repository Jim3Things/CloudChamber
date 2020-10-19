package inventory

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel/api/trace"

	"github.com/Jim3Things/CloudChamber/internal/common"
	"github.com/Jim3Things/CloudChamber/internal/sm"
	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters"
	ct "github.com/Jim3Things/CloudChamber/pkg/protos/common"
	"github.com/Jim3Things/CloudChamber/pkg/protos/services"
)

type TorTestSuite struct {
	suite.Suite

	utf *exporters.Exporter
}

func (ts *TorTestSuite) SetupSuite() {
	ts.utf = exporters.NewExporter(exporters.NewUTForwarder())
	exporters.ConnectToProvider(ts.utf)
}

func (ts *TorTestSuite) SetupTest() {
	_ = ts.utf.Open(ts.T())
}

func (ts *TorTestSuite) TearDownTest() {
	ts.utf.Close()
}

func (ts *TorTestSuite) completeWithin(ch <-chan *sm.Response, delay time.Duration) *sm.Response {
	select {
	case res := <-ch:
		return res
	case <-time.After(delay):
		return nil
	}
}

func (ts *TorTestSuite) TestCreateTor() {
	require := ts.Require()
	assert := ts.Assert()

	rackDef := createDummyRack(2)

	r := newRack(context.Background(), rackDef)
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

	rackDef := createDummyRack(2)

	r := newRack(ctx, rackDef)
	ctx = common.ContextWithTick(ctx, 2)

	t := r.tor
	require.NotNil(t)

	for i := range t.cables {
		t.cables[i] = newCable(true, false, 0)
	}

	rsp := make(chan *sm.Response)

	badMsg := &sm.Envelope{
		CH:   rsp,
		Span: trace.SpanContext{},
		Msg: &services.InventoryRepairMsg{
			Target: &services.InventoryAddress{
				Rack:    "",
				Element: &services.InventoryAddress_Tor{},
			},
			After: &ct.Timestamp{Ticks: common.TickFromContext(ctx)},
			Action: &services.InventoryRepairMsg_Connect{
				Connect: false,
			},
		},
	}

	go func() {
		t.Receive(ctx, badMsg)
	}()

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

	rackDef := createDummyRack(2)

	r := newRack(ctx, rackDef)
	require.NotNil(r)
	t := r.tor

	ctx = common.ContextWithTick(ctx, 2)

	rsp := make(chan *sm.Response)

	msg := &sm.Envelope{
		CH:   rsp,
		Span: trace.SpanContext{},
		Msg: &services.InventoryRepairMsg{
			Target: &services.InventoryAddress{
				Rack: "",
				Element: &services.InventoryAddress_BladeId{
					BladeId: 0,
				},
			},
			After: &ct.Timestamp{Ticks: common.TickFromContext(ctx)},
			Action: &services.InventoryRepairMsg_Connect{
				Connect: true,
			},
		},
	}

	go func() {
		t.Receive(ctx, msg)
	}()

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

	rackDef := createDummyRack(2)

	r := newRack(ctx, rackDef)
	t := r.tor
	ctx = common.ContextWithTick(ctx, 2)

	rsp := make(chan *sm.Response)

	msg := &sm.Envelope{
		CH:   rsp,
		Span: trace.SpanContext{},
		Msg: &services.InventoryRepairMsg{
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
	}

	go func() {
		t.Receive(ctx, msg)
	}()

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

	rackDef := createDummyRack(2)

	r := newRack(ctx, rackDef)
	t := r.tor

	startTime := common.TickFromContext(ctx)
	require.Nil(t.cables[0].fault(false, startTime, startTime))
	ctx = common.ContextWithTick(ctx, 2)

	rsp := make(chan *sm.Response)

	msg := &sm.Envelope{
		CH:   rsp,
		Span: trace.SpanContext{},
		Msg:  &services.InventoryRepairMsg{
			Target: &services.InventoryAddress{
				Rack: "",
				Element: &services.InventoryAddress_BladeId{
					BladeId: 0,
				},
			},
			After: &ct.Timestamp{Ticks: common.TickFromContext(ctx)},
			Action: &services.InventoryRepairMsg_Connect{
				Connect: true,
			},
		},
	}

	go func() {
		t.Receive(ctx, msg)
	}()

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
