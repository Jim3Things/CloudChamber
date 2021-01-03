// This module contains the defined errors related to inventory interactions with the store and/or frontend

package inventory

import (
	"fmt"
)

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
type ErrZoneAlreadyExists struct {
	Region string
	Zone string
}

func (ezae ErrZoneAlreadyExists) Error() string {
	return fmt.Sprintf("CloudChamber: zone %q already exists in region %q", ezae.Zone, ezae.Region)
}

// ErrZoneNotFound indicates the attempt to locate a zone record failed as that
// zone does not exist.
//
type ErrZoneNotFound struct {
	Region string
	Zone string
}

func (eznf ErrZoneNotFound) Error() string {
	return fmt.Sprintf("CloudChamber: zone %q not found in region %q", eznf.Zone, eznf.Region)
}

// ErrZoneStaleVersion indicates the attempt to locate a specific version of a
// zone record failed as either that zone does not exist, or the specific
// version is no longer present in the store.
//
type ErrZoneStaleVersion struct {
	Region string
	Zone string
}


func (ezsv ErrZoneStaleVersion) Error() string {
	return fmt.Sprintf("CloudChamber: zone %q in region %q has a newer version than expected", ezsv.Zone, ezsv.Region)
}

// ErrRackAlreadyExists indicates the attempt to create a new rack record
// failed as that rack already exists.
//
type ErrRackAlreadyExists struct {
	Region string
	Zone string
	Rack string
}

func (erae ErrRackAlreadyExists) Error() string {
	return fmt.Sprintf("CloudChamber: rack %q in zone %q already exists in region %q", erae.Rack, erae.Zone, erae.Region)
}

// ErrRackNotFound indicates the attempt to operate on a rack record failed
// as that record cannot be found.
//
type ErrRackNotFound struct {
	Region string
	Zone string
	Rack string
}

func (ernf ErrRackNotFound) Error() string {
	return fmt.Sprintf("CloudChamber: rack %q in zone %q was not found in region %q", ernf.Rack, ernf.Zone, ernf.Region)
}

// ErrPduIndexInvalid indicates the attempt to locate a record
// failed as the given index is invalid in some way.
//
type ErrPduIndexInvalid struct {
	Region string
	Zone string
	Rack string
	Pdu string
}

func (epii ErrPduIndexInvalid) Error() string {
	return fmt.Sprintf("CloudChamber: pdu %q in region %q, zone %q, rack %q was not valid", epii.Pdu, epii.Region, epii.Zone, epii.Rack)
}

// ErrPduNotFound indicates the attempt to operate on a pdu record
// failed as that record cannot be found.
//
type ErrPduNotFound struct {
	Region string
	Zone string
	Rack string
	Pdu int64
}

func (epae ErrPduNotFound) Error() string {
	return fmt.Sprintf("CloudChamber: pdu %v in region %q, zone %q, rack %q was not found", epae.Pdu, epae.Region, epae.Zone, epae.Rack)
}

// ErrPduAlreadyExists indicates the attempt to create a new pdu record
// failed as that pdu already exists.
//
type ErrPduAlreadyExists struct {
	Region string
	Zone string
	Rack string
	Pdu int64
}

func (epae ErrPduAlreadyExists) Error() string {
	return fmt.Sprintf("CloudChamber: pdu %v in region %q, zone %q, rack %q already exists", epae.Pdu, epae.Region, epae.Zone, epae.Rack)
}

// ErrTorIndexInvalid indicates the attempt to locate a record
// failed as the given index is invalid in some way.
//
type ErrTorIndexInvalid struct {
	Region string
	Zone string
	Rack string
	Tor string
}

func (etii ErrTorIndexInvalid) Error() string {
	return fmt.Sprintf("CloudChamber: tor %q in region %q, zone %q, rack %q was not valid", etii.Tor, etii.Region, etii.Zone, etii.Rack)
}

// ErrTorNotFound indicates the attempt to operate on a tor record
// failed as that record cannot be found.
//
type ErrTorNotFound struct {
	Region string
	Zone string
	Rack string
	Tor int64
}

func (etnf ErrTorNotFound) Error() string {
	return fmt.Sprintf("CloudChamber: tor %v in region %q, zone %q, rack %q was not found", etnf.Tor, etnf.Region, etnf.Zone, etnf.Rack)
}

// ErrTorAlreadyExists indicates the attempt to create a new zone record
// failed as that zone already exists.
//
type ErrTorAlreadyExists  struct {
	Region string
	Zone string
	Rack string
	Tor int64
}

func (etae ErrTorAlreadyExists) Error() string {
	return fmt.Sprintf("CloudChamber: tor %v in region %q, zone %q, rack %q already exists", etae.Tor, etae.Region, etae.Zone, etae.Rack)
}

// ErrBladeIndexInvalid indicates the attempt to locate a record
// failed as the given index is invalid in some way.
//
type ErrBladeIndexInvalid struct {
	Region string
	Zone string
	Rack string
	Blade string
}

func (ebii ErrBladeIndexInvalid) Error() string {
	return fmt.Sprintf("CloudChamber: blade %q in region %q, zone %q, rack %q was not valid", ebii.Blade, ebii.Region, ebii.Zone, ebii.Rack)
}

// ErrBladeNotFound indicates the attempt to operate on a blade record
// failed as that record cannot be found.
//
type ErrBladeNotFound struct {
	Region string
	Zone string
	Rack string
	Blade int64
}

func (ebnf ErrBladeNotFound) Error() string {
	return fmt.Sprintf("CloudChamber: blade %v in region %q, zone %q, rack %q was not found", ebnf.Blade, ebnf.Region, ebnf.Zone, ebnf.Rack)
}

// ErrBladeAlreadyExists indicates the attempt to create a new blade record
// failed as that blade already exists.
//
type ErrBladeAlreadyExists struct {
	Region string
	Zone string
	Rack string
	Blade int64
}

func (ebae ErrBladeAlreadyExists) Error() string {
	return fmt.Sprintf("CloudChamber: blade %v in region %q, zone %q, rack %q already exists", ebae.Blade, ebae.Region, ebae.Zone, ebae.Rack)
}

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

// ErrRevisionNotAvailable indicates the requested detail for the item have not
// yet been establiehed.
//
type ErrRevisionNotAvailable string

func (erna ErrRevisionNotAvailable) Error() string {
	return fmt.Sprintf("CloudChamber: %q revision not available", string(erna))
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

// ErrIndexKeyValueMismatch indicates the requested boot information for the item
// have not yet been establiehed.
//
type ErrIndexKeyValueMismatch struct {
	Namespace string
	Key       string
	Value     string
}

func (ekvm ErrIndexKeyValueMismatch) Error() string {
	return fmt.Sprintf("CloudChamber: mismatch in index key %q for returned value %q in the %q namespace", ekvm.Key, ekvm.Value, ekvm.Namespace)
}


// ErrfRegionNotFound is a wrapper around the composite literal based error of the related name
//
func ErrfRegionNotFound(region string) error {return ErrRegionNotFound(region)}

// ErrfRegionAlreadyExists is a wrapper around the composite literal based error of the related name
//
func ErrfRegionAlreadyExists(region string) error {return ErrRegionAlreadyExists(region)}

// ErrfZoneNotFound is a wrapper around the composite literal based error of the related name
//
func ErrfZoneNotFound(region string, zone string) error {return ErrZoneNotFound{region, zone}}

// ErrfZoneAlreadyExists is a wrapper around the composite literal based error of the related name
//
func ErrfZoneAlreadyExists(region string, zone string) error {return ErrZoneAlreadyExists{region, zone}}

// ErrfRackNotFound is a wrapper around the composite literal based error of the related name
//
func ErrfRackNotFound(region string, zone string, rack string) error {return ErrRackNotFound{region, zone, rack}}

// ErrfRackAlreadyExists is a wrapper around the composite literal based error of the related name
//
func ErrfRackAlreadyExists(region string, zone string, rack string) error {return ErrRackAlreadyExists{region, zone, rack}}

// ErrfPduIndexInvalid is a wrapper around the composite literal based error of the related name
//
func ErrfPduIndexInvalid(region string, zone string, rack string, pdu string) error {return ErrPduIndexInvalid{region, zone, rack, pdu}}

// ErrfPduNotFound is a wrapper around the composite literal based error of the related name
//
func ErrfPduNotFound(region string, zone string, rack string, pdu int64) error {return ErrPduNotFound{region, zone, rack, pdu}}

// ErrfPduAlreadyExists is a wrapper around the composite literal based error of the related name
//
func ErrfPduAlreadyExists(region string, zone string, rack string, pdu int64) error {return ErrPduAlreadyExists{region, zone, rack, pdu}}

// ErrfTorIndexInvalid is a wrapper around the composite literal based error of the related name
//
func ErrfTorIndexInvalid(region string, zone string, rack string, tor string) error {return ErrTorIndexInvalid{region, zone, rack, tor}}

// ErrfTorNotFound is a wrapper around the composite literal based error of the related name
//
func ErrfTorNotFound(region string, zone string, rack string, tor int64) error {return ErrTorNotFound{region, zone, rack, tor}}

// ErrfTorAlreadyExists is a wrapper around the composite literal based error of the related name
//
func ErrfTorAlreadyExists(region string, zone string, rack string, tor int64) error {return ErrTorAlreadyExists{region, zone, rack, tor}}

// ErrfBladeIndexInvalid is a wrapper around the composite literal based error of the related name
//
func ErrfBladeIndexInvalid(region string, zone string, rack string, blade string) error {return ErrBladeIndexInvalid{region, zone, rack, blade}}

// ErrfBladeNotFound is a wrapper around the composite literal based error of the related name
//
func ErrfBladeNotFound(region string, zone string, rack string, blade int64) error {return ErrBladeNotFound{region, zone, rack, blade}}

// ErrfBladeAlreadyExists is a wrapper around the composite literal based error of the related name
//
func ErrfBladeAlreadyExists(region string, zone string, rack string, blade int64) error {return ErrBladeAlreadyExists{region, zone, rack, blade}}

// ErrfIndexKeyValueMismatch is a wrapper around the composite literal based error of the related name
//
func ErrfIndexKeyValueMismatch(namespace string, key string, value string) error {return ErrIndexKeyValueMismatch{namespace, key, value}}


