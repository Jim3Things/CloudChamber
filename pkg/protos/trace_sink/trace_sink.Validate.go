package trace_sink

import (
    "github.com/Jim3Things/CloudChamber/pkg/protos/common"
)

func (x *AppendRequest) Validate() error {
    return x.Entry.Validate("Entry.")
}

func (x *GetAfterRequest) Validate() error {
    if x.Id < 0 {
        return common.ErrMustBeGTE{
            Field:    "Id",
            Actual:   x.Id,
            Required: 0,
        }
    }

    // A GetAfter call must be able to make progress, in spite of the call
    // itself producing trace entries.  Verify that enough entries can be
    // returned to ensure that at least one entry written prior to the
    // GetAfter call is returned as well.

    if x.MaxEntries < 10 {
        return common.ErrMustBeGTE{
            Field:    "MaxEntries",
            Actual:   x.MaxEntries,
            Required: 10,
        }
    }

    return nil
}
