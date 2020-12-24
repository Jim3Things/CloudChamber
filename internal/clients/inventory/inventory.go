package inventory

import (
	"context"
	"errors"
	"fmt"

	pb "github.com/Jim3Things/CloudChamber/pkg/protos/inventory"
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
	DefaultZone   = "DefZone"

	// DefinitionTable is a constant to indicate the inventory operation should be
	// performed against the inventory definition table for the item of interest.
	//
	DefinitionTable = "definition"

	// ActualTable is a constant to indicate the inventory operation should be
	// performed against the inventory actual state table for the item of interest.
	//
	ActualTable     = "actual"

	// ObservedTable is a constant to indicate the inventory operation should be
	// performed against the inventory observed state table for the item of interest.
	//
	ObservedTable   = "observed"

	// TargetTable is a constant to indicate the inventory operation should be
	// performed against the inventory target state table for the item of interest.
	//
	TargetTable     = "target"

	prefixRegion = "region"
	prefixZone   = "zone"
	prefixRack   = "rack"
	prefixBlade  = "blade"
	prefixPdu    = "pdu"
	prefixTor    = "tor"

	keyFormatIndexRegions  = "%s/" + prefixRegion + "s"
	keyFormatIndexZones    = "%s/" + prefixRegion + "/%s/" + prefixZone + "s"
	keyFormatIndexRacks    = "%s/" + prefixRegion + "/%s/" + prefixZone + "/%s/" + prefixRack + "s"
	keyFormatIndexPdus     = "%s/" + prefixRegion + "/%s/" + prefixZone + "/%s/" + prefixRack + "/%s/" + prefixPdu   + "s"
	keyFormatIndexTors     = "%s/" + prefixRegion + "/%s/" + prefixZone + "/%s/" + prefixRack + "/%s/" + prefixTor   + "s"
	keyFormatIndexBlades   = "%s/" + prefixRegion + "/%s/" + prefixZone + "/%s/" + prefixRack + "/%s/" + prefixBlade + "s"

	keyFormatRegion = "%s/" + prefixRegion + "/%s"
	keyFormatZone   = "%s/" + prefixRegion + "/%s/" + prefixZone + "/%s"
	keyFormatRack   = "%s/" + prefixRegion + "/%s/" + prefixZone + "/%s/" + prefixRack + "/%s"
	keyFormatPdu    = "%s/" + prefixRegion + "/%s/" + prefixZone + "/%s/" + prefixRack + "/%s/" + prefixPdu   + "/%v"
	keyFormatTor    = "%s/" + prefixRegion + "/%s/" + prefixZone + "/%s/" + prefixRack + "/%s/" + prefixTor   + "/%v"
	keyFormatBlade  = "%s/" + prefixRegion + "/%s/" + prefixZone + "/%s/" + prefixRack + "/%s/" + prefixBlade + "/%v"

	maxBladeID = int64(10 * 1000 * 1000)
	maxPduID   = int64(2)
	maxTorID   = int64(2)
)


func verifyTable(table string) error {
	switch (table) {
	case DefinitionTable: return nil
	case ActualTable:     return nil
	case ObservedTable:   return nil
	case TargetTable:     return nil

	case "":
		return ErrTableNameMissing(table)

	default:
		return ErrTableNameInvalid(table)
	}
}

func verifyRegion(val string) error {

	if "" == val {
		return ErrRegionNameMissing(val)
	}

	return nil
}

func verifyZone(val string) error {

	if "" == val {
		return ErrZoneNameMissing(val)
	}

	return nil
}

func verifyRack(val string) error {

	if "" == val {
		return ErrRackNameMissing(val)
	}

	return nil
}

func verifyBlade(val int64) error {

	if val < 0 || val > maxBladeID {
		return ErrBladeIDInvalid(val)
	}

	return nil
}

func verifyPdu(val int64) error {

	if val < 0 || val > maxPduID {
		return ErrPduIDInvalid(val)
	}

	return nil
}

func verifyTor(val int64) error {

	if val < 0 || val > maxTorID {
		return ErrTorIDInvalid(val)
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

	return fmt.Sprintf(keyFormatIndexRegions, table), nil
}


// GetKeyForIndexZone generates the key to discover the list of zones within a
// specific table (definition, actual, observed, target)
// 
func GetKeyForIndexZone(table string, region string) (key string, err error) {

	if err = verifyTable(table); err != nil {
		return key, err
	}

	if err = verifyRegion(region); err != nil {
		return key, err
	}

	return fmt.Sprintf(keyFormatIndexZones, table, region), nil
}

// GetKeyForIndexRack generates the key to discover the list of reracksgions within a
// specific table (definition, actual, observed, target)
// 
func GetKeyForIndexRack(table string, region string, zone string) (key string, err error) {

	if err = verifyTable(table); err != nil {
		return key, err
	}

	if err := verifyRegion(region); err != nil {
		return key, err
	}

	if err := verifyZone(zone); err != nil {
		return key, err
	}

	return fmt.Sprintf(keyFormatIndexRacks, table, region, zone), nil
}

// GetKeyForIndexPdu generates the key to discover the list of pdus within a
// specific table (definition, actual, observed, target)
// 
func GetKeyForIndexPdu(table string, region string, zone string, rack string) (key string, err error) {
	
	if err = verifyTable(table); err != nil {
		return key, err
	}

	if err = verifyRegion(region); err != nil {
		return key, err
	}

	if err = verifyZone(zone); err != nil {
		return key, err
	}

	if err = verifyRack(rack); err != nil {
		return key, err
	}

	return fmt.Sprintf(keyFormatIndexPdus, table, region, zone, rack), nil
}

// GetKeyForIndexTor generates the key to discover the list of tors within a
// specific table (definition, actual, observed, target)
// 
func GetKeyForIndexTor(table string, region string, zone string, rack string) (key string, err error) {

	if err = verifyTable(table); err != nil {
		return key, err
	}

	if err:= verifyRegion(region); err != nil {
		return key, err
	}

	if err = verifyZone(zone); err != nil {
		return key, err
	}

	if err = verifyRack(rack); err != nil {
		return key, err
	}

	return fmt.Sprintf(keyFormatIndexTors, table, region, zone, rack), nil
}

// GetKeyForIndexBlade generates the key to discover the list of blades within a
// specific table (definition, actual, observed, target)
// 
func GetKeyForIndexBlade(table string, region string, zone string, rack string) (key string, err error) {

	if err = verifyTable(table); err != nil {
		return key, err
	}

	if err = verifyRegion(region); err != nil {
		return key, err
	}

	if err = verifyZone(zone); err != nil {
		return key, err
	}

	if err = verifyRack(rack); err != nil {
		return key, err
	}

	return fmt.Sprintf(keyFormatIndexBlades, table, region, zone, rack), nil
}


// GetKeyForRegion generates the key to operate on the record for a region within a
// specific table (definition, actual, observed, target)
// 
func GetKeyForRegion(table string, region string) (key string, err error) {

	if err = verifyTable(table); err != nil {
		return key, err
	}

	if err = verifyRegion(region); err != nil {
		return key, err
	}

	return fmt.Sprintf(keyFormatRegion, table, region), nil
}

// GetKeyForZone generates the key to operate on the record for a zone within a
// specific table (definition, actual, observed, target)
// 
func GetKeyForZone(table string, region string, zone string) (key string, err error) {

	if err = verifyTable(table); err != nil {
		return key, err
	}

	if err = verifyRegion(region); err != nil {
		return key, err
	}

	if err = verifyZone(zone); err != nil {
		return key, err
	}

	return fmt.Sprintf(keyFormatZone, table, region, zone), nil
}

// GetKeyForRack generates the key to operate on the record for a rack within a
// specific table (definition, actual, observed, target)
// 
func GetKeyForRack(table string, region string, zone string, rack string) (key string, err error) {

	if err = verifyTable(table); err != nil {
		return key, err
	}

	if err = verifyRegion(region); err != nil {
		return key, err
	}

	if err = verifyZone(zone); err != nil {
		return key, err
	}

	if err = verifyRack(rack); err != nil {
		return key, err
	}

	return fmt.Sprintf(keyFormatRack, table, region, zone, rack), nil
}

// GetKeyForPdu generates the key to operate on the record for a pdu within a
// specific table (definition, actual, observed, target)
// 
func GetKeyForPdu(table string, region string, zone string, rack string, pdu int64) (key string, err error) {
	
	if err = verifyTable(table); err != nil {
		return key, err
	}

	if err = verifyRegion(region); err != nil {
		return key, err
	}

	if err = verifyZone(zone); err != nil {
		return key, err
	}

	if err = verifyRack(rack); err != nil {
		return key, err
	}

	if err = verifyBlade(pdu); err != nil {
		return key, err
	}

	return fmt.Sprintf(keyFormatPdu, table, region, zone, rack, pdu), nil
}

// GetKeyForTor generates the key to operate on the record for a tor within a
// specific table (definition, actual, observed, target)
// 
func GetKeyForTor(table string, region string, zone string, rack string, tor int64) (key string, err error) {

	if err = verifyTable(table); err != nil {
		return key, err
	}

	if err = verifyRegion(region); err != nil {
		return key, err
	}

	if err = verifyZone(zone); err != nil {
		return key, err
	}

	if err = verifyRack(rack); err != nil {
		return key, err
	}

	if err:= verifyTor(tor); err != nil {
		return key, err
	}

	return fmt.Sprintf(keyFormatTor, table, region, zone, rack, tor), nil
}

// GetKeyForBlade generates the key to operate on the record for a blade within a
// specific table (definition, actual, observed, target)
// 
func GetKeyForBlade(table string, region string, zone string, rack string, blade int64) (key string, err error) {

	if err = verifyTable(table); err != nil {
		return key, err
	}

	if err = verifyRegion(region); err != nil {
		return key, err
	}

	if err = verifyZone(zone); err != nil {
		return key, err
	}

	if err = verifyRack(rack); err != nil {
		return key, err
	}

	if err = verifyBlade(blade); err != nil {
		return key, err
	}

	return fmt.Sprintf(keyFormatBlade, table, region, zone, rack, blade), nil
}


// Address is a
//
type Address struct {
	region string
	zone string
	rack string
	item *pb.Hardware
}

type item struct {
	address Address
	key string
}

// GetKeyForAddress generates the key to operate on the record for an inventory item within a
// specific table (definition, actual, observed, target)
//
func GetKeyForAddress(table string, addr *Address) (string, error) {

	switch  {
	case addr.region != "" && addr.zone != "" && addr.rack != "" && addr.item != nil:
		switch addr.item.Type {
		case pb.Hardware_pdu: 
			return GetKeyForPdu(table, addr.region, addr.zone, addr.rack, addr.item.Id)

		case pb.Hardware_tor:
			return GetKeyForPdu(table, addr.region, addr.zone, addr.rack, addr.item.Id)

		case pb.Hardware_blade:
			return GetKeyForPdu(table, addr.region, addr.zone, addr.rack, addr.item.Id)
		}

	case addr.region != "" && addr.zone != "" && addr.rack != "":
		return GetKeyForRack(table, addr.region, addr.zone, addr.rack)

	case addr.region != "" && addr.zone != "":
		return GetKeyForZone(table, addr.region, addr.zone)

	case addr.region != "":
		return GetKeyForRegion(table, addr.region)
	}

	return "", ErrInvalidAddress{table, addr}
}

// Region, zone and rack are "containers" whereas tor, pdu and blade are "things". You can send operations and commands to things, but not containers.
type inventoryItem interface {
	SetAddress(ctx context.Context, addr Address) error
	SetName(ctx context.Context, name string) error
	GetKey(ctx context.Context) (*string, error)

	SetDetails(ctx context.Context, details *interface{}) error
	GetDetails(ctx context.Context) (*interface{}, error)

	Create(ctx context.Context) (int64, error)
	Read(ctx context.Context) (*interface{}, error)
	Update(ctx context.Context) (int64, error)
	Delete(ctx context.Context) error

	NewChild(ctx context.Context, name string) (*interface{}, error)
	ListChildren(ctx context.Context) (*map[string]interface{}, error)
}

// type inventoryRoot interface {
// 	inventoryItem
// }
// type inventoryRegion interface {
// 	inventoryItem
// }

// type inventoryZone interface {
// 	inventoryItem
// }

type inventoryItemRack interface {
	inventoryItem

	NewPdu(ctx context.Context,   name string) (*interface{}, error)
	NewTor(ctx context.Context,   name string) (*interface{}, error)
	NewBlade(ctx context.Context, name string) (*interface{}, error)

	ListPdus(ctx context.Context)   (*map[int64]*interface{}, error)
	ListTors(ctx context.Context)   (*map[int64]*interface{}, error)
	ListBlades(ctx context.Context) (*map[int64]*interface{}, error)
}

// type inventoryPdu interface {
// 	inventoryItem
// }

// type inventoryTor interface {
// 	inventoryItem
// }

// type inventoryBlade interface {
// 	inventoryItem
// }


var (
	// ErrNullItem indicates the supplied item does not exist
	//
	ErrNullItem             = errors.New("item not initialized")

	// ErrFunctionNotAvailable indicates the specified object does
	// not have the requested method.
	//
	ErrFunctionNotAvailable = errors.New("function not available")
)

type nullItem struct {}

func (n *nullItem) SetAddress(addr Address) error {
	return ErrNullItem
}

// SetName is a
//
func (n *nullItem) SetName(addr Address) error {
	return 	ErrNullItem
}

func (n *nullItem) GetKey() (*string, error) {
	return nil, ErrNullItem
}


func (n *nullItem) SetDetails(ctx context.Context, detaila *nullItem) error {
	return ErrNullItem
}

func (n *nullItem) GetDetails(ctx context.Context) (*nullItem, error) {
	return nil, ErrNullItem
}

func (n *nullItem) Create(ctx context.Context) error {
	return ErrNullItem
}

func (n *nullItem) Read(ctx context.Context) (*nullItem, error){
	return nil, ErrNullItem
}

func (n *nullItem) Update(ctx context.Context) error {
	return ErrNullItem
}

func (n *nullItem) Delete(ctx context.Context) error {
	return ErrNullItem
}


// Additional functions for the rack specialization of the basic inventory item
//

func (n *nullItem) NewPdu(ctx context.Context,   name string) (*interface{}, error) {
	return nil, ErrNullItem
}

func (n *nullItem) NewTor(ctx context.Context,   name string) (*interface{}, error) {
	return nil, ErrNullItem
}

func (n *nullItem) NewBlade(ctx context.Context, name string) (*interface{}, error) {
	return nil, ErrNullItem
}

func (n *nullItem) ListPdus(ctx context.Context)   (*map[int64]*interface{}, error) {
	return nil, ErrNullItem
}

func (n *nullItem) ListTors(ctx context.Context)   (*map[int64]*interface{}, error) {
	return nil, ErrNullItem
}

func (n *nullItem) ListBlades(ctx context.Context) (*map[int64]*interface{}, error) {
	return nil, ErrNullItem
}
