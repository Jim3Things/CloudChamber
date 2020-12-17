package inventory

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	tsc "github.com/Jim3Things/CloudChamber/internal/clients/timestamp"
	"github.com/Jim3Things/CloudChamber/internal/common"
	"github.com/Jim3Things/CloudChamber/internal/services/inventory/messages"
	"github.com/Jim3Things/CloudChamber/internal/sm"
	"github.com/Jim3Things/CloudChamber/internal/tracing"
	"github.com/Jim3Things/CloudChamber/test/utilities"
)

type RackTestSuite struct {
	testSuiteCore
}

func (ts *RackTestSuite) TestCreateRack() {
	require := ts.Require()
	assert := ts.Assert()

	ctx := ts.advance(context.Background())

	rackDef := ts.createDummyRack(2)
	ctx, span := tracing.StartSpan(
		context.Background(),
		tracing.WithName("test rack creation"),
		tracing.WithContextValue(tsc.EnsureTickInContext))

	r := newRack(ctx, ts.rackName(), rackDef, ts.timers)

	span.End()

	require.NotNil(r)
	assert.Equal(len(rackDef.Blades), len(r.blades))
}

func (ts *RackTestSuite) TestStartStopRack() {
	require := ts.Require()
	assert := ts.Assert()

	ctx := ts.advance(context.Background())

	rackDef := ts.createDummyRack(2)

	ctx, span := tracing.StartSpan(
		context.Background(),
		tracing.WithName("test rack start and stop"),
		tracing.WithContextValue(tsc.EnsureTickInContext))

	r := newRack(ctx, ts.rackName(), rackDef, ts.timers)
	require.NotNil(r)
	assert.Equal(len(rackDef.Blades), len(r.blades))
	assert.Equal(rackAwaitingStartState, r.sm.CurrentIndex)

	err := r.start(ctx)
	assert.NoError(err)

	ok := utilities.WaitForStateChange(1, func() bool {
		return r.sm.CurrentIndex == rackWorkingState
	})

	require.True(ok, "state is %v", r.sm.CurrentIndex)

	r.stop(ctx)
	ok = utilities.WaitForStateChange(1, func() bool {
		return r.sm.CurrentIndex == rackTerminalState
	})

	require.True(ok, "state is %v", r.sm.CurrentIndex)

	span.End()
}

func (ts *RackTestSuite) TestStartStartStopRack() {
	require := ts.Require()
	assert := ts.Assert()

	ctx := ts.advance(context.Background())

	rackDef := ts.createDummyRack(2)

	ctx, span := tracing.StartSpan(
		context.Background(),
		tracing.WithName("test rack start, start, and stop"),
		tracing.WithContextValue(tsc.EnsureTickInContext))

	r := newRack(ctx, ts.rackName(), rackDef, ts.timers)
	require.NotNil(r)
	assert.Equal(len(rackDef.Blades), len(r.blades))
	assert.Equal(rackAwaitingStartState, r.sm.CurrentIndex)

	err := r.start(ctx)
	assert.NoError(err)

	ok := utilities.WaitForStateChange(1, func() bool {
		return r.sm.CurrentIndex == rackWorkingState
	})

	require.True(ok, "state is %v", r.sm.CurrentIndex)

	err = r.start(ctx)
	assert.Error(err)
	assert.Equal(ErrAlreadyStarted, err)

	assert.Equal(rackWorkingState, r.sm.CurrentIndex)

	r.stop(ctx)
	ok = utilities.WaitForStateChange(1, func() bool {
		return r.sm.CurrentIndex == rackTerminalState
	})

	require.True(ok, "state is %v", r.sm.CurrentIndex)

	span.End()
}

func (ts *RackTestSuite) TestStartStopStopRack() {
	require := ts.Require()
	assert := ts.Assert()

	ctx := ts.advance(context.Background())

	rackDef := ts.createDummyRack(2)

	ctx, span := tracing.StartSpan(
		context.Background(),
		tracing.WithName("test rack start, stop, and stop"),
		tracing.WithContextValue(tsc.EnsureTickInContext))

	r := newRack(ctx, ts.rackName(), rackDef, ts.timers)
	require.NotNil(r)
	assert.Equal(len(rackDef.Blades), len(r.blades))
	assert.Equal(rackAwaitingStartState, r.sm.CurrentIndex)

	err := r.start(ctx)
	assert.NoError(err)

	ok := utilities.WaitForStateChange(1, func() bool {
		return r.sm.CurrentIndex == rackWorkingState
	})

	require.True(ok, "state is %v", r.sm.CurrentIndex)

	r.stop(ctx)
	ok = utilities.WaitForStateChange(1, func() bool {
		return r.sm.CurrentIndex == rackTerminalState
	})

	require.True(ok, "state is %v", r.sm.CurrentIndex)

	r.stop(ctx)
	assert.Equal(rackTerminalState, r.sm.CurrentIndex)

	span.End()
}

func (ts *RackTestSuite) TestStopNoStartRack() {
	require := ts.Require()
	assert := ts.Assert()

	ctx := ts.advance(context.Background())

	rackDef := ts.createDummyRack(2)

	ctx, span := tracing.StartSpan(
		context.Background(),
		tracing.WithName("test rack stop without a start"),
		tracing.WithContextValue(tsc.EnsureTickInContext))

	r := newRack(ctx, ts.rackName(), rackDef, ts.timers)
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

	ctx := ts.advance(context.Background())

	rackDef := ts.createDummyRack(2)

	ctx, span := tracing.StartSpan(
		context.Background(),
		tracing.WithName("test powering on PDU from rack"),
		tracing.WithContextValue(tsc.EnsureTickInContext))

	r := newRack(ctx, ts.rackName(), rackDef, ts.timers)
	require.NotNil(r)
	assert.Equal(len(rackDef.Blades), len(r.blades))
	assert.Equal(rackAwaitingStartState, r.sm.CurrentIndex)

	err := r.start(ctx)
	assert.NoError(err)

	ok := utilities.WaitForStateChange(1, func() bool {
		return r.sm.CurrentIndex == rackWorkingState
	})

	require.True(ok, "state is %v", r.sm.CurrentIndex)

	ctx = ts.advance(ctx)

	rsp := make(chan *sm.Response)

	msg := messages.NewSetPower(
		ctx,
		messages.NewTargetPdu(ts.rackName()),
		common.TickFromContext(ctx),
		true,
		rsp)

	r.Receive(msg)
	span.End()

	res, ok := ts.completeWithin(rsp, time.Duration(1)*time.Second)
	require.True(ok)
	require.Nil(res)

	for _, c := range r.pdu.cables {
		assert.False(c.on)
	}

	assert.Equal(pduWorkingState, r.pdu.sm.CurrentIndex)

	r.stop(ctx)
	assert.Equal(rackTerminalState, r.sm.CurrentIndex)

	span.End()
}

func TestRackTestSuite(t *testing.T) {
	suite.Run(t, new(RackTestSuite))
}
