package inventory

import (
	"context"

	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/namespace"
	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/store"
	"github.com/Jim3Things/CloudChamber/simulation/internal/tracing"
	"github.com/Jim3Things/CloudChamber/simulation/pkg/errors"
	pb "github.com/Jim3Things/CloudChamber/simulation/pkg/protos/inventory"
)

// Zone is a structure representing a zone object. This object can be used
// to operate on the associated zone records in the underlying store, or to
// navigate to child rack objects. The object can store information fetched
// from the underlying store, or as a staging area in preparation for updates
// to the store.
//
type Zone struct {
	zoneNode

	revisionInfo
    itemStore

	details *pb.ZoneDetails
}

// NewZone is a convenience function used to construct a Zone object
// from scratch rather than relative to its parent.
//
func (m *Inventory) NewZone(table namespace.TableName, region string, zone string) (*Zone, error) {
	return newZone(m.Store, table, region, zone)
}

func newZone(store *store.Store, table namespace.TableName, region string, zone string) (*Zone, error) {

	keyIndex, err := namespace.GetKeyForIndexRacks(table, region, zone)
	if err != nil {
		return nil, err
	}

	keyIndexEntry, err := namespace.GetKeyForIndexEntryZone(table, region, zone)
	if nil != err {
		return nil, err
	}

	key, err := namespace.GetKeyForZone(table, region, zone)
	if nil != err {
		return nil, err
	}

	z := &Zone{
		zoneNode: newZoneNode(store, keyIndexEntry, key, table, keyIndex, region, zone),
	}

	z.itemStore.isp = z

	return z, nil
}

// SetDetails is used to attach some attribute information to the object.
//
// The attribute information is not persisted to the store until an Update()
// call is made.
//
// The current revision of the zone object is reset
//
func (z *Zone) SetDetails(details *pb.ZoneDetails) {
	z.details = details.Clone()
	z.resetRevision()
}

// GetDetails is used to extract the attribute information from the object. The
// attribute information must have been previously read from the store (see
// the Read() method) or attached via a SetDetails() call.
//
// May return nil if there are no attributes currently held in the object.
//
func (z *Zone) GetDetails() *pb.ZoneDetails {
	return z.details.Clone()
}

// GetDefinitionZone returns a copy of the rack definition based on the contents
// of the current object.
//
func (z *Zone) GetDefinitionZone() *pb.Definition_Zone {

	return &pb.Definition_Zone{
		Details: z.GetDetails(),
		Racks:   make(map[string]*pb.Definition_Rack),
	}
}

// Equal is used to provide a simple equality check for use to determine
// if the current zone matches the supplied definition. Typically used
// when looking to see if the record has been changed.
//
// Note: only considers information relating to the parent object and does
//       not include comparisons for any descendants.
//
func (z *Zone) Equal(d *pb.Definition_Zone) bool {
	return z.details.Equal(d.GetDetails())
}

// NotEqual is used to provide a simple equality check for use to determine
// if the current zone does not match the supplied definition. Typically used
// when looking to see if the record has been changed.
//
// Note: only considers information relating to the parent object and does
//       not include comparisons for any descendants.
//
func (z *Zone) NotEqual(d *pb.Definition_Zone) bool {
	return !z.Equal(d)
}

// Load converts the string representation that was previously returned from a
// Save call back into the view's specific content, such that it can be used as
// part of normal operations.
func (z *Zone) Load(view ViewType, value string) error {
	if view != ViewDefinition {
		return errors.ErrUnimplementedView
	}

	record := &pb.Store_RecordDefinition_Zone{}

	if err := store.Decode(value, record); err != nil {
		return err
	}

	z.details = record.Details
	return nil
}

// Save returns the view's specific content as a string which can be stored
// for later retrieval.  If no view data is available, an error is returned.
func (z *Zone) Save(view ViewType) (string, error) {
	if view != ViewDefinition {
		return "", errors.ErrUnimplementedView
	}

	if z.details == nil {
		return "", errors.ErrDetailsNotAvailable("zone")
	}

	record := &pb.Store_RecordDefinition_Zone{
		Details: z.GetDetails(),
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
func (z *Zone) NewChild(name string) (*Rack, error) {

	return newRack(z.Store, z.Table, z.Region, z.Zone, name)
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

		_, err = child.Read(ctx, ViewDefinition)

		if err != nil {
			return store.RevisionInvalid, nil, err
		}

		children[v] = *child
	}

	return rev, &children, nil
}

// Watch is used to establish a watchpoint on a zone such that any updates
// to any names within the zone will generate a notification via the Event
// channel in the returned inventory.Watch structure.
//
// Once the watchpoint is no longer required, the caller should invoke the
// Close() method on the returned inventory.Watch structure.
//
func (z *Zone) Watch(ctx context.Context) (*Watch, error) {

	storeWatch, err := z.Store.Watch(ctx, namespace.KeyRootInventory, z.Key)

	if err != nil {
		return nil, err
	}

	notifications := make(chan WatchEvent)

	go func ()  {
		for ev := range storeWatch.Events {
			var we WatchEvent

			addr, err := namespace.GetAddressFromKey(ev.Key)

			if err != nil {
				_ = tracing.Error(ctx, "Invalid key format in watch event channel for key: %s", ev.Key)

				we = WatchEvent{
					Err:      err,
					Type:     ev.Type,
					Revision: ev.Revision,
					NewRev:   ev.NewRev,
					OldRev:   ev.OldRev,
					NewVal:   ev.NewVal,
					OldVal:   ev.OldVal,
				}
			} else {
				we = WatchEvent{
					Address:  addr,
					Type:     ev.Type,
					Revision: ev.Revision,
					NewRev:   ev.NewRev,
					OldRev:   ev.OldRev,
					NewVal:   ev.NewVal,
					OldVal:   ev.OldVal,
				}
			}

			notifications <- we
		}

		close(notifications)
	}()

	response := &Watch{
		watch:  storeWatch,
		Events: notifications,
	}

	return response, nil
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

