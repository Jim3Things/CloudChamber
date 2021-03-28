package namespace

import (
	"fmt"
	"strings"

	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/limits"

	"github.com/Jim3Things/CloudChamber/simulation/pkg/errors"
)

const (
	// DefinitionTable is a constant to indicate the inventory operation should be
	// performed against the inventory definition table for the item of interest.
	//
	DefinitionTable TableName = "definition"

	// DefinitionTableStdTest is a constant to indicate the inventory operation should be
	// performed against the test inventory definition table for the standard test
	// inventory for the item of interest.
	//
	DefinitionTableStdTest TableName = "definitionstdtest"

	// ActualTable is a constant to indicate the inventory operation should be
	// performed against the inventory actual state table for the item of interest.
	//
	ActualTable TableName = "actual"

	// ObservedTable is a constant to indicate the inventory operation should be
	// performed against the inventory observed state table for the item of interest.
	//
	ObservedTable TableName = "observed"

	// TargetTable is a constant to indicate the inventory operation should be
	// performed against the inventory target state table for the item of interest.
	//
	TargetTable TableName = "target"

	// InvalidTable is a default value for the case where a tablename in not one
	// of the valid set.
	//
	InvalidTable TableName = ""

	// InvalidRegion is used when a return valid is needed for a region but where
	// the value cannot be used in any way other than for comparison.
	//
	InvalidRegion string = ""

	// InvalidZone is used when a return valid is needed for a zone but where
	// the value cannot be used in any way other than for comparison.
	//
	InvalidZone   string = ""

	// InvalidRack is used when a return valid is needed for a rack but where
	// the value cannot be used in any way other than for comparison.
	//
	InvalidRack   string = ""

	// InvalidPdu is used when a return valid is needed for a pdu but where
	// the value cannot be used in any way other than for comparison.
	//
	InvalidPdu    int64  = -1

	// InvalidTor is used when a return valid is needed for a tor but where
	// the value cannot be used in any way other than for comparison.
	//
	invalidTor    int64  = -1

	// InvalidBlade is used when a return valid is needed for a blade but where
	// the value cannot be used in any way other than for comparison.
	//
	InvalidBlade  int64  = -1

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
)

// KeyRoot is used to describe which part of the store namespace
// should be used for the corresponding record access.
//
type KeyRoot int

// The set of available namespace roots used by various record types
//
const (
	KeyRootStoreTest KeyRoot = iota
	KeyRootUsers
	KeyRootInventory
	KeyRootWorkloads
)

const (
	namespaceRootStoreTest = "storetest"
	namespaceRootUsers     = "users"
	namespaceRootInventory = "inventory"
	namespaceRootWorkloads = "workload"
)

var namespaceRoots = map[KeyRoot]string{
	KeyRootStoreTest: namespaceRootStoreTest,
	KeyRootUsers:     namespaceRootUsers,
	KeyRootInventory: namespaceRootInventory,
	KeyRootWorkloads: namespaceRootWorkloads,
}



type TableName string

func (t TableName) String() string {
	return string(t)
}

func (t TableName) NormalizeName() string {
	return GetNormalizedName(t.String())
}

func (t TableName)  Validate() error {
	switch t {
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
		return errors.ErrTableNameMissing(t.String())

	default:
		return errors.ErrTableNameInvalid{
			Name:            t.String(),
			DefinitionTable: DefinitionTable.String(),
			ActualTable:     ActualTable.String(),
			ObservedTable:   ObservedTable.String(),
			TargetTable:     TargetTable.String(),
		}
	}
}

func GetTableNameFromString(name string) (TableName, error) {
	switch name {
	case DefinitionTable.String():
		return DefinitionTable, nil
	case ActualTable.String():
		return ActualTable, nil
	case ObservedTable.String():
		return ObservedTable, nil
	case TargetTable.String():
		return TargetTable, nil

	default:
		return InvalidTable, errors.ErrTableNameInvalid{
			Name:            name,
			DefinitionTable: DefinitionTable.String(),
			ActualTable:     ActualTable.String(),
			ObservedTable:   ObservedTable.String(),
			TargetTable:     TargetTable.String(),
		}
	}
}

func verifyTableName(table TableName) error {
	switch table {
	case
	DefinitionTable,
	ActualTable,
	ObservedTable,
	TargetTable:
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
			Name:            table.String(),
			DefinitionTable: DefinitionTable.String(),
			ActualTable:     ActualTable.String(),
			ObservedTable:   ObservedTable.String(),
			TargetTable:     TargetTable.String(),
		}
	}
}

func requireNotEmpty(s string, e error) error {
	if s == "" {
		return e
	}

	return nil
}

func verifyRegion(val string) error {
	return requireNotEmpty(val, errors.ErrRegionNameMissing)
}


func verifyZone(val string) error {
	return requireNotEmpty(val, errors.ErrZoneNameMissing)
}

func verifyRack(val string) error {
	return requireNotEmpty(val, errors.ErrRackNameMissing)
}

func verifyPdu(val int64) error {

	if val < 0 || val > limits.MaxPduID {
		return errors.ErrPduIDInvalid{Value: val, Limit: limits.MaxPduID}
	}

	return nil
}

func verifyTor(val int64) error {

	if val < 0 || val > limits.MaxTorID {
		return errors.ErrTorIDInvalid{Value: val, Limit: limits.MaxTorID}
	}

	return nil
}

func verifyBlade(val int64) error {

	if val < 0 || val > limits.MaxBladeID {
		return errors.ErrBladeIDInvalid{Value: val, Limit: limits.MaxBladeID}
	}

	return nil
}

func verifyNamesRegion(table TableName, region string) error {

	if err := verifyTableName(table); err != nil {
		return err
	}

	return verifyRegion(region)
}

func verifyNamesZone(table TableName, region string, zone string) error {

	if err := verifyNamesRegion(table, region); err != nil {
		return err
	}

	return verifyZone(zone)
}

func verifyNamesRack(table TableName, region string, zone string, rack string) error {

	if err := verifyNamesZone(table, region, zone); err != nil {
		return err
	}

	return verifyRack(rack)
}

func verifyNamesPdu(table TableName, region string, zone string, rack string, index int64) error {

	if err := verifyNamesRack(table, region, zone, rack); err != nil {
		return err
	}

	return verifyPdu(index)
}

func verifyNamesTor(table TableName, region string, zone string, rack string, index int64) error {

	if err := verifyNamesRack(table, region, zone, rack); err != nil {
		return err
	}

	return verifyTor(index)
}

func verifyNamesBlade(table TableName, region string, zone string, rack string, index int64) error {

	if err := verifyNamesRack(table, region, zone, rack); err != nil {
		return err
	}

	return verifyBlade(index)
}

// GetKeyForIndexRegions generates the key to discover the list of regions within a
// specific table (definition, actual, observed, target)
//
func GetKeyForIndexRegions(table TableName) (key string, err error) {

	if err = verifyTableName(table); err != nil {
		return key, err
	}

	key = fmt.Sprintf(
		keyFormatIndexRegions,
		table.NormalizeName())

	return key, nil
}

// GetKeyForIndexZones generates the key to discover the list of zones within a
// specific table (definition, actual, observed, target)
//
func GetKeyForIndexZones(table TableName, region string) (key string, err error) {

	if err = verifyNamesRegion(table, region); err != nil {
		return key, err
	}

	key = fmt.Sprintf(
		keyFormatIndexZones,
		table.NormalizeName(),
		GetNormalizedName(region))

	return key, nil
}

// GetKeyForIndexRacks generates the key to discover the list of racks within a
// specific table (definition, actual, observed, target)
//
func GetKeyForIndexRacks(table TableName, region string, zone string) (key string, err error) {

	if err = verifyNamesZone(table, region, zone); err != nil {
		return key, err
	}

	key = fmt.Sprintf(
		keyFormatIndexRacks,
		table.NormalizeName(),
		GetNormalizedName(region),
		GetNormalizedName(zone))

	return key, nil
}

// GetKeyForIndexPdu generates the key to discover the list of pdus within a
// specific table (definition, actual, observed, target)
//
func GetKeyForIndexPdus(table TableName, region string, zone string, rack string) (key string, err error) {

	if err = verifyNamesRack(table, region, zone, rack); err != nil {
		return key, err
	}

	key = fmt.Sprintf(
		keyFormatIndexPdus,
		table.NormalizeName(),
		GetNormalizedName(region),
		GetNormalizedName(zone),
		GetNormalizedName(rack))

	return key, nil
}

// GetKeyForIndexTors generates the key to discover the list of tors within a
// specific table (definition, actual, observed, target)
//
func GetKeyForIndexTors(table TableName, region string, zone string, rack string) (key string, err error) {

	if err = verifyNamesRack(table, region, zone, rack); err != nil {
		return key, err
	}

	key = fmt.Sprintf(
		keyFormatIndexTors,
		table.NormalizeName(),
		GetNormalizedName(region),
		GetNormalizedName(zone),
		GetNormalizedName(rack))

	return key, nil
}

// GetKeyForIndexBlades generates the key to discover the list of blades within a
// specific table (definition, actual, observed, target)
//
func GetKeyForIndexBlades(table TableName, region string, zone string, rack string) (key string, err error) {

	if err = verifyNamesRack(table, region, zone, rack); err != nil {
		return key, err
	}

	key = fmt.Sprintf(
		keyFormatIndexBlades,
		table.NormalizeName(),
		GetNormalizedName(region),
		GetNormalizedName(zone),
		GetNormalizedName(rack))

	return key, nil
}

// GetKeyForIndexEntryRegion generates the key to create an index entry for a region within a
// specific table (definition, actual, observed, target)
//
func GetKeyForIndexEntryRegion(table TableName, region string) (key string, err error) {

	if err = verifyNamesRegion(table, region); err != nil {
		return key, err
	}

	key = fmt.Sprintf(
		keyFormatIndexEntryRegion,
		table.NormalizeName(),
		GetNormalizedName(region))

	return key, nil
}

// GetKeyForIndexEntryZone generates the key to create an index entry for a zone within a
// specific table (definition, actual, observed, target)
//
func GetKeyForIndexEntryZone(table TableName, region string, zone string) (key string, err error) {

	if err = verifyNamesZone(table, region, zone); err != nil {
		return key, err
	}

	key = fmt.Sprintf(
		keyFormatIndexEntryZone,
		table.NormalizeName(),
		GetNormalizedName(region),
		GetNormalizedName(zone))

	return key, nil
}

// GetKeyForIndexEntryRack generates the key to create an index entry for a rack within a
// specific table (definition, actual, observed, target)
//
func GetKeyForIndexEntryRack(table TableName, region string, zone string, rack string) (key string, err error) {

	if err = verifyNamesRack(table, region, zone, rack); err != nil {
		return key, err
	}

	key = fmt.Sprintf(
		keyFormatIndexEntryRack,
		table.NormalizeName(),
		GetNormalizedName(region),
		GetNormalizedName(zone),
		GetNormalizedName(rack))

	return key, nil
}

// GetKeyForIndexEntryPdu generates the key to create an index entry for a pdu within a
// specific table (definition, actual, observed, target)
//
func GetKeyForIndexEntryPdu(table TableName, region string, zone string, rack string, pdu int64) (key string, err error) {

	if err = verifyNamesPdu(table, region, zone, rack, pdu); err != nil {
		return key, err
	}

	key = fmt.Sprintf(
		keyFormatIndexEntryPdu,
		table.NormalizeName(),
		GetNormalizedName(region),
		GetNormalizedName(zone),
		GetNormalizedName(rack),
		pdu)

	return key, nil
}

// GetKeyForIndexEntryTor generates the key to create an index entry for a tor within a
// specific table (definition, actual, observed, target)
//
func GetKeyForIndexEntryTor(table TableName, region string, zone string, rack string, tor int64) (key string, err error) {

	if err = verifyNamesTor(table, region, zone, rack, tor); err != nil {
		return key, err
	}

	key = fmt.Sprintf(
		keyFormatIndexEntryTor,
		table.NormalizeName(),
		GetNormalizedName(region),
		GetNormalizedName(zone),
		GetNormalizedName(rack),
		tor)

	return key, nil
}

// GetKeyForIndexEntryBlade generates the key to create an index entry for a blade within a
// specific table (definition, actual, observed, target)
//
func GetKeyForIndexEntryBlade(table TableName, region string, zone string, rack string, blade int64) (key string, err error) {

	if err = verifyNamesBlade(table, region, zone, rack, blade); err != nil {
		return key, err
	}

	key = fmt.Sprintf(
		keyFormatIndexEntryBlade,
		table.NormalizeName(),
		GetNormalizedName(region),
		GetNormalizedName(zone),
		GetNormalizedName(rack),
		blade)

	return key, nil
}

// GetKeyForRegion generates the key to operate on the record for a region within a
// specific table (definition, actual, observed, target)
//
func GetKeyForRegion(table TableName, region string) (key string, err error) {

	if err = verifyNamesRegion(table, region); err != nil {
		return key, err
	}

	key = fmt.Sprintf(
		keyFormatRegion,
		table.NormalizeName(),
		GetNormalizedName(region))

	return key, nil
}

// GetKeyForZone generates the key to operate on the record for a zone within a
// specific table (definition, actual, observed, target)
//
func GetKeyForZone(table TableName, region string, zone string) (key string, err error) {

	if err = verifyNamesZone(table, region, zone); err != nil {
		return key, err
	}

	key = fmt.Sprintf(
		keyFormatZone,
		table.NormalizeName(),
		GetNormalizedName(region),
		GetNormalizedName(zone))

	return key, nil
}

// GetKeyForRack generates the key to operate on the record for a rack within a
// specific table (definition, actual, observed, target)
//
func GetKeyForRack(table TableName, region string, zone string, rack string) (key string, err error) {

	if err = verifyNamesRack(table, region, zone, rack); err != nil {
		return key, err
	}

	key = fmt.Sprintf(
		keyFormatRack,
		table.NormalizeName(),
		GetNormalizedName(region),
		GetNormalizedName(zone),
		GetNormalizedName(rack))

	return key, nil
}

// GetKeyForPdu generates the key to operate on the record for a pdu within a
// specific table (definition, actual, observed, target)
//
func GetKeyForPdu(table TableName, region string, zone string, rack string, pdu int64) (key string, err error) {

	if err = verifyNamesPdu(table, region, zone, rack, pdu); err != nil {
		return key, err
	}

	key = fmt.Sprintf(
		keyFormatPdu,
		table.NormalizeName(),
		GetNormalizedName(region),
		GetNormalizedName(zone),
		GetNormalizedName(rack),
		pdu)

	return key, nil
}

// GetKeyForTor generates the key to operate on the record for a tor within a
// specific table (definition, actual, observed, target)
//
func GetKeyForTor(table TableName, region string, zone string, rack string, tor int64) (key string, err error) {

	if err = verifyNamesTor(table, region, zone, rack, tor); err != nil {
		return key, err
	}

	key = fmt.Sprintf(
		keyFormatTor,
		table.NormalizeName(),
		GetNormalizedName(region),
		GetNormalizedName(zone),
		GetNormalizedName(rack),
		tor)

	return key, nil
}

// GetKeyForBlade generates the key to operate on the record for a blade within a
// specific table (definition, actual, observed, target)
//
func GetKeyForBlade(table TableName, region string, zone string, rack string, blade int64) (key string, err error) {

	if err = verifyNamesBlade(table, region, zone, rack, blade); err != nil {
		return key, err
	}

	key = fmt.Sprintf(
		keyFormatBlade,
		table.NormalizeName(),
		GetNormalizedName(region),
		GetNormalizedName(zone),
		GetNormalizedName(rack),
		blade)

	return key, nil
}

func getNamespaceRootFromKeyRoot(r KeyRoot) string {
	return namespaceRoots[r]
}

func GetNamespacePrefixFromKeyRoot(r KeyRoot) string {
	return getNamespaceRootFromKeyRoot(r) + "/"
}

func GetKeyFromKeyRootAndName(r KeyRoot, n string) string {
	return getNamespaceRootFromKeyRoot(r) + "/" + GetNormalizedName(n)
}

func GetNameFromKeyRootAndKey(r KeyRoot, k string) string {
	n := strings.TrimPrefix(k, getNamespaceRootFromKeyRoot(r)+"/")
	return n
}

// GetKeyFromUsername1 is a utility function to convert a supplied username to
// a store usable key for use when operating with user records.
//
func GetKeyFromUsername(name string) string {
	return GetKeyFromKeyRootAndName(KeyRootUsers, name)
}

// GetNormalizedName is a utility function to prepare a name for use when
// building a key suitable for operating with records in the store
//
func GetNormalizedName(name string) string {
	return strings.ToLower(name)
}
