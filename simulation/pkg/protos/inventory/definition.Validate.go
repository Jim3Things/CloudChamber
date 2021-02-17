// Validation methods for the types from external.proto

package inventory

import (
	"fmt"

	"github.com/Jim3Things/CloudChamber/simulation/pkg/errors"
)

const (
	minPortsPerPdu   = int64(1)
	maxPortsPerPdu   = int64(128)
	minPortsPerTor   = int64(1)
	maxPortsPerTor   = int64(1024)
	minPortsPerBlade = int64(1)
	maxPortsPerBlade = int64(10000)
	minPdusPerRack   = int64(1)
	maxPdusPerRack   = int64(2)
	minTorsPerRack   = int64(1)
	maxTorsPerRack   = int64(2)
	minBladesPerRack = int64(1)
	maxBladesPerRack = int64(1000)
)

// Validate is a method that verifies that the associated DefinitionPdu instance
// has the expected number of ports and is semantically legal
//
// ports is the minimum number of ports that must be present.
//
func (x *Definition_Pdu) Validate(prefix string, minPorts int64) error {

	actual := int64(len(x.Ports))

	if actual < minPorts {
		return errors.ErrMustBeGTE{
			Field:    fmt.Sprintf("%sPorts", prefix),
			Actual:   actual,
			Required: minPorts,
		}
	}

	return x.check(prefix)
}

// check is a method that verifies that the associated DefinitionPdu instance
// is semantically legal
//
// For example, check for errors such as the pdu being wired to itself, the
// port being wired but without an associated item etc.
//
func (x *Definition_Pdu) check(prefix string) error {

	prefixedPorts := prefix + "Ports"
	prefixedItem := prefix + "Item"

	portCount := int64(len(x.Ports))

	if portCount < minPortsPerPdu {
		return errors.ErrMinLenMap{
			Field:    prefixedPorts,
			Actual:   portCount,
			Required: minPortsPerPdu,
		}
	}

	if portCount > maxPortsPerPdu {
		return errors.ErrMaxLenMap{
			Field:  prefixedPorts,
			Actual: portCount,
			Limit:  maxPortsPerPdu,
		}
	}

	for i, p := range x.Ports {
		if !p.Wired {
			if p.Item != nil {
				// port not wired but has an (unexpected) associated item
				//
				return errors.ErrItemMustBeEmpty{
					Field:  prefixedItem,
					Item:   "PDU",
					Port:   i,
					Actual: p.Item.String(),
				}
			}
		} else {
			// port is wired
			//
			if p.Item == nil {
				// port is wired but is missing an (expected) associated item
				//
				return errors.ErrItemMissingValue{
					Field: prefixedItem,
					Item:  "PDU",
					Port:  i,
				}
			}

			// Port is wired and we have a description. To keep things
			// simple, a Pdu is not allowed to be wired to another Pdu.
			// While this prevents a potentially legal case of chained
			// Pdus, it also prevents wiring a Pdu to itself.
			//
			if p.Item.Type == Hardware_pdu {
				return errors.ErrInvalidItemSelf{
					Field:  prefixedItem,
					Item:   "PDU",
					Port:   i,
					Actual: "PDU",
				}
			}
		}
	}

	return nil
}

// Validate is a method that verifies that the associated DefinitionTor instance
// has the expected number of ports and is semantically legal
//
// ports is the minimum number of ports that must be present.
//
func (x *Definition_Tor) Validate(prefix string, minPorts int64) error {

	actual := int64(len(x.Ports))

	if actual < minPorts {
		return errors.ErrMustBeEQ{
			Field:    fmt.Sprintf("%sPorts", prefix),
			Actual:   actual,
			Required: minPorts,
		}
	}

	return x.check(prefix)
}

// check is a method that verifies that the associated DefinitionTor instance
// is semantically legal
//
// For example, check for errors such as the tor being wired to itself, the
// port being wired but without an associated item etc.
//
func (x *Definition_Tor) check(prefix string) error {

	prefixedPorts := prefix + "Ports"
	prefixedItem := prefix + "Item"

	portCount := int64(len(x.Ports))

	if portCount < minPortsPerTor {
		return errors.ErrMinLenMap{
			Field:    prefixedPorts,
			Actual:   portCount,
			Required: minPortsPerTor,
		}
	}

	if portCount > maxPortsPerTor {
		return errors.ErrMaxLenMap{
			Field:  prefixedPorts,
			Actual: portCount,
			Limit:  maxPortsPerTor,
		}
	}

	for i, p := range x.Ports {
		if !p.Wired {
			if p.Item != nil {
				// port not wired but has an (unexpected) associated item
				//
				return errors.ErrItemMustBeEmpty{
					Field:  prefixedItem,
					Item:   "TOR",
					Port:   i,
					Actual: p.Item.String(),
				}
			}
		} else {
			// port is wired
			//
			if p.Item == nil {
				// port is wired but is missing an (expected) associated item
				//
				return errors.ErrItemMissingValue{
					Field: prefixedItem,
					Item:  "TOR",
					Port:  i,
				}
			}

			// Port is wired and we have a description. To keep things
			// simple, a Pdu is not allowed to be wired to another Pdu.
			// While this prevents a potentially legal case of chained
			// Pdus, it also prevents wiring a Pdu to itself.
			//
			if p.Item.Type == Hardware_tor {
				return errors.ErrInvalidItemSelf{
					Field:  prefixedItem,
					Item:   "TOR",
					Port:   i,
					Actual: "TOR",
				}
			}
		}
	}

	return nil
}

// Validate is a method that verifies that the associated DefinitionBlade instance
// has the expected number of cors etc and is semantically legal
//
func (x *Definition_Blade) Validate(prefix string) error {

	// Validate that the capacity is valid
	//
	if err := x.Capacity.Validate(fmt.Sprintf("%s", prefix)); err != nil {
			return err
	}

	return nil
}

// Validate is a method that verifies that the associated DefinitionRack instance
// is structurally legal
//
func (x *Definition_Rack) Validate(prefix string) error {

	// Verify that rack has at least the minimum number of pdus..., 
	//
	countPdus := int64(len(x.Pdus))

	if countPdus < minPdusPerRack {
		return errors.ErrMustBeGTE{
			Field:    fmt.Sprintf("%sPdus", prefix),
			Actual:   countPdus,
			Required: minPdusPerRack,
		}
	}

	// ... but no more than the maximum number of pdus
	//
	if countPdus > maxPdusPerRack {
		return errors.ErrMustBeLTE{
			Field:    fmt.Sprintf("%sPdus", prefix),
			Actual:   countPdus,
			Required: maxPdusPerRack,
		}
	}

	// Verify that rack has at least the minimum number of tors..., 
	//
	countTors := int64(len(x.Tors))

	if countTors < minTorsPerRack {
		return errors.ErrMustBeGTE{
			Field:    fmt.Sprintf("%sTors", prefix),
			Actual:   countTors,
			Required: minTorsPerRack,
		}
	}

	// ... but no more than the maximum number of tors
	//
	if countTors > maxTorsPerRack {
		return errors.ErrMustBeLTE{
			Field:    fmt.Sprintf("%sTors", prefix),
			Actual:   countTors,
			Required: maxTorsPerRack,
		}
	}

	// Verify that rack has at least the minimum number of blades..., 
	//
	countBlades := int64(len(x.Blades))

	if countBlades < minBladesPerRack {
		return errors.ErrMustBeGTE{
			Field:    fmt.Sprintf("%sBlades", prefix),
			Actual:   countBlades,
			Required: minBladesPerRack,
		}
	}

	// ... but no more than the maximum number of blades
	//
	if countBlades > maxBladesPerRack {
		return errors.ErrMustBeLTE{
			Field:    fmt.Sprintf("%sBlades", prefix),
			Actual:   countBlades,
			Required: maxBladesPerRack,
		}
	}


	// Check that there is one Pdu port for each blade
	//
	for k, v := range x.Pdus {
		if err := v.Validate(fmt.Sprintf("%sPdus[%d].", prefix, k), countBlades); err != nil {
			return err
		}
	}

	// Check that there is one Tor port for each blade
	//
	for k, v := range x.Tors {
		if err := v.Validate(fmt.Sprintf("%sTors[%d].", prefix, k), countBlades); err != nil {
			return err
		}
	}

	// .. And then validate that each blade is valid
	//
	for k, v := range x.Blades {
		if err := v.Validate(fmt.Sprintf("%sBlades[%d].", prefix, k)); err != nil {
			return err
		}
	}

	// All correct
	//
	return nil
}

// Validate is a method that verifies that the associated DefinitionZone instance
// is structurally legal
//
func (x *Definition_Zone) Validate(prefix string) error {
	// Verify that zone has at least one rack
	//
	actual := int64(len(x.Racks))
	if actual < 1 {
		return errors.ErrMinLenMap{
			Field:    "Racks",
			Actual:   actual,
			Required: 1,
		}
	}

	// .. And then validate that each rack is valid
	//
	for k, v := range x.Racks {
		if err := v.Validate(fmt.Sprintf("%sRacks[%s].", prefix, k)); err != nil {
			return err
		}
	}

	// All correct
	//
	return nil
}

// Validate is a method that verifies that the associated DefinitionZone instance
// is structurally legal
//
func (x *Definition_Region) Validate(prefix string) error {
	// Verify that zone has at least one rack
	//
	actual := int64(len(x.Zones))
	if actual < 1 {
		return errors.ErrMinLenMap{
			Field:    "Zones",
			Actual:   actual,
			Required: 1,
		}
	}

	// .. And then validate that each rack is valid
	//
	for k, v := range x.Zones {
		if err := v.Validate(fmt.Sprintf("%sZones[%s].", prefix, k)); err != nil {
			return err
		}
	}

	// All correct
	//
	return nil
}
