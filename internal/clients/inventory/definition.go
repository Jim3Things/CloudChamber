// This module contain the structures and methods to operate on the persisted definition
// table within the inventory package.

package inventory

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/Jim3Things/CloudChamber/internal/clients/store"
	"github.com/Jim3Things/CloudChamber/pkg/errors"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/inventory"
)

// NewRoot returns a root object which acts as a well-knwon point in a namespace
// and which can be used to navigate the namespace for a given table.
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

// NewRegion is a convenience function used to construct a Region object
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

// NewZone is a convenience function used to construct a Zone object
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

// NewRack is a convenience function used to construct a Rack object
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

// NewPdu is a convenience function used to construct a Pdu object
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

// NewTor is a convenience function used to construct a Tor object
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

// NewBlade is a convenience function used to construct a Blade object
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

// Root is a structure representing the well-known root of the namespace. It 
// is used to locate the regions within the namespace represented by the table
// field.
//
// The root object is an in-memory construct only and cannot be persisted to
// the store, or retrieved from it.
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

// SetDetails is used to attach some attribute information to the object.
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
	return store.RevisionInvalid, errors.ErrFunctionNotAvailable
}

// Read is a
//
func (r *Root) Read(ctx context.Context) (int64, error) {
	return 	store.RevisionInvalid, errors.ErrFunctionNotAvailable
}

// Update is a
//
func (r *Root) Update(ctx context.Context, unconditional bool) (int64, error) {
	return store.RevisionInvalid, errors.ErrFunctionNotAvailable
}

// Delete is a
//
func (r *Root) Delete(ctx context.Context, unconditional bool) (int64, error) {
	return 	store.RevisionInvalid, errors.ErrFunctionNotAvailable
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

	if err == errors.ErrStoreIndexNotFound(r.KeyChildIndex) {
		return store.RevisionInvalid, nil, errors.ErrIndexNotFound(r.KeyChildIndex)
	}

	if err != nil {
		return store.RevisionInvalid, nil, err
	}

	names := make([]string, 0, len(*records))

	for k, v := range *records {
	
		name := strings.TrimPrefix(k, r.KeyChildIndex)

		if name != store.GetNormalizedName(v.Value) {
			return store.RevisionInvalid, nil, errors.ErrIndexKeyValueMismatch{
				Namespace: r.Table,
				Key:       name,
				Value:     v.Value,
			}
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

// Region is a structure representing a region object. This object can be used
// to operate on the associated region records in the underlying store, or to
// navigate to child zone objects. The object can store information fetched
// from the underlying store, or as a staging area in preparation for updates
// to the store.
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

// SetDetails is used to attach some attribute information to the object.
//
// The attribute information is not persisted to the store until an Update()
// call is made.
//
// The current revision of the region object is reset
//
func (r *Region) SetDetails(ctx context.Context, details *pb.RegionDetails) {

	r.details  = details
	r.revision = store.RevisionInvalid
}

// GetDetails is used to extract the attribute information from the object. The
// attribute information must have been previously read from the store (see
// the Read() method) or attached via a SetDetails() call.
//
// May return nil if there are no attributes currently held in the object.
//
func (r *Region) GetDetails(ctx context.Context) *pb.RegionDetails {
	return r.details
}

// GetRevision returns the revision of the details field within the object. This
// will be either the revision of the object in the store after a Create() or
// Read() call or be store.RevisionInvalid if the details have been set or no
// Create() or Read() call has been executed.
//
func (r *Region) GetRevision(ctx context.Context) int64 {
	return r.revision
}

// GetRevisionRecord returns the revision of the underlying store object as 
// determined at the time of the last Create() or Read() for the object. The
// record revision is not reset by a SetDetails() call and is used when
// performing either a conditional update or conditional delete using the
// object.
//
func (r *Region) GetRevisionRecord(ctx context.Context) int64 {
	return r.revisionRecord
}

// GetRevisionStore returns the revison of the underlying store ifself as 
// determined at the time of the last Read() for the object. The
// store revision is not reset by a SetDetails() call and is provided 
// for information only.
//
func (r *Region) GetRevisionStore(ctx context.Context) int64 {
	return r.revisionStore
}

// Create is used to create a record in the underlying store for the
// object along with the associated index information.
//
// The underlying store record will contain the information currently
// held in the object.
//
// Once the store operation completes successfully, the revision fields
// in the object will be updated to that returned by the store. These can
// either be reetrieved by one of the GetRevisionXxx() call or used for
// subsequent conditional operaions such as a conditional Update() call.
//
func (r *Region) Create(ctx context.Context) (int64, error) {

	if r.details == nil {
		return store.RevisionInvalid, errors.ErrDetailsNotAvailable("region")
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
	case errors.ErrStoreAlreadyExists(r.KeyIndexEntry): return store.RevisionInvalid, errors.ErrRegionAlreadyExists(r.Region)
	case errors.ErrStoreAlreadyExists(r.Key):           return store.RevisionInvalid, errors.ErrRegionAlreadyExists(r.Region)
	default:                                            return store.RevisionInvalid, err
	}

	r.record         = record
	r.revision       = rev
	r.revisionRecord = rev
	r.revisionStore  = rev

	return r.revision, nil
}

// Read is used to load a record from the underlying store to populate the
// fields in the object and determine the revision values associated with
// that record.
//
// Once the Read() has completed successfully the details and other
// information for the object can be retrieved by any of the GetXxx() methods
// for that obect.
//
func (r *Region) Read(ctx context.Context) (int64, error) {

	v, rev, err := r.Store.Read(ctx, store.KeyRootInventory, r.Key)

	if err == errors.ErrStoreKeyNotFound(r.Key) {
		return store.RevisionInvalid, errors.ErrRegionNotFound(r.Region)
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

// Update is used to persist the information in the fields of the object to
// a record in the underlying store. The update can be either unconditional
// by setting the unconditional parameter to true, or conditional based on
// the revision of the object compared to the revision of the associated
// record in the underlying store.
//
// Once the store operation completes successfully, the revision information
// in the object is updated with that returned from the store.
//
// Update() has no effect on the index information for the object.
//
func (r *Region) Update(ctx context.Context, unconditional bool) (int64, error) {

	var revUpdate = r.revisionRecord

	if unconditional == true {
		revUpdate = store.RevisionInvalid
	}

	if r.details == nil {
		return store.RevisionInvalid, errors.ErrDetailsNotAvailable("region")
	}

	record := &pb.StoreRecordDefinitionRegion{
		Details: r.details,
	}

	v, err := store.Encode(record)

	if err != nil {
		return store.RevisionInvalid, err
	}

	rev, err := r.Store.Update(ctx, store.KeyRootInventory, r.Key, revUpdate, v)

	if err == errors.ErrStoreKeyNotFound(r.Key) {
		return store.RevisionInvalid, errors.ErrRegionNotFound(r.Region)
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

// Delete is used to remove the persisted copy of the object from the
// store along with any index information needed to navigate to or
// through that object. The delete can be either unconditional by
// setting the unconditional parameter to true, or conditional based
// on the revision of the object compared to the revision of the
// associated record in the underlying store.
//
// Deleting the record from the underlying store  has no effect on the
// values held in the fields of the object other than updating the
// revision information using the information returned by the store
// operation.
//
func (r *Region) Delete(ctx context.Context, unconditional bool) (int64, error) {

	var revDelete = r.revisionRecord

	if unconditional == true {
		revDelete = store.RevisionInvalid
	}

	// TODO - use delete multiple to remove object and index?
	//
	rev, err := r.Store.Delete(ctx, store.KeyRootInventory, r.Key, revDelete)

	if err == errors.ErrStoreKeyNotFound(r.Key) {
		return store.RevisionInvalid, errors.ErrRegionNotFound(r.Region)
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

// NewChild creates a new child object for the zone within the current
// region using the supplied name. This new object can be used for
// further navigation or for actions involving operations against the
// associated record in the underlying store.
//
// No information is fetched from the underlying store so the attribute
// and revisions fields within the oject are not valid.
//
func (r *Region) NewChild(ctx context.Context, name string) (*Zone, error) {

	z, err := NewZone(ctx, r.Store, r.Table, r.Region, name)

	if err != nil {
		return nil, err
	}

	return z, err
}

// ListChildren uses the current object to discover the names of all the
// zone child objects in the underlying store for the the current region
// object, The elements of the returned list can be used in subsequent
// NewChild() calls to create new zone objects.
//
func (r *Region) ListChildren(ctx context.Context) (int64, []string, error) {
	
	records, rev, err := r.Store.List(ctx, store.KeyRootInventory, r.KeyChildIndex)

	if err == errors.ErrStoreIndexNotFound(r.KeyChildIndex) {
		return store.RevisionInvalid, nil, errors.ErrIndexNotFound(r.KeyChildIndex)
	}

	if err != nil {
		return store.RevisionInvalid, nil, err
	}

	names := make([]string, 0, len(*records))

	for k, v := range *records {
	
		name := strings.TrimPrefix(k, r.KeyChildIndex)

		if name != store.GetNormalizedName(v.Value) {
			return store.RevisionInvalid, nil, errors.ErrIndexKeyValueMismatch{Namespace: r.Table, Key: name, Value: v.Value}
		}

		names = append(names, v.Value)
	}

	return rev, names, nil
}

// FetchChildren is used to discover all the child zone objects in the
// underlying store for the current region object and to generate a new
// zone object for each of those children. It is a convenience wrapper
// around ListChildren() followed by a NewChild() on each name discovered.
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

// Zone is a structure representing a zone object. This object can be used
// to operate on the associated zone records in the underlying store, or to
// navigate to child rack objects. The object can store information fetched
// from the underlying store, or as a staging area in preparation for updates
// to the store.
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

// SetDetails is used to attach some attribute information to the object.
//
// The attribute information is not persisted to the store until an Update()
// call is made.
//
// The current revision of the zone object is reset
//
func (z *Zone) SetDetails(ctx context.Context, details *pb.ZoneDetails) {

	z.details = details
	z.revision = store.RevisionInvalid
}

// GetDetails is used to extract the attribute information from the object. The
// attribute information must have been previously read from the store (see
// the Read() method) or attached via a SetDetails() call.
//
// May return nil if there are no attributes currently held in the object.
//
func (z *Zone) GetDetails(ctx context.Context) *pb.ZoneDetails {
	return z.details
}

// GetRevision returns the revision of the details field within the object. This
// will be either the revision of the object in the store after a Create() or
// Read() call or be store.RevisionInvalid if the details have been set or no
// Create() or Read() call has been executed.
//
func (z *Zone) GetRevision(ctx context.Context) (int64) {
	return z.revision
}

// GetRevisionRecord returns the revision of the underlying store object as 
// determined at the time of the last Create() or Read() for the object. The
// record revision is not reset by a SetDetails() call and is used when
// performing either a conditional update or conditional delete using the
// object.
//
func (z *Zone) GetRevisionRecord(ctx context.Context) (int64) {
	return z.revisionRecord
}

// GetRevisionStore returns the revison of the underlying store ifself as 
// determined at the time of the last Read() for the object. The
// store revision is not reset by a SetDetails() call and is provided 
// for information only.
//
func (z *Zone) GetRevisionStore(ctx context.Context) (int64) {
	return z.revisionStore
}

// Create is used to create a record in the underlying store for the
// object along with the associated index information.
//
// The underlying store record will contain the information currently
// held in the object.
//
// Once the store operation completes successfully, the revision fields
// in the object will be updated to that returned by the store. These can
// either be reetrieved by one of the GetRevisionXxx() call or used for
// subsequent conditional operaions such as a conditional Update() call.
//
func (z *Zone) Create(ctx context.Context) (int64, error) {

	if z.details == nil {
		return store.RevisionInvalid, errors.ErrDetailsNotAvailable("zone")
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
	case errors.ErrStoreAlreadyExists(z.KeyIndexEntry): return store.RevisionInvalid, errors.ErrZoneAlreadyExists{Region: z.Region, Zone: z.Zone}
	case errors.ErrStoreAlreadyExists(z.Key):           return store.RevisionInvalid, errors.ErrZoneAlreadyExists{Region: z.Region, Zone: z.Zone}
	default:                                            return store.RevisionInvalid, err
	}

	z.record         = record
	z.revision       = rev
	z.revisionRecord = rev
	z.revisionStore  = rev

	return z.revision, nil
}

// Read is used to load a record from the underlying store to populate the
// fields in the object and determine the revision values associated with
// that record.
//
// Once the Read() has completed successfully the details and other
// information for the object can be retrieved by any of the GetXxx() methods
// for that obect.
//
func (z *Zone) Read(ctx context.Context) (int64, error) {

	v, rev, err := z.Store.Read(ctx, store.KeyRootInventory, z.Key)

	if err == errors.ErrStoreKeyNotFound(z.Key) {
		return store.RevisionInvalid, errors.ErrZoneNotFound{Region: z.Region, Zone: z.Zone}
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

// Update is used to persist the information in the fields of the object to
// a record in the underlying store. The update can be either unconditional
// by setting the unconditional parameter to true, or conditional based on
// the revision of the object compared to the revision of the associated
// record in the underlying store.
//
// Once the store operation completes successfully, the revision information
// in the object is updated with that returned from the store.
//
// Update() has no effect on the index information for the object.
//
func (z *Zone) Update(ctx context.Context, unconditional bool) (int64, error) {

	var revUpdate = z.revisionRecord

	if unconditional == true {
		revUpdate = store.RevisionInvalid
	}

	if z.details == nil {
		return store.RevisionInvalid, errors.ErrDetailsNotAvailable("zone")
	}

	record := &pb.StoreRecordDefinitionZone{
		Details: z.details,
	}

	v, err := store.Encode(record)

	if err != nil {
		return store.RevisionInvalid, err
	}

	rev, err := z.Store.Update(ctx, store.KeyRootInventory, z.Key, revUpdate, v)

	if err == errors.ErrStoreKeyNotFound(z.Key) {
		return store.RevisionInvalid, errors.ErrZoneNotFound{Region: z.Region, Zone: z.Zone}
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

// Delete is used to remove the persisted copy of the object from the
// store along with any index information needed to navigate to or
// through that object. The delete can be either unconditional by
// setting the unconditional parameter to true, or conditional based
// on the revision of the object compared to the revision of the
// associated record in the underlying store.
//
// Deleting the record from the underlying store  has no effect on the
// values held in the fields of the object other than updating the
// revision information using the information returned by the store
// operation.
//
func (z *Zone) Delete(ctx context.Context, unconditional bool) (int64, error) {

	var revDelete = z.revisionRecord

	if unconditional == true {
		revDelete = store.RevisionInvalid
	}

	rev, err := z.Store.Delete(ctx, store.KeyRootInventory, z.Key, revDelete)

	if err == errors.ErrStoreKeyNotFound(z.Key) {
		return store.RevisionInvalid, errors.ErrZoneNotFound{Region: z.Region, Zone: z.Zone}
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

// NewChild creates a new child object for the zone within the current
// region using the supplied name. This new object can be used for
// further navigation or for actions involving operations against the
// associated record in the underlying store.
//
// No information is fetched from the underlying store so the attribute
// and revisions fields within the oject are not valid.
//
func (z *Zone) NewChild(ctx context.Context, name string) (*Rack, error) {

	r, err := NewRack(ctx, z.Store, z.Table, z.Region, z.Zone, name)

	if err != nil {
		return nil, err
	}

	return r, err
}

// ListChildren uses the current object to discover the names of all the
// rack child objects in the underlying store for the the current zone
// object, The elements of the returned list can be used in subsequent
// NewChild() calls to create new rack objects.
//
func (z *Zone) ListChildren(ctx context.Context) (int64, []string, error) {
	
	records, rev, err := z.Store.List(ctx, store.KeyRootInventory, z.KeyChildIndex)

	if err == errors.ErrStoreIndexNotFound(z.KeyChildIndex) {
		return store.RevisionInvalid, nil, errors.ErrIndexNotFound(z.KeyChildIndex)
	}

	if err != nil {
		return store.RevisionInvalid, nil, err
	}

	names := make([]string, 0, len(*records))

	for k, v := range *records {
	
		name := strings.TrimPrefix(k, z.KeyChildIndex)

		if name != store.GetNormalizedName(v.Value) {
			return store.RevisionInvalid, nil, errors.ErrIndexKeyValueMismatch{Namespace: z.Table, Key: name, Value: v.Value}
		}

		names = append(names, v.Value)
	}

	return rev, names, nil
}

// FetchChildren is used to discover all the child rack objects in the
// underlying store for the current zone object and to generate a new
// rack object for each of those children. It is a convenience wrapper
// around ListChildren() followed by a NewChild() on each name discovered.
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

// Rack is a structure representing a rack object. This object can be used
// to operate on the associated rack records in the underlying store, or to
// navigate to child pdu, tor or blade objects. The object can store
// information fetched from the underlying store, or as a staging area in
// preparation for updates to the store.
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

// SetDetails is used to attach some attribute information to the object.
//
// The attribute information is not persisted to the store until an Update()
// call is made.
//
// The current revision of the rack object is reset
//
func (r *Rack) SetDetails(ctx context.Context, details *pb.RackDetails) {

	r.details  = details
	r.revision = store.RevisionInvalid
}

// GetDetails is used to extract the attribute information from the object. The
// attribute information must have been previously read from the store (see
// the Read() method) or attached via a SetDetails() call.
//
// May return nil if there are no attributes currently held in the object.
//
func (r *Rack) GetDetails(ctx context.Context) *pb.RackDetails {
	return r.details
}

// GetRevision returns the revision of the details field within the object. This
// will be either the revision of the object in the store after a Create() or
// Read() call or be store.RevisionInvalid if the details have been set or no
// Create() or Read() call has been executed.
//
func (r *Rack) GetRevision(ctx context.Context) (int64) {
	return r.revision
}

// GetRevisionRecord returns the revision of the underlying store object as 
// determined at the time of the last Create() or Read() for the object. The
// record revision is not reset by a SetDetails() call and is used when
// performing either a conditional update or conditional delete using the
// object.
//
func (r *Rack) GetRevisionRecord(ctx context.Context) (int64) {
	return r.revisionRecord
}

// GetRevisionStore returns the revison of the underlying store ifself as 
// determined at the time of the last Read() for the object. The
// store revision is not reset by a SetDetails() call and is provided 
// for information only.
//
func (r *Rack) GetRevisionStore(ctx context.Context) (int64) {
	return r.revisionStore
}

// Create is used to create a record in the underlying store for the
// object along with the associated index information.
//
// The underlying store record will contain the information currently
// held in the object.
//
// Once the store operation completes successfully, the revision fields
// in the object will be updated to that returned by the store. These can
// either be reetrieved by one of the GetRevisionXxx() call or used for
// subsequent conditional operaions such as a conditional Update() call.
//
func (r *Rack) Create(ctx context.Context) (int64, error) {

	if r.details == nil {
		return store.RevisionInvalid, errors.ErrDetailsNotAvailable("rack")
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
	case errors.ErrStoreAlreadyExists(r.KeyIndexEntry): return store.RevisionInvalid, errors.ErrRackAlreadyExists{Region: r.Region, Zone: r.Zone, Rack: r.Rack}
	case errors.ErrStoreAlreadyExists(r.Key):           return store.RevisionInvalid, errors.ErrRackAlreadyExists{Region: r.Region, Zone: r.Zone, Rack: r.Rack}
	default:                                            return store.RevisionInvalid, err
	}

	r.record         = record
	r.revision       = rev
	r.revisionRecord = rev
	r.revisionStore  = rev

	return r.revision, nil
}

// Read is used to load a record from the underlying store to populate the
// fields in the object and determine the revision values associated with
// that record.
//
// Once the Read() has completed successfully the details and other
// information for the object can be retrieved by any of the GetXxx() methods
// for that obect.
//
func (r *Rack) Read(ctx context.Context) (int64, error) {

	v, rev, err := r.Store.Read(ctx, store.KeyRootInventory, r.Key)

	if err == errors.ErrStoreKeyNotFound(r.Key) {
		return store.RevisionInvalid, errors.ErrRackNotFound{Region: r.Region, Zone: r.Zone, Rack: r.Rack}
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

// Update is used to persist the information in the fields of the object to
// a record in the underlying store. The update can be either unconditional
// by setting the unconditional parameter to true, or conditional based on
// the revision of the object compared to the revision of the associated
// record in the underlying store.
//
// Once the store operation completes successfully, the revision information
// in the object is updated with that returned from the store.
//
// Update() has no effect on the index information for the object.
//
func (r *Rack) Update(ctx context.Context, unconditional bool) (int64, error) {

	var revUpdate = r.revisionRecord

	if unconditional == true {
		revUpdate = store.RevisionInvalid
	}

	if r.details == nil {
		return store.RevisionInvalid, errors.ErrDetailsNotAvailable("rack")
	}

	record := &pb.StoreRecordDefinitionRack{
		Details: r.details,
	}

	v, err := store.Encode(record)

	if err != nil {
		return store.RevisionInvalid, err
	}

	rev, err := r.Store.Update(ctx, store.KeyRootInventory, r.Key, revUpdate, v)

	if err == errors.ErrStoreKeyNotFound(r.Key) {
		return store.RevisionInvalid, errors.ErrRackNotFound{Region: r.Region, Zone: r.Zone, Rack: r.Rack}
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

// Delete is used to remove the persisted copy of the object from the
// store along with any index information needed to navigate to or
// through that object. The delete can be either unconditional by
// setting the unconditional parameter to true, or conditional based
// on the revision of the object compared to the revision of the
// associated record in the underlying store.
//
// Deleting the record from the underlying store  has no effect on the
// values held in the fields of the object other than updating the
// revision information using the information returned by the store
// operation.
//
func (r *Rack) Delete(ctx context.Context, unconditional bool) (int64, error) {

	var revDelete = r.revisionRecord

	if unconditional == true {
		revDelete = store.RevisionInvalid
	}

	rev, err := r.Store.Delete(ctx, store.KeyRootInventory, r.Key, revDelete)

	if err == errors.ErrStoreKeyNotFound(r.Key) {
		return store.RevisionInvalid, errors.ErrRackNotFound{Region: r.Region, Zone: r.Zone, Rack: r.Rack}
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

// NewChild is a stub function for rack objects as there are no generic
// child objects. Instead one of the specialized functions
//
//    NewPdu()
//    NewTor()
//    NewBlade()
//
// should be called to construct an object for the appropriate specialized
// child.
//
func (r *Rack) NewChild(ctx context.Context, name string) (*Zone, error) {
	return nil, errors.ErrFunctionNotAvailable
}

// NewPdu creates a new child object for the pdu within the current
// rack using the supplied identifier. This new object can be used for
// for actions involving operations against the associated record in
// the underlying store.
//
// No information is fetched from the underlying store so the attribute
// and revisions fields within the oject are not valid.
//
func (r *Rack) NewPdu(ctx context.Context, ID int64) (*Pdu, error) {

	p, err := NewPdu(ctx, r.Store, r.Table, r.Region, r.Zone, r.Rack, ID)

	if err != nil {
		return nil, err
	}

	return p, err
}

// NewTor creates a new child object for the tor within the current
// rack using the supplied identifier. This new object can be used for
// for actions involving operations against the associated record in
// the underlying store.
//
// No information is fetched from the underlying store so the attribute
// and revisions fields within the oject are not valid.
//
func (r *Rack) NewTor(ctx context.Context, ID int64) (*Tor, error) {

	t, err := NewTor(ctx, r.Store, r.Table, r.Region, r.Zone, r.Rack, ID)

	if err != nil {
		return nil, err
	}

	return t, err
}

// NewBlade creates a new child object for the blade within the current
// rack using the supplied identifier. This new object can be used for
// for actions involving operations against the associated record in
// the underlying store.
//
// No information is fetched from the underlying store so the attribute
// and revisions fields within the oject are not valid.
//
func (r *Rack) NewBlade(ctx context.Context, ID int64) (*Blade, error) {

	b, err := NewBlade(ctx, r.Store, r.Table, r.Region, r.Zone, r.Rack, ID)

	if err != nil {
		return nil, err
	}

	return b, err
}

// ListChildren is a stub function for rack objects as there are no
// generic child objects. Instead one of the specialized functions
//
//    ListPdu()
//    ListTor()
//    ListBlade()
//
// should be called to construct an object for the appropriate specialized
// child.
//
func (r *Rack) ListChildren(ctx context.Context) (int64, *[]string, error) {
	return store.RevisionInvalid, nil, errors.ErrFunctionNotAvailable
}

// FetchChildren is a stub function for rack objects as there are no
// generic child objects. Instead one of the specialized functions
//
//    FetchPdu()
//    FetchTor()
//    FetchBlade()
//
// should be called to construct an object for the appropriate specialized
// child.
//
func (r *Rack) FetchChildren(ctx context.Context) (int64, *map[string]interface{}, error) {
	return store.RevisionInvalid, nil, errors.ErrFunctionNotAvailable
}

// ListPdus uses the current object to discover the names of all the
// pdu child objects in the underlying store for the the current rack
// object, The elements of the returned list can be used in subsequent
// NewPdu() calls to create new pdu objects.
//
func (r *Rack) ListPdus(ctx context.Context) (int64, []int64, error) {

	records, rev, err := r.Store.List(ctx, store.KeyRootInventory, r.KeyIndexPdu)

	if err == errors.ErrStoreIndexNotFound(r.KeyIndexPdu) {
		return store.RevisionInvalid, nil, errors.ErrIndexNotFound(r.KeyIndexPdu)
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
			return store.RevisionInvalid, nil, errors.ErrPduIndexInvalid{Region: r.Region, Zone: r.Zone, Rack: r.Rack, Pdu: name}
		}

		intValue, err := strconv.ParseInt(v.Value, 10, 0)

		if err != nil {
			return store.RevisionInvalid, nil, errors.ErrPduIndexInvalid{Region: r.Region, Zone: r.Zone, Rack: r.Rack, Pdu: v.Value}
		}

		if intName != intValue {
			return store.RevisionInvalid, nil, errors.ErrIndexKeyValueMismatch{Namespace: r.Table, Key: name, Value: v.Value}
		}

		names = append(names, intValue)
	}

	return rev, names, nil
}

// ListTors uses the current object to discover the names of all the
// tor child objects in the underlying store for the the current rack
// object, The elements of the returned list can be used in subsequent
// NewTor() calls to create new tor objects.
//
func (r *Rack) ListTors(ctx context.Context) (int64, []int64, error) {

	records, rev, err := r.Store.List(ctx, store.KeyRootInventory, r.KeyIndexTor)

	if err == errors.ErrStoreIndexNotFound(r.KeyIndexTor) {
		return store.RevisionInvalid, nil, errors.ErrIndexNotFound(r.KeyIndexTor)
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
			return store.RevisionInvalid, nil, errors.ErrTorIndexInvalid{Region: r.Region, Zone: r.Zone, Rack: r.Rack, Tor: name}
		}

		intValue, err := strconv.ParseInt(v.Value, 10, 0)

		if err != nil {
			return store.RevisionInvalid, nil, errors.ErrTorIndexInvalid{Region: r.Region, Zone: r.Zone, Rack: r.Rack, Tor: v.Value}
		}

		if intName != intValue {
			return store.RevisionInvalid, nil, errors.ErrIndexKeyValueMismatch{Namespace: r.Table, Key: name, Value: v.Value}
		}

		names = append(names, intValue)
	}

	return rev, names, nil
}

// ListBlades uses the current object to discover the names of all the
// blade child objects in the underlying store for the the current rack
// object, The elements of the returned list can be used in subsequent
// NewBlade() calls to create new blade objects.
//
func (r *Rack) ListBlades(ctx context.Context) (int64, []int64, error) {

	records, rev, err := r.Store.List(ctx, store.KeyRootInventory, r.KeyIndexBlade)

	if err == errors.ErrStoreIndexNotFound(r.KeyIndexBlade) {
		return store.RevisionInvalid, nil, errors.ErrIndexNotFound(r.KeyIndexBlade)
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
			return store.RevisionInvalid, nil, errors.ErrBladeIndexInvalid{Region: r.Table, Zone: r.Zone, Rack: r.Rack, Blade: name}
		}

		intValue, err := strconv.ParseInt(v.Value, 10, 0)

		if err != nil {
			return store.RevisionInvalid, nil, errors.ErrBladeIndexInvalid{Region: r.Table, Zone: r.Zone, Rack: r.Rack, Blade: v.Value}
		}

		if intName != intValue {
			return store.RevisionInvalid, nil, errors.ErrIndexKeyValueMismatch{Namespace: r.Table, Key: name, Value: v.Value}
		}

		names = append(names, intValue)
	}

	return rev, names, nil
}

// FetchPdus is used to discover all the child pdu objects in the
// underlying store for the current rack object and to generate a new
// pdu object for each of those children. It is a convenience wrapper
// around ListPdus() followed by a NewPdu() on each name discovered.
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

// FetchTors is used to discover all the child tor objects in the
// underlying store for the current rack object and to generate a new
// tor object for each of those children. It is a convenience wrapper
// around ListTors() followed by a NewTor() on each name discovered.
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

// FetchBlades is used to discover all the child blade objects in the
// underlying store for the current rack object and to generate a new
// blade object for each of those children. It is a convenience wrapper
// around ListBlades() followed by a NewBlade() on each name discovered.
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

// Pdu is a structure representing a pdu object. This object can be used
// to operate on the associated pdu records in the underlying store. The
// object can hold information fetched from the underlying store, or as
// a staging area in preparation for updates to the store.
//
// Pdu is a specialization of a child object for a rack parent.
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

// SetDetails is used to attach some attribute information to the object.
//
// The attribute information is not persisted to the store until an Update()
// call is made.
//
// The current revision of the pdu object is reset
//
func (p *Pdu) SetDetails(ctx context.Context, details *pb.PduDetails) {
	p.details  = details
	p.revision = store.RevisionInvalid
}

// GetDetails is used to extract the attribute information from the object. The
// attribute information must have been previously read from the store (see
// the Read() method) or attached via a SetDetails() call.
//
// May return nil if there are no attributes currently held in the object.
//
func (p *Pdu) GetDetails(ctx context.Context) *pb.PduDetails {
	return 	p.details
}

// GetRevision returns the revision of the details field within the object. This
// will be either the revision of the object in the store after a Create() or
// Read() call or be store.RevisionInvalid if the details have been set or no
// Create() or Read() call has been executed.
//
func (p *Pdu) GetRevision(ctx context.Context) int64 {
	return p.revision
}

// GetRevisionRecord returns the revision of the underlying store object as 
// determined at the time of the last Create() or Read() for the object. The
// record revision is not reset by a SetDetails() call and is used when
// performing either a conditional update or conditional delete using the
// object.
//
func (p *Pdu) GetRevisionRecord(ctx context.Context) int64 {
	return p.revisionRecord
}

// GetRevisionStore returns the revison of the underlying store ifself as 
// determined at the time of the last Read() for the object. The
// store revision is not reset by a SetDetails() call and is provided 
// for information only.
//
func (p *Pdu) GetRevisionStore(ctx context.Context) int64 {
	return p.revisionStore
}

// SetPorts is used to attach some power port information to the object.
//
// The port information is not persisted to the store until an Update()
// call is made.
//
// The current revision of the pdu object is reset
//
func (p *Pdu) SetPorts(ctx context.Context, ports *map[int64]*pb.PowerPort) {
	p.ports    = ports
	p.revision = store.RevisionInvalid
}

// GetPorts is used to extract the power port information from the object.
// The port information must have been previously read from the store (see
// the Read() method) or attached via a SetPorts() call.
//
// May return nil if there are no power port information currently held
// in the object.
//
func (p *Pdu) GetPorts(ctx context.Context) *map[int64]*pb.PowerPort {
	return p.ports
}

// Create is used to create a record in the underlying store for the
// object along with the associated index information.
//
// The underlying store record will contain the information currently
// held in the object.
//
// Once the store operation completes successfully, the revision fields
// in the object will be updated to that returned by the store. These can
// either be reetrieved by one of the GetRevisionXxx() call or used for
// subsequent conditional operaions such as a conditional Update() call.
//
func (p *Pdu) Create(ctx context.Context) (int64, error) {

	if p.details == nil {
		return store.RevisionInvalid, errors.ErrDetailsNotAvailable("pdu")
	}

	if p.ports == nil {
		return store.RevisionInvalid, errors.ErrPortsNotAvailable("pdu")
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
	case errors.ErrStoreAlreadyExists(p.KeyIndexEntry): return store.RevisionInvalid, errors.ErrPduAlreadyExists{Region: p.Region, Zone: p.Zone, Rack: p.Rack, Pdu: p.ID}
	case errors.ErrStoreAlreadyExists(p.Key):           return store.RevisionInvalid, errors.ErrPduAlreadyExists{Region: p.Region, Zone: p.Zone, Rack: p.Rack, Pdu: p.ID}
	default:                                            return store.RevisionInvalid, err
	}

	p.record         = record
	p.revision       = rev
	p.revisionRecord = rev
	p.revisionStore  = rev

	return p.revision, nil
}

// Read is used to load a record from the underlying store to populate the
// fields in the object and determine the revision values associated with
// that record.
//
// Once the Read() has completed successfully the details and other
// information for the object can be retrieved by any of the GetXxx() methods
// for that obect.
//
func (p *Pdu) Read(ctx context.Context) (int64, error) {

	v, rev, err := p.Store.Read(ctx, store.KeyRootInventory, p.Key)

	if err == errors.ErrStoreKeyNotFound(p.Key) {
		return store.RevisionInvalid, errors.ErrPduNotFound{Region: p.Region, Zone: p.Zone, Rack: p.Rack, Pdu: p.ID}
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

// Update is used to persist the information in the fields of the object to
// a record in the underlying store. The update can be either unconditional
// by setting the unconditional parameter to true, or conditional based on
// the revision of the object compared to the revision of the associated
// record in the underlying store.
//
// Once the store operation completes successfully, the revision information
// in the object is updated with that returned from the store.
//
// Update() has no effect on the index information for the object.
//
func (p *Pdu) Update(ctx context.Context, unconditional bool) (int64, error) {

	var revUpdate = p.revisionRecord

	if unconditional == true {
		revUpdate = store.RevisionInvalid
	}

	if p.details == nil {
		return store.RevisionInvalid, errors.ErrDetailsNotAvailable("pdu")
	}

	if p.ports == nil {
		return store.RevisionInvalid, errors.ErrPortsNotAvailable("pdu")
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

	if err == errors.ErrStoreKeyNotFound(p.Key) {
		return store.RevisionInvalid, errors.ErrPduNotFound{Region: p.Region, Zone: p.Zone, Rack: p.Rack, Pdu: p.ID}
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

// Delete is used to remove the persisted copy of the object from the
// store along with any index information needed to navigate to or
// through that object. The delete can be either unconditional by
// setting the unconditional parameter to true, or conditional based
// on the revision of the object compared to the revision of the
// associated record in the underlying store.
//
// Deleting the record from the underlying store  has no effect on the
// values held in the fields of the object other than updating the
// revision information using the information returned by the store
// operation.
//
func (p *Pdu) Delete(ctx context.Context, unconditional bool) (int64, error) {

	var revDelete = p.revisionRecord

	if unconditional == true {
		revDelete = store.RevisionInvalid
	}

	rev, err := p.Store.Delete(ctx, store.KeyRootInventory, p.Key, revDelete)

	if err == errors.ErrStoreKeyNotFound(p.Key) {
		return store.RevisionInvalid, errors.ErrPduNotFound{Region: p.Region, Zone: p.Zone, Rack: p.Rack, Pdu: p.ID}
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

// Tor is a structure representing a tor object. This object can be used
// to operate on the associated tor records in the underlying store. The
// object can hold information fetched from the underlying store, or as
// a staging area in preparation for updates to the store.
//
// Tor is a specialization of a child object for a rack parent.
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

// SetDetails is used to attach some attribute information to the object.
//
// The attribute information is not persisted to the store until an Update()
// call is made.
//
// The current revision of the tor object is reset
//
func (t *Tor) SetDetails(ctx context.Context, details *pb.TorDetails) {
	t.details  = details
	t.revision = store.RevisionInvalid
}

// GetDetails is used to extract the attribute information from the object. The
// attribute information must have been previously read from the store (see
// the Read() method) or attached via a SetDetails() call.
//
// May return nil if there are no attributes currently held in the object.
//
func (t *Tor) GetDetails(ctx context.Context) *pb.TorDetails {
	return 	t.details
}

// GetRevision returns the revision of the details field within the object. This
// will be either the revision of the object in the store after a Create() or
// Read() call or be store.RevisionInvalid if the details have been set or no
// Create() or Read() call has been executed.
//
func (t *Tor) GetRevision(ctx context.Context) int64 {
	return t.revision
}

// GetRevisionRecord returns the revision of the underlying store object as 
// determined at the time of the last Create() or Read() for the object. The
// record revision is not reset by a SetDetails() call and is used when
// performing either a conditional update or conditional delete using the
// object.
//
func (t *Tor) GetRevisionRecord(ctx context.Context) int64 {
	return t.revisionRecord
}

// GetRevisionStore returns the revison of the underlying store ifself as 
// determined at the time of the last Read() for the object. The
// store revision is not reset by a SetDetails() call and is provided 
// for information only.
//
func (t *Tor) GetRevisionStore(ctx context.Context) int64 {
	return t.revisionStore
}

// SetPorts is used to attach some network port information to the object.
//
// The port information is not persisted to the store until an Update()
// call is made.
//
// The current revision of the tor object is reset
//
func (t *Tor) SetPorts(ctx context.Context, ports *map[int64]*pb.NetworkPort) {
	t.ports    = ports
	t.revision = store.RevisionInvalid
}

// GetPorts is used to extract the network port information from the object.
// The port information must have been previously read from the store (see
// the Read() method) or attached via a SetPorts() call.
//
// May return nil if there are no network port information currently held
// in the object.
//
func (t *Tor) GetPorts(ctx context.Context) *map[int64]*pb.NetworkPort {
	return t.ports
}

// Create is used to create a record in the underlying store for the
// object along with the associated index information.
//
// The underlying store record will contain the information currently
// held in the object.
//
// Once the store operation completes successfully, the revision fields
// in the object will be updated to that returned by the store. These can
// either be reetrieved by one of the GetRevisionXxx() call or used for
// subsequent conditional operaions such as a conditional Update() call.
//
func (t *Tor) Create(ctx context.Context) (int64, error) {

	if t.details == nil {
		return store.RevisionInvalid, errors.ErrDetailsNotAvailable("tor")
	}

	if t.ports == nil {
		return store.RevisionInvalid, errors.ErrPortsNotAvailable("tor")
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
	case errors.ErrStoreAlreadyExists(t.KeyIndexEntry): return store.RevisionInvalid, errors.ErrTorAlreadyExists{Region: t.Region, Zone: t.Zone, Rack: t.Rack, Tor: t.ID}
	case errors.ErrStoreAlreadyExists(t.Key):           return store.RevisionInvalid, errors.ErrTorAlreadyExists{Region: t.Region, Zone: t.Zone, Rack: t.Rack, Tor: t.ID}
	default:                                            return store.RevisionInvalid, err
	}

	t.record         = record
	t.revision       = rev
	t.revisionRecord = rev
	t.revisionStore  = rev

	return t.revision, nil
}

// Read is used to load a record from the underlying store to populate the
// fields in the object and determine the revision values associated with
// that record.
//
// Once the Read() has completed successfully the details and other
// information for the object can be retrieved by any of the GetXxx() methods
// for that obect.
//
func (t *Tor) Read(ctx context.Context) (int64, error) {

	v, rev, err := t.Store.Read(ctx, store.KeyRootInventory, t.Key)

	if err == errors.ErrStoreKeyNotFound(t.Key) {
		return store.RevisionInvalid, errors.ErrTorNotFound{Region: t.Region, Zone: t.Zone, Rack: t.Rack, Tor: t.ID}
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

// Update is used to persist the information in the fields of the object to
// a record in the underlying store. The update can be either unconditional
// by setting the unconditional parameter to true, or conditional based on
// the revision of the object compared to the revision of the associated
// record in the underlying store.
//
// Once the store operation completes successfully, the revision information
// in the object is updated with that returned from the store.
//
// Update() has no effect on the index information for the object.
//
func (t *Tor) Update(ctx context.Context, unconditional bool) (int64, error) {

	var revUpdate = t.revisionRecord

	if unconditional == true {
		revUpdate = store.RevisionInvalid
	}

	if t.details == nil {
		return store.RevisionInvalid, errors.ErrDetailsNotAvailable("tor")
	}

	if t.ports == nil {
		return store.RevisionInvalid, errors.ErrPortsNotAvailable("tor")
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

	if err == errors.ErrStoreKeyNotFound(t.Key) {
		return store.RevisionInvalid, errors.ErrTorNotFound{Region: t.Region, Zone: t.Zone, Rack: t.Rack, Tor: t.ID}
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

// Delete is used to remove the persisted copy of the object from the
// store along with any index information needed to navigate to or
// through that object. The delete can be either unconditional by
// setting the unconditional parameter to true, or conditional based
// on the revision of the object compared to the revision of the
// associated record in the underlying store.
//
// Deleting the record from the underlying store  has no effect on the
// values held in the fields of the object other than updating the
// revision information using the information returned by the store
// operation.
//
func (t *Tor) Delete(ctx context.Context, unconditional bool) (int64, error) {

	var revDelete = t.revisionRecord

	if unconditional == true {
		revDelete = store.RevisionInvalid
	}

	rev, err := t.Store.Delete(ctx, store.KeyRootInventory, t.Key, revDelete)

	if err == errors.ErrStoreKeyNotFound(t.Key) {
		return store.RevisionInvalid, errors.ErrTorNotFound{Region: t.Region, Zone: t.Zone, Rack: t.Rack, Tor: t.ID}
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

// Blade is a structure representing a blade object. This object can be
// used to operate on the associated blade records in the underlying
// store. The object can hold information fetched from the underlying
// store, or as a staging area in preparation for updates to the store.
//
// Blade is a specialization of a child object for a rack parent.
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

// SetDetails is used to attach some attribute information to the object.
//
// The attribute information is not persisted to the store until an Update()
// call is made.
//
// The current revision of the blade object is reset
//
func (b *Blade) SetDetails(ctx context.Context, details *pb.BladeDetails) {
	b.details  = details
	b.revision = store.RevisionInvalid
}

// GetDetails is used to extract the attribute information from the object. The
// attribute information must have been previously read from the store (see
// the Read() method) or attached via a SetDetails() call.
//
// May return nil if there are no attributes currently held in the object.
//
func (b *Blade) GetDetails(ctx context.Context) *pb.BladeDetails {
	return 	b.details
}

// GetRevision returns the revision of the details field within the object. This
// will be either the revision of the object in the store after a Create() or
// Read() call or be store.RevisionInvalid if the details have been set or no
// Create() or Read() call has been executed.
//
func (b *Blade) GetRevision(ctx context.Context) int64 {
	return b.revision
}

// GetRevisionRecord returns the revision of the underlying store object as 
// determined at the time of the last Create() or Read() for the object. The
// record revision is not reset by a SetDetails() call and is used when
// performing either a conditional update or conditional delete using the
// object.
//
func (b *Blade) GetRevisionRecord(ctx context.Context) int64 {
	return b.revisionRecord
}

// GetRevisionStore returns the revison of the underlying store ifself as 
// determined at the time of the last Read() for the object. The
// store revision is not reset by a SetDetails() call and is provided 
// for information only.
//
func (b *Blade) GetRevisionStore(ctx context.Context) int64 {
	return b.revisionStore
}

// SetCapacity is used to attach some capacity information to the object.
//
// The capacity information is not persisted to the store until an Update()
// call is made.
//
// The current revision of the blade object is reset
//
func (b *Blade) SetCapacity(ctx context.Context, capacity *pb.BladeCapacity) {
	b.capacity = capacity
	b.revision = store.RevisionInvalid
}

// SetBootInfo is used to attach some boot information to the object.
//
// The boot information is not persisted to the store until an Update()
// call is made.
//
// The current revision of the blade object is reset
//
func (b *Blade) SetBootInfo(ctx context.Context, bootOnPowerOn bool, bootInfo *pb.BladeBootInfo) {
	b.bootOnPowerOn = bootOnPowerOn
	b.bootInfo      = bootInfo
	b.revision      = store.RevisionInvalid
}

// GetCapacity is used to extract the capacity information from the object.
// The capacity information must have been previously read from the store (see
// the Read() method) or attached via a SetCapacity() call.
//
// May return nil if there are no capacity information currently held
// in the object.
//
func (b *Blade) GetCapacity(ctx context.Context) *pb.BladeCapacity {
	return b.capacity
}

// GetBootInfo is used to extract the boot information from the object.
// The boot information must have been previously read from the store (see
// the Read() method) or attached via a SetBootInfo() call.
//
// May return nil if there are no boot information currently held
// in the object.
//
func (b *Blade) GetBootInfo(ctx context.Context) (bool, *pb.BladeBootInfo) {
	return b.bootOnPowerOn, b.bootInfo
}

// Create is used to create a record in the underlying store for the
// object along with the associated index information.
//
// The underlying store record will contain the information currently
// held in the object.
//
// Once the store operation completes successfully, the revision fields
// in the object will be updated to that returned by the store. These can
// either be reetrieved by one of the GetRevisionXxx() call or used for
// subsequent conditional operaions such as a conditional Update() call.
//
func (b *Blade) Create(ctx context.Context)  (int64, error)  {

	if b.details == nil {
		return store.RevisionInvalid, errors.ErrDetailsNotAvailable("blade")
	}

	if b.capacity == nil {
		return store.RevisionInvalid, errors.ErrCapacityNotAvailable("blade")
	}

	if b.bootInfo == nil {
		return store.RevisionInvalid, errors.ErrBootInfoNotAvailable("blade")
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
	case errors.ErrStoreAlreadyExists(b.KeyIndexEntry): return store.RevisionInvalid, errors.ErrBladeAlreadyExists{Region: b.Region, Zone: b.Zone, Rack: b.Rack, Blade: b.ID}
	case errors.ErrStoreAlreadyExists(b.Key):           return store.RevisionInvalid, errors.ErrBladeAlreadyExists{Region: b.Region, Zone: b.Zone, Rack: b.Rack, Blade: b.ID}
	default:                                            return store.RevisionInvalid, err
	}

	b.record         = record
	b.revision       = rev
	b.revisionRecord = rev
	b.revisionStore  = rev

	return b.revision, nil
}

// Read is used to load a record from the underlying store to populate the
// fields in the object and determine the revision values associated with
// that record.
//
// Once the Read() has completed successfully the details and other
// information for the object can be retrieved by any of the GetXxx() methods
// for that obect.
//
func (b *Blade) Read(ctx context.Context) (int64, error) {

	v, rev, err := b.Store.Read(ctx, store.KeyRootInventory, b.Key)

	if err == errors.ErrStoreKeyNotFound(b.Key) {
		return store.RevisionInvalid, errors.ErrBladeNotFound{Region: b.Region, Zone: b.Zone, Rack: b.Rack, Blade: b.ID}
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

// Update is used to persist the information in the fields of the object to
// a record in the underlying store. The update can be either unconditional
// by setting the unconditional parameter to true, or conditional based on
// the revision of the object compared to the revision of the associated
// record in the underlying store.
//
// Once the store operation completes successfully, the revision information
// in the object is updated with that returned from the store.
//
// Update() has no effect on the index information for the object.
//
func (b *Blade) Update(ctx context.Context, unconditional bool) (int64, error) {

	var revUpdate = b.revisionRecord

	if unconditional == true {
		revUpdate = store.RevisionInvalid
	}

	if b.details == nil {
		return store.RevisionInvalid, errors.ErrDetailsNotAvailable("blade")
	}

	if b.capacity == nil {
		return store.RevisionInvalid, errors.ErrCapacityNotAvailable("blade")
	}

	if b.bootInfo == nil {
		return store.RevisionInvalid, errors.ErrBootInfoNotAvailable("blade")
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

	if err == errors.ErrStoreKeyNotFound(b.Key) {
		return store.RevisionInvalid, errors.ErrBladeNotFound{Region: b.Region, Zone: b.Zone, Rack: b.Rack, Blade: b.ID}
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

// Delete is used to remove the persisted copy of the object from the
// store along with any index information needed to navigate to or
// through that object. The delete can be either unconditional by
// setting the unconditional parameter to true, or conditional based
// on the revision of the object compared to the revision of the
// associated record in the underlying store.
//
// Deleting the record from the underlying store  has no effect on the
// values held in the fields of the object other than updating the
// revision information using the information returned by the store
// operation.
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

	if err == errors.ErrStoreKeyNotFound(b.Key) {
		return store.RevisionInvalid, errors.ErrBladeNotFound{Region: b.Region, Zone: b.Zone, Rack: b.Rack, Blade: b.ID}
	}

	if err != nil {
		return store.RevisionInvalid, err
	}

	_, err = b.Store.Delete(ctx, store.KeyRootInventory, b.KeyIndexEntry, store.RevisionInvalid)

	if err == errors.ErrStoreKeyNotFound(b.KeyIndexEntry) {
		return store.RevisionInvalid, errors.ErrBladeNotFound{Region: b.Region, Zone: b.Zone, Rack: b.Rack, Blade: b.ID}
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
