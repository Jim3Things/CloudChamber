// This package is used to access the persisted inventory state in an underlying
// store and to provide an abstraction to the other CloudChamber services such that
// they need not concern themselves with how the data is laid out or manipulated.
//
// The access method is based around a 'cursor' object which can be used to
// operate on the underlying persisted version of that object just by
// using a few basic access methods, e.g. Read, Update, ListChildren etc.
//
// The objects providing such a cursor are regions, zones, racks, pdus, tor and
// blades. There is also a root object which provides a well-known entry point
// and can be used to start the discovery process. These object form a hierarchy
// which allows for navigating from parent objects to an associated child.
//
// The complete inventory is split into a number of separate namespaces (or tables)
// which are used to compartmentalize different usage groups within the inventory.
// For example, the primary namespaces are
//
//		definition
//		observed
//		actual
//		target
//
// See the CloudChamber documentation for a description of these namespaces and
// their uses.
//
//
// The objects within a table form a hierarchy of navigable "nodes" from root,
// to region, to zone to rack. Within a rack there are the further child objects
// for pdus, tors and blades. These last child objects do not themselves have
// children and are accessed as complete entities which are operated on as an
// atomic unit.
//
// NOTE: At least so far. It is conceivable that we may wish to extend this
//       to allow individual entries within these leaf items to be manipulated
//       directly, e.g. we may wish to update the state of a network port
//       within a tor without having to re-write the complete tor object.
//
// To allow for object discovery, the inventory allows a caller to request a
// list of all the children of a navigable node, e.g. search for all of the
// racks within a zone. To achieve this, the package maintains an index for
// child objects which is kept separately from the objects themselves. Thus
// when creating a new object, the package will also create an appropriate
// index entry to allow that object to be discovered from its parent. In
// this way, by knowing the root of a namespace, all the regions can be
// located. Once a region is know, all the zones within that region can be
// located. And so on for racks within zones, etc.
//
// NOTE: Currently there is no way to discover a parent for a given object.
//       However, each object contains sufficient information that a parent
//       can be readily identified if required which would allow this feature
//       to be added if needed.

package inventory

import (
	"context"
	"fmt"
	"sync"

	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/namespace"
	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/store"
	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/timestamp"
	"github.com/Jim3Things/CloudChamber/simulation/internal/common"
	"github.com/Jim3Things/CloudChamber/simulation/internal/config"
	"github.com/Jim3Things/CloudChamber/simulation/internal/tracing"
	"github.com/Jim3Things/CloudChamber/simulation/pkg/errors"
	pb "github.com/Jim3Things/CloudChamber/simulation/pkg/protos/inventory"
)

const (

	// DefaultRegion is used to provide a value for the non-existing region name
	// while the transition to the new inventory extended schema continues.
	// Eventually these will disappear as the front-end and higher layers learn
	// abouts regions, zones, multiple pdus and tors.
	//
	DefaultRegion = "standard"

	// DefaultZone is used to provide a value for the non-existing zone name
	// while the transition to the new inventory extended schema continues.
	// Eventually these will disappear as the front-end and higher layers learn
	// abouts regions, zones, multiple pdus and tors.
	//
	DefaultZone = "standard"
)

// ViewType and the associated enum values define the various views that the
// store has.
//
// Note: only ViewDefinition is currently in use.
type ViewType int
const (
	ViewDefinition ViewType = iota
	ViewTarget
	ViewObserved
	ViewActual
)

type inventoryItemRack interface {
	NewPdu(name string) (interface{}, error)
	NewTor(name string) (interface{}, error)
	NewBlade(name string) (interface{}, error)

	ListPdus(ctx context.Context) (int64, map[int64]interface{}, error)
	ListTors(ctx context.Context) (int64, map[int64]interface{}, error)
	ListBlades(ctx context.Context) (int64, map[int64]interface{}, error)

	FetchPdus(ctx context.Context) (int64, map[int64]interface{}, error)
	FetchTors(ctx context.Context) (int64, map[int64]interface{}, error)
	FetchBlades(ctx context.Context) (int64, map[int64]interface{}, error)
}

type inventoryItemPdu interface {
	SetPorts(ports map[int64]*pb.PowerPort)
	GetPorts() map[int64]*pb.PowerPort
}

type inventoryTor interface {
	SetPorts(cports map[int64]*pb.NetworkPort)
	GetPorts() map[int64]*pb.NetworkPort
}

type inventoryBlade interface {
	SetCapacity(capacity *pb.BladeCapacity)
	GetCapacity() *pb.BladeCapacity

	SetBootInfo(bootInfo *pb.BladeBootInfo)
	GetBootInfo() *pb.BladeBootInfo

	SetBootPowerOn(bootOnPowerOn bool)
	GetBootOnPowerOn() bool
}


// RackSizeSummary contains the maximum content values for a set of racks.  The
// values can be used to form a visual fixed rack size that is certain to be
// sufficient to hold any rack's contents covered by this summary instance.
type RackSizeSummary struct {
	MaxTorCount   int
	MaxPduCount   int
	MaxBladeCount int
	MaxConnectors int
	MaxCapacity   *pb.BladeCapacity
}

// String returns a formatted description of the summary instance.
func (s RackSizeSummary) String() string {
	return fmt.Sprintf(
		"max TORs/rack=%d, max PDUs/rack=%d, maxConnectors in either=%d, max blade capacity=%s",
		s.MaxTorCount, s.MaxPduCount, s.MaxConnectors, s.MaxCapacity.String())
}

// setToMax updates the instance such that it encompasses the boundaries set by
// both itself and the supplied summary instance.  It does this by setting each
// member field to be the maximum of two values.
func (s *RackSizeSummary) setToMax(t RackSizeSummary) {
	s.MaxTorCount = common.MaxInt(s.MaxTorCount, t.MaxTorCount)
	s.MaxPduCount = common.MaxInt(s.MaxPduCount, t.MaxPduCount)
	s.MaxBladeCount = common.MaxInt(s.MaxBladeCount, t.MaxBladeCount)
	s.MaxConnectors = common.MaxInt(s.MaxConnectors, t.MaxConnectors)
	s.MaxCapacity = maxBladeCapacity(s.MaxCapacity, t.MaxCapacity)
}

// maxBladeCapacity is a helper function that returns a BladeCapacity instance
// in which the core, memory, disk, and network capacity fields are the maximum
// of the two supplied values.
func maxBladeCapacity(s, t *pb.BladeCapacity) *pb.BladeCapacity {
	return &pb.BladeCapacity{
		Cores:                  common.MaxInt64(s.Cores, t.Cores),
		MemoryInMb:             common.MaxInt64(s.MemoryInMb, t.MemoryInMb),
		DiskInGb:               common.MaxInt64(s.DiskInGb, t.DiskInGb),
		NetworkBandwidthInMbps: common.MaxInt64(s.NetworkBandwidthInMbps, t.NetworkBandwidthInMbps),
		Arch:                   "",
		Accelerators:           nil,
	}
}

// ZoneSummary contains the summary data for a single zone. The contents
// are either zero or are (re-)computed whenever the inventory definitions
// are loaded from file. They are a cache of data in the store to avoid
// having to scan the entire zone whenever a simple query for the basic
// data is being handled.
//
type ZoneSummary struct {
	RackCount int
	RackSizeSummary
}

// RegionSummary contains the summary data for a single region. The contents
// are either zero or are (re-)computed whenever the inventory definitions
// are loaded from file. They are a cache of data in the store to avoid
// having to scan the entire region whenever a simple query for the basic
// data is being handled.
//
type RegionSummary struct {
	ZoneCount    int
	MaxRackCount int
	RackSizeSummary
}

// RootSummary contains the summary data for the entire inventory. The contents
// are either zero or are (re-)computed whenever the inventory definitions
// are loaded from file. They are a cache of data in the store to avoid
// having to scan the entire zone whenever a simple query for the basic
// data is being handled.
//
type RootSummary struct {
	RegionCount  int
	MaxZoneCount int
	MaxRackCount int
	RackSizeSummary
}

// Inventory is a structure used to established synchronized access to values
// required to make use of the inventory layer.
//
type Inventory struct {
	mutex              sync.RWMutex
	cfg                *config.GlobalConfig
	Store              *store.Store
	RootSummary        *RootSummary
	DefaultZoneSummary *ZoneSummary
}

// NewInventory is a helper routine to construct an empty Inventory structure
// as a convenience to avoid clients having to do all the details themselves.
//
func NewInventory(cfg *config.GlobalConfig, store *store.Store) *Inventory {
	return &Inventory{
		mutex:       sync.RWMutex{},
		cfg:         cfg,
		Store:       store,
		RootSummary: &RootSummary{},
		DefaultZoneSummary: &ZoneSummary{
			RackSizeSummary: RackSizeSummary{
				MaxTorCount:   0,
				MaxPduCount:   0,
				MaxBladeCount: 0,
				MaxConnectors: 0,
				MaxCapacity: &pb.BladeCapacity{
					Accelerators: []*pb.Accelerator{},
				},
			},
		},
	}
}

// Start is a function to get the inventory ready for use.
//
func (m *Inventory) Start(ctx context.Context) error {

	if err := m.Store.Connect(); err != nil {
		return err
	}

	return m.updateSummaryInformation(ctx)
}

// NewRoot returns a root object which acts as a well-known point in a namespace
// and which can be used to navigate the namespace for a given table.
//
// Valid tables are
//	- DefinitionTable
//	- ActualTable
//	- ObservedTable
//	- TargetTable
//
func (m *Inventory) NewRoot(table namespace.TableName) (*Root, error) {

	return newRoot(m.Store, table)
}

// GetDefaultZoneSummary returns the maximum number of blades held in any rack
// in the inventory.
func (m *Inventory) GetDefaultZoneSummary() *ZoneSummary {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.DefaultZoneSummary
}

// UpdateInventoryDefinition is a method to load a new inventory definition from
// the configured file. Once read, the store will be updated with the differences
// which will in turn trigger a set of previously established watch routines to
// issue a number of arrival and/or departure notifications.
//
func (m *Inventory) UpdateInventoryDefinition(ctx context.Context, path string) error {
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
	rootFile, err := ReadInventoryDefinitionFromFile(ctx, path)
	if err != nil {
		return err
	}

	tracing.UpdateSpanName(ctx,
		"Read inventory definition with name %q from path %s",
		rootFile.Details.Name,
		path)

	if err = m.reconcileNewInventory(ctx, rootFile); err != nil {
		return err
	}

	tracing.UpdateSpanName(ctx,
		"Reconciled inventory definition into store with name %q from path %s",
		rootFile.Details.Name,
		path)

	return m.updateSummaryInformation(ctx)
}

// DeleteInventoryDefinition is a
//
func (m *Inventory) DeleteInventoryDefinition(ctx context.Context) error {
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

	return m.updateSummaryInformation(ctx)
}

func (m *Inventory) reconcileNewInventoryRegion(
	ctx context.Context,
	regionStore *Region,
	regionFile *pb.Definition_Region,
) error {
	_, err := regionStore.Read(ctx, ViewDefinition)

	if err == error(errors.ErrRegionNotFound{Region: regionStore.Region}) {
		tracing.Info(ctx, "creating new region %s", regionStore.Region)

		regionStore.SetDetails(regionFile.GetDetails())

		_, err = regionStore.Create(ctx, ViewDefinition)
	} else if err == nil && regionStore.NotEqual(regionFile) {
		tracing.Info(ctx, "updating region %s", regionStore.Region)

		regionStore.SetDetails(regionFile.GetDetails())

		_, err = regionStore.Update(ctx, true, ViewDefinition)
	}

	return err
}

func (m *Inventory) reconcileNewInventoryZone(
	ctx context.Context,
	zoneStore *Zone,
	zoneFile *pb.Definition_Zone,
) error {
	_, err := zoneStore.Read(ctx, ViewDefinition)

	if err == error(errors.ErrZoneNotFound{Region: zoneStore.Region, Zone: zoneStore.Zone}) {
		tracing.Info(ctx, "creating new zone %s/%s", zoneStore.Region, zoneStore.Zone)

		zoneStore.SetDetails(zoneFile.GetDetails())

		_, err = zoneStore.Create(ctx, ViewDefinition)
	} else if err == nil && zoneStore.NotEqual(zoneFile) {
		tracing.Info(ctx, "updating zone %s/%s", zoneStore.Region, zoneStore.Zone)

		zoneStore.SetDetails(zoneFile.GetDetails())

		_, err = zoneStore.Update(ctx, true, ViewDefinition)
	}

	return err
}

func (m *Inventory) reconcileNewInventoryRack(
	ctx context.Context,
	rackStore *Rack,
	rackFile *pb.Definition_Rack,
) error {
	_, err := rackStore.Read(ctx, ViewDefinition)

	if err == error(errors.ErrRackNotFound{Region: rackStore.Region, Zone: rackStore.Zone, Rack: rackStore.Rack}) {
		tracing.Info(ctx, "creating new rack %s/%s/%s", rackStore.Region, rackStore.Zone, rackStore.Rack)

		rackStore.SetDetails(rackFile.GetDetails())

		_, err = rackStore.Create(ctx, ViewDefinition)
	} else if err == nil && rackStore.NotEqual(rackFile) {
		tracing.Info(ctx, "updating rack %s/%s/%s", rackStore.Region, rackStore.Zone, rackStore.Rack)

		rackStore.SetDetails(rackFile.GetDetails())

		_, err = rackStore.Update(ctx, true, ViewDefinition)
	}

	if err == nil {
		if err = m.reconcileNewInventoryPdus(ctx, rackStore, rackFile.Pdus); err != nil {
			return err
		}

		if err = m.reconcileNewInventoryTors(ctx, rackStore, rackFile.Tors); err != nil {
			return err
		}

		err = m.reconcileNewInventoryBlades(ctx, rackStore, rackFile.Blades)
	}

	return err
}

func (m *Inventory) reconcileNewInventoryPdu(
	ctx context.Context,
	pduStore *Pdu,
	pduFile *pb.Definition_Pdu,
) error {
	_, err := pduStore.Read(ctx, ViewDefinition)

	if err == error(errors.ErrPduNotFound{Region: pduStore.Region, Zone: pduStore.Zone, Rack: pduStore.Rack, Pdu: pduStore.ID}) {
		tracing.Info(ctx, "creating new pdu %s/%s/%s/%d", pduStore.Region, pduStore.Zone, pduStore.Rack, pduStore.ID)

		ports := pduFile.GetPorts()
		pduStore.SetPorts(ports)
		pduStore.SetDetails(pduFile.GetDetails())

		_, err = pduStore.Create(ctx, ViewDefinition)
	} else if err == nil && pduStore.NotEqual(pduFile) {
		tracing.Info(ctx, "updating pdu %s/%s/%s/%d", pduStore.Region, pduStore.Zone, pduStore.Rack, pduStore.ID)

		ports := pduFile.GetPorts()
		pduStore.SetPorts(ports)
		pduStore.SetDetails(pduFile.GetDetails())

		_, err = pduStore.Update(ctx, true, ViewDefinition)
	}

	return err
}

func (m *Inventory) reconcileNewInventoryTor(
	ctx context.Context,
	torStore *Tor,
	torFile *pb.Definition_Tor,
) error {
	_, err := torStore.Read(ctx, ViewDefinition)

	if err == error(errors.ErrTorNotFound{Region: torStore.Region, Zone: torStore.Zone, Rack: torStore.Rack, Tor: torStore.ID}) {
		tracing.Info(ctx, "creating new %s/%s/%s/%d", torStore.Region, torStore.Zone, torStore.Rack, torStore.ID)

		ports := torFile.GetPorts()
		torStore.SetPorts(ports)
		torStore.SetDetails(torFile.GetDetails())

		_, err = torStore.Create(ctx, ViewDefinition)
	} else if err == nil && torStore.NotEqual(torFile) {
		tracing.Info(ctx, "updating %s/%s/%s/%d", torStore.Region, torStore.Zone, torStore.Rack, torStore.ID)

		ports := torFile.GetPorts()
		torStore.SetPorts(ports)
		torStore.SetDetails(torFile.GetDetails())

		_, err = torStore.Update(ctx, true, ViewDefinition)
	}

	return err
}

func (m *Inventory) reconcileNewInventoryBlade(
	ctx context.Context,
	bladeStore *Blade,
	bladeFile *pb.Definition_Blade,
) error {
	_, err := bladeStore.Read(ctx, ViewDefinition)

	if err == error(errors.ErrBladeNotFound{Region: bladeStore.Region, Zone: bladeStore.Zone, Rack: bladeStore.Rack, Blade: bladeStore.ID}) {
		tracing.Info(ctx, "creating new blade %s/%s/%s/%d", bladeStore.Region, bladeStore.Zone, bladeStore.Rack, bladeStore.ID)

		bladeStore.SetDetails(bladeFile.GetDetails())
		bladeStore.SetCapacity(bladeFile.GetCapacity())
		bladeStore.SetBootInfo(bladeFile.GetBootInfo())
		bladeStore.SetBootPowerOn(bladeFile.GetBootOnPowerOn())

		_, err = bladeStore.Create(ctx, ViewDefinition)
	} else if err == nil && bladeStore.NotEqual(bladeFile) {
		tracing.Info(ctx, "updating blade %s/%s/%s/%d", bladeStore.Region, bladeStore.Zone, bladeStore.Rack, bladeStore.ID)

		bladeStore.SetDetails(bladeFile.GetDetails())
		bladeStore.SetCapacity(bladeFile.GetCapacity())
		bladeStore.SetBootInfo(bladeFile.GetBootInfo())
		bladeStore.SetBootPowerOn(bladeFile.GetBootOnPowerOn())

		_, err = bladeStore.Update(ctx, true, ViewDefinition)
	}

	return err
}

func (m *Inventory) reconcileNewInventoryPdus(
	ctx context.Context,
	rackStore *Rack,
	pdus map[int64]*pb.Definition_Pdu,
) error {
	for index, pduFile := range pdus {
		pduStore, err := rackStore.NewPdu(index)
		if err != nil {
			return err
		}

		if err = m.reconcileNewInventoryPdu(ctx, pduStore, pduFile); err != nil {
			return err
		}
	}

	return nil
}

func (m *Inventory) reconcileNewInventoryTors(
	ctx context.Context,
	rackStore *Rack,
	tors map[int64]*pb.Definition_Tor,
) error {
	for index, torFile := range tors {
		torStore, err := rackStore.NewTor(index)
		if err != nil {
			return err
		}

		if err = m.reconcileNewInventoryTor(ctx, torStore, torFile); err != nil {
			return err
		}
	}

	return nil
}

func (m *Inventory) reconcileNewInventoryBlades(
	ctx context.Context,
	rackStore *Rack,
	blades map[int64]*pb.Definition_Blade,
) error {
	for index, bladeFile := range blades {
		bladeStore, err := rackStore.NewBlade(index)
		if err != nil {
			return err
		}

		if err = m.reconcileNewInventoryBlade(ctx, bladeStore, bladeFile); err != nil {
			return err
		}
	}

	return nil
}

func (m *Inventory) reconcileOldInventoryRacks(
	ctx context.Context,
	zone *Zone,
	zoneStore *pb.Definition_Zone,
	zoneFile *pb.Definition_Zone,
) error {
	for rackName, rackStore := range zoneStore.Racks {
		rack, err := zone.NewChild(rackName)
		if err != nil {
			return err
		}

		rackFile, ok := zoneFile.Racks[rackName]

		if !ok {
			if _, err = rack.Delete(ctx, true, ViewDefinition); err != nil {
				return err
			}
		} else {
			if err = m.reconcileOldInventoryPdus(ctx, rack, &rackStore.Pdus, &rackFile.Pdus); err != nil {
				return err
			}

			if err = m.reconcileOldInventoryTors(ctx, rack, &rackStore.Tors, &rackFile.Tors); err != nil {
				return err
			}

			if err = m.reconcileOldInventoryBlades(ctx, rack, &rackStore.Blades, &rackFile.Blades); err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *Inventory) reconcileOldInventoryPdus(
	ctx context.Context,
	rack *Rack,
	pdusStore *map[int64]*pb.Definition_Pdu,
	pdusFile *map[int64]*pb.Definition_Pdu,
) error {
	for index := range *pdusStore {
		pdu, err := rack.NewPdu(index)
		if err != nil {
			return err
		}

		if _, ok := (*pdusFile)[index]; !ok {
			if _, err = pdu.Delete(ctx, true, ViewDefinition); err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *Inventory) reconcileOldInventoryTors(
	ctx context.Context,
	rack *Rack,
	torsStore *map[int64]*pb.Definition_Tor,
	torsFile *map[int64]*pb.Definition_Tor,
) error {
	for index := range *torsStore {
		tor, err := rack.NewTor(index)
		if err != nil {
			return err
		}

		if _, ok := (*torsFile)[index]; !ok {
			if _, err = tor.Delete(ctx, true, ViewDefinition); err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *Inventory) reconcileOldInventoryBlades(
	ctx context.Context,
	rack *Rack,
	bladesStore *map[int64]*pb.Definition_Blade,
	bladesFile *map[int64]*pb.Definition_Blade,
) error {
	for index := range *bladesStore {
		blade, err := rack.NewBlade(index)
		if err != nil {
			return err
		}

		if _, ok := (*bladesFile)[index]; !ok {
			if _, err = blade.Delete(ctx, true, ViewDefinition); err != nil {
				return err
			}
		}
	}

	return nil
}

// reconcileNewInventory compares the newly loaded inventory definition,
// presumably from a configuration file, with the current inventory and
// updates the store accordingly. This will trigger the various watches
// which any currently running services have previously established and
// deliver a set of arrival and/or departure notifications as appropriate.
//
// NOTE: As a temporary measure, reconciliation just deletes the old inventory
//       from the store and completely replaces it with the newly read inventory
//       from the configured file
//
func (m *Inventory) reconcileNewInventory(
	ctx context.Context,
	rootFile *pb.Definition_Root) error {
	ctx, span := tracing.StartSpan(ctx,
		tracing.WithName("Reconcile current inventory with update"),
		tracing.WithContextValue(timestamp.EnsureTickInContext),
		tracing.AsInternal())
	defer span.End()

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// First iterate over the file based inventory and search for items which have
	// been added, or changed. For items that are new, create new items and write
	// them to the store. For updates, update the existing records.
	//
	root, err := m.NewRoot(namespace.DefinitionTable)
	if err != nil {
		return err
	}

	for regionName, regionFile := range rootFile.Regions {
		regionStore, err := root.NewChild(regionName)
		if err != nil {
			return err
		}

		if err = m.reconcileNewInventoryRegion(ctx, regionStore, regionFile); err != nil {
			return err
		}

		for zoneName, zoneFile := range regionFile.Zones {
			zoneStore, err := regionStore.NewChild(zoneName)
			if err != nil {
				return err
			}

			if err = m.reconcileNewInventoryZone(ctx, zoneStore, zoneFile); err != nil {
				return err
			}

			for rackName, rackFile := range zoneFile.Racks {
				rackStore, err := zoneStore.NewChild(rackName)
				if err != nil {
					return err
				}

				if err = m.reconcileNewInventoryRack(ctx, rackStore, rackFile); err != nil {
					return err
				}
			}
		}
	}

	// Now that new items and updates have been processed, iterate over the store
	// based inventory and check that all the items are still present in the file
	// based inventory. If not, then delete the items from the store.
	//
	rootStore, err := m.readInventoryDefinitionFromStore(ctx)

	if err != nil {
		return err
	}

	for regionName, regionStore := range rootStore.Regions {
		region, err := root.NewChild(regionName)
		if err != nil {
			return err
		}

		regionFile, ok := rootFile.Regions[regionName]

		if !ok {
			if _, err = region.Delete(ctx, true, ViewDefinition); err != nil {
				return err
			}
		} else {
			for zoneName, zoneStore := range regionStore.Zones {
				zone, err := region.NewChild(zoneName)
				if err != nil {
					return err
				}

				zoneFile, ok := regionFile.Zones[zoneName]

				if !ok {
					if _, err = zone.Delete(ctx, true, ViewDefinition); err != nil {
						return err
					}
				} else {
					if err = m.reconcileOldInventoryRacks(ctx, zone, zoneStore, zoneFile); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

// updateSummaryInformation is called after any update to the store that might have
// added or removed items from the inventory and generates a new set of summary
// information to reflect those updates.
//
func (m *Inventory) updateSummaryInformation(ctx context.Context) error {

	root, err := m.readInventoryDefinitionFromStore(ctx)

	if err != nil {
		return err
	}

	m.RootSummary, m.DefaultZoneSummary = m.buildSummaryInformation(ctx, root)

	return nil
}

func (m *Inventory) buildSummaryInformation(ctx context.Context, root *pb.Definition_Root) (*RootSummary, *ZoneSummary) {

	var zoneSummary *ZoneSummary

	rootSummary := m.buildSummaryForRoot(root)

	tracing.Info(
		ctx,
		"Updated inventory summary - RegionCount: %d MaxZoneCount: %d MaxRackCount: %d MaxBladeCount: %d MaxCapacity: %v",
		rootSummary.RegionCount,
		rootSummary.MaxZoneCount,
		rootSummary.MaxRackCount,
		rootSummary.MaxBladeCount,
		rootSummary.MaxCapacity)

	zone, err := m.getDefaultZone(root)

	if err != nil {
		zoneSummary = &ZoneSummary{}

		_ = tracing.Error(
			ctx,
			"Reset DEFAULT inventory summary - MaxRackCount: %d MaxBladeCount: %d MaxCapacity: %v - %v",
			zoneSummary.RackCount,
			zoneSummary.MaxBladeCount,
			zoneSummary.MaxCapacity,
			err,
		)
	} else {
		zoneSummary = m.buildSummaryForZone(zone)

		tracing.Info(
			ctx,
			"Updated DEFAULT inventory summary - MaxRackCount: %d MaxBladeCount: %d MaxCapacity: %v",
			zoneSummary.RackCount,
			zoneSummary.MaxBladeCount,
			zoneSummary.MaxCapacity,
		)
	}

	return rootSummary, zoneSummary
}

func (m *Inventory) getDefaultZone(root *pb.Definition_Root) (*pb.Definition_Zone, error) {

	region, ok := root.Regions[DefaultRegion]

	if !ok {
		return nil, errors.ErrRegionNotFound{Region: DefaultRegion}
	}

	zone, ok := region.Zones[DefaultZone]

	if !ok {
		return nil, errors.ErrZoneNotFound{Region: DefaultRegion, Zone: DefaultZone}
	}

	return zone, nil
}

// readInventoryDefinitionFromStore is used to read all the inventory
// definitions from the store, regardless of how they got there and
// return them in the hierarchical in-memory form.
//
func (m *Inventory) readInventoryDefinitionFromStore(ctx context.Context) (*pb.Definition_Root, error) {
	ctx, span := tracing.StartSpan(ctx,
		tracing.WithName("Read inventory definition from store"),
		tracing.WithContextValue(timestamp.EnsureTickInContext),
		tracing.AsInternal())
	defer span.End()

	root, err := m.NewRoot(namespace.DefinitionTable)
	if err != nil {
		return nil, err
	}

	_, regions, err := root.FetchChildren(ctx)
	if err != nil {
		return nil, err
	}

	defRoot := &pb.Definition_Root{
		Details: root.GetDetails(),
		Regions: make(map[string]*pb.Definition_Region, len(*regions)),
	}

	for regionName, region := range *regions {
		_, zones, err := region.FetchChildren(ctx)
		if err != nil {
			return nil, err
		}

		defRegion := &pb.Definition_Region{
			Details: region.GetDetails(),
			Zones:   make(map[string]*pb.Definition_Zone, len(*zones)),
		}

		for zoneName, zone := range *zones {
			_, racks, err := zone.FetchChildren(ctx)
			if err != nil {
				return nil, err
			}

			defZone := &pb.Definition_Zone{
				Details: zone.GetDetails(),
				Racks:   make(map[string]*pb.Definition_Rack, len(*racks)),
			}

			for rackName, rack := range *racks {
				defRack, err := rack.GetDefinitionRackWithChildren(ctx)
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
func (m *Inventory) writeInventoryDefinitionToStore(ctx context.Context, root *pb.Definition_Root) error {
	ctx, span := tracing.StartSpan(ctx,
		tracing.WithName("Write inventory definition to store"),
		tracing.WithContextValue(timestamp.EnsureTickInContext),
		tracing.AsInternal())
	defer span.End()

	storeRoot, err := m.NewRoot(namespace.DefinitionTable)
	if err != nil {
		return err
	}

	storeRoot.SetDetails(root.Details)

	for regionName, region := range root.Regions {
		if err = m.writeOneRegion(ctx, regionName, storeRoot, region); err != nil {
			return err
		}
	}

	return nil
}

// writeOneRegion writes the records associated with the supplied region to the
// store.
func (m *Inventory) writeOneRegion(
	ctx context.Context,
	regionName string,
	storeRoot *Root,
	region *pb.Definition_Region) error {
	ctx, span := tracing.StartSpan(ctx,
		tracing.WithName("Write inventory definition for region %q", regionName),
		tracing.WithContextValue(timestamp.EnsureTickInContext),
		tracing.AsInternal())
	defer span.End()

	storeRegion, err := storeRoot.NewChild(regionName)
	if err != nil {
		return err
	}

	storeRegion.SetDetails(region.GetDetails())

	if _, err = storeRegion.Create(ctx, ViewDefinition); err != nil {
		return err
	}

	for zoneName, zone := range region.Zones {
		if err = m.writeOneZone(ctx, zoneName, storeRegion, zone); err != nil {
			return err
		}
	}

	return nil
}

// writeOneZone writes the records associated with the supplied zone to the store.
func (m *Inventory) writeOneZone(
	ctx context.Context,
	zoneName string,
	storeRegion *Region,
	zone *pb.Definition_Zone) error {
	ctx, span := tracing.StartSpan(ctx,
		tracing.WithName(
			"Write inventory definition for region %q, zone %q",
			storeRegion.Region, zoneName),
		tracing.WithContextValue(timestamp.EnsureTickInContext),
		tracing.AsInternal())
	defer span.End()

	storeZone, err := storeRegion.NewChild(zoneName)
	if err != nil {
		return err
	}

	storeZone.SetDetails(zone.GetDetails())

	if _, err = storeZone.Create(ctx, ViewDefinition); err != nil {
		return err
	}

	for rackName, rack := range zone.Racks {
		if err = m.writeOneRack(ctx, rackName, storeZone, rack); err != nil {
			return err
		}
	}

	return nil
}

// writeOneRack writes the records associated with the supplied rack to the store.
func (m *Inventory) writeOneRack(
	ctx context.Context,
	rackName string,
	storeZone *Zone,
	rack *pb.Definition_Rack) error {
	ctx, span := tracing.StartSpan(ctx,
		tracing.WithName(
			"Write inventory definition for region %q, zone %q, rack %q",
			storeZone.Region, storeZone.Zone, rackName),
		tracing.WithContextValue(timestamp.EnsureTickInContext),
		tracing.AsInternal())
	defer span.End()

	storeRack, err := storeZone.NewChild(rackName)
	if err != nil {
		return err
	}

	storeRack.SetDetails(rack.GetDetails())

	if _, err = storeRack.Create(ctx, ViewDefinition); err != nil {
		return err
	}

	for index, pdu := range rack.Pdus {
		if err = m.writeOnePdu(ctx, index, storeRack, pdu); err != nil {
			return err
		}
	}

	for index, tor := range rack.Tors {
		if err = m.writeOneTor(ctx, index, storeRack, tor); err != nil {
			return err
		}
	}

	for index, blade := range rack.Blades {
		if err = m.writeOneBlade(ctx, index, storeRack, blade); err != nil {
			return err
		}
	}

	return nil
}

// writeOnePdu writes the record associated with the supplied PDU.
func (m *Inventory) writeOnePdu(
	ctx context.Context,
	index int64,
	storeRack *Rack,
	pdu *pb.Definition_Pdu) error {
	ctx, span := tracing.StartSpan(ctx,
		tracing.WithName(
			"Write inventory definition for region %q, zone %q, rack %q, pdu %d",
			storeRack.Region, storeRack.Zone, storeRack.Rack, index),
		tracing.WithContextValue(timestamp.EnsureTickInContext),
		tracing.AsInternal())
	defer span.End()

	storePdu, err := storeRack.NewPdu(index)
	if err != nil {
		return err
	}
	ports := pdu.GetPorts()

	storePdu.SetDetails(pdu.GetDetails())
	storePdu.SetPorts(ports)

	if _, err = storePdu.Create(ctx, ViewDefinition); err != nil {
		return err
	}

	return nil
}

// writeOneTor writes the record associated with the supplied TOR.
func (m *Inventory) writeOneTor(
	ctx context.Context,
	index int64,
	storeRack *Rack,
	tor *pb.Definition_Tor) error {
	ctx, span := tracing.StartSpan(ctx,
		tracing.WithName(
			"Write inventory definition for region %q, zone %q, rack %q, tor %d",
			storeRack.Region, storeRack.Zone, storeRack.Rack, index),
		tracing.WithContextValue(timestamp.EnsureTickInContext),
		tracing.AsInternal())
	defer span.End()

	storeTor, err := storeRack.NewTor(index)
	if err != nil {
		return err
	}

	ports := tor.GetPorts()
	storeTor.SetDetails(tor.GetDetails())
	storeTor.SetPorts(ports)

	if _, err = storeTor.Create(ctx, ViewDefinition); err != nil {
		return err
	}

	return nil
}

// writeOneBlade writes the record associated with the supplied blade.
func (m *Inventory) writeOneBlade(ctx context.Context, index int64, storeRack *Rack, blade *pb.Definition_Blade) error {
	ctx, span := tracing.StartSpan(ctx,
		tracing.WithName(
			"Write inventory definition for region %q, zone %q, rack %q, blade %d",
			storeRack.Region, storeRack.Zone, storeRack.Rack, index),
		tracing.WithContextValue(timestamp.EnsureTickInContext),
		tracing.AsInternal())
	defer span.End()

	storeBlade, err := storeRack.NewBlade(index)
	if err != nil {
		return err
	}

	storeBlade.SetDetails(blade.GetDetails())
	storeBlade.SetCapacity(blade.GetCapacity())
	storeBlade.SetBootInfo(blade.GetBootInfo())
	storeBlade.SetBootPowerOn(blade.GetBootOnPowerOn())

	if _, err = storeBlade.Create(ctx, ViewDefinition); err != nil {
		return err
	}

	return nil
}

// deleteInventoryDefinitionFromStore is used to completely remove all
// inventory definitions from the store as identified by the storeRoot
// parameter, regardless of how they got there.
//
func (m *Inventory) deleteInventoryDefinitionFromStore(ctx context.Context, storeRoot *pb.Definition_Root) error {

	root, err := m.NewRoot(namespace.DefinitionTable)
	if err != nil {
		return err
	}

	for regionName, storeRegion := range storeRoot.Regions {

		region, err := root.NewChild(regionName)
		if err != nil {
			return err
		}

		for zoneName, storeZone := range storeRegion.Zones {

			zone, err := region.NewChild(zoneName)
			if err != nil {
				return err
			}

			for rackName, storeRack := range storeZone.Racks {

				rack, err := zone.NewChild(rackName)
				if err != nil {
					return err
				}

				for i := range storeRack.Pdus {

					pdu, err := rack.NewPdu(i)
					if err != nil {
						return err
					}

					pdu.Delete(ctx, true, ViewDefinition)
				}

				for i := range storeRack.Tors {

					tor, err := rack.NewTor(i)
					if err != nil {
						return err
					}

					tor.Delete(ctx, true, ViewDefinition)
				}

				for i := range storeRack.Blades {

					blade, err := rack.NewBlade(i)
					if err != nil {
						return err
					}

					blade.Delete(ctx, true, ViewDefinition)
				}

				rack.Delete(ctx, true, ViewDefinition)
			}

			zone.Delete(ctx, true, ViewDefinition)
		}

		region.Delete(ctx, true, ViewDefinition)
	}

	return nil
}

// buildSummaryForRoot constructs the memo-ed summary data for the root. This should
// be called whenever the configured inventory changes. This includes
//
// - the region count
// - the maximum number of zones in a region
// - the maximum number of racks in a zone
// - the maximum number of TORs, PDUs, and blades in a rack
// - the maximum number of connectors in any TOR or PDU
// - the maximum blade capacity
//
func (m *Inventory) buildSummaryForRoot(
	root *pb.Definition_Root) *RootSummary {

	summary := RackSizeSummary{
		MaxCapacity: &pb.BladeCapacity{},
	}

	maxZoneCount := 0
	maxRackCount := 0

	for _, region := range root.Regions {
		regionSummary := m.buildSummaryForRegion(region)

		summary.setToMax(regionSummary.RackSizeSummary)
		maxRackCount = common.MaxInt(maxRackCount, regionSummary.MaxRackCount)
		maxZoneCount = common.MaxInt(maxZoneCount, regionSummary.ZoneCount)
	}

	return &RootSummary{
		RegionCount:     len(root.Regions),
		MaxZoneCount:    maxZoneCount,
		MaxRackCount:    maxRackCount,
		RackSizeSummary: summary,
	}
}

// buildSummaryForRegion constructs the memo-ed summary data for the region.
// This should be called whenever the configured inventory changes. This includes
//
// - the zone count
// - the maximum number of racks in a zone
// - the maximum number of TORs, PDUs, and blades in a rack
// - the maximum number of connectors in any TOR or PDU
// - the maximum blade capacity
//
func (m *Inventory) buildSummaryForRegion(
	region *pb.Definition_Region) *RegionSummary {

	summary := RackSizeSummary{
		MaxCapacity: &pb.BladeCapacity{},
	}

	maxRackCount := 0

	for _, zone := range region.Zones {
		zoneSummary := m.buildSummaryForZone(zone)
		summary.setToMax(zoneSummary.RackSizeSummary)

		maxRackCount = common.MaxInt(maxRackCount, zoneSummary.RackCount)
	}

	return &RegionSummary{
		ZoneCount:       len(region.Zones),
		MaxRackCount:    maxRackCount,
		RackSizeSummary: summary,
	}
}

// buildSummaryForZone constructs the memo-ed summary data for the zone. This
// should be called whenever the configured inventory changes. This includes
//
// - the number of racks in a zone
// - the maximum number of TORs, PDUs, and blades in a rack
// - the maximum number of connectors in any TOR or PDU
// - the maximum blade capacity
//
func (m *Inventory) buildSummaryForZone(zone *pb.Definition_Zone) *ZoneSummary {

	summary := RackSizeSummary{
		MaxCapacity: &pb.BladeCapacity{},
	}

	for _, rack := range zone.Racks {
		summary.setToMax(m.buildSummaryForRack(rack))
	}

	return &ZoneSummary{
		RackCount:       len(zone.Racks),
		RackSizeSummary: summary,
	}
}

// buildSummaryForZone constructs the memo-ed summary data for the rack. This
// should be called whenever the configured inventory changes. This includes
//
// - the number of TORs, PDUs, and blades in a rack
// - the maximum number of connectors in any TOR or PDU
// - the maximum blade capacity
//
func (m *Inventory) buildSummaryForRack(rack *pb.Definition_Rack) RackSizeSummary {

	memo := &pb.BladeCapacity{}

	for _, blade := range rack.Blades {
		memo = maxBladeCapacity(memo, blade.Capacity)
	}

	connectors := 0
	for _, tor := range rack.Tors {
		connectors = common.MaxInt(connectors, len(tor.Ports))
	}

	for _, pdu := range rack.Pdus {
		connectors = common.MaxInt(connectors, len(pdu.Ports))
	}

	return RackSizeSummary{
		MaxTorCount:   len(rack.Tors),
		MaxPduCount:   len(rack.Pdus),
		MaxBladeCount: len(rack.Blades),
		MaxConnectors: connectors,
		MaxCapacity:   memo,
	}
}
