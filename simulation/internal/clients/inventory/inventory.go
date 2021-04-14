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
// racks within a zone. To achieve this, the pacakge maintains an index for
// child objects which is kept separately from the objects themselves. Thus
// when creating a new object, the package will also create an appropriate
// index entry to allow that object to be discovered from its parent. In
// this way, by knowing the root of a namespace, all the regions can be
// located. Once a region is know, all the zones within that region can be
// located. And so on for racks within zones, etc.
//
// NOTE: Currently there is no way to discover a parent for a given oject.
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
	// while the transition to the new inventory extended schemaa continues.
	// Eventually these will dissapear as the front-end and higher layers learn
	// abouts regions, zones, multiple pdus and tors.
	//
	DefaultRegion = "standard"

	// DefaultZone is used to provide a value for the non-existing zone name
	// while the transition to the new inventory extended schemaa continues.
	// Eventually these will dissapear as the front-end and higher layers learn
	// abouts regions, zones, multiple pdus and tors.
	//
	DefaultZone   = "standard"

	// DefaultPdu is used to provide a value for the non-existing pdu ID
	// while the transition to the new inventory extended schemaa continues.
	// Eventually these will dissapear as the front-end and higher layers learn
	// abouts regions, zones, multiple pdus and tors.
	//
	DefaultPdu    = int64(0)

	// DefaultTor is used to provide a value for the non-existing tor ID
	// while the transition to the new inventory extended schemaa continues.
	// Eventually these will dissapear as the front-end and higher layers learn
	// abouts regions, zones, multiple pdus and tors.
	//
	DefaultTor    = int64(0)

)

type inventoryRevision interface {
	// GetRevision returns the revision of the details field within the object.
	// This will be either the revision of the object in the store after a
	// Create(), Read() or Update() call or be store.RevisionInvalid if the
	// details have been set or no Create(), Read() or Update() call has been
	// executed.
	//
	GetRevision() int64

	// GetRevisionRecord returns the revision of the underlying store object as
	// determined at the time of the last Create(), Read() or Update() for the
	// object. The record revision is not reset by a SetDetails() call and is
	// used when performing either a conditional update or conditional delete
	// using the object.
	//
	GetRevisionRecord() int64

	// GetRevisionStore returns the revison of the underlying store ifself as
	// determined at the time of the last Create() Read() for the object. The
	// store revision is not reset by a SetDetails() call and is provided
	// for information only.
	//
	GetRevisionStore() int64

	// GetRevisionForRequest returns the appropriate revision for the update
	// for either a conditional update based upon the revision of the most
	// recently read record, or an unconditional update.
	//
	GetRevisionForRequest(unconditional bool) int64

	// resetRevision resets the revision for the details field within the object.
	// Subsequent calls to GetRevision() will return store.RevisionInvalid until
	// a successful call is made to one of the routines which invoke the store
	//
	resetRevision() int64

	// updateRevision is used to set/update the current revision information
	// as part of a successful invokation of a store routine.
	//
	updateRevisionInfo(rev int64) int64
}

// Region, zone and rack are "containers" whereas tor, pdu and blade are "things".
// You can send operations and commands to things, but not containers.
//
// Of the following SetXxx() and GetXxx() methods fetch or set values within the
// fields of an object (cursor) and do NOT perform any operations to the underlying
// store.
//
// the Create(), Read(), Update() and Delete() methods perform the appropriate
// operation to the persisted object based upon the current values of fields
// within the object (cursor) being used and will return an error if needed fields
// are not set.
//
type inventoryItem interface {
	inventoryRevision

	// Use to set the attribues of an object within the inventory
	//
	SetDetails(ctx context.Context, details *interface{})
	GetDetails(ctx context.Context) *interface{}

	// Create uses the current object to persist the object to the underlying
	// store  and also create any index entries that may be required.
	//
	// This is not valid to call on a root object and doing so will return
	// an error.
	//
	Create(ctx context.Context) (int64, error)

	// Read issues a request to the underlying store to populate all the fields
	// within the cursor object, including any attributes for that object or
	// other information specific to that object.
	//
	// This is not valid to call on a root object and doing so will return
	// an error.
	//
	Read(ctx context.Context) (int64, error)

	// Update will write a record the underlying store using the currenty
	// information  in the fields of the object. The update can be either
	// unconditional by setting the unconditional parameter to true, or
	// conditional based on the revision of the object compared to the
	// revision of the associated record in the underlying store.
	//
	// Note the object maintains revision information returned from the
	// store for any actions involving the store, e.g. Create(), Read() or
	// Update().
	//
	Update(ctx context.Context, unconditional bool) (int64, error)

	// Delete is used to remove the persisted copy of the object from the
	// store along with any index information needed to navigate to or
	// through that object. The delete can be either unconditional by
	// setting the unconditional parameter to true, or conditional based
	// on the revision of the object compared to the revision of the
	// associated record in the underlying store.
	//
	// Note the object maintains revision information returned from the
	// store for any actions involving the store, e.g. Create(), Read() or
	// Update().
	//
	Delete(ctx context.Context, unconditional bool) (int64, error)
}

// A node object within the inventory is one which in addition to having its own
// attributes also allows for navigation of the namespace
//
type inventoryItemNode interface {
	inventoryItem

	// NewChild creates a child of the current object but uses only on the
	// current object to create a new child object and no store
	// operations are involved.
	//
	// Note, this method does not populate the attributes of the new child
	// object. To retrieve the attributes a Read() using the child must be
	// performed.
	//
	NewChild(name string) (*interface{}, error)

	// ListChildren uses the current object to discover the names of all the
	// child objects of the current object, The elements of the returned list
	// can be used in subsequent operations to create child objects.
	//
	ListChildren(ctx context.Context) (int64, *[]string, error)

	// FetchChildren is used to locate all the children of the current object
	// and to generate an object for each of those children. It is a
	// convenience wrapper around ListChildren() followed by a NewChild() on
	// each name discovered.
	//
	FetchChildren(ctx context.Context) (int64, *map[string]interface{}, error)
}

// Provide a set of definitions to cope with calls to a "null" object.
//
func (n *nullItem) GetRevision() int64 {
	return store.RevisionInvalid
}

func (n *nullItem) GetRevisionRecord() int64 {
	return store.RevisionInvalid
}

func (n *nullItem) GetRevisionStore() int64 {
	return store.RevisionInvalid
}

func (n *nullItem) GetRevisionForRequest(unconditional bool) int64 {
	return store.RevisionInvalid
}

func (n *nullItem) resetRevision() int64 {
	return store.RevisionInvalid
}

func (n *nullItem) updateRevisionInfo(rev int64) int64 {
	return store.RevisionInvalid
}

type inventoryItemRack interface {
	inventoryItemNode

	NewPdu(name string) (*interface{}, error)
	NewTor(name string) (*interface{}, error)
	NewBlade(name string) (*interface{}, error)

	ListPdus(ctx context.Context) (int64, *map[int64]*interface{}, error)
	ListTors(ctx context.Context) (int64, *map[int64]*interface{}, error)
	ListBlades(ctx context.Context) (int64, *map[int64]*interface{}, error)

	FetchPdus(ctx context.Context) (int64, *map[int64]*interface{}, error)
	FetchTors(ctx context.Context) (int64, *map[int64]*interface{}, error)
	FetchBlades(ctx context.Context) (int64, *map[int64]*interface{}, error)
}

type inventoryItemPdu interface {
	inventoryItem

	SetPorts(ports *map[int64]*pb.PowerPort)
	GetPorts() *map[int64]*pb.PowerPort
}

type inventoryTor interface {
	inventoryItem

	SetPorts(cports *map[int64]*pb.NetworkPort)
	GetPorts() *map[int64]*pb.NetworkPort
}

type inventoryBlade interface {
	inventoryItem

	SetCapacity(capacity *pb.BladeCapacity)
	GetCapacity() *pb.BladeCapacity

	SetBootInfo(bootInfo *pb.BladeBootInfo)
	GetBootInfo() *pb.BladeBootInfo

	SetBootPowerOn(bootOnPowerOn bool)
	GetBootOnPowerOn() bool

}

// Provide a set of definitions to cope with calls to a "null" object.
//
type nullItem struct{}

func (n *nullItem) SetDetails(details *nullItem) {
}

func (n *nullItem) GetDetails() *nullItem {
	return nil
}

func (n *nullItem) Create(ctx context.Context) (int64, error) {
	return store.RevisionInvalid, errors.ErrNullItem
}

func (n *nullItem) Read(ctx context.Context) (int64, error) {
	return store.RevisionInvalid, errors.ErrNullItem
}

func (n *nullItem) Update(ctx context.Context) (int64, error) {
	return store.RevisionInvalid, errors.ErrNullItem
}

func (n *nullItem) Delete(ctx context.Context) (int64, error) {
	return store.RevisionInvalid, errors.ErrNullItem
}

// Additional functions for the node specialization of the basic inventory item
//
func (n *nullItem) NewChild(name string) (*interface{}, error) {
	return nil, errors.ErrNullItem
}

func (n *nullItem) ListChildren(ctx context.Context) (int64, *[]string, error) {
	return store.RevisionInvalid, nil, errors.ErrNullItem
}
func (n *nullItem) FetchChildren(ctx context.Context) (int64, *map[string]interface{}, error) {
	return store.RevisionInvalid, nil, errors.ErrNullItem
}

// Additional functions for the rack specialization of the basic inventory item
//
func (n *nullItem) NewPdu(name string) (*interface{}, error) {
	return nil, errors.ErrNullItem
}

func (n *nullItem) NewTor( name string) (*interface{}, error) {
	return nil, errors.ErrNullItem
}

func (n *nullItem) NewBlade(name string) (*interface{}, error) {
	return nil, errors.ErrNullItem
}

func (n *nullItem) ListPdus(ctx context.Context) (int64, *map[int64]*interface{}, error) {
	return store.RevisionInvalid, nil, errors.ErrNullItem
}

func (n *nullItem) ListTors(ctx context.Context) (int64, *map[int64]*interface{}, error) {
	return store.RevisionInvalid, nil, errors.ErrNullItem
}

func (n *nullItem) ListBlades(ctx context.Context) (int64, *map[int64]*interface{}, error) {
	return store.RevisionInvalid, nil, errors.ErrNullItem
}

func (n *nullItem) FetchPdus(ctx context.Context) (int64, *[]string, error) {
	return store.RevisionInvalid, nil, errors.ErrNullItem
}

func (n *nullItem) FetchTors(ctx context.Context) (int64, *[]string, error) {
	return store.RevisionInvalid, nil, errors.ErrNullItem
}

func (n *nullItem) FetchBlades(ctx context.Context) (int64, *[]string, error) {
	return store.RevisionInvalid, nil, errors.ErrNullItem
}

// Additional functions for the pdu and tor specializations of the basic inventory item
//
func (n *nullItem) SetPorts(ports *map[int64]*interface{}) {
}

func (n *nullItem) GetPorts() *map[int64]*interface{} {
	return nil
}

// Additional functions for the blade specialization of the basic inventory item
//
func (n *nullItem) SetCapacity(capacity *interface{}) {
}

func (n *nullItem) GetCapacity() *interface{} {
	return nil
}

func (n *nullItem) SetBootInfo(bootOnPowerOn bool, bootInfo *interface{}) {
}

func (n *nullItem) GetBootInfo() (bool, *interface{}) {
	return false, nil
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

// Inventory is a structure used to estblished synchronized access to values
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
		mutex:              sync.RWMutex{},
		cfg:                cfg,
		Store:              store,
		RootSummary:        &RootSummary{},
		DefaultZoneSummary: &ZoneSummary{
			MaxCapacity:   &pb.BladeCapacity{
				Accelerators: []*pb.Accelerator{},
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
	ctx         context.Context,
	regionStore *Region,
	regionFile  *pb.Definition_Region,
) error {
	_, err := regionStore.Read(ctx)

	if err == error(errors.ErrRegionNotFound{Region: regionStore.Region}) {
		tracing.Info(ctx, "creating new region %s", regionStore.Region)

		regionStore.SetDetails(regionFile.GetDetails())

		_, err = regionStore.Create(ctx)
	} else if err == nil && regionStore.NotEqual(regionFile) {
		tracing.Info(ctx, "updating region %s", regionStore.Region)

		regionStore.SetDetails(regionFile.GetDetails())

		_, err = regionStore.Update(ctx, true)
	}

	return err
}

func (m *Inventory) reconcileNewInventoryZone(
	ctx         context.Context,
	zoneStore   *Zone,
	zoneFile    *pb.Definition_Zone,
) error {
	_, err := zoneStore.Read(ctx)

	if err == error(errors.ErrZoneNotFound{Region: zoneStore.Region, Zone: zoneStore.Zone}) {
		tracing.Info(ctx, "creating new zone %s/%s", zoneStore.Region, zoneStore.Zone)

		zoneStore.SetDetails(zoneFile.GetDetails())

		_, err = zoneStore.Create(ctx)
	} else if err == nil && zoneStore.NotEqual(zoneFile) {
		tracing.Info(ctx, "updating zone %s/%s", zoneStore.Region, zoneStore.Zone)

		zoneStore.SetDetails(zoneFile.GetDetails())

		_, err = zoneStore.Update(ctx, true)
	}

	return err
}

func (m *Inventory) reconcileNewInventoryRack(
	ctx         context.Context,
	rackStore   *Rack,
	rackFile    *pb.Definition_Rack,
) error {
	_, err := rackStore.Read(ctx)

	if err == error(errors.ErrRackNotFound{Region: rackStore.Region, Zone: rackStore.Zone, Rack: rackStore.Rack}) {
		tracing.Info(ctx, "creating new rack %s/%s/%s", rackStore.Region, rackStore.Zone, rackStore.Rack)

		rackStore.SetDetails(rackFile.GetDetails())

		_, err = rackStore.Create(ctx)
	} else if err == nil && rackStore.NotEqual(rackFile) {
		tracing.Info(ctx, "updating rack %s/%s/%s", rackStore.Region, rackStore.Zone, rackStore.Rack)

		rackStore.SetDetails(rackFile.GetDetails())

		_, err = rackStore.Update(ctx, true)
	}

	if err == nil {
		if err = m.reconcileNewInventoryPdus(ctx, rackStore, &rackFile.Pdus); err != nil {
			return err
		}

		if err = m.reconcileNewInventoryTors(ctx, rackStore, &rackFile.Tors); err != nil {
			return err
		}

		err = m.reconcileNewInventoryBlades(ctx, rackStore, &rackFile.Blades)
	}

	return err
}

func (m *Inventory) reconcileNewInventoryPdu(
	ctx      context.Context,
	pduStore *Pdu,
	pduFile  *pb.Definition_Pdu,
) error {
	_, err := pduStore.Read(ctx)

	if err == error(errors.ErrPduNotFound{Region: pduStore.Region, Zone: pduStore.Zone, Rack: pduStore.Rack, Pdu: pduStore.ID}) {
		tracing.Info(ctx, "creating new pdu %s/%s/%s/%d", pduStore.Region, pduStore.Zone, pduStore.Rack, pduStore.ID)

		ports := pduFile.GetPorts()
		pduStore.SetPorts(&ports)
		pduStore.SetDetails(pduFile.GetDetails())

		_, err = pduStore.Create(ctx)
	} else if err == nil && pduStore.NotEqual(pduFile) {
		tracing.Info(ctx, "updating pdu %s/%s/%s/%d", pduStore.Region, pduStore.Zone, pduStore.Rack, pduStore.ID)

		ports := pduFile.GetPorts()
		pduStore.SetPorts(&ports)
		pduStore.SetDetails(pduFile.GetDetails())

		_, err = pduStore.Update(ctx, true)
	}

	return err
}

func (m *Inventory) reconcileNewInventoryTor(
	ctx      context.Context,
	torStore *Tor,
	torFile  *pb.Definition_Tor,
) error {
	_, err := torStore.Read(ctx)

	if err == error(errors.ErrTorNotFound{Region: torStore.Region, Zone: torStore.Zone, Rack: torStore.Rack, Tor: torStore.ID}) {
		tracing.Info(ctx, "creating new tor %s/%s/%s/%d", torStore.Region, torStore.Zone, torStore.Rack, torStore.ID)

		ports := torFile.GetPorts()
		torStore.SetPorts(&ports)
		torStore.SetDetails(torFile.GetDetails())

		_, err = torStore.Create(ctx)
	} else if err == nil && torStore.NotEqual(torFile) {
		tracing.Info(ctx, "updating tor %s/%s/%s/%d", torStore.Region, torStore.Zone, torStore.Rack, torStore.ID)

		ports := torFile.GetPorts()
		torStore.SetPorts(&ports)
		torStore.SetDetails(torFile.GetDetails())

		_, err = torStore.Update(ctx, true)
	}

	return err
}

func (m *Inventory) reconcileNewInventoryBlade(
	ctx      context.Context,
	bladeStore *Blade,
	bladeFile  *pb.Definition_Blade,
) error {
	_, err := bladeStore.Read(ctx)

	if err == error(errors.ErrBladeNotFound{Region: bladeStore.Region, Zone: bladeStore.Zone, Rack: bladeStore.Rack, Blade: bladeStore.ID}) {
		tracing.Info(ctx, "creating new blade %s/%s/%s/%d", bladeStore.Region, bladeStore.Zone, bladeStore.Rack, bladeStore.ID)

		bladeStore.SetDetails(bladeFile.GetDetails())
		bladeStore.SetCapacity(bladeFile.GetCapacity())
		bladeStore.SetBootInfo(bladeFile.GetBootInfo())
		bladeStore.SetBootPowerOn(bladeFile.GetBootOnPowerOn())

		_, err = bladeStore.Create(ctx)
	} else if err == nil && bladeStore.NotEqual(bladeFile) {
		tracing.Info(ctx, "updating blade %s/%s/%s/%d", bladeStore.Region, bladeStore.Zone, bladeStore.Rack, bladeStore.ID)

		bladeStore.SetDetails(bladeFile.GetDetails())
		bladeStore.SetCapacity(bladeFile.GetCapacity())
		bladeStore.SetBootInfo(bladeFile.GetBootInfo())
		bladeStore.SetBootPowerOn(bladeFile.GetBootOnPowerOn())

		_, err = bladeStore.Update(ctx, true)
	}

	return err
}

func (m *Inventory) reconcileNewInventoryPdus(
	ctx       context.Context,
	rackStore *Rack,
	pdus      *map[int64]*pb.Definition_Pdu,
) error {
	for index, pduFile := range *pdus {
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
	ctx       context.Context,
	rackStore *Rack,
	tors      *map[int64]*pb.Definition_Tor,
) error {
	for index, torFile := range *tors {
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
	ctx       context.Context,
	rackStore *Rack,
	blades    *map[int64]*pb.Definition_Blade,
) error {
	for index, bladeFile := range *blades {
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
	ctx       context.Context,
	zone      *Zone,
	zoneStore *pb.Definition_Zone,
	zoneFile  *pb.Definition_Zone,
) error {
	for rackName, rackStore := range zoneStore.Racks {
		rack, err := zone.NewChild(rackName)
		if err != nil {
			return err
		}

		rackFile, ok := zoneFile.Racks[rackName]

		if !ok {
			if _, err = rack.Delete(ctx, true); err != nil {
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
	ctx       context.Context,
	rack      *Rack,
	pdusStore *map[int64]*pb.Definition_Pdu,
	pdusFile  *map[int64]*pb.Definition_Pdu,
) error {
	for index := range *pdusStore {
		pdu, err := rack.NewPdu(index)
		if err != nil {
			return err
		}

		_, ok := (*pdusFile)[index]
		if !ok {
			if _, err = pdu.Delete(ctx, true); err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *Inventory) reconcileOldInventoryTors(
	ctx       context.Context,
	rack      *Rack,
	torsStore *map[int64]*pb.Definition_Tor,
	torsFile  *map[int64]*pb.Definition_Tor,
) error {
	for index := range *torsStore {
		tor, err := rack.NewTor(index)
		if err != nil {
			return err
		}

		_, ok := (*torsFile)[index]
		if !ok {
			if _, err = tor.Delete(ctx, true); err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *Inventory) reconcileOldInventoryBlades(
	ctx         context.Context,
	rack        *Rack,
	bladesStore *map[int64]*pb.Definition_Blade,
	bladesFile  *map[int64]*pb.Definition_Blade,
) error {
	for index := range *bladesStore {
		blade, err := rack.NewBlade(index)
		if err != nil {
			return err
		}

		_, ok := (*bladesFile)[index]
		if !ok {
			if _, err = blade.Delete(ctx, true); err != nil {
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
			if _, err = region.Delete(ctx, true); err != nil {
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
					if _, err = zone.Delete(ctx, true); err != nil {
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

		tracing.Error(
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
		ctx, span := tracing.StartSpan(ctx,
			tracing.WithName(fmt.Sprintf("Write inventory definition for region %q", regionName)),
			tracing.WithContextValue(timestamp.EnsureTickInContext),
			tracing.AsInternal())
		defer span.End()

		storeRegion, err := storeRoot.NewChild(regionName)
		if err != nil {
			return err
		}

		storeRegion.SetDetails(region.GetDetails())

		if _, err = storeRegion.Create(ctx); err != nil {
			return err
		}

		for zoneName, zone := range region.Zones {
			ctx, span := tracing.StartSpan(ctx,
				tracing.WithName(fmt.Sprintf("Write inventory definition for region %q, zone %q", regionName, zoneName)),
				tracing.WithContextValue(timestamp.EnsureTickInContext),
				tracing.AsInternal())
			defer span.End()

			storeZone, err := storeRegion.NewChild(zoneName)
			if err != nil {
				return err
			}

			storeZone.SetDetails(zone.GetDetails())

			if _, err = storeZone.Create(ctx); err != nil {
				return err
			}

			for rackName, rack := range zone.Racks {
				ctx, span := tracing.StartSpan(ctx,
					tracing.WithName(fmt.Sprintf("Write inventory definition for region %q, zone %q, rack %q", regionName, zoneName, rackName)),
					tracing.WithContextValue(timestamp.EnsureTickInContext),
					tracing.AsInternal())
				defer span.End()

				storeRack, err := storeZone.NewChild(rackName)
				if err != nil {
					return err
				}

				storeRack.SetDetails(rack.GetDetails())

				if _, err = storeRack.Create(ctx); err != nil {
					return err
				}

				for index, pdu := range rack.Pdus {
					ctx, span := tracing.StartSpan(ctx,
						tracing.WithName(fmt.Sprintf("Write inventory definition for region %q, zone %q, rack %q, pdu %d", regionName, zoneName, rackName, index)),
						tracing.WithContextValue(timestamp.EnsureTickInContext),
						tracing.AsInternal())
					defer span.End()

					storePdu, err := storeRack.NewPdu(index)
					if err != nil {
						return err
					}
					ports := pdu.GetPorts()

					storePdu.SetDetails(pdu.GetDetails())
					storePdu.SetPorts(&ports)

					if _, err = storePdu.Create(ctx); err != nil {
						return err
					}
				}

				for index, tor := range rack.Tors {
					ctx, span := tracing.StartSpan(ctx,
						tracing.WithName(fmt.Sprintf("Write inventory definition for region %q, zone %q, rack %q, tor %d", regionName, zoneName, rackName, index)),
						tracing.WithContextValue(timestamp.EnsureTickInContext),
						tracing.AsInternal())
					defer span.End()

					storeTor, err := storeRack.NewTor(index)
					if err != nil {
						return err
					}

					ports := tor.GetPorts()
					storeTor.SetDetails(tor.GetDetails())
					storeTor.SetPorts(&ports)

					if _, err = storeTor.Create(ctx); err != nil {
						return err
					}
				}

				for index, blade := range rack.Blades {
					ctx, span := tracing.StartSpan(ctx,
						tracing.WithName(fmt.Sprintf("Write inventory definition for region %q, zone %q, rack %q, blade %d", regionName, zoneName, rackName, index)),
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

					if _, err = storeBlade.Create(ctx); err != nil {
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

					pdu.Delete(ctx, true)
				}

				for i := range storeRack.Tors {

					tor, err := rack.NewTor(i)
					if err != nil {
						return err
					}

					tor.Delete(ctx, true)
				}

				for i := range storeRack.Blades {

					blade, err := rack.NewBlade(i)
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

// buildSummaryForRoot constructs the memo-ed summary data for the root. This should
// be called whenever the configured inventory changes. This includes
//
// - the zone count
// - the maximum number of blades in a rack
// - the memo data itself
//
func (m *Inventory) buildSummaryForRoot(
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

// buildSummary constructs the memo-ed summary data for the zone.  This should
// be called whenever the configured inventory changes. This includes
//
// - the zone count
// - the maximum number of blades in a rack
// - the memo data itself
//
func (m *Inventory) buildSummaryForRegion(
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

func (m *Inventory) buildSummaryForZone(zone *pb.Definition_Zone) *ZoneSummary {

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

func (m *Inventory) buildSummaryForRack(rack *pb.Definition_Rack) (int, *pb.BladeCapacity) {

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
