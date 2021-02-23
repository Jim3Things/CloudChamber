// This module contain the structures and methods to operate on the persisted definition
// table within the inventory package.

package inventory

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	pb "github.com/Jim3Things/CloudChamber/simulation/pkg/protos/inventory"

	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/store"
	"github.com/Jim3Things/CloudChamber/simulation/pkg/errors"
)

// NewRoot returns a root object which acts as a well-known point in a namespace
// and which can be used to navigate the namespace for a given table.
//
// Valid tables are
//	- DefinitionTable
//	- ActualTable
//	- ObservedTable
//	- TargetTable
//
func (m *Inventory) NewRoot(table string) (*Root, error) {

	return newRoot(m.Store, table)
}

func newRoot(store *store.Store, table string) (*Root, error) {

	k, err := getKeyForIndexRegion(table)

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
func (m *Inventory) NewRegion(table string, region string) (*Region, error) {

	return newRegion(m.Store, table, region)
}

func newRegion(store *store.Store, table string, region string) (*Region, error) {

	keyIndex, err := getKeyForIndexZone(table, region)

	if err != nil {
		return nil, err
	}

	keyIndexEntry, err := getKeyForIndexEntryRegion(table, region)

	if nil != err {
		return nil, err
	}

	key, err := getKeyForRegion(table, region)

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
func (m *Inventory) NewZone(table string, region string, zone string) (*Zone, error) {
	return newZone(m.Store, table, region, zone)
}

func newZone(store *store.Store, table string, region string, zone string) (*Zone, error) {

	keyIndex, err := getKeyForIndexRack(table, region, zone)

	if err != nil {
		return nil, err
	}

	keyIndexEntry, err := getKeyForIndexEntryZone(table, region, zone)

	if nil != err {
		return nil, err
	}

	key, err := getKeyForZone(table, region, zone)

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
func (m *Inventory) NewRack(
	table string,
	region string,
	zone string,
	rack string) (*Rack, error) {

	return newRack(m.Store, table, region, zone, rack)
}

func newRack(
	store *store.Store,
	table string,
	region string,
	zone string,
	rack string) (*Rack, error) {

	keyIndexPdu, err := getKeyForIndexPdu(table, region, zone, rack)

	if err != nil {
		return nil, err
	}

	keyIndexTor, err := getKeyForIndexTor(table, region, zone, rack)

	if err != nil {
		return nil, err
	}

	keyIndexBlade, err := getKeyForIndexBlade(table, region, zone, rack)

	if err != nil {
		return nil, err
	}

	keyIndexEntry, err := getKeyForIndexEntryRack(table, region, zone, rack)

	if nil != err {
		return nil, err
	}

	key, err := getKeyForRack(table, region, zone, rack)

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
func (m *Inventory) NewPdu(
	table string,
	region string,
	zone string,
	rack string,
	id int64) (*Pdu, error) {

	return newPdu(m.Store, table, region, zone, rack, id)
}

func newPdu(
	store *store.Store,
	table string,
	region string,
	zone string,
	rack string,
	id int64) (*Pdu, error) {

	keyIndexEntry, err := getKeyForIndexEntryPdu(table, region, zone, rack, id)

	if nil != err {
		return nil, err
	}

	key, err := getKeyForPdu(table, region, zone, rack, id)

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
func (m *Inventory) NewTor(
	table string,
	region string,
	zone string,
	rack string,
	id int64) (*Tor, error) {

	return newTor(m.Store, table, region, zone, rack, id)
	}

func newTor(
	store *store.Store,
	table string,
	region string,
	zone string,
	rack string,
	id int64) (*Tor, error) {

	keyIndexEntry, err := getKeyForIndexEntryTor(table, region, zone, rack, id)

	if nil != err {
		return nil, err
	}

	key, err := getKeyForTor(table, region, zone, rack, id)

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
func (m *Inventory) NewBlade(
	table string,
	region string,
	zone string,
	rack string,
	id int64) (*Blade, error) {

	return newBlade(m.Store, table, region, zone, rack, id)
}

func newBlade(
	store *store.Store,
	table string,
	region string,
	zone string,
	rack string,
	id int64) (*Blade, error) {

	keyIndexEntry, err := getKeyForIndexEntryBlade(table, region, zone, rack, id)

	if nil != err {
		return nil, err
	}

	key, err := getKeyForBlade(table, region, zone, rack, id)

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
func (r *revisionInfo) GetRevision() int64 {
	return r.revision
}

// GetRevisionRecord returns the revision of the underlying store object as
// determined at the time of the last Create(), Read() or Update() for the
// object. The record revision is not reset by a SetDetails() call and is
// used when performing either a conditional update or conditional delete
// using the object.
//
func (r *revisionInfo) GetRevisionRecord() int64 {
	return r.revisionRecord
}

// GetRevisionStore returns the revision of the underlying store itself as
// determined at the time of the last Create() Read() for the object. The
// store revision is not reset by a SetDetails() call and is provided
// for information only.
//
func (r *revisionInfo) GetRevisionStore() int64 {
	return r.revisionStore
}

// GetRevisionForRequest returns the appropriate revision for the update
// for either a conditional update based upon the revision of the most
// recently read record, or an unconditional update.
//
func (r *revisionInfo) GetRevisionForRequest(unconditional bool) int64 {

	if unconditional == true {
		return store.RevisionInvalid
	}

	return r.revisionRecord
}

// resetRevision resets the revision for the details field within the object.
// Subsequent calls to GetRevision() will return store.RevisionInvalid until
// a successful call is made to one of the routines which invoke the store
//
func (r *revisionInfo) resetRevision() int64 {
	r.revision = store.RevisionInvalid

	return store.RevisionInvalid
}

// updateRevision is used to set/update the current revision information
// as part of a successful invocation of a store routine.
//
func (r *revisionInfo) updateRevisionInfo(rev int64) int64 {
	r.revision = rev
	r.revisionRecord = rev
	r.revisionStore = rev

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
	KeyChildIndex string
	Table         string

	revisionInfo

	details *pb.RootDetails
}

// SetDetails is used to attach some attribute information to the object.
//
// For a Root object, the information is not persisted.
//
// The current revision of the region object is reset
//
func (r *Root) SetDetails(details *pb.RootDetails) {
	r.details = details
	r.resetRevision()
}

// GetDetails is used to extract the attribute information from the object.
//
// As the Root object is not persisted, the attribute information will either
// be the initialisation value, or whatever was last set using SetDetails()
//
func (r *Root) GetDetails() *pb.RootDetails {
	return r.details
}

// Create is not used for a Root object as there is no persistence for this
// object.
//
func (r *Root) Create(_ context.Context) (int64, error) {
	return store.RevisionInvalid, errors.ErrFunctionNotAvailable
}

// Read is not used for a Root object as there is no persistence for this
// object.
//
func (r *Root) Read(_ context.Context) (int64, error) {
	return store.RevisionInvalid, errors.ErrFunctionNotAvailable
}

// Update is not used for a Root object as there is no persistence for this
// object.
//
func (r *Root) Update(_ context.Context, _ bool) (int64, error) {
	return store.RevisionInvalid, errors.ErrFunctionNotAvailable
}

// Delete is not used for a Root object as there is no persistence for this
// object.
//
func (r *Root) Delete(_ context.Context, _ bool) (int64, error) {
	return store.RevisionInvalid, errors.ErrFunctionNotAvailable
}

// NewChild creates a new region child object within the current
// root using the supplied name. This new object can be used for
// further navigation or for actions involving operations against the
// associated record in the underlying store.
//
// No information is fetched from the underlying store so the attribute
// and revisions fields within the object are not valid.
//
func (r *Root) NewChild(name string) (*Region, error) {

	return newRegion(r.Store, r.Table, name)
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
// around ListChildren() followed by a NewChild() and Read() on each name
// discovered.
//
func (r *Root) FetchChildren(ctx context.Context) (int64, *map[string]Region, error) {

	rev, names, err := r.ListChildren(ctx)

	if err != nil {
		return store.RevisionInvalid, nil, err
	}

	children := make(map[string]Region, len(names))

	for _, v := range names {

		child, err := r.NewChild(v)

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
	Store         *store.Store
	KeyChildIndex string
	KeyIndexEntry string
	Key           string
	Table         string
	Region        string

	revisionInfo

	details *pb.RegionDetails
	record  *pb.Store_RecordDefinition_Region
}

// SetDetails is used to attach some attribute information to the object.
//
// The attribute information is not persisted to the store until an Update()
// call is made.
//
// The current revision of the region object is reset
//
func (r *Region) SetDetails(details *pb.RegionDetails) {
	r.details = details
	r.resetRevision()
}

// GetDetails is used to extract the attribute information from the object. The
// attribute information must have been previously read from the store (see
// the Read() method) or attached via a SetDetails() call.
//
// May return nil if there are no attributes currently held in the object.
//
func (r *Region) GetDetails() *pb.RegionDetails {
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
// either be retrieved by one of the GetRevisionXxx() call or used for
// subsequent conditional operations such as a conditional Update() call.
//
func (r *Region) Create(ctx context.Context) (int64, error) {

	if r.details == nil {
		return store.RevisionInvalid, errors.ErrDetailsNotAvailable("region")
	}

	record := &pb.Store_RecordDefinition_Region{
		Details: r.details,
	}

	v, err := store.Encode(record)

	if err != nil {
		return store.RevisionInvalid, err
	}

	// Create the child and its index as an atomic pair.
	//
	keySet := &map[string]string{
		r.KeyIndexEntry: r.Region,
		r.Key:           v,
	}

	rev, err := r.Store.CreateMultiple(ctx, store.KeyRootInventory, keySet)

	if err = r.mapErrStoreValue(err); err != nil {
		return store.RevisionInvalid, err
	}

	r.record = record

	return r.updateRevisionInfo(rev), nil
}

// Read is used to load a record from the underlying store to populate the
// fields in the object and determine the revision values associated with
// that record.
//
// Once the Read() has completed successfully the details and other
// information for the object can be retrieved by any of the GetXxx() methods
// for that object.
//
func (r *Region) Read(ctx context.Context) (int64, error) {

	v, rev, err := r.Store.Read(ctx, store.KeyRootInventory, r.Key)

	if err = r.mapErrStoreValue(err); err != nil {
		return store.RevisionInvalid, err
	}

	record := &pb.Store_RecordDefinition_Region{}

	if err = store.Decode(*v, record); err != nil {
		return store.RevisionInvalid, err
	}

	r.details = record.Details
	r.record = record

	return r.updateRevisionInfo(rev), nil
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

	record := &pb.Store_RecordDefinition_Region{
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
		r.GetRevisionForRequest(unconditional),
		v)

	if err = r.mapErrStoreValue(err); err != nil {
		return store.RevisionInvalid, err
	}

	r.record = record

	return r.updateRevisionInfo(rev), nil
}

// Delete is used to remove the persisted copy of the object from the
// store along with any index information needed to navigate to or
// through that object. The delete can be either unconditional by
// setting the unconditional parameter to true, or conditional based
// on the revision of the object compared to the revision of the
// associated record in the underlying store.
//
// Deleting the record from the underlying store has no effect on the
// values held in the fields of the object other than updating the
// revision information using the information returned by the store
// operation.
//
func (r *Region) Delete(ctx context.Context, unconditional bool) (int64, error) {

	// Delete the record and its index as an atomic pair.
	//
	keySet := &map[string]int64{
		r.KeyIndexEntry: store.RevisionInvalid,
		r.Key:           r.GetRevisionForRequest(unconditional),
	}

	rev, err := r.Store.DeleteMultiple(ctx, store.KeyRootInventory, keySet)

	if err = r.mapErrStoreValue(err); err != nil {
		return store.RevisionInvalid, err
	}

	r.record = nil

	return r.updateRevisionInfo(rev), nil
}

// NewChild creates a new child object for the zone within the current
// region using the supplied name. This new object can be used for
// further navigation or for actions involving operations against the
// associated record in the underlying store.
//
// No information is fetched from the underlying store so the attribute
// and revisions fields within the object are not valid.
//
func (r *Region) NewChild(name string) (*Zone, error) {

	return newZone(r.Store, r.Table, r.Region, name)
}

// ListChildren uses the current object to discover the names of all the
// zone child objects in the underlying store for the the current region
// object, The elements of the returned list can be used in subsequent
// NewChild() calls to create new zone objects.
//
func (r *Region) ListChildren(ctx context.Context) (int64, []string, error) {

	records, rev, err := r.Store.List(ctx, store.KeyRootInventory, r.KeyChildIndex)

	if err = r.mapErrStoreValue(err); err != nil {
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
// around ListChildren() followed by a NewChild() and Read() on each
// name discovered.
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

		child, err := r.NewChild(v)

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

func (r *Region) mapErrStoreValue(err error) error {
	switch err {
	case errors.ErrStoreKeyNotFound(r.Key):
		return errors.ErrRegionNotFound{Region: r.Region}

	case errors.ErrStoreKeyNotFound(r.KeyIndexEntry):
		return errors.ErrRegionIndexNotFound{Region: r.Region}

	case errors.ErrStoreKeyNotFound(r.KeyChildIndex):
		return errors.ErrRegionChildIndexNotFound{Region: r.Region}

	case errors.ErrStoreAlreadyExists(r.KeyIndexEntry), errors.ErrStoreAlreadyExists(r.Key):
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
	Store         *store.Store
	KeyChildIndex string
	KeyIndexEntry string
	Key           string
	Table         string
	Region        string
	Zone          string

	revisionInfo

	details *pb.ZoneDetails
	record  *pb.Store_RecordDefinition_Zone
}

// SetDetails is used to attach some attribute information to the object.
//
// The attribute information is not persisted to the store until an Update()
// call is made.
//
// The current revision of the zone object is reset
//
func (z *Zone) SetDetails(details *pb.ZoneDetails) {
	z.details = details
	z.resetRevision()
}

// GetDetails is used to extract the attribute information from the object. The
// attribute information must have been previously read from the store (see
// the Read() method) or attached via a SetDetails() call.
//
// May return nil if there are no attributes currently held in the object.
//
func (z *Zone) GetDetails() *pb.ZoneDetails {
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
// either be retrieved by one of the GetRevisionXxx() call or used for
// subsequent conditional operations such as a conditional Update() call.
//
func (z *Zone) Create(ctx context.Context) (int64, error) {

	if z.details == nil {
		return store.RevisionInvalid, errors.ErrDetailsNotAvailable("zone")
	}

	record := &pb.Store_RecordDefinition_Zone{
		Details: z.details,
	}

	v, err := store.Encode(record)

	if err != nil {
		return store.RevisionInvalid, err
	}

	// Create the child and its index as an atomic pair.
	//
	keySet := &map[string]string{
		z.KeyIndexEntry: z.Zone,
		z.Key:           v,
	}

	rev, err := z.Store.CreateMultiple(ctx, store.KeyRootInventory, keySet)

	if err = z.mapErrStoreValue(err); err != nil {
		return store.RevisionInvalid, err
	}

	z.record = record

	return z.updateRevisionInfo(rev), nil
}

// Read is used to load a record from the underlying store to populate the
// fields in the object and determine the revision values associated with
// that record.
//
// Once the Read() has completed successfully the details and other
// information for the object can be retrieved by any of the GetXxx() methods
// for that object.
//
func (z *Zone) Read(ctx context.Context) (int64, error) {

	v, rev, err := z.Store.Read(ctx, store.KeyRootInventory, z.Key)

	if err = z.mapErrStoreValue(err); err != nil {
		return store.RevisionInvalid, err
	}

	record := &pb.Store_RecordDefinition_Zone{}

	if err = store.Decode(*v, record); err != nil {
		return store.RevisionInvalid, err
	}

	z.details = record.Details
	z.record = record

	return z.updateRevisionInfo(rev), nil
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

	record := &pb.Store_RecordDefinition_Zone{
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
		z.GetRevisionForRequest(unconditional),
		v)

	if err = z.mapErrStoreValue(err); err != nil {
		return store.RevisionInvalid, err
	}

	z.record = record

	return z.updateRevisionInfo(rev), nil
}

// Delete is used to remove the persisted copy of the object from the
// store along with any index information needed to navigate to or
// through that object. The delete can be either unconditional by
// setting the unconditional parameter to true, or conditional based
// on the revision of the object compared to the revision of the
// associated record in the underlying store.
//
// Deleting the record from the underlying store has no effect on the
// values held in the fields of the object other than updating the
// revision information using the information returned by the store
// operation.
//
func (z *Zone) Delete(ctx context.Context, unconditional bool) (int64, error) {

	// Delete the record and its index as an atomic pair.
	//
	keySet := &map[string]int64{
		z.KeyIndexEntry: store.RevisionInvalid,
		z.Key:           z.GetRevisionForRequest(unconditional),
	}

	rev, err := z.Store.DeleteMultiple(ctx, store.KeyRootInventory, keySet)

	if err = z.mapErrStoreValue(err); err != nil {
		return store.RevisionInvalid, err
	}

	z.record = nil

	return z.updateRevisionInfo(rev), nil
}

// NewChild creates a new child object for the zone within the current
// region using the supplied name. This new object can be used for
// further navigation or for actions involving operations against the
// associated record in the underlying store.
//
// No information is fetched from the underlying store so the attribute
// and revisions fields within the object are not valid.
//
func (z *Zone) NewChild(name string) (*Rack, error) {

	return newRack(z.Store, z.Table, z.Region, z.Zone, name)
}

// ListChildren uses the current object to discover the names of all the
// rack child objects in the underlying store for the the current zone
// object, The elements of the returned list can be used in subsequent
// NewChild() calls to create new rack objects.
//
func (z *Zone) ListChildren(ctx context.Context) (int64, []string, error) {

	records, rev, err := z.Store.List(ctx, store.KeyRootInventory, z.KeyChildIndex)

	if err = z.mapErrStoreValue(err); err != nil {
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
// around ListChildren() followed by a NewChild() and Read() on each
// name discovered.
//
func (z *Zone) FetchChildren(ctx context.Context) (int64, *map[string]Rack, error) {

	rev, names, err := z.ListChildren(ctx)

	if err != nil {
		return store.RevisionInvalid, nil, err
	}

	children := make(map[string]Rack, len(names))

	for _, v := range names {

		child, err := z.NewChild(v)

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

func (z *Zone) mapErrStoreValue(err error) error {
	switch err {
	case errors.ErrStoreKeyNotFound(z.Key):
		return errors.ErrZoneNotFound{Region: z.Region, Zone: z.Zone}

	case errors.ErrStoreKeyNotFound(z.KeyIndexEntry):
		return errors.ErrZoneIndexNotFound{Region: z.Region, Zone: z.Zone}

	case errors.ErrStoreKeyNotFound(z.KeyChildIndex):
		return errors.ErrZoneChildIndexNotFound{Region: z.Region, Zone: z.Zone}

	case errors.ErrStoreAlreadyExists(z.KeyIndexEntry), errors.ErrStoreAlreadyExists(z.Key):
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
	Store         *store.Store
	KeyIndexPdu   string
	KeyIndexTor   string
	KeyIndexBlade string
	KeyIndexEntry string
	Key           string
	Table         string
	Region        string
	Zone          string
	Rack          string

	revisionInfo

	details *pb.RackDetails
	record  *pb.Store_RecordDefinition_Rack
}

// SetDetails is used to attach some attribute information to the object.
//
// The attribute information is not persisted to the store until an Update()
// call is made.
//
// The current revision of the rack object is reset
//
func (r *Rack) SetDetails(details *pb.RackDetails) {
	r.details = details
	r.resetRevision()
}

// GetDetails is used to extract the attribute information from the object. The
// attribute information must have been previously read from the store (see
// the Read() method) or attached via a SetDetails() call.
//
// May return nil if there are no attributes currently held in the object.
//
func (r *Rack) GetDetails() *pb.RackDetails {
	return r.details
}

// Copy returns a copy of the rack definition based on the contents of the
// current object.
//
func (r *Rack) Copy(ctx context.Context) (*pb.Definition_Rack, error) {

	_, pdus, err := r.FetchPdus(ctx)
	if err != nil {
		return nil, err
	}

	_, tors, err := r.FetchTors(ctx)
	if err != nil {
		return nil, err
	}

	_, blades, err := r.FetchBlades(ctx)
	if err != nil {
		return nil, err
	}

	rack := &pb.Definition_Rack{
		Details: r.GetDetails(),
		Pdus:    make(map[int64]*pb.Definition_Pdu, len(*pdus)),
		Tors:    make(map[int64]*pb.Definition_Tor, len(*tors)),
		Blades:  make(map[int64]*pb.Definition_Blade, len(*blades)),
	}

	for pduIndex, pdu := range *pdus {
		rack.Pdus[pduIndex] = pdu.Copy()
	}

	for torIndex, tor := range *tors {
		rack.Tors[torIndex] = tor.Copy()
	}

	for bladeIndex, blade := range *blades {
		rack.Blades[bladeIndex] = blade.Copy()
	}

	return rack, nil
}

// Create is used to create a record in the underlying store for the
// object along with the associated index information.
//
// The underlying store record will contain the information currently
// held in the object.
//
// Once the store operation completes successfully, the revision fields
// in the object will be updated to that returned by the store. These can
// either be retrieved by one of the GetRevisionXxx() call or used for
// subsequent conditional operations such as a conditional Update() call.
//
func (r *Rack) Create(ctx context.Context) (int64, error) {

	if r.details == nil {
		return store.RevisionInvalid, errors.ErrDetailsNotAvailable("rack")
	}

	record := &pb.Store_RecordDefinition_Rack{
		Details: r.details,
	}

	v, err := store.Encode(record)

	if err != nil {
		return store.RevisionInvalid, err
	}

	// Create the child and its index as an atomic pair.
	//
	keySet := &map[string]string{
		r.KeyIndexEntry: r.Rack,
		r.Key:           v,
	}

	rev, err := r.Store.CreateMultiple(ctx, store.KeyRootInventory, keySet)

	if err = r.mapErrStoreValue(err); err != nil {
		return store.RevisionInvalid, err
	}

	r.record = record

	return r.updateRevisionInfo(rev), nil
}

// Read is used to load a record from the underlying store to populate the
// fields in the object and determine the revision values associated with
// that record.
//
// Once the Read() has completed successfully the details and other
// information for the object can be retrieved by any of the GetXxx() methods
// for that object.
//
func (r *Rack) Read(ctx context.Context) (int64, error) {

	v, rev, err := r.Store.Read(ctx, store.KeyRootInventory, r.Key)

	if err = r.mapErrStoreValue(err); err != nil {
		return store.RevisionInvalid, err
	}

	record := &pb.Store_RecordDefinition_Rack{}

	if err = store.Decode(*v, record); err != nil {
		return store.RevisionInvalid, err
	}

	r.details = record.Details
	r.record = record

	return r.updateRevisionInfo(rev), nil
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

	record := &pb.Store_RecordDefinition_Rack{
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
		r.GetRevisionForRequest(unconditional),
		v)

	if err = r.mapErrStoreValue(err); err != nil {
		return store.RevisionInvalid, err
	}

	r.record = record

	return r.updateRevisionInfo(rev), nil
}

// Delete is used to remove the persisted copy of the object from the
// store along with any index information needed to navigate to or
// through that object. The delete can be either unconditional by
// setting the unconditional parameter to true, or conditional based
// on the revision of the object compared to the revision of the
// associated record in the underlying store.
//
// Deleting the record from the underlying store has no effect on the
// values held in the fields of the object other than updating the
// revision information using the information returned by the store
// operation.
//
func (r *Rack) Delete(ctx context.Context, unconditional bool) (int64, error) {

	// Delete the record and its index as an atomic pair.
	//
	keySet := &map[string]int64{
		r.KeyIndexEntry: store.RevisionInvalid,
		r.Key:           r.GetRevisionForRequest(unconditional),
	}

	rev, err := r.Store.DeleteMultiple(ctx, store.KeyRootInventory, keySet)

	if err = r.mapErrStoreValue(err); err != nil {
		return store.RevisionInvalid, err
	}

	r.record = nil

	return r.updateRevisionInfo(rev), nil
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
func (r *Rack) NewChild(_ string) (*Zone, error) {
	return nil, errors.ErrFunctionNotAvailable
}

// NewPdu creates a new child object for the pdu within the current
// rack using the supplied identifier. This new object can be used for
// for actions involving operations against the associated record in
// the underlying store.
//
// No information is fetched from the underlying store so the attribute
// and revisions fields within the object are not valid.
//
func (r *Rack) NewPdu(ID int64) (*Pdu, error) {

	return newPdu(r.Store, r.Table, r.Region, r.Zone, r.Rack, ID)
}

// NewTor creates a new child object for the tor within the current
// rack using the supplied identifier. This new object can be used for
// for actions involving operations against the associated record in
// the underlying store.
//
// No information is fetched from the underlying store so the attribute
// and revisions fields within the object are not valid.
//
func (r *Rack) NewTor(ID int64) (*Tor, error) {

	return newTor(r.Store, r.Table, r.Region, r.Zone, r.Rack, ID)
}

// NewBlade creates a new child object for the blade within the current
// rack using the supplied identifier. This new object can be used for
// for actions involving operations against the associated record in
// the underlying store.
//
// No information is fetched from the underlying store so the attribute
// and revisions fields within the object are not valid.
//
func (r *Rack) NewBlade(ID int64) (*Blade, error) {

	return newBlade(r.Store, r.Table, r.Region, r.Zone, r.Rack, ID)
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
func (r *Rack) ListChildren(_ context.Context) (int64, *[]string, error) {
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
func (r *Rack) FetchChildren(_ context.Context) (int64, *map[string]interface{}, error) {
	return store.RevisionInvalid, nil, errors.ErrFunctionNotAvailable
}

// ListPdus uses the current object to discover the names of all the
// pdu child objects in the underlying store for the the current rack
// object, The elements of the returned list can be used in subsequent
// NewPdu() calls to create new pdu objects.
//
func (r *Rack) ListPdus(ctx context.Context) (int64, []int64, error) {

	records, rev, err := r.Store.List(ctx, store.KeyRootInventory, r.KeyIndexPdu)

	if err = r.mapErrStoreValue(err); err != nil {
		return store.RevisionInvalid, nil, err
	}

	names := make([]int64, 0, len(*records))

	for k, v := range *records {

		name := strings.TrimPrefix(k, r.KeyIndexPdu)

		// Verify that the "index" part of the name is numeric
		//
		intName, err := strconv.ParseInt(name, 10, 64)

		if err != nil {
			return store.RevisionInvalid, nil, errors.ErrPduIndexInvalid{
				Region: r.Region,
				Zone:   r.Zone,
				Rack:   r.Rack,
				Pdu:    name,
			}
		}

		intValue, err := strconv.ParseInt(v.Value, 10, 64)

		if err != nil {
			return store.RevisionInvalid, nil, errors.ErrPduIndexInvalid{
				Region: r.Region,
				Zone:   r.Zone,
				Rack:   r.Rack,
				Pdu:    v.Value,
			}
		}

		if intName != intValue {
			return store.RevisionInvalid, nil, errors.ErrIndexKeyValueMismatch{
				Namespace: r.Table,
				Key:       name,
				Value:     v.Value,
			}
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

	if err = r.mapErrStoreValue(err); err != nil {
		return store.RevisionInvalid, nil, err
	}

	names := make([]int64, 0, len(*records))

	for k, v := range *records {

		name := strings.TrimPrefix(k, r.KeyIndexTor)

		// Verify that the "index" part of the name is numeric
		//
		intName, err := strconv.ParseInt(name, 10, 64)

		if err != nil {
			return store.RevisionInvalid, nil, errors.ErrTorIndexInvalid{
				Region: r.Region,
				Zone:   r.Zone,
				Rack:   r.Rack,
				Tor:    name,
			}
		}

		intValue, err := strconv.ParseInt(v.Value, 10, 64)

		if err != nil {
			return store.RevisionInvalid, nil, errors.ErrTorIndexInvalid{
				Region: r.Region,
				Zone:   r.Zone,
				Rack:   r.Rack,
				Tor:    v.Value,
			}
		}

		if intName != intValue {
			return store.RevisionInvalid, nil, errors.ErrIndexKeyValueMismatch{
				Namespace: r.Table,
				Key:       name,
				Value:     v.Value,
			}
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

	if err = r.mapErrStoreValue(err); err != nil {
		return store.RevisionInvalid, nil, err
	}

	names := make([]int64, 0, len(*records))

	for k, v := range *records {

		name := strings.TrimPrefix(k, r.KeyIndexBlade)

		// Verify that the "index" part of the name is numeric
		//
		intName, err := strconv.ParseInt(name, 10, 64)

		if err != nil {
			return store.RevisionInvalid, nil, errors.ErrBladeIndexInvalid{
				Region: r.Table,
				Zone:   r.Zone,
				Rack:   r.Rack,
				Blade:  name,
			}
		}

		intValue, err := strconv.ParseInt(v.Value, 10, 64)

		if err != nil {
			return store.RevisionInvalid, nil, errors.ErrBladeIndexInvalid{
				Region: r.Table,
				Zone:   r.Zone,
				Rack:   r.Rack,
				Blade:  v.Value,
			}
		}

		if intName != intValue {
			return store.RevisionInvalid, nil, errors.ErrIndexKeyValueMismatch{
				Namespace: r.Table,
				Key:       name,
				Value:     v.Value,
			}
		}

		names = append(names, intValue)
	}

	return rev, names, nil
}

// FetchPdus is used to discover all the child pdu objects in the
// underlying store for the current rack object and to generate a new
// pdu object for each of those children. It is a convenience wrapper
// around ListPdus() followed by a NewPdu() and Read() on each name
// discovered.
//
func (r *Rack) FetchPdus(ctx context.Context) (int64, *map[int64]Pdu, error) {

	rev, names, err := r.ListPdus(ctx)

	if err != nil {
		return store.RevisionInvalid, nil, err
	}

	pdus := make(map[int64]Pdu, len(names))

	for _, v := range names {

		pdu, err := r.NewPdu(v)

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
// around ListTors() followed by a NewTor() and Read() on each name
// discovered.
//
func (r *Rack) FetchTors(ctx context.Context) (int64, *map[int64]Tor, error) {

	rev, names, err := r.ListTors(ctx)

	if err != nil {
		return store.RevisionInvalid, nil, err
	}

	tors := make(map[int64]Tor, len(names))

	for _, v := range names {

		tor, err := r.NewTor(v)

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
// around ListBlades() followed by a NewBlade() and Read() on each name
// discovered.
//
func (r *Rack) FetchBlades(ctx context.Context) (int64, *map[int64]Blade, error) {

	rev, names, err := r.ListBlades(ctx)

	if err != nil {
		return store.RevisionInvalid, nil, err
	}

	blades := make(map[int64]Blade, len(names))

	for _, v := range names {

		blade, err := r.NewBlade(v)

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

func (r *Rack) mapErrStoreValue(err error) error {
	switch err {
	case errors.ErrStoreKeyNotFound(r.Key):
		return errors.ErrRackNotFound{Region: r.Region, Zone: r.Zone, Rack: r.Rack}

	case errors.ErrStoreKeyNotFound(r.KeyIndexEntry):
		return errors.ErrRackIndexNotFound{Region: r.Region, Zone: r.Zone, Rack: r.Rack}

	case errors.ErrStoreKeyNotFound(r.KeyIndexPdu):
		return errors.ErrRackPduIndexNotFound{Region: r.Region, Zone: r.Zone, Rack: r.Rack}

	case errors.ErrStoreKeyNotFound(r.KeyIndexTor):
		return errors.ErrRackTorIndexNotFound{Region: r.Region, Zone: r.Zone, Rack: r.Rack}

	case errors.ErrStoreKeyNotFound(r.KeyIndexBlade):
		return errors.ErrRackBladeIndexNotFound{Region: r.Region, Zone: r.Zone, Rack: r.Rack}

	case errors.ErrStoreAlreadyExists(r.KeyIndexEntry), errors.ErrStoreAlreadyExists(r.Key):
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
	Store         *store.Store
	Key           string
	KeyIndexEntry string
	Table         string
	Region        string
	Zone          string
	Rack          string
	ID            int64

	revisionInfo

	details *pb.PduDetails
	ports   *map[int64]*pb.PowerPort
	record  *pb.Store_RecordDefinition_Pdu
}

// SetDetails is used to attach some attribute information to the object.
//
// The attribute information is not persisted to the store until an Update()
// call is made.
//
// The current revision of the pdu object is reset
//
func (p *Pdu) SetDetails(details *pb.PduDetails) {
	p.details = details
	p.resetRevision()
}

// GetDetails is used to extract the attribute information from the object. The
// attribute information must have been previously read from the store (see
// the Read() method) or attached via a SetDetails() call.
//
// May return nil if there are no attributes currently held in the object.
//
func (p *Pdu) GetDetails() *pb.PduDetails {
	return p.details
}

// SetPorts is used to attach some power port information to the object.
//
// The port information is not persisted to the store until an Update()
// call is made.
//
// The current revision of the pdu object is reset
//
func (p *Pdu) SetPorts(ports *map[int64]*pb.PowerPort) {
	p.ports = ports
	p.resetRevision()
}

// GetPorts is used to extract the power port information from the object.
// The port information must have been previously read from the store (see
// the Read() method) or attached via a SetPorts() call.
//
// May return nil if there are no power port information currently held
// in the object.
//
func (p *Pdu) GetPorts() *map[int64]*pb.PowerPort {
	return p.ports
}

// Copy returns a copy of the pdu definition based on the contents of the
// current object.
//
func (p *Pdu) Copy() *pb.Definition_Pdu {
	pdu := &pb.Definition_Pdu{
		Details: p.GetDetails(),
		Ports:   *p.GetPorts(),
	}

	return pdu
}

// Create is used to create a record in the underlying store for the
// object along with the associated index information.
//
// The underlying store record will contain the information currently
// held in the object.
//
// Once the store operation completes successfully, the revision fields
// in the object will be updated to that returned by the store. These can
// either be retrieved by one of the GetRevisionXxx() call or used for
// subsequent conditional operations such as a conditional Update() call.
//
func (p *Pdu) Create(ctx context.Context) (int64, error) {

	if p.details == nil {
		return store.RevisionInvalid, errors.ErrDetailsNotAvailable("pdu")
	}

	if p.ports == nil {
		return store.RevisionInvalid, errors.ErrPortsNotAvailable("pdu")
	}

	record := &pb.Store_RecordDefinition_Pdu{
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
		p.KeyIndexEntry: fmt.Sprintf("%d", p.ID),
		p.Key:           v,
	}

	rev, err := p.Store.CreateMultiple(ctx, store.KeyRootInventory, keySet)

	if err = p.mapErrStoreValue(err); err != nil {
		return store.RevisionInvalid, err
	}

	p.record = record

	return p.updateRevisionInfo(rev), nil
}

// Read is used to load a record from the underlying store to populate the
// fields in the object and determine the revision values associated with
// that record.
//
// Once the Read() has completed successfully the details and other
// information for the object can be retrieved by any of the GetXxx() methods
// for that object.
//
func (p *Pdu) Read(ctx context.Context) (int64, error) {

	v, rev, err := p.Store.Read(ctx, store.KeyRootInventory, p.Key)

	if err = p.mapErrStoreValue(err); err != nil {
		return store.RevisionInvalid, err
	}

	record := &pb.Store_RecordDefinition_Pdu{}

	if err = store.Decode(*v, record); err != nil {
		return store.RevisionInvalid, err
	}

	p.details = record.Details
	p.ports = &record.Ports
	p.record = record

	return p.updateRevisionInfo(rev), nil
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

	record := &pb.Store_RecordDefinition_Pdu{
		Details: p.details,
		Ports:   *p.ports,
	}

	v, err := store.Encode(record)

	if err != nil {
		return store.RevisionInvalid, err
	}

	rev, err := p.Store.Update(
		ctx,
		store.KeyRootInventory,
		p.Key,
		p.GetRevisionForRequest(unconditional),
		v)

	if err = p.mapErrStoreValue(err); err != nil {
		return store.RevisionInvalid, err
	}

	p.record = record

	return p.updateRevisionInfo(rev), nil
}

// Delete is used to remove the persisted copy of the object from the
// store along with any index information needed to navigate to or
// through that object. The delete can be either unconditional by
// setting the unconditional parameter to true, or conditional based
// on the revision of the object compared to the revision of the
// associated record in the underlying store.
//
// Deleting the record from the underlying store has no effect on the
// values held in the fields of the object other than updating the
// revision information using the information returned by the store
// operation.
//
func (p *Pdu) Delete(ctx context.Context, unconditional bool) (int64, error) {

	// Delete the record and its index as an atomic pair.
	//
	keySet := &map[string]int64{
		p.KeyIndexEntry: store.RevisionInvalid,
		p.Key:           p.GetRevisionForRequest(unconditional),
	}

	rev, err := p.Store.DeleteMultiple(ctx, store.KeyRootInventory, keySet)

	if err = p.mapErrStoreValue(err); err != nil {
		return store.RevisionInvalid, err
	}

	p.record = nil

	return p.updateRevisionInfo(rev), nil
}

func (p *Pdu) mapErrStoreValue(err error) error {
	switch err {
	case errors.ErrStoreKeyNotFound(p.Key):
		return errors.ErrPduNotFound{Region: p.Region, Zone: p.Zone, Rack: p.Rack, Pdu: p.ID}

	case errors.ErrStoreKeyNotFound(p.KeyIndexEntry):
		return errors.ErrPduIndexNotFound{Region: p.Region, Zone: p.Zone, Rack: p.Rack, Pdu: p.ID}

	case errors.ErrStoreAlreadyExists(p.KeyIndexEntry), errors.ErrStoreAlreadyExists(p.Key):
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
	Store         *store.Store
	Key           string
	KeyIndexEntry string
	Table         string
	Region        string
	Zone          string
	Rack          string
	ID            int64

	revisionInfo

	details *pb.TorDetails
	ports   *map[int64]*pb.NetworkPort
	record  *pb.Store_RecordDefinition_Tor
}

// SetDetails is used to attach some attribute information to the object.
//
// The attribute information is not persisted to the store until an Update()
// call is made.
//
// The current revision of the tor object is reset
//
func (t *Tor) SetDetails(details *pb.TorDetails) {
	t.details = details
	t.resetRevision()
}

// GetDetails is used to extract the attribute information from the object. The
// attribute information must have been previously read from the store (see
// the Read() method) or attached via a SetDetails() call.
//
// May return nil if there are no attributes currently held in the object.
//
func (t *Tor) GetDetails() *pb.TorDetails {
	return t.details
}

// SetPorts is used to attach some network port information to the object.
//
// The port information is not persisted to the store until an Update()
// call is made.
//
// The current revision of the tor object is reset
//
func (t *Tor) SetPorts(ports *map[int64]*pb.NetworkPort) {
	t.ports = ports
	t.resetRevision()
}

// GetPorts is used to extract the network port information from the object.
// The port information must have been previously read from the store (see
// the Read() method) or attached via a SetPorts() call.
//
// May return nil if there are no network port information currently held
// in the object.
//
func (t *Tor) GetPorts() *map[int64]*pb.NetworkPort {
	return t.ports
}

// Copy returns a copy of the tor definition based on the contents of the
// current object.
//
func (t *Tor) Copy() *pb.Definition_Tor {
	tor := &pb.Definition_Tor{
		Details: t.GetDetails(),
		Ports:   *t.GetPorts(),
	}

	return tor
}

// Create is used to create a record in the underlying store for the
// object along with the associated index information.
//
// The underlying store record will contain the information currently
// held in the object.
//
// Once the store operation completes successfully, the revision fields
// in the object will be updated to that returned by the store. These can
// either be retrieved by one of the GetRevisionXxx() call or used for
// subsequent conditional operations such as a conditional Update() call.
//
func (t *Tor) Create(ctx context.Context) (int64, error) {

	if t.details == nil {
		return store.RevisionInvalid, errors.ErrDetailsNotAvailable("tor")
	}

	if t.ports == nil {
		return store.RevisionInvalid, errors.ErrPortsNotAvailable("tor")
	}

	record := &pb.Store_RecordDefinition_Tor{
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
		t.KeyIndexEntry: fmt.Sprintf("%d", t.ID),
		t.Key:           v,
	}

	rev, err := t.Store.CreateMultiple(ctx, store.KeyRootInventory, keySet)

	if err = t.mapErrStoreValue(err); err != nil {
		return store.RevisionInvalid, err
	}

	t.record = record

	return t.updateRevisionInfo(rev), nil
}

// Read is used to load a record from the underlying store to populate the
// fields in the object and determine the revision values associated with
// that record.
//
// Once the Read() has completed successfully the details and other
// information for the object can be retrieved by any of the GetXxx() methods
// for that object.
//
func (t *Tor) Read(ctx context.Context) (int64, error) {

	v, rev, err := t.Store.Read(ctx, store.KeyRootInventory, t.Key)

	if err = t.mapErrStoreValue(err); err != nil {
		return store.RevisionInvalid, err
	}

	record := &pb.Store_RecordDefinition_Tor{}

	if err = store.Decode(*v, record); err != nil {
		return store.RevisionInvalid, err
	}

	t.details = record.Details
	t.ports = &record.Ports
	t.record = record

	return t.updateRevisionInfo(rev), nil
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

	record := &pb.Store_RecordDefinition_Tor{
		Details: t.details,
		Ports:   *t.ports,
	}

	v, err := store.Encode(record)

	if err != nil {
		return store.RevisionInvalid, err
	}

	rev, err := t.Store.Update(
		ctx,
		store.KeyRootInventory,
		t.Key,
		t.GetRevisionForRequest(unconditional),
		v)

	if err = t.mapErrStoreValue(err); err != nil {
		return store.RevisionInvalid, err
	}

	t.record = record

	return t.updateRevisionInfo(rev), nil
}

// Delete is used to remove the persisted copy of the object from the
// store along with any index information needed to navigate to or
// through that object. The delete can be either unconditional by
// setting the unconditional parameter to true, or conditional based
// on the revision of the object compared to the revision of the
// associated record in the underlying store.
//
// Deleting the record from the underlying store has no effect on the
// values held in the fields of the object other than updating the
// revision information using the information returned by the store
// operation.
//
func (t *Tor) Delete(ctx context.Context, unconditional bool) (int64, error) {

	// Delete the record and its index as an atomic pair.
	//
	keySet := &map[string]int64{
		t.KeyIndexEntry: store.RevisionInvalid,
		t.Key:           t.GetRevisionForRequest(unconditional),
	}

	rev, err := t.Store.DeleteMultiple(ctx, store.KeyRootInventory, keySet)

	if err = t.mapErrStoreValue(err); err != nil {
		return store.RevisionInvalid, err
	}

	t.record = nil

	return t.updateRevisionInfo(rev), nil
}

func (t *Tor) mapErrStoreValue(err error) error {
	switch err {
	case errors.ErrStoreKeyNotFound(t.Key):
		return errors.ErrTorNotFound{Region: t.Region, Zone: t.Zone, Rack: t.Rack, Tor: t.ID}

	case errors.ErrStoreKeyNotFound(t.KeyIndexEntry):
		return errors.ErrTorIndexNotFound{Region: t.Region, Zone: t.Zone, Rack: t.Rack, Tor: t.ID}

	case errors.ErrStoreAlreadyExists(t.KeyIndexEntry), errors.ErrStoreAlreadyExists(t.Key):
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
	Store         *store.Store
	Key           string
	KeyIndexEntry string
	Table         string
	Region        string
	Zone          string
	Rack          string
	ID            int64

	revisionInfo

	details       *pb.BladeDetails
	capacity      *pb.BladeCapacity
	bootInfo      *pb.BladeBootInfo
	bootOnPowerOn bool
	record        *pb.Store_RecordDefinition_Blade
}

// SetDetails is used to attach some attribute information to the object.
//
// The attribute information is not persisted to the store until an Update()
// call is made.
//
// The current revision of the blade object is reset
//
func (b *Blade) SetDetails(details *pb.BladeDetails) {
	b.details = details
	b.resetRevision()
}

// GetDetails is used to extract the attribute information from the object. The
// attribute information must have been previously read from the store (see
// the Read() method) or attached via a SetDetails() call.
//
// May return nil if there are no attributes currently held in the object.
//
func (b *Blade) GetDetails() *pb.BladeDetails {
	return b.details
}

// SetCapacity is used to attach some capacity information to the object.
//
// The capacity information is not persisted to the store until an Update()
// call is made.
//
// The current revision of the blade object is reset
//
func (b *Blade) SetCapacity(capacity *pb.BladeCapacity) {
	b.capacity = capacity
	b.resetRevision()
}

// SetBootInfo is used to attach some boot information to the object.
//
// The boot information is not persisted to the store until an Update()
// call is made.
//
// The current revision of the blade object is reset
//
func (b *Blade) SetBootInfo(bootOnPowerOn bool, bootInfo *pb.BladeBootInfo) {
	b.bootOnPowerOn = bootOnPowerOn
	b.bootInfo = bootInfo
	b.resetRevision()
}

// GetCapacity is used to extract the capacity information from the object.
// The capacity information must have been previously read from the store (see
// the Read() method) or attached via a SetCapacity() call.
//
// May return nil if there are no capacity information currently held
// in the object.
//
func (b *Blade) GetCapacity() *pb.BladeCapacity {
	return b.capacity
}

// GetBootInfo is used to extract the boot information from the object.
// The boot information must have been previously read from the store (see
// the Read() method) or attached via a SetBootInfo() call.
//
// May return nil if there are no boot information currently held
// in the object.
//
func (b *Blade) GetBootInfo() (bool, *pb.BladeBootInfo) {
	return b.bootOnPowerOn, b.bootInfo
}

// Copy returns a copy of the blade definition based on the contents of the
// current object.
//
func (b *Blade) Copy() *pb.Definition_Blade {
	bootOnPowerOn, bootInfo := b.GetBootInfo()

	blade := &pb.Definition_Blade{
		Details:       b.GetDetails(),
		Capacity:      b.GetCapacity(),
		BootInfo:      bootInfo,
		BootOnPowerOn: bootOnPowerOn,
	}

	return blade
}

// Create is used to create a record in the underlying store for the
// object along with the associated index information.
//
// The underlying store record will contain the information currently
// held in the object.
//
// Once the store operation completes successfully, the revision fields
// in the object will be updated to that returned by the store. These can
// either be retrieved by one of the GetRevisionXxx() call or used for
// subsequent conditional operations such as a conditional Update() call.
//
func (b *Blade) Create(ctx context.Context) (int64, error) {

	if b.details == nil {
		return store.RevisionInvalid, errors.ErrDetailsNotAvailable("blade")
	}

	if b.capacity == nil {
		return store.RevisionInvalid, errors.ErrCapacityNotAvailable("blade")
	}

	if b.bootInfo == nil {
		return store.RevisionInvalid, errors.ErrBootInfoNotAvailable("blade")
	}

	record := &pb.Store_RecordDefinition_Blade{
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
		b.KeyIndexEntry: fmt.Sprintf("%d", b.ID),
		b.Key:           v,
	}

	rev, err := b.Store.CreateMultiple(ctx, store.KeyRootInventory, keySet)

	if err = b.mapErrStoreValue(err); err != nil {
		return store.RevisionInvalid, err
	}

	b.record = record

	return b.updateRevisionInfo(rev), nil
}

// Read is used to load a record from the underlying store to populate the
// fields in the object and determine the revision values associated with
// that record.
//
// Once the Read() has completed successfully the details and other
// information for the object can be retrieved by any of the GetXxx() methods
// for that object.
//
func (b *Blade) Read(ctx context.Context) (int64, error) {

	v, rev, err := b.Store.Read(ctx, store.KeyRootInventory, b.Key)

	if err = b.mapErrStoreValue(err); err != nil {
		return store.RevisionInvalid, err
	}

	record := &pb.Store_RecordDefinition_Blade{}

	if err = store.Decode(*v, record); err != nil {
		return store.RevisionInvalid, err
	}

	b.details = record.Details
	b.capacity = record.Capacity
	b.bootInfo = record.BootInfo
	b.bootOnPowerOn = record.BootOnPowerOn
	b.record = record

	return b.updateRevisionInfo(rev), nil
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

	record := &pb.Store_RecordDefinition_Blade{
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
		b.GetRevisionForRequest(unconditional),
		v)

	if err = b.mapErrStoreValue(err); err != nil {
		return store.RevisionInvalid, err
	}

	b.record = record

	return b.updateRevisionInfo(rev), nil
}

// Delete is used to remove the persisted copy of the object from the
// store along with any index information needed to navigate to or
// through that object. The delete can be either unconditional by
// setting the unconditional parameter to true, or conditional based
// on the revision of the object compared to the revision of the
// associated record in the underlying store.
//
// Deleting the record from the underlying store has no effect on the
// values held in the fields of the object other than updating the
// revision information using the information returned by the store
// operation.
//
func (b *Blade) Delete(ctx context.Context, unconditional bool) (int64, error) {

	// Delete the record and its index as an atomic pair.
	//
	keySet := &map[string]int64{
		b.KeyIndexEntry: store.RevisionInvalid,
		b.Key:           b.GetRevisionForRequest(unconditional),
	}

	rev, err := b.Store.DeleteMultiple(ctx, store.KeyRootInventory, keySet)

	if err = b.mapErrStoreValue(err); err != nil {
		return store.RevisionInvalid, err
	}

	b.record = nil

	return b.updateRevisionInfo(rev), nil
}

func (b *Blade) mapErrStoreValue(err error) error {
	switch err {
	case errors.ErrStoreKeyNotFound(b.Key):
		return errors.ErrBladeNotFound{Region: b.Region, Zone: b.Zone, Rack: b.Rack, Blade: b.ID}

	case errors.ErrStoreKeyNotFound(b.KeyIndexEntry):
		return errors.ErrBladeIndexNotFound{Region: b.Region, Zone: b.Zone, Rack: b.Rack, Blade: b.ID}

	case errors.ErrStoreAlreadyExists(b.KeyIndexEntry), errors.ErrStoreAlreadyExists(b.Key):
		return errors.ErrBladeAlreadyExists{Region: b.Region, Zone: b.Zone, Rack: b.Rack, Blade: b.ID}
	}

	return err
}
