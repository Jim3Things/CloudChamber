package inventory

import (
	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/namespace"
	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/store"
	"github.com/Jim3Things/CloudChamber/simulation/pkg/errors"
	pb "github.com/Jim3Things/CloudChamber/simulation/pkg/protos/inventory"
)

// Tor is a structure representing a tor object. This object can be used
// to operate on the associated tor records in the underlying store. The
// object can hold information fetched from the underlying store, or as
// a staging area in preparation for updates to the store.
//
// Tor is a specialization of a child object for a rack parent.
//
type Tor struct {
	rackElement

	revisionInfo
    itemStoreLeaf

	details *pb.TorDetails
	ports   map[int64]*pb.NetworkPort
}

// NewTor is a convenience function used to construct a Tor object
// from scratch rather than relative to its parent.
//
func (m *Inventory) NewTor(
	table namespace.TableName,
	region string,
	zone string,
	rack string,
	id int64) (*Tor, error) {

	return newTor(m.Store, table, region, zone, rack, id)
}

func newTor(
	store *store.Store,
	table namespace.TableName,
	region string,
	zone string,
	rack string,
	id int64) (*Tor, error) {

	keyIndexEntry, err := namespace.GetKeyForIndexEntryTor(table, region, zone, rack, id)
	if nil != err {
		return nil, err
	}

	key, err := namespace.GetKeyForTor(table, region, zone, rack, id)
	if nil != err {
		return nil, err
	}

	t := &Tor{
		rackElement: newRackElement(store, keyIndexEntry, key, table, region, zone, rack, id),
	}

	t.itemStoreLeaf.isp = t

	return t, nil
}

func cloneNetworkPorts(ports map[int64]*pb.NetworkPort) map[int64]*pb.NetworkPort {
	if ports == nil {
		return nil
	}

	portMap := make(map[int64]*pb.NetworkPort, len(ports))

	for k, p := range ports {
		portMap[k] = p.Clone()
	}

	return portMap
}

// SetDetails is used to attach some attribute information to the object.
//
// The attribute information is not persisted to the store until an Update()
// call is made.
//
// The current revision of the tor object is reset
//
func (t *Tor) SetDetails(details *pb.TorDetails) {
	t.details = details.Clone()
	t.resetRevision()
}

// GetDetails is used to extract the attribute information from the object. The
// attribute information must have been previously read from the store (see
// the Read() method) or attached via a SetDetails() call.
//
// May return nil if there are no attributes currently held in the object.
//
func (t *Tor) GetDetails() *pb.TorDetails {
	return t.details.Clone()
}

// SetPorts is used to attach some network port information to the object.
//
// The port information is not persisted to the store until an Update()
// call is made.
//
// The current revision of the tor object is reset
//
func (t *Tor) SetPorts(ports map[int64]*pb.NetworkPort) {
	t.ports = cloneNetworkPorts(ports)
	t.resetRevision()
}

// GetPorts is used to extract the network port information from the object.
// The port information must have been previously read from the store (see
// the Read() method) or attached via a SetPorts() call.
//
// May return nil if there are no network port information currently held
// in the object.
//
func (t *Tor) GetPorts() map[int64]*pb.NetworkPort {
	return cloneNetworkPorts(t.ports)
}

// EqualPorts is used to provide a simple equality check for use to determine
// if the current ports match those supplied. Typically used when looking
// to see if the record has been changed.
//
func (t *Tor) EqualPorts(ports map[int64]*pb.NetworkPort) bool {
	if len(t.ports) != len(ports) {
		return false
	}

	for i, np := range t.ports {
		if !np.Equal(ports[i]) {
			return false
		}
	}

	return true
}

// GetDefinitionTor returns a copy of the tor definition based on the contents of the
// current object.
//
func (t *Tor) GetDefinitionTor() *pb.Definition_Tor {
	tor := &pb.Definition_Tor{
		Details: t.GetDetails(),
		Ports:   t.GetPorts(),
	}

	return tor
}

// Equal is used to provide a simple equality check for use to determine
// if the current tor matches the supplied definition. Typically used
// when looking to see if the record has been changed.
//
func (t *Tor) Equal(d *pb.Definition_Tor) bool {
	return t.details.Equal(d.GetDetails()) && t.EqualPorts(d.GetPorts())
}

// NotEqual is used to provide a simple equality check for use to determine
// if the current tor does not match the supplied definition. Typically used
// when looking to see if the record has been changed.
//
func (t *Tor) NotEqual(d *pb.Definition_Tor) bool {
	return !t.Equal(d)
}

// Load converts the string representation that was previously returned from a
// Save call back into the view's specific content, such that it can be used as
// part of normal operations.
func (t *Tor) Load(view ViewType, value string) error {
	if view != ViewDefinition {
		return errors.ErrUnimplementedView
	}

	record := &pb.Store_RecordDefinition_Tor{}

	if err := store.Decode(value, record); err != nil {
		return err
	}

	t.details = record.GetDetails()
	t.ports = record.GetPorts()

	return nil
}

// Save returns the view's specific content as a string which can be stored
// for later retrieval.  If no view data is available, an error is returned.
func (t *Tor) Save(view ViewType) (string, error) {
	if view != ViewDefinition {
		return "", errors.ErrUnimplementedView
	}

	if t.details == nil {
		return "", errors.ErrDetailsNotAvailable("tor")
	}

	if t.ports == nil {
		return "", errors.ErrPortsNotAvailable("tor")
	}

	record := t.GetDefinitionTor()

	return store.Encode(record)
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
