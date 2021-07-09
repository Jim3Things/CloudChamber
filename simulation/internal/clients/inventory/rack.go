package inventory

import (
	"context"
	"strconv"
	"strings"

	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/namespace"
	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/store"
	"github.com/Jim3Things/CloudChamber/simulation/pkg/errors"
	pb "github.com/Jim3Things/CloudChamber/simulation/pkg/protos/inventory"
)

// Rack is a structure representing a rack object. This object can be used
// to operate on the associated rack records in the underlying store, or to
// navigate to child pdu, tor or blade objects. The object can store
// information fetched from the underlying store, or as a staging area in
// preparation for updates to the store.
//
type Rack struct {
	rackNode

	revisionInfo
    itemStoreLeaf

	details *pb.RackDetails
}

// NewRack is a convenience function used to construct a Rack object
// from scratch rather than relative to its parent.
//
func (m *Inventory) NewRack(
	table namespace.TableName,
	region string,
	zone string,
	rack string) (*Rack, error) {

	return newRack(m.Store, table, region, zone, rack)
}

func newRack(
	store *store.Store,
	table namespace.TableName,
	region string,
	zone string,
	rack string) (*Rack, error) {

	keyIndexPdu, err := namespace.GetKeyForIndexPdus(table, region, zone, rack)
	if err != nil {
		return nil, err
	}

	keyIndexTor, err := namespace.GetKeyForIndexTors(table, region, zone, rack)
	if err != nil {
		return nil, err
	}

	keyIndexBlade, err := namespace.GetKeyForIndexBlades(table, region, zone, rack)
	if err != nil {
		return nil, err
	}

	keyIndexEntry, err := namespace.GetKeyForIndexEntryRack(table, region, zone, rack)
	if nil != err {
		return nil, err
	}

	key, err := namespace.GetKeyForRack(table, region, zone, rack)
	if nil != err {
		return nil, err
	}

	r := &Rack{
		rackNode: newRackNode(
			store,
			keyIndexEntry,
			key,
			table,
			region,
			zone,
			keyIndexPdu,
			keyIndexTor,
			keyIndexBlade,
			rack),
	}

	r.itemStoreLeaf.provider = r

	return r, nil
}

// SetDetails is used to attach some attribute information to the object.
//
// The attribute information is not persisted to the store until an Update()
// call is made.
//
// The current revision of the rack object is reset
//
func (r *Rack) SetDetails(details *pb.RackDetails) {
	r.details = details.Clone()
	r.resetRevision()
}

// GetDetails is used to extract the attribute information from the object. The
// attribute information must have been previously read from the store (see
// the Read() method) or attached via a SetDetails() call.
//
// May return nil if there are no attributes currently held in the object.
//
func (r *Rack) GetDetails() *pb.RackDetails {
	return r.details.Clone()
}

// GetDefinitionRack returns a copy of the rack definition based on the contents of the
// current object.
//
func (r *Rack) GetDefinitionRack() *pb.Definition_Rack {

	return &pb.Definition_Rack{
		Details: r.GetDetails(),
		Pdus:    make(map[int64]*pb.Definition_Pdu),
		Tors:    make(map[int64]*pb.Definition_Tor),
		Blades:  make(map[int64]*pb.Definition_Blade),
	}
}

// Equal is used to provide a simple equality check for use to determine
// if the current rack matches the supplied definition. Typically used
// when looking to see if the record has been changed.
//
// Note: only considers information relating to the parent object and does
//       not include comparisons for any descendants.
//
func (r *Rack) Equal(d *pb.Definition_Rack) bool {
	return r.details.Equal(d.GetDetails())
}

// NotEqual is used to provide a simple equality check for use to determine
// if the current rack do not match the supplied definition. Typically used
// when looking to see if the record has been changed.
//
// Note: only considers information relating to the parent object and does
//       not include comparisons for any descendants.
//
func (r *Rack) NotEqual(d *pb.Definition_Rack) bool {
	return !r.Equal(d)
}

// GetDefinitionRackWithChildren returns a copy of the rack definition based on the contents of the
// current object and fully populates all the child maps..
//
func (r *Rack) GetDefinitionRackWithChildren(ctx context.Context) (*pb.Definition_Rack, error) {

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
		Pdus:    make(map[int64]*pb.Definition_Pdu, len(pdus)),
		Tors:    make(map[int64]*pb.Definition_Tor, len(tors)),
		Blades:  make(map[int64]*pb.Definition_Blade, len(blades)),
	}

	for pduIndex, pdu := range pdus {
		rack.Pdus[pduIndex] = pdu.GetDefinitionPdu()
	}

	for torIndex, tor := range tors {
		rack.Tors[torIndex] = tor.GetDefinitionTor()
	}

	for bladeIndex, blade := range blades {
		rack.Blades[bladeIndex] = blade.GetDefinitionBlade()
	}

	return rack, nil
}

// Load converts the string representation that was previously returned from a
// Save call back into the view's specific content, such that it can be used as
// part of normal operations.
func (r *Rack) Load(view ViewType, value string) error {
	if view != ViewDefinition {
		return errors.ErrUnimplementedView
	}

	record := &pb.Store_RecordDefinition_Rack{}

	if err := store.Decode(value, record); err != nil {
		return err
	}

	r.details = record.GetDetails()

	return nil
}

// Save returns the view's specific content as a string which can be stored
// for later retrieval.  If no view data is available, an error is returned.
func (r *Rack) Save(view ViewType) (string, error) {
	if view != ViewDefinition {
		return "", errors.ErrUnimplementedView
	}

	if r.details == nil {
		return "", errors.ErrDetailsNotAvailable("rack")
	}

	record := &pb.Store_RecordDefinition_Rack{
		Details: r.details,
	}

	return store.Encode(record)
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

func (r *Rack) listChildrenOfType(
	ctx context.Context,
	index string,
	indexInvalid func(value string) error) (int64, []int64, error) {

	records, rev, err := r.Store.List(ctx, namespace.KeyRootInventory, index)
	if err = r.mapErrStoreValue(err); err != nil {
		return store.RevisionInvalid, nil, err
	}

	names := make([]int64, 0, len(*records))

	for k, v := range *records {

		name := strings.TrimPrefix(k, index)

		// Verify that the "index" part of the name is numeric
		//
		intName, err := strconv.ParseInt(name, 10, 64)
		if err != nil {
			return store.RevisionInvalid, nil, indexInvalid(name)
		}

		intValue, err := strconv.ParseInt(v.Value, 10, 64)
		if err != nil {
			return store.RevisionInvalid, nil, indexInvalid(v.Value)
		}

		if intName != intValue {
			return store.RevisionInvalid, nil, indexInvalid(v.Value)
		}

		names = append(names, intValue)
	}

	return rev, names, nil
}

// ListPdus uses the current object to discover the names of all the
// pdu child objects in the underlying store for the the current rack
// object, The elements of the returned list can be used in subsequent
// NewPdu() calls to create new pdu objects.
//
func (r *Rack) ListPdus(ctx context.Context) (int64, []int64, error) {
	return r.listChildrenOfType(ctx, r.KeyIndexPdu, func(val string) error {
		return errors.ErrPduIndexInvalid{
			Region: r.Region,
			Zone:   r.Zone,
			Rack:   r.Rack,
			Pdu:    val,
		}})
}

// ListTors uses the current object to discover the names of all the
// tor child objects in the underlying store for the the current rack
// object, The elements of the returned list can be used in subsequent
// NewTor() calls to create new tor objects.
//
func (r *Rack) ListTors(ctx context.Context) (int64, []int64, error) {
	return r.listChildrenOfType(ctx, r.KeyIndexTor, func(val string) error {
		return errors.ErrTorIndexInvalid{
			Region: r.Region,
			Zone:   r.Zone,
			Rack:   r.Rack,
			Tor:    val,
		}})
}

// ListBlades uses the current object to discover the names of all the
// blade child objects in the underlying store for the the current rack
// object, The elements of the returned list can be used in subsequent
// NewBlade() calls to create new blade objects.
//
func (r *Rack) ListBlades(ctx context.Context) (int64, []int64, error) {
	return r.listChildrenOfType(ctx, r.KeyIndexBlade, func(val string) error {
		return errors.ErrBladeIndexInvalid{
			Region: r.Region,
			Zone:   r.Zone,
			Rack:   r.Rack,
			Blade:    val,
		}})
}

// FetchPdus is used to discover all the child pdu objects in the
// underlying store for the current rack object and to generate a new
// pdu object for each of those children. It is a convenience wrapper
// around ListPdus() followed by a NewPdu() and Read() on each name
// discovered.
//
func (r *Rack) FetchPdus(ctx context.Context) (int64, map[int64]Pdu, error) {

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

		_, err = pdu.Read(ctx, ViewDefinition)
		if err != nil {
			return store.RevisionInvalid, nil, err
		}

		pdus[v] = *pdu
	}

	return rev, pdus, nil
}

// FetchTors is used to discover all the child tor objects in the
// underlying store for the current rack object and to generate a new
// tor object for each of those children. It is a convenience wrapper
// around ListTors() followed by a NewTor() and Read() on each name
// discovered.
//
func (r *Rack) FetchTors(ctx context.Context) (int64, map[int64]Tor, error) {

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

		_, err = tor.Read(ctx, ViewDefinition)
		if err != nil {
			return store.RevisionInvalid, nil, err
		}

		tors[v] = *tor
	}

	return rev, tors, nil
}

// FetchBlades is used to discover all the child blade objects in the
// underlying store for the current rack object and to generate a new
// blade object for each of those children. It is a convenience wrapper
// around ListBlades() followed by a NewBlade() and Read() on each name
// discovered.
//
func (r *Rack) FetchBlades(ctx context.Context) (int64, map[int64]Blade, error) {

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

		_, err = blade.Read(ctx, ViewDefinition)
		if err != nil {
			return store.RevisionInvalid, nil, err
		}

		blades[v] = *blade
	}

	return rev, blades, nil
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

