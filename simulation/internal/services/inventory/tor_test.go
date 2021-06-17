package inventory

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/stretchr/testify/suite"

	tsc "github.com/Jim3Things/CloudChamber/simulation/internal/clients/timestamp"
	"github.com/Jim3Things/CloudChamber/simulation/internal/common"
	"github.com/Jim3Things/CloudChamber/simulation/internal/services/inventory/messages"
	"github.com/Jim3Things/CloudChamber/simulation/internal/sm"
	"github.com/Jim3Things/CloudChamber/simulation/internal/tracing"
	"github.com/Jim3Things/CloudChamber/simulation/pkg/errors"
	pb "github.com/Jim3Things/CloudChamber/simulation/pkg/protos/inventory"
)

type TorTestSuite struct {
	testSuiteCore
}

func (ts *TorTestSuite) TestCreateTor() {
	require := ts.Require()
	assert := ts.Assert()

	rackDef := ts.createDummyRack(2)

	r := newRack(
		context.Background(),
		ts.rackName(),
		rackDef,
		fmt.Sprintf("racks/%s/pdus/", ts.rackName()),
		fmt.Sprintf("racks/%s/tors/", ts.rackName()),
		fmt.Sprintf("racks/%s/blades/", ts.rackName()),
		ts.timers)
	require.NotNil(r)

	require.Len(r.tors, 1)
	t := r.tors[0]
	require.NotNil(t)

	assert.Equal(2, len(t.cables))

	assert.Equal(pb.Actual_Tor_working, t.sm.CurrentIndex)

	for _, c := range t.cables {
		assert.False(c.on)
		assert.False(c.faulted)
	}
}

func (ts *TorTestSuite) TestBadConnectionTarget() {
	require := ts.Require()
	assert := ts.Assert()

	ctx, r := ts.createAndStartRack(context.Background(), 2, true, false)

	require.Len(r.tors, 1)
	t := r.tors[0]
	require.NotNil(t)

	at := common.TickFromContext(ctx)

	for _,  c := range t.cables {
		_, err := c.set(true, at, at)
		require.NoError(err)
	}

	rsp := make(chan *sm.Response)

	r.Receive(
		messages.NewSetConnection(
			ctx,
			messages.NewTargetTor(ts.rackName(), 0, 0),
			common.TickFromContext(ctx),
			false,
			rsp))

	res, ok := ts.completeWithin(rsp, time.Second)
	require.True(ok)
	require.NotNil(res)
	assert.Error(res.Err)
	assert.Equal(errors.ErrInvalidTarget, res.Err)
	assert.Equal(common.TickFromContext(ctx), res.At)
	assert.Nil(res.Msg)

	assert.Equal(pb.Actual_Tor_working, t.sm.CurrentIndex)

	for _, c := range t.cables {
		assert.True(c.on)
	}
}

func (ts *TorTestSuite) TestConnectTooLate() {
	require := ts.Require()
	assert := ts.Assert()

	ctx := ts.advance(context.Background())
	commandTime := common.TickFromContext(ctx)

	ctx, r := ts.createAndStartRack(ctx, 2, true, false)

	require.Len(r.tors, 1)
	t := r.tors[0]
	require.NotNil(t)

	ctx = ts.advance(ctx)
	checkTime := common.TickFromContext(ctx)

	rsp := make(chan *sm.Response)
	target := messages.NewTargetBlade(ts.rackName(), 0, 0)

	msg := messages.NewSetConnection(
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

	assert.Less(t.sm.Guard, checkTime)
	c := t.cables[target.Key()]
	assert.Less(c.Guard, checkTime)
	assert.False(c.on)

	assert.Equal(pb.Actual_Tor_working, t.sm.CurrentIndex)
}

func (ts *TorTestSuite) TestConnectBlade() {
	require := ts.Require()
	assert := ts.Assert()

	ctx, r := ts.createAndStartRack(context.Background(), 2, false, false)

	require.Eventually(func() bool {
		return r.tors[0].sm.CurrentIndex == pb.Actual_Tor_working
	}, time.Second, 10*time.Millisecond,
		"state is %v", r.tors[0].sm.CurrentIndex)

	require.Eventually(func() bool {
		return r.blades[0].sm.CurrentIndex == pb.BladeSmState_off_disconnected
	}, time.Second, 10*time.Millisecond,
		"state is %v", r.blades[0].sm.CurrentIndex)

	require.Len(r.tors, 1)
	t := r.tors[0]
	require.NotNil(t)

	ctx, span := tracing.StartSpan(
		ctx,
		tracing.WithName("test connecting up a blade"),
		tracing.WithContextValue(tsc.EnsureTickInContext))

	rsp := make(chan *sm.Response)
	target := messages.NewTargetBlade(ts.rackName(), 0, 0)

	r.Receive(
		messages.NewSetConnection(
			ctx,
			target,
			common.TickFromContext(ctx),
			true,
			rsp))

	span.End()

	res, ok := ts.completeWithin(rsp, time.Second)
	require.True(ok)
	require.NotNil(res)

	assert.NoError(res.Err)

	assert.Equal(common.TickFromContext(ctx), res.At)
	assert.Nil(res.Msg)

	assert.Equal(common.TickFromContext(ctx), t.sm.Guard)

	c := t.cables[target.Key()]
	assert.Equal(common.TickFromContext(ctx), c.Guard)
	assert.True(c.on)
	assert.False(c.faulted)

	require.Eventually(func() bool {
		return r.blades[0].sm.CurrentIndex == pb.BladeSmState_off_connected
	}, time.Second, 10*time.Millisecond,
		"state is %v", r.blades[0].sm.CurrentIndex)

	assert.Equal(pb.Actual_Tor_working, t.sm.CurrentIndex)

	saved, err := r.tors[0].Save()
	require.NoError(err)

	m := jsonpb.Marshaler{}
	s, err := m.MarshalToString(saved)
	require.NoError(err)

	require.JSONEq(
		`{`+
			`"condition":"operational",`+
			`"cables":{"0":{"state":"on"},"1":{"state":"off"}},`+
			`"smState":"working",`+
			`"core":{"guard": "2"}`+
			`}`,
		s,
	)
}

func (ts *TorTestSuite) TestConnectBladeWhileWorking() {
	require := ts.Require()
	assert := ts.Assert()

	ctx, r := ts.createAndStartRack(context.Background(), 2, false, false)

	target := messages.NewTargetBlade(ts.rackName(), 0, 0)
	ctx = ts.bootBlade(ctx, r, target)

	require.Len(r.tors, 1)
	t := r.tors[0]
	require.NotNil(t)

	c := t.cables[target.Key()]
	require.True(c.on)
	require.False(c.faulted)

	require.Equal(pb.BladeSmState_working, r.blades[0].sm.CurrentIndex)

	ctx, span := tracing.StartSpan(
		ctx,
		tracing.WithName("test connecting up a blade"),
		tracing.WithContextValue(tsc.EnsureTickInContext))

	rsp := make(chan *sm.Response)

	r.Receive(
		messages.NewSetConnection(
			ctx,
			target,
			common.TickFromContext(ctx),
			false,
			rsp))

	span.End()

	res, ok := ts.completeWithin(rsp, time.Second)
	require.True(ok)
	require.NotNil(res)

	assert.NoError(res.Err)

	assert.Equal(common.TickFromContext(ctx), res.At)
	assert.Nil(res.Msg)

	assert.Equal(common.TickFromContext(ctx), t.sm.Guard)

	c = t.cables[target.Key()]

	assert.Equal(common.TickFromContext(ctx), c.Guard)
	assert.False(c.on)
	assert.False(c.faulted)
	require.Equal(c.target.Key(), target.Key())

	require.Eventually(func() bool {
		return r.blades[0].sm.CurrentIndex == pb.BladeSmState_isolated
	}, time.Second, 10*time.Millisecond,
		"state is %v", r.blades[0].sm.CurrentIndex)

	assert.Equal(pb.Actual_Tor_working, t.sm.CurrentIndex)
}

func (ts *TorTestSuite) TestStuckCable() {
	require := ts.Require()
	assert := ts.Assert()

	ctx, r := ts.createAndStartRack(context.Background(), 2, true, false)

	require.Len(r.tors, 1)
	t := r.tors[0]
	require.NotNil(t)

	rsp := make(chan *sm.Response)
	target := messages.NewTargetBlade(ts.rackName(), 0, 0)

	t.cables[target.Key()].faulted = true

	commandTime := common.TickFromContext(ctx)
	msg := messages.NewSetConnection(
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

	assert.Less(t.sm.Guard, commandTime)

	c := t.cables[target.Key()]
	assert.Less(c.Guard, commandTime)
	assert.False(c.on)
	assert.Equal(true, c.faulted)

	assert.Equal(pb.Actual_Tor_working, t.sm.CurrentIndex)
}

func (ts *TorTestSuite) TestWorkingGetStatus() {
	require := ts.Require()
	assert := ts.Assert()

	ctx, r := ts.createAndStartRack(context.Background(), 2, true, false)

	rsp := make(chan *sm.Response)
	msg := messages.NewGetStatus(
		ctx,
		messages.NewTargetTor(ts.rackName(), 0, 0),
		common.TickFromContext(ctx),
		rsp)

	r.Receive(msg)

	res, ok := ts.completeWithin(rsp, time.Second)
	require.True(ok)
	require.NotNil(res)

	assert.NoError(res.Err)
	assert.Equal(common.TickFromContext(ctx), res.At)
	require.NotNil(res.Msg)

	status := res.Msg.(*messages.TorStatus)

	assert.Equal(pb.Actual_Tor_working.String(), status.State)
	assert.Equal(int64(0), status.EnteredAt)

	for i, c := range status.Cables {
		assert.False(c.On, "Cable.On[%d]", i)
		assert.False(c.Faulted, "Cable.Faulted[%d]", i)
	}
}

func TestTorTestSuite(t *testing.T) {
	suite.Run(t, new(TorTestSuite))
}
