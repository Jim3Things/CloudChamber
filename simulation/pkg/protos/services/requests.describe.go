package services

import (
	"fmt"
)

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
