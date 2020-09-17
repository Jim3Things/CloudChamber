// This module encapsulates storage and retrieval of known inventory
//
// Racks are held in a zone.  The zone has the racks, and a memoized summary of
// the maximum number of blades held in any rack.
//
// The full rack contains attributes about tor, pdU and blades.  It also has a
// memoized summary of the maximum capacity values.
//
// The memoized values are used by callers, such as the Cloud Chamber UI, to
// quickly shape the dimensions of the inventory display

// Each rack has an associated key which is the string name of the rack.

package frontend

import (
	"fmt"
	"math/rand"
	"sync"

	"github.com/Jim3Things/CloudChamber/pkg/protos/common"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/inventory"
)

// DBInventory is a structure used to establish synchronized access to
// the in-memory version of the inventory.
//
// The structure consists of three parts:
// 	- the Mutex field controls access to the rest of the structure, avoiding
// 	  collisions between two simultaneous web server calls.
//
// 	- the Zone field contains the inventory definition data.  This data
//    originated externally, via a definition file, and is read only.
//
//  - the remains fields, MaxBladeCount and MaxCapacity, contain summary memo
// 	  calculations that simplify the sizing and placement calculations for the
//	  Cloud Chamber UI.
//
type DBInventory struct {
	Mutex sync.Mutex

	Zone *pb.ExternalZone

	MaxBladeCount int64
	MaxCapacity   *common.BladeCapacity
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
			Zone: &pb.ExternalZone{
				Racks: make(map[string]*pb.ExternalRack),
			},
			MaxBladeCount: 0,
			MaxCapacity:   &common.BladeCapacity{},
		}

		dbInventory.Zone.Racks["rack1"] = &pb.ExternalRack{
			Tor:    &pb.ExternalTor{},
			Pdu:    &pb.ExternalPdu{},
			Blades: make(map[int64]*common.BladeCapacity),
		}

		dbInventory.Zone.Racks["rack2"] = &pb.ExternalRack{
			Tor:    &pb.ExternalTor{},
			Pdu:    &pb.ExternalPdu{},
			Blades: make(map[int64]*common.BladeCapacity),
		}

		// First blade for rack 1.
		dbInventory.Zone.Racks["rack1"].Blades[1] = &common.BladeCapacity{
			Cores:                  8,
			MemoryInMb:             16384,
			DiskInGb:               120,
			NetworkBandwidthInMbps: 1024,
			Arch:                   "X64",
		}

		// Second blade for rack 1.
		dbInventory.Zone.Racks["rack1"].Blades[2] = &common.BladeCapacity{
			Cores:                  16,
			MemoryInMb:             16384,
			DiskInGb:               240,
			NetworkBandwidthInMbps: 2048,
			Arch:                   "X64",
		}

		// First blade for rack 2.
		dbInventory.Zone.Racks["rack2"].Blades[1] = &common.BladeCapacity{
			Cores:                  24,
			MemoryInMb:             16384,
			DiskInGb:               120,
			NetworkBandwidthInMbps: 1024,
			Arch:                   "X64",
		}

		// Second blade for rack 2.
		dbInventory.Zone.Racks["rack2"].Blades[2] = &common.BladeCapacity{
			Cores:                  32,
			MemoryInMb:             16384,
			DiskInGb:               120,
			NetworkBandwidthInMbps: 1024,
			Arch:                   "X64",
		}

		// Add some more arbitrary inventory
		//
		for r := 3; r <= 4; r++ {

			rack := fmt.Sprintf("test%d", r)

			dbInventory.Zone.Racks[rack] = &pb.ExternalRack{
				Tor:    &pb.ExternalTor{},
				Pdu:    &pb.ExternalPdu{},
				Blades: make(map[int64]*common.BladeCapacity),
			}

			for b := 1; b <= 8; b++ {

				dbInventory.Zone.Racks[rack].Blades[int64(b)] = &common.BladeCapacity{
					Cores:                  3 * rand.Int63n(16),
					MemoryInMb:             1024 * rand.Int63n(128),
					DiskInGb:               1024 * rand.Int63n(16),
					NetworkBandwidthInMbps: 10 * 1024,
					Arch:                   "X64",
				}
			}
		}

		for r := 5; r <= 8; r++ {

			rack := fmt.Sprintf("test%d", r)

			dbInventory.Zone.Racks[rack] = &pb.ExternalRack{
				Tor:    &pb.ExternalTor{},
				Pdu:    &pb.ExternalPdu{},
				Blades: make(map[int64]*common.BladeCapacity),
			}

			for b := 1; b <= 8; b++ {

				dbInventory.Zone.Racks[rack].Blades[int64(b)] = &common.BladeCapacity{
					Cores:                  48,
					MemoryInMb:             128 * 1024,
					DiskInGb:               16 * 1024,
					NetworkBandwidthInMbps: 10 * 1024,
					Arch:                   "X64",
				}
			}
		}

		dbInventory.buildSummary()
	}

	return nil
}

// GetMemoData returns the maximum number of blades held in any rack
// in the inventory.
func (m *DBInventory) GetMemoData() (int64, *common.BladeCapacity) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	return m.MaxBladeCount, m.MaxCapacity
}

// Scan the set of known blades the store, invoking the supplied
// function with each entry.
func (m *DBInventory) Scan(action func(entry string) error) error {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	for name := range dbInventory.Zone.Racks {
		if err := action(name); err != nil {
			return err
		}
	}

	return nil
}

// Get returns the rack details to match the supplied rackID
func (m *DBInventory) Get(rackID string) (*pb.ExternalRack, error) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	r, ok := m.Zone.Racks[rackID]
	if !ok {
		return nil, NewErrRackNotFound(rackID)
	}

	return r, nil
}

// ScanBladesInRack enumerates over all the blades in a rack of the given
// rackID, and invokes the supplied action on each discovered bladeID in
// turn.
func (m *DBInventory) ScanBladesInRack(rackID string, action func(bladeID int64) error) error {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	r, ok := m.Zone.Racks[rackID]
	if !ok {
		return NewErrRackNotFound(rackID)
	}

	for name := range r.Blades {
		if err := action(name); err != nil {
			return err
		}
	}

	return nil
}

// GetBlade returns the details of a blade matching the supplied rackID and
// bladeID
func (m *DBInventory) GetBlade(rackID string, bladeID int64) (*common.BladeCapacity, error) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	r, ok := m.Zone.Racks[rackID]
	if !ok {
		return nil, NewErrRackNotFound(rackID)
	}

	b, ok := r.Blades[bladeID]
	if !ok {
		return nil, NewErrBladeNotFound(rackID, bladeID)
	}

	return b, nil
}

// buildSummary constructs the memo-ed summary data for the zone.  This should
// be called whenever the configured inventory changes.
//
// This method assumes it is called from within a DBInventory function, and the
// required mutex is already held.
func (m *DBInventory) buildSummary() {
	m.MaxBladeCount = 0

	memo := &common.BladeCapacity{}
	for _, rack := range m.Zone.Racks {
		for _, blade := range rack.Blades {
			memo.Cores = maxInt64(memo.Cores, blade.Cores)
			memo.DiskInGb = maxInt64(memo.DiskInGb, blade.DiskInGb)
			memo.MemoryInMb = maxInt64(memo.MemoryInMb, blade.MemoryInMb)
			memo.NetworkBandwidthInMbps = maxInt64(
				memo.NetworkBandwidthInMbps,
				blade.NetworkBandwidthInMbps)
		}

		m.MaxBladeCount = maxInt64(
			m.MaxBladeCount,
			int64(len(rack.Blades)))
	}

	m.MaxCapacity = memo
}

// maxInt64 is a helper function to return the maximum of two int64 values
func maxInt64(a int64, b int64) int64 {
	if a < b {
		return b
	}

	return a
}


//Cloud chamber config file needs to have a setting that point to inventory.yaml file
//We need dbinventory to create the inventory via the inventory reader