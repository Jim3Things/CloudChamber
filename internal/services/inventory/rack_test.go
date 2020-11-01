package inventory

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/Jim3Things/CloudChamber/internal/common"
	"github.com/Jim3Things/CloudChamber/internal/sm"
	"github.com/Jim3Things/CloudChamber/internal/tracing"
)

type RackTestSuite struct {
	testSuiteCore
}

func (ts *RackTestSuite) TestCreateRack() {
	require := ts.Require()
	assert := ts.Assert()

	ctx := common.ContextWithTick(context.Background(), 1)

	rackDef := ts.createDummyRack(2)
	ctx, span := tracing.StartSpan(ctx, tracing.WithName("test rack creation"))
	r := newRack(ctx, ts.rackName(), rackDef)
	span.End()

	require.NotNil(r)
	assert.Equal(len(rackDef.Blades), len(r.blades))
}

func (ts *RackTestSuite) TestStartStopRack() {
	require := ts.Require()
	assert := ts.Assert()

	ctx := common.ContextWithTick(context.Background(), 1)

	rackDef := ts.createDummyRack(2)

	ctx, span := tracing.StartSpan(
		ctx,
		tracing.WithName("test rack start and stop"))

	r := newRack(ctx, ts.rackName(), rackDef)
	require.NotNil(r)
	assert.Equal(len(rackDef.Blades), len(r.blades))
	assert.Equal(rackAwaitingStartState, r.sm.CurrentIndex)

	err := r.start(ctx)
	assert.NoError(err)

	assert.Equal(rackWorkingState, r.sm.CurrentIndex)

	r.stop(ctx)
	assert.Equal(rackTerminalState, r.sm.CurrentIndex)

	span.End()
}

func (ts *RackTestSuite) TestStartStartStopRack() {
	require := ts.Require()
	assert := ts.Assert()

	ctx := common.ContextWithTick(context.Background(), 1)

	rackDef := ts.createDummyRack(2)

	ctx, span := tracing.StartSpan(
		ctx,
		tracing.WithName("test rack start and stop"))

	r := newRack(ctx, ts.rackName(), rackDef)
	require.NotNil(r)
	assert.Equal(len(rackDef.Blades), len(r.blades))
	assert.Equal(rackAwaitingStartState, r.sm.CurrentIndex)

	err := r.start(ctx)
	assert.NoError(err)

	assert.Equal(rackWorkingState, r.sm.CurrentIndex)

	err = r.start(ctx)
	assert.Error(err)
	assert.Equal(ErrRepairMessageDropped, err)

	assert.Equal(rackWorkingState, r.sm.CurrentIndex)

	r.stop(ctx)
	assert.Equal(rackTerminalState, r.sm.CurrentIndex)

	span.End()
}

func (ts *RackTestSuite) TestStartStopStopRack() {
	require := ts.Require()
	assert := ts.Assert()

	ctx := common.ContextWithTick(context.Background(), 1)

	rackDef := ts.createDummyRack(2)

	ctx, span := tracing.StartSpan(
		ctx,
		tracing.WithName("test rack start and stop"))

	r := newRack(ctx, ts.rackName(), rackDef)
	require.NotNil(r)
	assert.Equal(len(rackDef.Blades), len(r.blades))
	assert.Equal(rackAwaitingStartState, r.sm.CurrentIndex)

	err := r.start(ctx)
	assert.NoError(err)

	assert.Equal(rackWorkingState, r.sm.CurrentIndex)

	r.stop(ctx)
	assert.Equal(rackTerminalState, r.sm.CurrentIndex)

	r.stop(ctx)
	assert.Equal(rackTerminalState, r.sm.CurrentIndex)

	span.End()
}

func (ts *RackTestSuite) TestStopNoStartRack() {
	require := ts.Require()
	assert := ts.Assert()

	ctx := common.ContextWithTick(context.Background(), 1)

	rackDef := ts.createDummyRack(2)

	ctx, span := tracing.StartSpan(
		ctx,
		tracing.WithName("test rack start and stop"))

	r := newRack(ctx, ts.rackName(), rackDef)
	require.NotNil(r)
	assert.Equal(len(rackDef.Blades), len(r.blades))
	assert.Equal(rackAwaitingStartState, r.sm.CurrentIndex)

	r.stop(ctx)
	assert.Equal(rackTerminalState, r.sm.CurrentIndex)

	span.End()
}

func (ts *RackTestSuite) TestPowerOnPdu() {
	require := ts.Require()
	assert := ts.Assert()

	ctx := common.ContextWithTick(context.Background(), 1)

	rackDef := ts.createDummyRack(2)

	ctx, span := tracing.StartSpan(
		ctx,
		tracing.WithName("test powering on PDU from rack"))

	r := newRack(ctx, ts.rackName(), rackDef)
	require.NotNil(r)
	assert.Equal(len(rackDef.Blades), len(r.blades))
	assert.Equal(rackAwaitingStartState, r.sm.CurrentIndex)

	err := r.start(ctx)
	assert.NoError(err)

	assert.Equal(rackWorkingState, r.sm.CurrentIndex)

	ctx = common.ContextWithTick(ctx, 2)

	rsp := make(chan *sm.Response)

	msg := newSetPower(ctx, newTargetPdu(ts.rackName()), common.TickFromContext(ctx), true, rsp)


	r.Receive(ctx, msg, rsp)
	span.End()

	res := ts.completeWithin(rsp, time.Duration(1)*time.Second)
	require.NotNil(res)
	require.Error(res.Err)
	assert.Equal(ErrRepairMessageDropped, res.Err)

	// Since the stepper service is not set up, we should expect a false time here.
	assert.Equal(int64(-1), res.At)
	assert.Nil(res.Msg)

	for _, c := range r.pdu.cables {
		assert.False(c.on)
	}

	assert.Equal("working", r.pdu.sm.Current.Name())

	r.stop(ctx)
	assert.Equal(rackTerminalState, r.sm.CurrentIndex)

	span.End()
}
func TestRackTestSuite(t *testing.T) {
	suite.Run(t, new(RackTestSuite))
}
