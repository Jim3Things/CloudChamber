package inventory

// This module contains the support structures and interfaces for handling the
// key and value components of the storage indices associated with the simulated
// inventory items in CloudChamber

import (
	"fmt"

	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/namespace"
	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/store"
	"github.com/Jim3Things/CloudChamber/simulation/pkg/errors"
)

// keyProvider specifies the functions that a structure that has a location in
// the store must provide in order to support the addressing functions.
//
// Note that it may be that view either won't be required, or that we'll never
// return an error...  Both would simplify use.
type keyProvider interface {
	// ChildIndexName supplies the name of the index for the set of children of
	// the instance with the specified view.  If no index exists, an error is
	// returned.
	ChildIndexName(view ViewType) (string, error)

	// IndexName supplies the name of the index for this instance given the
	// specified view.  It returns an error if no such index exists.
	IndexName(view ViewType) (string, error)

	// IndexValue supplies the value of the associated index for this instance.
	IndexValue() string

	// KeyName supplies the string used as a key to find this instance.
	KeyName(view ViewType) (string, error)

	// DbStore supplies the data store currently in use.
	DbStore() *store.Store

	// TableName supplies the store's table name to use when searching, reading,
	// or writing this instance.
	TableName() namespace.TableName
}

// baseIndex holds the common index substructures and capabilities that the more
// specific index types derive from
type baseIndex struct {
	Store         *store.Store
	KeyIndexEntry string
	Key           string
	Table         namespace.TableName
}

func newBaseIndex(
	store *store.Store,
	keyIndexEntry string,
	key string,
	table namespace.TableName) baseIndex {
	return baseIndex{
		Store:         store,
		KeyIndexEntry: keyIndexEntry,
		Key:           key,
		Table:         table,
	}
}

func (x *baseIndex) ChildIndexName(_ ViewType) (string, error) {
	return "", errors.ErrFunctionNotAvailable
}

func (x *baseIndex) IndexName(view ViewType) (string, error) {
	if view != ViewDefinition {
		return "", errors.ErrUnimplementedView
	}

	return x.KeyIndexEntry, nil
}

func (x *baseIndex) IndexValue() string { return "invalid" }

func (x *baseIndex) KeyName(view ViewType) (string, error) {
	if view != ViewDefinition {
		return "", errors.ErrUnimplementedView
	}

	return x.Key, nil
}

func (x *baseIndex) DbStore() *store.Store          { return x.Store }
func (x *baseIndex) TableName() namespace.TableName { return x.Table }

// regionNode is the topmost index substructure.  It extends the baseIndex by
// the addition of region key information and an index holding all children.
type regionNode struct {
	baseIndex

	KeyChildIndex string
	Region        string
}

func newRegionNode(
	store *store.Store,
	keyIndexEntry string,
	key string,
	table namespace.TableName,
	keyChildIndex string,
	region string) regionNode {
	return regionNode{
		baseIndex:     newBaseIndex(store, keyIndexEntry, key, table),
		KeyChildIndex: keyChildIndex,
		Region:        region,
	}
}

func (re *regionNode) ChildIndexName(view ViewType) (string, error) {
	if view != ViewDefinition {
		return "", errors.ErrUnimplementedView
	}

	return re.KeyChildIndex, nil
}

func (re *regionNode) IndexValue() string { return re.Region }

// zoneNode holds the index structures for zones, which are children of regions.
type zoneNode struct {
	regionNode
	Zone string
}

func newZoneNode(
	store *store.Store,
	keyIndexEntry string,
	key string,
	table namespace.TableName,
	keyChildIndex string,
	region string,
	zone string) zoneNode {
	return zoneNode{
		regionNode: newRegionNode(store, keyIndexEntry, key, table, keyChildIndex, region),
		Zone:       zone,
	}
}

func (re *zoneNode) IndexValue() string { return re.Zone }

// rackNode holds the index structures for racks, which are children of zones.
// Note that racks have a number of different types of children, rather than a
// single common type.  Because of this, it prohibits the use of the common
// ChildIndexName function.
type rackNode struct {
	zoneNode
	KeyIndexPdu   string
	KeyIndexTor   string
	KeyIndexBlade string

	Rack string
}

func newRackNode(
	store *store.Store,
	keyIndexEntry string,
	key string,
	table namespace.TableName,
	region string,
	zone string,
	keyIndexPdu string,
	keyIndexTor string,
	keyIndexBlade string,
	rack string) rackNode {
	return rackNode{
		zoneNode:      newZoneNode(store, keyIndexEntry, key, table, "", region, zone),
		KeyIndexPdu:   keyIndexPdu,
		KeyIndexTor:   keyIndexTor,
		KeyIndexBlade: keyIndexBlade,
		Rack:          rack,
	}
}

func (re *rackNode) ChildIndexName(_ ViewType) (string, error) {
	return "", errors.ErrFunctionNotAvailable
}

func (re *rackNode) IndexValue() string { return re.Rack }

// rackElement holds the index structures used by the various elements that are
// placed in racks, such as PDUs, TORs, or blades.  These are leaf entries that
// do not have any further children.
type rackElement struct {
	zoneNode
	Rack string

	ID int64
}

func newRackElement(
	store *store.Store,
	keyIndexEntry string,
	key string,
	table namespace.TableName,
	region string,
	zone string,
	rack string,
	ID int64) rackElement {

	return rackElement{
		zoneNode: newZoneNode(store, keyIndexEntry, key, table, "", region, zone),
		Rack:     rack,
		ID:       ID,
	}
}

func (re *rackElement) ChildIndexName(_ ViewType) (string, error) {
	return "", errors.ErrFunctionNotAvailable
}

func (re *rackElement) IndexValue() string { return fmt.Sprintf("%d", re.ID) }
