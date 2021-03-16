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
	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/limits"
	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/namespace"
	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/store"
	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/timestamp"
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

	cfg                *config.GlobalConfig
	inventory          *inventory.Inventory
	RootSummary        *inventory.RootSummary
	DefaultZoneSummary *inventory.ZoneSummary
	actualLoaded       bool
	Actual             *actualZone
}

var dbInventory *DBInventory


// InitDBInventory initializes the base state for the inventory.
//
func InitDBInventory(ctx context.Context, cfg *config.GlobalConfig, store *store.Store) (err error) {
	if dbInventory == nil {
		db := NewDbInventory(cfg, store)

		if err := db.Start(ctx); err != nil {
			return tracing.Error(ctx, "unable to start inventory: error: %v", err)
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
func NewDbInventory(cfg *config.GlobalConfig, store *store.Store) *DBInventory {
	return  &DBInventory{
		mutex:              sync.RWMutex{},
		Started:            false,
		cfg:                cfg,
		inventory:          inventory.NewInventory(cfg, store),
		actualLoaded:       false,
		Actual:             &actualZone{Racks: make(map[string]*actualRack)},
	}
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

	if !m.Started {
		if err := m.inventory.Start(ctx); err != nil {
			return err
		}

		if err := m.LoadInventoryActual(ctx, true); err != nil {
			return err
		}

		m.DefaultZoneSummary = m.GetMemoData()
		m.Started = true
	}

	return nil
}

// GetMemoData returns the maximum number of blades held in any rack
// in the inventory.
func (m *DBInventory) GetMemoData() *inventory.ZoneSummary {
	return m.inventory.GetDefaultZoneSummary()
}

// ScanRegions scans the set of known regions in the store, invoking the supplied
// function with each entry.
//
func (m *DBInventory) ScanRegions(ctx context.Context, action func(entry string) error) error {
	root, err := m.inventory.NewRoot(namespace.DefinitionTable)
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
func (m *DBInventory) GetRegion(ctx context.Context, regionName string) (*pb.Definition_Region, error) {
	region, err := m.inventory.NewRegion(namespace.DefinitionTable, regionName)
	if err != nil {
		return nil, m.transformError(err, regionName, "", "", 0)
	}

	_, err = region.Read(ctx)
	if err != nil {
		return nil, m.transformError(err, regionName, "", "", 0)
	}

	r := &pb.Definition_Region{
		Details: region.GetDetails(),
		Zones: make(map[string]*pb.Definition_Zone),
	}

	return r, nil
}

// ScanZonesInRegion scans the set of known zones in the store within
// the specified region, invoking the supplied function with each entry.
//
func (m *DBInventory) ScanZonesInRegion(ctx context.Context, regionName string, action func(entry string) error) error {
	root, err := m.inventory.NewRoot(namespace.DefinitionTable)
	if err != nil {
		return m.transformError(err, regionName, "", "", 0)
	}

	region, err := root.NewChild(regionName)
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
func (m *DBInventory) GetZone(ctx context.Context, regionName string, zoneName string) (*pb.Definition_Zone, error) {
	zone, err := m.inventory.NewZone(namespace.DefinitionTable, regionName, zoneName)
	if err != nil {
		return nil, m.transformError(err, regionName, zoneName, "", 0)
	}

	_, err = zone.Read(ctx)
	if err != nil {
		return nil, m.transformError(err, regionName, zoneName, "", 0)
	}

	z := &pb.Definition_Zone{
		Details: zone.GetDetails(),
		Racks: make(map[string]*pb.Definition_Rack),
	}

	return z, nil
}


// ScanRacksInZone scans the set of known racks in the store, invoking the supplied
// function with each entry.
//
func (m *DBInventory) ScanRacksInZone(
	ctx context.Context,
	regionName string,
	zoneName string,
	action func(entry string) error) error {
	zone, err := m.inventory.NewZone(namespace.DefinitionTable, regionName, zoneName)

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
func (m *DBInventory) GetRackInZone(ctx context.Context, regionName string, zoneName string, rackName string) (*pb.Definition_Rack, error) {
	rack, err := m.inventory.NewRack(namespace.DefinitionTable, regionName, zoneName, rackName)
	if err != nil {
		return nil, m.transformError(err, regionName, zoneName, rackName, 0)
	}

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
		Details: rack.GetDetails(),
		Pdus:    make(map[int64]*pb.Definition_Pdu, len(*pdus)),
		Tors:    make(map[int64]*pb.Definition_Tor, len(*tors)),
		Blades:  make(map[int64]*pb.Definition_Blade, len(*blades)),
	}

	for index, pdu := range *pdus {
		r.Pdus[index] = pdu.GetDefinitionPdu()
	}

	for index, tor := range *tors {
		r.Tors[index] = tor.GetDefinitionTor()
	}

	for index, blade := range *blades {
		r.Blades[index] = blade.GetDefinitionBlade()
	}

	return r, nil
}

// ScanBladesInRack enumerates over all the blades in a rack of the given
// rackID, and invokes the supplied action on each discovered bladeID in
// turn.
func (m *DBInventory) ScanBladesInRack(ctx context.Context, regionName string, zoneName string, rackName string, action func(bladeID int64) error) error {
	rack, err := m.inventory.NewRack(
		namespace.DefinitionTable,
		regionName,
		zoneName,
		rackName)

	if err != nil {
		return m.transformError(err, regionName, zoneName, rackName, 0)
	}

	_, IDs, err := rack.ListBlades(ctx)
	if err != nil {
		return m.transformError(err, regionName, zoneName, rackName, 0)
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
func (m *DBInventory) GetBlade(ctx context.Context, regionName string, zoneName string, rackName string, bladeID int64) (*pb.BladeCapacity, error) {
	blade, err := m.inventory.NewBlade(
		namespace.DefinitionTable,
		regionName,
		zoneName,
		rackName,
		bladeID)

	if err != nil {
		return nil, m.transformError(err, regionName, zoneName, rackName, bladeID)
	}

	_, err = blade.Read(ctx)
	if err != nil {
		return nil, m.transformError(err, regionName, zoneName, rackName, bladeID)
	}

	return blade.GetCapacity(), nil
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
		return NewErrPduNotFound(rack, index)

	case err == errors.ErrPduIDInvalid{Value: index, Limit: limits.MaxPduID}:
		return NewErrPduNotFound(rack, index)

	case err == errors.ErrPduIndexInvalid{Region: region, Zone: zone, Rack: rack, Pdu: fmt.Sprintf("%d", index)}:
		return NewErrPduNotFound(rack, index)

	case err == errors.ErrTorNotFound{Region: region, Zone: zone, Rack: rack, Tor: index}:
		return NewErrTorNotFound(rack, index)

	case err == errors.ErrTorIndexNotFound{Region: region, Zone: zone, Rack: rack, Tor: index}:
		return NewErrTorNotFound(rack, index)

	case err == errors.ErrTorIDInvalid{Value: index, Limit: limits.MaxTorID}:
		return NewErrTorNotFound(rack, index)

	case err == errors.ErrTorIndexInvalid{Region: region, Zone: zone, Rack: rack, Tor: fmt.Sprintf("%d", index)}:
		return NewErrTorNotFound(rack, index)

	case err == errors.ErrBladeNotFound{Region: region, Zone: zone, Rack: rack, Blade: index}:
		return NewErrBladeNotFound(rack, index)

	case err == errors.ErrBladeIndexNotFound{Region: region, Zone: zone, Rack: rack, Blade: index}:
		return NewErrBladeNotFound(rack, index)

	case err == errors.ErrBladeIDInvalid{Value: index, Limit: limits.MaxBladeID}:
		return NewErrBladeNotFound(rack, index)

	case err == errors.ErrBladeIndexInvalid{Region: region, Zone: zone, Rack: rack, Blade: fmt.Sprintf("%d", index)}:
		return NewErrBladeNotFound(rack, index)

	default:
		return err
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

// // DefinitionTor describes the revision and value for the request TOR
// //
// type DefinitionTor struct {
// 	revision int64
// 	tor      *pb.Definition_Tor
// }

// // DefinitionPdu describes the revision and value for the request PDU
// //
// type DefinitionPdu struct {
// 	revision int64
// 	pdu      *pb.Definition_Pdu
// }

// // DefinitionBlade describes the revision and value for the request blade
// //
// type DefinitionBlade struct {
// 	revision int64
// 	blade    *pb.Definition_Blade
// }

// // DefinitionRack describes the revision and value for the request rack
// //
// type DefinitionRack struct {
// 	revision int64
// 	rack     *pb.Definition_Rack
// }

// // DefinitionZone defines the set of values used to summarize the contents of a zone to
// // allow a query on the zone without incorporating the entire inventory definition
// //
// type DefinitionZone struct {
// 	revision int64
// 	zone     *pb.Definition_Zone
// }

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
func WithRevision(rev int64) InventoryOption {
	return func(options *InventoryOptions) { options.revision = rev }
}

// Region options

// InventoryRegionOption is a
//
type InventoryRegionOption func(*InventoryRegionOptions)

// InventoryRegionOptions is a struct
//
type InventoryRegionOptions struct {
	revision      int64
	includeZones  bool
	includeRacks  bool
	includePdus   bool
	includeTors   bool
	includeBlades bool
}

func (options *InventoryRegionOptions) applyOpts(optionsArray []InventoryRegionOption) {
	for _, option := range optionsArray {
		option(options)
	}
}

// WithRegionRevision is a
//
func WithRegionRevision(rev int64) InventoryRegionOption {
	return func(options *InventoryRegionOptions) { options.revision = rev }
}

// WithRegionZones is a
//
func WithRegionZones() InventoryRegionOption {
	return func(options *InventoryRegionOptions) { options.includeZones = true }
}

// WithRegionRacks is a
//
func WithRegionRacks() InventoryRegionOption {
	return func(options *InventoryRegionOptions) { options.includeRacks = true }
}

// WithRegionTors is a
//
func WithRegionTors() InventoryRegionOption {
	return func(options *InventoryRegionOptions) { options.includeTors = true }
}

// WithRegionPdus is a
//
func WithRegionPdus() InventoryRegionOption {
	return func(options *InventoryRegionOptions) { options.includePdus = true }
}

// WithRegionBlades is a
//
func WithRegionBlades() InventoryRegionOption {
	return func(options *InventoryRegionOptions) { options.includeBlades = true }
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

func (options *InventoryZoneOptions) applyOpts(optionsArray []InventoryZoneOption) {
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

func (options *InventoryRackOptions) applyOpts(optionsArray []InventoryRackOption) {
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

// // ListZones returns the basic zone record for all the discovered zones. Optionally,
// // racks along with the rack component PDU, TOR and blades can also be returned.
// //
// func (m *DBInventory) ListZones(
// 	ctx context.Context,
// 	options ...InventoryZoneOption) (map[string]*DefinitionZone, int64, error) {
// 	return nil, InvalidRev, nil
// }

// // ListRacks returns the basic zone record for all the discovered racks within
// // the specified zone. Optionally, the rack component PDU, TOR and blades can
// // also be returned.
// //
// func (m *DBInventory) ListRacks(
// 	ctx context.Context,
// 	zone string,
// 	options ...InventoryRackOption) (map[string]*DefinitionRack, int64, error) {
// 	return nil, InvalidRev, nil
// }

// // ListPdus returns the basic records for the PDUs in the specified rack.
// //
// func (m *DBInventory) ListPdus(
// 	ctx context.Context,
// 	zone string,
// 	rack string,
// 	options ...InventoryOption) (map[string]*DefinitionPdu, int64, error) {
// 	return nil, InvalidRev, nil
// }

// // ListTors returns the basic records for the TORs in the specified rack.
// //
// func (m *DBInventory) ListTors(
// 	ctx context.Context,
// 	zone string,
// 	rack string,
// 	options ...InventoryOption) (map[string]*DefinitionTor, int64, error) {
// 	return nil, InvalidRev, nil
// }

// // ListBlades returns the basic records for the blades in the specified rack.
// //
// func (m *DBInventory) ListBlades(
// 	ctx context.Context,
// 	zone string,
// 	rack string,
// 	options ...InventoryOption) (map[string]*DefinitionBlade, int64, error) {
// 	return nil, InvalidRev, nil
// }

// CreateRegion is used to create a basic region record in the store.
//
// This record created will contain just the region level details
// and any additional zone, rack, blade, tor or pdu data will be ignored
// and not included in the stored record.
//
func (m *DBInventory) CreateRegion(
	ctx context.Context,
	name string,
	region *pb.Definition_Region,
	options ...InventoryRegionOption) (int64, error) {

	r, err := m.inventory.NewRegion(
		namespace.DefinitionTable,
		name)

	if err != nil {
		return InvalidRev, err
	}

	r.SetDetails(region.Details)

	rev, err := r.Create(ctx)

	if err != nil {
		return InvalidRev, err
	}

	return rev, nil
}

// CreateZone is used to create a basic zone record in the store.
//
// This record created will contain just the zone level details
// and any additional rack, blade, tor or pdu data will be ignored
// and not included in the stored record.
//
func (m *DBInventory) CreateZone(
	ctx context.Context,
	region string,
	name string,
	zone *pb.Definition_Zone,
	options ...InventoryZoneOption) (int64, error) {

	z, err := m.inventory.NewZone(
		namespace.DefinitionTable,
		region,
		name)

	if err != nil {
		return InvalidRev, err
	}

	z.SetDetails(zone.Details)

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
	region string,
	zone string,
	name string,
	rack *pb.Definition_Rack,
	options ...InventoryRackOption) (int64, error) {

	r, err := m.inventory.NewRack(
		namespace.DefinitionTable,
		region,
		zone,
		name)

	if err != nil {
		return InvalidRev, err
	}

	r.SetDetails(rack.Details)

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
	region string,
	zone string,
	rack string,
	index int64,
	pdu *pb.Definition_Pdu,
	options ...InventoryOption) (int64, error) {

	if err := pdu.Validate("", 0); err != nil {
		return InvalidRev, err
	}

	p, err := m.inventory.NewPdu(
		namespace.DefinitionTable,
		region,
		zone,
		rack,
		index,
	)

	if err != nil {
		return InvalidRev, err
	}

	p.SetDetails(pdu.Details)
	p.SetPorts(&pdu.Ports)

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
	region string,
	zone string,
	rack string,
	index int64,
	tor *pb.Definition_Tor,
	options ...InventoryOption) (int64, error) {

	if err := tor.Validate("", 0); err != nil {
		return InvalidRev, err
	}

	t, err := m.inventory.NewTor(
		namespace.DefinitionTable,
		region,
		zone,
		rack,
		index,
	)

	if err != nil {
		return InvalidRev, err
	}

	t.SetDetails(tor.Details)
	t.SetPorts(&tor.Ports)

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
	region string,
	zone string,
	rack string,
	index int64,
	blade *pb.Definition_Blade,
	options ...InventoryOption) (int64, error) {

	b, err := m.inventory.NewBlade(
		namespace.DefinitionTable,
		region,
		zone,
		rack,
		index,
	)

	if err != nil {
		return InvalidRev, err
	}

	b.SetDetails(blade.Details)
	b.SetCapacity(blade.Capacity)
	b.SetBootInfo(blade.BootInfo)
	b.SetBootPowerOn(blade.BootOnPowerOn)

	rev, err := b.Create(ctx)

	if err != nil {
		return InvalidRev, err
	}

	return rev, nil
}

// ReadRegion returns the region information with optionally additional
// zone, rack, blade, tor and pdu details for an optionally specified
// revision.
//
func (m *DBInventory) ReadRegion(
	ctx context.Context,
	name string,
	options ...InventoryRegionOption) (*pb.Definition_Region, int64, error) {

	r, err := m.inventory.NewRegion(
		namespace.DefinitionTable,
		name)

	if err != nil {
		return nil, InvalidRev, err
	}

	rev, err := r.Read(ctx)

	if err != nil {
		return nil, InvalidRev, err
	}

	return r.GetDefinitionRegion(), rev, nil
}

// ReadZone returns the zone information with optionally additional
// rack, blade, tor and pdu details for an optionally specified
// revision.
//
func (m *DBInventory) ReadZone(
	ctx context.Context,
	region string,
	name string,
	options ...InventoryZoneOption) (*pb.Definition_Zone, int64, error) {

	z, err := m.inventory.NewZone(
		namespace.DefinitionTable,
		region,
		name)

	if err != nil {
		return nil, InvalidRev, err
	}

	rev, err := z.Read(ctx)

	if err != nil {
		return nil, InvalidRev, err
	}

	return z.GetDefinitionZone(), rev, nil
}

// ReadRack returns the rack information with optionally additional
// blade, tor and pdu details for an optionally specified revision.
//
func (m *DBInventory) ReadRack(
	ctx context.Context,
	region string,
	zone string,
	name string,
	options ...InventoryRackOption) (*pb.Definition_Rack, int64, error) {

	r, err := m.inventory.NewRack(
		namespace.DefinitionTable,
		region,
		zone,
		name)

	if err != nil {
		return nil, InvalidRev, err
	}

	rev, err := r.Read(ctx)

	if err != nil {
		return nil, InvalidRev, err
	}

	return r.GetDefinitionRack(), rev, nil
}

// ReadPdu returns the PDU information for an optionally specified revision.
//
func (m *DBInventory) ReadPdu(
	ctx context.Context,
	region string,
	zone string,
	rack string,
	index int64,
	options ...InventoryOption) (*pb.Definition_Pdu, int64, error) {

	p, err := m.inventory.NewPdu(
		namespace.DefinitionTable,
		region,
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

	return p.GetDefinitionPdu(), rev, nil
}

// ReadTor returns the TOR information for an optionally specified revision.
//
func (m *DBInventory) ReadTor(
	ctx context.Context,
	region string,
	zone string,
	rack string,
	index int64,
	options ...InventoryOption) (*pb.Definition_Tor, int64, error) {

	t, err := m.inventory.NewTor(
		namespace.DefinitionTable,
		region,
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

	return t.GetDefinitionTor(), rev, nil
}

// ReadBlade returns the blade information for an optionally specified revision.
//
func (m *DBInventory) ReadBlade(
	ctx context.Context,
	region string,
	zone string,
	rack string,
	index int64,
	options ...InventoryOption) (*pb.Definition_Blade, int64, error) {

	b, err := m.inventory.NewBlade(
		namespace.DefinitionTable,
		region,
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

	return b.GetDefinitionBlade(), rev, nil
}

// UpdateRegion is used to update the Region basic details record.
//
// Only the Region level details will be updated and any
// additional rack, blade, tor or pdu data will be ignored
// and not included in the updated record.
//
func (m *DBInventory) UpdateRegion(
	ctx context.Context,
	name string,
	region *pb.Definition_Region,
	options ...InventoryRegionOption) (int64, error) {

	r, err := m.inventory.NewRegion(
		namespace.DefinitionTable,
		name)

	if err != nil {
		return InvalidRev, err
	}

	r.SetDetails(region.Details)

	rev, err := r.Update(ctx, true)

	if err != nil {
		return InvalidRev, err
	}

	return rev, nil
}

// UpdateZone is used to update the zone basic details record.
//
// Only the zone level details will be updated and any
// additional rack, blade, tor or pdu data will be ignored
// and not included in the updated record.
//
func (m *DBInventory) UpdateZone(
	ctx context.Context,
	region string,
	name string,
	zone *pb.Definition_Zone,
	options ...InventoryZoneOption) (int64, error) {

	z, err := m.inventory.NewZone(
		namespace.DefinitionTable,
		region,
		name)

	if err != nil {
		return InvalidRev, err
	}

	z.SetDetails(zone.Details)

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
	region string,
	zone string,
	name string,
	rack *pb.Definition_Rack, options ...InventoryRackOption) (int64, error) {

	r, err := m.inventory.NewRack(
		namespace.DefinitionTable,
		region,
		zone,
		name)

	if err != nil {
		return InvalidRev, err
	}

	r.SetDetails(rack.Details)

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
	region string,
	zone string,
	rack string,
	index int64,
	pdu *pb.Definition_Pdu,
	options ...InventoryOption) (int64, error) {

	p, err := m.inventory.NewPdu(
		namespace.DefinitionTable,
		region,
		zone,
		rack,
		index,
	)

	if err != nil {
		return InvalidRev, err
	}

	p.SetDetails(pdu.Details)
	p.SetPorts(&pdu.Ports)

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
	region string,
	zone string,
	rack string,
	index int64,
	tor *pb.Definition_Tor,
	options ...InventoryOption) (int64, error) {

	t, err := m.inventory.NewTor(
		namespace.DefinitionTable,
		region,
		zone,
		rack,
		index,
	)

	if err != nil {
		return InvalidRev, err
	}

	t.SetDetails(tor.Details)
	t.SetPorts(&tor.Ports)

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
	region string,
	zone string,
	rack string,
	index int64,
	blade *pb.Definition_Blade,
	options ...InventoryOption) (int64, error) {

	b, err := m.inventory.NewBlade(
		namespace.DefinitionTable,
		region,
		zone,
		rack,
		index,
	)

	if err != nil {
		return InvalidRev, err
	}

	b.SetDetails(blade.Details)
	b.SetCapacity(blade.Capacity)
	b.SetBootInfo(blade.BootInfo)
	b.SetBootPowerOn(blade.BootOnPowerOn)

	rev, err := b.Update(ctx, true)

	if err != nil {
		return InvalidRev, err
	}

	return rev, nil
}

// DeleteRegion is used to delete the Region record and any
// contained rack records. That is it will delete the
// entire Region and all related records.
//
func (m *DBInventory) DeleteRegion(
	ctx context.Context,
	name string,
	options ...InventoryOption) (int64, error) {

	r, err := m.inventory.NewRegion(
		namespace.DefinitionTable,
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

// DeleteZone is used to delete the zone record and any
// contained rack records. That is it will delete the
// entire zone and all related records.
//
func (m *DBInventory) DeleteZone(
	ctx context.Context,
	region string,
	name string,
	options ...InventoryOption) (int64, error) {

	z, err := m.inventory.NewZone(
		namespace.DefinitionTable,
		region,
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
	region string,
	zone string,
	name string,
	options ...InventoryOption) (int64, error) {

	r, err := m.inventory.NewRack(
		namespace.DefinitionTable,
		region,
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
	region string,
	zone string,
	rack string,
	index int64,
	options ...InventoryOption) (int64, error) {

	p, err := m.inventory.NewPdu(
		namespace.DefinitionTable,
		region,
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
	region string,
	zone string,
	rack string,
	index int64,
	options ...InventoryOption) (int64, error) {

	t, err := m.inventory.NewTor(
		namespace.DefinitionTable,
		region,
		zone,
		rack,
		index,
	)

	if err != nil {
		return InvalidRev, err
	}

	rev, err := t.Delete(ctx, true)

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
	region string,
	zone string,
	rack string,
	index int64,
	options ...InventoryOption) (int64, error) {

	b, err := m.inventory.NewBlade(
		namespace.DefinitionTable,
		region,
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
