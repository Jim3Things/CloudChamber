package common

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

const (
	firstId        = 100
	firstSecondary = 1000
)

type item struct {
	id        int
	secondary int64
	value     string
}

func (i *item) Key() int         { return i.id }
func (i *item) Secondary() int64 { return i.secondary }

type bimapTestSuite struct {
	suite.Suite
}

func (ts *bimapTestSuite) createMap(
	count int,
	firstId int,
	firstSecondary int64,
	collisions int) *Bimap {
	require := ts.Require()

	m := NewBimap()

	sidx := firstSecondary
	dups := 0

	for i := 0; i < count; i++ {
		require.True(m.Add(&item{
			id:        firstId + i,
			secondary: sidx,
			value:     fmt.Sprintf("value%d", i),
		}))

		if dups++; dups >= collisions {
			dups = 0
			sidx++
		}
	}

	return m
}

func (ts *bimapTestSuite) TestSimpleAdd() {
	require := ts.Require()

	m := ts.createMap(10, firstId, firstSecondary, 1)

	require.Equal(10, len(m.idMap))
	require.Equal(10, len(m.secondary))

	for i, bimapItem := range m.idMap {
		idx := i - firstId
		v := bimapItem.(*item)
		require.Equal(fmt.Sprintf("value%d", idx), v.value)
		require.EqualValues(firstSecondary+idx, v.secondary)
	}
}

func (ts *bimapTestSuite) TestDuplicateAdd() {
	require := ts.Require()

	m := ts.createMap(0, 0, 0, 0)

	require.True(m.Add(&item{id: 1, secondary: 11, value: "1"}))
	require.False(m.Add(&item{id: 1, secondary: 11, value: "1"}))
	require.False(m.Add(&item{id: 1, secondary: 12, value: "2"}))
}

func (ts *bimapTestSuite) TestGet() {
	require := ts.Require()

	m := ts.createMap(10, firstId, firstSecondary, 1)

	v, ok := m.Get(firstId + 5)
	require.True(ok)
	require.EqualValues(firstId+5, v.(*item).id)

	v, ok = m.Get(firstId + 15)
	require.False(ok)
	require.Nil(v)
}

func (ts *bimapTestSuite) TestCollisionsAdd() {
	require := ts.Require()

	m := ts.createMap(10, firstId, firstSecondary, 2)

	require.Equal(10, len(m.idMap))
	require.Equal(5, len(m.secondary))

	for _, keys := range m.secondary {
		require.Equal(2, len(keys))
		require.Equal(keys[0]+1, keys[1])
	}
}

func (ts *bimapTestSuite) TestDeletions() {
	require := ts.Require()

	m := ts.createMap(10, firstId, firstSecondary, 2)

	require.True(m.Remove(firstId + 2))

	require.Equal(9, len(m.idMap))
	require.Equal(5, len(m.secondary))

	item, ok := m.Get(firstId + 2)
	require.False(ok)
	require.Nil(item)

	require.False(m.Remove(firstId + 2))

	require.True(m.Remove(firstId + 3))

	require.Equal(8, len(m.idMap))
	require.Equal(4, len(m.secondary))
}

func (ts *bimapTestSuite) TestSecondary() {
	require := ts.Require()

	m := ts.createMap(0, 0, 0, 0)

	m.Add(&item{id: 1, secondary: 10, value: "1"})
	m.Add(&item{id: 3, secondary: 11, value: "3"})
	m.Add(&item{id: 5, secondary: 20, value: "5"})

	count := 0
	m.ForEachSecondary(
		func(key int64) bool { return key == 11 },
		func(item BimapItem) {
			require.EqualValues(11, item.Secondary())
			count++
		})
	require.Equal(1, count)

	count = 0
	m.ForEachSecondary(
		func(key int64) bool { return key <= 11 },
		func(item BimapItem) {
			require.GreaterOrEqual(int64(11), item.Secondary())
			count++
		})
	require.Equal(2, count)

	count = 0
	m.ForEachSecondary(
		func(key int64) bool { return key <= 0 },
		func(item BimapItem) {
			require.GreaterOrEqual(int64(0), item.Secondary())
			count++
		})
	require.Equal(0, count)
}

func (ts *bimapTestSuite) TestClear() {
	require := ts.Require()

	m := ts.createMap(10, firstId, firstSecondary, 1)

	require.Equal(10, len(m.idMap))
	require.Equal(10, len(m.secondary))

	m.Clear()

	require.Equal(0, len(m.idMap))
	require.Equal(0, len(m.secondary))
}

func TestBimapTestSuite(t *testing.T) {
	suite.Run(t, new(bimapTestSuite))
}
