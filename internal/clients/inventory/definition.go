package inventory

import (
	"context"
	"fmt"
	"strconv"
	"strings"

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
		Table:         table,
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

	keyIndexEntry, err := GetKeyForIndexEntryRegion(table, region)

	if nil != err {
		return nil, err
	}

	key, err := GetKeyForRegion(table, region)

	if nil != err {
		return nil, err
	}

	r := &Region{
		Store:         store,
		KeyChildIndex: keyIndex,
		KeyIndexEntry: keyIndexEntry,
		Key:           key,
		Table:         table,
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

	keyIndexEntry, err := GetKeyForIndexEntryZone(table, region, zone)

	if nil != err {
		return nil, err
	}

	key, err := GetKeyForZone(table, region, zone)

	if nil != err {
		return nil, err
	}

	z := &Zone{
		Store:         store,
		KeyChildIndex: keyIndex,
		KeyIndexEntry: keyIndexEntry,
		Key:           key,
		Table:         table,
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

	keyIndexEntry, err := GetKeyForIndexEntryRack(table, region, zone, rack)

	if nil != err {
		return nil, err
	}

	key, err := GetKeyForRack(table, region, zone, rack)

	if nil != err {
		return nil, err
	}

	r := &Rack{
		Store:         store,
		KeyIndexPdu:   keyIndexPdu,
		KeyIndexTor:   keyIndexTor,
		KeyIndexBlade: keyIndexBlade,
		KeyIndexEntry: keyIndexEntry,
		Key:           key,
		Table:         table,
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

	keyIndexEntry, err := GetKeyForIndexEntryPdu(table, region, zone, rack, id)

	if nil != err {
		return nil, err
	}

	key, err := GetKeyForPdu(table, region, zone, rack, id)

	if nil != err {
		return nil, err
	}

	p := &Pdu{
		Store:         store,
		KeyIndexEntry: keyIndexEntry,
		Key:           key,
		Table:         table,
		Region:        region,
		Zone:          zone,
		Rack:          rack,
		ID:            id,
	}

	return p, nil
}

// NewTor is a convenience function used to construct a Pdu struct
// from scratch rather than relative to its parent.
//
func NewTor(ctx context.Context, store *store.Store, table string, region string, zone string, rack string, id int64) (*Tor, error) {

	keyIndexEntry, err := GetKeyForIndexEntryTor(table, region, zone, rack, id)

	if nil != err {
		return nil, err
	}

	key, err := GetKeyForTor(table, region, zone, rack, id)

	if nil != err {
		return nil, err
	}

	t := &Tor{
		Store:         store,
		KeyIndexEntry: keyIndexEntry,
		Key:           key,
		Table:         table,
		Region:        region,
		Zone:          zone,
		Rack:          rack,
		ID:            id,
	}

	return t, nil
}

// NewBlade is a convenience function used to construct a Pdu struct
// from scratch rather than relative to its parent.
//
func NewBlade(ctx context.Context, store *store.Store, table string, region string, zone string, rack string, id int64) (*Blade, error) {

	keyIndexEntry, err := GetKeyForIndexEntryBlade(table, region, zone, rack, id)

	if nil != err {
		return nil, err
	}

	key, err := GetKeyForBlade(table, region, zone, rack, id)

	if nil != err {
		return nil, err
	}

	b := &Blade{
		Store:         store,
		KeyIndexEntry: keyIndexEntry,
		Key:           key,
		Table:         table,
		Region:        region,
		Zone:          zone,
		Rack:          rack,
		ID:            id,
	}

	return b, nil
}




// Root is a 
//
type Root struct {

	Store         *store.Store
	KeyChildIndex  string
	Table          string

	revision       int64
	revisionRecord int64
	revisionStore  int64
	details        *pb.RootDetails
}


// SetDetails is a
//
func (r *Root) SetDetails(ctx context.Context, details *pb.RootDetails) {

	r.details  = details
	r.revision = store.RevisionInvalid
}

// GetDetails is a
//
func (r *Root) GetDetails(ctx context.Context) *pb.RootDetails {
	return 	r.details
}

// GetRevision is a
//
func (r *Root) GetRevision(ctx context.Context) int64 {
	return r.revision
}

// GetRevisionRecord is a
//
func (r *Root) GetRevisionRecord(ctx context.Context) int64 {
	return r.revisionRecord
}

// GetRevisionStore is a
//
func (r *Root) GetRevisionStore(ctx context.Context) int64 {
	return r.revisionStore
}

// Create is a
//
func (r *Root) Create(ctx context.Context) (int64, error) {
	return store.RevisionInvalid, ErrFunctionNotAvailable
}

// Read is a
//
func (r *Root) Read(ctx context.Context) (int64, error) {
	return 	store.RevisionInvalid, ErrFunctionNotAvailable
}

// Update is a
//
func (r *Root) Update(ctx context.Context, unconditional bool) (int64, error) {
	return store.RevisionInvalid, ErrFunctionNotAvailable
}

// Delete is a
//
func (r *Root) Delete(ctx context.Context, unconditional bool) (int64, error) {
	return 	store.RevisionInvalid, ErrFunctionNotAvailable
}

// NewChild is a
//
func (r *Root) NewChild(ctx context.Context, name string) (*Region, error) {

	region, err := NewRegion(ctx, r.Store, r.Table, name)

	if err != nil {
		return nil, err
	}

	return region, nil
}

// ListChildren is a
//
func (r *Root) ListChildren(ctx context.Context) (int64, []string, error) {
	
	records, rev, err := r.Store.List(ctx, store.KeyRootInventory, r.KeyChildIndex)

	if err == store.ErrStoreIndexNotFound(r.KeyChildIndex) {
		return store.RevisionInvalid, nil, ErrIndexNotFound(r.KeyChildIndex)
	}

	if err != nil {
		return store.RevisionInvalid, nil, err
	}

	names := make([]string, 0, len(*records))

	for k, v := range *records {
	
		name := strings.TrimPrefix(k, r.KeyChildIndex)

		if name != store.GetNormalizedName(v.Value) {
			return store.RevisionInvalid, nil, ErrfIndexKeyValueMismatch(r.Table, name, v.Value)
		}

		names = append(names, v.Value)
	}

	return rev, names, nil
}

// FetchChildren is a
//
func (r *Root) FetchChildren(ctx context.Context) (int64, *map[string]Region, error) {

	rev, names, err := r.ListChildren(ctx)

	if err != nil {
		return store.RevisionInvalid, nil, err
	}

	children := make(map[string]Region, len(names))

	for _, v := range names {

		child, err := r.NewChild(ctx, v)

		if err != nil {
			return store.RevisionInvalid, nil, err
		}
	
		_, err = child.Read(ctx)

		if err != nil {
			return store.RevisionInvalid, nil, err
		}

		children[v] = *child
	}

	return rev, &children, nil
}




// Region is a
//
type Region struct {

	Store          *store.Store
	KeyChildIndex  string
	KeyIndexEntry  string
	Key            string
	Table          string
	Region         string

	revision       int64
	revisionRecord int64
	revisionStore  int64

	details       *pb.RegionDetails
	record        *pb.StoreRecordDefinitionRegion
}

// SetDetails is a
//
func (r *Region) SetDetails(ctx context.Context, details *pb.RegionDetails) {

	r.details  = details
	r.revision = store.RevisionInvalid
}

// GetDetails is a
//
func (r *Region) GetDetails(ctx context.Context) *pb.RegionDetails {
	return r.details
}

// GetRevision is a
//
func (r *Region) GetRevision(ctx context.Context) int64 {
	return r.revision
}

// GetRevisionRecord is a
//
func (r *Region) GetRevisionRecord(ctx context.Context) int64 {
	return r.revisionRecord
}

// GetRevisionStore is a
//
func (r *Region) GetRevisionStore(ctx context.Context) int64 {
	return r.revisionStore
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

	// Create the child and its index as an atomic pair.
	//
	keySet := &map[string]string{
		r.KeyIndexEntry : r.Region,
		r.Key           : v,
	}

	rev, err := r.Store.CreateMultiple(ctx, store.KeyRootInventory, keySet)

	switch (err) {
	case nil:
	case store.ErrStoreAlreadyExists(r.KeyIndexEntry): return store.RevisionInvalid, ErrfRegionAlreadyExists(r.Region)
	case store.ErrStoreAlreadyExists(r.Key):           return store.RevisionInvalid, ErrfRegionAlreadyExists(r.Region)
	default:                                           return store.RevisionInvalid, err
	}

	r.record         = record
	r.revision       = rev
	r.revisionRecord = rev
	r.revisionStore  = rev

	return r.revision, nil
}

// Read is
//
func (r *Region) Read(ctx context.Context) (int64, error) {

	v, rev, err := r.Store.Read(ctx, store.KeyRootInventory, r.Key)

	if err == store.ErrStoreKeyNotFound(r.Key) {
		return store.RevisionInvalid, ErrfRegionNotFound(r.Region)
	}

	record := &pb.StoreRecordDefinitionRegion{}

	if err = store.Decode(*v, record); err != nil {
		return store.RevisionInvalid, err
	}

	r.details        = record.Details
	r.record         = record
	r.revision       = rev
	r.revisionRecord = rev
	r.revisionStore  = rev

	return r.revision, nil
}

// Update is
//
func (r *Region) Update(ctx context.Context, unconditional bool) (int64, error) {

	var revUpdate = r.revisionRecord

	if unconditional == true {
		revUpdate = store.RevisionInvalid
	}

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

	rev, err := r.Store.Update(ctx, store.KeyRootInventory, r.Key, revUpdate, v)

	if err == store.ErrStoreKeyNotFound(r.Key) {
		return store.RevisionInvalid, ErrfRegionNotFound(r.Region)
	}

	if err != nil {
		return store.RevisionInvalid, err
	}

	r.record         = record
	r.revision       = rev
	r.revisionRecord = rev
	r.revisionStore  = rev

	return rev, nil
}

// Delete is
//
func (r *Region) Delete(ctx context.Context, unconditional bool) (int64, error) {

	var revDelete = r.revisionRecord

	if unconditional == true {
		revDelete = store.RevisionInvalid
	}

	// TODO - use delete multiple to remove object and index?
	//
	rev, err := r.Store.Delete(ctx, store.KeyRootInventory, r.Key, revDelete)

	if err == store.ErrStoreKeyNotFound(r.Key) {
		return store.RevisionInvalid, ErrfRegionNotFound(r.Region)
	}

	if err != nil {
		return store.RevisionInvalid, err
	}

	r.record         = nil
	r.revision       = rev
	r.revisionRecord = rev
	r.revisionStore  = rev

	return rev, nil
}

// NewChild is a 
//
func (r *Region) NewChild(ctx context.Context, name string) (*Zone, error) {

	z, err := NewZone(ctx, r.Store, r.Table, r.Region, name)

	if err != nil {
		return nil, err
	}

	return z, err
}

// ListChildren is a
//
func (r *Region) ListChildren(ctx context.Context) (int64, []string, error) {
	
	records, rev, err := r.Store.List(ctx, store.KeyRootInventory, r.KeyChildIndex)

	if err == store.ErrStoreIndexNotFound(r.KeyChildIndex) {
		return store.RevisionInvalid, nil, ErrIndexNotFound(r.KeyChildIndex)
	}

	if err != nil {
		return store.RevisionInvalid, nil, err
	}

	names := make([]string, 0, len(*records))

	for k, v := range *records {
	
		name := strings.TrimPrefix(k, r.KeyChildIndex)

		if name != store.GetNormalizedName(v.Value) {
			return store.RevisionInvalid, nil, ErrfIndexKeyValueMismatch(r.Table, name, v.Value)
		}

		names = append(names, v.Value)
	}

	return rev, names, nil
}

// FetchChildren is a
//
func (r *Region) FetchChildren(ctx context.Context) (int64, *map[string]Zone, error) {

	rev, names, err := r.ListChildren(ctx)

	if err != nil {
		return store.RevisionInvalid, nil, err
	}

	children := make(map[string]Zone, len(names))

	// TODO - use read multiple? If broken out into "groups", returned rev is the highest found 
	//
	for _, v := range names {

		child, err := r.NewChild(ctx, v)

		if err != nil {
			return store.RevisionInvalid, nil, err
		}
	
		rev, err = child.Read(ctx)

		if err != nil {
			return store.RevisionInvalid, nil, err
		}

		children[v] = *child
	}

	return rev, &children, nil
}




// Zone is a
//
type Zone struct {

	Store          *store.Store
	KeyChildIndex  string
	KeyIndexEntry  string
	Key            string
	Table          string
	Region         string
	Zone           string

	revision       int64
	revisionRecord int64
	revisionStore  int64
	details        *pb.ZoneDetails
	record         *pb.StoreRecordDefinitionZone
}

// SetDetails is a
//
func (z *Zone) SetDetails(ctx context.Context, details *pb.ZoneDetails) {

	z.details = details
	z.revision = store.RevisionInvalid
}

// GetDetails is a
//
func (z *Zone) GetDetails(ctx context.Context) *pb.ZoneDetails {
	return z.details
}

// GetRevision is a
//
func (z *Zone) GetRevision(ctx context.Context) (int64) {
	return z.revision
}

// GetRevisionRecord is a
//
func (z *Zone) GetRevisionRecord(ctx context.Context) (int64) {
	return z.revisionRecord
}

// GetRevisionStore is a
//
func (z *Zone) GetRevisionStore(ctx context.Context) (int64) {
	return z.revisionStore
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

	// Create the child and its index as an atomic pair.
	//
	keySet := &map[string]string{
		z.KeyIndexEntry : z.Zone,
		z.Key           : v,
	}

	rev, err := z.Store.CreateMultiple(ctx, store.KeyRootInventory, keySet)

	switch (err) {
	case nil:
	case store.ErrStoreAlreadyExists(z.KeyIndexEntry): return store.RevisionInvalid, ErrfZoneAlreadyExists(z.Region, z.Zone)
	case store.ErrStoreAlreadyExists(z.Key):           return store.RevisionInvalid, ErrfZoneAlreadyExists(z.Region, z.Zone)
	default:                                           return store.RevisionInvalid, err
	}

	z.record         = record
	z.revision       = rev
	z.revisionRecord = rev
	z.revisionStore  = rev

	return z.revision, nil
}

// Read is
//
func (z *Zone) Read(ctx context.Context) (int64, error) {

	v, rev, err := z.Store.Read(ctx, store.KeyRootInventory, z.Key)

	if err == store.ErrStoreKeyNotFound(z.Key) {
		return store.RevisionInvalid, ErrfZoneNotFound(z.Region, z.Zone)
	}

	record := &pb.StoreRecordDefinitionZone{}

	if err = store.Decode(*v, record); err != nil {
		return store.RevisionInvalid, err
	}

	z.details        = record.Details
	z.record         = record
	z.revision       = rev
	z.revisionRecord = rev
	z.revisionStore  = rev

	return z.revision, nil
}

// Update is
//
func (z *Zone) Update(ctx context.Context, unconditional bool) (int64, error) {

	var revUpdate = z.revisionRecord

	if unconditional == true {
		revUpdate = store.RevisionInvalid
	}

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

	rev, err := z.Store.Update(ctx, store.KeyRootInventory, z.Key, revUpdate, v)

	if err == store.ErrStoreKeyNotFound(z.Key) {
		return store.RevisionInvalid, ErrfZoneNotFound(z.Region, z.Zone)
	}

	if err != nil {
		return store.RevisionInvalid, err
	}

	z.record         = record
	z.revision       = rev
	z.revisionRecord = rev
	z.revisionStore  = rev

	return rev, nil
}

// Delete is
//
func (z *Zone) Delete(ctx context.Context, unconditional bool) (int64, error) {

	var revDelete = z.revisionRecord

	if unconditional == true {
		revDelete = store.RevisionInvalid
	}

	rev, err := z.Store.Delete(ctx, store.KeyRootInventory, z.Key, revDelete)

	if err == store.ErrStoreKeyNotFound(z.Key) {
		return store.RevisionInvalid, ErrfZoneNotFound(z.Region, z.Zone)
	}

	if err != nil {
		return store.RevisionInvalid, err
	}

	z.record         = nil
	z.revision       = rev
	z.revisionRecord = rev
	z.revisionStore  = rev

	return rev, nil
}

// NewChild is a 
//
func (z *Zone) NewChild(ctx context.Context, name string) (*Rack, error) {

	r, err := NewRack(ctx, z.Store, z.Table, z.Region, z.Zone, name)

	if err != nil {
		return nil, err
	}

	return r, err
}

// ListChildren is a
//
func (z *Zone) ListChildren(ctx context.Context) (int64, []string, error) {
	
	records, rev, err := z.Store.List(ctx, store.KeyRootInventory, z.KeyChildIndex)

	if err == store.ErrStoreIndexNotFound(z.KeyChildIndex) {
		return store.RevisionInvalid, nil, ErrIndexNotFound(z.KeyChildIndex)
	}

	if err != nil {
		return store.RevisionInvalid, nil, err
	}

	names := make([]string, 0, len(*records))

	for k, v := range *records {
	
		name := strings.TrimPrefix(k, z.KeyChildIndex)

		if name != store.GetNormalizedName(v.Value) {
			return store.RevisionInvalid, nil, ErrfIndexKeyValueMismatch(z.Table, name, v.Value)
		}

		names = append(names, v.Value)
	}

	return rev, names, nil
}

// FetchChildren is a
//
func (z *Zone) FetchChildren(ctx context.Context) (int64, *map[string]Rack, error) {

	rev, names, err := z.ListChildren(ctx)

	if err != nil {
		return store.RevisionInvalid, nil, err
	}

	children := make(map[string]Rack, len(names))

	for _, v := range names {

		child, err := z.NewChild(ctx, v)

		if err != nil {
			return store.RevisionInvalid, nil, err
		}
	
		_, err = child.Read(ctx)

		if err != nil {
			return store.RevisionInvalid, nil, err
		}

		children[v] = *child
	}

	return rev, &children, nil
}




// Rack is a
//
type Rack struct {

	Store          *store.Store
	KeyIndexPdu    string
	KeyIndexTor    string
	KeyIndexBlade  string
	KeyIndexEntry  string
	Key            string
	Table          string
	Region         string
	Zone           string
	Rack           string

	revision       int64
	revisionRecord int64
	revisionStore  int64

	details       *pb.RackDetails
	record        *pb.StoreRecordDefinitionRack

}

// SetDetails is a
//
func (r *Rack) SetDetails(ctx context.Context, details *pb.RackDetails) {

	r.details  = details
	r.revision = store.RevisionInvalid
}

// GetDetails is a
//
func (r *Rack) GetDetails(ctx context.Context) *pb.RackDetails {
	return r.details
}

// GetRevision is a
//
func (r *Rack) GetRevision(ctx context.Context) (int64) {
	return r.revision
}

// GetRevisionRecord is a
//
func (r *Rack) GetRevisionRecord(ctx context.Context) (int64) {
	return r.revisionRecord
}

// GetRevisionStore is a
//
func (r *Rack) GetRevisionStore(ctx context.Context) (int64) {
	return r.revisionStore
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

	// Create the child and its index as an atomic pair.
	//
	keySet := &map[string]string{
		r.KeyIndexEntry : r.Rack,
		r.Key           : v,
	}

	rev, err := r.Store.CreateMultiple(ctx, store.KeyRootInventory, keySet)

	switch (err) {
	case nil:
	case store.ErrStoreAlreadyExists(r.KeyIndexEntry): return store.RevisionInvalid, ErrfRackAlreadyExists(r.Region, r.Zone, r.Rack)
	case store.ErrStoreAlreadyExists(r.Key):           return store.RevisionInvalid, ErrfRackAlreadyExists(r.Region, r.Zone, r.Rack)
	default:                                           return store.RevisionInvalid, err
	}

	r.record         = record
	r.revision       = rev
	r.revisionRecord = rev
	r.revisionStore  = rev

	return r.revision, nil
}

// Read is
//
func (r *Rack) Read(ctx context.Context) (int64, error) {

	v, rev, err := r.Store.Read(ctx, store.KeyRootInventory, r.Key)

	if err == store.ErrStoreKeyNotFound(r.Key) {
		return store.RevisionInvalid, ErrfRackNotFound(r.Region, r.Zone, r.Rack)
	}

	record := &pb.StoreRecordDefinitionRack{}

	if err = store.Decode(*v, record); err != nil {
		return store.RevisionInvalid, err
	}

	r.details        = record.Details
	r.record         = record
	r.revision       = rev
	r.revisionRecord = rev
	r.revisionStore  = rev

	return r.revision, nil
}

// Update is
//
func (r *Rack) Update(ctx context.Context, unconditional bool) (int64, error) {

	var revUpdate = r.revisionRecord

	if unconditional == true {
		revUpdate = store.RevisionInvalid
	}

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

	rev, err := r.Store.Update(ctx, store.KeyRootInventory, r.Key, revUpdate, v)

	if err == store.ErrStoreKeyNotFound(r.Key) {
		return store.RevisionInvalid, ErrfRegionNotFound(r.Region)
	}

	if err != nil {
		return store.RevisionInvalid, err
	}

	r.record         = record
	r.revision       = rev
	r.revisionRecord = rev
	r.revisionStore  = rev

	return rev, nil
}

// Delete is
//
func (r *Rack) Delete(ctx context.Context, unconditional bool) (int64, error) {

	var revDelete = r.revisionRecord

	if unconditional == true {
		revDelete = store.RevisionInvalid
	}

	rev, err := r.Store.Delete(ctx, store.KeyRootInventory, r.Key, revDelete)

	if err == store.ErrStoreKeyNotFound(r.Key) {
		return store.RevisionInvalid, ErrfRegionNotFound(r.Region)
	}

	if err != nil {
		return store.RevisionInvalid, err
	}

	r.record         = nil
	r.revision       = rev
	r.revisionRecord = rev
	r.revisionStore  = rev

	return rev, nil
}

// NewChild is a 
//
func (r *Rack) NewChild(ctx context.Context, name string) (*Zone, error) {
	return nil, ErrFunctionNotAvailable
}

// NewPdu is a 
//
func (r *Rack) NewPdu(ctx context.Context, ID int64) (*Pdu, error) {

	p, err := NewPdu(ctx, r.Store, r.Table, r.Region, r.Zone, r.Rack, ID)

	if err != nil {
		return nil, err
	}

	return p, err
}

// NewTor is a 
//
func (r *Rack) NewTor(ctx context.Context, ID int64) (*Tor, error) {

	t, err := NewTor(ctx, r.Store, r.Table, r.Region, r.Zone, r.Rack, ID)

	if err != nil {
		return nil, err
	}

	return t, err
}

// NewBlade is a 
//
func (r *Rack) NewBlade(ctx context.Context, ID int64) (*Blade, error) {

	b, err := NewBlade(ctx, r.Store, r.Table, r.Region, r.Zone, r.Rack, ID)

	if err != nil {
		return nil, err
	}

	return b, err
}

// FetchChildren is a
//
func (r *Rack) FetchChildren(ctx context.Context) (int64, *map[string]Zone, error) {
	return store.RevisionInvalid, nil, ErrFunctionNotAvailable
}


// ListPdus is a
//
func (r *Rack) ListPdus(ctx context.Context) (int64, []int64, error) {

	records, rev, err := r.Store.List(ctx, store.KeyRootInventory, r.KeyIndexPdu)

	if err == store.ErrStoreIndexNotFound(r.KeyIndexPdu) {
		return store.RevisionInvalid, nil, ErrIndexNotFound(r.KeyIndexPdu)
	}

	if err != nil {
		return store.RevisionInvalid, nil, err
	}

	names := make([]int64, 0, len(*records))

	for k, v := range *records {
	
		name := strings.TrimPrefix(k, r.KeyIndexPdu)

		// Verify that the "index" part of the name is numeric
		//
		intName, err := strconv.ParseInt(name, 10, 0)

		if err != nil {
			return store.RevisionInvalid, nil, ErrfPduIndexInvalid(r.Table, r.Zone, r.Rack, name)
		}

		intValue, err := strconv.ParseInt(v.Value, 10, 0)

		if err != nil {
			return store.RevisionInvalid, nil, ErrfPduIndexInvalid(r.Table, r.Zone, r.Rack, v.Value)
		}

		if intName != intValue {
			return store.RevisionInvalid, nil, ErrfIndexKeyValueMismatch(r.Table, name, v.Value)
		}

		names = append(names, intValue)
	}

	return rev, names, nil
}

// ListTors is a
//
func (r *Rack) ListTors(ctx context.Context) (int64, []int64, error) {

	records, rev, err := r.Store.List(ctx, store.KeyRootInventory, r.KeyIndexTor)

	if err == store.ErrStoreIndexNotFound(r.KeyIndexTor) {
		return store.RevisionInvalid, nil, ErrIndexNotFound(r.KeyIndexTor)
	}

	if err != nil {
		return store.RevisionInvalid, nil, err
	}

	names := make([]int64, 0, len(*records))

	for k, v := range *records {
	
		name := strings.TrimPrefix(k, r.KeyIndexTor)

		// Verify that the "index" part of the name is numeric
		//
		intName, err := strconv.ParseInt(name, 10, 0)

		if err != nil {
			return store.RevisionInvalid, nil, ErrfTorIndexInvalid(r.Table, r.Zone, r.Rack, name)
		}

		intValue, err := strconv.ParseInt(v.Value, 10, 0)

		if err != nil {
			return store.RevisionInvalid, nil, ErrfTorIndexInvalid(r.Table, r.Zone, r.Rack, v.Value)
		}

		if intName != intValue {
			return store.RevisionInvalid, nil, ErrfIndexKeyValueMismatch(r.Table, name, v.Value)
		}

		names = append(names, intValue)
	}

	return rev, names, nil
}

// ListBlades is a
//
func (r *Rack) ListBlades(ctx context.Context) (int64, []int64, error) {

	records, rev, err := r.Store.List(ctx, store.KeyRootInventory, r.KeyIndexBlade)

	if err == store.ErrStoreIndexNotFound(r.KeyIndexBlade) {
		return store.RevisionInvalid, nil, ErrIndexNotFound(r.KeyIndexBlade)
	}

	if err != nil {
		return store.RevisionInvalid, nil, err
	}

	names := make([]int64, 0, len(*records))

	for k, v := range *records {
	
		name := strings.TrimPrefix(k, r.KeyIndexBlade)

		// Verify that the "index" part of the name is numeric
		//
		intName, err := strconv.ParseInt(name, 10, 0)

		if err != nil {
			return store.RevisionInvalid, nil, ErrfBladeIndexInvalid(r.Table, r.Zone, r.Rack, name)
		}

		intValue, err := strconv.ParseInt(v.Value, 10, 0)

		if err != nil {
			return store.RevisionInvalid, nil, ErrfBladeIndexInvalid(r.Table, r.Zone, r.Rack, v.Value)
		}

		if intName != intValue {
			return store.RevisionInvalid, nil, ErrfIndexKeyValueMismatch(r.Table, name, v.Value)
		}

		names = append(names, intValue)
	}

	return rev, names, nil
}

// FetchPdus is a
//
func (r *Rack) FetchPdus(ctx context.Context) (int64, *map[int64]Pdu, error) {

	rev, names, err := r.ListPdus(ctx)

	if err != nil {
		return store.RevisionInvalid, nil, err
	}

	pdus := make(map[int64]Pdu, len(names))

	for _, v := range names {

		pdu, err := r.NewPdu(ctx, v)

		if err != nil {
			return store.RevisionInvalid, nil, err
		}
	
		_, err = pdu.Read(ctx)

		if err != nil {
			return store.RevisionInvalid, nil, err
		}

		pdus[v] = *pdu
	}

	return rev, &pdus, nil
}

// FetchTors is a
//
func (r *Rack) FetchTors(ctx context.Context) (int64, *map[int64]Tor, error) {

	rev, names, err := r.ListTors(ctx)

	if err != nil {
		return store.RevisionInvalid, nil, err
	}

	tors := make(map[int64]Tor, len(names))

	for _, v := range names {

		tor, err := r.NewTor(ctx, v)

		if err != nil {
			return store.RevisionInvalid, nil, err
		}
	
		_, err = tor.Read(ctx)

		if err != nil {
			return store.RevisionInvalid, nil, err
		}

		tors[v] = *tor
	}

	return rev, &tors, nil
}

// FetchBlades is a
//
func (r *Rack) FetchBlades(ctx context.Context) (int64, *map[int64]Blade, error) {

	rev, names, err := r.ListBlades(ctx)

	if err != nil {
		return store.RevisionInvalid, nil, err
	}

	blades := make(map[int64]Blade, len(names))

	for _, v := range names {

		blade, err := r.NewBlade(ctx, v)

		if err != nil {
			return store.RevisionInvalid, nil, err
		}
	
		_, err = blade.Read(ctx)

		if err != nil {
			return store.RevisionInvalid, nil, err
		}

		blades[v] = *blade
	}

	return rev, &blades, nil
}




// Pdu is a
//
type Pdu struct {

	Store          *store.Store
	Key            string
	KeyIndexEntry  string
	Table          string
	Region         string
	Zone           string
	Rack           string
	ID             int64

	revision       int64
	revisionRecord int64
	revisionStore  int64
	details        *pb.PduDetails
	ports          *map[int64]*pb.PowerPort
	record         *pb.StoreRecordDefinitionPdu
}

// SetDetails is a
//
func (p *Pdu) SetDetails(ctx context.Context, details *pb.PduDetails) {
	p.details  = details
	p.revision = store.RevisionInvalid
}

// GetDetails is a
//
func (p *Pdu) GetDetails(ctx context.Context) *pb.PduDetails {
	return 	p.details
}

// GetRevision is a
//
func (p *Pdu) GetRevision(ctx context.Context) int64 {
	return p.revision
}

// GetRevisionRecord is a
//
func (p *Pdu) GetRevisionRecord(ctx context.Context) int64 {
	return p.revisionRecord
}

// GetRevisionStore is a
//
func (p *Pdu) GetRevisionStore(ctx context.Context) int64 {
	return p.revisionStore
}

// SetPorts is a
//
func (p *Pdu) SetPorts(ctx context.Context, ports *map[int64]*pb.PowerPort) {
	p.ports    = ports
	p.revision = store.RevisionInvalid
}

// GetPorts is a 
//
func (p *Pdu) GetPorts(ctx context.Context) *map[int64]*pb.PowerPort {
	return p.ports
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

	// Create the child and its index as an atomic pair.
	//
	keySet := &map[string]string{
		p.KeyIndexEntry : fmt.Sprintf("%d", p.ID),
		p.Key           : v,
	}

	rev, err := p.Store.CreateMultiple(ctx, store.KeyRootInventory, keySet)

	switch (err) {
	case nil:
	case store.ErrStoreAlreadyExists(p.KeyIndexEntry): return store.RevisionInvalid, ErrfPduAlreadyExists(p.Region, p.Zone, p.Rack, p.ID)
	case store.ErrStoreAlreadyExists(p.Key):           return store.RevisionInvalid, ErrfPduAlreadyExists(p.Region, p.Zone, p.Rack, p.ID)
	default:                                           return store.RevisionInvalid, err
	}

	p.record         = record
	p.revision       = rev
	p.revisionRecord = rev
	p.revisionStore  = rev

	return p.revision, nil
}

// Read is
//
func (p *Pdu) Read(ctx context.Context) (int64, error) {

	v, rev, err := p.Store.Read(ctx, store.KeyRootInventory, p.Key)

	if err == store.ErrStoreKeyNotFound(p.Key) {
		return store.RevisionInvalid, ErrfPduNotFound(p.Region, p.Zone, p.Rack, p.ID)
	}

	record := &pb.StoreRecordDefinitionPdu{}

	if err = store.Decode(*v, record); err != nil {
		return store.RevisionInvalid, err
	}

	p.details        = record.Details
	p.ports          = &record.Ports
	p.record         = record
	p.revision       = rev
	p.revisionRecord = rev
	p.revisionStore  = rev

	return p.revision, nil
}

// Update is
//
func (p *Pdu) Update(ctx context.Context, unconditional bool) (int64, error) {

	var revUpdate = p.revisionRecord

	if unconditional == true {
		revUpdate = store.RevisionInvalid
	}

	if p.details == nil {
		return store.RevisionInvalid, ErrDetailsNotAvailable("pdu")
	}

	if p.ports == nil {
		return store.RevisionInvalid, ErrPortsNotAvailable("pdu")
	}

	record := &pb.StoreRecordDefinitionPdu{
		Details: p.details,
		Ports: *p.ports,
	}

	v, err := store.Encode(record)

	if err != nil {
		return store.RevisionInvalid, err
	}

	rev, err := p.Store.Update(ctx, store.KeyRootInventory, p.Key, revUpdate, v)

	if err == store.ErrStoreKeyNotFound(p.Key) {
		return store.RevisionInvalid, ErrfPduNotFound(p.Region, p.Zone, p.Rack, p.ID)
	}

	if err != nil {
		return store.RevisionInvalid, err
	}

	p.record         = record
	p.revision       = rev
	p.revisionRecord = rev
	p.revisionStore  = rev

	return rev, nil
}

// Delete is
//
func (p *Pdu) Delete(ctx context.Context, unconditional bool) (int64, error) {

	var revDelete = p.revisionRecord

	if unconditional == true {
		revDelete = store.RevisionInvalid
	}

	rev, err := p.Store.Delete(ctx, store.KeyRootInventory, p.Key, revDelete)

	if err == store.ErrStoreKeyNotFound(p.Key) {
		return store.RevisionInvalid, ErrfPduNotFound(p.Region, p.Zone, p.Rack, p.ID)
	}

	if err != nil {
		return store.RevisionInvalid, err
	}

	p.record         = nil
	p.revision       = rev
	p.revisionRecord = rev
	p.revisionStore  = rev

	return rev, nil
}




// Tor is a
//
type Tor struct {

	Store          *store.Store
	Key            string
	KeyIndexEntry  string
	Table          string
	Region         string
	Zone           string
	Rack           string
	ID             int64

	revision       int64
	revisionRecord int64
	revisionStore  int64
	details        *pb.TorDetails
	ports          *map[int64]*pb.NetworkPort
	record         *pb.StoreRecordDefinitionTor
}

// SetDetails is a
//
func (t *Tor) SetDetails(ctx context.Context, details *pb.TorDetails) {
	t.details  = details
	t.revision = store.RevisionInvalid
}

// GetDetails is a
//
func (t *Tor) GetDetails(ctx context.Context) *pb.TorDetails {
	return 	t.details
}

// GetRevision is a
//
func (t *Tor) GetRevision(ctx context.Context) int64 {
	return t.revision
}

// GetRevisionRecord is a
//
func (t *Tor) GetRevisionRecord(ctx context.Context) int64 {
	return t.revisionRecord
}

// GetRevisionStore is a
//
func (t *Tor) GetRevisionStore(ctx context.Context) int64 {
	return t.revisionStore
}

// SetPorts is a
//
func (t *Tor) SetPorts(ctx context.Context, ports *map[int64]*pb.NetworkPort) {
	t.ports    = ports
	t.revision = store.RevisionInvalid
}

// GetPorts is a 
//
func (t *Tor) GetPorts(ctx context.Context) *map[int64]*pb.NetworkPort {
	return t.ports
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

	// Create the child and its index as an atomic pair.
	//
	keySet := &map[string]string{
		t.KeyIndexEntry : fmt.Sprintf("%d", t.ID),
		t.Key           : v,
	}

	rev, err := t.Store.CreateMultiple(ctx, store.KeyRootInventory, keySet)

	switch (err) {
	case nil:
	case store.ErrStoreAlreadyExists(t.KeyIndexEntry): return store.RevisionInvalid, ErrfTorAlreadyExists(t.Region, t.Zone, t.Rack, t.ID)
	case store.ErrStoreAlreadyExists(t.Key):           return store.RevisionInvalid, ErrfTorAlreadyExists(t.Region, t.Zone, t.Rack, t.ID)
	default:                                           return store.RevisionInvalid, err
	}

	t.record         = record
	t.revision       = rev
	t.revisionRecord = rev
	t.revisionStore  = rev

	return t.revision, nil
}

// Read is
//
func (t *Tor) Read(ctx context.Context) (int64, error) {

	v, rev, err := t.Store.Read(ctx, store.KeyRootInventory, t.Key)

	if err == store.ErrStoreKeyNotFound(t.Key) {
		return store.RevisionInvalid, ErrfTorNotFound(t.Region, t.Zone, t.Rack, t.ID)
	}

	record := &pb.StoreRecordDefinitionTor{}

	if err = store.Decode(*v, record); err != nil {
		return store.RevisionInvalid, err
	}

	t.details        = record.Details
	t.ports          = &record.Ports
	t.record         = record
	t.revision       = rev
	t.revisionRecord = rev
	t.revisionStore  = rev

	return t.revision, nil
}

// Update is
//
func (t *Tor) Update(ctx context.Context, unconditional bool) (int64, error) {

	var revUpdate = t.revisionRecord

	if unconditional == true {
		revUpdate = store.RevisionInvalid
	}

	if t.details == nil {
		return store.RevisionInvalid, ErrDetailsNotAvailable("tor")
	}

	if t.ports == nil {
		return store.RevisionInvalid, ErrPortsNotAvailable("tor")
	}

	record := &pb.StoreRecordDefinitionTor{
		Details: t.details,
		Ports: *t.ports,
	}

	v, err := store.Encode(record)

	if err != nil {
		return store.RevisionInvalid, err
	}

	rev, err := t.Store.Update(ctx, store.KeyRootInventory, t.Key, revUpdate, v)

	if err == store.ErrStoreKeyNotFound(t.Key) {
		return store.RevisionInvalid, ErrfTorNotFound(t.Region, t.Zone, t.Rack, t.ID)
	}

	if err != nil {
		return store.RevisionInvalid, err
	}

	t.record         = record
	t.revision       = rev
	t.revisionRecord = rev
	t.revisionStore  = rev

	return rev, nil
}

// Delete is
//
func (t *Tor) Delete(ctx context.Context, unconditional bool) (int64, error) {

	var revDelete = t.revisionRecord

	if unconditional == true {
		revDelete = store.RevisionInvalid
	}

	rev, err := t.Store.Delete(ctx, store.KeyRootInventory, t.Key, revDelete)

	if err == store.ErrStoreKeyNotFound(t.Key) {
		return store.RevisionInvalid, ErrfTorNotFound(t.Region, t.Zone, t.Rack, t.ID)
	}

	if err != nil {
		return store.RevisionInvalid, err
	}

	t.record         = nil
	t.revision       = rev
	t.revisionRecord = rev
	t.revisionStore  = rev

	return rev, nil
}




// Blade is a
//
type Blade struct {

	Store          *store.Store
	Key            string
	KeyIndexEntry  string
	Table          string
	Region         string
	Zone           string
	Rack           string
	ID             int64

	revision       int64
	revisionRecord int64
	revisionStore  int64
	details        *pb.BladeDetails
	capacity       *pb.BladeCapacity
	bootInfo       *pb.BladeBootInfo
	bootOnPowerOn  bool
	record         *pb.StoreRecordDefinitionBlade
}

// SetDetails is a
//
func (b *Blade) SetDetails(ctx context.Context, details *pb.BladeDetails) {
	b.details  = details
	b.revision = store.RevisionInvalid
}

// GetDetails is a
//
func (b *Blade) GetDetails(ctx context.Context) *pb.BladeDetails {
	return 	b.details
}

// GetRevision is a
//
func (b *Blade) GetRevision(ctx context.Context) int64 {
	return b.revision
}

// GetRevisionRecord is a
//
func (b *Blade) GetRevisionRecord(ctx context.Context) int64 {
	return b.revisionRecord
}

// GetRevisionStore is a
//
func (b *Blade) GetRevisionStore(ctx context.Context) int64 {
	return b.revisionStore
}

// SetCapacity is a
//
func (b *Blade) SetCapacity(ctx context.Context, capacity *pb.BladeCapacity) {
	b.capacity = capacity
	b.revision = store.RevisionInvalid
}

// SetBootInfo is a
//
func (b *Blade) SetBootInfo(ctx context.Context, bootOnPowerOn bool, bootInfo *pb.BladeBootInfo) {
	b.bootOnPowerOn = bootOnPowerOn
	b.bootInfo      = bootInfo
	b.revision      = store.RevisionInvalid
}

// GetCapacity is a
//
func (b *Blade) GetCapacity(ctx context.Context) *pb.BladeCapacity {
	return b.capacity
}

// GetBootInfo is a
//
func (b *Blade) GetBootInfo(ctx context.Context) (bool, *pb.BladeBootInfo) {
	return b.bootOnPowerOn, b.bootInfo
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

	// Create the child and its index as an atomic pair.
	//
	keySet := &map[string]string{
		b.KeyIndexEntry : fmt.Sprintf("%d", b.ID),
		b.Key           : v,
	}

	rev, err := b.Store.CreateMultiple(ctx, store.KeyRootInventory, keySet)

	switch (err) {
	case nil:
	case store.ErrStoreAlreadyExists(b.KeyIndexEntry): return store.RevisionInvalid, ErrfBladeAlreadyExists(b.Region, b.Zone, b.Rack, b.ID)
	case store.ErrStoreAlreadyExists(b.Key):           return store.RevisionInvalid, ErrfBladeAlreadyExists(b.Region, b.Zone, b.Rack, b.ID)
	default:                                           return store.RevisionInvalid, err
	}

	b.record         = record
	b.revision       = rev
	b.revisionRecord = rev
	b.revisionStore  = rev

	return b.revision, nil
}

// Read is
//
func (b *Blade) Read(ctx context.Context) (int64, error) {

	v, rev, err := b.Store.Read(ctx, store.KeyRootInventory, b.Key)

	if err == store.ErrStoreKeyNotFound(b.Key) {
		return store.RevisionInvalid, ErrfPduNotFound(b.Region, b.Zone, b.Rack, b.ID)
	}

	record := &pb.StoreRecordDefinitionBlade{}

	if err = store.Decode(*v, record); err != nil {
		return store.RevisionInvalid, err
	}

	b.details        = record.Details
	b.capacity       = record.Capacity
	b.bootInfo       = record.BootInfo
	b.bootOnPowerOn  = record.BootOnPowerOn
	b.record         = record
	b.revision       = rev
	b.revisionRecord = rev
	b.revisionStore  = rev

	return b.revision, nil
}

// Update is
//
func (b *Blade) Update(ctx context.Context, unconditional bool) (int64, error) {

	var revUpdate = b.revisionRecord

	if unconditional == true {
		revUpdate = store.RevisionInvalid
	}

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
		BootInfo:      b.bootInfo,
		BootOnPowerOn: b.bootOnPowerOn,
	}

	v, err := store.Encode(record)

	if err != nil {
		return store.RevisionInvalid, err
	}

	rev, err := b.Store.Update(ctx, store.KeyRootInventory, b.Key, revUpdate, v)

	if err == store.ErrStoreKeyNotFound(b.Key) {
		return store.RevisionInvalid, ErrfBladeNotFound(b.Region, b.Zone, b.Rack, b.ID)
	}

	if err != nil {
		return store.RevisionInvalid, err
	}

	b.record         = record
	b.revision       = rev
	b.revisionRecord = rev
	b.revisionStore  = rev

	return rev, nil
}

// Delete is
//
func (b *Blade) Delete(ctx context.Context, unconditional bool) (int64, error) {

	var revDelete = b.revisionRecord

	if unconditional == true {
		revDelete = store.RevisionInvalid
	}

	// This should really use store.DeleteMultiple() except that doesn't exist
	// yet to allow multiple key,value pairs to removed as an atomic update.
	//
	// Oops.
	//
	// Once the DeleteMultiple() routine has been added, this should change to
	// make use of the new call.
	//
	rev, err := b.Store.Delete(ctx, store.KeyRootInventory, b.Key, revDelete)

	if err == store.ErrStoreKeyNotFound(b.Key) {
		return store.RevisionInvalid, ErrfBladeNotFound(b.Region, b.Zone, b.Rack, b.ID)
	}

	if err != nil {
		return store.RevisionInvalid, err
	}

	_, err = b.Store.Delete(ctx, store.KeyRootInventory, b.KeyIndexEntry, store.RevisionInvalid)

	if err == store.ErrStoreKeyNotFound(b.KeyIndexEntry) {
		return store.RevisionInvalid, ErrfBladeNotFound(b.Region, b.Zone, b.Rack, b.ID)
	}

	if err != nil {
		return store.RevisionInvalid, err
	}

	b.record         = nil
	b.revision       = rev
	b.revisionRecord = rev
	b.revisionStore  = rev

	return rev, nil
}
