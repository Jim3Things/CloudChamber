// This package is used to access the persisted inventory state in an underlying
// store and to provide an abstraction to the other CloudChamber services such that
// they need not concern themselves with how the data is laid out or manipulated.
//
// The access method is based around a 'cursor' object which can be used to
// operate on the underlying persisted version of that object just by
// using a few basic access methods, e.g. Read, Update, ListChildren etc.
//
// The objects providing such a cursor are regions, zones, racks, pdus, tor and
// blades. There is also a root object which provides a well-known entry point
// and can be used to start the discovery process. These object form a hierarchy
// which allows for navigating from parent objects to an associated child.
//
// The complete inventory is split into a number of separate namespaces (or tables)
// which are used to compartmentalize different usage groups within the inventory.
// For example, the primary namespaces are
//
//		definition
//		observed
//		actual
//		target
//
// See the CloudChamber documentation for a description of these namespaces and
// their uses.
//
//
// The objects within a table form a hierarchy of navigable "nodes" from root,
// to region, to zone to rack. Within a rack there are the further child objects
// for pdus, tors and blades. These last child objects do not themselves have
// children and are accessed as complete entities which are operated on as an
// atomic unit.
//
// NOTE: At least so far. It is conceivable that we may wish to extend this
//       to allow individual entries within these leaf items to be manipulated
//       directly, e.g. we may wish to update the state of a network port
//       within a tor without having to re-write the complete tor object.
//
// To allow for object discovery, the inventory allows a caller to request a
// list of all the children of a navigable node, e.g. search for all of the
// racks within a zone. To achieve this, the pacakge maintains an index for
// child objects which is kept separately from the objects themselves. Thus
// when creating a new object, the package will also create an appropriate
// index entry to allow that object to be discovered from its parent. In
// this way, by knowing the root of a namespace, all the regions can be
// located. Once a region is know, all the zones within that region can be
// located. And so on for racks within zones, etc.
//
// NOTE: Currently there is no way to discover a parent for a given oject.
//       However, each object contains sufficient information that a parent
//       can be readily identified if required which would allow this feature
//       to be added if needed.

package inventory

import (
	"context"
	"fmt"
	"sync"

	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/store"
	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/timestamp"
	"github.com/Jim3Things/CloudChamber/simulation/internal/common"
	"github.com/Jim3Things/CloudChamber/simulation/internal/config"
	"github.com/Jim3Things/CloudChamber/simulation/internal/tracing"
	"github.com/Jim3Things/CloudChamber/simulation/pkg/errors"
	pb "github.com/Jim3Things/CloudChamber/simulation/pkg/protos/inventory"
)

const (

	// DefaultRegion is used to provide a value for the non-existing region name
	// while the transition to the new inventory extended schemaa continues.
	// Eventually these will dissapear as the front-end and higher layers learn
	// abouts regions, zones, multiple pdus and tors.
	//
	DefaultRegion = "standard"

	// DefaultZone is used to provide a value for the non-existing zone name
	// while the transition to the new inventory extended schemaa continues.
	// Eventually these will dissapear as the front-end and higher layers learn
	// abouts regions, zones, multiple pdus and tors.
	//
	DefaultZone   = "standard"

	// DefaultPdu is used to provide a value for the non-existing pdu ID
	// while the transition to the new inventory extended schemaa continues.
	// Eventually these will dissapear as the front-end and higher layers learn
	// abouts regions, zones, multiple pdus and tors.
	//
	DefaultPdu    = int64(0)

	// DefaultTor is used to provide a value for the non-existing tor ID
	// while the transition to the new inventory extended schemaa continues.
	// Eventually these will dissapear as the front-end and higher layers learn
	// abouts regions, zones, multiple pdus and tors.
	//
	DefaultTor    = int64(0)

	// DefinitionTable is a constant to indicate the inventory operation should be
	// performed against the inventory definition table for the item of interest.
	//
	DefinitionTable = "definition"

	// DefinitionTableStdTest is a constant to indicate the inventory operation should be
	// performed against the test inventory definition table for the standard test
	// inventory for the item of interest.
	//
	DefinitionTableStdTest = "definitionstdtest"

	// ActualTable is a constant to indicate the inventory operation should be
	// performed against the inventory actual state table for the item of interest.
	//
	ActualTable = "actual"

	// ObservedTable is a constant to indicate the inventory operation should be
	// performed against the inventory observed state table for the item of interest.
	//
	ObservedTable = "observed"

	// TargetTable is a constant to indicate the inventory operation should be
	// performed against the inventory target state table for the item of interest.
	//
	TargetTable = "target"

	prefixRegion = "region"
	prefixZone   = "zone"
	prefixRack   = "rack"
	prefixBlade  = "blade"
	prefixPdu    = "pdu"
	prefixTor    = "tor"

	keyFormatIndexRegions = "%s/index/" + prefixRegion + "s/"
	keyFormatIndexZones   = "%s/index/" + prefixRegion + "/%s/" + prefixZone + "s/"
	keyFormatIndexRacks   = "%s/index/" + prefixRegion + "/%s/" + prefixZone + "/%s/" + prefixRack + "s/"
	keyFormatIndexPdus    = "%s/index/" + prefixRegion + "/%s/" + prefixZone + "/%s/" + prefixRack + "/%s/" + prefixPdu + "s/"
	keyFormatIndexTors    = "%s/index/" + prefixRegion + "/%s/" + prefixZone + "/%s/" + prefixRack + "/%s/" + prefixTor + "s/"
	keyFormatIndexBlades  = "%s/index/" + prefixRegion + "/%s/" + prefixZone + "/%s/" + prefixRack + "/%s/" + prefixBlade + "s/"

	keyFormatIndexEntryRegion = keyFormatIndexRegions + "%s"
	keyFormatIndexEntryZone   = keyFormatIndexZones + "%s"
	keyFormatIndexEntryRack   = keyFormatIndexRacks + "%s"
	keyFormatIndexEntryPdu    = keyFormatIndexPdus + "%d"
	keyFormatIndexEntryTor    = keyFormatIndexTors + "%d"
	keyFormatIndexEntryBlade  = keyFormatIndexBlades + "%d"

	keyFormatRegion = "%s/data/" + prefixRegion + "/%s"
	keyFormatZone   = "%s/data/" + prefixRegion + "/%s/" + prefixZone + "/%s"
	keyFormatRack   = "%s/data/" + prefixRegion + "/%s/" + prefixZone + "/%s/" + prefixRack + "/%s"
	keyFormatPdu    = "%s/data/" + prefixRegion + "/%s/" + prefixZone + "/%s/" + prefixRack + "/%s/" + prefixPdu + "/%d"
	keyFormatTor    = "%s/data/" + prefixRegion + "/%s/" + prefixZone + "/%s/" + prefixRack + "/%s/" + prefixTor + "/%d"
	keyFormatBlade  = "%s/data/" + prefixRegion + "/%s/" + prefixZone + "/%s/" + prefixRack + "/%s/" + prefixBlade + "/%d"

	// MaxBladeID is the highest blade number accepted as valid. This is
	// an arbitrary choice intended to help prevent configuration issues.
	//
	MaxBladeID = int64(10 * 1000 * 1000)

	// MaxPduID defines the larget number of Pdus that can be configured within
	// a single rack.
	//
	MaxPduID = int64(2)

	// MaxTorID defines the larget number of PTorsdus that can be configured within
	// a single rack.
	//
	MaxTorID = int64(2)
)

func verifyTable(table string) error {
	switch table {
	case DefinitionTable:
		return nil
	case ActualTable:
		return nil
	case ObservedTable:
		return nil
	case TargetTable:
		return nil

	// Special case for a namespace only ever expected to be used
	// in CloudChamber tests
	//
	case DefinitionTableStdTest:
		return nil

	case "":
		return errors.ErrTableNameMissing(table)

	default:
		return errors.ErrTableNameInvalid{
			Name:            table,
			DefinitionTable: DefinitionTable,
			ActualTable:     ActualTable,
			ObservedTable:   ObservedTable,
			TargetTable:     TargetTable,
		}
	}
}

func verifyRegion(val string) error {

	if "" == val {
		return errors.ErrRegionNameMissing(val)
	}

	return nil
}

func verifyZone(val string) error {

	if "" == val {
		return errors.ErrZoneNameMissing(val)
	}

	return nil
}

func verifyRack(val string) error {

	if "" == val {
		return errors.ErrRackNameMissing(val)
	}

	return nil
}

func verifyPdu(val int64) error {

	if val < 0 || val > MaxPduID {
		return errors.ErrPduIDInvalid{Value: val, Limit: MaxPduID}
	}

	return nil
}

func verifyTor(val int64) error {

	if val < 0 || val > MaxTorID {
		return errors.ErrTorIDInvalid{Value: val, Limit: MaxTorID}
	}

	return nil
}

func verifyBlade(val int64) error {

	if val < 0 || val > MaxBladeID {
		return errors.ErrBladeIDInvalid{Value: val, Limit: MaxBladeID}
	}

	return nil
}

func verifyNamesRegion(table string, region string) error {

	if err := verifyTable(table); err != nil {
		return err
	}

	if err := verifyRegion(region); err != nil {
		return err
	}

	return nil
}

func verifyNamesZone(table string, region string, zone string) error {

	if err := verifyNamesRegion(table, region); err != nil {
		return err
	}

	if err := verifyZone(zone); err != nil {
		return err
	}

	return nil
}

func verifyNamesRack(table string, region string, zone string, rack string) error {

	if err := verifyNamesZone(table, region, zone); err != nil {
		return err
	}

	if err := verifyRack(rack); err != nil {
		return err
	}

	return nil
}

func verifyNamesPdu(table string, region string, zone string, rack string, index int64) error {

	if err := verifyNamesRack(table, region, zone, rack); err != nil {
		return err
	}

	if err := verifyPdu(index); err != nil {
		return err
	}

	return nil
}

func verifyNamesTor(table string, region string, zone string, rack string, index int64) error {

	if err := verifyNamesRack(table, region, zone, rack); err != nil {
		return err
	}

	if err := verifyTor(index); err != nil {
		return err
	}

	return nil
}

func verifyNamesBlade(table string, region string, zone string, rack string, index int64) error {

	if err := verifyNamesRack(table, region, zone, rack); err != nil {
		return err
	}

	if err := verifyBlade(index); err != nil {
		return err
	}

	return nil
}

// getKeyForIndexRegions generates the key to discover the list of regions within a
// specific table (definition, actual, observed, target)
//
func getKeyForIndexRegions(table string) (key string, err error) {

	if err = verifyTable(table); err != nil {
		return key, err
	}

	key = fmt.Sprintf(
		keyFormatIndexRegions,
		store.GetNormalizedName(table))

	return key, nil
}

// getKeyForIndexZones generates the key to discover the list of zones within a
// specific table (definition, actual, observed, target)
//
func getKeyForIndexZones(table string, region string) (key string, err error) {

	if err = verifyNamesRegion(table, region); err != nil {
		return key, err
	}

	key = fmt.Sprintf(
		keyFormatIndexZones,
		store.GetNormalizedName(table),
		store.GetNormalizedName(region))

	return key, nil
}

// getKeyForIndexRacks generates the key to discover the list of racks within a
// specific table (definition, actual, observed, target)
//
func getKeyForIndexRacks(table string, region string, zone string) (key string, err error) {

	if err = verifyNamesZone(table, region, zone); err != nil {
		return key, err
	}

	key = fmt.Sprintf(
		keyFormatIndexRacks,
		store.GetNormalizedName(table),
		store.GetNormalizedName(region),
		store.GetNormalizedName(zone))

	return key, nil
}

// GetKeyForIndexPdu generates the key to discover the list of pdus within a
// specific table (definition, actual, observed, target)
//
func getKeyForIndexPdus(table string, region string, zone string, rack string) (key string, err error) {

	if err = verifyNamesRack(table, region, zone, rack); err != nil {
		return key, err
	}

	key = fmt.Sprintf(
		keyFormatIndexPdus,
		store.GetNormalizedName(table),
		store.GetNormalizedName(region),
		store.GetNormalizedName(zone),
		store.GetNormalizedName(rack))

	return key, nil
}

// getKeyForIndexTors generates the key to discover the list of tors within a
// specific table (definition, actual, observed, target)
//
func getKeyForIndexTors(table string, region string, zone string, rack string) (key string, err error) {

	if err = verifyNamesRack(table, region, zone, rack); err != nil {
		return key, err
	}

	key = fmt.Sprintf(
		keyFormatIndexTors,
		store.GetNormalizedName(table),
		store.GetNormalizedName(region),
		store.GetNormalizedName(zone),
		store.GetNormalizedName(rack))

	return key, nil
}

// getKeyForIndexBlades generates the key to discover the list of blades within a
// specific table (definition, actual, observed, target)
//
func getKeyForIndexBlades(table string, region string, zone string, rack string) (key string, err error) {

	if err = verifyNamesRack(table, region, zone, rack); err != nil {
		return key, err
	}

	key = fmt.Sprintf(
		keyFormatIndexBlades,
		store.GetNormalizedName(table),
		store.GetNormalizedName(region),
		store.GetNormalizedName(zone),
		store.GetNormalizedName(rack))

	return key, nil
}

// getKeyForIndexEntryRegion generates the key to create an index entry for a region within a
// specific table (definition, actual, observed, target)
//
func getKeyForIndexEntryRegion(table string, region string) (key string, err error) {

	if err = verifyNamesRegion(table, region); err != nil {
		return key, err
	}

	key = fmt.Sprintf(
		keyFormatIndexEntryRegion,
		store.GetNormalizedName(table),
		store.GetNormalizedName(region))

	return key, nil
}

// getKeyForIndexEntryZone generates the key to create an index entry for a zone within a
// specific table (definition, actual, observed, target)
//
func getKeyForIndexEntryZone(table string, region string, zone string) (key string, err error) {

	if err = verifyNamesZone(table, region, zone); err != nil {
		return key, err
	}

	key = fmt.Sprintf(
		keyFormatIndexEntryZone,
		store.GetNormalizedName(table),
		store.GetNormalizedName(region),
		store.GetNormalizedName(zone))

	return key, nil
}

// getKeyForIndexEntryRack generates the key to create an index entry for a rack within a
// specific table (definition, actual, observed, target)
//
func getKeyForIndexEntryRack(table string, region string, zone string, rack string) (key string, err error) {

	if err = verifyNamesRack(table, region, zone, rack); err != nil {
		return key, err
	}

	key = fmt.Sprintf(
		keyFormatIndexEntryRack,
		store.GetNormalizedName(table),
		store.GetNormalizedName(region),
		store.GetNormalizedName(zone),
		store.GetNormalizedName(rack))

	return key, nil
}

// getKeyForIndexEntryPdu generates the key to create an index entry for a pdu within a
// specific table (definition, actual, observed, target)
//
func getKeyForIndexEntryPdu(table string, region string, zone string, rack string, pdu int64) (key string, err error) {

	if err = verifyNamesPdu(table, region, zone, rack, pdu); err != nil {
		return key, err
	}

	key = fmt.Sprintf(
		keyFormatIndexEntryPdu,
		store.GetNormalizedName(table),
		store.GetNormalizedName(region),
		store.GetNormalizedName(zone),
		store.GetNormalizedName(rack),
		pdu)

	return key, nil
}

// getKeyForIndexEntryTor generates the key to create an index entry for a tor within a
// specific table (definition, actual, observed, target)
//
func getKeyForIndexEntryTor(table string, region string, zone string, rack string, tor int64) (key string, err error) {

	if err = verifyNamesTor(table, region, zone, rack, tor); err != nil {
		return key, err
	}

	key = fmt.Sprintf(
		keyFormatIndexEntryTor,
		store.GetNormalizedName(table),
		store.GetNormalizedName(region),
		store.GetNormalizedName(zone),
		store.GetNormalizedName(rack),
		tor)

	return key, nil
}

// getKeyForIndexEntryBlade generates the key to create an index entry for a blade within a
// specific table (definition, actual, observed, target)
//
func getKeyForIndexEntryBlade(table string, region string, zone string, rack string, blade int64) (key string, err error) {

	if err = verifyNamesBlade(table, region, zone, rack, blade); err != nil {
		return key, err
	}

	key = fmt.Sprintf(
		keyFormatIndexEntryBlade,
		store.GetNormalizedName(table),
		store.GetNormalizedName(region),
		store.GetNormalizedName(zone),
		store.GetNormalizedName(rack),
		blade)

	return key, nil
}

// getKeyForRegion generates the key to operate on the record for a region within a
// specific table (definition, actual, observed, target)
//
func getKeyForRegion(table string, region string) (key string, err error) {

	if err = verifyNamesRegion(table, region); err != nil {
		return key, err
	}

	key = fmt.Sprintf(
		keyFormatRegion,
		store.GetNormalizedName(table),
		store.GetNormalizedName(region))

	return key, nil
}

// getKeyForZone generates the key to operate on the record for a zone within a
// specific table (definition, actual, observed, target)
//
func getKeyForZone(table string, region string, zone string) (key string, err error) {

	if err = verifyNamesZone(table, region, zone); err != nil {
		return key, err
	}

	key = fmt.Sprintf(
		keyFormatZone,
		store.GetNormalizedName(table),
		store.GetNormalizedName(region),
		store.GetNormalizedName(zone))

	return key, nil
}

// getKeyForRack generates the key to operate on the record for a rack within a
// specific table (definition, actual, observed, target)
//
func getKeyForRack(table string, region string, zone string, rack string) (key string, err error) {

	if err = verifyNamesRack(table, region, zone, rack); err != nil {
		return key, err
	}

	key = fmt.Sprintf(
		keyFormatRack,
		store.GetNormalizedName(table),
		store.GetNormalizedName(region),
		store.GetNormalizedName(zone),
		store.GetNormalizedName(rack))

	return key, nil
}

// getKeyForPdu generates the key to operate on the record for a pdu within a
// specific table (definition, actual, observed, target)
//
func getKeyForPdu(table string, region string, zone string, rack string, pdu int64) (key string, err error) {

	if err = verifyNamesPdu(table, region, zone, rack, pdu); err != nil {
		return key, err
	}

	key = fmt.Sprintf(
		keyFormatPdu,
		store.GetNormalizedName(table),
		store.GetNormalizedName(region),
		store.GetNormalizedName(zone),
		store.GetNormalizedName(rack),
		pdu)

	return key, nil
}

// getKeyForTor generates the key to operate on the record for a tor within a
// specific table (definition, actual, observed, target)
//
func getKeyForTor(table string, region string, zone string, rack string, tor int64) (key string, err error) {

	if err = verifyNamesTor(table, region, zone, rack, tor); err != nil {
		return key, err
	}

	key = fmt.Sprintf(
		keyFormatTor,
		store.GetNormalizedName(table),
		store.GetNormalizedName(region),
		store.GetNormalizedName(zone),
		store.GetNormalizedName(rack),
		tor)

	return key, nil
}

// getKeyForBlade generates the key to operate on the record for a blade within a
// specific table (definition, actual, observed, target)
//
func getKeyForBlade(table string, region string, zone string, rack string, blade int64) (key string, err error) {

	if err = verifyNamesBlade(table, region, zone, rack, blade); err != nil {
		return key, err
	}

	key = fmt.Sprintf(
		keyFormatBlade,
		store.GetNormalizedName(table),
		store.GetNormalizedName(region),
		store.GetNormalizedName(zone),
		store.GetNormalizedName(rack),
		blade)

	return key, nil
}

type inventoryRevision interface {
	// GetRevision returns the revision of the details field within the object.
	// This will be either the revision of the object in the store after a
	// Create(), Read() or Update() call or be store.RevisionInvalid if the
	// details have been set or no Create(), Read() or Update() call has been
	// executed.
	//
	GetRevision() int64

	// GetRevisionRecord returns the revision of the underlying store object as
	// determined at the time of the last Create(), Read() or Update() for the
	// object. The record revision is not reset by a SetDetails() call and is
	// used when performing either a conditional update or conditional delete
	// using the object.
	//
	GetRevisionRecord() int64

	// GetRevisionStore returns the revison of the underlying store ifself as
	// determined at the time of the last Create() Read() for the object. The
	// store revision is not reset by a SetDetails() call and is provided
	// for information only.
	//
	GetRevisionStore() int64

	// GetRevisionForRequest returns the appropriate revision for the update
	// for either a conditional update based upon the revision of the most
	// recently read record, or an unconditional update.
	//
	GetRevisionForRequest(unconditional bool) int64

	// resetRevision resets the revision for the details field within the object.
	// Subsequent calls to GetRevision() will return store.RevisionInvalid until
	// a successful call is made to one of the routines which invoke the store
	//
	resetRevision() int64

	// updateRevision is used to set/update the current revision information
	// as part of a successful invokation of a store routine.
	//
	updateRevisionInfo(rev int64) int64
}

// Region, zone and rack are "containers" whereas tor, pdu and blade are "things".
// You can send operations and commands to things, but not containers.
//
// Of the following SetXxx() and GetXxx() methods fetch or set values within the
// fields of an object (cursor) and do NOT perform any operations to the underlying
// store.
//
// the Create(), Read(), Update() and Delete() methods perform the appropriate
// operation to the persisted object based upon the current values of fields
// within the object (cursor) being used and will return an error if needed fields
// are not set.
//
type inventoryItem interface {
	inventoryRevision

	// Use to set the attribues of an object within the inventory
	//
	SetDetails(ctx context.Context, details *interface{})
	GetDetails(ctx context.Context) *interface{}

	// Create uses the current object to persist the object to the underlying
	// store  and also create any index entries that may be required.
	//
	// This is not valid to call on a root object and doing so will return
	// an error.
	//
	Create(ctx context.Context) (int64, error)

	// Read issues a request to the underlying store to populate all the fields
	// within the cursor object, including any attributes for that object or
	// other information specific to that object.
	//
	// This is not valid to call on a root object and doing so will return
	// an error.
	//
	Read(ctx context.Context) (int64, error)

	// Update will write a record the underlying store using the currenty
	// information  in the fields of the object. The update can be either
	// unconditional by setting the unconditional parameter to true, or
	// conditional based on the revision of the object compared to the
	// revision of the associated record in the underlying store.
	//
	// Note the object maintains revision information returned from the
	// store for any actions involving the store, e.g. Create(), Read() or
	// Update().
	//
	Update(ctx context.Context, unconditional bool) (int64, error)

	// Delete is used to remove the persisted copy of the object from the
	// store along with any index information needed to navigate to or
	// through that object. The delete can be either unconditional by
	// setting the unconditional parameter to true, or conditional based
	// on the revision of the object compared to the revision of the
	// associated record in the underlying store.
	//
	// Note the object maintains revision information returned from the
	// store for any actions involving the store, e.g. Create(), Read() or
	// Update().
	//
	Delete(ctx context.Context, unconditional bool) (int64, error)
}

// A node object within the inventory is one which in addition to having its own
// attributes also allows for navigation of the namespace
//
type inventoryItemNode interface {
	inventoryItem

	// NewChild creates a child of the current object but uses only on the
	// current object to create a new child object and no store
	// operations are involved.
	//
	// Note, this method does not populate the attributes of the new child
	// object. To retrieve the attributes a Read() using the child must be
	// performed.
	//
	NewChild(name string) (*interface{}, error)

	// ListChildren uses the current object to discover the names of all the
	// child objects of the current object, The elements of the returned list
	// can be used in subsequent operations to create child objects.
	//
	ListChildren(ctx context.Context) (int64, *[]string, error)

	// FetchChildren is used to locate all the children of the current object
	// and to generate an object for each of those children. It is a
	// convenience wrapper around ListChildren() followed by a NewChild() on
	// each name discovered.
	//
	FetchChildren(ctx context.Context) (int64, *map[string]interface{}, error)
}

// Provide a set of definitions to cope with calls to a "null" object.
//
func (n *nullItem) GetRevision() int64 {
	return store.RevisionInvalid
}

func (n *nullItem) GetRevisionRecord() int64 {
	return store.RevisionInvalid
}

func (n *nullItem) GetRevisionStore() int64 {
	return store.RevisionInvalid
}

func (n *nullItem) GetRevisionForRequest(unconditional bool) int64 {
	return store.RevisionInvalid
}

func (n *nullItem) resetRevision() int64 {
	return store.RevisionInvalid
}

func (n *nullItem) updateRevisionInfo(rev int64) int64 {
	return store.RevisionInvalid
}

type inventoryItemRack interface {
	inventoryItemNode

	NewPdu(name string) (*interface{}, error)
	NewTor(name string) (*interface{}, error)
	NewBlade(name string) (*interface{}, error)

	ListPdus(ctx context.Context) (int64, *map[int64]*interface{}, error)
	ListTors(ctx context.Context) (int64, *map[int64]*interface{}, error)
	ListBlades(ctx context.Context) (int64, *map[int64]*interface{}, error)

	FetchPdus(ctx context.Context) (int64, *map[int64]*interface{}, error)
	FetchTors(ctx context.Context) (int64, *map[int64]*interface{}, error)
	FetchBlades(ctx context.Context) (int64, *map[int64]*interface{}, error)
}

type inventoryItemPdu interface {
	inventoryItem

	SetPorts(ports *map[int64]*pb.PowerPort)
	GetPorts() *map[int64]*pb.PowerPort
}

type inventoryTor interface {
	inventoryItem

	SetPorts(cports *map[int64]*pb.NetworkPort)
	GetPorts() *map[int64]*pb.NetworkPort
}

type inventoryBlade interface {
	inventoryItem

	SetCapacity(capacity *pb.BladeCapacity)
	GetCapacity() *pb.BladeCapacity

	SetBootInfo(bootOnPowerOn bool, bootInfo *pb.BladeBootInfo)
	GetBootInfo() (bool, *pb.BladeBootInfo)
}

// Provide a set of definitions to cope with calls to a "null" object.
//
type nullItem struct{}

func (n *nullItem) SetDetails(details *nullItem) {
}

func (n *nullItem) GetDetails() *nullItem {
	return nil
}

func (n *nullItem) Create(ctx context.Context) (int64, error) {
	return store.RevisionInvalid, errors.ErrNullItem
}

func (n *nullItem) Read(ctx context.Context) (int64, error) {
	return store.RevisionInvalid, errors.ErrNullItem
}

func (n *nullItem) Update(ctx context.Context) (int64, error) {
	return store.RevisionInvalid, errors.ErrNullItem
}

func (n *nullItem) Delete(ctx context.Context) (int64, error) {
	return store.RevisionInvalid, errors.ErrNullItem
}

// Additional functions for the node specialization of the basic inventory item
//
func (n *nullItem) NewChild(name string) (*interface{}, error) {
	return nil, errors.ErrNullItem
}

func (n *nullItem) ListChildren(ctx context.Context) (int64, *[]string, error) {
	return store.RevisionInvalid, nil, errors.ErrNullItem
}
func (n *nullItem) FetchChildren(ctx context.Context) (int64, *map[string]interface{}, error) {
	return store.RevisionInvalid, nil, errors.ErrNullItem
}

// Additional functions for the rack specialization of the basic inventory item
//
func (n *nullItem) NewPdu(name string) (*interface{}, error) {
	return nil, errors.ErrNullItem
}

func (n *nullItem) NewTor( name string) (*interface{}, error) {
	return nil, errors.ErrNullItem
}

func (n *nullItem) NewBlade(name string) (*interface{}, error) {
	return nil, errors.ErrNullItem
}

func (n *nullItem) ListPdus(ctx context.Context) (int64, *map[int64]*interface{}, error) {
	return store.RevisionInvalid, nil, errors.ErrNullItem
}

func (n *nullItem) ListTors(ctx context.Context) (int64, *map[int64]*interface{}, error) {
	return store.RevisionInvalid, nil, errors.ErrNullItem
}

func (n *nullItem) ListBlades(ctx context.Context) (int64, *map[int64]*interface{}, error) {
	return store.RevisionInvalid, nil, errors.ErrNullItem
}

func (n *nullItem) FetchPdus(ctx context.Context) (int64, *[]string, error) {
	return store.RevisionInvalid, nil, errors.ErrNullItem
}

func (n *nullItem) FetchTors(ctx context.Context) (int64, *[]string, error) {
	return store.RevisionInvalid, nil, errors.ErrNullItem
}

func (n *nullItem) FetchBlades(ctx context.Context) (int64, *[]string, error) {
	return store.RevisionInvalid, nil, errors.ErrNullItem
}

// Additional functions for the pdu and tor specializations of the basic inventory item
//
func (n *nullItem) SetPorts(ports *map[int64]*interface{}) {
	return
}

func (n *nullItem) GetPorts() *map[int64]*interface{} {
	return nil
}

// Additional functions for the blade specialization of the basic inventory item
//
func (n *nullItem) SetCapacity(capacity *interface{}) {
	return
}

func (n *nullItem) GetCapacity() *interface{} {
	return nil
}

func (n *nullItem) SetBootInfo(bootOnPowerOn bool, bootInfo *interface{}) {
	return
}

func (n *nullItem) GetBootInfo() (bool, *interface{}) {
	return false, nil
}


// ZoneSummary contains the summary data for a single zone. The contents
// are either zero or are (re-)computed whenever the inventory definitions
// are loaded from file. They are a cache of data in the store to avoid
// having to scan the entire zone whenever a simple query for the basic
// data is being handled.
//
type ZoneSummary struct {
	RackCount     int
	MaxBladeCount int
	MaxCapacity   *pb.BladeCapacity
}

// RegionSummary contains the summary data for a single region. The contents
// are either zero or are (re-)computed whenever the inventory definitions
// are loaded from file. They are a cache of data in the store to avoid
// having to scan the entire region whenever a simple query for the basic
// data is being handled.
//
type RegionSummary struct {
	ZoneCount     int
	MaxRackCount  int
	MaxBladeCount int
	MaxCapacity   *pb.BladeCapacity
}

// RootSummary contains the summary data for the entire invnetory. The contents
// are either zero or are (re-)computed whenever the inventory definitions
// are loaded from file. They are a cache of data in the store to avoid
// having to scan the entire zone whenever a simple query for the basic
// data is being handled.
//
type RootSummary struct {
	RegionCount   int
	MaxZoneCount  int
	MaxRackCount  int
	MaxBladeCount int
	MaxCapacity   *pb.BladeCapacity
}

// Inventory is a structure used to estblished synchronized access to values 
// required to make use of the inventory layer.
//
type Inventory struct {
	mutex              sync.RWMutex
	cfg                *config.GlobalConfig	
	Store              *store.Store
	RootSummary        *RootSummary
	DefaultZoneSummary *ZoneSummary
}

// NewInventory is a helper routine to construct an empty Inventory structure
// as a convenience to avoid clients having to do all the details themselves.
//
func NewInventory(cfg *config.GlobalConfig, store *store.Store) *Inventory {
	return &Inventory{
		mutex:              sync.RWMutex{},
		cfg:                cfg,
		Store:              store,
		RootSummary:        &RootSummary{},
		DefaultZoneSummary: &ZoneSummary{
			MaxCapacity:   &pb.BladeCapacity{
				Accelerators: []*pb.Accelerator{},
			},
		},
	}
}

// Start is a function to get the inventory ready for use.
//
func (m *Inventory) Start(ctx context.Context) error {

	if err := m.Store.Connect(); err != nil {
		return err
	}

	if err := m.UpdateInventoryDefinition(ctx, m.cfg.Inventory.InventoryDefinition); err != nil {
		return err
	}

	return nil
}

// GetDefaultZoneSummary returns the maximum number of blades held in any rack
// in the inventory.
func (m *Inventory) GetDefaultZoneSummary() *ZoneSummary {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.DefaultZoneSummary
}

// UpdateInventoryDefinition is a method to load a new inventory definition from
// the configured file. Once read, the store will be updated with the differences
// which will in turn trigger a set of previously established watch routines to
// issue a number of arrival and/or departure notifications.
//
func (m *Inventory) UpdateInventoryDefinition(ctx context.Context, path string) error {
	ctx, span := tracing.StartSpan(ctx,
		tracing.WithName("Update inventory definition from file"),
		tracing.WithContextValue(timestamp.EnsureTickInContext),
		tracing.AsInternal())
	defer span.End()

	// We have the basic initialization done. Now go read the current inventory
	// from the file indicated by the configuration. Once we have that, load it
	// into the store looking to see if there are any material changes between
	// what is already in the store and what is now found in the file.
	//
	rootFile, err := ReadInventoryDefinitionFromFileEx(ctx, path)
	if err != nil {
		return err
	}

	rootStore, err := m.readInventoryDefinitionFromStore(ctx)

	if err != nil {
		return err
	}

	if err = m.reconcileNewInventory(ctx, rootFile, rootStore); err != nil {
		return err
	}

	return nil
}

// DeleteInventoryDefinition is a
//
func (m *Inventory) DeleteInventoryDefinition(ctx context.Context) error {
	ctx, span := tracing.StartSpan(ctx,
		tracing.WithName("Delete inventory definition from store"),
		tracing.WithContextValue(timestamp.EnsureTickInContext),
		tracing.AsInternal())
	defer span.End()

	root, err := m.readInventoryDefinitionFromStore(ctx)
	if err != nil {
		return err
	}

	return m.deleteInventoryDefinitionFromStore(ctx, root)
}

// reconcileNewInventory compares the newly loaded inventory definition,
// presumably from a configuration file, with the currently loaded inventory
// and updates the store accordingly. This will trigger the various watches
// which any currently running services have previously established and deliver
// a set of arrival and/or departure notifications as appropriate.
//
// NOTE: As a temporary measure, reconciliation just deletes the old inventory
//       from the store and completely replaces it with the newly read inventory
//       from the configured file
//
func (m *Inventory) reconcileNewInventory(
	ctx context.Context,
	rootFile *pb.Definition_Root,
	rootStore *pb.Definition_Root) error {
	ctx, span := tracing.StartSpan(ctx,
		tracing.WithName("Reconcile current inventory with update"),
		tracing.WithContextValue(timestamp.EnsureTickInContext),
		tracing.AsInternal())
	defer span.End()

	m.mutex.Lock()
	defer m.mutex.Unlock()

	err := m.deleteInventoryDefinitionFromStore(ctx, rootStore)
	if err != nil {
		return err
	}

	err = m.writeInventoryDefinitionToStore(ctx, rootFile)
	if err != nil {
		return err
	}

	m.RootSummary = m.buildSummaryForRoot(rootFile)

	tracing.Info(
		ctx,
		"   Updated inventory summary - RegionCount: %d MaxZoneCount: %d MaxRackCount: %d MaxBladeCount: %d MaxCapacity: %v",
		m.RootSummary.RegionCount,
		m.RootSummary.MaxZoneCount,
		m.RootSummary.MaxRackCount,
		m.RootSummary.MaxBladeCount,
		&m.RootSummary.MaxCapacity)


	zone, err := m.getDefaultZone(rootFile)

	if err != nil {
		m.DefaultZoneSummary = &ZoneSummary{}
	
		tracing.Error(
			ctx,
			"   Reset DEFAULT inventory summary - MaxRackCount: %d MaxBladeCount: %d MaxCapacity: %v - %v",
			m.DefaultZoneSummary.RackCount,
			m.DefaultZoneSummary.MaxBladeCount,
			&m.DefaultZoneSummary.MaxCapacity,
			err,
		)
	} else {
		m.DefaultZoneSummary = m.buildSummaryForZone(zone)

		tracing.Info(
			ctx,
			"   Updated DEFAULT inventory summary - MaxRackCount: %d MaxBladeCount: %d MaxCapacity: %v",
			m.DefaultZoneSummary.RackCount,
			m.DefaultZoneSummary.MaxBladeCount,
			&m.DefaultZoneSummary.MaxCapacity,
		)
	}

	return nil
}

func (m *Inventory) getDefaultZone(root *pb.Definition_Root) (*pb.Definition_Zone, error) {

	region, ok := root.Regions[DefaultRegion]

	if !ok {
		return nil, errors.ErrRegionNotFound{Region: DefaultRegion}
	}

	zone, ok := region.Zones[DefaultZone]

	if !ok {
		return nil, errors.ErrZoneNotFound{Region: DefaultRegion, Zone: DefaultZone}
	}

	return zone, nil
}

// readInventoryDefinitionFromStore is used to read all the inventory
// definitions from the store, regardless of how they got there and 
// return them in the hierarchical in-memory form.
//
func (m *Inventory) readInventoryDefinitionFromStore(ctx context.Context) (*pb.Definition_Root, error) {
	ctx, span := tracing.StartSpan(ctx,
		tracing.WithName("Read inventory definition from store"),
		tracing.WithContextValue(timestamp.EnsureTickInContext),
		tracing.AsInternal())
	defer span.End()

	root, err := m.NewRoot(DefinitionTable)
	if err != nil {
		return nil, err
	}

	_, regions, err := root.FetchChildren(ctx)
	if err != nil {
		return nil, err
	}

	defRoot := &pb.Definition_Root{
		Details: root.GetDetails(),
		Regions: make(map[string]*pb.Definition_Region, len(*regions)),
	}

	for regionName, region := range *regions {
		_, zones, err := region.FetchChildren(ctx)
		if err != nil {
			return nil, err
		}

		defRegion := &pb.Definition_Region{
			Details: region.GetDetails(),
			Zones:   make(map[string]*pb.Definition_Zone, len(*zones)),
		}

		for zoneName, zone := range *zones {
			_, racks, err := zone.FetchChildren(ctx)
			if err != nil {
				return nil, err
			}

			defZone := &pb.Definition_Zone{
				Details: zone.GetDetails(),
				Racks:   make(map[string]*pb.Definition_Rack, len(*racks)),
			}

			for rackName, rack := range *racks {
				defRack, err := rack.Copy(ctx)
				if err != nil {
					return nil, err
				}
	
				defZone.Racks[rackName] = defRack
			}

			defRegion.Zones[zoneName] = defZone
		}

		defRoot.Regions[regionName] = defRegion
	}

	return defRoot, nil
}

// writeInventoryDefinitionToStore will use the supplied root parameter and 
// persist a record for each item in the supplied Definition_Root contents.
//
// In the case of collisions, the routine will fail with an "already exists"
// error.
//
func (m *Inventory) writeInventoryDefinitionToStore(ctx context.Context, root *pb.Definition_Root) error {
	ctx, span := tracing.StartSpan(ctx,
		tracing.WithName("Write inventory definition to store"),
		tracing.WithContextValue(timestamp.EnsureTickInContext),
		tracing.AsInternal())
	defer span.End()

	storeRoot, err := m.NewRoot(DefinitionTable)
	if err != nil {
		return err
	}

	storeRoot.SetDetails(root.Details)

	for regionName, region := range root.Regions {
		ctx, span := tracing.StartSpan(ctx,
			tracing.WithName(fmt.Sprintf("Write inventory definition for region %q", regionName)),
			tracing.WithContextValue(timestamp.EnsureTickInContext),
			tracing.AsInternal())
		defer span.End()

		storeRegion, err := storeRoot.NewChild(regionName)
		if err != nil {
			return err
		}

		storeRegion.SetDetails(region.GetDetails())

		_, err = storeRegion.Create(ctx)
		if err != nil {
			return err
		}

		for zoneName, zone := range region.Zones {
			ctx, span := tracing.StartSpan(ctx,
				tracing.WithName(fmt.Sprintf("Write inventory definition for region %q, zone %q", regionName, zoneName)),
				tracing.WithContextValue(timestamp.EnsureTickInContext),
				tracing.AsInternal())
			defer span.End()

			storeZone, err := storeRegion.NewChild(zoneName)
			if err != nil {
				return err
			}
	
			storeZone.SetDetails(zone.GetDetails())
	
			_, err = storeZone.Create(ctx)
			if err != nil {
				return err
			}

			for rackName, rack := range zone.Racks {
				ctx, span := tracing.StartSpan(ctx,
					tracing.WithName(fmt.Sprintf("Write inventory definition for region %q, zone %q, rack %q", regionName, zoneName, rackName)),
					tracing.WithContextValue(timestamp.EnsureTickInContext),
					tracing.AsInternal())
				defer span.End()

				storeRack, err := storeZone.NewChild(rackName)
				if err != nil {
					return err
				}
		
				storeRack.SetDetails(rack.GetDetails())
		
				_, err = storeRack.Create(ctx)
				if err != nil {
					return err
				}

				for index, pdu := range rack.Pdus {
					ctx, span := tracing.StartSpan(ctx,
						tracing.WithName(fmt.Sprintf("Write inventory definition for region %q, zone %q, rack %q, pdu %d", regionName, zoneName, rackName, index)),
						tracing.WithContextValue(timestamp.EnsureTickInContext),
						tracing.AsInternal())
					defer span.End()

					storePdu, err := storeRack.NewPdu(index)
					if err != nil {
						return err
					}
					ports := pdu.GetPorts()

					storePdu.SetDetails(pdu.GetDetails())
					storePdu.SetPorts(&ports)

					_, err = storePdu.Create(ctx)
					if err != nil {
						return err
					}
				}

				for index, tor := range rack.Tors {
					ctx, span := tracing.StartSpan(ctx,
						tracing.WithName(fmt.Sprintf("Write inventory definition for region %q, zone %q, rack %q, tor %d", regionName, zoneName, rackName, index)),
						tracing.WithContextValue(timestamp.EnsureTickInContext),
						tracing.AsInternal())
					defer span.End()

					storeTor, err := storeRack.NewTor(index)
					if err != nil {
						return err
					}

					ports := tor.GetPorts()
					storeTor.SetDetails(tor.GetDetails())
					storeTor.SetPorts(&ports)

					_, err = storeTor.Create(ctx)
					if err != nil {
						return err
					}
				}

				for index, blade := range rack.Blades {
					ctx, span := tracing.StartSpan(ctx,
						tracing.WithName(fmt.Sprintf("Write inventory definition for region %q, zone %q, rack %q, blade %d", regionName, zoneName, rackName, index)),
						tracing.WithContextValue(timestamp.EnsureTickInContext),
						tracing.AsInternal())
					defer span.End()

					storeBlade, err := storeRack.NewBlade(index)
					if err != nil {
						return err
					}

					storeBlade.SetDetails(blade.GetDetails())
					storeBlade.SetCapacity(blade.GetCapacity())
					storeBlade.SetBootInfo(blade.BootOnPowerOn, blade.GetBootInfo())

					_, err = storeBlade.Create(ctx)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

// deleteInventoryDefinitionFromStore is used to completely remove all
// inventory definitions from the store as identified by the storeRoot
// parameter, regardless of how they got there.
//
func (m *Inventory) deleteInventoryDefinitionFromStore(ctx context.Context, storeRoot *pb.Definition_Root) error {

	root, err := m.NewRoot(DefinitionTable)
	if err != nil {
		return err
	}

	for regionName, storeRegion := range storeRoot.Regions {

		region, err := root.NewChild(regionName)
		if err != nil {
			return err
		}

		for zoneName, storeZone := range storeRegion.Zones {

			zone, err := region.NewChild(zoneName)
			if err != nil {
				return err
			}

			for rackName, storeRack := range storeZone.Racks {

				rack, err := zone.NewChild(rackName)
				if err != nil {
					return err
				}

				for i := range storeRack.Pdus {

					pdu, err := rack.NewPdu(i)
					if err != nil {
						return err
					}

					pdu.Delete(ctx, true)
				}

				for i := range storeRack.Tors {

					tor, err := rack.NewTor(i)
					if err != nil {
						return err
					}

					tor.Delete(ctx, true)
				}

				for i := range storeRack.Blades {

					blade, err := rack.NewBlade(i)
					if err != nil {
						return err
					}

					blade.Delete(ctx, true)
				}

			rack.Delete(ctx, true)
			}

		zone.Delete(ctx, true)
		}

	region.Delete(ctx, true)
	}

	return nil
}

// buildSummaryForRoot constructs the memo-ed summary data for the root. This should
// be called whenever the configured inventory changes. This includes
//
// - the zone count
// - the maximum number of blades in a rack
// - the memo data itself
//
func (m *Inventory) buildSummaryForRoot(
	root *pb.Definition_Root) *RootSummary {

	maxCapacity := &pb.BladeCapacity{}
	maxZoneCount := int(0)
	maxRackCount := int(0)
	maxBladeCount := int(0)

	for _, region := range root.Regions {
		regionSummary := m.buildSummaryForRegion(region)

		maxBladeCount = common.MaxInt(maxBladeCount, regionSummary.MaxBladeCount)
		maxRackCount = common.MaxInt(maxRackCount, regionSummary.MaxRackCount)
		maxZoneCount = common.MaxInt(maxZoneCount, regionSummary.ZoneCount)

		maxCapacity.Cores = common.MaxInt64(maxCapacity.Cores, regionSummary.MaxCapacity.Cores)
		maxCapacity.DiskInGb = common.MaxInt64(maxCapacity.DiskInGb, regionSummary.MaxCapacity.DiskInGb)
		maxCapacity.MemoryInMb = common.MaxInt64(maxCapacity.MemoryInMb, regionSummary.MaxCapacity.MemoryInMb)

		maxCapacity.NetworkBandwidthInMbps = common.MaxInt64(
				maxCapacity.NetworkBandwidthInMbps,
				regionSummary.MaxCapacity.NetworkBandwidthInMbps)
	}

	return &RootSummary{
		RegionCount:   len(root.Regions),
		MaxZoneCount:  maxZoneCount,
		MaxRackCount:  maxRackCount,
		MaxBladeCount: maxBladeCount,
		MaxCapacity:   maxCapacity,
	}
}

// buildSummary constructs the memo-ed summary data for the zone.  This should
// be called whenever the configured inventory changes. This includes
//
// - the zone count
// - the maximum number of blades in a rack
// - the memo data itself
//
func (m *Inventory) buildSummaryForRegion(
	region *pb.Definition_Region) *RegionSummary {

	maxCapacity := &pb.BladeCapacity{}
	maxRackCount := int(0)
	maxBladeCount := int(0)

	for _, zone := range region.Zones {
		zoneSummary := m.buildSummaryForZone(zone)

		maxBladeCount = common.MaxInt(maxBladeCount, zoneSummary.MaxBladeCount)
		maxRackCount = common.MaxInt(maxRackCount, zoneSummary.RackCount)

		maxCapacity.Cores = common.MaxInt64(maxCapacity.Cores, zoneSummary.MaxCapacity.Cores)
		maxCapacity.DiskInGb = common.MaxInt64(maxCapacity.DiskInGb, zoneSummary.MaxCapacity.DiskInGb)
		maxCapacity.MemoryInMb = common.MaxInt64(maxCapacity.MemoryInMb, zoneSummary.MaxCapacity.MemoryInMb)

		maxCapacity.NetworkBandwidthInMbps = common.MaxInt64(
				maxCapacity.NetworkBandwidthInMbps,
				zoneSummary.MaxCapacity.NetworkBandwidthInMbps)

		maxBladeCount = common.MaxInt(maxBladeCount, zoneSummary.MaxBladeCount)
	}

	return &RegionSummary{
		ZoneCount:     len(region.Zones),
		MaxRackCount:  maxRackCount,
		MaxBladeCount: maxBladeCount,
		MaxCapacity:   maxCapacity,
	}
}

func (m *Inventory) buildSummaryForZone(zone *pb.Definition_Zone) *ZoneSummary {

	maxCapacity := &pb.BladeCapacity{}
	maxBladeCount := int(0)

	for _, rack := range zone.Racks {
		bladeCount, capacity := m.buildSummaryForRack(rack)

		maxBladeCount = common.MaxInt(maxBladeCount, bladeCount)
		maxCapacity.Cores = common.MaxInt64(maxCapacity.Cores, capacity.Cores)
		maxCapacity.DiskInGb = common.MaxInt64(maxCapacity.DiskInGb, capacity.DiskInGb)
		maxCapacity.MemoryInMb = common.MaxInt64(maxCapacity.MemoryInMb, capacity.MemoryInMb)
		maxCapacity.NetworkBandwidthInMbps = common.MaxInt64(
			maxCapacity.NetworkBandwidthInMbps,
			capacity.NetworkBandwidthInMbps)
	}

	return &ZoneSummary{
		RackCount:     len(zone.Racks),
		MaxBladeCount: maxBladeCount,
		MaxCapacity:   maxCapacity,
	}
}

func (m *Inventory) buildSummaryForRack(rack *pb.Definition_Rack) (int, *pb.BladeCapacity) {

	memo := &pb.BladeCapacity{}

	for _, blade := range rack.Blades {
		memo.Cores = common.MaxInt64(memo.Cores, blade.Capacity.Cores)
		memo.DiskInGb = common.MaxInt64(memo.DiskInGb, blade.Capacity.DiskInGb)
		memo.MemoryInMb = common.MaxInt64(memo.MemoryInMb, blade.Capacity.MemoryInMb)
		memo.NetworkBandwidthInMbps = common.MaxInt64(
			memo.NetworkBandwidthInMbps,
			blade.Capacity.NetworkBandwidthInMbps)
	}

	return len(rack.Blades), memo
}

