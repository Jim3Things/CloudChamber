package common

import (
    "context"
    "errors"

    "github.com/Jim3Things/CloudChamber/pkg/protos/log"
)

type Deferrable struct {
    maxEntries int
    deferred []*log.Entry
}

func NewDeferrable(limit int) *Deferrable {
    return &Deferrable{
        maxEntries: limit,
        deferred:   []*log.Entry{},
    }
}

func (d *Deferrable) Defer(entry *log.Entry) error {
    if d.maxEntries > 0 && len(d.deferred) >= d.maxEntries {
        return errors.New("maximum deferred limit exceeded")
    }

    d.deferred = append(d.deferred, entry)
    return nil
}

func (d *Deferrable) Flush(ctx context.Context, action func(ctx context.Context, entry *log.Entry) error) error {
    return nil
}
