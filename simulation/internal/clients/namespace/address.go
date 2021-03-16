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


type AddressRegion struct {
	table TableName
	region string
}

type AddressZone struct {
	AddressRegion
	zone string
}

type AddressRack struct {
	AddressZone
	rack string
}

type AddressPdu struct {
	AddressRack
	pdu int64
}

type AddressTor struct {
	AddressRack
	tor int64
}

type AddressBlade struct {
	AddressRack
	blade int64
}

type Address struct {
	AddressRegion
	AddressZone
	AddressRack
	AddressPdu
	AddressTor
	AddressBlade

	nodeType AddressType
}


func (a *AddressRegion) Validate() bool {
	return a.region != ""
}

func (a *AddressRegion) String() string {
	return fmt.Sprintf("Region %s", a.region)
}

func (a *AddressRegion) Region() string {
	return a.region
}

func (a *AddressRegion) Table() TableName {
	return a.table
}


func (a *AddressZone) Validate() bool {
	return a.zone != ""  && a.AddressRegion.Validate()
}

func (a *AddressZone) String() string {
	return fmt.Sprintf("Region: %s Zone: %s", a.region, a.zone)
}

func (a *AddressZone) Zone() string {
	return a.zone
}


func (a *AddressRack) Validate() bool {
	return a.rack != "" && a.AddressZone.Validate()
}

func (a *AddressRack) String() string {
	return fmt.Sprintf("Region: %s Zone: %s Rack: %s", a.region, a.zone, a.rack)
}

func (a *AddressRack) Rack() string {
	return a.rack
}


func (a *AddressPdu) Validate() bool {
	return a.pdu < limits.MaxPduID && a.AddressRack.Validate()
}

func (a *AddressPdu) String() string {
	return fmt.Sprintf("Region: %s Zone: %s Rack: %s Pdu: %d", a.region, a.zone, a.rack, a.pdu)
}

func (a *AddressPdu) Pdu() int64 {
	return a.pdu
}


func (a *AddressTor) Validate() bool {
	return a.tor < limits.MaxTorID && a.AddressRack.Validate()
}

func (a *AddressTor) String() string {
	return fmt.Sprintf("Region: %s Zone: %s Rack: %s Pdu: %d", a.region, a.zone, a.rack, a.tor)
}

func (a *AddressTor) Tor() int64 {
	return a.tor
}


func (a *AddressBlade) Validate() bool {
	return a.blade < limits.MaxBladeID && a.AddressRack.Validate()
}

func (a *AddressBlade) String() string {
	return fmt.Sprintf("Region: %s Zone: %s Rack: %s Pdu: %d", a.region, a.zone, a.rack, a.blade)
}

func (a *AddressBlade) Blade() int64 {
	return a.blade
}



func (a *Address) GetType() AddressType {
	return a.nodeType
}

func (a *Address) Validate() bool {
	switch a.nodeType {
	case AddressTypeRegion: return a.AddressRegion.Validate()
	case AddressTypeZone: return a.AddressZone.Validate()
	case AddressTypeRack: return a.AddressRack.Validate()
	case AddressTypePdu: return a.AddressPdu.Validate()
	case AddressTypeTor: return a.AddressTor.Validate()
	case AddressTypeBlade: return a.AddressBlade.Validate()
	default: return false
	}
}

func (a *Address) String() string {
	switch a.nodeType {
	case AddressTypeRegion: return a.AddressRegion.String()
	case AddressTypeZone: return a.AddressZone.String()
	case AddressTypeRack: return a.AddressRack.String()
	case AddressTypePdu: return a.AddressPdu.String()
	case AddressTypeTor: return a.AddressTor.String()
	case AddressTypeBlade: return a.AddressBlade.String()
	default: return "<UNRECOGNIZED ADDRESS>"
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