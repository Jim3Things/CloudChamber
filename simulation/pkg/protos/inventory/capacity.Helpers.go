package inventory


func (x *Accelerator) Clone() *Accelerator {
	if x == nil {
		return nil
	}

	return &Accelerator{
		AcceleratorType: x.AcceleratorType,
	}
}

func (x *Accelerator) Equal(c *Accelerator) bool {
	return x.GetAcceleratorType() == c.GetAcceleratorType()
}

func (x *BladeCapacity) Clone() *BladeCapacity {
	if x == nil {
		return nil
	}

	var accelerators []*Accelerator = nil

	if x.Accelerators != nil {
		accelerators = make([]*Accelerator, len(x.Accelerators))

		for i, a := range x.Accelerators {
			accelerators[i] = a.Clone()
		}
	}

	return &BladeCapacity{
		Cores:                  x.Cores,
		MemoryInMb:             x.MemoryInMb,
		DiskInGb:               x.DiskInGb,
		NetworkBandwidthInMbps: x.NetworkBandwidthInMbps,
		Arch:                   x.Arch,
		Accelerators:           accelerators,
	}
}

func (x *BladeCapacity) Equal(c *BladeCapacity) bool {
	accelMatch := len(x.GetAccelerators()) == len(c.GetAccelerators())

	for i := 0; accelMatch && i < len(x.Accelerators); i++ {
		accelMatch = accelMatch && x.Accelerators[i].Equal(c.Accelerators[i])
	}

	return accelMatch &&
		x.GetArch() == c.GetArch() &&
		x.GetCores() == c.GetCores() &&
		x.GetDiskInGb() == c.GetDiskInGb() &&
		x.GetMemoryInMb() == c.GetMemoryInMb() &&
		x.GetNetworkBandwidthInMbps() == c.GetNetworkBandwidthInMbps()
}
