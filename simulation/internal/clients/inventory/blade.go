package inventory

import (
	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/namespace"
	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/store"
	"github.com/Jim3Things/CloudChamber/simulation/pkg/errors"
	"github.com/Jim3Things/CloudChamber/simulation/pkg/protos/inventory"
)

// Blade is a structure representing a blade object. This object can be
// used to operate on the associated blade records in the underlying
// store. The object can hold information fetched from the underlying
// store, or as a staging area in preparation for updates to the store.
//
// Blade is a specialization of a child object for a rack parent.
//
type Blade struct {
	rackElement

	revisionInfo
	itemStoreLeaf

	details       *inventory.BladeDetails
	capacity      *inventory.BladeCapacity
	bootInfo      *inventory.BladeBootInfo
	bootOnPowerOn bool
}

// NewBlade is a convenience function used to construct a Blade object
// from scratch rather than relative to its parent.
//
func (m *Inventory) NewBlade(
	table namespace.TableName,
	region string,
	zone string,
	rack string,
	id int64) (*Blade, error) {

	return newBlade(m.Store, table, region, zone, rack, id)
}

func newBlade(
	store *store.Store,
	table namespace.TableName,
	region string,
	zone string,
	rack string,
	id int64) (*Blade, error) {

	keyIndexEntry, err := namespace.GetKeyForIndexEntryBlade(table, region, zone, rack, id)
	if nil != err {
		return nil, err
	}

	key, err := namespace.GetKeyForBlade(table, region, zone, rack, id)
	if nil != err {
		return nil, err
	}

	b := &Blade{
		rackElement: newRackElement(store, keyIndexEntry, key, table, region, zone, rack, id),
	}

	b.itemStoreLeaf.provider = b

	return b, nil
}

// SetDetails is used to attach some attribute information to the object.
//
// The attribute information is not persisted to the store until an Update()
// call is made.
//
// The current revision of the blade object is reset
//
func (b *Blade) SetDetails(details *inventory.BladeDetails) {
	b.details = details.Clone()
	b.resetRevision()
}

// GetDetails is used to extract the attribute information from the object. The
// attribute information must have been previously read from the store (see
// the Read() method) or attached via a SetDetails() call.
//
// May return nil if there are no attributes currently held in the object.
//
func (b *Blade) GetDetails() *inventory.BladeDetails {
	return b.details.Clone()
}

// SetCapacity is used to attach some capacity information to the object.
//
// The capacity information is not persisted to the store until an Update()
// call is made.
//
// The current revision of the blade object is reset
//
func (b *Blade) SetCapacity(capacity *inventory.BladeCapacity) {
	b.capacity = capacity.Clone()
	b.resetRevision()
}

// GetCapacity is used to extract the capacity information from the object.
// The capacity information must have been previously read from the store (see
// the Read() method) or attached via a SetCapacity() call.
//
// May return nil if there are no capacity information currently held
// in the object.
//
func (b *Blade) GetCapacity() *inventory.BladeCapacity {
	return b.capacity.Clone()
}

// SetBootInfo is used to attach some boot information to the object.
//
// The boot information is not persisted to the store until an Update()
// call is made.
//
// The current revision of the blade object is reset
//
func (b *Blade) SetBootInfo(bootInfo *inventory.BladeBootInfo) {
	b.bootInfo = bootInfo.Clone()
	b.resetRevision()
}

// GetBootInfo is used to extract the boot information from the object.
// The boot information must have been previously read from the store (see
// the Read() method) or attached via a SetBootInfo() call.
//
// May return nil if there are no boot information currently held
// in the object.
//
func (b *Blade) GetBootInfo() *inventory.BladeBootInfo {
	return b.bootInfo.Clone()
}

// SetBootPowerOn is used to set the boot power on flag
//
func (b *Blade) SetBootPowerOn(bootOnPowerOn bool) {
	b.bootOnPowerOn = bootOnPowerOn
	b.resetRevision()
}

// GetBootOnPowerOn returns a new copy of the BootOnPowerOn field
// within the blade object.
//
func (b *Blade) GetBootOnPowerOn() bool {
	return b.bootOnPowerOn
}

// GetDefinitionBlade returns a copy of the blade definition based on the contents of the
// current object.
//
func (b *Blade) GetDefinitionBlade() *inventory.Definition_Blade {
	return &inventory.Definition_Blade{
		Details:       b.GetDetails(),
		Capacity:      b.GetCapacity(),
		BootInfo:      b.GetBootInfo(),
		BootOnPowerOn: b.GetBootOnPowerOn(),
	}
}

// Equal is used to provide a simple equality check for use to determine
// if the current blade matches the supplied definition. Typically used
// when looking to see if the record has been changed.
//
func (b *Blade) Equal(d *inventory.Definition_Blade) bool {
	return b.details.Equal(d.GetDetails()) &&
		b.capacity.Equal(d.GetCapacity()) &&
		b.bootInfo.Equal(d.GetBootInfo()) &&
		b.GetBootOnPowerOn() == d.GetBootOnPowerOn()
}

// NotEqual is used to provide a simple equality check for use to determine
// if the current blade matches the supplied definition. Typically used
// when looking to see if the record has been changed.
//
func (b *Blade) NotEqual(d *inventory.Definition_Blade) bool {
	return !b.Equal(d)
}

// Load converts the string representation that was previously returned from a
// Save call back into the view's specific content, such that it can be used as
// part of normal operations.
func (b *Blade) Load(view ViewType, value string) error {
	if view != ViewDefinition {
		return errors.ErrUnimplementedView
	}

	record := &inventory.Store_RecordDefinition_Blade{}

	if err := store.Decode(value, record); err != nil {
		return err
	}

	b.details = record.GetDetails()
	b.capacity = record.GetCapacity()
	b.bootInfo = record.GetBootInfo()
	b.bootOnPowerOn = record.GetBootOnPowerOn()

	return nil
}

// Save returns the view's specific content as a string which can be stored
// for later retrieval.  If no view data is available, an error is returned.
func (b *Blade) Save(view ViewType) (string, error) {
	if view != ViewDefinition {
		return "", errors.ErrUnimplementedView
	}

	if b.details == nil {
		return "", errors.ErrDetailsNotAvailable("blade")
	}

	if b.capacity == nil {
		return "", errors.ErrCapacityNotAvailable("blade")
	}

	if b.bootInfo == nil {
		return "", errors.ErrBootInfoNotAvailable("blade")
	}

	record := b.GetDefinitionBlade()

	return store.Encode(record)
}

// mapErrStoreValue is a helper function that converts generic store access
// errors to equivalent view-specific ones, or returns the original error if
// no mapping for the supplied error is found.
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
