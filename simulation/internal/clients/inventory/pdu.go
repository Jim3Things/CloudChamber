package inventory

import (
	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/namespace"
	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/store"
	"github.com/Jim3Things/CloudChamber/simulation/pkg/errors"
	pb "github.com/Jim3Things/CloudChamber/simulation/pkg/protos/inventory"
)

// Pdu is a structure representing a pdu object. This object can be used
// to operate on the associated pdu records in the underlying store. The
// object can hold information fetched from the underlying store, or as
// a staging area in preparation for updates to the store.
//
// Pdu is a specialization of a child object for a rack parent.
//
type Pdu struct {
	rackElement

	revisionInfo
    itemStoreLeaf

	details *pb.PduDetails
	ports   map[int64]*pb.PowerPort
}

// NewPdu is a convenience function used to construct a Pdu object
// from scratch rather than relative to its parent.
//
func (m *Inventory) NewPdu(
	table namespace.TableName,
	region string,
	zone string,
	rack string,
	id int64) (*Pdu, error) {

	return newPdu(m.Store, table, region, zone, rack, id)
}

func newPdu(
	store *store.Store,
	table namespace.TableName,
	region string,
	zone string,
	rack string,
	id int64) (*Pdu, error) {

	keyIndexEntry, err := namespace.GetKeyForIndexEntryPdu(table, region, zone, rack, id)

	if nil != err {
		return nil, err
	}

	key, err := namespace.GetKeyForPdu(table, region, zone, rack, id)

	if nil != err {
		return nil, err
	}

	p := &Pdu{
		rackElement: newRackElement(store, keyIndexEntry, key, table, region, zone, rack, id),
	}

	p.itemStoreLeaf.isp = p

	return p, nil
}

func clonePowerPorts(ports map[int64]*pb.PowerPort) map[int64]*pb.PowerPort {
	if ports == nil {
		return nil
	}

	portMap := make(map[int64]*pb.PowerPort, len(ports))

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
// The current revision of the pdu object is reset
//
func (p *Pdu) SetDetails(details *pb.PduDetails) {
	p.details = details.Clone()
	p.resetRevision()
}

// GetDetails is used to extract the attribute information from the object. The
// attribute information must have been previously read from the store (see
// the Read() method) or attached via a SetDetails() call.
//
// May return nil if there are no attributes currently held in the object.
//
func (p *Pdu) GetDetails() *pb.PduDetails {
	return p.details.Clone()
}

// SetPorts is used to attach some power port information to the object.
//
// The port information is not persisted to the store until an Update()
// call is made.
//
// The current revision of the pdu object is reset
//
func (p *Pdu) SetPorts(ports map[int64]*pb.PowerPort) {
	p.ports = clonePowerPorts(ports)
	p.resetRevision()
}

// GetPorts is used to extract the power port information from the object.
// The port information must have been previously read from the store (see
// the Read() method) or attached via a SetPorts() call.
//
// May return nil if there are no power port information currently held
// in the object.
//
func (p *Pdu) GetPorts() map[int64]*pb.PowerPort {
	return clonePowerPorts(p.ports)
}

// EqualPorts is used to provide a simple equality check for use to determine
// if the current ports match those supplied. Typically used when looking
// to see if the record has been changed.
//
func (p *Pdu) EqualPorts(ports map[int64]*pb.PowerPort) bool {
	if len(p.ports) != len(ports) {
		return false
	}

	for i, pp := range p.ports {
		if !pp.Equal(ports[i]) {
			return false
		}
	}

	return true
}

// GetDefinitionPdu returns a copy of the pdu definition based on the contents of the
// current object.
//
func (p *Pdu) GetDefinitionPdu() *pb.Definition_Pdu {
	pdu := &pb.Definition_Pdu{
		Details: p.GetDetails(),
		Ports:   p.GetPorts(),
	}

	return pdu
}

// Equal is used to provide a simple equality check for use to determine
// if the current pdu matches the supplied definition. Typically used when
// looking to see if the record has been changed.
//
func (p *Pdu) Equal(d *pb.Definition_Pdu) bool {
	return p.details.Equal(d.GetDetails()) && p.EqualPorts(d.GetPorts())
}

// NotEqual is used to provide a simple equality check for use to determine
// if the current pdu do not match the supplied definition. Typically used when
// looking to see if the record has been changed.
//
func (p *Pdu) NotEqual(d *pb.Definition_Pdu) bool {
	return !p.Equal(d)
}

// Load converts the string representation that was previously returned from a
// Save call back into the view's specific content, such that it can be used as
// part of normal operations.
func (p *Pdu) Load(view ViewType, value string) error {
	if view != ViewDefinition {
		return errors.ErrUnimplementedView
	}

	record := &pb.Store_RecordDefinition_Pdu{}

	if err := store.Decode(value, record); err != nil {
		return err
	}

	p.details = record.GetDetails()
	p.ports = record.GetPorts()

	return nil
}

// Save returns the view's specific content as a string which can be stored
// for later retrieval.  If no view data is available, an error is returned.
func (p *Pdu) Save(view ViewType) (string, error) {
	if view != ViewDefinition {
		return "", errors.ErrUnimplementedView
	}

	if p.details == nil {
		return "", errors.ErrDetailsNotAvailable("pdu")
	}

	if p.ports == nil {
		return "", errors.ErrPortsNotAvailable("pdu")
	}

	record := p.GetDefinitionPdu()

	return store.Encode(record)
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
