package messages

// This file contains the implementation for the component in the repair
// messages that designates the target simulated inventory element for the
// repair.

import (
	"fmt"
)

const (
	targetTor int = iota
	targetPdu
	targetBlade
)

// MessageTarget describes the target element for the enclosing message.
type MessageTarget struct {
	element int
	Rack    string
	blade   int64
}

// NewTargetTor creates a new target instance that specifies the designated
// rack's TOR.
func NewTargetTor(name string) *MessageTarget {
	return &MessageTarget{
		element: targetTor,
		Rack:    name,
		blade:   0,
	}
}

// NewTargetPdu creates a new target instance that specifies the designated
// rack's PDU.
func NewTargetPdu(name string) *MessageTarget {
	return &MessageTarget{
		element: targetPdu,
		Rack:    name,
		blade:   0,
	}
}

// NewTargetBlade creates a new target instance that specifies the designated
// blade in the specified rack.
func NewTargetBlade(name string, id int64) *MessageTarget {
	return &MessageTarget{
		element: targetBlade,
		Rack:    name,
		blade:   id,
	}
}

func (m *MessageTarget) rackName() string { return m.Rack }
func (m *MessageTarget) IsTor() bool      { return m.element == targetTor }
func (m *MessageTarget) IsPdu() bool      { return m.element == targetPdu }
func (m *MessageTarget) BladeID() (int64, bool) {
	if m.element == targetBlade {
		return m.blade, true
	}

	return -1, false
}

// Describe produces a string that describes the target element's logical
// address.
func (m *MessageTarget) Describe() string {
	preamble := ""

	if m.IsTor() {
		preamble = "the TOR"
	} else if m.IsPdu() {
		preamble = "the PDU"
	} else if id, ok := m.BladeID(); ok {
		preamble = fmt.Sprintf("blade %d", id)
	}

	return fmt.Sprintf("%s in rack %q", preamble, m.rackName())
}

func (m *MessageTarget) String() string {
	component := "unknown"
	switch m.element {
	case targetBlade:
		component = "blade"
		break

	case targetTor:
		component = "tor"
		break

	case targetPdu:
		component = "pdu"
		break

	default:
		break
	}

	return fmt.Sprintf("racks/%s/%s/%d", m.Rack, component, m.blade)
}
