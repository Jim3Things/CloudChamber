package messages

// This file contains the implementation for the component in the repair
// messages that designates the target simulated inventory element for the
// repair.

import (
	"fmt"

	"github.com/Jim3Things/CloudChamber/simulation/pkg/protos/inventory"
)

// Define the hardware type as used for message addressing.
//
// NB: This should eventually be merged with / superseded by the address types
// that we're setting up in the store.  They should fully overlap in feature,
// at which point we can remove this one and convert the inventory simulation to
// use the common structure.

type target int

const (
	targetInvalid target = iota
	targetTor
	targetPdu
	targetBlade
)

func (t target) String() string {
	switch t {

	case targetTor:
		return "tor"

	case targetPdu:
		return "pdu"

	case targetBlade:
		return "blade"

	default:
		return "unknown"
	}
}

// MessageTarget describes the target element for the enclosing message.
type MessageTarget struct {
	element target
	Rack    string
	itemId  int64
	port    int64
}

// NewTargetTor creates a new target instance that specifies the designated
// rack's TOR.
func NewTargetTor(name string, id int64, port int64) *MessageTarget {
	return &MessageTarget{
		element: targetTor,
		Rack:    name,
		itemId:  id,
		port:    port,
	}
}

// NewTargetPdu creates a new target instance that specifies the designated
// rack's PDU.
func NewTargetPdu(name string, id int64, port int64) *MessageTarget {
	return &MessageTarget{
		element: targetPdu,
		Rack:    name,
		itemId:  id,
		port:    port,
	}
}

// NewTargetBlade creates a new target instance that specifies the designated
// itemId in the specified rack.
func NewTargetBlade(name string, id int64, port int64) *MessageTarget {
	return &MessageTarget{
		element: targetBlade,
		Rack:    name,
		itemId:  id,
		port:    port,
	}
}

func (m *MessageTarget) rackName() string { return m.Rack }
func (m *MessageTarget) IsTor() bool      { return m.element == targetTor }
func (m *MessageTarget) IsPdu() bool      { return m.element == targetPdu }
func (m *MessageTarget) IsBlade() bool    { return m.element == targetBlade }
func (m *MessageTarget) ElementId() int64 { return m.itemId }
func (m *MessageTarget) Port() int64      { return m.port }

// Describe produces a string that describes the target element's logical
// address.
func (m *MessageTarget) Describe() string {
	return fmt.Sprintf(
		"%s %d:%d in rack %q",
		m.element.String(), m.ElementId(), m.Port(), m.rackName())
}

// Key produces a structured string that can be used simply as a rack-local id
// for an element in the rack.
func (m *MessageTarget) Key() string {
	return fmt.Sprintf("%d:%d:%d", m.element, m.ElementId(), m.Port())
}

func (m *MessageTarget) String() string {
	return fmt.Sprintf(
		"racks/%s/%s/%d/%d",
		m.rackName(), m.element.String(), m.ElementId(), m.Port())
}

// HardwareToTarget is a temporary function that converts a storage hardware
// element's address into a rack-local message target.  See note about the
// target type above for why this is temporary.
func HardwareToTarget(hw *inventory.Hardware) *MessageTarget {
	t := &MessageTarget{
		element: 0,
		Rack:    "",
		itemId:  hw.GetId(),
		port:    hw.GetPort(),
	}

	switch hw.GetType() {
	case inventory.Hardware_blade:
		t.element = targetBlade

	case inventory.Hardware_pdu:
		t.element = targetPdu

	case inventory.Hardware_tor:
		t.element = targetTor

	default:
		t.element = targetInvalid
	}

	return t
}
