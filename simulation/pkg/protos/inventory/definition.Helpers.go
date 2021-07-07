// Assorted utility routines for types from the various protobuf definitions

package inventory

func (x *Definition_Pdu) EqualPorts(d *Definition_Pdu) bool {
	switch {
	case x == nil && d == nil:
		return true

	case (x == nil) != (d == nil):
		return false
	}

	dtPorts := x.GetPorts()
	dPorts := d.GetPorts()

	if len(dtPorts) != len(dPorts) {
		return false
	}

	for i, pp := range dtPorts {
		if !pp.Equal(dPorts[i]) {
			return false
		}
	}

	return true
}

func (x *Definition_Pdu) Equal(d *Definition_Pdu) bool {
	switch {
	case x == nil && d == nil:
		return true

	case (x == nil) != (d == nil):
		return false
	}

	return x.Details.Equal(d.GetDetails()) && x.EqualPorts(d)
}

func (x *Definition_Tor) EqualPorts(d *Definition_Tor) bool {
	switch {
	case x == nil && d == nil:
		return true

	case (x == nil) != (d == nil):
		return false
	}

	dtPorts := x.GetPorts()
	dPorts := d.GetPorts()

	if len(dtPorts) != len(dPorts) {
		return false
	}

	for i, pp := range dtPorts {
		if !pp.Equal(dPorts[i]) {
			return false
		}
	}

	return true
}

func (x *Definition_Tor) Equal(d *Definition_Tor) bool {
	switch {
	case x == nil && d == nil:
		return true

	case (x == nil) != (d == nil):
		return false
	}

	return x.Details.Equal(d.GetDetails()) && x.EqualPorts(d)
}

func (x *Definition_Blade) Equal(d *Definition_Blade) bool {
	switch {
	case x == nil && d == nil:
		return true

	case (x == nil) != (d == nil):
		return false
	}

	return x.Details.Equal(d.GetDetails()) &&
		   x.Capacity.Equal(d.GetCapacity()) &&
		   x.BootInfo.Equal(d.GetBootInfo()) &&
		   x.GetBootOnPowerOn() == d.GetBootOnPowerOn()
}
