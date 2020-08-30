// Validation methods for the types from Stepper.proto
package services

import (
	"github.com/Jim3Things/CloudChamber/pkg/protos/common"
)

// Validate is a function to ensure that the policy request values are
// legal
func (x *PolicyRequest) Validate() error {
	// The policy must be within the valid range for the enum
	switch x.Policy {
	case StepperPolicy_NoWait,
		StepperPolicy_Measured,
		StepperPolicy_Manual:
		return nil

	default:
		// Not a valid policy choice
		return common.ErrInvalidEnum{
			Field:  "Policy",
			Actual: int64(x.Policy),
		}
	}
}

// Validate is a function to ensure that the single step request values are
// legal
func (x *StepRequest) Validate() error {
	return nil
}

// Validate is a function to ensure that the get current time request values
// are legal
func (x *NowRequest) Validate() error {
	return nil
}

// Validate is a function to ensure that the wait until request values are
// legal
func (x *DelayRequest) Validate() error {
	// The minimum time must be valid
	if err := x.AtLeast.Validate(); err != nil {
		return err
	}

	// .. and the jitter allowance cannot be negative
	if x.Jitter < 0 {
		return common.ErrMustBeGTE{
			Field:    "Jitter",
			Actual:   x.Jitter,
			Required: 0,
		}
	}

	// All ok
	return nil
}

// Validate is a function to ensure that the reset service request values are
// legal
func (x *ResetRequest) Validate() error {
	return nil
}

func (x *GetStatusRequest) Validate() error {
	return nil
}
