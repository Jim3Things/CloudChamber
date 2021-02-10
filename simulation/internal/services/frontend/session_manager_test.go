package frontend

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/Jim3Things/CloudChamber/simulation/internal/tracing/exporters"
)

type SessionManagerTestSuite struct {
	suite.Suite

	utf *exporters.Exporter

	sessions *sessionTable
}

func (ts *SessionManagerTestSuite) SetupSuite() {
	ts.utf = exporters.NewExporter(exporters.NewUTForwarder())
	exporters.ConnectToProvider(ts.utf)
}

func (ts *SessionManagerTestSuite) SetupTest() {
	_ = ts.utf.Open(ts.T())

	ts.sessions = newSessionTable(100, expirationTimeout).(*sessionTable)
}

func (ts *SessionManagerTestSuite) TearDownTest() {
	ts.utf.Close()
}

// forceCollision forces a timeout collision with a number of test entries.
func (ts *SessionManagerTestSuite) forceCollision(
	count int,
	dueTime time.Time) []int64 {
	var ids []int64

	for i := 0; i < count; i++ {
		s := sessionState{
			name:    fmt.Sprintf("test-%d", i),
			timeout: dueTime,
		}

		ts.sessions.lastID++
		id := ts.sessions.lastID
		ids = append(ids, id)

		ts.sessions.addToTables(id, s)
	}

	return ids
}

func (ts *SessionManagerTestSuite) add(count int, tag string) []int64 {
	require := ts.Require()
	var ids []int64

	for i := 0; i < count; i++ {
		id, err := ts.sessions.add(sessionState{
			name: fmt.Sprintf("%s-%d", tag, i),
		})

		require.NoError(err)
		time.Sleep(time.Duration(10) * time.Millisecond)

		ids = append(ids, id)
	}

	return ids
}

func (ts *SessionManagerTestSuite) TestSimple() {
	require := ts.Require()

	ids := ts.add(3, "base")

	require.Equal(3, len(ts.sessions.known))
	require.Equal(3, len(ts.sessions.timeouts))
	require.Equal(3, ts.sessions.count())

	require.ElementsMatch(ids, ts.sessions.knownIDs())

	var items []sessionState

	for _, id := range ids {
		item, ok := ts.sessions.get(id)
		require.True(ok)

		items = append(items, item)
	}

	for i, id := range ids {
		removed, ok := ts.sessions.delete(id)
		require.True(ok)
		require.Equal(items[i], removed)
	}

	require.Equal(0, len(ts.sessions.known))
	require.Equal(0, len(ts.sessions.timeouts))
	require.Equal(0, ts.sessions.count())

	require.Equal(0, len(ts.sessions.knownIDs()))
}

func (ts *SessionManagerTestSuite) TestTouch() {
	require := ts.Require()

	ids := ts.add(3, "base")

	expected, ok := ts.sessions.get(ids[1])
	require.True(ok)

	s, ok := ts.sessions.touch(ids[1])
	require.True(ok)

	require.True(expected.timeout.Before(s.timeout))

	require.Equal(3, len(ts.sessions.known))
	require.Equal(3, len(ts.sessions.timeouts))
	require.Equal(3, ts.sessions.count())
}

func (ts *SessionManagerTestSuite) TestTouchCollision() {
	require := ts.Require()

	ids := ts.add(3, "base")

	expected, ok := ts.sessions.get(ids[1])
	require.True(ok)

	collisions := ts.forceCollision(1, expected.timeout)

	s, ok := ts.sessions.touch(ids[1])
	require.True(ok)

	require.True(expected.timeout.Before(s.timeout))

	require.Equal(4, len(ts.sessions.known))
	require.Equal(4, len(ts.sessions.timeouts))
	require.Equal(4, ts.sessions.count())

	time.Sleep(time.Duration(10) * time.Millisecond)

	s, ok = ts.sessions.touch(collisions[0])
	require.True(ok)

	require.True(expected.timeout.Before(s.timeout))

	require.Equal(4, len(ts.sessions.known))
	require.Equal(4, len(ts.sessions.timeouts))
	require.Equal(4, ts.sessions.count())
}

func (ts *SessionManagerTestSuite) TestPurge() {
	require := ts.Require()

	expiry := time.Now()

	ids := ts.forceCollision(2, expiry)

	require.Equal(2, len(ids))
	require.Equal(2, ts.sessions.count())
	require.Equal(1, len(ts.sessions.timeouts))

	time.Sleep(time.Duration(10) * time.Millisecond)

	newIds := ts.add(1, "base")
	require.Equal(1, ts.sessions.count())
	require.Equal(1, len(ts.sessions.timeouts))

	for _, id := range ids {
		_, ok := ts.sessions.get(id)
		require.False(ok)
	}

	s, ok := ts.sessions.get(newIds[0])
	require.True(ok)
	require.True(expiry.Before(s.timeout))
}

func (ts *SessionManagerTestSuite) TestCollisionSimple() {
	require := ts.Require()

	expiry := time.Now().Add(expirationTimeout)

	ids := ts.forceCollision(2, expiry)

	require.Equal(2, len(ids))
	require.Equal(2, ts.sessions.count())
	require.Equal(1, len(ts.sessions.timeouts))

	for _, id := range ids {
		expected, ok := ts.sessions.get(id)
		require.True(ok)

		s, ok := ts.sessions.delete(id)
		require.True(ok)
		require.Equal(expected, s)
	}

	require.Equal(0, ts.sessions.count())
	require.Equal(0, len(ts.sessions.timeouts))
}

func (ts *SessionManagerTestSuite) TestCollisionMulti() {
	require := ts.Require()

	ids := ts.add(3, "base")

	mid, ok := ts.sessions.get(ids[1])
	require.True(ok)

	collisions := ts.forceCollision(1, mid.timeout)
	require.Equal(1, len(collisions))
	require.Equal(4, ts.sessions.count())
	require.Equal(3, len(ts.sessions.timeouts))

	s, ok := ts.sessions.delete(collisions[0])
	require.True(ok)
	require.Equal(3, ts.sessions.count())
	require.Equal(3, len(ts.sessions.timeouts))
	require.Equal(mid.timeout, s.timeout)

	s, ok = ts.sessions.get(collisions[0])
	require.False(ok)

	s, ok = ts.sessions.delete(ids[1])
	require.True(ok)
	require.Equal(2, ts.sessions.count())
	require.Equal(2, len(ts.sessions.timeouts))
	require.Equal(mid.timeout, s.timeout)
	require.Equal(mid.timeout, s.timeout)
}

func (ts *SessionManagerTestSuite) TestCollisionLongChain() {
	require := ts.Require()

	expiry := time.Now().Add(expirationTimeout)

	ids := ts.forceCollision(10, expiry)

	require.Equal(10, len(ids))
	require.Equal(10, ts.sessions.count())
	require.Equal(1, len(ts.sessions.timeouts))

	for _, id := range ids {
		e, ok := ts.sessions.get(id)
		require.True(ok)
		require.NotNil(e)
		require.Equal(expiry, e.timeout)
	}

	entry, ok := ts.sessions.get(ids[3])
	require.True(ok)

	ts.sessions.delete(ids[3])
	ids = append(ids[0:2], ids[4:]...)

	require.Equal(9, ts.sessions.count())

	for _, id := range ids {
		e, ok := ts.sessions.get(id)
		require.True(ok)
		require.NotNil(e)
		require.NotEqual(entry.name, e.name)
		require.Equal(expiry, e.timeout)
	}
}

func TestSessionManagerTestSuite(t *testing.T) {
	suite.Run(t, new(SessionManagerTestSuite))
}
