package inventory

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

// messageTarget describes the target element for the enclosing message.
type messageTarget struct {
	element int
	rack string
	blade int64
}

// newTargetTor creates a new target instance that specifies the designated
// rack's TOR.
func newTargetTor(name string) *messageTarget {
	return &messageTarget{
		element: targetTor,
		rack:  name,
		blade: 0,
	}
}

// newTargetPdu creates a new target instance that specifies the designated
// rack's PDU.
func newTargetPdu(name string) *messageTarget {
	return &messageTarget{
		element: targetPdu,
		rack:  name,
		blade: 0,
	}
}

// newTargetBlade creates a new target instance that specifies the designated
// blade in the specified rack.
func newTargetBlade(name string, id int64) *messageTarget {
	return &messageTarget{
		element: targetBlade,
		rack:  name,
		blade: id,
	}
}

func (m *messageTarget) rackName() string { return m.rack }
func (m *messageTarget) isTor() bool { return m.element == targetTor }
func (m *messageTarget) isPdu() bool { return m.element == targetPdu }
func (m *messageTarget) bladeID() (int64, bool) {
	if m.element == targetBlade {
		return m.blade, true
	}

	return -1, false
}

// describe produces a string that describes the target element's logical
// address.
func (m *messageTarget) describe() string {
	preamble := ""

	if m.isTor() {
		preamble = "the TOR"
	} else if m.isPdu() {
		preamble = "the PDU"
	} else if id, ok := m.bladeID(); ok {
		preamble = fmt.Sprintf("blade %d", id)
	}

	return fmt.Sprintf("%s in rack %q", preamble, m.rackName())
}
