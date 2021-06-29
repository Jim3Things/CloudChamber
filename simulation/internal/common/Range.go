package common

import (
	"fmt"
	"math/rand"

	"github.com/Jim3Things/CloudChamber/simulation/pkg/errors"
)

// Range contains a pair of values that form a range of acceptable int64 values,
// low to high, inclusive.
type Range struct {
	Low  int64
	High int64
}

// String returns a formatted description of the range of values as '[low:high]'
func (dr Range) String() string {
	return fmt.Sprintf("%d:%d", dr.Low, dr.High)
}

// Pick returns a value in the acceptable range.
func (dr Range) Pick() int64 {
	delta := dr.High - dr.Low
	switch {
	case delta > 0:
		delta = rand.Int63n(delta + 1)

	case delta < 0:
		// High is illegal.  Since that should have been checked earlier, either
		// in writing the code or via Validate, we simply assert that it is the
		// same as Low here, and therefore force a return of Low.
		delta = 0
	}

	return dr.Low + delta
}

// Validate checks whether or not the values are within an acceptable range,
// and that Low is not greater than High.  It returns an error if the validation
// fails, or nil, if the range is valid.
func (dr Range) Validate(name string, minLow int64, maxHigh int64) error {
	if dr.Low < minLow {
		return &errors.ErrMustBeGTE{
			Field:    fmt.Sprintf("%s.Low", name),
			Actual:   dr.Low,
			Required: minLow,
		}
	}

	if dr.High > maxHigh {
		return &errors.ErrMustBeLTE{
			Field:    fmt.Sprintf("%s.High", name),
			Actual:   dr.High,
			Required: maxHigh,
		}
	}

	if dr.High < dr.Low {
		return &errors.ErrMustBeGTE{
			Field:    fmt.Sprintf("%s.High", name),
			Actual:   dr.High,
			Required: dr.Low,
		}
	}

	return nil
}
