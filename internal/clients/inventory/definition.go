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

type revisionInfo struct {

	revision       int64
	revisionRecord int64
	revisionStore  int64
}

// GetRevision returns the revision of the details field within the object.
// This will be either the revision of the object in the store after a
// Create(), Read() or Update() call or be store.RevisionInvalid if the
// details have been set or no Create(), Read() or Update() call has been
// executed.
//
func (r *revisionInfo) GetRevision(ctx context.Context) int64 {
	return r.revision
}

// GetRevisionRecord returns the revision of the underlying store object as 
// determined at the time of the last Create(), Read() or Update() for the
// object. The record revision is not reset by a SetDetails() call and is
// used when performing either a conditional update or conditional delete
// using the object.
//
func (r *revisionInfo) GetRevisionRecord(ctx context.Context) int64 {
	return r.revisionRecord
}

// GetRevisionStore returns the revison of the underlying store ifself as 
// determined at the time of the last Create() Read() for the object. The
// store revision is not reset by a SetDetails() call and is provided 
// for information only.
//
func (r *revisionInfo) GetRevisionStore(ctx context.Context) int64 {
	return r.revisionStore
}

// GetRevisionForRequest returns the appropriate revision for the update
// for either a conditional update based upon the revision of the most
// recently read record, or an unconditional update.
//
func (r *revisionInfo) GetRevisionForRequest(ctx context.Context, unconditional bool) int64 {

	if unconditional == true {
		return store.RevisionInvalid
	}

	return r.revisionRecord
}

// resetRevision resets the revision for the details field within the object.
// Subsequent calls to GetRevision() will return store.RevisionInvalid until
// a successful call is made to one of the routines which invoke the store
//
func (r *revisionInfo) resetRevision(ctx context.Context) int64 {
	r.revision = store.RevisionInvalid

	return store.RevisionInvalid
}

// updateRevision is used to set/update the current revision information 
// as part of a successful invokation of a store routine.
//
func (r *revisionInfo) updateRevisionInfo(ctx context.Context, rev int64) int64 {
	r.revision       = rev
	r.revisionRecord = rev
	r.revisionStore  = rev

	return rev
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

	revisionInfo

	details        *pb.RootDetails
}

// SetDetails is used to attach some attribute information to the object.
//
// For a Root object, the information is not persisted.
//
// The current revision of the region object is reset
//
func (r *Root) SetDetails(ctx context.Context, details *pb.RootDetails) {
	r.details  = details
	r.resetRevision(ctx)
}

// GetDetails is used to extract the attribute information from the object.
//
// As the Root object is not persisted, the attribute information will either
// be the initialisation value, or whatever was last set using SetDetails()
//
func (r *Root) GetDetails(ctx context.Context) *pb.RootDetails {
	return 	r.details
}

// Create is not used for a Root object as there is no persistence for this
// object.

//


func (r *Root) Create(ctx context.Context) (int64, error) {
	return store.RevisionInvalid, errors.ErrFunctionNotAvailable
}

// Read is not used for a Root object as there is no persistence for this
// object.
//
func (r *Root) Read(ctx context.Context) (int64, error) {
	return 	store.RevisionInvalid, errors.ErrFunctionNotAvailable
}

// Update is not used for a Root object as there is no persistence for this
// object.
//
func (r *Root) Update(ctx context.Context, unconditional bool) (int64, error) {
	return store.RevisionInvalid, errors.ErrFunctionNotAvailable
}

// Delete is not used for a Root object as there is no persistence for this
// object.
//
func (r *Root) Delete(ctx context.Context, unconditional bool) (int64, error) {
	return 	store.RevisionInvalid, errors.ErrFunctionNotAvailable
}

// NewChild creates a new region child object within the current
// root using the supplied name. This new object can be used for
// further navigation or for actions involving operations against the
// associated record in the underlying store.
//
// No information is fetched from the underlying store so the attribute
// and revisions fields within the oject are not valid.
//
func (r *Root) NewChild(ctx context.Context, name string) (*Region, error) {

	return NewRegion(ctx, r.Store, r.Table, name)


}

// ListChildren uses the current object to discover the names of all the
// region child objects in the underlying store for the the current root
// object, The elements of the returned list can be used in subsequent
// NewChild() calls to create new region objects.
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

// FetchChildren is used to discover all the child region objects in the
// underlying store for the current root object and to generate a new
// region object for each of those children. It is a convenience wrapper
// around ListChildren() followed by a NewChild() on each name discovered.
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

	revisionInfo

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
	r.resetRevision(ctx)
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

	if err = r.mapErrStoreAlreadyExists(err); err != nil {
		return store.RevisionInvalid, err
	}

	r.record = record

	return r.updateRevisionInfo(ctx, rev), nil
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

	if err = r.mapErrStoreKeyNotFound(err); err != nil {
		return store.RevisionInvalid, err
	}

	record := &pb.StoreRecordDefinitionRegion{}

	if err = store.Decode(*v, record); err != nil {
		return store.RevisionInvalid, err
	}

	r.details = record.Details
	r.record  = record

	return r.updateRevisionInfo(ctx, rev), nil
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

	rev, err := r.Store.Update(
		ctx,
		store.KeyRootInventory,
		r.Key,
		r.GetRevisionForRequest(ctx, unconditional),
		v)

	if err = r.mapErrStoreKeyNotFound(err); err != nil {
		return store.RevisionInvalid, err
	}

	r.record = record

	return r.updateRevisionInfo(ctx, rev), nil
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

	// TODO - use delete multiple to remove object and index?
	//
	rev, err := r.Store.Delete(
		ctx,
		store.KeyRootInventory,
		r.Key,
		r.GetRevisionForRequest(ctx, unconditional))

	if err = r.mapErrStoreKeyNotFound(err); err != nil {
		return store.RevisionInvalid, err
	}

	r.record = nil

	return r.updateRevisionInfo(ctx, rev), nil
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

	return NewZone(ctx, r.Store, r.Table, r.Region, name)
}

// ListChildren uses the current object to discover the names of all the
// zone child objects in the underlying store for the the current region
// object, The elements of the returned list can be used in subsequent
// NewChild() calls to create new zone objects.
//
func (r *Region) ListChildren(ctx context.Context) (int64, []string, error) {
	
	records, rev, err := r.Store.List(ctx, store.KeyRootInventory, r.KeyChildIndex)

	if err = r.mapErrStoreIndexNotFound(err); err != nil {
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

func (r *Region) mapErrStoreKeyNotFound(err error) error {
	if err == errors.ErrStoreKeyNotFound(r.Key) {
		return errors.ErrRegionNotFound{Region: r.Region}
	} 
	
	return err
}

func (r *Region) mapErrStoreIndexNotFound(err error) error {
	if err == errors.ErrStoreIndexNotFound(r.KeyChildIndex) {
		return errors.ErrIndexNotFound(r.KeyChildIndex)
	}

	return err
}

func (r *Region) mapErrStoreAlreadyExists(err error) error {
	if err == errors.ErrStoreAlreadyExists(r.KeyIndexEntry) || err == errors.ErrStoreAlreadyExists(r.Key) {
		return errors.ErrRegionAlreadyExists{Region: r.Region}
	}
	
	return err
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

	revisionInfo

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
	z.resetRevision(ctx)
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

	if err = z.mapErrStoreAlreadyExists(err); err != nil {
		return store.RevisionInvalid, err
	}

	z.record = record

	return z.updateRevisionInfo(ctx, rev), nil
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

	if err = z.mapErrStoreKeyNotFound(err); err != nil {
		return store.RevisionInvalid, err
	}

	record := &pb.StoreRecordDefinitionZone{}

	if err = store.Decode(*v, record); err != nil {
		return store.RevisionInvalid, err
	}

	z.details = record.Details
	z.record  = record

	return z.updateRevisionInfo(ctx, rev), nil
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

	rev, err := z.Store.Update(
		ctx,
		store.KeyRootInventory,
		z.Key,
		z.GetRevisionForRequest(ctx, unconditional),
		v)

	if err = z.mapErrStoreKeyNotFound(err); err != nil {
		return store.RevisionInvalid, err
	}

	z.record = record

	return z.updateRevisionInfo(ctx, rev), nil
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

	rev, err := z.Store.Delete(
		ctx,
		store.KeyRootInventory,
		z.Key,
		z.GetRevisionForRequest(ctx, unconditional))

	if err = z.mapErrStoreKeyNotFound(err); err != nil {
		return store.RevisionInvalid, err
	}

	z.record = nil

	return z.updateRevisionInfo(ctx, rev), nil
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

	return NewRack(ctx, z.Store, z.Table, z.Region, z.Zone, name)
}

// ListChildren uses the current object to discover the names of all the
// rack child objects in the underlying store for the the current zone
// object, The elements of the returned list can be used in subsequent
// NewChild() calls to create new rack objects.
//
func (z *Zone) ListChildren(ctx context.Context) (int64, []string, error) {
	
	records, rev, err := z.Store.List(ctx, store.KeyRootInventory, z.KeyChildIndex)

	if err = z.mapErrStoreIndexNotFound(err); err != nil {
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

func (z *Zone) mapErrStoreKeyNotFound(err error) error {
	if err == errors.ErrStoreKeyNotFound(z.Key) {
		return errors.ErrZoneNotFound{Region: z.Region, Zone: z.Zone}
	} 
	
	return err
}

func (z *Zone) mapErrStoreIndexNotFound(err error) error {
	if err == errors.ErrStoreIndexNotFound(z.KeyChildIndex) {
		return errors.ErrIndexNotFound(z.KeyChildIndex)
	}

	return err
}

func (z *Zone) mapErrStoreAlreadyExists(err error) error {
	if err == errors.ErrStoreAlreadyExists(z.KeyIndexEntry) || err == errors.ErrStoreAlreadyExists(z.Key) {
		return errors.ErrZoneAlreadyExists{Region: z.Region, Zone: z.Zone}
	}
	
	return err
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

	revisionInfo

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
	r.resetRevision(ctx)
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

	if err = r.mapErrStoreAlreadyExists(err); err != nil {
		return store.RevisionInvalid, err
	}

	r.record = record

	return r.updateRevisionInfo(ctx, rev), nil
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

	if err = r.mapErrStoreKeyNotFound(err); err != nil {
		return store.RevisionInvalid, err
	}

	record := &pb.StoreRecordDefinitionRack{}

	if err = store.Decode(*v, record); err != nil {
		return store.RevisionInvalid, err
	}

	r.details = record.Details
	r.record  = record

	return r.updateRevisionInfo(ctx, rev), nil
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

	rev, err := r.Store.Update(
		ctx,
		store.KeyRootInventory,
		r.Key,
		r.GetRevisionForRequest(ctx, unconditional),
		v)

	if err = r.mapErrStoreKeyNotFound(err); err != nil {
		return store.RevisionInvalid, err
	}

	r.record = record

	return r.updateRevisionInfo(ctx, rev), nil
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

	rev, err := r.Store.Delete(
		ctx,
		store.KeyRootInventory,
		r.Key,
		r.GetRevisionForRequest(ctx, unconditional))

	if err = r.mapErrStoreKeyNotFound(err); err != nil {
		return store.RevisionInvalid, err
	}

	r.record = nil

	return r.updateRevisionInfo(ctx, rev), nil
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

	return NewPdu(ctx, r.Store, r.Table, r.Region, r.Zone, r.Rack, ID)
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

	return NewTor(ctx, r.Store, r.Table, r.Region, r.Zone, r.Rack, ID)
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

	return NewBlade(ctx, r.Store, r.Table, r.Region, r.Zone, r.Rack, ID)
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

	if err = r.mapErrStoreIndexNotFound(err, r.KeyIndexPdu); err != nil {
		return store.RevisionInvalid, nil, err
	}

	names := make([]int64, 0, len(*records))

	for k, v := range *records {
	
		name := strings.TrimPrefix(k, r.KeyIndexPdu)

		// Verify that the "index" part of the name is numeric
		//
		intName, err := strconv.ParseInt(name, 10, 64)

		if err != nil {
			return store.RevisionInvalid, nil, errors.ErrPduIndexInvalid{Region: r.Region, Zone: r.Zone, Rack: r.Rack, Pdu: name}
		}

		intValue, err := strconv.ParseInt(v.Value, 10, 64)

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

	if err = r.mapErrStoreIndexNotFound(err, r.KeyIndexTor); err != nil {
		return store.RevisionInvalid, nil, err
	}

	names := make([]int64, 0, len(*records))

	for k, v := range *records {
	
		name := strings.TrimPrefix(k, r.KeyIndexTor)

		// Verify that the "index" part of the name is numeric
		//
		intName, err := strconv.ParseInt(name, 10, 64)

		if err != nil {
			return store.RevisionInvalid, nil, errors.ErrTorIndexInvalid{Region: r.Region, Zone: r.Zone, Rack: r.Rack, Tor: name}
		}

		intValue, err := strconv.ParseInt(v.Value, 10, 64)

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

	if err = r.mapErrStoreIndexNotFound(err, r.KeyIndexBlade); err != nil {
		return store.RevisionInvalid, nil, err
	}

	names := make([]int64, 0, len(*records))

	for k, v := range *records {
	
		name := strings.TrimPrefix(k, r.KeyIndexBlade)

		// Verify that the "index" part of the name is numeric
		//
		intName, err := strconv.ParseInt(name, 10, 64)

		if err != nil {
			return store.RevisionInvalid, nil, errors.ErrBladeIndexInvalid{Region: r.Table, Zone: r.Zone, Rack: r.Rack, Blade: name}
		}

		intValue, err := strconv.ParseInt(v.Value, 10, 64)

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

func (r *Rack) mapErrStoreKeyNotFound(err error) error {
	if err == errors.ErrStoreKeyNotFound(r.Key) {
		return errors.ErrRackNotFound{Region: r.Region, Zone: r.Zone, Rack: r.Rack}
	} 
	
	return err
}

func (r *Rack) mapErrStoreIndexNotFound(err error, index string) error {
	if err == errors.ErrStoreIndexNotFound(index) {
		return errors.ErrIndexNotFound(index)
	}

	return err
}

func (r *Rack) mapErrStoreAlreadyExists(err error) error {
	if err == errors.ErrStoreAlreadyExists(r.KeyIndexEntry) || err == errors.ErrStoreAlreadyExists(r.Key) {
		return errors.ErrRackAlreadyExists{Region: r.Region, Zone: r.Zone, Rack: r.Rack}
	}
	
	return err
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

	revisionInfo

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
	p.resetRevision(ctx)
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

// SetPorts is used to attach some power port information to the object.
//
// The port information is not persisted to the store until an Update()
// call is made.
//
// The current revision of the pdu object is reset
//
func (p *Pdu) SetPorts(ctx context.Context, ports *map[int64]*pb.PowerPort) {
	p.ports    = ports
	p.resetRevision(ctx)
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

	if err = p.mapErrStoreAlreadyExists(err); err != nil {
		return store.RevisionInvalid, err
	}

	p.record = record

	return p.updateRevisionInfo(ctx, rev), nil
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

	if err = p.mapErrStoreKeyNotFound(err); err != nil {
		return store.RevisionInvalid, err
	}

	record := &pb.StoreRecordDefinitionPdu{}

	if err = store.Decode(*v, record); err != nil {
		return store.RevisionInvalid, err
	}

	p.details = record.Details
	p.ports   = &record.Ports
	p.record  = record

	return p.updateRevisionInfo(ctx, rev), nil
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

	rev, err := p.Store.Update(
		ctx,
		store.KeyRootInventory,
		p.Key,
		p.GetRevisionForRequest(ctx, unconditional),
		v)

	if err = p.mapErrStoreKeyNotFound(err); err != nil {
		return store.RevisionInvalid, err
	}

	p.record = record

	return p.updateRevisionInfo(ctx, rev), nil
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

	rev, err := p.Store.Delete(
		ctx,
		store.KeyRootInventory,
		p.Key,
		p.GetRevisionForRequest(ctx, unconditional))

	if err = p.mapErrStoreKeyNotFound(err); err != nil {
		return store.RevisionInvalid, err
	}

	p.record = nil

	return p.updateRevisionInfo(ctx, rev), nil
}

func (p *Pdu) mapErrStoreKeyNotFound(err error) error {
	if err == errors.ErrStoreKeyNotFound(p.Key) {
		return errors.ErrPduNotFound{Region: p.Region, Zone: p.Zone, Rack: p.Rack, Pdu: p.ID}
	} 
	
	return err
}

func (p *Pdu) mapErrStoreAlreadyExists(err error) error {
	if err == errors.ErrStoreAlreadyExists(p.KeyIndexEntry) || err == errors.ErrStoreAlreadyExists(p.Key) {
		return errors.ErrPduAlreadyExists{Region: p.Region, Zone: p.Zone, Rack: p.Rack, Pdu: p.ID}
	}
	
	return err
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

	revisionInfo

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
	t.resetRevision(ctx)
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

// SetPorts is used to attach some network port information to the object.
//
// The port information is not persisted to the store until an Update()
// call is made.
//
// The current revision of the tor object is reset
//
func (t *Tor) SetPorts(ctx context.Context, ports *map[int64]*pb.NetworkPort) {
	t.ports    = ports
	t.resetRevision(ctx)
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

	if err = t.mapErrStoreAlreadyExists(err); err != nil {
		return store.RevisionInvalid, err
	}

	t.record = record

	return t.updateRevisionInfo(ctx, rev), nil
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

	if err = t.mapErrStoreKeyNotFound(err); err != nil {
		return store.RevisionInvalid, err
	}

	record := &pb.StoreRecordDefinitionTor{}

	if err = store.Decode(*v, record); err != nil {
		return store.RevisionInvalid, err
	}

	t.details = record.Details
	t.ports   = &record.Ports
	t.record  = record

	return t.updateRevisionInfo(ctx, rev), nil
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

	rev, err := t.Store.Update(
		ctx,
		store.KeyRootInventory,
		t.Key,
		t.GetRevisionForRequest(ctx, unconditional),
		v)

	if err = t.mapErrStoreKeyNotFound(err); err != nil {
		return store.RevisionInvalid, err
	}

	t.record = record

	return t.updateRevisionInfo(ctx, rev), nil
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

	rev, err := t.Store.Delete(
		ctx,
		store.KeyRootInventory,
		t.Key,
		t.GetRevisionForRequest(ctx, unconditional))

	if err = t.mapErrStoreKeyNotFound(err); err != nil {
		return store.RevisionInvalid, err
	}

	t.record = nil

	return t.updateRevisionInfo(ctx, rev), nil
}

func (t *Tor) mapErrStoreKeyNotFound(err error) error {
	if err == errors.ErrStoreKeyNotFound(t.Key) {
		return errors.ErrTorNotFound{Region: t.Region, Zone: t.Zone, Rack: t.Rack, Tor: t.ID}
	} 
	
	return err
}

func (t *Tor) mapErrStoreAlreadyExists(err error) error {
	if err == errors.ErrStoreAlreadyExists(t.KeyIndexEntry) || err == errors.ErrStoreAlreadyExists(t.Key) {
		return errors.ErrTorAlreadyExists{Region: t.Region, Zone: t.Zone, Rack: t.Rack, Tor: t.ID}
	}
	
	return err
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

	revisionInfo

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
	b.resetRevision(ctx)
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

// SetCapacity is used to attach some capacity information to the object.
//
// The capacity information is not persisted to the store until an Update()
// call is made.
//
// The current revision of the blade object is reset
//
func (b *Blade) SetCapacity(ctx context.Context, capacity *pb.BladeCapacity) {
	b.capacity = capacity
	b.resetRevision(ctx)
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
	b.resetRevision(ctx)
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

	if err = b.mapErrStoreAlreadyExists(err); err != nil {
		return store.RevisionInvalid, err
	}

	b.record = record

	return b.updateRevisionInfo(ctx, rev), nil
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

	if err = b.mapErrStoreKeyNotFound(err); err != nil {
		return store.RevisionInvalid, err
	}

	record := &pb.StoreRecordDefinitionBlade{}

	if err = store.Decode(*v, record); err != nil {
		return store.RevisionInvalid, err
	}

	b.details       = record.Details
	b.capacity      = record.Capacity
	b.bootInfo      = record.BootInfo
	b.bootOnPowerOn = record.BootOnPowerOn
	b.record        = record

	return b.updateRevisionInfo(ctx, rev), nil
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

	rev, err := b.Store.Update(
		ctx,
		store.KeyRootInventory,
		b.Key,
		b.GetRevisionForRequest(ctx, unconditional),
		v)

	if err = b.mapErrStoreKeyNotFound(err); err != nil {
		return store.RevisionInvalid, err
	}

	b.record = record

	return b.updateRevisionInfo(ctx, rev), nil
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

	// This should really use store.DeleteMultiple() except that doesn't exist
	// yet to allow multiple key,value pairs to removed as an atomic update.
	//
	// Oops.
	//
	// Once the DeleteMultiple() routine has been added, this should change to
	// make use of the new call.
	//
	rev, err := b.Store.Delete(
		ctx,
		store.KeyRootInventory,
		b.Key,
		b.GetRevisionForRequest(ctx, unconditional))

	if err == errors.ErrStoreKeyNotFound(b.Key) {
		return store.RevisionInvalid, errors.ErrBladeNotFound{Region: b.Region, Zone: b.Zone, Rack: b.Rack, Blade: b.ID}
	}

	if err != nil {
		return store.RevisionInvalid, err
	}

	_, err = b.Store.Delete(ctx, store.KeyRootInventory, b.KeyIndexEntry, store.RevisionInvalid)

	if err = b.mapErrStoreKeyNotFound(err); err != nil {
		return store.RevisionInvalid, err
	}

	b.record = nil

	return b.updateRevisionInfo(ctx, rev), nil
}

func (b *Blade) mapErrStoreKeyNotFound(err error) error {
	if err == errors.ErrStoreKeyNotFound(b.Key) {
		return errors.ErrBladeNotFound{Region: b.Region, Zone: b.Zone, Rack: b.Rack, Blade: b.ID}
	} 
	
	return err
}

func (b *Blade) mapErrStoreAlreadyExists(err error) error {
	if err == errors.ErrStoreAlreadyExists(b.KeyIndexEntry) || err == errors.ErrStoreAlreadyExists(b.Key) {
		return errors.ErrBladeAlreadyExists{Region: b.Region, Zone: b.Zone, Rack: b.Rack, Blade: b.ID}
	}
	
	return err
}
