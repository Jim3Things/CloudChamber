package messages

const (
	CapacityCores   = "_cores"
	CapacityMemory  = "_memoryInMB"
	CapacityDisk    = "_diskInGB"
	CapacityNetwork = "_networkBandwidthInMbps"
)

const (
	// AcceleratorPrefix is put in front of any accelerator name to ensure that
	// there is no collision with the core Capacity categories listed above.
	AcceleratorPrefix = "a_"
)

// Capacity defines the consumable and capability portions of a blade or
// workload.
type Capacity struct {
	// Consumables are named units of capacity that are used by a workload such
	// that the amount available to other workloads is reduced by that amount.
	// For example, a core may only be used by one workload at a time.
	Consumables map[string]float64

	// Features are statements of capabilities that are available for use, but
	// that are not consumed when used.  For example, the presence of security
	// enclave support would be a feature.
	Features map[string]bool
}

// Clone produces a copy of this Capacity instance.
func (s *Capacity) Clone() *Capacity {
	c := NewCapacity()

	for k, v := range s.Consumables {
		c.Consumables[k] = v
	}

	for k, v := range s.Features {
		c.Features[k] = v
	}

	return c
}

// NewCapacity creates a new, blank, Capacity instance.
func NewCapacity() *Capacity {
	return &Capacity{
		Consumables: make(map[string]float64),
		Features:    make(map[string]bool),
	}
}
