package common

// Bimap is a map with a secondary key.  The secondary key values are not
// required to be unique.
//
// Note Bimap is not fully general purpose.  The keys are designed around its
// use to store time-related structures, indexed by some ID and by simulated
// time tick.
type Bimap struct {
	idMap     map[int]BimapItem
	secondary map[int64][]int
}

// BimapItem defines the required values for an item in a Bimap.
type BimapItem interface {
	// Key returns the primary key value.
	Key() int

	// Secondary returns the secondary key value.
	Secondary() int64
}

// NewBimap constructs a new Bimap instance.
func NewBimap() *Bimap {
	return &Bimap{
		idMap:     make(map[int]BimapItem),
		secondary: make(map[int64][]int),
	}
}

// Count returns the number of entries held in this Bimap instance.
func (m *Bimap) Count() int {
	return len(m.idMap)
}

// SecondaryCount returns the number of unique secondary keys held in this
// Bimap instance.
func (m *Bimap) SecondaryCount() int {
	return len(m.secondary)
}

// Clear removes all entries from this Bimap instance.
func (m *Bimap) Clear() {
	m.idMap = make(map[int]BimapItem)
	m.secondary = make(map[int64][]int)
}

// Get returns the entry with the specified primary key.  If no such entry is
// found, this function returns false as the second return value.
func (m *Bimap) Get(id int) (BimapItem, bool) {
	v, ok := m.idMap[id]
	return v, ok
}

// ForEachSecondary examines each secondary key, and processes the entries
// according to whether or not the secondary key value matches the caller's
// criteria.
//
// The match argument is the function supplied by the caller that returns
// true if the secondary key is to be processed.
//
// The action argument is the function that operates on a supplied entry from
// a secondary key that passed the match filtering.
func (m *Bimap) ForEachSecondary(match func(key int64) bool, action func(item BimapItem)) {
	for i, keys := range m.secondary {
		if match(i) {
			for _, key := range keys {
				action(m.idMap[key])
			}
		}
	}
}

// Add is a function to insert a new entry into this Bimap instance.  It returns
// true if the supplied entry is inserted; false if the id for the supplied
// entry is already present in this Bimap instance.
func (m *Bimap) Add(item BimapItem) bool {
	if _, ok := m.idMap[item.Key()]; ok {
		return false
	}

	m.idMap[item.Key()] = item
	list, _ := m.secondary[item.Secondary()]
	list = append(list, item.Key())

	m.secondary[item.Secondary()] = list
	return true
}

// Remove is a function that removes the entry with the supplied primary key
// value.  It returns true if the entry was found and removed; false if it was
// not found.
func (m *Bimap) Remove(id int) bool {
	item, ok := m.idMap[id]
	if !ok {
		return false
	}

	list := m.secondary[item.Secondary()]

	for i, key := range list {
		if key == id {
			list = append(list[:i], list[i+1:]...)

			if len(list) > 0 {
				m.secondary[item.Secondary()] = list
			} else {
				delete(m.secondary, item.Secondary())
			}

			break
		}
	}

	delete(m.idMap, id)
	return true
}
