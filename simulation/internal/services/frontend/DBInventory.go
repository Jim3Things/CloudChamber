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
	"sync"

	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/inventory"
	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/store"
	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/timestamp"
	"github.com/Jim3Things/CloudChamber/simulation/internal/common"
	"github.com/Jim3Things/CloudChamber/simulation/internal/config"
	"github.com/Jim3Things/CloudChamber/simulation/internal/tracing"
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

	Zone *pb.External_Zone

	ZoneCount     int
	MaxBladeCount int64
	MaxCapacity   *pb.BladeCapacity
	Store         *store.Store

	Root *pb.Definition_Root
}

var dbInventory *DBInventory

// InitDBInventory initializes the base state for the inventory.
//
// At present the primary state is sufficient data in an in-memory db sufficient
// for testing purposes. Eventually, this will be removed and the calls will be
// connected to the store in order to persist the inventory read from an external
// definition file
//
func InitDBInventory(ctx context.Context, cfg *config.GlobalConfig) (err error) {
	if dbInventory == nil {
		db := &DBInventory{
			mutex:         sync.RWMutex{},
			Zone:          nil,
			MaxBladeCount: 0,
			MaxCapacity:   &pb.BladeCapacity{},
			Store:         store.NewStore(),
			Root:          &pb.Definition_Root{},
		}

		if err = db.Initialize(ctx, cfg); err != nil {
			return err
		}

		if dbInventory == nil {
			dbInventory = db
		}

		// For temporary backwards compatibilities sake, need to have the older
		// version here to allow all the current tests to run before we have a
		// complete cut-over to the store based inventory definition.
		//
		zone, err := inventory.ReadInventoryDefinition(ctx, cfg.Inventory.InventoryDefinition)

		if err != nil {
			return err
		}

		dbInventory.Zone = zone

		dbInventory.buildSummary(ctx)
	}

	return nil
}

// Initialize initializes an existing, but currently un-initialized DB Inventory
// structure. This involves connecting to the store, loading the current inventory
// definition from the definition file, and sending updates to the store to
// reconcile the persisted state. Note that changes to the store may in turn
// trigger any currently established watch handlers leading to updates elsewhere
// in the system.
//
func (m *DBInventory) Initialize(ctx context.Context, cfg *config.GlobalConfig) (err error) {
	ctx, span := tracing.StartSpan(ctx,
		tracing.WithName("Initialize Inventory DB state"),
		tracing.WithContextValue(timestamp.EnsureTickInContext),
		tracing.AsInternal())
	defer span.End()

	if err = m.Store.Connect(); err != nil {
		return err
	}

	if err = m.LoadFromStore(ctx); err != nil {
		return err
	}

	if err = m.UpdateInventoryDefinition(ctx, cfg); err != nil {
		return err
	}

	return nil
}

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
				_, pdus, err := rack.FetchPdus(ctx)
				if err != nil {
					return nil, err
				}
			
				_, tors, err := rack.FetchTors(ctx)
				if err != nil {
					return nil, err
				}
			
				_, blades, err := rack.FetchBlades(ctx)
				if err != nil {
					return nil, err
				}

				defRack := &pb.Definition_Rack{
					Details: rack.GetDetails(ctx),
					Pdus:    make(map[int64]*pb.Definition_Pdu, len(*pdus)),
					Tors:    make(map[int64]*pb.Definition_Tor, len(*tors)),
					Blades:  make(map[int64]*pb.Definition_Blade, len(*blades)),
				}

				for pduIndex, pdu := range *pdus {
					defRack.Pdus[pduIndex] = &pb.Definition_Pdu{
						Details: pdu.GetDetails(ctx),
						Ports:   *pdu.GetPorts(ctx),
					}
				}

				for torIndex, tor := range *tors {
					defRack.Tors[torIndex] = &pb.Definition_Tor{
						Details: tor.GetDetails(ctx),
						Ports:   *tor.GetPorts(ctx),
					}
				}

				for bladeIndex, blade := range *blades {
					bootOnPowerOn, bootInfo := blade.GetBootInfo(ctx)

					defRack.Blades[bladeIndex] = &pb.Definition_Blade{
						Details:       blade.GetDetails(ctx),
						Capacity:      blade.GetCapacity(ctx),
						BootInfo:      bootInfo,
						BootOnPowerOn: bootOnPowerOn,
					}
				}

				defZone.Racks[rackName] = defRack
			}

			defRegion.Zones[zoneName] = defZone
		}

		defRoot.Regions[regionName] = defRegion
	}

	return defRoot, nil
}

func (m *DBInventory) writeInventoryDefinitionToStore(ctx context.Context, root *pb.Definition_Root) error {
	ctx, span := tracing.StartSpan(ctx,
		tracing.WithName("Read inventory definition from store"),
		tracing.WithContextValue(timestamp.EnsureTickInContext),
		tracing.AsInternal())
	defer span.End()

	storeRoot, err := inventory.NewRoot(ctx, m.Store, inventory.DefinitionTable)
	if err != nil {
		return err
	}

	storeRoot.SetDetails(ctx, root.Details)

	for regionName, region := range root.Regions {
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

				for i := range storeRack.Pdus {

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

// LoadFromStore is a method to load the currently known inventory from the store and in
// expected to use used on service startup. Subsequent to this, once all the
// component services are running, the inventory in the configuration file will
// be loaded and a reconciliation pass will take place with all the appropriate
// notifications for arrival and/or departures of various items in the inventory.
//
func (m *DBInventory) LoadFromStore(ctx context.Context) error {
	ctx, span := tracing.StartSpan(ctx,
		tracing.WithName("Load inventory definition from store"),
		tracing.WithContextValue(timestamp.EnsureTickInContext),
		tracing.AsInternal())
	defer span.End()

	return nil
}

// UpdateInventoryDefinition is a method to load a new inventory definition from
// the configured file. Once read, the store will be updated with the differences
// which will in turn trigger a set of previously established watch routines to
// issue a number of arrival and/or departure notifications.
//
func (m *DBInventory) UpdateInventoryDefinition(
	ctx context.Context,
	cfg *config.GlobalConfig,
) error {
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
	rootFile, err := inventory.ReadInventoryDefinitionFromFileEx(ctx, cfg.Inventory.InventoryDefinition)
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

// reconcileNewInventory compares the newly loaded inventory definition,
// presumably from a configuration file, with the currently loaded inventory
// and updates the store accordingly. This will trigger the various watches
// which any currently running services have previously established and deliver
// a set of arrival and/or departure notifications as appropriate.
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


	err := dbInventory.deleteInventoryDefinitionFromStore(ctx, rootStore)
	if err != nil {
		return err
	}


	err = dbInventory.writeInventoryDefinitionToStore(ctx, rootFile)
	if err != nil {
		return err
	}

	m.Root = rootFile

	m.ZoneCount, m.MaxBladeCount, m.MaxCapacity =  m.buildSummaryForRoot(ctx, rootFile)

	return nil
}

// GetMemoData returns the maximum number of blades held in any rack
// in the inventory.
func (m *DBInventory) GetMemoData() (int, int64, *pb.BladeCapacity) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return len(m.Zone.Racks), m.MaxBladeCount, m.MaxCapacity
}

// ScanRacks scans the set of known racks in the store, invoking the supplied
// function with each entry.
func (m *DBInventory) ScanRacks(action func(entry string) error) error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for name := range m.Zone.Racks {
		if err := action(name); err != nil {
			return err
		}
	}

	return nil
}

// GetRack returns the rack details to match the supplied rackID
func (m *DBInventory) GetRack(rackID string) (*pb.External_Rack, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

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
	m.mutex.RLock()
	defer m.mutex.RUnlock()

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
func (m *DBInventory) GetBlade(rackID string, bladeID int64) (*pb.BladeCapacity, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

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
// Assumptions: dbInventory (write)lock is already held.
//
func (m *DBInventory) buildSummary(ctx context.Context) {

	maxBladeCount := int64(0)
	memo := &pb.BladeCapacity{}

	for _, rack := range m.Zone.Racks {
		for _, blade := range rack.Blades {
			memo.Cores = common.MaxInt64(memo.Cores, blade.Cores)
			memo.DiskInGb = common.MaxInt64(memo.DiskInGb, blade.DiskInGb)
			memo.MemoryInMb = common.MaxInt64(memo.MemoryInMb, blade.MemoryInMb)
			memo.NetworkBandwidthInMbps = common.MaxInt64(
				memo.NetworkBandwidthInMbps,
				blade.NetworkBandwidthInMbps)
		}

		maxBladeCount = common.MaxInt64(maxBladeCount, int64(len(rack.Blades)))
	}

	m.MaxBladeCount = maxBladeCount
	m.MaxCapacity = memo

	tracing.Info(ctx, "   Updated inventory summary - MaxBladeCount: %d MaxCapacity: %v", m.MaxBladeCount, m.MaxCapacity)
}

// buildSummary constructs the memo-ed summary data for the zone.  This should
// be called whenever the configured inventory changes. This includes
//
// - the zone count
// - the maximum number of blades in a rack
// - the memo data itself
//
func (m *DBInventory) buildSummaryForRegion(
	ctx context.Context,
	zm *pb.Definition_Region) (int, int64, *pb.BladeCapacity) {

	maxBladeCount := int64(0)
	maxCapacity := &pb.BladeCapacity{}

	for _, zone := range zm.Zones {
		for _, rack := range zone.Racks {
			for _, blade := range rack.Blades {
				maxCapacity.Cores = common.MaxInt64(maxCapacity.Cores, blade.Capacity.Cores)
				maxCapacity.DiskInGb = common.MaxInt64(maxCapacity.DiskInGb, blade.Capacity.DiskInGb)
				maxCapacity.MemoryInMb = common.MaxInt64(maxCapacity.MemoryInMb, blade.Capacity.MemoryInMb)

				maxCapacity.NetworkBandwidthInMbps = common.MaxInt64(
					maxCapacity.NetworkBandwidthInMbps,
					blade.Capacity.NetworkBandwidthInMbps)
			}

			maxBladeCount = common.MaxInt64(maxBladeCount, int64(len(rack.Blades)))
		}
	}

	tracing.Info(ctx, "   Updated inventory summary - MaxBladeCount: %d MaxCapacity: %v", maxBladeCount, maxCapacity)

	return len(zm.Zones), maxBladeCount, maxCapacity
}

// buildSummary constructs the memo-ed summary data for the zone.  This should
// be called whenever the configured inventory changes. This includes
//
// - the zone count
// - the maximum number of blades in a rack
// - the memo data itself
//
func (m *DBInventory) buildSummaryForRoot(
	ctx context.Context,
	root *pb.Definition_Root) (int, int64, *pb.BladeCapacity) {

	zoneCount :=  int(0)
	maxBladeCount := int64(0)
	maxCapacity := &pb.BladeCapacity{}

	for _, region := range root.Regions {

		zoneCount += len(region.Zones)

		for _, zone := range region.Zones {
			for _, rack := range zone.Racks {
				for _, blade := range rack.Blades {
					maxCapacity.Cores = common.MaxInt64(maxCapacity.Cores, blade.Capacity.Cores)
					maxCapacity.DiskInGb = common.MaxInt64(maxCapacity.DiskInGb, blade.Capacity.DiskInGb)
					maxCapacity.MemoryInMb = common.MaxInt64(maxCapacity.MemoryInMb, blade.Capacity.MemoryInMb)

					maxCapacity.NetworkBandwidthInMbps = common.MaxInt64(
						maxCapacity.NetworkBandwidthInMbps,
						blade.Capacity.NetworkBandwidthInMbps)
				}

				maxBladeCount = common.MaxInt64(maxBladeCount, int64(len(rack.Blades)))
			}
		}
	}

	tracing.Info(ctx, "   Updated inventory summary - MaxBladeCount: %d MaxCapacity: %v", maxBladeCount, maxCapacity)

	return zoneCount, maxBladeCount, maxCapacity
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

// Definition_Tor describes the revision and value for the request TOR
//
type Definition_Tor struct {
	revision int64
	tor      *pb.Definition_Tor
}

// Definition_Pdu describes the revision and value for the request PDU
//
type Definition_Pdu struct {
	revision int64
	pdu      *pb.Definition_Pdu
}

// Definition_Blade describes the revision and value for the request blade
//
type Definition_Blade struct {
	revision int64
	blade    *pb.Definition_Blade
}

// Definition_Rack describes the revision and value for the request rack
//
type Definition_Rack struct {
	revision int64
	rack     *pb.Definition_Rack
}

// Definition_Zone defines the set of values used to summarize the contents of a zone to
// allow a query on the zone without incorporating the entire inventory definition
//
type Definition_Zone struct {
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
	options ...InventoryZoneOption) (map[string]*Definition_Zone, int64, error) {
	return nil, InvalidRev, nil
}

// ListRacks returns the basic zone record for all the discovered racks within
// the specified zone. Optionally, the rack component PDU, TOR and blades can
// also be returned.
//
func (m *DBInventory) ListRacks(
	ctx context.Context,
	zone string,
	options ...InventoryRackOption) (map[string]*Definition_Rack, int64, error) {
	return nil, InvalidRev, nil
}

// ListPdus returns the basic records for the PDUs in the specified rack.
//
func (m *DBInventory) ListPdus(
	ctx context.Context,
	zone string,
	rack string,
	options ...InventoryOption) (map[string]*Definition_Pdu, int64, error) {
	return nil, InvalidRev, nil
}

// ListTors returns the basic records for the TORs in the specified rack.
//
func (m *DBInventory) ListTors(
	ctx context.Context,
	zone string,
	rack string,
	options ...InventoryOption) (map[string]*Definition_Tor, int64, error) {
	return nil, InvalidRev, nil
}

// ListBlades returns the basic records for the blades in the specified rack.
//
func (m *DBInventory) ListBlades(
	ctx context.Context,
	zone string,
	rack string,
	options ...InventoryOption) (map[string]*Definition_Blade, int64, error) {
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
