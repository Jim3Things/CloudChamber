package common

// Validation methods for the types from capacity.proto.

import (
	"fmt"
)

// Validate is a function to ensure that the blade capacity values are legal.
// Note that since BladeCapacity is always used as a subfield, the Validate
// function takes a prefix to use on the field name in order to place the
// error correctly.
func (x *BladeCapacity) Validate(prefix string) error {
	// A blade must have at least one core
	if x.Cores < 1 {
		return ErrMustBeGTE{
			Field:    fmt.Sprintf("%sCores", prefix),
			Actual:   x.Cores,
			Required: 1,
		}
	}

	// .. and it must have some memory
	if x.MemoryInMb < 1 {
		return ErrMustBeGTE{
			Field:    fmt.Sprintf("%sMemoryInMb", prefix),
			Actual:   x.MemoryInMb,
			Required: 1,
		}
	}

	// .. and some disk space
	if x.DiskInGb < 1 {
		return ErrMustBeGTE{
			Field:    fmt.Sprintf("%sDiskInGb", prefix),
			Actual:   x.DiskInGb,
			Required: 1,
		}
	}

	// .. and a network bandwidth allowance
	if x.NetworkBandwidthInMbps < 1 {
		return ErrMustBeGTE{
			Field:    fmt.Sprintf("%sNetworkBandwidthInMbps", prefix),
			Actual:   x.NetworkBandwidthInMbps,
			Required: 1,
		}
	}

	return nil
}
