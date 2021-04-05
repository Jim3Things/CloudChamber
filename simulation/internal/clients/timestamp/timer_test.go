package timestamp

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/Jim3Things/CloudChamber/simulation/internal/common"
)

type timerTestSuite struct {
	stepperTestClientSuite
}

type doneMsg struct {
	done bool
}

func (ts *timerTestSuite) TestSimple() {
	require := ts.Require()
	assert := ts.Assert()

	timers := NewTimers(ts.ep, ts.dialOpts...)

	require.NotNil(timers)
	ts.validateTimerState(timers, 0, 0, 1, false, 1)

	assert.Equal(ts.ep, timers.dialName)
	assert.ElementsMatch(ts.dialOpts, timers.dialOpts)
}

func (ts *timerTestSuite) TestStartCancel() {
	require := ts.Require()

	ctx := common.ContextWithTick(context.Background(), Tick(context.Background()))

	ch := make(chan interface{}, 10)

	timers := NewTimers(ts.ep, ts.dialOpts...)

	id, err := timers.Timer(ctx, 1, doneMsg{done: true}, func(msg interface{}) {
		ch <- msg
	})
	require.Nil(err)
	require.Less(0, id)

	ts.validateTimerState(timers, 1, 1, 2, true, 1)

	err = timers.Cancel(id)
	require.Nil(err)

	ts.validateTimerState(timers, 0, 0, 2, false, 2)

	require.Nil(Advance(ctx))
}

func (ts *timerTestSuite) TestStartStartCancel() {
	require := ts.Require()

	ctx := common.ContextWithTick(context.Background(), Tick(context.Background()))

	ch := make(chan interface{}, 10)

	timers := NewTimers(ts.ep, ts.dialOpts...)

	id, err := timers.Timer(ctx, 1, doneMsg{done: true}, func(msg interface{}) {
		ch <- msg
	})
	require.Nil(err)
	require.Less(0, id)

	ts.validateTimerState(timers, 1, 1, 2, true, 1)

	id2, err := timers.Timer(ctx, 1, doneMsg{done: true}, func(msg interface{}) {
		ch <- msg
	})
	require.Nil(err)
	require.Less(0, id2)
	require.NotEqual(id, id2)

	ts.validateTimerState(timers, 2, 1, 3, true, 1)

	require.Nil(timers.Cancel(id))

	ts.validateTimerState(timers, 1, 1, 3, true, 1)

	require.Nil(timers.Cancel(id2))

	ts.validateTimerState(timers, 0, 0, 3, false, 2)

	require.Nil(Advance(ctx))
}

func (ts *timerTestSuite) TestStartExpire() {
	require := ts.Require()
	assert := ts.Assert()

	ctx := common.ContextWithTick(context.Background(), Tick(context.Background()))

	ch := make(chan interface{}, 10)

	timers := NewTimers(ts.ep, ts.dialOpts...)

	id, err := timers.Timer(ctx, 1, doneMsg{done: true}, func(msg interface{}) {
		ch <- msg
	})
	require.Nil(err)

	assert.Nil(common.DoNotCompleteWithinInterface(ch, time.Duration(2)*time.Second))

	require.Nil(Advance(ctx))
	ctx = common.ContextWithTick(ctx, Tick(ctx))

	assert.NotNil(common.CompleteWithinInterface(ch, time.Duration(2)*time.Second))

	ts.validateTimerState(timers, 0, 0, 2, false, 2)

	require.NotNil(timers.Cancel(id))
	require.Nil(Advance(ctx))
}

func (ts *timerTestSuite) TestStartExpireTwice() {
	require := ts.Require()
	assert := ts.Assert()

	ctx := common.ContextWithTick(context.Background(), Tick(context.Background()))

	ch := make(chan interface{}, 10)

	timers := NewTimers(ts.ep, ts.dialOpts...)

	id, err := timers.Timer(ctx, 1, doneMsg{done: true}, func(msg interface{}) {
		ch <- msg
	})
	require.Nil(err)

	assert.Nil(common.DoNotCompleteWithinInterface(ch, time.Duration(2)*time.Second))

	require.Nil(Advance(ctx))
	ctx = common.ContextWithTick(ctx, Tick(ctx))

	assert.NotNil(common.CompleteWithinInterface(ch, time.Duration(2)*time.Second))

	ts.validateTimerState(timers, 0, 0, 2, false, 2)

	require.NotNil(timers.Cancel(id))

	ch = make(chan interface{}, 10)

	id, err = timers.Timer(ctx, 1, doneMsg{done: true}, func(msg interface{}) {
		ch <- msg
	})
	require.Nil(err)

	assert.Nil(common.DoNotCompleteWithinInterface(ch, time.Duration(2)*time.Second))

	require.Nil(Advance(ctx))
	ctx = common.ContextWithTick(ctx, Tick(ctx))

	assert.NotNil(common.CompleteWithinInterface(ch, time.Duration(2)*time.Second))

	ts.validateTimerState(timers, 0, 0, 3, false, 3)

	require.NotNil(timers.Cancel(id))
}

func (ts *timerTestSuite) validateTimerState(timers *Timers, idLen int, waiterLen int, nextID int, active bool, epoch int) {
	assert := ts.Assert()

	timers.m.Lock()
	defer timers.m.Unlock()

	assert.Equal(idLen, timers.waiters.Count())
	assert.Equal(waiterLen, timers.waiters.SecondaryCount())
	assert.Equal(nextID, timers.nextID)
	assert.Equal(active, timers.active)
	assert.Equal(epoch, timers.epoch)
}

func TestTimerSuite(t *testing.T) {
	suite.Run(t, new(timerTestSuite))
}
