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

	"github.com/Jim3Things/CloudChamber/internal/clients/store"
	clients "github.com/Jim3Things/CloudChamber/internal/clients/timestamp"
	"github.com/Jim3Things/CloudChamber/internal/common"
	"github.com/Jim3Things/CloudChamber/internal/config"
	"github.com/Jim3Things/CloudChamber/internal/tracing"
	ct "github.com/Jim3Things/CloudChamber/pkg/protos/common"
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
	mutex sync.RWMutex

	Zone *pb.ExternalZone

	MaxBladeCount int64
	MaxCapacity   *ct.BladeCapacity
	Store *store.Store
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
	ctx, span := tracing.StartSpan(ctx,
		tracing.WithName("Initialize Inventory DB Connection"),
		tracing.WithContextValue(clients.EnsureTickInContext),
		tracing.AsInternal())
	defer span.End()

	if dbInventory == nil {
		dbInventory = &DBInventory{
			mutex: sync.RWMutex{},
			Zone: nil,
			MaxBladeCount: 0,
			MaxCapacity:   &ct.BladeCapacity{},
			Store: store.NewStore(),
		}

		if err = dbInventory.Store.Connect(); err != nil {
			return err
		}

		if err = dbInventory.LoadFromStore(ctx); err != nil {
			return err
		}

		if err = dbInventory.UpdateFromFile(ctx, cfg); err != nil {
			return err
		}
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

	return nil
}

// UpdateFromFile is a method to load a new inventory definition from the configured
// file. Once read, the store will be updated with the differences which will in
// turn trigger a set of previously established watch routines to issue a number or
// arrival and/or departure notifications.
//
func (m *DBInventory) UpdateFromFile(ctx context.Context, cfg *config.GlobalConfig) error {
	ctx, span := tracing.StartSpan(ctx,
		tracing.WithName("Initialize User DB Connection"),
		tracing.WithContextValue(clients.EnsureTickInContext),
		tracing.AsInternal())
	defer span.End()


	// We have the basic initialization done. Now go read the current inventory
	// from the file indicated by the configuration. Once we have that, load it
	// into the store looking to see if there are any material changes between
	// what is already in the store and what is now found in the file.
	//
	zone, err := config.ReadInventoryDefinition(cfg.Inventory.InventoryDefinition) 
	if err != nil {
			return err 
	}

	if err = m.reconcileNewInventory(ctx, zone); err != nil {
		return err
	}

	return nil
}


// reconcileNewInventory compares the newly loaded inventory definition,
// presumably from a configuration file, with the currently loaded inventory
// and updates the store accordinly. This will trigger the various watches
// which any currently running services have previously established and deliver
// a set of arrival and/or departure notifications as appropriate.
//
func (m *DBInventory) reconcileNewInventory(ctx context.Context, zone *pb.ExternalZone) error {
	ctx, span := tracing.StartSpan(ctx,
		tracing.WithName("Reconcile current inventory with update"),
		tracing.WithContextValue(clients.EnsureTickInContext),
		tracing.AsInternal())
	defer span.End()

	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.Zone = zone

	m.buildSummary(ctx)

	return nil
}

// GetMemoData returns the maximum number of blades held in any rack
// in the inventory.
func (m *DBInventory) GetMemoData() (int, int64, *ct.BladeCapacity) {
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
func (m *DBInventory) GetRack(rackID string) (*pb.ExternalRack, error) {
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
func (m *DBInventory) GetBlade(rackID string, bladeID int64) (*ct.BladeCapacity, error) {
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
	memo := &ct.BladeCapacity{}

	for _, rack := range m.Zone.Racks {
		for _, blade := range rack.Blades {
			memo.Cores = common.MaxInt64(memo.Cores, blade.Cores)
			memo.DiskInGb = common.MaxInt64(memo.DiskInGb, blade.DiskInGb)
			memo.MemoryInMb = common.MaxInt64(memo.MemoryInMb, blade.MemoryInMb)
			memo.NetworkBandwidthInMbps = common.MaxInt64(
				memo.NetworkBandwidthInMbps,
				blade.NetworkBandwidthInMbps)
		}

		maxBladeCount = common.MaxInt64(maxBladeCount,	int64(len(rack.Blades)))
	}

	m.MaxBladeCount = maxBladeCount
	m.MaxCapacity   = memo

	tracing.Info(ctx, "   Updated inventory summary - MaxBladeCount: %d MaxCapacity: %v", m.MaxBladeCount, m.MaxCapacity)
}


// Condition describes the currtent operational condition of an item in the inventory.
//
type Condition int

// Thw set of available conditions used to describe items in the inventory. Related
// to the conditions defined in inventory.proto
//
const (
	ConditionNotInService Condition = iota
	ConditionOperational
	ConditionSuspect
	ConditionOutForRepair
	ConditionRetiring
	ConditionRetired
)


// DefinitionTor describes the revision and value for the request TOR
//
type DefinitionTor struct {
	revision int64
	tor *pb.DefinitionTor
}

// DefinitionPdu describes the revision and value for the request PDU
//
type DefinitionPdu struct {
	revision int64
	pdu *pb.DefinitionPdu
}

// DefinitionBlade describes the revision and value for the request blade
//
type DefinitionBlade struct {
	revision int64
	blade *pb.DefinitionBlade
}

// DefinitionRack describes the revision and value for the request rack
//
type DefinitionRack struct {
	revision int64
	rack *pb.DefinitionRack
}

// DefinitionZone defines the set of values used to summarize the contents of a zone to
// allow a query on the zone without incorporating the entire inventory definition
//
type DefinitionZone struct {
	revision int64
	zone *pb.DefinitionZone
}

// InventoryOptions is a struct 
//
type InventoryOptions struct {
	revision int64
	includeRackValues bool
	includeBladeValues bool
	includePduValues bool
	includeTorValues bool
	includeRackLocation bool
	includeRackNotes bool
	includeZoneLocation bool
	includeZoneNotes bool
}

func (options *InventoryOptions) applyOpts(optionsArray []InventoryOption) {
	for _, option := range optionsArray {
		option(options)
	}
}



// InventoryOption is a 
// 
type InventoryOption func(*InventoryOptions)

// WithRevision is a
//
func WithRevision(rev int64) InventoryOption {
	return func(options *InventoryOptions) {options.revision = rev}
}

// WithRackValues is a
//
func WithRackValues() InventoryOption {
	return func(options *InventoryOptions) {options.includeRackValues = true}
}

// WithBladeValues is a
//
func WithBladeValues() InventoryOption {
	return func(options *InventoryOptions) {options.includeBladeValues = true}
}

// WithTorValues is a
//
func WithTorValues() InventoryOption {
	return func(options *InventoryOptions) {options.includeTorValues = true}
}

// WithPduValues is a
//
func WithPduValues() InventoryOption {
	return func(options *InventoryOptions) {options.includePduValues = true}
}

// WithRackNotes is a 
//
func WithRackNotes() InventoryOption {
	return func(options *InventoryOptions) {options.includeRackNotes = true}
}

// WithZoneNotes is a 
//
func WithZoneNotes() InventoryOption {
	return func(options *InventoryOptions) {options.includeZoneNotes = true}
}

// WithRackLocation is a
//
func WithRackLocation() InventoryOption {
	return func(options *InventoryOptions) {options.includeRackLocation = true}
}

// WithZoneLocation is a
//
func WithZoneLocation() InventoryOption {
	return func(options *InventoryOptions) {options.includeZoneLocation = true}
}





// ListZones returns
//
func (m *DBInventory) ListZones(ctx context.Context, options ...InventoryOption) (map[string]*DefinitionZone, int64, error) {return nil, InvalidRev, nil}

// ListRacks returns
//
func (m *DBInventory) ListRacks(ctx context.Context, zone string, options ...InventoryOption) (map[string]*DefinitionRack, int64, error) {return nil, InvalidRev, nil}

// ListPdus returns
//
func (m *DBInventory) ListPdus(ctx context.Context, zone string, rack string, options ...InventoryOption) (map[string]*DefinitionPdu, int64, error) {return nil, InvalidRev, nil}

// ListTors returns
//
func (m *DBInventory) ListTors(ctx context.Context, zone string, rack string, options ...InventoryOption) (map[string]*DefinitionTor, int64, error) {return nil,InvalidRev,  nil}

// ListBlades returns
//
func (m *DBInventory) ListBlades(ctx context.Context, zone string, rack string, options ...InventoryOption) (map[string]*DefinitionBlade, int64, error) {return nil, InvalidRev, nil}

// CreateZone is used
//
func (m *DBInventory) CreateZone(ctx context.Context, zone string, u *pb.DefinitionZone, options ...InventoryOption) (int64, error) {return InvalidRev, nil}

// CreateRack is used
//
func (m *DBInventory) CreateRack(ctx context.Context, zone string, rack string, u *pb.DefinitionRack, options ...InventoryOption) (int64, error) {return InvalidRev, nil}

// CreatePdu is used
//
func (m *DBInventory) CreatePdu(ctx context.Context, zone string, rack string, pdu int64, u *pb.DefinitionPdu, options ...InventoryOption) (int64, error) {return InvalidRev, nil}

// CreateTor is used
//
func (m *DBInventory) CreateTor(ctx context.Context, zone string, rack string, tor int64, u *pb.DefinitionTor, options ...InventoryOption) (int64, error) {return InvalidRev, nil}

// CreateBlade is used
//
func (m *DBInventory) CreateBlade(ctx context.Context, zone string, rack string, blade int64, u *pb.DefinitionBlade, options ...InventoryOption) (int64, error) {return InvalidRev, nil}


// ReadZone returns
//
func (m *DBInventory) ReadZone(ctx context.Context, zone string, options ...InventoryOption) (*DefinitionZone, int64, error) {return nil, InvalidRev, nil}

// ReadRack returns
//
func (m *DBInventory) ReadRack(ctx context.Context, zone string, rack string, blade int64, options ...InventoryOption) (*DefinitionRack, int64, error) {return nil, InvalidRev, nil}

// ReadPdus returns
//
func (m *DBInventory) ReadPdus(ctx context.Context, zone string, rack string, pdu int64, options ...InventoryOption) (*DefinitionPdu, int64, error) {return nil, InvalidRev, nil}

// ReadTor returns
//
func (m *DBInventory) ReadTor(ctx context.Context, zone string, rack string, tor int64, options ...InventoryOption) (*DefinitionTor, int64, error) {return nil, InvalidRev, nil}

// ReadBlade returns
//
func (m *DBInventory) ReadBlade(ctx context.Context, zone string, rack string, blade int64, options ...InventoryOption) (*DefinitionBlade, int64, error) {return nil, InvalidRev, nil}


// UpdateZone is used
//
func (m *DBInventory) UpdateZone(ctx context.Context, zone string, u *pb.DefinitionZone, options ...InventoryOption) (int64, error) {return InvalidRev, nil}

// UpdateRack is used
//
func (m *DBInventory) UpdateRack(ctx context.Context, zone string, rack string, u *pb.DefinitionRack, options ...InventoryOption) (int64, error) {return InvalidRev, nil}

// UpdatePdu is used
//
func (m *DBInventory) UpdatePdu(ctx context.Context, zone string, rack string, pdu int64, u *pb.DefinitionPdu, options ...InventoryOption) (int64, error) {return InvalidRev, nil}

// UpdateTor is used
//
func (m *DBInventory) UpdateTor(ctx context.Context, zone string, rack string, tor int64, u *pb.DefinitionTor, options ...InventoryOption) (int64, error) {return InvalidRev, nil}

// UpdateBlade is used
//
func (m *DBInventory) UpdateBlade(ctx context.Context, zone string, rack string, blade int64, u *pb.DefinitionBlade, options ...InventoryOption) (int64, error) {return InvalidRev, nil}


// DeleteZone is used
//
func (m *DBInventory) DeleteZone(ctx context.Context, zone string, options ...InventoryOption) (int64, error) {return InvalidRev, nil}

// DeleteRack is used
//
func (m *DBInventory) DeleteRack(ctx context.Context, zone string, rack string, options ...InventoryOption) (int64, error) {return InvalidRev, nil}

// DeletePdu is used
//
func (m *DBInventory) DeletePdu(ctx context.Context, zone string, rack string, pdu int64, options ...InventoryOption) (int64, error) {return InvalidRev, nil}

// DeleteTor is used
//
func (m *DBInventory) DeleteTor(ctx context.Context, zone string, rack string, tor int64, options ...InventoryOption) (int64, error) {return InvalidRev, nil}

// DeleteBlade is used
//
func (m *DBInventory) DeleteBlade(ctx context.Context, zone string, rack string, blade int64, options ...InventoryOption) (int64, error) {return InvalidRev, nil}

