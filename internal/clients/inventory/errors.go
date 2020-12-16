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
	return fmt.Sprintf("CloudChamber: bladeID %q is out of range (0 to %v)", string(ebii), maxBladeID)
}

// ErrPduIDInvalid indicates the supplied pduID was out of range, either < less than 0 or greater than maxPduID
//
type ErrPduIDInvalid int64

func (epii ErrPduIDInvalid) Error() string {
	return fmt.Sprintf("CloudChamber: pduID %q is out of range (0 to %v)", string(epii), maxPduID)
}

// ErrTorIDInvalid indicates the supplied torID was out of range, either < less than 0 or greater than maxTorID
//
type ErrTorIDInvalid int64

func (etii ErrTorIDInvalid) Error() string {
	return fmt.Sprintf("CloudChamber: torID %q is out of range (0 to %v)", string(etii), maxTorID)
}

