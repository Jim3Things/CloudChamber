package namespace

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Jim3Things/CloudChamber/simulation/pkg/errors"
)

// This module contains structures and routines to provide the notion
// of a generailized address notation for items within the inventory.
//

type AddressType int64

const (
	AddressTypeInvalid AddressType = iota
	AddressTypeRegion
	AddressTypeZone
	AddressTypeRack
	AddressTypePdu
	AddressTypeTor
	AddressTypeBlade
)

type Address struct {
	nodeType AddressType
	table    TableName
	region   string
	zone     string
	rack     string
	index    int64
}

func normalize(a *Address) *Address {
	a.region = strings.ToLower(a.region)
	a.zone   = strings.ToLower(a.zone)
	a.rack   = strings.ToLower(a.rack)

	return a
}

func NewRegion(table TableName, region string) *Address {
	return normalize(&Address{
		nodeType: AddressTypeRegion,
		table:    table,
		region:   region,
	})
}

func NewZone(table TableName, region string, zone string) *Address {
	return normalize(&Address{
		nodeType: AddressTypeZone,
		table:    table,
		region:   region,
		zone:     zone,
	})
}

func NewRack(table TableName, region string, zone string, rack string) *Address {
	return normalize(&Address{
		nodeType: AddressTypeRack,
		table:    table,
		region:   region,
		zone:     zone,
		rack:     rack,
	})
}

func NewPdu(table TableName, region string, zone string, rack string, pdu int64) *Address {
	return normalize(&Address{
		nodeType: AddressTypeBlade,
		table:    table,
		region:   region,
		zone:     zone,
		rack:     rack,
		index:    pdu,
	})
}

func NewTor(table TableName, region string, zone string, rack string, tor int64) *Address {
	return normalize(&Address{
		nodeType: AddressTypeBlade,
		table:    table,
		region:   region,
		zone:     zone,
		rack:     rack,
		index:    tor,
	})
}

func NewBlade(table TableName, region string, zone string, rack string, blade int64) *Address {
	return normalize(&Address{
		nodeType: AddressTypeBlade,
		table:    table,
		region:   region,
		zone:     zone,
		rack:     rack,
		index:    blade,
	})
}

func (a *Address) regionValidate() error {
	return verifyRegion(a.region)
}

func (a *Address) zoneValidate() error {
	if err := a.regionValidate(); err != nil {
		return err
	}

	return verifyZone(a.zone)
}

func (a *Address) rackValidate() error {
	if err := a.zoneValidate(); err != nil {
		return err
	}

	return verifyRack(a.rack)
}

func (a *Address) pduValidate() error {
	if err := a.rackValidate(); err != nil {
		return err
	}

	return verifyPdu(a.index)
}

func (a *Address) torValidate() error {
	if err := a.rackValidate(); err != nil {
		return err
	}

	return verifyTor(a.index)
}

func (a *Address) bladeValidate() error {
	if err := a.rackValidate(); err != nil {
		return err
	}

	return verifyBlade(a.index)
}



func (a *Address) regionString() string {
	return fmt.Sprintf("Region %s", a.region)
}


func (a *Address) zoneString() string {
	return fmt.Sprintf("Region: %s Zone: %s", a.region, a.zone)
}


func (a *Address) rackString() string {
	return fmt.Sprintf("Region: %s Zone: %s Rack: %s", a.region, a.zone, a.rack)
}

func (a *Address) pduString() string {
	return fmt.Sprintf("Region: %s Zone: %s Rack: %s Pdu: %d", a.region, a.zone, a.rack, a.index)
}


func (a *Address) torString() string {
	return fmt.Sprintf("Region: %s Zone: %s Rack: %s Tor: %d", a.region, a.zone, a.rack, a.index)
}


func (a *Address) bladeString() string {
	return fmt.Sprintf("Region: %s Zone: %s Rack: %s Blade: %d", a.region, a.zone, a.rack, a.index)
}


func (a *Address) Validate() error {
	switch a.nodeType {
	case AddressTypeRegion: return a.regionValidate()
	case AddressTypeZone: return a.zoneValidate()
	case AddressTypeRack: return a.rackValidate()
	case AddressTypePdu: return a.pduValidate()
	case AddressTypeTor: return a.torValidate()
	case AddressTypeBlade: return a.bladeValidate()

	default: return errors.ErrAddrInvalidType{Type: int64(a.nodeType)}
	}
}

func (a *Address) String() string {
	switch a.nodeType {
	case AddressTypeRegion: return a.regionString()
	case AddressTypeZone: return a.zoneString()
	case AddressTypeRack: return a.rackString()
	case AddressTypePdu: return a.pduString()
	case AddressTypeTor: return a.torString()
	case AddressTypeBlade: return a.bladeString()

	default: return "<UNRECOGNIZED ADDRESS>"
	}
}

func (a *Address) Type() AddressType {
	return a.nodeType
}

func (a *Address) Table() TableName {
	switch a.nodeType {
	case
	AddressTypeRegion,
	AddressTypeZone,
	AddressTypeRack,
	AddressTypePdu,
	AddressTypeTor,
	AddressTypeBlade:
		return a.table

	default:
		return InvalidTable
	}
}

func (a *Address) Region() string {
	switch a.nodeType {
	case
	AddressTypeRegion,
	AddressTypeZone,
	AddressTypeRack,
	AddressTypePdu,
	AddressTypeTor,
	AddressTypeBlade:
		return a.region

	default:
		return InvalidRegion
	}
}

func (a *Address) Zone() string {
	switch a.nodeType {
	case
	AddressTypeZone,
	AddressTypeRack,
	AddressTypePdu,
	AddressTypeTor,
	AddressTypeBlade:
		return a.zone

	default:
		return InvalidZone
	}
}

func (a *Address) Rack() string {
	switch a.nodeType {
	case
	AddressTypeRack,
	AddressTypePdu,
	AddressTypeTor,
	AddressTypeBlade:
		return a.rack

	default: return InvalidRack
	}
}

func (a *Address) Pdu() int64 {
	switch a.nodeType {
	case AddressTypePdu:
		return a.index

	default:
		return InvalidPdu
	}
}

func (a *Address) Tor() int64 {
	switch a.nodeType {
	case AddressTypeTor:
		return a.index

	default:
		return invalidTor
	}
}

func (a *Address) Blade() int64 {
	switch a.nodeType {
	case AddressTypeBlade:
		return a.index

	default:
		return InvalidBlade
	}
}


// GetAddressFromKey is a function to parse a key string, typically
// as returned by a watch event and generate a generic address for
// that string.
//
// Keys follow a number of rules
//
// must contain no spaces
// use fields separated by a "/" character
// must begin with either "index" or "data"
// must contain a second field which is one of the legal table names
// must contain a set of 1 or more pairs of fields where the first field is the item type and the second field is the item instance.
// The item type is one of
//   - "region"
//   - "zone"
//   - "rack"
//   - "pdu"
//   - "tor"
//   - "blade"
//
// Item type region, zone, and rack have string instance fields.
// Item types pdu, tor and blade have integer instance fields
//
func GetAddressFromKey(key string) (*Address, error) {
	fields := strings.Split(key, "/")

	var err error

	// Ony certain values for the size of the array are valid. Since
	// the region is not optional, there must be at least 4 fields
	//  - table name
	//  - keyword "data"
	//  - keyword "region"
	//  - value of region field
	//
	// Then optionally we can have the pair
	//  - keyword "zone"
	//  - value of zone field
	//
	// Then optionally we can have the pair
	//  - keyword "rack"
	//  - value of rack field
	//
	// Then optionally we can have any one of the following pairs
	//  - keyword "pdu"
	//  - value of pdu field
	//
	// or
	//  - keyword "tor"
	//  - value of tor field
	//
	// or
	//  - keyword "blade"
	//  - value of blade field
	//
	addr := &Address{}

	switch len(fields)  {
	case 10:
		switch fields[8] {
		case prefixPdu:
			if addr.index, err = strconv.ParseInt(fields[9], 10, 64); err != nil {
				return nil, errors.ErrNoValidAddressFromKey{Key: key}
			}
			addr.nodeType = AddressTypePdu

		case prefixTor:
			if addr.index, err = strconv.ParseInt(fields[9], 10, 64); err != nil {
				return nil, errors.ErrNoValidAddressFromKey{Key: key}
			}
			addr.nodeType = AddressTypeTor

		case prefixBlade:
			if addr.index, err = strconv.ParseInt(fields[9], 10, 64); err != nil {
				return nil, errors.ErrNoValidAddressFromKey{Key: key}
			}
			addr.nodeType = AddressTypeBlade

		default:
			return nil, errors.ErrNoValidAddressFromKey{Key: key}
		}

		fallthrough

	case 8:
		if fields[6] != prefixRack {
			return nil, errors.ErrNoValidAddressFromKey{Key: key}
		}

		if addr.nodeType == AddressTypeInvalid {
			addr.nodeType = AddressTypeRack
		}

		addr.rack = fields[7]

		fallthrough

	case 6:
		if fields[4] != prefixZone {
			return nil, errors.ErrNoValidAddressFromKey{Key: key}
		}

		if addr.nodeType == AddressTypeInvalid {
			addr.nodeType = AddressTypeZone
		}

		addr.zone = fields[5]

		fallthrough

	case 4:
		// The next value, field[1] should be either "index" or "data".
		// Everything else is invalid in an address. In addition, we never
		// build an address for index keys. As a result we reject everything
		// except "data"
		//
		if fields[1] != "data" {
			return nil, errors.ErrNoValidAddressFromKey{Key: key}
		}

		if fields[2] != prefixRegion {
			return nil, errors.ErrNoValidAddressFromKey{Key: key}
		}

		if addr.nodeType == AddressTypeInvalid {
			addr.nodeType = AddressTypeRegion
		}

		addr.region = fields[3]

		// We expect the first field to be one of the well know table names
		//
		addr.table, err = GetTableNameFromString(fields[0])

		if err != nil {
			return nil, err
		}

	default:
		return nil, errors.ErrNoValidAddressFromKey{Key: key}
	}

	return addr, nil
}