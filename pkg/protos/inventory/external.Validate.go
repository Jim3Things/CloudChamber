// Validation methods for the types from external.proto

package inventory

import (
    "fmt"

    "github.com/Jim3Things/CloudChamber/pkg/errors"
)

// Validate is a method that verifies that the associated ExternalRack instance
// is structurally legal
func (x *ExternalRack) Validate() error {
	// Verify that a rack has at least one blade
	actual := int64(len(x.Blades))
	if actual < 1 {
		return errors.ErrMinLenMap{
			Field:    "Blades",
			Actual:   actual,
			Required: 1,
		}
	}

	// .. And then validate that each blade is valid
	for k, v := range x.Blades {
		if err := v.Validate(fmt.Sprintf("Blades[%d].", k)); err != nil {
			return err
		}
	}

	// All correct
	return nil
}
