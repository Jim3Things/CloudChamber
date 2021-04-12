package common

// This module contains the interface and implementation for a multi-key map
// collection.
//
// The map is defined with a number of secondary keys, and each entry provides
// an implementation of an interface that returns that entry's primary and
// secondary keys.
//
// Primary keys are unique, secondary keys are not.  Secondary keys may also be
// sparse, in that not all entries need be present in all secondary keys.

// PrimaryKey is the abstract type for an entry's primary key
type PrimaryKey int

// SecondaryKey is the abstract type for each of the secondary keys for an entry
type SecondaryKey interface{}

// MultiMapEntry is the interface that an entry instance is required to support
// in order to be placed in a MultiMap.
type MultiMapEntry interface {
	// Primary returns the primary key value for this entry.
	Primary() PrimaryKey

	// Secondary returns the key value for the secondary key specified by the
	// index parameter.  If the entry does not support that secondary key, it
	// must return nil.
	Secondary(index int) SecondaryKey
}

// MultiMap is the interface callers use for accessing a MultiMap instance.
type MultiMap interface {
	// Get returns the instance associated with the supplied primary key, or
	// nil, if it was not found.  The second return value is true, if it was
	// found, or false, if it was not.
	Get(key PrimaryKey) (MultiMapEntry, bool)

	// GetSecondary returns the primary keys for entries that match the supplied
	// value for the secondary key at the index.  The second return value is
	// true, if it was found, or false, if it was not.
	GetSecondary(index int, key SecondaryKey) ([]PrimaryKey, bool)

	// Add attempts to add the supplied entry to the MultiMap instance.  It
	// returns true if the new entry was inserted; false, if it was not.  The
	// reason for insertion failing would be due to the primary key value for
	// this new entry matching one already in the MultiMap instance.
	Add(entry MultiMapEntry) bool

	// Remove attempts to remove the entry identified by the supplied primary
	// key.  It returns the current entry, if it was found, and true, if the
	// entry was found and removed, or false, if no such entry was found.
	Remove(key PrimaryKey) (MultiMapEntry, bool)

	// ForEach calls the supplied action function with each entry held in the
	// MultiMap instance, in random order.
	ForEach(action func(item MultiMapEntry))

	// ForEachSecondary processes each entry for the specified secondary key.
	// It calls the supplied action function for each key value, passing that
	// value and the set of primary keys associated with it.
	ForEachSecondary(index int, action func(key SecondaryKey, items []PrimaryKey))

	// Count returns the number of entries held in this MultiMap instance.
	Count() int

	// Count returns the number of unique secondary keys held in this MultiMap
	// instance for the specified secondary key index.
	SecondaryCount(index int) int

	// Clear removes all existing entries and resets all secondary keys indices.
	Clear()
}

// MultiMapImpl is an implementation of the MultiMap interface.  This
// implementation uses multiple maps and slice for the indices.
type MultiMapImpl struct {
	items       map[PrimaryKey]MultiMapEntry
	secondaries []map[SecondaryKey][]PrimaryKey
}

// NewMultiMap creates a new MultiMap instance.  It requires that the caller
// supply the number of secondary key indices.
func NewMultiMap(count int) MultiMap {
	m := &MultiMapImpl{
		items:       make(map[PrimaryKey]MultiMapEntry),
		secondaries: make([]map[SecondaryKey][]PrimaryKey, count),
	}

	for i := 0; i < count; i++ {
		m.secondaries[i] = make(map[SecondaryKey][]PrimaryKey)
	}

	return m
}

func (m *MultiMapImpl) Get(key PrimaryKey) (MultiMapEntry, bool) {
	v, ok := m.items[key]
	return v, ok
}

func (m *MultiMapImpl) GetSecondary(index int, key SecondaryKey) ([]PrimaryKey, bool) {
	if index < 0 || index >= len(m.secondaries) {
		return nil, false
	}

	slice := m.secondaries[index]
	v, ok := slice[key]
	return v, ok
}

func (m *MultiMapImpl) Add(entry MultiMapEntry) bool {
	// First, ensure that there is no collision on the primary key value.
	p := entry.Primary()
	if _, ok := m.items[p]; ok {
		return false
	}

	// Second, since it is unique, go ahead and add it.
	m.items[p] = entry

	// Finally, now add all supplied secondary keys from the entry.
	//
	for i := 0; i < len(m.secondaries); i++ {
		item := entry.Secondary(i)
		if item != nil {
			v, _ := m.secondaries[i][item]
			v = append(v, p)
			m.secondaries[i][item] = v
		}
	}

	return true
}

func (m *MultiMapImpl) Remove(key PrimaryKey) (MultiMapEntry, bool) {
	v, ok := m.items[key]
	if !ok {
		return nil, false
	}

	for i, secondary := range m.secondaries {
		sk := v.Secondary(i)
		if sk != nil {
			sl := removeFromSlice(key, secondary[sk])
			if len(sl) == 0 {
				delete(m.secondaries[i], sk)
			} else {
				m.secondaries[i][sk] = sl
			}
		}
	}

	delete(m.items, key)

	return v, true
}

func (m *MultiMapImpl) ForEach(action func(item MultiMapEntry)) {
	for _, entry := range m.items {
		action(entry)
	}
}

func (m *MultiMapImpl) ForEachSecondary(index int, action func(key SecondaryKey, items []PrimaryKey)) {
	if index < 0 || index >= len(m.secondaries) {
		return
	}

	for key, items := range m.secondaries[index] {
		action(key, items)
	}
}

func (m *MultiMapImpl) Count() int {
	return len(m.items)
}

func (m *MultiMapImpl) SecondaryCount(index int) int {
	if index < 0 || index >= len(m.secondaries) {
		return 0
	}

	return len(m.secondaries[index])
}

func (m *MultiMapImpl) Clear() {
	m.items = make(map[PrimaryKey]MultiMapEntry)
	for i := 0; i < len(m.secondaries); i++ {
		m.secondaries[i] = make(map[SecondaryKey][]PrimaryKey)
	}
}

func removeFromSlice(key PrimaryKey, slice []PrimaryKey) []PrimaryKey {
	res := make([]PrimaryKey, 0, len(slice))

	for _, item := range slice {
		if item != key {
			res = append(res, item)
		}
	}

	return res
}
