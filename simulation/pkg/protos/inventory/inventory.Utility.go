// Assorted utility routines for types from the various protobuf definitions

package inventory

func (hw *Hardware) Equal(h *Hardware) bool {
	return hw.GetType() == h.GetType() && hw.GetId() == h.GetId()
}

func (pp *PowerPort) Equal(p *PowerPort) bool {
	return pp.GetWired() == p.GetWired() && pp.Item.Equal(p.Item)
}

func (np *NetworkPort) Equal(p *NetworkPort) bool {
	return np.GetWired() == p.GetWired() && np.Item.Equal(p.Item)
}

func (rd *RootDetails) Equal(d *RootDetails) bool {
	return rd.GetName() == d.GetName() && rd.GetNotes() == d.GetNotes()
}

func (rd *RegionDetails) Equal(d *RegionDetails) bool {
	return rd.GetName() == d.GetName() &&
		   rd.GetState() == d.GetState() &&
		   rd.GetLocation() == d.GetLocation() &&
		   rd.GetNotes() == d.GetNotes()
}

func (zd *ZoneDetails) Equal(d *ZoneDetails) bool {
	return zd.GetEnabled() == d.GetEnabled() &&
		   zd.GetState() == d.GetState() &&
		   zd.GetLocation() == d.GetLocation() &&
		   zd.GetNotes() == d.GetNotes()
}

func (rd *RackDetails) Equal(d *RackDetails) bool {
	return rd.GetEnabled() == d.GetEnabled() &&
		   rd.GetCondition() == d.GetCondition() &&
		   rd.GetLocation() == d.GetLocation() &&
		   rd.GetNotes() == d.GetNotes()
}

func (pd *PduDetails) Equal(d *PduDetails) bool {
	return pd.GetEnabled() == d.GetEnabled() && pd.GetCondition() == d.GetCondition()
}

func (td *TorDetails) Equal(d *TorDetails) bool {
	return td.GetEnabled() == d.GetEnabled() && td.GetCondition() == d.GetCondition()
}

func (bd *BladeDetails) Equal(d *BladeDetails) bool {
	return bd.GetEnabled() == d.GetEnabled() && bd.GetCondition() == d.GetCondition()
}

func (bc *BladeCapacity) Equal(c *BladeCapacity) bool {
	return bc.GetArch() == c.GetArch() &&
		   bc.GetCores() == c.GetCores() &&
		   bc.GetDiskInGb() == c.GetDiskInGb() &&
		   bc.GetMemoryInMb() == c.GetMemoryInMb() &&
		   bc.GetNetworkBandwidthInMbps() == c.GetNetworkBandwidthInMbps()
}

func (bi *BladeBootInfo) Equal(i *BladeBootInfo) bool {
	return bi.GetSource() == i.GetSource() &&
		   bi.GetImage() == i.GetImage() &&
		   bi.GetVersion() == i.GetVersion() &&
		   bi.GetParameters() == i.GetParameters()
}

func (dt *Definition_Pdu) EqualPorts(d *Definition_Pdu) bool {
	switch {
	case dt == nil && d == nil:
		return true

	case (dt == nil) != (d == nil):
		return false
	}

	dtPorts := dt.GetPorts()
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

func (dt *Definition_Pdu) Equal(d *Definition_Pdu) bool {
	switch {
	case dt == nil && d == nil:
		return true

	case (dt == nil) != (d == nil):
		return false
	}

	return dt.Details.Equal(d.GetDetails()) && dt.EqualPorts(d)
}

func (dt *Definition_Tor) EqualPorts(d *Definition_Tor) bool {
	switch {
	case dt == nil && d == nil:
		return true

	case (dt == nil) != (d == nil):
		return false
	}

	dtPorts := dt.GetPorts()
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

func (dt *Definition_Tor) Equal(d *Definition_Tor) bool {
	switch {
	case dt == nil && d == nil:
		return true

	case (dt == nil) != (d == nil):
		return false
	}

	return dt.Details.Equal(d.GetDetails()) && dt.EqualPorts(d)
}

func (db *Definition_Blade) Equal(d *Definition_Blade) bool {
	switch {
	case db == nil && d == nil:
		return true

	case (db == nil) != (d == nil):
		return false
	}

	return db.Details.Equal(d.GetDetails()) &&
		   db.Capacity.Equal(d.GetCapacity()) &&
		   db.BootInfo.Equal(d.GetBootInfo()) &&
		   db.GetBootOnPowerOn() == d.GetBootOnPowerOn()
}
