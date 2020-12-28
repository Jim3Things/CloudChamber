package inventory

import (
	"context"
	"strconv"

	"github.com/Jim3Things/CloudChamber/internal/clients/store"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/inventory"
)

// NewRoot returns a root struct which can be used to navigate
// the namespace for a given table
//
// Valid tables are
//	- DefinitionTable
//	- ActualTable
//	- ObservedTable
//	- TargetTable
//
func NewRoot(ctx context.Context, store *store.Store, table string) (*Root, error) {

	k, err := GetKeyForIndexRegion(table)

	if err != nil {
		return nil, err
	}

	r := &Root{
		Store:         store,
		KeyChildIndex: k,
	}

	return r, nil
}

// NewRegion is a convenience function used to construct a Region struct
// from scratch rather than relative to its parent.
//
func NewRegion(ctx context.Context, store *store.Store, table string, region string) (*Region, error) {

	keyIndex, err := GetKeyForIndexZone(table, region)

	if err != nil {
		return nil, err
	}

	key, err := GetKeyForRegion(DefinitionTable, region)

	if nil != err {
		return nil, err
	}

	r := &Region{
		Store:         store,
		KeyChildIndex: keyIndex,
		Key:           key,
		Region:        region,
	}

	return r, nil
}

// NewZone is a convenience function used to construct a Zone struct
// from scratch rather than relative to its parent.
//
func NewZone(ctx context.Context, store *store.Store, table string, region string, zone string) (*Zone, error) {

	keyIndex, err := GetKeyForIndexRack(table, region, zone)

	if err != nil {
		return nil, err
	}

	key, err := GetKeyForZone(DefinitionTable, region, zone)

	if nil != err {
		return nil, err
	}

	z := &Zone{
		Store:         store,
		KeyChildIndex: keyIndex,
		Key:           key,
		Region:        region,
		Zone:          zone,
	}

	return z, nil
}

// NewRack is a convenience function used to construct a Rack struct
// from scratch rather than relative to its parent.
//
func NewRack(ctx context.Context, store *store.Store, table string, region string, zone string, rack string) (*Rack, error) {

	keyIndexPdu, err := GetKeyForIndexPdu(table, region, zone, rack)

	if err != nil {
		return nil, err
	}

	keyIndexTor, err := GetKeyForIndexTor(table, region, zone, rack)

	if err != nil {
		return nil, err
	}

	keyIndexBlade, err := GetKeyForIndexBlade(table, region, zone, rack)

	if err != nil {
		return nil, err
	}

	key, err := GetKeyForRack(DefinitionTable, region, zone, rack)

	if nil != err {
		return nil, err
	}

	r := &Rack{
		Store:         store,
		KeyIndexPdu:   keyIndexPdu,
		KeyIndexTor:   keyIndexTor,
		KeyIndexBlade: keyIndexBlade,
		Key:           key,
		Region:        region,
		Zone:          zone,
		Rack:          rack,
	}

	return r, nil
}

// NewPdu is a convenience function used to construct a Pdu struct
// from scratch rather than relative to its parent.
//
func NewPdu(ctx context.Context, store *store.Store, table string, region string, zone string, rack string, id int64) (*Pdu, error) {

	key, err := GetKeyForPdu(DefinitionTable, region, zone, rack, id)

	if nil != err {
		return nil, err
	}

	p := &Pdu{
		Store:  store,
		Key:    key,
		Region: region,
		Zone:   zone,
		Rack:   rack,
		ID:     id,
	}

	return p, nil
}

// NewTor is a convenience function used to construct a Pdu struct
// from scratch rather than relative to its parent.
//
func NewTor(ctx context.Context, store *store.Store, table string, region string, zone string, rack string, id int64) (*Tor, error) {

	key, err := GetKeyForTor(DefinitionTable, region, zone, rack, id)

	if nil != err {
		return nil, err
	}

	t := &Tor{
		Store:  store,
		Key:    key,
		Region: region,
		Zone:   zone,
		Rack:   rack,
		ID:     id,
	}

	return t, nil
}

// NewBlade is a convenience function used to construct a Pdu struct
// from scratch rather than relative to its parent.
//
func NewBlade(ctx context.Context, store *store.Store, table string, region string, zone string, rack string, id int64) (*Blade, error) {

	key, err := GetKeyForBlade(DefinitionTable, region, zone, rack, id)

	if nil != err {
		return nil, err
	}

	b := &Blade{
		Store:  store,
		Key:    key,
		Region: region,
		Zone:   zone,
		Rack:   rack,
		ID:     id,
	}

	return b, nil
}




// Root is a 
//
type Root struct {

	Store        *store.Store
	KeyChildIndex string

	details      *pb.RootDetails
}


// SetName is a
//
func (r *Root) SetName(ctx context.Context, name string) error {
	
	keyIndex , err := GetKeyForIndexZone(DefinitionTable, name)

	if err != nil {
		return err
	}

	r.KeyChildIndex = keyIndex

	return nil
}

// SetDetails is a
//
func (r *Root) SetDetails(ctx context.Context, details *pb.RootDetails) error {

	r.details = details

	return 	nil
}

// GetDetails is a
//
func (r *Root) GetDetails(ctx context.Context) (*pb.RootDetails, error) {
	if r.details == nil {
		return nil, ErrDetailsNotAvailable("root")
	}
	
	return 	r.details, nil
}

// Create is a
//
func (r *Root) Create(ctx context.Context) (int64, error) {
	return store.RevisionInvalid, ErrFunctionNotAvailable
}

// Read is a
//
func (r *Root) Read(ctx context.Context) (*interface{}, error) {
	return 	nil, ErrFunctionNotAvailable
}

// Update is a
//
func (r *Root) Update(ctx context.Context) (int64, error) {
	return store.RevisionInvalid, ErrFunctionNotAvailable
}

// Delete is a
//
func (r *Root) Delete(ctx context.Context) error {
	return 	ErrFunctionNotAvailable
}

// NewChild is a
//
func (r *Root) NewChild(ctx context.Context, name string) (*Region, error) {

	region, err := NewRegion(ctx, r.Store, DefinitionTable, name)

	if err != nil {
		return nil, err
	}

	return region, nil
}

// ListChildren is a
//
func (r *Root) ListChildren(ctx context.Context) (*map[string]Region, error) {

	records, _, err := r.Store.List(ctx, store.KeyRootInventory, r.KeyChildIndex)

	if err == store.ErrStoreKeyNotFound(r.KeyChildIndex) {
		return nil, ErrRootNotFound{DefinitionTable}
	}

	if err != nil {
		return nil, err
	}

	regions := make(map[string]Region, len(*records))

	for k, v := range *records {

		record := &pb.StoreRecordDefinitionRegion{}

		if err = store.Decode(v.Value, record); err != nil {
			return nil, err
		}

		region, err := r.NewChild(ctx, k)

		if err != nil {
			return nil, err
		}

		region.revision = v.Revision
		region.record   = record
		region.details  = record.Details

		regions[k] = *region
	}

	return &regions, nil
}






// Region is a
//
type Region struct {

	Store         *store.Store
	KeyChildIndex string
	Key           string
	Region        string

	revision int64
	record   *pb.StoreRecordDefinitionRegion
	details  *pb.RegionDetails
}

// SetName is a 
//
func (r *Region) SetName(ctx context.Context, name string) error {

	key, err := GetKeyForRegion(DefinitionTable, name)

	if nil != err {
		return err
	}

	r.Key = key

	return nil
}

// SetDetails is a
//
func (r *Region) SetDetails(ctx context.Context, details *pb.RegionDetails) error {

	r.details = details

	return 	nil
}

// GetDetails is a
//
func (r *Region) GetDetails(ctx context.Context) (*pb.RegionDetails, error) {
	if r.details == nil {
		return nil, ErrDetailsNotAvailable("region")
	}
	
	return 	r.details, nil
}

// Create is
//
func (r *Region) Create(ctx context.Context) (int64, error) {

	if r.details == nil {
		return store.RevisionInvalid, ErrDetailsNotAvailable("region")
	}

	record := &pb.StoreRecordDefinitionRegion{
		Details: r.details,
	}

	v, err := store.Encode(record)

	if err != nil {
		return store.RevisionInvalid, err
	}

	rev, err := r.Store.Create(ctx, store.KeyRootInventory, r.Key, v)

	if err == store.ErrStoreAlreadyExists(r.Key) {
		return store.RevisionInvalid, ErrZoneAlreadyExists(r.Region)
	}

	r.revision = rev

	return r.revision, nil
}

// Read is
//
func (r *Region) Read(ctx context.Context) (*pb.RegionDetails, error) {

	return nil, nil
}

// Update is
//
func (r *Region) Update(ctx context.Context) (int64, error) {

	return store.RevisionInvalid, nil
}

// Delete is
//
func (r *Region) Delete(ctx context.Context) error {

	return nil
}

// NewChild is a 
//
func (r *Region) NewChild(ctx context.Context, name string) (*Zone, error) {

	z, err := NewZone(ctx, r.Store, DefinitionTable, r.Region, name)

	if err != nil {
		return nil, err
	}

	return z, err
}


// ListChildren is a
//
func (r *Region) ListChildren(ctx context.Context) (*map[string]Zone, error) {

	records, _, err := r.Store.List(ctx, store.KeyRootInventory, r.KeyChildIndex)

	if err == store.ErrStoreIndexNotFound(r.KeyChildIndex) {
		return nil, ErrIndexNotFound(r.KeyChildIndex)
	}

	if err != nil {
		return nil, err
	}

	zones := make(map[string]Zone, len(*records))

	for k, v := range *records {

		record := &pb.StoreRecordDefinitionZone{}

		if err = store.Decode(v.Value, record); err != nil {
			return nil, err
		}

		zone, err := r.NewChild(ctx, k)

		if err != nil {
			return nil, err
		}
	
		zone.revision = v.Revision
		zone.record   = record
		zone.details  = record.Details
		
		zones[k] = *zone
	}

	return &zones, nil
}





// Zone is a
//
type Zone struct {

	Store         *store.Store
	KeyChildIndex string
	Key           string
	Region        string
	Zone          string

	revision int64
	record   *pb.StoreRecordDefinitionZone
	details  *pb.ZoneDetails
}

// SetName is a 
//
func (z *Zone) SetName(ctx context.Context, name string) error {

	key, err := GetKeyForZone(DefinitionTable, z.Region, name)

	if nil != err {
		return err
	}

	z.Key = key

	return nil
}

// SetDetails is a
//
func (z *Zone) SetDetails(ctx context.Context, details *pb.ZoneDetails) error {

	z.details = details

	return 	nil
}

// GetDetails is a
//
func (z *Zone) GetDetails(ctx context.Context) (*pb.ZoneDetails, error) {
	if z.details == nil {
		return nil, ErrDetailsNotAvailable("zone")
	}
	
	return 	z.details, nil
}

// Create is
//
func (z *Zone) Create(ctx context.Context) (int64, error) {

	if z.details == nil {
		return store.RevisionInvalid, ErrDetailsNotAvailable("zone")
	}

	record := &pb.StoreRecordDefinitionZone{
		Details: z.details,
	}

	v, err := store.Encode(record)

	if err != nil {
		return store.RevisionInvalid, err
	}

	rev, err := z.Store.Create(ctx, store.KeyRootInventory, z.Key, v)

	if err == store.ErrStoreAlreadyExists(z.Key) {
		return store.RevisionInvalid, ErrZoneAlreadyExists(z.Zone)
	}

	z.revision = rev

	return z.revision, nil
}

// Read is
//
func (z *Zone) Read(ctx context.Context) (*pb.ZoneDetails, error) {

	return nil, nil
}

// Update is
//
func (z *Zone) Update(ctx context.Context) error {

	return nil
}

// Delete is
//
func (z *Zone) Delete(ctx context.Context) error {

	return nil
}

// NewChild is a 
//
func (z *Zone) NewChild(ctx context.Context, name string) (*Rack, error) {

	r, err := NewRack(ctx, z.Store, DefinitionTable, z.Region, z.Zone, name)

	if err != nil {
		return nil, err
	}

	return r, err
}


// ListChildren is a
//
func (z *Zone) ListChildren(ctx context.Context) (*map[string]Rack, error) {

	records, _, err := z.Store.List(ctx, store.KeyRootInventory, z.KeyChildIndex)

	if err == store.ErrStoreIndexNotFound(z.KeyChildIndex) {
		return nil, ErrIndexNotFound(z.KeyChildIndex)
	}

	if err != nil {
		return nil, err
	}

	racks := make(map[string]Rack, len(*records))

	for k, v := range *records {

		record := &pb.StoreRecordDefinitionRack{}

		if err = store.Decode(v.Value, record); err != nil {
			return nil, err
		}

		rack, err := z.NewChild(ctx, k)

		if err != nil {
			return nil, err
		}
	
		rack.revision = v.Revision
		rack.record   = record
		rack.details  = record.Details
		
		racks[k] = *rack
	}

	return &racks, nil
}




// Rack is a
//
type Rack struct {

	Store         *store.Store
	KeyIndexPdu   string
	KeyIndexTor   string
	KeyIndexBlade string
	Key           string
	Region        string
	Zone          string
	Rack          string

	revision int64
	record   *pb.StoreRecordDefinitionRack
	details  *pb.RackDetails

}

// SetName is a 
//
func (r *Rack) SetName(ctx context.Context, name string) error {

	key, err := GetKeyForRegion(DefinitionTable, name)

	if nil != err {
		return err
	}

	r.Key = key

	return nil
}

// SetDetails is a
//
func (r *Rack) SetDetails(ctx context.Context, details *pb.RackDetails) error {

	r.details = details

	return 	nil
}

// GetDetails is a
//
func (r *Rack) GetDetails(ctx context.Context) (*pb.RackDetails, error) {
	if r.details == nil {
		return nil, ErrDetailsNotAvailable("rack")
	}
	
	return 	r.details, nil
}

// Create is
//
func (r *Rack) Create(ctx context.Context) (int64, error) {

	if r.details == nil {
		return store.RevisionInvalid, ErrDetailsNotAvailable("rack")
	}

	record := &pb.StoreRecordDefinitionRack{
		Details: r.details,
	}

	v, err := store.Encode(record)

	if err != nil {
		return store.RevisionInvalid, err
	}

	rev, err := r.Store.Create(ctx, store.KeyRootInventory, r.Key, v)

	if err == store.ErrStoreAlreadyExists(r.Key) {
		return store.RevisionInvalid, ErrZoneAlreadyExists(r.Region)
	}

	r.revision = rev

	return r.revision, nil
}

// Read is
//
func (r *Rack) Read(ctx context.Context) (*pb.RackDetails, error) {

	return nil, nil
}

// Update is
//
func (r *Rack) Update(ctx context.Context) (int64, error) {

	return store.RevisionInvalid, nil
}

// Delete is
//
func (r *Rack) Delete(ctx context.Context) error {

	return nil
}

// NewChild is a 
//
func (r *Rack) NewChild(ctx context.Context, name string) (*Zone, error) {
	return nil, ErrFunctionNotAvailable
}

// NewPdu is a 
//
func (r *Rack) NewPdu(ctx context.Context, ID int64) (*Pdu, error) {

	p, err := NewPdu(ctx, r.Store, DefinitionTable, r.Region, r.Zone, r.Rack, ID)

	if err != nil {
		return nil, err
	}

	return p, err
}

// NewTor is a 
//
func (r *Rack) NewTor(ctx context.Context, ID int64) (*Tor, error) {

	t, err := NewTor(ctx, r.Store, DefinitionTable, r.Region, r.Zone, r.Rack, ID)

	if err != nil {
		return nil, err
	}

	return t, err
}

// NewBlade is a 
//
func (r *Rack) NewBlade(ctx context.Context, ID int64) (*Blade, error) {

	b, err := NewBlade(ctx, r.Store, DefinitionTable, r.Region, r.Zone, r.Rack, ID)

	if err != nil {
		return nil, err
	}

	return b, err
}

// ListChildren is a
//
func (r *Rack) ListChildren(ctx context.Context) (*map[string]Zone, error) {
	return nil, ErrFunctionNotAvailable
}


// ListPdus is a
//
func (r *Rack) ListPdus(ctx context.Context) (*map[int64]Pdu, error) {

	records, _, err := r.Store.List(ctx, store.KeyRootInventory, r.KeyIndexPdu)

	if err == store.ErrStoreIndexNotFound(r.KeyIndexPdu) {
		return nil, ErrIndexNotFound(r.KeyIndexPdu)
	}

	if err != nil {
		return nil, err
	}

	pdus := make(map[int64]Pdu, len(*records))

	for k, v := range *records {

		i, err := strconv.ParseInt(k, 10, 0)

		if err != nil {
			return nil, err
		}

		record := &pb.StoreRecordDefinitionPdu{}

		if err = store.Decode(v.Value, record); err != nil {
			return nil, err
		}

		pdu, err := r.NewPdu(ctx, i)

		if err != nil {
			return nil, err
		}
	
		pdu.revision = v.Revision
		pdu.record   = record
		pdu.details  = record.Details
		pdu.ports    = &record.Ports
		
		pdus[i] = *pdu
	}

	return &pdus, nil
}

// ListTors is a
//
func (r *Rack) ListTors(ctx context.Context) (*map[int64]Tor, error) {

	records, _, err := r.Store.List(ctx, store.KeyRootInventory, r.KeyIndexTor)

	if err == store.ErrStoreIndexNotFound(r.KeyIndexTor) {
		return nil, ErrIndexNotFound(r.KeyIndexTor)
	}

	if err != nil {
		return nil, err
	}

	tors := make(map[int64]Tor, len(*records))

	for k, v := range *records {

		i, err := strconv.ParseInt(k, 10, 0)

		if err != nil {
			return nil, err
		}

		record := &pb.StoreRecordDefinitionTor{}

		if err = store.Decode(v.Value, record); err != nil {
			return nil, err
		}

		tor, err := r.NewTor(ctx, i)

		if err != nil {
			return nil, err
		}
	
		tor.revision = v.Revision
		tor.record   = record
		tor.details  = record.Details
		tor.ports    = &record.Ports
	
		tors[i] = *tor
	}

	return &tors, nil
}

// ListBlades is a
//
func (r *Rack) ListBlades(ctx context.Context) (*map[int64]Blade, error) {

	records, _, err := r.Store.List(ctx, store.KeyRootInventory, r.KeyIndexBlade)

	if err == store.ErrStoreIndexNotFound(r.KeyIndexBlade) {
		return nil, ErrIndexNotFound(r.KeyIndexBlade)
	}

	if err != nil {
		return nil, err
	}

	blades := make(map[int64]Blade, len(*records))

	for k, v := range *records {

		i, err := strconv.ParseInt(k, 10, 0)

		if err != nil {
			return nil, err
		}

		record := &pb.StoreRecordDefinitionBlade{}

		if err = store.Decode(v.Value, record); err != nil {
			return nil, err
		}

		blade, err := r.NewBlade(ctx, i)

		if err != nil {
			return nil, err
		}
	
		blade.revision      = v.Revision
		blade.record        = record
		blade.details       = record.Details
		blade.capacity      = record.Capacity
		blade.bootOnPowerOn = record.BootOnPowerOn
		blade.bootInfo      = record.BootInfo
		
		blades[i] = *blade
	}

	return &blades, nil
}







// Pdu is a
//
type Pdu struct {

	Store  *store.Store
	Key    string
	Region string
	Zone   string
	Rack   string
	ID     int64

	revision int64
	record   *pb.StoreRecordDefinitionPdu
	details  *pb.PduDetails
	ports    *map[int64]*pb.PowerPort
}

// SetName is a 
//
func (p *Pdu) SetName(ctx context.Context, ID int64) error {

	key, err := GetKeyForPdu(DefinitionTable, p.Region, p.Zone, p.Rack, ID)

	if nil != err {
		return err
	}

	p.Key = key

	return nil
}

// SetDetails is a
//
func (p *Pdu) SetDetails(ctx context.Context, details *pb.PduDetails) error {

	p.details = details

	return 	nil
}

// GetDetails is a
//
func (p *Pdu) GetDetails(ctx context.Context) (*pb.PduDetails, error) {

	if p.details == nil {
		return nil, ErrDetailsNotAvailable("pdu")
	}
	
	return 	p.details, nil
}

// SetPorts is a
//
func (p *Pdu) SetPorts(ctx context.Context, ports *map[int64]*pb.PowerPort) error {

	p.ports = ports

	return nil
}

// GetPorts is a 
//
func (p *Pdu) GetPorts(ctx context.Context) (*map[int64]*pb.PowerPort, error) {
	if p.ports == nil {
		return nil, ErrPortsNotAvailable("pdu")
	}

	return p.ports, nil
}

// Create is
//
func (p *Pdu) Create(ctx context.Context) (int64, error) {

	if p.details == nil {
		return store.RevisionInvalid, ErrDetailsNotAvailable("pdu")
	}

	if p.ports == nil {
		return store.RevisionInvalid, ErrPortsNotAvailable("pdu")
	}

	record := &pb.StoreRecordDefinitionPdu{
		Details: p.details,
		Ports:   *p.ports,
	}
	v, err := store.Encode(record)

	if err != nil {
		return store.RevisionInvalid, err
	}

	rev, err := p.Store.Create(ctx, store.KeyRootInventory, p.Key, v)

	if err == store.ErrStoreAlreadyExists(p.Key) {
		return store.RevisionInvalid, ErrfPduAlreadyExists(p.Zone, p.Rack, p.ID)
	}

	p.revision = rev

	return p.revision, nil
}

// Read is
//
func (p *Pdu) Read(ctx context.Context) (*pb.PduDetails, error) {

	return nil, nil
}

// Update is
//
func (p *Pdu) Update(ctx context.Context) error {

	return nil
}

// Delete is
//
func (p *Pdu) Delete(ctx context.Context) error {

	return nil
}







// Tor is a
//
type Tor struct {

	Store  *store.Store
	Key    string
	Region string
	Zone   string
	Rack   string
	ID     int64

	revision int64
	record   *pb.StoreRecordDefinitionTor
	details  *pb.TorDetails
	ports    *map[int64]*pb.NetworkPort
}

// SetName is a 
//
func (t *Tor) SetName(ctx context.Context, ID int64) error {

	key, err := GetKeyForPdu(DefinitionTable, t.Region, t.Zone, t.Rack, ID)

	if nil != err {
		return err
	}

	t.Key = key

	return nil
}

// SetDetails is a
//
func (t *Tor) SetDetails(ctx context.Context, details *pb.TorDetails) error {

	t.details = details

	return 	nil
}

// GetDetails is a
//
func (t *Tor) GetDetails(ctx context.Context) (*pb.TorDetails, error) {
	if t.details == nil {
		return nil, ErrDetailsNotAvailable("tor")
	}
	
	return 	t.details, nil
}

// SetPorts is a
//
func (t *Tor) SetPorts(ctx context.Context, ports *map[int64]*pb.NetworkPort) error {
	t.ports = ports

	return nil
}

// GetPorts is a 
//
func (t *Tor) GetPorts(ctx context.Context) (*map[int64]*pb.NetworkPort, error) {
	if t.ports == nil {
		return nil, ErrPortsNotAvailable("tor")
	}

	return t.ports, nil
}

// Create is
//
func (t *Tor) Create(ctx context.Context) (int64, error) {

	if t.details == nil {
		return store.RevisionInvalid, ErrDetailsNotAvailable("tor")
	}

	if t.ports == nil {
		return store.RevisionInvalid, ErrPortsNotAvailable("tor")
	}

	record := &pb.StoreRecordDefinitionTor{
		Details: t.details,
		Ports:   *t.ports,
	}

	v, err := store.Encode(record)

	if err != nil {
		return store.RevisionInvalid, err
	}

	rev, err := t.Store.Create(ctx, store.KeyRootInventory, t.Key, v)

	if err == store.ErrStoreAlreadyExists(t.Key) {
		return store.RevisionInvalid, ErrfPduAlreadyExists(t.Zone, t.Rack, t.ID)
	}

	t.revision = rev

	return t.revision, nil
}

// Read is
//
func (t *Tor) Read(ctx context.Context) (*pb.TorDetails, error) {

	return nil, nil
}

// Update is
//
func (t *Tor) Update(ctx context.Context) error {

	return nil
}

// Delete is
//
func (t *Tor) Delete(ctx context.Context) error {

	return nil
}







// Blade is a
//
type Blade struct {

	Store  *store.Store
	Key    string
	Region string
	Zone   string
	Rack   string
	ID     int64

	revision      int64
	record        *pb.StoreRecordDefinitionBlade
	details       *pb.BladeDetails
	capacity      *pb.BladeCapacity
	bootInfo      *pb.BladeBootInfo
	bootOnPowerOn bool
}

// SetName is a 
//
func (b *Blade) SetName(ctx context.Context, region string, zone string, rack string, blade int64) error {

	key, err := GetKeyForBlade(DefinitionTable, region, zone, rack, blade)

	if nil != err {
		return err
	}

	b.Region = region
	b.Zone   = zone
	b.Rack   = rack
	b.ID     = blade

	b.Key    = key
	b.record = nil

	return nil
}

// SetDetails is a
//
func (b *Blade) SetDetails(ctx context.Context, details *pb.BladeDetails) error {

	b.details = details

	return 	nil
}

// GetDetails is a
//
func (b *Blade) GetDetails(ctx context.Context) (*pb.BladeDetails, error) {
	if b.details == nil {
		return nil, ErrDetailsNotAvailable("blade")
	}
	
	return 	b.details, nil
}


// SetCapacity is a
//
func (b *Blade) SetCapacity(ctx context.Context, capacity *pb.BladeCapacity) error {
	b.capacity = capacity

	return nil
}

// SetBootInfo is a
//
func (b *Blade) SetBootInfo(ctx context.Context, bootOnPowerOn bool, bootInfo *pb.BladeBootInfo) error {
	b.bootOnPowerOn = bootOnPowerOn
	b.bootInfo      = bootInfo

	return nil
}

// GetCapacity is a
//
func (b *Blade) GetCapacity(ctx context.Context) (*pb.BladeCapacity, error) {
	if b.capacity == nil {
		return nil, ErrCapacityNotAvailable("blade")
	}

	return b.capacity, nil
}

// GetBootInfo is a
//
func (b *Blade) GetBootInfo(ctx context.Context) (bool, *pb.BladeBootInfo, error) {
	if b.bootInfo == nil {
		return false, nil, ErrBootInfoNotAvailable("blade")
	}

	return b.bootOnPowerOn, b.bootInfo, nil
}

// Create is
//
func (b *Blade) Create(ctx context.Context)  (int64, error)  {

	if b.details == nil {
		return store.RevisionInvalid, ErrDetailsNotAvailable("blade")
	}

	if b.capacity == nil {
		return store.RevisionInvalid, ErrCapacityNotAvailable("blade")
	}

	if b.bootInfo == nil {
		return store.RevisionInvalid, ErrBootInfoNotAvailable("blade")
	}

	record := &pb.StoreRecordDefinitionBlade{
		Details:       b.details,
		Capacity:      b.capacity,
		BootOnPowerOn: b.bootOnPowerOn,
		BootInfo:      b.bootInfo,
	}

	v, err := store.Encode(record)

	if err != nil {
		return store.RevisionInvalid, err
	}

	rev, err := b.Store.Create(ctx, store.KeyRootInventory, b.Key, v)

	if err == store.ErrStoreAlreadyExists(b.Key) {
		return store.RevisionInvalid, ErrfPduAlreadyExists(b.Zone, b.Rack, b.ID)
	}

	b.revision = rev

	return b.revision, nil
}

// Read is
//
func (b *Blade) Read(ctx context.Context) (*pb.DefinitionBlade, error) {

	return nil, nil
}

// Update is
//
func (b *Blade) Update(ctx context.Context) error {

	return nil
}

// Delete is
//
func (b *Blade) Delete(ctx context.Context) error {

	return nil
}



