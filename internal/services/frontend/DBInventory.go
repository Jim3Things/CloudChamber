// This module encapsulates storage and retrieval of known racks

// The full rack contains attributes about tor, pdU and blades.

// Each rack has an associated key which is the lowercased form of the
// username.  The supplied name is retained as an attribute in order to present
// the form that the caller originally used for display purposes.

package frontend

import (
	"sync"

	common "github.com/Jim3Things/CloudChamber/pkg/protos/common"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/inventory"
)

// DBInventory is a container used to establish synchronized access to
// the in-memory set of Racks records.
//
type DBInventory struct { //A struct for the collection of Racks
	Mutex sync.Mutex
	Racks map[string]*pb.ExternalRack
}

var dbInventory *DBInventory

// InitDBInventory initializes the base state for the inventory.
//
// At present the primary state is sufficient data in an in-memory db sufficient
// for testing purposes. Eventually, this will be removed and the calls will be
// connected to the store in order to persist the inventory read from an external
// definition file
//
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
		dbInventory.Racks["rack1"].Blades[1] = &common.BladeCapacity{
			Cores:                  8,
			MemoryInMb:             16384,
			DiskInGb:               120,
			NetworkBandwidthInMbps: 1024,
			Arch:                   "X64",
		} //First blade for rack 1.

		dbInventory.Racks["rack1"].Blades[2] = &common.BladeCapacity{
			Cores:                  16,
			MemoryInMb:             16384,
			DiskInGb:               240,
			NetworkBandwidthInMbps: 2048,
			Arch:                   "X64",
		} //Second blade for rack 1.

		dbInventory.Racks["rack2"].Blades[1] = &common.BladeCapacity{
			Cores:                  24,
			MemoryInMb:             16384,
			DiskInGb:               120,
			NetworkBandwidthInMbps: 1024,
			Arch:                   "X64",
		} //First blade for rack 2.
		dbInventory.Racks["rack2"].Blades[2] = &common.BladeCapacity{
			Cores:                  32,
			MemoryInMb:             16384,
			DiskInGb:               120,
			NetworkBandwidthInMbps: 1024,
			Arch:                   "X64",
		} //Second blade for rack 2.

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

// Get returns the rack details to match the supplied rackId
//
func (m *DBInventory) Get(rackid string) (*pb.ExternalRack, error) {

	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	r, ok := m.Racks[rackid]
	if !ok {
		return nil, NewErrRackNotFound(rackid)
	}
	return r, nil

}

// ScanBladesInRack enumerates over all the blades in a rack of the
// given rackId, and invokes the supplied action on each discovered
// bladeId in turn.
//
func (m *DBInventory) ScanBladesInRack(rackid string, action func(bladeid int64) error) error {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	r, ok := m.Racks[rackid]
	if !ok {
		return NewErrRackNotFound(rackid)
	}
	for name := range r.Blades {
		if err := action(name); err != nil {
			return err
		}
	}
	return nil
}

// GetBlade returns the details of a blade matching the
// supplied rackId and bladeId
//
func (m *DBInventory) GetBlade(rackid string, bladeid int64) (*common.BladeCapacity, error) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	r, ok := m.Racks[rackid]
	if !ok {
		return nil, NewErrRackNotFound(rackid)
	}

	b, ok := r.Blades[bladeid]
	if !ok {
		return nil, NewErrBladeNotFound(rackid, bladeid)
	}
	return b, nil
}
