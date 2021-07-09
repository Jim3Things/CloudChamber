package inventory

import (
	"context"

	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/namespace"
	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/store"
	"github.com/Jim3Things/CloudChamber/simulation/pkg/errors"
	pb "github.com/Jim3Things/CloudChamber/simulation/pkg/protos/inventory"
)

// Region is a structure representing a region object. This object can be used
// to operate on the associated region records in the underlying store, or to
// navigate to child zone objects. The object can store information fetched
// from the underlying store, or as a staging area in preparation for updates
// to the store.
//
type Region struct {
	regionNode

	revisionInfo
	itemStore

	details *pb.RegionDetails
}

// NewRegion is a convenience function used to construct a Region object
// from scratch rather than relative to its parent.
//
func (m *Inventory) NewRegion(table namespace.TableName, region string) (*Region, error) {

	return newRegion(m.Store, table, region)
}

func newRegion(store *store.Store, table namespace.TableName, region string) (*Region, error) {

	keyIndex, err := namespace.GetKeyForIndexZones(table, region)
	if err != nil {
		return nil, err
	}

	keyIndexEntry, err := namespace.GetKeyForIndexEntryRegion(table, region)
	if nil != err {
		return nil, err
	}

	key, err := namespace.GetKeyForRegion(table, region)
	if nil != err {
		return nil, err
	}

	r := &Region{
		regionNode: newRegionNode(store, keyIndexEntry, key, table, keyIndex, region),
	}

	r.itemStore.provider = r

	return r, nil
}

// SetDetails is used to attach some attribute information to the object.
//
// The attribute information is not persisted to the store until an Update()
// call is made.
//
// The current revision of the region object is reset
//
func (r *Region) SetDetails(details *pb.RegionDetails) {
	r.details = details.Clone()
	r.resetRevision()
}

// GetDetails is used to extract the attribute information from the object. The
// attribute information must have been previously read from the store (see
// the Read() method) or attached via a SetDetails() call.
//
// May return nil if there are no attributes currently held in the object.
//
func (r *Region) GetDetails() *pb.RegionDetails {
	return r.details.Clone()
}

// GetDefinitionRegion returns a copy of the rack definition based on the contents of the
// current object.
//
func (r *Region) GetDefinitionRegion() *pb.Definition_Region {
	return &pb.Definition_Region{
		Details: r.GetDetails(),
		Zones:   make(map[string]*pb.Definition_Zone),
	}
}

// Equal is used to provide a simple equality check for use to determine
// if the current region matches the supplied definition. Typically used
// when looking to see if the record has been changed.
//
// Note: only considers information relating to the parent object and does
//       not include comparisons for any descendants.
//
func (r *Region) Equal(d *pb.Definition_Region) bool {
	return r.details.Equal(d.GetDetails())
}

// NotEqual is used to provide a simple equality check for use to determine
// if the current region does not match the supplied definition. Typically used
// when looking to see if the record has been changed.
//
// Note: only considers information relating to the parent object and does
//       not include comparisons for any descendants.
//
func (r *Region) NotEqual(d *pb.Definition_Region) bool {
	return !r.Equal(d)
}

// Load converts the string representation that was previously returned from a
// Save call back into the view's specific content, such that it can be used as
// part of normal operations.
func (r *Region) Load(view ViewType, value string) error {
	if view != ViewDefinition {
		return errors.ErrUnimplementedView
	}

	record := &pb.Store_RecordDefinition_Region{}

	if err := store.Decode(value, record); err != nil {
		return err
	}

	r.details = record.Details
	return nil
}

// Save returns the view's specific content as a string which can be stored
// for later retrieval.  If no view data is available, an error is returned.
func (r *Region) Save(view ViewType) (string, error) {
	if view != ViewDefinition {
		return "", errors.ErrUnimplementedView
	}

	if r.details == nil {
		return "", errors.ErrDetailsNotAvailable("region")
	}

	record := &pb.Store_RecordDefinition_Region{
		Details: r.details,
	}

	return store.Encode(record)
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

		rev, err = child.Read(ctx, ViewDefinition)

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
