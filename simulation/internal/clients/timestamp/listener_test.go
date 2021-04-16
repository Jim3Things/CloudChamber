package timestamp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
)

type listenerTestSuite struct {
	stepperTestClientSuite
}

func (ts *listenerTestSuite) TestStartListener() {
	require := ts.Require()

	l := NewListener(ts.ep, ts.dialOpts...)
	require.NotNil(l)

	l.Stop()
}

func (ts *listenerTestSuite) TestSubscribeCancelOnStop() {
	require := ts.Require()

	l := NewListener(ts.ep, ts.dialOpts...)
	require.NotNil(l)

	id, notify, err := l.After("foo", 1, NoEpochCheck)
	require.NoError(err)
	require.NotZero(id)

	l.Stop()

	res := <-notify
	require.Error(res.Err)

	_, ok := <-notify
	require.False(ok)
}

func (ts *listenerTestSuite) TestSubscribeCancel() {
	require := ts.Require()

	l := NewListener(ts.ep, ts.dialOpts...)
	require.NotNil(l)
	defer l.Stop()

	id, notify, err := l.After("foo", 1, NoEpochCheck)
	require.NoError(err)
	require.NotZero(id)

	canceler, err := l.Cancel(id)
	require.NoError(err)

	res := <-canceler
	require.NoError(res.Err)

	res = <-notify
	require.Error(res.Err)

	_, ok := <-notify
	require.False(ok)
}

func (ts *listenerTestSuite) TestSubscribeExpire() {
	require := ts.Require()

	l := NewListener(ts.ep, ts.dialOpts...)
	require.NotNil(l)
	defer l.Stop()

	id, notify, err := l.After("foo", 1, NoEpochCheck)
	require.NoError(err)
	require.NotZero(id)

	require.NoError(Advance(context.Background()))
	res := <-notify
	require.NoError(res.Err)
	require.EqualValues(1, res.Status.Now)

	_, ok := <-notify
	require.False(ok)
}

func (ts *listenerTestSuite) TestMulti() {
	require := ts.Require()

	l := NewListener(ts.ep, ts.dialOpts...)
	require.NotNil(l)
	defer l.Stop()

	id, notify, err := l.After("foo", 1, NoEpochCheck)
	require.NoError(err)
	require.NotZero(id)

	id2, notify2, err := l.After("bar", 2, NoEpochCheck)
	require.NoError(err)
	require.NotZero(id2)
	require.NotEqual(id, id2)

	require.NoError(Advance(context.Background()))
	res := <-notify
	require.NoError(res.Err)
	require.EqualValues(1, res.Status.Now)

	_, ok := <-notify
	require.False(ok)

	require.NoError(Advance(context.Background()))
	res = <-notify2
	require.NoError(res.Err)
	require.EqualValues(2, res.Status.Now)

	_, ok = <-notify2
	require.False(ok)
}

func TestListener(t *testing.T) {
	suite.Run(t, new(listenerTestSuite))
}
