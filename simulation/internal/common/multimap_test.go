package common

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type MultiMapTestSuite struct {
	suite.Suite
}

type multiMapEntry struct {
	key         int64
	secondaries []int64
}

func (m *multiMapEntry) Primary() PrimaryKey {
	return PrimaryKey(m.key)
}

func (m *multiMapEntry) Secondary(index int) SecondaryKey {
	if index < 0 || index >= len(m.secondaries) {
		return nil
	}

	return m.secondaries[index]
}

func newEntry(key int64, secondaries ...int64) MultiMapEntry {
	return &multiMapEntry{
		key:         key,
		secondaries: secondaries,
	}
}

func (ts *MultiMapTestSuite) TestAdd() {
	require := ts.Require()

	item := newEntry(1, 2, 3)

	m := NewMultiMap(2)
	require.True(m.Add(item))

	require.Equal(1, m.Count())
	require.Equal(1, m.SecondaryCount(0))
	require.Equal(1, m.SecondaryCount(1))

	v, ok := m.Get(1)
	require.True(ok)
	require.Equal(item, v)

	slots, ok := m.GetPrimaryKeysFromSecondary(0, int64(2))
	require.True(ok)
	require.Equal(1, len(slots))
	require.EqualValues(1, slots[0])

	slots, ok = m.GetPrimaryKeysFromSecondary(1, int64(3))
	require.True(ok)
	require.Equal(1, len(slots))
	require.EqualValues(1, slots[0])

	m.Clear()
	require.Zero(m.Count())
	require.Zero(m.SecondaryCount(0))
	require.Zero(m.SecondaryCount(1))
}

func (ts *MultiMapTestSuite) TestAddMissingSecondary() {
	require := ts.Require()

	item := newEntry(1, 2)

	m := NewMultiMap(2)
	require.True(m.Add(item))

	require.Equal(1, m.Count())
	require.Equal(1, m.SecondaryCount(0))
	require.Equal(0, m.SecondaryCount(1))

	v, ok := m.Get(1)
	require.True(ok)
	require.Equal(item, v)

	slots, ok := m.GetPrimaryKeysFromSecondary(0, int64(2))
	require.True(ok)
	require.Equal(1, len(slots))
	require.EqualValues(1, slots[0])

	slots, ok = m.GetPrimaryKeysFromSecondary(1, int64(3))
	require.False(ok)
}

func (ts *MultiMapTestSuite) TestAddDup() {
	require := ts.Require()

	item := newEntry(1, 2)

	m := NewMultiMap(2)

	require.True(m.Add(item))
	require.False(m.Add(item))

	require.EqualValues(1, m.Count())
}

func (ts *MultiMapTestSuite) TestRemove() {
	require := ts.Require()

	item := newEntry(1, 2)

	m := NewMultiMap(2)

	require.True(m.Add(item))

	v, ok := m.Remove(item.Primary())
	require.True(ok)
	require.EqualValues(item, v)
}

func (ts *MultiMapTestSuite) TestRemoveComplex() {
	require := ts.Require()

	item := newEntry(1, 2)

	m := NewMultiMap(2)

	require.True(m.Add(item))
	require.True(m.Add(newEntry(2, 2, 3)))
	require.True(m.Add(newEntry(3, 3, 4)))

	require.EqualValues(3, m.Count())
	require.EqualValues(2, m.SecondaryCount(0))
	require.EqualValues(2, m.SecondaryCount(1))

	v, ok := m.Remove(item.Primary())
	require.True(ok)
	require.EqualValues(item, v)

	require.EqualValues(2, m.Count())
	require.EqualValues(2, m.SecondaryCount(0))
	require.EqualValues(2, m.SecondaryCount(1))
}

func (ts *MultiMapTestSuite) TestRemoveDup() {
	require := ts.Require()

	item := newEntry(1, 2)

	m := NewMultiMap(2)

	require.True(m.Add(item))

	v, ok := m.Remove(item.Primary())
	require.True(ok)
	require.EqualValues(item, v)

	v, ok = m.Remove(item.Primary())
	require.False(ok)
	require.Nil(v)
}

func (ts *MultiMapTestSuite) TestForEach() {
	require := ts.Require()

	m := NewMultiMap(2)

	require.True(m.Add(newEntry(1, 2, 3)))
	require.True(m.Add(newEntry(2, 3, 4)))
	require.True(m.Add(newEntry(3, 3, 4)))

	var found []PrimaryKey

	m.ForEach(func(item MultiMapEntry) {
		found = append(found, item.Primary())
	})

	require.EqualValues(3, len(found))
	require.ElementsMatch([]PrimaryKey{1, 2, 3}, found)
}

func (ts *MultiMapTestSuite) TestForEachSecondary() {
	require := ts.Require()

	m := NewMultiMap(2)

	require.True(m.Add(newEntry(1, 2, 3)))
	require.True(m.Add(newEntry(2, 3, 4)))
	require.True(m.Add(newEntry(3, 3, 4)))

	var found []PrimaryKey
	var secondaries []SecondaryKey

	m.ForEachSecondary(0, func(key SecondaryKey, keys []PrimaryKey) {
		secondaries = append(secondaries, key)
		found = append(found, keys...)
	})

	require.EqualValues(2, len(secondaries))
	require.ElementsMatch([]SecondaryKey{int64(2), int64(3)}, secondaries)

	require.EqualValues(3, len(found))
	require.ElementsMatch([]PrimaryKey{1, 2, 3}, found)
}

func (ts *MultiMapTestSuite) TestForEachSecondaryMissing() {
	require := ts.Require()

	m := NewMultiMap(2)

	require.True(m.Add(newEntry(1, 2, 3)))
	require.True(m.Add(newEntry(2, 3)))
	require.True(m.Add(newEntry(3, 3, 4)))

	var found []PrimaryKey
	var secondaries []SecondaryKey

	m.ForEachSecondary(0, func(key SecondaryKey, keys []PrimaryKey) {
		secondaries = append(secondaries, key)
		found = append(found, keys...)
	})

	require.EqualValues(2, len(secondaries))
	require.ElementsMatch([]SecondaryKey{int64(2), int64(3)}, secondaries)

	require.EqualValues(3, len(found))
	require.ElementsMatch([]PrimaryKey{1, 2, 3}, found)

	found = []PrimaryKey{}
	secondaries = []SecondaryKey{}

	m.ForEachSecondary(1, func(key SecondaryKey, keys []PrimaryKey) {
		secondaries = append(secondaries, key)
		found = append(found, keys...)
	})

	require.EqualValues(2, len(secondaries))
	require.ElementsMatch([]SecondaryKey{int64(3), int64(4)}, secondaries)

	require.EqualValues(2, len(found))
	require.ElementsMatch([]PrimaryKey{1, 3}, found)
}

func (ts *MultiMapTestSuite) TestForEachSecondaryBadIndex() {
	require := ts.Require()

	m := NewMultiMap(2)

	require.True(m.Add(newEntry(1, 2, 3)))
	require.True(m.Add(newEntry(2, 3)))
	require.True(m.Add(newEntry(3, 3, 4)))

	m.ForEachSecondary(3, func(key SecondaryKey, keys []PrimaryKey) {
		require.Fail("Should not have found any entries")
	})
}

func (ts *MultiMapTestSuite) TestGetSecondaryBadIndex() {
	require := ts.Require()

	m := NewMultiMap(2)

	require.True(m.Add(newEntry(1, 2, 3)))
	require.True(m.Add(newEntry(2, 3)))
	require.True(m.Add(newEntry(3, 3, 4)))

	v, ok := m.GetPrimaryKeysFromSecondary(3, 1)
	require.False(ok)
	require.Nil(v)

	v, ok = m.GetPrimaryKeysFromSecondary(-1, 1)
	require.False(ok)
	require.Nil(v)
}

func (ts *MultiMapTestSuite) TestGetSecondaryCountBadIndex() {
	require := ts.Require()

	m := NewMultiMap(2)

	require.True(m.Add(newEntry(1, 2, 3)))
	require.True(m.Add(newEntry(2, 3)))
	require.True(m.Add(newEntry(3, 3, 4)))

	require.Zero(m.SecondaryCount(-1))
	require.EqualValues(2, m.SecondaryCount(0))
	require.EqualValues(2, m.SecondaryCount(1))
	require.Zero(m.SecondaryCount(2))
}

func TestMultiMapTestSuite(t *testing.T) {
	suite.Run(t, new(MultiMapTestSuite))
}
