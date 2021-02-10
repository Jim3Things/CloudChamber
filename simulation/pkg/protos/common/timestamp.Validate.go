// Validation methods for the types from timestamp.proto

package common

import (
    "github.com/Jim3Things/CloudChamber/simulation/pkg/errors"
)

// Validate is a method that verifies that the associated Timestamp instance
// is structurally legal
func (x *Timestamp) Validate() error {
	// No negative time allowed
	if x.Ticks < 0 {
		return errors.ErrMustBeGTE{Field: "Ticks", Actual: x.Ticks, Required: 0}
	}

	// All correct
	return nil
}
