package services

import (
	"fmt"
)

// Describe returns a formatted string with the details of this message.
func (x *StatusResponse) Describe() string {
	if x == nil {
		return "Missing"
	}

	return fmt.Sprintf("At: %d, Policy: %s(%s), waiters: %d, epoch: %d",
		x.Now,
		x.Policy.String(),
		x.MeasuredDelay.String(),
		x.WaiterCount,
		x.Epoch)
}
