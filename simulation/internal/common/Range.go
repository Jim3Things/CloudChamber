package common

import (
	"fmt"
	"math/rand"
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
	if delta > 0 {
		delta = rand.Int63n(delta + 1)
	}

	return dr.Low + delta
}
