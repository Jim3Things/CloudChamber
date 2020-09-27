package clients

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Jim3Things/CloudChamber/internal/common"
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
	assert.Equal(0, len(timers.idMap))
	assert.Equal(0, len(timers.waiters))
	assert.Equal(1, timers.nextID)
	assert.False(timers.active)
	assert.Equal(1, timers.epoch)
	assert.Equal(ts.ep, timers.dialName)
	assert.ElementsMatch(ts.dialOpts, timers.dialOpts)
}

func (ts *timerTestSuite) TestStartCancel() {
	require := ts.Require()
	assert := ts.Assert()

	ctx := common.ContextWithTick(context.Background(), 10)

	ch := make(chan interface{})

	timers := NewTimers(ts.ep, ts.dialOpts...)

	id, err := timers.Timer(ctx, 1, ch, doneMsg{done: true})
	require.Nil(err)
	require.Less(0, id)

	assert.Equal(1, len(timers.idMap))
	assert.Equal(1, len(timers.waiters))
	assert.Equal(2, timers.nextID)
	assert.True(timers.active)
	assert.Equal(1, timers.epoch)

	err = timers.Cancel(id)
	require.Nil(err)

	assert.Equal(0, len(timers.idMap))
	assert.Equal(0, len(timers.waiters))
	assert.Equal(2, timers.nextID)
	assert.False(timers.active)
	assert.Equal(2, timers.epoch)
}

func TestTimerSuite(t *testing.T) {
	suite.Run(t, new(timerTestSuite))
}
