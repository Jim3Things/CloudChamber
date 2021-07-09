package inventory

func (x *BladeBootInfo) Clone() *BladeBootInfo {
	if x == nil {
		return nil
	}

	return &BladeBootInfo{
		Source:     x.Source,
		Image:      x.Image,
		Version:    x.Version,
		Parameters: x.Parameters,
	}
}

func (x *BladeBootInfo) Equal(i *BladeBootInfo) bool {
	return x.GetSource() == i.GetSource() &&
		x.GetImage() == i.GetImage() &&
		x.GetVersion() == i.GetVersion() &&
		x.GetParameters() == i.GetParameters()
}


func (x *BladeDetails) Clone() *BladeDetails {
	if x == nil {
		return nil
	}

	return &BladeDetails{
		Enabled: x.Enabled,
		Condition: x.Condition,
	}
}

func (x *BladeDetails) Equal(d *BladeDetails) bool {
	return x.GetEnabled() == d.GetEnabled() && x.GetCondition() == d.GetCondition()
}

func (x *Hardware) Clone() *Hardware {
	if x == nil {
		return nil
	}

	return &Hardware{
		Type: x.Type,
		Id:   x.Id,
		Port: x.Port,
	}
}

func (x *Hardware) Equal(h *Hardware) bool {
	return x.GetType() == h.GetType() && x.GetId() == h.GetId()
}

func (x *NetworkPort) Clone() *NetworkPort {
	if x == nil {
		return nil
	}

	return &NetworkPort{
		Wired: x.Wired,
		Item:  x.Item.Clone(),
	}
}

func (x *NetworkPort) Equal(p *NetworkPort) bool {
	return x.GetWired() == p.GetWired() && x.Item.Equal(p.Item)
}

func (x *PduDetails) Clone() *PduDetails {
	if x == nil {
		return nil
	}

	return &PduDetails {
		Enabled:   x.Enabled,
		Condition: x.Condition,
	}
}

func (x *PduDetails) Equal(d *PduDetails) bool {
	return x.GetEnabled() == d.GetEnabled() && x.GetCondition() == d.GetCondition()
}

func (x *PowerPort) Clone() *PowerPort {
	if x == nil {
		return nil
	}

	return &PowerPort{
		Wired: x.Wired,
		Item:  x.Item.Clone(),
	}
}

func (x *PowerPort) Equal(p *PowerPort) bool {
	return x.GetWired() == p.GetWired() && x.Item.Equal(p.Item)
}

func (x *RackDetails) Clone() *RackDetails {
	if x == nil {
		return nil
	}

	return &RackDetails{
		Enabled:   x.Enabled,
		Condition: x.Condition,
		Location:  x.Location,
		Notes:     x.Notes,
	}
}

func (x *RackDetails) Equal(d *RackDetails) bool {
	return x.GetEnabled() == d.GetEnabled() &&
		x.GetCondition() == d.GetCondition() &&
		x.GetLocation() == d.GetLocation() &&
		x.GetNotes() == d.GetNotes()
}

func (x *RegionDetails) Clone() *RegionDetails {
	if x == nil {
		return nil
	}

	return &RegionDetails{
		Name:     x.Name,
		State:    x.State,
		Location: x.Location,
		Notes:    x.Notes,
	}
}

func (x *RegionDetails) Equal(d *RegionDetails) bool {
	return x.GetName() == d.GetName() &&
		x.GetState() == d.GetState() &&
		x.GetLocation() == d.GetLocation() &&
		x.GetNotes() == d.GetNotes()
}

func (x *RootDetails) Clone() *RootDetails {
	if x == nil {
		return nil
	}

	return &RootDetails{
		Name:  x.Name,
		Notes: x.Notes,
	}
}

func (x *RootDetails) Equal(d *RootDetails) bool {
	return x.GetName() == d.GetName() && x.GetNotes() == d.GetNotes()
}

func (x *TorDetails) Clone() *TorDetails {
	if x == nil {
		return nil
	}

	return &TorDetails{
		Enabled:   x.Enabled,
		Condition: x.Condition,
	}
}

func (x *TorDetails) Equal(d *TorDetails) bool {
	return x.GetEnabled() == d.GetEnabled() && x.GetCondition() == d.GetCondition()
}

func (x *ZoneDetails) Clone() *ZoneDetails {
	if x == nil {
		return nil
	}

	return &ZoneDetails{
		Enabled:  x.Enabled,
		State:    x.State,
		Location: x.Location,
		Notes:    x.Notes,
	}
}

func (x *ZoneDetails) Equal(d *ZoneDetails) bool {
	return x.GetEnabled() == d.GetEnabled() &&
		x.GetState() == d.GetState() &&
		x.GetLocation() == d.GetLocation() &&
		x.GetNotes() == d.GetNotes()
}
