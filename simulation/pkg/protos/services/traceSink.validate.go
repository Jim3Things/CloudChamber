package services

import (
	"github.com/Jim3Things/CloudChamber/simulation/pkg/errors"
)

func (x *AppendRequest) Validate() error {
	return x.Entry.Validate("Entry.")
}

func (x *GetAfterRequest) Validate() error {
	if x.Id < -1 {
		return errors.ErrMustBeGTE{
			Field:    "Id",
			Actual:   x.Id,
			Required: -1,
		}
	}

	// A GetAfter call must be able to make progress, in spite of the call
	// itself producing trace entries.  Verify that enough entries can be
	// returned to ensure that at least one entry written prior to the
	// GetAfter call is returned as well.

	if x.MaxEntries < 10 {
		return errors.ErrMustBeGTE{
			Field:    "MaxEntries",
			Actual:   x.MaxEntries,
			Required: 10,
		}
	}

	return nil
}
