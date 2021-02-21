package inventory

// Validation methods for the types from capacity.proto.

import (
	"fmt"

	"github.com/Jim3Things/CloudChamber/simulation/pkg/errors"
)

const (
	minCors                   = int64(1)
	maxCors                   = int64(16384)
	minMemoryInMb             = int64(1)
	maxMemoryInMb             = int64(16 * 1024 * 1024) // 16TB
	minDiskInGb               = int64(1)
	maxDiskInGb               = int64(1024 * 1024 * 1024) // 1PB
	minNetworkBandwidthInMbps = int64(1)
	maxNetworkBandwidthInMbps = int64(1024 * 1024) // 1Tbps
)

// Validate is a function to ensure that the blade capacity values are legal.
// Note that since BladeCapacity is always used as a subfield, the Validate
// function takes a prefix to use on the field name in order to place the
// error correctly.
func (x *BladeCapacity) Validate(prefix string) error {
	// A blade must have at least the minimum number of cores...
	//
	if x.Cores < minCors {
		return errors.ErrMustBeGTE{
			Field:    fmt.Sprintf("%sCores", prefix),
			Actual:   x.Cores,
			Required: minCors,
		}
	}

	// ... but no more than the maximum number of cores
	//
	if x.Cores > maxCors {
		return errors.ErrMustBeLTE{
			Field:    fmt.Sprintf("%sCores", prefix),
			Actual:   x.Cores,
			Required: maxCors,
		}
	}

	// And it must have some memory...
	//
	if x.MemoryInMb < minMemoryInMb {
		return errors.ErrMustBeGTE{
			Field:    fmt.Sprintf("%sMemoryInMb", prefix),
			Actual:   x.MemoryInMb,
			Required: minMemoryInMb,
		}
	}

	//...but no more than some sensible upper amount
	//
	if x.MemoryInMb > maxMemoryInMb {
		return errors.ErrMustBeLTE{
			Field:    fmt.Sprintf("%sMemoryInMb", prefix),
			Actual:   x.MemoryInMb,
			Required: maxMemoryInMb,
		}
	}

	// And some disk space...
	//
	if x.DiskInGb < minDiskInGb {
		return errors.ErrMustBeGTE{
			Field:    fmt.Sprintf("%sDiskInGb", prefix),
			Actual:   x.DiskInGb,
			Required: minDiskInGb,
		}
	}

	// ...but no more than some sensible upper amount
	//
	if x.DiskInGb > maxDiskInGb {
		return errors.ErrMustBeGTE{
			Field:    fmt.Sprintf("%sDiskInGb", prefix),
			Actual:   x.DiskInGb,
			Required: maxDiskInGb,
		}
	}

	// And a network bandwidth allowance...
	//
	if x.NetworkBandwidthInMbps < minNetworkBandwidthInMbps {
		return errors.ErrMustBeGTE{
			Field:    fmt.Sprintf("%sNetworkBandwidthInMbps", prefix),
			Actual:   x.NetworkBandwidthInMbps,
			Required: minNetworkBandwidthInMbps,
		}
	}

	// ...but no more than some sensible upper amount
	//
	if x.NetworkBandwidthInMbps > maxNetworkBandwidthInMbps {
		return errors.ErrMustBeLTE{
			Field:    fmt.Sprintf("%sNetworkBandwidthInMbps", prefix),
			Actual:   x.NetworkBandwidthInMbps,
			Required: maxNetworkBandwidthInMbps,
		}
	}

	return nil
}
