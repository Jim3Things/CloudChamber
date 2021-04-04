// Assorted utility routines for types from the various protobuf definitions

package inventory

func (hw *Hardware) EqualHardware(item *Hardware) bool {
	return hw.Type == item.Type && hw.Id == item.Id
}

func (pp *PowerPort) EqualPort(port *PowerPort) bool {
	return pp.Wired == port.Wired && pp.Item.EqualHardware(port.Item)
}

func (np *NetworkPort) EqualPort(port *NetworkPort) bool {
	return np.Wired == port.Wired && np.Item.EqualHardware(port.Item)
}
