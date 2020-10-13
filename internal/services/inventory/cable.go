package inventory

import (
	"errors"

	"github.com/Jim3Things/CloudChamber/internal/common"
)

var (
	errStuck = errors.New("cable is faulted")
	errTooLate = errors.New("cable modified after the requested time")
)

// cable describes the active state of a connecting cable from the one element
// to a connected blade.  The connecting cable may be a power cable from the
// PDU, a network cable from the TOR, or some other control cable.
type cable struct {
	common.Guarded

	// on is true if the cable is actually connected and 'functioning' (i.e.
	// providing power, if a PDU power cable)
	on bool

	// faulted is true if the cable connection is broken in such a way that the
	// on state cannot be changed, and cannot be reliably determined for
	// control.
	faulted bool
}

// newCable creates a new cable instance with the specified state and guard
// values.
func newCable(on bool, faulted bool, at int64) *cable {
	return &cable{
		Guarded: common.Guarded{
			Guard: at,
		},
		on: on,
		faulted: faulted,
	}
}

// set changes the on/off state for the cable, contingent upon the guard being
// met and the cable not being faulted.  It returns whether the cable state
// changed, and an error if the conditions for changing were not met.
func (c *cable) set(offOn bool, guard int64, at int64) (bool, error) {
	if c.faulted {
		return false, errStuck
	}

	if !c.Pass(guard, at) {
		return false, errTooLate
	}

	startState := c.on
	c.on = offOn

	return c.on != startState, nil
}

// fault sets the cable to faulted, as well as the actual underlying state.  It
// also requires that the guard check be met.
func (c *cable) fault(offOn bool, guard int64, at int64) error {
	if !c.Pass(guard, at) {
		return errTooLate
	}

	c.faulted = true
	c.on = offOn

	return nil
}

// fix is the inverse of fault, in that it clears the faulted state.
func (c *cable) fix(guard int64, at int64) error {
	if !c.Pass(guard, at) {
		return errTooLate
	}

	c.faulted = false

	return nil
}

// force sets the underlying state of the cable, even if it is faulted.
func (c *cable) force(offOn bool, guard int64, at int64) (bool, error) {
	if !c.Pass(guard, at) {
		return false, errTooLate
	}

	startState := c.on

	c.on = offOn
	return c.on != startState, nil
}
