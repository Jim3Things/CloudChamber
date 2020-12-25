// This module contains the defined errors related to inventory interactions with the store and/or frontend

package inventory

import (
	"fmt"
)

// ErrInvalidAddress indicates the supplied region name is absent or otherwise not properly specified.
//
type ErrInvalidAddress struct {
	table string
	addr  *Address
}

func (eia ErrInvalidAddress) Error() string {
	return fmt.Sprintf("CloudChamber: invalid address: table: %q address: %v", eia.table, eia.addr)
}

// ErrTableNameInvalid indicates the supplied region name is not one of the valid options.
//
type ErrTableNameInvalid string

func (etni ErrTableNameInvalid) Error() string {
	return fmt.Sprintf(
		"CloudChamber: table name %q is not one of the valid options [%q, %q, %q, %q",
		string(etni),
		ActualTable,
		DefinitionTable,
		ObservedTable,
		TargetTable,
	)
}

// ErrTableNameMissing indicates the supplied region name is absent or otherwise not properly specified.
//
type ErrTableNameMissing string

func (etnm ErrTableNameMissing) Error() string {
	return fmt.Sprintf("CloudChamber: table name %q is missing or not properly specified", string(etnm))
}

// ErrRegionNameMissing indicates the supplied region name is absent or otherwise not properly specified.
//
type ErrRegionNameMissing string

func (ernm ErrRegionNameMissing) Error() string {
	return fmt.Sprintf("CloudChamber: region name %q is missing or not properly specified", string(ernm))
}

// ErrZoneNameMissing indicates the supplied zone name is absent or otherwise not properly specified.
//
type ErrZoneNameMissing string

func (eznm ErrZoneNameMissing) Error() string {
	return fmt.Sprintf("CloudChamber: zone name %q is missing or not properly specified", string(eznm))
}

// ErrRackNameMissing indicates the supplied rack name is absent or otherwise not properly specified.
//
type ErrRackNameMissing string

func (ernm ErrRackNameMissing) Error() string {
	return fmt.Sprintf("CloudChamber: zone name %q is missing or not properly specified", string(ernm))
}

// ErrBladeIDInvalid indicates the supplied bladeID was out of range, either < less than 0 or greater than maxBladeID
//
type ErrBladeIDInvalid int64

func (ebii ErrBladeIDInvalid) Error() string {
	return fmt.Sprintf("CloudChamber: bladeID %v is out of range (0 to %v)", int64(ebii), maxBladeID)
}

// ErrPduIDInvalid indicates the supplied pduID was out of range, either < less than 0 or greater than maxPduID
//
type ErrPduIDInvalid int64

func (epii ErrPduIDInvalid) Error() string {
	return fmt.Sprintf("CloudChamber: pduID %v is out of range (0 to %v)", int64(epii), maxPduID)
}

// ErrTorIDInvalid indicates the supplied torID was out of range, either < less than 0 or greater than maxTorID
//
type ErrTorIDInvalid int64

func (etii ErrTorIDInvalid) Error() string {
	return fmt.Sprintf("CloudChamber: torID %v is out of range (0 to %v)", int64(etii), maxTorID)
}

// ErrRegionAlreadyExists indicates the attempt to create a new region record
// failed as that region already exists.
//
type ErrRegionAlreadyExists string

func (ezae ErrRegionAlreadyExists) Error() string {
	return fmt.Sprintf("CloudChamber: region %q already exists", string(ezae))
}

// ErrRegionNotFound indicates the attempt to locate a region record failed as that
// region does not exist.
//
type ErrRegionNotFound string

func (eznf ErrRegionNotFound) Error() string {
	return fmt.Sprintf("CloudChamber: region %q not found", string(eznf))
}

// ErrRegionStaleVersion indicates the attempt to locate a specific version of a
// region record failed as either that region does not exist, or the specific
// version is no longer present in the store.
//
type ErrRegionStaleVersion string

func (ezsv ErrRegionStaleVersion) Error() string {
	return fmt.Sprintf("CloudChamber: region %q has a newer version than expected", string(ezsv))
}

// ErrZoneAlreadyExists indicates the attempt to create a new zone record
// failed as that zone already exists.
//
type ErrZoneAlreadyExists string

func (ezae ErrZoneAlreadyExists) Error() string {
	return fmt.Sprintf("CloudChamber: zone %q already exists", string(ezae))
}

// ErrZoneNotFound indicates the attempt to locate a zone record failed as that
// zone does not exist.
//
type ErrZoneNotFound string

func (eznf ErrZoneNotFound) Error() string {
	return fmt.Sprintf("CloudChamber: zone %q not found", string(eznf))
}

// ErrZoneStaleVersion indicates the attempt to locate a specific version of a
// zone record failed as either that zone does not exist, or the specific
// version is no longer present in the store.
//
type ErrZoneStaleVersion string

func (ezsv ErrZoneStaleVersion) Error() string {
	return fmt.Sprintf("CloudChamber: zone %q has a newer version than expected", string(ezsv))
}

// ErrRackAlreadyExists indicates the attempt to create a new rack record
// failed as that rack already exists.
//
type ErrRackAlreadyExists struct {
	Zone string
	Rack string
}

func (erae ErrRackAlreadyExists) Error() string {
	return fmt.Sprintf("CloudChamber: rack %q in zone %q already exists", erae.Rack, erae.Zone)
}

// ErrRackNotFound indicates the attempt to operate on a rack record failed
// as that record cannot be found.
//
type ErrRackNotFound struct {
	Zone string
	Rack string
}

func (ernf ErrRackNotFound) Error() string {
	return fmt.Sprintf("CloudChamber: rack %q in zone %q was not found", ernf.Rack, ernf.Zone)
}

// ErrPduAlreadyExists indicates the attempt to create a new pdu record
// failed as that pdu already exists.
//
type ErrPduAlreadyExists struct {
	Zone string
	Rack string
	Pdu int64
}

func (epae ErrPduAlreadyExists) Error() string {
	return fmt.Sprintf("CloudChamber: pdu %v in zone %q, rack %q already exists", epae.Pdu, epae.Zone, epae.Rack)
}

// ErrPduNotFound indicates the attempt to operate on a pdu record
// failed as that record cannot be found.
//
type ErrPduNotFound struct {
	Zone string
	Rack string
	Pdu int64
}

func (epae ErrPduNotFound) Error() string {
	return fmt.Sprintf("CloudChamber: pdu %v in zone %q, rack %q was not found", epae.Pdu, epae.Zone, epae.Rack)
}

// ErrTorAlreadyExists indicates the attempt to create a new zone record
// failed as that zone already exists.
//
type ErrTorAlreadyExists  struct {
	Zone string
	Rack string
	Tor int64
}

func (etae ErrTorAlreadyExists) Error() string {
	return fmt.Sprintf("CloudChamber: tor %v in zone %q, rack %q already exists", etae.Tor, etae.Zone, etae.Rack)
}

// ErrTorNotFound indicates the attempt to operate on a tor record
// failed as that record cannot be found.
//
type ErrTorNotFound struct {
	Zone string
	Rack string
	Tor int64
}

func (etnf ErrTorNotFound) Error() string {
	return fmt.Sprintf("CloudChamber: tor %v in zone %q, rack %q was not found", etnf.Tor, etnf.Zone, etnf.Rack)
}

// ErrBladeAlreadyExists indicates the attempt to create a new blade record
// failed as that blade already exists.
//
type ErrBladeAlreadyExists struct {
	Zone string
	Rack string
	Blade int64
}

func (ebae ErrBladeAlreadyExists) Error() string {
	return fmt.Sprintf("CloudChamber: blade %v in zone %q, rack %q already exists", ebae.Blade, ebae.Zone, ebae.Rack)
}

// ErrBladeNotFound indicates the attempt to operate on a blade record
// failed as that record cannot be found.
//
type ErrBladeNotFound struct {
	Zone string
	Rack string
	Blade int64
}

func (ebnf ErrBladeNotFound) Error() string {
	return fmt.Sprintf("CloudChamber: blade %v in zone %q, rack %q was not found", ebnf.Blade, ebnf.Zone, ebnf.Rack)
}

// func ErrBladeNotFound(z string, r string , b int64) error {
// 	return ErrBladeNotFound{Zone: z, Rack: r, Blade: b}
// }


// ErrRootNotFound indicates the attempt to operate on the specified namespace table
// failed as that part of the namespace cannot be found.
//
type ErrRootNotFound struct {
	namespace string
}

func (ernf ErrRootNotFound) Error() string {
	return fmt.Sprintf("CloudChamber: unable to find the root of the %q namespace", ernf.namespace)
}

// ErrIndexNotFound indicates the requested index was not found when the store
// lookup/fetch was attempted.
//
type ErrIndexNotFound string

func (einf ErrIndexNotFound) Error() string {
	return fmt.Sprintf("CloudChamber: index %q not found", string(einf))
}

// ErrDetailsNotAvailable indicates the requested detail for the item have not
// yet been establiehed.
//
type ErrDetailsNotAvailable string

func (edna ErrDetailsNotAvailable) Error() string {
	return fmt.Sprintf("CloudChamber: %q details not available", string(edna))
}

// ErrPortsNotAvailable indicates the requested detail for the item have not
// yet been establiehed.
//
type ErrPortsNotAvailable string

func (epna ErrPortsNotAvailable) Error() string {
	return fmt.Sprintf("CloudChamber: %q ports not available", string(epna))
}

// ErrCapacityNotAvailable indicates the requested capacity information for 
// the item have not yet been establiehed.
//
type ErrCapacityNotAvailable string

func (ecna ErrCapacityNotAvailable) Error() string {
	return fmt.Sprintf("CloudChamber: %q capacity not available", string(ecna))
}

// ErrBootInfoNotAvailable indicates the requested boot information for the item
// have not yet been establiehed.
//
type ErrBootInfoNotAvailable string

func (ebina ErrBootInfoNotAvailable) Error() string {
	return fmt.Sprintf("CloudChamber: %q boot information not available", string(ebina))
}

// ErrfRegionNotFound is a wrapper around the composite literal based error of the related name
//
func ErrfRegionNotFound(zone string, rack string, blade int64) error {return ErrBladeNotFound{zone, rack, blade}}

// ErrfZoneNotFound is a wrapper around the composite literal based error of the related name
//
func ErrfZoneNotFound(zone string) error {return ErrZoneNotFound(zone)}

// ErrfRackNotFound is a wrapper around the composite literal based error of the related name
//
func ErrfRackNotFound(zone string, rack string) error {return ErrRackNotFound{zone, rack}}

// ErrfRackAlreadyExists is a wrapper around the composite literal based error of the related name
//
func ErrfRackAlreadyExists(zone string, rack string) error {return ErrRackAlreadyExists{zone, rack}}

// ErrfPduNotFound is a wrapper around the composite literal based error of the related name
//
func ErrfPduNotFound(zone string, rack string, pdu int64) error {return ErrPduNotFound{zone, rack, pdu}}

// ErrfPduAlreadyExists is a wrapper around the composite literal based error of the related name
//
func ErrfPduAlreadyExists(zone string, rack string, pdu int64) error {return ErrPduAlreadyExists{zone, rack, pdu}}

// ErrfTorNotFound is a wrapper around the composite literal based error of the related name
//
func ErrfTorNotFound(zone string, rack string, tor int64) error {return ErrTorNotFound{zone, rack, tor}}

// ErrfTorAlreadyExists is a wrapper around the composite literal based error of the related name
//
func ErrfTorAlreadyExists(zone string, rack string, tor int64) error {return ErrTorAlreadyExists{zone, rack, tor}}

// ErrfBladeNotFound is a wrapper around the composite literal based error of the related name
//
func ErrfBladeNotFound(zone string, rack string, blade int64) error {return ErrBladeNotFound{zone, rack, blade}}

// ErrfBladeAlreadyExists is a wrapper around the composite literal based error of the related name
//
func ErrfBladeAlreadyExists(zone string, rack string, blade int64) error {return ErrBladeAlreadyExists{zone, rack, blade}}

