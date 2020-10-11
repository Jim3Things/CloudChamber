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


func newCable(on bool, faulted bool, at int64) *cable {
	return &cable{
		Guarded: common.Guarded{
			Guard: at,
		},
		on: on,
		faulted: faulted,
	}
}

// set
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

func (c *cable) fault(offOn bool, guard int64, at int64) error {
	if !c.Pass(guard, at) {
		return errTooLate
	}

	c.faulted = true
	c.on = offOn

	return nil
}

func (c *cable) fix(guard int64, at int64) error {
	if !c.Pass(guard, at) {
		return errTooLate
	}

	c.faulted = false

	return nil
}

func (c *cable) force(offOn bool, guard int64, at int64) (bool, error) {
	if !c.Pass(guard, at) {
		return false, errTooLate
	}

	startState := c.on

	c.on = offOn
	return c.on != startState, nil
}
