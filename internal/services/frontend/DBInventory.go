// This module encapsulates storage and retrieval of known racks

// The full rack contains attributes about tor, pdU and blades.

// Each rack has an associated key which is the lowercased form of the
// username.  The supplied name is retained as an attribute in order to present
// the form that the caller originally used for display purposes.

package frontend

import (
	"sync"
)

type Tor struct {
}

type Pdu struct {
}

type blade struct {
}

type rack struct {
	tor    Tor
	pdu    Pdu
	blades map[int64]*blade
}

// DBInventory is a container used to establish synchronized access to
// the in-memory set of Racks records.
//
type DBInventory struct { //A struct for the collection of Racks
	Mutex sync.Mutex
	Racks map[string]*rack
}

var dbInventory *DBInventory

func InitDBInventory() error {
	if dbInventory == nil {
		dbInventory = &DBInventory{
			Mutex: sync.Mutex{},
			Racks: make(map[string]*rack),
		}
		dbInventory.Racks["rack1"] = &rack{ //

			tor:    Tor{},
			pdu:    Pdu{},
			blades: make(map[int64]*blade),
		}
		dbInventory.Racks["rack2"] = &rack{ //

			tor:    Tor{},
			pdu:    Pdu{},
			blades: make(map[int64]*blade),
		}
		dbInventory.Racks["rack1"].blades[1] = &blade{} //First blade for rack 1.
		dbInventory.Racks["rack1"].blades[2] = &blade{} //Second blade for rack 1.

		dbInventory.Racks["rack2"].blades[1] = &blade{} //First blade for rack 2.
		dbInventory.Racks["rack2"].blades[2] = &blade{} //Second blade for rack 2.

	}

	return nil
}

// Scan the set of known users in the store, invoking the supplied
// function with each entry.
func (m *DBInventory) Scan(action func(entry string) error) error {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	for name := range dbInventory.Racks {
		if err := action(name); err != nil {
			return err
		}
	}

	return nil
}
