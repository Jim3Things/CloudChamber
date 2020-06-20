// This module encapsulates storage and retrieval of known racks

// The full rack contains attributes about tor, pdU and blades.

// Each rack has an associated key which is the lowercased form of the
// username.  The supplied name is retained as an attribute in order to present
// the form that the caller originally used for display purposes.

package frontend

import (
	"errors"
	"sync"

	common "github.com/Jim3Things/CloudChamber/pkg/protos/common"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/inventory"
)

// DBInventory is a container used to establish synchronized access to
// the in-memory set of Racks records.

type DBInventory struct { //A struct for the collection of Racks
	Mutex sync.Mutex
	Racks map[string]*pb.ExternalRack
}

var dbInventory *DBInventory

func InitDBInventory() error {
	if dbInventory == nil {
		dbInventory = &DBInventory{
			Mutex: sync.Mutex{},
			Racks: make(map[string]*pb.ExternalRack),
		}
		dbInventory.Racks["rack1"] = &pb.ExternalRack{ //

			Tor:    &pb.ExternalTor{},
			Pdu:    &pb.ExternalPdu{},
			Blades: make(map[int64]*common.BladeCapacity),
		}
		dbInventory.Racks["rack2"] = &pb.ExternalRack{ //

			Tor:    &pb.ExternalTor{},
			Pdu:    &pb.ExternalPdu{},
			Blades: make(map[int64]*common.BladeCapacity),
		}
		dbInventory.Racks["rack1"].Blades[1] = &common.BladeCapacity{} //First blade for rack 1.
		dbInventory.Racks["rack1"].Blades[2] = &common.BladeCapacity{} //Second blade for rack 1.

		dbInventory.Racks["rack2"].Blades[1] = &common.BladeCapacity{} //First blade for rack 2.
		dbInventory.Racks["rack2"].Blades[2] = &common.BladeCapacity{} //Second blade for rack 2.

	}

	return nil
}

// Scan the set of known blades the store, invoking the supplied
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

func (m *DBInventory) Get(rackid string) (*pb.ExternalRack, error) {

	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	r, ok := m.Racks[rackid]
	if !ok {
		return nil, errors.New("Rack not found")
	}
	return r, nil

}
