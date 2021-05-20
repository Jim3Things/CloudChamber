package inventory

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	tsc "github.com/Jim3Things/CloudChamber/simulation/internal/clients/timestamp"
	"github.com/Jim3Things/CloudChamber/simulation/internal/common"
	"github.com/Jim3Things/CloudChamber/simulation/internal/services/inventory/messages"
	"github.com/Jim3Things/CloudChamber/simulation/internal/sm"
	"github.com/Jim3Things/CloudChamber/simulation/internal/tracing"
	"github.com/Jim3Things/CloudChamber/simulation/pkg/errors"
	pb "github.com/Jim3Things/CloudChamber/simulation/pkg/protos/inventory"
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

	r := newRack(
		ctx,
		ts.rackName(),
		rackDef,
		fmt.Sprintf("racks/%s/pdus/", ts.rackName()),
		fmt.Sprintf("racks/%s/tors/", ts.rackName()),
		fmt.Sprintf("racks/%s/blades/", ts.rackName()),
		ts.timers)

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

	r := newRack(
		ctx,
		ts.rackName(),
		rackDef,
		fmt.Sprintf("racks/%s/pdus/", ts.rackName()),
		fmt.Sprintf("racks/%s/tors/", ts.rackName()),
		fmt.Sprintf("racks/%s/blades/", ts.rackName()),
		ts.timers)
	require.NotNil(r)
	assert.Equal(len(rackDef.Blades), len(r.blades))
	assert.Equal(pb.Actual_Rack_awaiting_start, r.sm.CurrentIndex)

	err := r.start(ctx)
	assert.NoError(err)

	require.Eventually(func() bool {
		return r.sm.CurrentIndex == pb.Actual_Rack_working
	}, time.Second, 10*time.Millisecond,
		"state is %v", r.sm.CurrentIndex)

	r.stop(ctx)
	require.Eventually(func() bool {
		return r.sm.CurrentIndex == pb.Actual_Rack_terminated
	}, time.Second, 10*time.Millisecond,
		"state is %v", r.sm.CurrentIndex)

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

	r := newRack(
		ctx,
		ts.rackName(),
		rackDef,
		fmt.Sprintf("racks/%s/pdus/", ts.rackName()),
		fmt.Sprintf("racks/%s/tors/", ts.rackName()),
		fmt.Sprintf("racks/%s/blades/", ts.rackName()),
		ts.timers)
	require.NotNil(r)
	assert.Equal(len(rackDef.Blades), len(r.blades))
	assert.Equal(pb.Actual_Rack_awaiting_start, r.sm.CurrentIndex)

	err := r.start(ctx)
	assert.NoError(err)

	require.Eventually(func() bool {
		return r.sm.CurrentIndex == pb.Actual_Rack_working
	}, time.Second, 10*time.Millisecond,
		"state is %v", r.sm.CurrentIndex)

	err = r.start(ctx)
	assert.Error(err)
	assert.Equal(errors.ErrAlreadyStarted, err)

	assert.Equal(pb.Actual_Rack_working, r.sm.CurrentIndex)

	r.stop(ctx)
	require.Eventually(func() bool {
		return r.sm.CurrentIndex == pb.Actual_Rack_terminated
	}, time.Second, 10*time.Millisecond,
		"state is %v", r.sm.CurrentIndex)

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

	r := newRack(
		ctx,
		ts.rackName(),
		rackDef,
		fmt.Sprintf("racks/%s/pdus/", ts.rackName()),
		fmt.Sprintf("racks/%s/tors/", ts.rackName()),
		fmt.Sprintf("racks/%s/blades/", ts.rackName()),
		ts.timers)
	require.NotNil(r)
	assert.Equal(len(rackDef.Blades), len(r.blades))
	assert.Equal(pb.Actual_Rack_awaiting_start, r.sm.CurrentIndex)

	err := r.start(ctx)
	assert.NoError(err)

	require.Eventually(func() bool {
		return r.sm.CurrentIndex == pb.Actual_Rack_working
	}, time.Second, 10*time.Millisecond,
		"state is %v", r.sm.CurrentIndex)

	r.stop(ctx)
	require.Eventually(func() bool {
		return r.sm.CurrentIndex == pb.Actual_Rack_terminated
	}, time.Second, 10*time.Millisecond,
		"state is %v", r.sm.CurrentIndex)

	r.stop(ctx)
	assert.Equal(pb.Actual_Rack_terminated, r.sm.CurrentIndex)

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

	r := newRack(
		ctx,
		ts.rackName(),
		rackDef,
		fmt.Sprintf("racks/%s/pdus/", ts.rackName()),
		fmt.Sprintf("racks/%s/tors/", ts.rackName()),
		fmt.Sprintf("racks/%s/blades/", ts.rackName()),
		ts.timers)
	require.NotNil(r)
	assert.Equal(len(rackDef.Blades), len(r.blades))
	assert.Equal(pb.Actual_Rack_awaiting_start, r.sm.CurrentIndex)

	r.stop(ctx)
	assert.Equal(pb.Actual_Rack_terminated, r.sm.CurrentIndex)

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

	r := newRack(
		ctx,
		ts.rackName(),
		rackDef,
		fmt.Sprintf("racks/%s/pdus/", ts.rackName()),
		fmt.Sprintf("racks/%s/tors/", ts.rackName()),
		fmt.Sprintf("racks/%s/blades/", ts.rackName()),
		ts.timers)
	require.NotNil(r)
	assert.Equal(len(rackDef.Blades), len(r.blades))
	assert.Equal(pb.Actual_Rack_awaiting_start, r.sm.CurrentIndex)

	err := r.start(ctx)
	assert.NoError(err)

	require.Eventually(func() bool {
		return r.sm.CurrentIndex == pb.Actual_Rack_working
	}, time.Second, 10*time.Millisecond,
		"state is %v", r.sm.CurrentIndex)

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

	res, ok := ts.completeWithin(rsp, time.Second)
	require.True(ok)
	require.Nil(res)

	for _, c := range r.pdu.cables {
		assert.False(c.on)
	}

	assert.Equal(pb.Actual_Pdu_working, r.pdu.sm.CurrentIndex)

	r.stop(ctx)
	assert.Equal(pb.Actual_Rack_terminated, r.sm.CurrentIndex)

	span.End()
}

func TestRackTestSuite(t *testing.T) {
	suite.Run(t, new(RackTestSuite))
}
