// Validation methods for the types from external.proto

package inventory

import (
	"fmt"

	"github.com/Jim3Things/CloudChamber/pkg/protos/common"
)

// Validate is a method that verifies that the associated DefinitionPdu instance
// is structurally legal
//
func(x *DefinitionPdu) Validate(prefix string, ports int64) error {

	actual := int64(len(x.Ports))
	if actual != ports {
		return common.ErrMustBeEQ{
			Field:    fmt.Sprintf("%sPorts", prefix),
			Actual:   actual,
			Required: 1,
		}
	}

	return nil
}

// Validate is a method that verifies that the associated DefinitionTor instance
// is structurally legal
//
func(x *DefinitionTor) Validate(prefix string, ports int64) error {

	actual := int64(len(x.Ports))
	if actual != ports {
		return common.ErrMustBeEQ{
			Field:    fmt.Sprintf("%sPorts", prefix),
			Actual:   actual,
			Required: 1,
		}
	}

	return nil
}

// Validate is a method that verifies that the associated DefinitionRack instance
// is structurally legal
//
func (x *DefinitionRack) Validate(prefix string) error {
	// Verify that rack has at least one Pdu
	//
	// NOTE: at present we expect there to be exactly one Pdu per-rack
	//
	countPdus := int64(len(x.Pdus))
	if countPdus != 1 {
		return common.ErrMustBeEQ{
			Field:    fmt.Sprintf("%sPdus", prefix),
			Actual:   countPdus,
			Required: 1,
		}
	}

	// Verify that rack has at least one Tor
	//
	// NOTE: at present we expect there to be exactly one Tor per-rack
	//
	countTors := int64(len(x.Tors))
	if countTors != 1 {
		return common.ErrMustBeEQ{
			Field:    fmt.Sprintf("%sTors", prefix),
			Actual:   countTors,
			Required: 1,
		}
	}

	// Verify that a rack has at least one blade
	//
	countBlades := int64(len(x.Blades))
	if countBlades < 1 {
		return common.ErrMinLenMap{
			Field:    fmt.Sprintf("%sBlades", prefix),
			Actual:   countBlades,
			Required: 1,
		}
	}

	// Check that there is one Pdu port for each blade
	//
	for k, v := range x.Pdus {
		if err := v.Validate(fmt.Sprintf("%sPdu[%d].", prefix, k), countBlades); err != nil {
			return err
		}
	}

	// Check that there is one Tor port for each blade
	//
	for k, v := range x.Tors {
		if err := v.Validate(fmt.Sprintf("%sTor[%d].", prefix, k), countBlades); err != nil {
			return err
		}
	}	

	// .. And then validate that each blade is valid
	//
	for k, v := range x.Blades {
		if err := v.Capacity.Validate(fmt.Sprintf("%sBlades[%d].", prefix, k)); err != nil {
			return err
		}
	}

	// All correct
	return nil
}

// Validate is a method that verifies that the associated DefinitionZone instance
// is structurally legal
//
func (x *DefinitionZone) Validate() error {
	// Verify that zone has at least one rack
	//
	actual := int64(len(x.Racks))
	if actual < 1 {
		return common.ErrMinLenMap{
			Field:    "Racks",
			Actual:   actual,
			Required: 1,
		}
	}

	// .. And then validate that each rack is valid
	for k, v := range x.Racks {
		if err := v.Validate(fmt.Sprintf("Rack[%s].", k)); err != nil {
			return err
		}
	}


	// All correct
	return nil
}
