package namespace

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/limits"

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


type addressRegion struct {
	table TableName
	region string
}

type addressZone struct {
	addressRegion
	zone string
}

type addressRack struct {
	addressZone
	rack string
}

type addressPdu struct {
	addressRack
	pdu int64
}

type addressTor struct {
	addressRack
	tor int64
}

type addressBlade struct {
	addressRack
	blade int64
}

type Address struct {
	addressRegion
	addressZone
	addressRack
	addressPdu
	addressTor
	addressBlade

	nodeType AddressType
}

func NewRegion(table TableName, region string) *Address {
	addr := &Address{nodeType: AddressTypeRegion}

	addr.table  = table
	addr.region = strings.ToLower(region)

	return addr
}

func NewZone(table TableName, region string, zone string) *Address {
	addr := &Address{nodeType: AddressTypeZone}

	addr.table  = table
	addr.region = strings.ToLower(region)
	addr.zone   = strings.ToLower(zone)

	return addr
}

func NewRack(table TableName, region string, zone string, rack string) *Address {
	addr := &Address{nodeType: AddressTypeRack}

	addr.table  = table
	addr.region = strings.ToLower(region)
	addr.zone   = strings.ToLower(zone)
	addr.rack   = strings.ToLower(rack)

	return addr
}

func NewPdu(table TableName, region string, zone string, rack string, pdu int64) *Address {
	addr := &Address{nodeType: AddressTypeBlade}

	addr.table  = table
	addr.region = strings.ToLower(region)
	addr.zone   = strings.ToLower(zone)
	addr.rack   = strings.ToLower(rack)
	addr.pdu    = pdu

	return addr
}

func NewTor(table TableName, region string, zone string, rack string, tor int64) *Address {
	addr := &Address{nodeType: AddressTypeBlade}

	addr.table  = table
	addr.region = strings.ToLower(region)
	addr.zone   = strings.ToLower(zone)
	addr.rack   = strings.ToLower(rack)
	addr.tor    = tor

	return addr
}

func NewBlade(table TableName, region string, zone string, rack string, blade int64) *Address {
	addr := &Address{nodeType: AddressTypeBlade}

	addr.table  = table
	addr.region = strings.ToLower(region)
	addr.zone   = strings.ToLower(zone)
	addr.rack   = strings.ToLower(rack)
	addr.blade  = blade

	return addr
}

func (a *addressRegion) Validate() bool {
	return a.region != ""
}

func (a *addressRegion) String() string {
	return fmt.Sprintf("Region %s", a.region)
}

func (a *addressRegion) Region() string {
	return a.region
}

func (a *addressRegion) Table() TableName {
	return a.table
}


func (a *addressZone) Validate() bool {
	return a.zone != ""  && a.addressRegion.Validate()
}

func (a *addressZone) String() string {
	return fmt.Sprintf("Region: %s Zone: %s", a.region, a.zone)
}

func (a *addressZone) Zone() string {
	return a.zone
}


func (a *addressRack) Validate() bool {
	return a.rack != "" && a.addressZone.Validate()
}

func (a *addressRack) String() string {
	return fmt.Sprintf("Region: %s Zone: %s Rack: %s", a.region, a.zone, a.rack)
}

func (a *addressRack) Rack() string {
	return a.rack
}


func (a *addressPdu) Validate() bool {
	return a.pdu < limits.MaxPduID && a.addressRack.Validate()
}

func (a *addressPdu) String() string {
	return fmt.Sprintf("Region: %s Zone: %s Rack: %s Pdu: %d", a.region, a.zone, a.rack, a.pdu)
}

func (a *addressPdu) Pdu() int64 {
	return a.pdu
}


func (a *addressTor) Validate() bool {
	return a.tor < limits.MaxTorID && a.addressRack.Validate()
}

func (a *addressTor) String() string {
	return fmt.Sprintf("Region: %s Zone: %s Rack: %s Pdu: %d", a.region, a.zone, a.rack, a.tor)
}

func (a *addressTor) Tor() int64 {
	return a.tor
}


func (a *addressBlade) Validate() bool {
	return a.blade < limits.MaxBladeID && a.addressRack.Validate()
}

func (a *addressBlade) String() string {
	return fmt.Sprintf("Region: %s Zone: %s Rack: %s Pdu: %d", a.region, a.zone, a.rack, a.blade)
}

func (a *addressBlade) Blade() int64 {
	return a.blade
}



func (a *Address) Type() AddressType {
	return a.nodeType
}

func (a *Address) Validate() bool {
	switch a.nodeType {
	case AddressTypeRegion: return a.addressRegion.Validate()
	case AddressTypeZone: return a.addressZone.Validate()
	case AddressTypeRack: return a.addressRack.Validate()
	case AddressTypePdu: return a.addressPdu.Validate()
	case AddressTypeTor: return a.addressTor.Validate()
	case AddressTypeBlade: return a.addressBlade.Validate()
	default: return false
	}
}

func (a *Address) String() string {
	switch a.nodeType {
	case AddressTypeRegion: return a.addressRegion.String()
	case AddressTypeZone: return a.addressZone.String()
	case AddressTypeRack: return a.addressRack.String()
	case AddressTypePdu: return a.addressPdu.String()
	case AddressTypeTor: return a.addressTor.String()
	case AddressTypeBlade: return a.addressBlade.String()
	default: return "<UNRECOGNIZED ADDRESS>"
	}
}

func (a *Address) Region() string {
	switch a.nodeType {
	case AddressTypeRegion: return a.addressRegion.region
	case AddressTypeZone: return a.addressZone.region
	case AddressTypeRack: return a.addressRack.region
	case AddressTypePdu: return a.addressPdu.region
	case AddressTypeTor: return a.addressTor.region
	case AddressTypeBlade: return a.addressBlade.region
	default: return ""
	}
}

func (a *Address) Zone() string {
	switch a.nodeType {
	case AddressTypeZone: return a.addressZone.zone
	case AddressTypeRack: return a.addressRack.zone
	case AddressTypePdu: return a.addressPdu.zone
	case AddressTypeTor: return a.addressTor.zone
	case AddressTypeBlade: return a.addressBlade.zone
	default: return ""
	}
}

func (a *Address) Rack() string {
	switch a.nodeType {
	case AddressTypeRack: return a.addressRack.rack
	case AddressTypePdu: return a.addressPdu.rack
	case AddressTypeTor: return a.addressTor.rack
	case AddressTypeBlade: return a.addressBlade.rack
	default: return ""
	}
}

func (a *Address) Pdu() int64 {
	switch a.nodeType {
	case AddressTypePdu: return a.addressPdu.pdu
	default: return int64(-1)
	}
}

func (a *Address) Tor() int64 {
	switch a.nodeType {
	case AddressTypeTor: return a.addressTor.tor
	default: return int64(-1)
	}
}

func (a *Address) Blade() int64 {
	switch a.nodeType {
	case AddressTypeBlade: return a.addressBlade.blade
	default: return int64(-1)
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
			if addr.pdu, err = strconv.ParseInt(fields[9], 10, 64); err != nil {
				return nil, errors.ErrNoValidAddressFromKey{Key: key}
			}
			addr.nodeType = AddressTypePdu

		case prefixTor:
			if addr.tor, err = strconv.ParseInt(fields[9], 10, 64); err != nil {
				return nil, errors.ErrNoValidAddressFromKey{Key: key}
			}
			addr.nodeType = AddressTypeTor

		case prefixBlade:
			if addr.blade, err = strconv.ParseInt(fields[9], 10, 64); err != nil {
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

		// The next value, field[1] should be either "index" or "data".
		// Everything else is invalid in an address. In addition, we never
		// build an address for index keys. As a result we reject everything
		// except "data"
		//
		if fields[1] != "data" {
			return nil, errors.ErrNoValidAddressFromKey{Key: key}
		}

	default:
		return nil, errors.ErrNoValidAddressFromKey{Key: key}
	}

	return addr, nil
}