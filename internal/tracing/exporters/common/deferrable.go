package common

import (
	"context"
	"errors"

	"github.com/Jim3Things/CloudChamber/pkg/protos/log"
)

// Deferrable defines a helper type that exporters use to hold trace entries
// that arrived at an inconvenient time.
type Deferrable struct {
	maxEntries int
	deferred   []*log.Entry
}

// NewDeferrable is a function that creates a new Deferrable instance
func NewDeferrable(limit int) *Deferrable {
	return &Deferrable{
		maxEntries: limit,
		deferred:   []*log.Entry{},
	}
}

// Defer is a function that appends the specified log entry to the current
// list of waiting entries.
func (d *Deferrable) Defer(entry *log.Entry) error {
	if d.maxEntries > 0 && len(d.deferred) >= d.maxEntries {
		return errors.New("maximum deferred limit exceeded")
	}

	d.deferred = append(d.deferred, entry)
	return nil
}

// Flush is a function that attempts to post all deferred entries via the
// supplied action routine.  The operation ends when either all deferred
// entries have been processed, or when a call to the action routine returns
// an error.
//
// In the case of an error from the action routine, the entry is retained
// and processing will start with that entry on the next Flush call.
func (d *Deferrable) Flush(ctx context.Context, action func(ctx context.Context, entry *log.Entry) error) error {
	for i, item := range d.deferred {
		if err := action(ctx, item); err != nil {
			d.deferred = d.deferred[i:]
			return err
		}
	}

	d.deferred = []*log.Entry{}
	return nil
}

// GetCount returns the number of entries currently held
func (d *Deferrable) GetCount() int {
	return len(d.deferred)
}

// GetLimit returns the maximum number of entries that may be held.  Note that
// any value of zero or less is treated as 'no limit'.
func (d *Deferrable) GetLimit() int {
	return d.maxEntries
}
