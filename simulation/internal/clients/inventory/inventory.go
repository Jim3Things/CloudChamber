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

	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/store"
	"github.com/Jim3Things/CloudChamber/simulation/pkg/errors"
	pb "github.com/Jim3Things/CloudChamber/simulation/pkg/protos/inventory"
)

const (

	// DefaultRegion is a default value until the correct region can be
	// retrieved from the external configuration file. It is not intended
	// for permananet usage.
	//
	DefaultRegion = "DefRegion"

	// DefaultZone is a default value until the correct zone can be
	// retrieved from the external configuration file. It is not intended
	// for permananet usage.
	//
	DefaultZone = "DefZone"

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

	maxBladeID = int64(10 * 1000 * 1000)
	maxPduID   = int64(2)
	maxTorID   = int64(2)
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

	if val < 0 || val > maxPduID {
		return errors.ErrPduIDInvalid{Value: val, Limit: maxPduID}
	}

	return nil
}

func verifyTor(val int64) error {

	if val < 0 || val > maxTorID {
		return errors.ErrTorIDInvalid{Value: val, Limit: maxTorID}
	}

	return nil
}

func verifyBlade(val int64) error {

	if val < 0 || val > maxBladeID {
		return errors.ErrBladeIDInvalid{Value: val, Limit: maxBladeID}
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

// GetKeyForIndexRegion generates the key to discover the list of regions within a
// specific table (definition, actual, observed, target)
//
func GetKeyForIndexRegion(table string) (key string, err error) {

	if err = verifyTable(table); err != nil {
		return key, err
	}

	key = fmt.Sprintf(
		keyFormatIndexRegions,
		store.GetNormalizedName(table))

	return key, nil
}

// GetKeyForIndexZone generates the key to discover the list of zones within a
// specific table (definition, actual, observed, target)
//
func GetKeyForIndexZone(table string, region string) (key string, err error) {

	if err = verifyNamesRegion(table, region); err != nil {
		return key, err
	}

	key = fmt.Sprintf(
		keyFormatIndexZones,
		store.GetNormalizedName(table),
		store.GetNormalizedName(region))

	return key, nil
}

// GetKeyForIndexRack generates the key to discover the list of racks within a
// specific table (definition, actual, observed, target)
//
func GetKeyForIndexRack(table string, region string, zone string) (key string, err error) {

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
func GetKeyForIndexPdu(table string, region string, zone string, rack string) (key string, err error) {

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

// GetKeyForIndexTor generates the key to discover the list of tors within a
// specific table (definition, actual, observed, target)
//
func GetKeyForIndexTor(table string, region string, zone string, rack string) (key string, err error) {

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

// GetKeyForIndexBlade generates the key to discover the list of blades within a
// specific table (definition, actual, observed, target)
//
func GetKeyForIndexBlade(table string, region string, zone string, rack string) (key string, err error) {

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

// GetKeyForIndexEntryRegion generates the key to create an index entry for a region within a
// specific table (definition, actual, observed, target)
//
func GetKeyForIndexEntryRegion(table string, region string) (key string, err error) {

	if err = verifyNamesRegion(table, region); err != nil {
		return key, err
	}

	key = fmt.Sprintf(
		keyFormatIndexEntryRegion,
		store.GetNormalizedName(table),
		store.GetNormalizedName(region))

	return key, nil
}

// GetKeyForIndexEntryZone generates the key to create an index entry for a zone within a
// specific table (definition, actual, observed, target)
//
func GetKeyForIndexEntryZone(table string, region string, zone string) (key string, err error) {

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

// GetKeyForIndexEntryRack generates the key to create an index entry for a rack within a
// specific table (definition, actual, observed, target)
//
func GetKeyForIndexEntryRack(table string, region string, zone string, rack string) (key string, err error) {

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

// GetKeyForIndexEntryPdu generates the key to create an index entry for a pdu within a
// specific table (definition, actual, observed, target)
//
func GetKeyForIndexEntryPdu(table string, region string, zone string, rack string, pdu int64) (key string, err error) {

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

// GetKeyForIndexEntryTor generates the key to create an index entry for a tor within a
// specific table (definition, actual, observed, target)
//
func GetKeyForIndexEntryTor(table string, region string, zone string, rack string, tor int64) (key string, err error) {

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

// GetKeyForIndexEntryBlade generates the key to create an index entry for a blade within a
// specific table (definition, actual, observed, target)
//
func GetKeyForIndexEntryBlade(table string, region string, zone string, rack string, blade int64) (key string, err error) {

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

// GetKeyForRegion generates the key to operate on the record for a region within a
// specific table (definition, actual, observed, target)
//
func GetKeyForRegion(table string, region string) (key string, err error) {

	if err = verifyNamesRegion(table, region); err != nil {
		return key, err
	}

	key = fmt.Sprintf(
		keyFormatRegion,
		store.GetNormalizedName(table),
		store.GetNormalizedName(region))

	return key, nil
}

// GetKeyForZone generates the key to operate on the record for a zone within a
// specific table (definition, actual, observed, target)
//
func GetKeyForZone(table string, region string, zone string) (key string, err error) {

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

// GetKeyForRack generates the key to operate on the record for a rack within a
// specific table (definition, actual, observed, target)
//
func GetKeyForRack(table string, region string, zone string, rack string) (key string, err error) {

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

// GetKeyForPdu generates the key to operate on the record for a pdu within a
// specific table (definition, actual, observed, target)
//
func GetKeyForPdu(table string, region string, zone string, rack string, pdu int64) (key string, err error) {

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

// GetKeyForTor generates the key to operate on the record for a tor within a
// specific table (definition, actual, observed, target)
//
func GetKeyForTor(table string, region string, zone string, rack string, tor int64) (key string, err error) {

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

// GetKeyForBlade generates the key to operate on the record for a blade within a
// specific table (definition, actual, observed, target)
//
func GetKeyForBlade(table string, region string, zone string, rack string, blade int64) (key string, err error) {

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
	NewChild(ctx context.Context, name string) (*interface{}, error)

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
func (n *nullItem) GetRevision(ctx context.Context) int64 {
	return store.RevisionInvalid
}

func (n *nullItem) GetRevisionRecord(ctx context.Context) int64 {
	return store.RevisionInvalid
}

func (n *nullItem) GetRevisionStore(ctx context.Context) int64 {
	return store.RevisionInvalid
}

func (n *nullItem) GetRevisionForRequest(ctx context.Context, unconditional bool) int64 {
	return store.RevisionInvalid
}

func (n *nullItem) resetRevision(ctx context.Context) int64 {
	return store.RevisionInvalid
}

func (n *nullItem) updateRevisionInfo(ctx context.Context, rev int64) int64 {
	return store.RevisionInvalid
}

type inventoryItemRack interface {
	inventoryItemNode

	NewPdu(ctx context.Context, name string) (*interface{}, error)
	NewTor(ctx context.Context, name string) (*interface{}, error)
	NewBlade(ctx context.Context, name string) (*interface{}, error)

	ListPdus(ctx context.Context) (int64, *map[int64]*interface{}, error)
	ListTors(ctx context.Context) (int64, *map[int64]*interface{}, error)
	ListBlades(ctx context.Context) (int64, *map[int64]*interface{}, error)

	FetchPdus(ctx context.Context) (int64, *map[int64]*interface{}, error)
	FetchTors(ctx context.Context) (int64, *map[int64]*interface{}, error)
	FetchBlades(ctx context.Context) (int64, *map[int64]*interface{}, error)
}

type inventoryItemPdu interface {
	inventoryItem

	SetPorts(ctx context.Context, ports *map[int64]*pb.PowerPort)
	GetPorts(ctx context.Context) *map[int64]*pb.PowerPort
}

type inventoryTor interface {
	inventoryItem

	SetPorts(ctx context.Context, ports *map[int64]*pb.NetworkPort)
	GetPorts(ctx context.Context) *map[int64]*pb.NetworkPort
}

type inventoryBlade interface {
	inventoryItem

	SetCapacity(ctx context.Context, capacity *pb.BladeCapacity)
	GetCapacity(ctx context.Context) *pb.BladeCapacity

	SetBootInfo(ctx context.Context, bootOnPowerOn bool, bootInfo *pb.BladeBootInfo)
	GetBootInfo(ctx context.Context) (bool, *pb.BladeBootInfo)
}

// Provide a set of definitions to cope with calls to a "null" object.
//
type nullItem struct{}

func (n *nullItem) SetDetails(ctx context.Context, details *nullItem) {
}

func (n *nullItem) GetDetails(ctx context.Context) *nullItem {
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
func (n *nullItem) NewChild(ctx context.Context, name string) (*interface{}, error) {
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
func (n *nullItem) NewPdu(ctx context.Context, name string) (*interface{}, error) {
	return nil, errors.ErrNullItem
}

func (n *nullItem) NewTor(ctx context.Context, name string) (*interface{}, error) {
	return nil, errors.ErrNullItem
}

func (n *nullItem) NewBlade(ctx context.Context, name string) (*interface{}, error) {
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
func (n *nullItem) SetPorts(ctx context.Context, ports *map[int64]*interface{}) {
	return
}

func (n *nullItem) GetPorts(ctx context.Context) *map[int64]*interface{} {
	return nil
}

// Additional functions for the blade specialization of the basic inventory item
//
func (n *nullItem) SetCapacity(ctx context.Context, capacity *interface{}) {
	return
}

func (n *nullItem) GetCapacity(ctx context.Context) *interface{} {
	return nil
}

func (n *nullItem) SetBootInfo(ctx context.Context, bootOnPowerOn bool, bootInfo *interface{}) {
	return
}

func (n *nullItem) GetBootInfo(ctx context.Context) (bool, *interface{}) {
	return false, nil
}
