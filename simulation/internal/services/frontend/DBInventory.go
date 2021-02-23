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
	"context"
	"fmt"
	"sync"

	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/inventory"
	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/store"
	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/timestamp"
	"github.com/Jim3Things/CloudChamber/simulation/internal/common"
	"github.com/Jim3Things/CloudChamber/simulation/internal/config"
	"github.com/Jim3Things/CloudChamber/simulation/internal/tracing"
	"github.com/Jim3Things/CloudChamber/simulation/pkg/errors"
	pb "github.com/Jim3Things/CloudChamber/simulation/pkg/protos/inventory"
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
	mutex sync.RWMutex

	Started       bool

	Store              *store.Store
	cfg                *config.GlobalConfig
	RootSummary        *RootSummary
	DefaultZoneSummary *ZoneSummary
	actualLoaded       bool
	Actual             *actualZone
}

// ZoneSummary contains the summary data for a single zone. The contents
// are either zero or are (re-)computed whenever the inventory definitions
// are loaded from file. They are a cache of data in the store to avoid
// having to scan the entire zone whenever a simple query for the basic
// data is being handled.
//
type ZoneSummary struct {
	RackCount     int
	MaxBladeCount int
	MaxCapacity   *pb.BladeCapacity
}

// RegionSummary contains the summary data for a single region. The contents
// are either zero or are (re-)computed whenever the inventory definitions
// are loaded from file. They are a cache of data in the store to avoid
// having to scan the entire region whenever a simple query for the basic
// data is being handled.
//
type RegionSummary struct {
	ZoneCount     int
	MaxRackCount  int
	MaxBladeCount int
	MaxCapacity   *pb.BladeCapacity
}

// RootSummary contains the summary data for the entire invnetory. The contents
// are either zero or are (re-)computed whenever the inventory definitions
// are loaded from file. They are a cache of data in the store to avoid
// having to scan the entire zone whenever a simple query for the basic
// data is being handled.
//
type RootSummary struct {
	RegionCount   int
	MaxZoneCount  int
	MaxRackCount  int
	MaxBladeCount int
	MaxCapacity   *pb.BladeCapacity
}

var dbInventory *DBInventory

// While the transition to the new inventory extended schemaa continues, the 
// underlying store records use the following constants for determining values
// for the previously non-existent fields. Eventually these will dissapear as
// the front-end and higher layers learn abouts regions, zones, multiple pdus
// and tors.
//
const (
	defaultRegion = "standard"
	defaultZone   = "standard"
	defaultPdu    = int64(0)
	defaultTor    = int64(0)
)

// InitDBInventory initializes the base state for the inventory.
//
func InitDBInventory(ctx context.Context, cfg *config.GlobalConfig) (err error) {
	if dbInventory == nil {
		db := NewDbInventory()

		if err = db.Initialize(ctx, cfg); err != nil {
			return err
		}

		if dbInventory == nil {
			dbInventory = db
		}
	}

	return nil
}

// NewDbInventory is a helper routine to construct an empty DBInventory structure
// as a convenience to avoid clients having to do all the details themselves.
//
func NewDbInventory() *DBInventory {
	return  &DBInventory{
		mutex:              sync.RWMutex{},
		Started:            false,
		Store:              store.NewStore(),
		cfg:                &config.GlobalConfig{},
		RootSummary:        &RootSummary{},
		DefaultZoneSummary: &ZoneSummary{
			MaxCapacity:   &pb.BladeCapacity{
				Accelerators: []*pb.Accelerator{},
			},
		},
		actualLoaded:       false,
		Actual:             &actualZone{Racks: make(map[string]*actualRack)},
	}
}

// Initialize initializes an existing, but currently un-initialized DB Inventory
// structure. This involves connecting to the store.
//
// The inventory is not loaded here as the loading process triggers a large
// number of trace spans to be created and sent to the trace channel but it's
// possible that the trace sink is not yet available. While this is unlikely
// to be a big problem in the production code as the trace sink will show
// up soon, for testing it definitily creates an issue as the trace sink is
// often not opened up until the per-test setup phase rather than the suite
// setup (when using the Testify package) leading to the test stalling
// during initialization.
//
func (m *DBInventory) Initialize(ctx context.Context, cfg *config.GlobalConfig) (err error) {
	ctx, span := tracing.StartSpan(ctx,
		tracing.WithName("Initialize Inventory DB state"),
		tracing.WithContextValue(timestamp.EnsureTickInContext),
		tracing.AsInternal())
	defer span.End()

	m.cfg = cfg

	if err = m.Store.Connect(); err != nil {
		return err
	}

	return nil
}

// StartDBInventory is a 
//
func StartDBInventory(ctx context.Context) error {
	return dbInventory.Start(ctx)
}

// Start is a method used to start the inventory service as a part of normal
// product code paths loading the current inventory definition from the
// definition file, and sending updates to the store to reconcile the
// persisted state. Note that changes to the store may in turn trigger any
// currently established watch handlers leading to updates elsewhere
// in the system.
//
// NOTE: this method is not expected to be called as part of initialization 
//       when running tests so as to avoid unfortunate interactions with
//       the trace code as a result of generating lots of trace spans before
//       the trace sink client is available.
//
func (m *DBInventory) Start(ctx context.Context) error {
	ctx, span := tracing.StartSpan(ctx,
		tracing.WithName("Start Inventory DB"),
		tracing.WithContextValue(timestamp.EnsureTickInContext),
		tracing.AsInternal())
	defer span.End()
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.Started {
		if err := m.UpdateInventoryDefinition(ctx, m.cfg.Inventory.InventoryDefinition); err != nil {
			return err
		}
	}

	if err := m.LoadInventoryActual(true); err != nil {
		return err
	}

	return nil
}

// UpdateInventoryDefinition is a method to load a new inventory definition from
// the configured file. Once read, the store will be updated with the differences
// which will in turn trigger a set of previously established watch routines to
// issue a number of arrival and/or departure notifications.
//
func (m *DBInventory) UpdateInventoryDefinition(ctx context.Context, path string) error {
	ctx, span := tracing.StartSpan(ctx,
		tracing.WithName("Update inventory definition from file"),
		tracing.WithContextValue(timestamp.EnsureTickInContext),
		tracing.AsInternal())
	defer span.End()

	// We have the basic initialization done. Now go read the current inventory
	// from the file indicated by the configuration. Once we have that, load it
	// into the store looking to see if there are any material changes between
	// what is already in the store and what is now found in the file.
	//
	rootFile, err := inventory.ReadInventoryDefinitionFromFileEx(ctx, path)
	if err != nil {
		return err
	}

	rootStore, err := m.readInventoryDefinitionFromStore(ctx)

	if err != nil {
		return err
	}

	if err = m.reconcileNewInventory(ctx, rootFile, rootStore); err != nil {
		return err
	}

	return nil
}

// DeleteInventoryDefinition is a
//
func (m *DBInventory) DeleteInventoryDefinition(ctx context.Context) error {
	ctx, span := tracing.StartSpan(ctx,
		tracing.WithName("Delete inventory definition from store"),
		tracing.WithContextValue(timestamp.EnsureTickInContext),
		tracing.AsInternal())
	defer span.End()

	root, err := m.readInventoryDefinitionFromStore(ctx)
	if err != nil {
		return err
	}

	err = m.deleteInventoryDefinitionFromStore(ctx, root)
	if err != nil {
		return err
	}

	return nil
}

// reconcileNewInventory compares the newly loaded inventory definition,
// presumably from a configuration file, with the currently loaded inventory
// and updates the store accordingly. This will trigger the various watches
// which any currently running services have previously established and deliver
// a set of arrival and/or departure notifications as appropriate.
//
// NOTE: As a temporary measure, reconciliation just deletes the old inventory
//       from the store and completely replaces it with the newly read inventory
//       from the configured file
//
func (m *DBInventory) reconcileNewInventory(
	ctx context.Context,
	rootFile *pb.Definition_Root,
	rootStore *pb.Definition_Root) error {
	ctx, span := tracing.StartSpan(ctx,
		tracing.WithName("Reconcile current inventory with update"),
		tracing.WithContextValue(timestamp.EnsureTickInContext),
		tracing.AsInternal())
	defer span.End()

	m.mutex.Lock()
	defer m.mutex.Unlock()

	err := m.deleteInventoryDefinitionFromStore(ctx, rootStore)
	if err != nil {
		return err
	}

	err = m.writeInventoryDefinitionToStore(ctx, rootFile)
	if err != nil {
		return err
	}

	m.RootSummary = m.buildSummaryForRoot(rootFile)

	tracing.Info(
		ctx,
		"   Updated inventory summary - RegionCount: %d MaxZoneCount: %d MaxRackCount: %d MaxBladeCount: %d MaxCapacity: %v",
		m.RootSummary.RegionCount,
		m.RootSummary.MaxZoneCount,
		m.RootSummary.MaxRackCount,
		m.RootSummary.MaxBladeCount,
		&m.RootSummary.MaxCapacity)


	zone, err := m.getDefaultZone(rootFile)

	if err != nil {
		m.DefaultZoneSummary = &ZoneSummary{}
	
		tracing.Error(
			ctx,
			"   Reset DEFAULT inventory summary - MaxRackCount: %d MaxBladeCount: %d MaxCapacity: %v - %v",
			m.DefaultZoneSummary.RackCount,
			m.DefaultZoneSummary.MaxBladeCount,
			&m.DefaultZoneSummary.MaxCapacity,
			err,
		)
	} else {
		m.DefaultZoneSummary = m.buildSummaryForZone(zone)

		tracing.Info(
			ctx,
			"   Updated DEFAULT inventory summary - MaxRackCount: %d MaxBladeCount: %d MaxCapacity: %v",
			m.DefaultZoneSummary.RackCount,
			m.DefaultZoneSummary.MaxBladeCount,
			&m.DefaultZoneSummary.MaxCapacity,
		)
	}

	return nil
}

func (m *DBInventory) getDefaultZone(root *pb.Definition_Root) (*pb.Definition_Zone, error) {

	region, ok := root.Regions[defaultRegion]

	if !ok {
		return nil, errors.ErrRegionNotFound{Region: defaultRegion}
	}

	zone, ok := region.Zones[defaultZone]

	if !ok {
		return nil, errors.ErrZoneNotFound{Region: defaultRegion, Zone: defaultZone}
	}

	return zone, nil
}

// readInventoryDefinitionFromStore is used to read all the inventory
// definitions from the store, regardless of how they got there and 
// return them in the hierarchical in-memory form.
//
func (m *DBInventory) readInventoryDefinitionFromStore(ctx context.Context) (*pb.Definition_Root, error) {
	ctx, span := tracing.StartSpan(ctx,
		tracing.WithName("Read inventory definition from store"),
		tracing.WithContextValue(timestamp.EnsureTickInContext),
		tracing.AsInternal())
	defer span.End()

	root, err := inventory.NewRoot(ctx, m.Store, inventory.DefinitionTable)
	if err != nil {
		return nil, err
	}

	_, regions, err := root.FetchChildren(ctx)
	if err != nil {
		return nil, err
	}

	defRoot := &pb.Definition_Root{
		Details: root.GetDetails(ctx),
		Regions: make(map[string]*pb.Definition_Region, len(*regions)),
	}

	for regionName, region := range *regions {
		_, zones, err := region.FetchChildren(ctx)
		if err != nil {
			return nil, err
		}

		defRegion := &pb.Definition_Region{
			Details: region.GetDetails(ctx),
			Zones:   make(map[string]*pb.Definition_Zone, len(*zones)),
		}

		for zoneName, zone := range *zones {
			_, racks, err := zone.FetchChildren(ctx)
			if err != nil {
				return nil, err
			}

			defZone := &pb.Definition_Zone{
				Details: zone.GetDetails(ctx),
				Racks:   make(map[string]*pb.Definition_Rack, len(*racks)),
			}

			for rackName, rack := range *racks {
				defRack, err := rack.Copy(ctx)
				if err != nil {
					return nil, err
				}
	
				defZone.Racks[rackName] = defRack
			}

			defRegion.Zones[zoneName] = defZone
		}

		defRoot.Regions[regionName] = defRegion
	}

	return defRoot, nil
}

// writeInventoryDefinitionToStore will use the supplied root parameter and 
// persist a record for each item in the supplied Definition_Root contents.
//
// In the case of collisions, the routine will fail with an "already exists"
// error.
//
func (m *DBInventory) writeInventoryDefinitionToStore(ctx context.Context, root *pb.Definition_Root) error {
	ctx, span := tracing.StartSpan(ctx,
		tracing.WithName("Write inventory definition to store"),
		tracing.WithContextValue(timestamp.EnsureTickInContext),
		tracing.AsInternal())
	defer span.End()

	storeRoot, err := inventory.NewRoot(ctx, m.Store, inventory.DefinitionTable)
	if err != nil {
		return err
	}

	storeRoot.SetDetails(ctx, root.Details)

	for regionName, region := range root.Regions {
		ctx, span := tracing.StartSpan(ctx,
			tracing.WithName(fmt.Sprintf("Write inventory definition for region %q", regionName)),
			tracing.WithContextValue(timestamp.EnsureTickInContext),
			tracing.AsInternal())
		defer span.End()

		storeRegion, err := storeRoot.NewChild(ctx, regionName)
		if err != nil {
			return err
		}

		storeRegion.SetDetails(ctx, region.GetDetails())

		_, err = storeRegion.Create(ctx)
		if err != nil {
			return err
		}

		for zoneName, zone := range region.Zones {
			ctx, span := tracing.StartSpan(ctx,
				tracing.WithName(fmt.Sprintf("Write inventory definition for region %q, zone %q", regionName, zoneName)),
				tracing.WithContextValue(timestamp.EnsureTickInContext),
				tracing.AsInternal())
			defer span.End()

			storeZone, err := storeRegion.NewChild(ctx, zoneName)
			if err != nil {
				return err
			}
	
			storeZone.SetDetails(ctx, zone.GetDetails())
	
			_, err = storeZone.Create(ctx)
			if err != nil {
				return err
			}

			for rackName, rack := range zone.Racks {
				ctx, span := tracing.StartSpan(ctx,
					tracing.WithName(fmt.Sprintf("Write inventory definition for region %q, zone %q, rack %q", regionName, zoneName, rackName)),
					tracing.WithContextValue(timestamp.EnsureTickInContext),
					tracing.AsInternal())
				defer span.End()

				storeRack, err := storeZone.NewChild(ctx, rackName)
				if err != nil {
					return err
				}
		
				storeRack.SetDetails(ctx, rack.GetDetails())
		
				_, err = storeRack.Create(ctx)
				if err != nil {
					return err
				}

				for index, pdu := range rack.Pdus {
					ctx, span := tracing.StartSpan(ctx,
						tracing.WithName(fmt.Sprintf("Write inventory definition for region %q, zone %q, rack %q, pdu %d", regionName, zoneName, rackName, index)),
						tracing.WithContextValue(timestamp.EnsureTickInContext),
						tracing.AsInternal())
					defer span.End()

					storePdu, err := storeRack.NewPdu(ctx, index)
					if err != nil {
						return err
					}
					ports := pdu.GetPorts()

					storePdu.SetDetails(ctx, pdu.GetDetails())
					storePdu.SetPorts(ctx, &ports)

					_, err = storePdu.Create(ctx)
					if err != nil {
						return err
					}
				}

				for index, tor := range rack.Tors {
					ctx, span := tracing.StartSpan(ctx,
						tracing.WithName(fmt.Sprintf("Write inventory definition for region %q, zone %q, rack %q, tor %d", regionName, zoneName, rackName, index)),
						tracing.WithContextValue(timestamp.EnsureTickInContext),
						tracing.AsInternal())
					defer span.End()

					storeTor, err := storeRack.NewTor(ctx, index)
					if err != nil {
						return err
					}

					ports := tor.GetPorts()
					storeTor.SetDetails(ctx, tor.GetDetails())
					storeTor.SetPorts(ctx, &ports)

					_, err = storeTor.Create(ctx)
					if err != nil {
						return err
					}
				}

				for index, blade := range rack.Blades {
					ctx, span := tracing.StartSpan(ctx,
						tracing.WithName(fmt.Sprintf("Write inventory definition for region %q, zone %q, rack %q, blade %d", regionName, zoneName, rackName, index)),
						tracing.WithContextValue(timestamp.EnsureTickInContext),
						tracing.AsInternal())
					defer span.End()

					storeBlade, err := storeRack.NewBlade(ctx, index)
					if err != nil {
						return err
					}

					storeBlade.SetDetails(ctx, blade.GetDetails())
					storeBlade.SetCapacity(ctx, blade.GetCapacity())
					storeBlade.SetBootInfo(ctx, blade.BootOnPowerOn, blade.GetBootInfo())

					_, err = storeBlade.Create(ctx)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

// deleteInventoryDefinitionFromStore is used to completely remove all
// inventory definitions from the store as identified by the storeRoot
// parameter, regardless of how they got there.
//
func (m *DBInventory) deleteInventoryDefinitionFromStore(ctx context.Context, storeRoot *pb.Definition_Root) error {

	root, err := inventory.NewRoot(ctx, m.Store, inventory.DefinitionTable)
	if err != nil {
		return err
	}

	for regionName, storeRegion := range storeRoot.Regions {

		region, err := root.NewChild(ctx, regionName)
		if err != nil {
			return err
		}

		for zoneName, storeZone := range storeRegion.Zones {

			zone, err := region.NewChild(ctx, zoneName)
			if err != nil {
				return err
			}

			for rackName, storeRack := range storeZone.Racks {

				rack, err := zone.NewChild(ctx, rackName)
				if err != nil {
					return err
				}

				for i := range storeRack.Pdus {

					pdu, err := rack.NewPdu(ctx, i)
					if err != nil {
						return err
					}

					pdu.Delete(ctx, true)
				}

				for i := range storeRack.Tors {

					tor, err := rack.NewTor(ctx, i)
					if err != nil {
						return err
					}

					tor.Delete(ctx, true)
				}

				for i := range storeRack.Blades {

					blade, err := rack.NewBlade(ctx, i)
					if err != nil {
						return err
					}

					blade.Delete(ctx, true)
				}

			rack.Delete(ctx, true)
			}

		zone.Delete(ctx, true)
		}

	region.Delete(ctx, true)
	}

	return nil
}

// GetMemoData returns the maximum number of blades held in any rack
// in the inventory.
func (m *DBInventory) GetMemoData() *ZoneSummary {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.DefaultZoneSummary
}

// ScanRegions scans the set of known regions in the store, invoking the supplied
// function with each entry.
//
func (m *DBInventory) ScanRegions(action func(entry string) error) error {

	ctx := context.Background()

	root, err := inventory.NewRoot(ctx, m.Store, inventory.DefinitionTable)
	if err != nil {
		return m.transformError(err, "", "", "", 0)
	}

	_, regions, err := root.FetchChildren(ctx)
	if err != nil {
		return m.transformError(err, "", "", "", 0)
	}

	for name := range *regions {
		if err := action(name); err != nil {
			return err
		}
	}

	return nil
}

// GetRegion returns the region details to match the supplied region name
//
func (m *DBInventory) GetRegion(regionName string) (*pb.Definition_Region, error) {

	ctx := context.Background()

	region, err := inventory.NewRegion(ctx, m.Store, inventory.DefinitionTable, regionName)
	if err != nil {
		return nil, m.transformError(err, regionName, "", "", 0)
	}

	_, err = region.Read(ctx)
	if err != nil {
		return nil, m.transformError(err, regionName, "", "", 0)
	}

	r := &pb.Definition_Region{
		Details: region.GetDetails(ctx),
		Zones: make(map[string]*pb.Definition_Zone),
	}

	return r, nil
}

// ScanZonesInRegion scans the set of known zones in the store within
// the specified region, invoking the supplied function with each entry.
//
func (m *DBInventory) ScanZonesInRegion(regionName string, action func(entry string) error) error {

	ctx := context.Background()

	root, err := inventory.NewRoot(ctx, m.Store, inventory.DefinitionTable)
	if err != nil {
		return m.transformError(err, regionName, "", "", 0)
	}

	region, err := root.NewChild(ctx, regionName)
	if err != nil {
		return m.transformError(err, regionName, "", "", 0)
	}

	_, names, err := region.ListChildren(ctx)
	if err != nil {
		return m.transformError(err, regionName, "", "", 0)
	}

	for _, name := range names {
		if err := action(name); err != nil {
			return err
		}
	}

	return nil
}

// GetZone returns the zone details to match the supplied rackID
//
func (m *DBInventory) GetZone(regionName string, zoneName string) (*pb.Definition_Zone, error) {

	ctx := context.Background()

	zone, err := inventory.NewZone(ctx, m.Store, inventory.DefinitionTable, regionName, zoneName)
	if err != nil {
		return nil, m.transformError(err, regionName, zoneName, "", 0)
	}

	_, err = zone.Read(ctx)
	if err != nil {
		return nil, m.transformError(err, regionName, zoneName, "", 0)
	}

	z := &pb.Definition_Zone{
		Details: zone.GetDetails(ctx),
		Racks: make(map[string]*pb.Definition_Rack),
	}

	return z, nil
}


// ScanRacksInZone scans the set of known racks in the store, invoking the supplied
// function with each entry.
//
func (m *DBInventory) ScanRacksInZone(regionName string, zoneName string, action func(entry string) error) error {
	
	ctx := context.Background()

	zone, err := inventory.NewZone(ctx, m.Store, inventory.DefinitionTable, regionName, zoneName)

	if err != nil {
		return m.transformError(err, regionName, zoneName, "", 0)
	}

	_, names, err := zone.ListChildren(ctx)
	if err != nil {
		return m.transformError(err, regionName, zoneName, "", 0)
	}

	for _, name := range names {
		if err := action(name); err != nil {
			return err
		}
	}

	return nil
}

// GetRackInZone returns the rack details to match the supplied rackName
//
func (m *DBInventory) GetRackInZone(regionName string, zoneName string, rackName string) (*pb.Definition_Rack, error) {

	ctx := context.Background()

	rack, err := inventory.NewRack(ctx, m.Store, inventory.DefinitionTable, regionName, zoneName, rackName)

	_, err = rack.Read(ctx)
	if err != nil {
		return nil, m.transformError(err, regionName, zoneName, rackName, 0)
	}

	_, pdus, err := rack.FetchPdus(ctx)
	if err != nil {
		return nil, m.transformError(err, regionName, zoneName, rackName, 0)
	}

	_, tors, err := rack.FetchTors(ctx)
	if err != nil {
		return nil, m.transformError(err, regionName, zoneName, rackName, 0)
	}

	_, blades, err := rack.FetchBlades(ctx)
	if err != nil {
		return nil, m.transformError(err, regionName, zoneName, rackName, 0)
	}

	r := &pb.Definition_Rack{
		Details: rack.GetDetails(ctx),
		Pdus:    make(map[int64]*pb.Definition_Pdu),
		Tors:    make(map[int64]*pb.Definition_Tor),
		Blades:  make(map[int64]*pb.Definition_Blade),
	}

	for index, pdu := range *pdus {
		r.Pdus[index] = pdu.Copy(ctx)
	}

	for index, tor := range *tors {
		r.Tors[index] = tor.Copy(ctx)
	}

	for index, blade := range *blades {
		r.Blades[index] = blade.Copy(ctx)
	}

	return r, nil
}

// GetRack returns the rack details to match the supplied rackID
//
// By convention, during the transition, the routines which provide the
// External_Xxxx style structures *REQUIRE* that there be a region named
// "standard", there be a zone within that region named "standard", in
// each of the racks there must be a Pdu with id 0 and a Tor with Id 0.
//
// All other regions, zones, pdus and tors will be ignored.
//
func (m *DBInventory) GetRack(rackName string) (*pb.External_Rack, error) {

	ctx := context.Background()

	rack, err := inventory.NewRack(ctx, m.Store, inventory.DefinitionTable, defaultRegion, defaultZone, rackName)
	if err != nil {
		return nil, m.transformError(err, defaultRegion, defaultZone, rackName, 0)
	}

	_, err = rack.Read(ctx)
	if err != nil {
		return nil, m.transformError(err, defaultRegion, defaultZone, rackName, 0)
	}

	_, blades, err := rack.FetchBlades(ctx)
	if err != nil {
		return nil, m.transformError(err, defaultRegion, defaultZone, rackName, 0)
	}

	bladeMap := make(map[int64]*pb.BladeCapacity)

	for index, blade := range *blades {
		bladeMap[index] = blade.GetCapacity(ctx)
	}

	r := &pb.External_Rack{
		Pdu:    &pb.External_Pdu{},
		Tor:    &pb.External_Tor{},
		Blades: bladeMap,
	}

	return r, nil
}

// ScanBladesInRack enumerates over all the blades in a rack of the given
// rackID, and invokes the supplied action on each discovered bladeID in
// turn.
func (m *DBInventory) ScanBladesInRack(rackName string, action func(bladeID int64) error) error {

	ctx := context.Background()

	rack, err := inventory.NewRack(ctx, m.Store, inventory.DefinitionTable, defaultRegion, defaultZone, rackName)
	if err != nil {
		return m.transformError(err, defaultRegion, defaultZone, rackName, 0)
	}

	_, IDs, err := rack.ListBlades(ctx)
	if err != nil {
		return m.transformError(err, defaultRegion, defaultZone, rackName, 0)
	}

	for _, index := range IDs {
		if err := action(index); err != nil {
			return err
		}
	}

	return nil
}

// GetBlade returns the details of a blade matching the supplied rackID and
// bladeID
func (m *DBInventory) GetBlade(rackName string, bladeID int64) (*pb.BladeCapacity, error) {

	ctx := context.Background()

	blade, err := inventory.NewBlade(ctx, m.Store, inventory.DefinitionTable, defaultRegion, defaultZone, rackName, bladeID)
	if err != nil {
		return nil, m.transformError(err, defaultRegion, defaultZone, rackName, bladeID)
	}

	_, err = blade.Read(ctx)
	if err != nil {
		return nil, m.transformError(err, defaultRegion, defaultZone, rackName, bladeID)
	}

	return blade.GetCapacity(ctx), nil
}

func (m *DBInventory) buildSummaryForRack(rack *pb.Definition_Rack) (int, *pb.BladeCapacity) {

	memo := &pb.BladeCapacity{}

	for _, blade := range rack.Blades {
		memo.Cores = common.MaxInt64(memo.Cores, blade.Capacity.Cores)
		memo.DiskInGb = common.MaxInt64(memo.DiskInGb, blade.Capacity.DiskInGb)
		memo.MemoryInMb = common.MaxInt64(memo.MemoryInMb, blade.Capacity.MemoryInMb)
		memo.NetworkBandwidthInMbps = common.MaxInt64(
			memo.NetworkBandwidthInMbps,
			blade.Capacity.NetworkBandwidthInMbps)
	}

	return len(rack.Blades), memo
}

func (m *DBInventory) buildSummaryForZone(zone *pb.Definition_Zone) *ZoneSummary {

	maxCapacity := &pb.BladeCapacity{}
	maxBladeCount := int(0)

	for _, rack := range zone.Racks {
		bladeCount, capacity := m.buildSummaryForRack(rack)

		maxBladeCount = common.MaxInt(maxBladeCount, bladeCount)
		maxCapacity.Cores = common.MaxInt64(maxCapacity.Cores, capacity.Cores)
		maxCapacity.DiskInGb = common.MaxInt64(maxCapacity.DiskInGb, capacity.DiskInGb)
		maxCapacity.MemoryInMb = common.MaxInt64(maxCapacity.MemoryInMb, capacity.MemoryInMb)
		maxCapacity.NetworkBandwidthInMbps = common.MaxInt64(
			maxCapacity.NetworkBandwidthInMbps,
			capacity.NetworkBandwidthInMbps)
	}

	return &ZoneSummary{
		RackCount:     len(zone.Racks),
		MaxBladeCount: maxBladeCount,
		MaxCapacity:   maxCapacity,
	}
}

func (m *DBInventory) transformError(err error, region string, zone string, rack string, index int64) error {
	switch {
	case err == errors.ErrRegionNotFound{Region: region}:
		return NewErrRegionNotFound(rack)

	case err == errors.ErrZoneNotFound{Region: region, Zone: zone}:
		return NewErrZoneNotFound(rack)

	case err == errors.ErrRackNotFound{Region: region, Zone: zone, Rack: rack}:
		return NewErrRackNotFound(rack)

	case err == errors.ErrPduNotFound{Region: region, Zone: zone, Rack: rack, Pdu: index}:
		return NewErrPduNotFound(rack, index)

	case err == errors.ErrPduIndexNotFound{Region: region, Zone: zone, Rack: rack, Pdu: index}:
		return NewErrBladeNotFound(rack, index)

	case err == errors.ErrPduIDInvalid{Value: index, Limit: inventory.MaxPduID}:
		return NewErrBladeNotFound(rack, index)

	case err == errors.ErrPduIndexInvalid{Region: region, Zone: zone, Rack: rack, Pdu: fmt.Sprintf("%d", index)}:
		return NewErrBladeNotFound(rack, index)

	case err == errors.ErrTorNotFound{Region: region, Zone: zone, Rack: rack, Tor: index}:
		return NewErrTorNotFound(rack, index)

	case err == errors.ErrTorIndexNotFound{Region: region, Zone: zone, Rack: rack, Tor: index}:
		return NewErrBladeNotFound(rack, index)

	case err == errors.ErrTorIDInvalid{Value: index, Limit: inventory.MaxTorID}:
		return NewErrBladeNotFound(rack, index)

	case err == errors.ErrTorIndexInvalid{Region: region, Zone: zone, Rack: rack, Tor: fmt.Sprintf("%d", index)}:
		return NewErrBladeNotFound(rack, index)

	case err == errors.ErrBladeNotFound{Region: region, Zone: zone, Rack: rack, Blade: index}:
		return NewErrBladeNotFound(rack, index)

	case err == errors.ErrBladeIndexNotFound{Region: region, Zone: zone, Rack: rack, Blade: index}:
		return NewErrBladeNotFound(rack, index)

	case err == errors.ErrBladeIDInvalid{Value: index, Limit: inventory.MaxBladeID}:
		return NewErrBladeNotFound(rack, index)

	case err == errors.ErrBladeIndexInvalid{Region: region, Zone: zone, Rack: rack, Blade: fmt.Sprintf("%d", index)}:
		return NewErrBladeNotFound(rack, index)

	default:
		return err
	}
}

// buildSummary constructs the memo-ed summary data for the zone.  This should
// be called whenever the configured inventory changes. This includes
//
// - the zone count
// - the maximum number of blades in a rack
// - the memo data itself
//
func (m *DBInventory) buildSummaryForRegion(
	region *pb.Definition_Region) *RegionSummary {

	maxCapacity := &pb.BladeCapacity{}
	maxRackCount := int(0)
	maxBladeCount := int(0)

	for _, zone := range region.Zones {
		zoneSummary := m.buildSummaryForZone(zone)

		maxBladeCount = common.MaxInt(maxBladeCount, zoneSummary.MaxBladeCount)
		maxRackCount = common.MaxInt(maxRackCount, zoneSummary.RackCount)

		maxCapacity.Cores = common.MaxInt64(maxCapacity.Cores, zoneSummary.MaxCapacity.Cores)
		maxCapacity.DiskInGb = common.MaxInt64(maxCapacity.DiskInGb, zoneSummary.MaxCapacity.DiskInGb)
		maxCapacity.MemoryInMb = common.MaxInt64(maxCapacity.MemoryInMb, zoneSummary.MaxCapacity.MemoryInMb)

		maxCapacity.NetworkBandwidthInMbps = common.MaxInt64(
				maxCapacity.NetworkBandwidthInMbps,
				zoneSummary.MaxCapacity.NetworkBandwidthInMbps)

		maxBladeCount = common.MaxInt(maxBladeCount, zoneSummary.MaxBladeCount)
	}

	return &RegionSummary{
		ZoneCount:     len(region.Zones),
		MaxRackCount:  maxRackCount,
		MaxBladeCount: maxBladeCount,
		MaxCapacity:   maxCapacity,
	}
}

// buildSummaryForRoot constructs the memo-ed summary data for the root. This should
// be called whenever the configured inventory changes. This includes
//
// - the zone count
// - the maximum number of blades in a rack
// - the memo data itself
//
func (m *DBInventory) buildSummaryForRoot(
	root *pb.Definition_Root) *RootSummary {

	maxCapacity := &pb.BladeCapacity{}
	maxZoneCount := int(0)
	maxRackCount := int(0)
	maxBladeCount := int(0)

	for _, region := range root.Regions {
		regionSummary := m.buildSummaryForRegion(region)

		maxBladeCount = common.MaxInt(maxBladeCount, regionSummary.MaxBladeCount)
		maxRackCount = common.MaxInt(maxRackCount, regionSummary.MaxRackCount)
		maxZoneCount = common.MaxInt(maxZoneCount, regionSummary.ZoneCount)

		maxCapacity.Cores = common.MaxInt64(maxCapacity.Cores, regionSummary.MaxCapacity.Cores)
		maxCapacity.DiskInGb = common.MaxInt64(maxCapacity.DiskInGb, regionSummary.MaxCapacity.DiskInGb)
		maxCapacity.MemoryInMb = common.MaxInt64(maxCapacity.MemoryInMb, regionSummary.MaxCapacity.MemoryInMb)

		maxCapacity.NetworkBandwidthInMbps = common.MaxInt64(
				maxCapacity.NetworkBandwidthInMbps,
				regionSummary.MaxCapacity.NetworkBandwidthInMbps)
	}

	return &RootSummary{
		RegionCount:   len(root.Regions),
		MaxZoneCount:  maxZoneCount,
		MaxRackCount:  maxRackCount,
		MaxBladeCount: maxBladeCount,
		MaxCapacity:   maxCapacity,
	}
}

// Condition describes the current operational condition of an item in the inventory.
//
type Condition int

// Thw set of available conditions used to describe items in the inventory. Related
// to the conditions defined in inventory.proto
//
const (
	ConditionNotInService Condition = iota
	ConditionOperational
	ConditionBurnIn
	ConditionOutForRepair
	ConditionRetiring
	ConditionRetired
)

// DefinitionTor describes the revision and value for the request TOR
//
type DefinitionTor struct {
	revision int64
	tor      *pb.Definition_Tor
}

// DefinitionPdu describes the revision and value for the request PDU
//
type DefinitionPdu struct {
	revision int64
	pdu      *pb.Definition_Pdu
}

// DefinitionBlade describes the revision and value for the request blade
//
type DefinitionBlade struct {
	revision int64
	blade    *pb.Definition_Blade
}

// DefinitionRack describes the revision and value for the request rack
//
type DefinitionRack struct {
	revision int64
	rack     *pb.Definition_Rack
}

// DefinitionZone defines the set of values used to summarize the contents of a zone to
// allow a query on the zone without incorporating the entire inventory definition
//
type DefinitionZone struct {
	revision int64
	zone     *pb.Definition_Zone
}

// Generic options

// InventoryOption is a
//
type InventoryOption func(*InventoryOptions)

// InventoryOptions is a struct
//
type InventoryOptions struct {
	revision int64
}

func (options *InventoryOptions) applyOpts(optionsArray []InventoryOption) {
	for _, option := range optionsArray {
		option(options)
	}
}

// WithRevision is a
//
func WithRevision(rev int64) InventoryZoneOption {
	return func(options *InventoryZoneOptions) { options.revision = rev }
}

// Zone options

// InventoryZoneOption is a
//
type InventoryZoneOption func(*InventoryZoneOptions)

// InventoryZoneOptions is a struct
//
type InventoryZoneOptions struct {
	revision      int64
	includeRacks  bool
	includePdus   bool
	includeTors   bool
	includeBlades bool
}

func (options *InventoryZoneOptions) applyZoneOpts(optionsArray []InventoryZoneOption) {
	for _, option := range optionsArray {
		option(options)
	}
}

// WithZoneRevision is a
//
func WithZoneRevision(rev int64) InventoryZoneOption {
	return func(options *InventoryZoneOptions) { options.revision = rev }
}

// WithZoneRacks is a
//
func WithZoneRacks() InventoryZoneOption {
	return func(options *InventoryZoneOptions) { options.includeRacks = true }
}

// WithZoneTors is a
//
func WithZoneTors() InventoryZoneOption {
	return func(options *InventoryZoneOptions) { options.includeTors = true }
}

// WithZonePdus is a
//
func WithZonePdus() InventoryZoneOption {
	return func(options *InventoryZoneOptions) { options.includePdus = true }
}

// WithZoneBlades is a
//
func WithZoneBlades() InventoryZoneOption {
	return func(options *InventoryZoneOptions) { options.includeBlades = true }
}

// Rack options

// InventoryRackOption is a specific option to be applied to the
// operation of interest.
//
type InventoryRackOption func(*InventoryRackOptions)

// InventoryRackOptions is a struct used to collect the set of
// options to be applied to the operation of interest.
//
type InventoryRackOptions struct {
	revision      int64
	includePdus   bool
	includeTors   bool
	includeBlades bool
}

func (options *InventoryRackOptions) applyRackOpts(optionsArray []InventoryRackOption) {
	for _, option := range optionsArray {
		option(options)
	}
}

// WithRackRevision is an option to request a specific revision be
// used for the operation. If the revision in not available, or
// inappropriate, an error will be returned.
//
func WithRackRevision(rev int64) InventoryRackOption {
	return func(options *InventoryRackOptions) { options.revision = rev }
}

// WithRackPdus is an option to request that the Pdu details also
// be returned on a rack or zone read operation.
//
// Note, this can lead to much larger responses to read requests
// and should be used with care.
//
func WithRackPdus() InventoryRackOption {
	return func(options *InventoryRackOptions) { options.includePdus = true }
}

// WithRackTors is an option to request that the Tor details also
// be returned on a rack or zone read operation.
//
// Note, this can lead to much larger responses to read requests
// and should be used with care.
//
func WithRackTors() InventoryRackOption {
	return func(options *InventoryRackOptions) { options.includeTors = true }
}

// WithRackBlades is an option to request that the blade details
// also be returned on a rack or zone read operation.
//
// Note, this can lead to much larger responses to read requests
// and should be used with care.
//
func WithRackBlades() InventoryRackOption {
	return func(options *InventoryRackOptions) { options.includeBlades = true }
}

// ListZones returns the basic zone record for all the discovered zones. Optionally,
// racks along with the rack component PDU, TOR and blades can also be returned.
//
func (m *DBInventory) ListZones(
	ctx context.Context,
	options ...InventoryZoneOption) (map[string]*DefinitionZone, int64, error) {
	return nil, InvalidRev, nil
}

// ListRacks returns the basic zone record for all the discovered racks within
// the specified zone. Optionally, the rack component PDU, TOR and blades can
// also be returned.
//
func (m *DBInventory) ListRacks(
	ctx context.Context,
	zone string,
	options ...InventoryRackOption) (map[string]*DefinitionRack, int64, error) {
	return nil, InvalidRev, nil
}

// ListPdus returns the basic records for the PDUs in the specified rack.
//
func (m *DBInventory) ListPdus(
	ctx context.Context,
	zone string,
	rack string,
	options ...InventoryOption) (map[string]*DefinitionPdu, int64, error) {
	return nil, InvalidRev, nil
}

// ListTors returns the basic records for the TORs in the specified rack.
//
func (m *DBInventory) ListTors(
	ctx context.Context,
	zone string,
	rack string,
	options ...InventoryOption) (map[string]*DefinitionTor, int64, error) {
	return nil, InvalidRev, nil
}

// ListBlades returns the basic records for the blades in the specified rack.
//
func (m *DBInventory) ListBlades(
	ctx context.Context,
	zone string,
	rack string,
	options ...InventoryOption) (map[string]*DefinitionBlade, int64, error) {
	return nil, InvalidRev, nil
}

// CreateZone is used to create a basic zone record in the store.
//
// This record created will contain just the zone level details
// and any additional rack, blade, tor or pdu data will be ignored
// and not included in the stored record.
//
func (m *DBInventory) CreateZone(
	ctx context.Context,
	name string,
	zone *pb.Definition_Zone,
	options ...InventoryZoneOption) (int64, error) {

	z, err := inventory.NewZone(
		ctx,
		m.Store,
		inventory.DefinitionTable,
		inventory.DefaultRegion,
		name)

	if err != nil {
		return InvalidRev, err
	}

	z.SetDetails(ctx, zone.Details)

	rev, err := z.Create(ctx)

	if err != nil {
		return InvalidRev, err
	}

	return rev, nil
}

// CreateRack is used to create a basic rack record in the store.
//
// This record created will contain just the rack level details
// and any additional blade, tor or pdu data will be ignored
// and not included in the stored record.
//
func (m *DBInventory) CreateRack(
	ctx context.Context,
	zone string,
	name string,
	rack *pb.Definition_Rack,
	options ...InventoryRackOption) (int64, error) {

	r, err := inventory.NewRack(
		ctx,
		m.Store,
		inventory.DefinitionTable,
		inventory.DefaultRegion,
		zone,
		name)

	if err != nil {
		return InvalidRev, err
	}

	r.SetDetails(ctx, rack.Details)

	rev, err := r.Create(ctx)

	if err != nil {
		return InvalidRev, err
	}

	return rev, nil
}

// CreatePdu is used to create a basic pdu record in the store.
//
// This record created will contain just the pdu level details
// and any additional data will be ignored and not included in
// the stored record.
//
func (m *DBInventory) CreatePdu(
	ctx context.Context,
	zone string,
	rack string,
	index int64,
	pdu *pb.Definition_Pdu,
	options ...InventoryOption) (int64, error) {

	if err := pdu.Validate("", 0); err != nil {
		return InvalidRev, err
	}

	p, err := inventory.NewPdu(
		ctx,
		m.Store,
		inventory.DefinitionTable,
		inventory.DefaultRegion,
		zone,
		rack,
		index,
	)

	if err != nil {
		return InvalidRev, err
	}

	p.SetDetails(ctx, pdu.Details)
	p.SetPorts(ctx, &pdu.Ports)

	rev, err := p.Create(ctx)

	if err != nil {
		return InvalidRev, err
	}

	return rev, nil
}

// CreateTor is used to create a basic tor record in the store.
//
// This record created will contain just the tor level details
// and any additional data will be ignored and not included in
// the stored record.
//
func (m *DBInventory) CreateTor(
	ctx context.Context,
	zone string,
	rack string,
	index int64,
	tor *pb.Definition_Tor,
	options ...InventoryOption) (int64, error) {

	if err := tor.Validate("", 0); err != nil {
		return InvalidRev, err
	}

	t, err := inventory.NewTor(
		ctx,
		m.Store,
		inventory.DefinitionTable,
		inventory.DefaultRegion,
		zone,
		rack,
		index,
	)

	if err != nil {
		return InvalidRev, err
	}

	t.SetDetails(ctx, tor.Details)
	t.SetPorts(ctx, &tor.Ports)

	rev, err := t.Create(ctx)

	if err != nil {
		return InvalidRev, err
	}

	return rev, nil
}

// CreateBlade is used to create a basic blade record in the store.
//
// This record created will contain just the blade level details
// and any additional data will be ignored and not included in
// the stored record.
//
func (m *DBInventory) CreateBlade(
	ctx context.Context,
	zone string,
	rack string,
	index int64,
	blade *pb.Definition_Blade,
	options ...InventoryOption) (int64, error) {

	b, err := inventory.NewBlade(
		ctx,
		m.Store,
		inventory.DefinitionTable,
		inventory.DefaultRegion,
		zone,
		rack,
		index,
	)

	if err != nil {
		return InvalidRev, err
	}

	b.SetDetails(ctx, blade.Details)
	b.SetCapacity(ctx, blade.Capacity)
	b.SetBootInfo(ctx, blade.BootOnPowerOn, blade.BootInfo)

	rev, err := b.Create(ctx)

	if err != nil {
		return InvalidRev, err
	}

	return rev, nil
}

// ReadZone returns the zone information with optionally additional
// rack, blade, tor and pdu details for an optionally specified
// revision.
//
func (m *DBInventory) ReadZone(
	ctx context.Context,
	name string,
	options ...InventoryZoneOption) (*pb.Definition_Zone, int64, error) {

	z, err := inventory.NewZone(
		ctx,
		m.Store,
		inventory.DefinitionTable,
		inventory.DefaultRegion,
		name)

	if err != nil {
		return nil, InvalidRev, err
	}

	rev, err := z.Read(ctx)

	if err != nil {
		return nil, InvalidRev, err
	}

	details := z.GetDetails(ctx)

	return &pb.Definition_Zone{Details: details}, rev, nil
}

// ReadRack returns the rack information with optionally additional
// blade, tor and pdu details for an optionally specified revision.
//
func (m *DBInventory) ReadRack(
	ctx context.Context,
	zone string,
	name string,
	options ...InventoryRackOption) (*pb.Definition_Rack, int64, error) {

	r, err := inventory.NewRack(
		ctx,
		m.Store,
		inventory.DefinitionTable,
		inventory.DefaultRegion,
		zone,
		name)

	if err != nil {
		return nil, InvalidRev, err
	}

	rev, err := r.Read(ctx)

	if err != nil {
		return nil, InvalidRev, err
	}

	details := r.GetDetails(ctx)

	return &pb.Definition_Rack{Details: details}, rev, nil
}

// ReadPdu returns the PDU information for an optionally specified revision.
//
func (m *DBInventory) ReadPdu(
	ctx context.Context,
	zone string,
	rack string,
	index int64,
	options ...InventoryOption) (*pb.Definition_Pdu, int64, error) {

	p, err := inventory.NewPdu(
		ctx,
		m.Store,
		inventory.DefinitionTable,
		inventory.DefaultRegion,
		zone,
		rack,
		index)

	if err != nil {
		return nil, InvalidRev, err
	}

	rev, err := p.Read(ctx)

	if err != nil {
		return nil, InvalidRev, err
	}

	details := p.GetDetails(ctx)
	ports := p.GetPorts(ctx)

	return &pb.Definition_Pdu{Details: details, Ports: *ports}, rev, nil
}

// ReadTor returns the TOR information for an optionally specified revision.
//
func (m *DBInventory) ReadTor(
	ctx context.Context,
	zone string,
	rack string,
	index int64,
	options ...InventoryOption) (*pb.Definition_Tor, int64, error) {

	t, err := inventory.NewTor(
		ctx,
		m.Store,
		inventory.DefinitionTable,
		inventory.DefaultRegion,
		zone,
		rack,
		index)

	if err != nil {
		return nil, InvalidRev, err
	}

	rev, err := t.Read(ctx)

	if err != nil {
		return nil, InvalidRev, err
	}

	details := t.GetDetails(ctx)
	ports := t.GetPorts(ctx)

	return &pb.Definition_Tor{Details: details, Ports: *ports}, rev, nil
}

// ReadBlade returns the blade information for an optionally specified revision.
//
func (m *DBInventory) ReadBlade(
	ctx context.Context,
	zone string,
	rack string,
	index int64,
	options ...InventoryOption) (*pb.Definition_Blade, int64, error) {

	b, err := inventory.NewBlade(
		ctx,
		m.Store,
		inventory.DefinitionTable,
		inventory.DefaultRegion,
		zone,
		rack,
		index)

	if err != nil {
		return nil, InvalidRev, err
	}

	rev, err := b.Read(ctx)

	if err != nil {
		return nil, InvalidRev, err
	}

	details := b.GetDetails(ctx)
	capacity := b.GetCapacity(ctx)
	bootOnPowerOn, bootInfo := b.GetBootInfo(ctx)

	blade := &pb.Definition_Blade{
		Details:       details,
		Capacity:      capacity,
		BootOnPowerOn: bootOnPowerOn,
		BootInfo:      bootInfo,
	}

	return blade, rev, nil
}

// UpdateZone is used to update the zone basic details record.
//
// Only the zone level details will be updated and any
// additional rack, blade, tor or pdu data will be ignored
// and not included in the updated record.
//
func (m *DBInventory) UpdateZone(
	ctx context.Context,
	name string,
	zone *pb.Definition_Zone,
	options ...InventoryZoneOption) (int64, error) {

	z, err := inventory.NewZone(
		ctx,
		m.Store,
		inventory.DefinitionTable,
		inventory.DefaultRegion,
		name)

	if err != nil {
		return InvalidRev, err
	}

	z.SetDetails(ctx, zone.Details)

	rev, err := z.Update(ctx, true)

	if err != nil {
		return InvalidRev, err
	}

	return rev, nil
}

// UpdateRack is used to update the rack basic details record.
//
// Only the rack level details will be updated and any
// additional blade, tor or pdu data will be ignored
// and not included in the updated record.
//
func (m *DBInventory) UpdateRack(
	ctx context.Context,
	zone string,
	name string,
	rack *pb.Definition_Rack, options ...InventoryRackOption) (int64, error) {

	r, err := inventory.NewRack(
		ctx,
		m.Store,
		inventory.DefinitionTable,
		inventory.DefaultRegion,
		zone,
		name)

	if err != nil {
		return InvalidRev, err
	}

	r.SetDetails(ctx, rack.Details)

	rev, err := r.Update(ctx, true)

	if err != nil {
		return InvalidRev, err
	}

	return rev, nil
}

// UpdatePdu is used to update the PDU record.
//
func (m *DBInventory) UpdatePdu(
	ctx context.Context,
	zone string,
	rack string,
	index int64,
	pdu *pb.Definition_Pdu,
	options ...InventoryOption) (int64, error) {

	p, err := inventory.NewPdu(
		ctx,
		m.Store,
		inventory.DefinitionTable,
		inventory.DefaultRegion,
		zone,
		rack,
		index,
	)

	if err != nil {
		return InvalidRev, err
	}

	p.SetDetails(ctx, pdu.Details)
	p.SetPorts(ctx, &pdu.Ports)

	rev, err := p.Update(ctx, true)

	if err != nil {
		return InvalidRev, err
	}

	return rev, nil
}

// UpdateTor is used to update the TOR record.
//
func (m *DBInventory) UpdateTor(
	ctx context.Context,
	zone string,
	rack string,
	index int64,
	tor *pb.Definition_Tor,
	options ...InventoryOption) (int64, error) {

	t, err := inventory.NewTor(
		ctx,
		m.Store,
		inventory.DefinitionTable,
		inventory.DefaultRegion,
		zone,
		rack,
		index,
	)

	if err != nil {
		return InvalidRev, err
	}

	t.SetDetails(ctx, tor.Details)
	t.SetPorts(ctx, &tor.Ports)

	rev, err := t.Update(ctx, true)

	if err != nil {
		return InvalidRev, err
	}

	return rev, nil
}

// UpdateBlade is used to update the blade record.
//
func (m *DBInventory) UpdateBlade(
	ctx context.Context,
	zone string,
	rack string,
	index int64,
	blade *pb.Definition_Blade,
	options ...InventoryOption) (int64, error) {

	b, err := inventory.NewBlade(
		ctx,
		m.Store,
		inventory.DefinitionTable,
		inventory.DefaultRegion,
		zone,
		rack,
		index,
	)

	if err != nil {
		return InvalidRev, err
	}

	b.SetDetails(ctx, blade.Details)
	b.SetCapacity(ctx, blade.Capacity)
	b.SetBootInfo(ctx, blade.BootOnPowerOn, blade.BootInfo)

	rev, err := b.Update(ctx, true)

	if err != nil {
		return InvalidRev, err
	}

	return rev, nil
}

// DeleteZone is used to delete the zone record and any
// contained rack records. That is it will delete the
// entire zone and all related records.
//
func (m *DBInventory) DeleteZone(
	ctx context.Context,
	name string,
	options ...InventoryOption) (int64, error) {

	z, err := inventory.NewZone(
		ctx,
		m.Store,
		inventory.DefinitionTable,
		inventory.DefaultRegion,
		name)

	if err != nil {
		return InvalidRev, err
	}

	rev, err := z.Delete(ctx, true)

	if err != nil {
		return InvalidRev, err
	}

	return rev, nil
}

// DeleteRack is used to delete the rack record and any
// contained PDU, TOR or blade records. That is it will
// delete the entire rack and all related records.
//
func (m *DBInventory) DeleteRack(
	ctx context.Context,
	zone string,
	name string,
	options ...InventoryOption) (int64, error) {

	r, err := inventory.NewRack(
		ctx,
		m.Store,
		inventory.DefinitionTable,
		inventory.DefaultRegion,
		zone,
		name)

	if err != nil {
		return InvalidRev, err
	}

	rev, err := r.Delete(ctx, true)

	if err != nil {
		return InvalidRev, err
	}

	return rev, nil
}

// DeletePdu is used to delete the PDU record from a rack.
// If there is only a single PDU defined in the rack it
// will leave the rack without a PDU description which is
// at least an unexpected condition.
//
// TODO - should the deletion of the last PDU return a distinct
//        status to indicate it is the last PDU?
//
func (m *DBInventory) DeletePdu(
	ctx context.Context,
	zone string,
	rack string,
	index int64,
	options ...InventoryOption) (int64, error) {

	p, err := inventory.NewPdu(
		ctx,
		m.Store,
		inventory.DefinitionTable,
		inventory.DefaultRegion,
		zone,
		rack,
		index,
	)

	if err != nil {
		return InvalidRev, err
	}

	rev, err := p.Delete(ctx, true)

	if err != nil {
		return InvalidRev, err
	}

	return rev, nil
}

// DeleteTor is used to delete the TOR record from a rack.
// If there is only a single TOR defined in the rack it
// will leave the rack without a TOR description which is
// at least an unexpected condition.
//
// TODO - should the deletion of the last TOR return a distinct
//        status to indicate it is the last TOR?
//
func (m *DBInventory) DeleteTor(
	ctx context.Context,
	zone string,
	rack string,
	index int64,
	options ...InventoryOption) (int64, error) {

	t, err := inventory.NewTor(
		ctx,
		m.Store,
		inventory.DefinitionTable,
		inventory.DefaultRegion,
		zone,
		rack,
		index,
	)

	if err != nil {
		return InvalidRev, err
	}

	rev, err := t.Create(ctx)

	if err != nil {
		return InvalidRev, err
	}

	return rev, nil
}

// DeleteBlade is used to delete the blade record from a rack.
// If there is only a single blade defined in the rack it
// will leave the rack without a blade description which is
// at least an unexpected condition.
//
// TODO - should the deletion of the last blade return a distinct
//        status to indicate it is the last blade?
//
func (m *DBInventory) DeleteBlade(
	ctx context.Context,
	zone string,
	rack string,
	index int64,
	options ...InventoryOption) (int64, error) {

	b, err := inventory.NewBlade(
		ctx,
		m.Store,
		inventory.DefinitionTable,
		inventory.DefaultRegion,
		zone,
		rack,
		index,
	)

	if err != nil {
		return InvalidRev, err
	}

	rev, err := b.Delete(ctx, true)

	if err != nil {
		return InvalidRev, err
	}

	return rev, nil
}
