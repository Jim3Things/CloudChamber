// This module contain the structures and methods to operate on the persisted definition
// table within the inventory package.

package inventory

import (
	"context"

	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/namespace"
	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/store"
)

// Watch is a structure returned from the Watch() function and provides
// the channel used to report changes within the namespace covered by the
// watchpoint, and also is used to terminate the watchpoint by means of
// the Close() method.
//
type Watch struct {
	watch *store.Watch

	// Events is the channel on which the notifications are delivered. The
	// caller should pull the WatchEvent structures from this channel to
	// receive the event notifications.
	//
	Events chan WatchEvent
}

// WatchEvent is a structure used to describe a change to the portion of
// a namespace that is being monitored by a watchpoint established via a
// call to the Watch() function.
//
type WatchEvent struct {
	// Err indicates if some sort of error occurred during the construction
	// of the WatchEvent notification itself, likely an issue resulting from
	// processing the key responsible for generating the event. The value
	// of the key leading to the problem is included in the error.
	//
	Err error

	// Type indicates the type of change to the store that lead to the
	// event such as a create, a modify/update or a deletion of the
	// indicated key
	//
	Type store.WatchEventType

	// Address is name of the object that was changed.
	//
	Address *namespace.Address

	// Revision is the revision of the store itself when the change occurred.
	// For creates and updates, this will be the same as the new revision
	// of the item that was the subject of the create/update.
	//
	Revision int64

	// NewRev is the revision value for the item that was modified for
	// create and update changes. For delete operation, this will be set
	// to store.RevisionInvalid
	//
	NewRev int64

	// NewVal is the value associated with the key after the completion
	// of the operation. That is, the value after the create or update
	// operation. For a delete operation, this will be set to the empty
	// string ""
	//
	NewVal string

	// OldRev is the revision of the key, value pair prior to the change
	// that lead to the notification. For a create operation, this is set
	// to store.RevisionInvalid as there was no previous key, value pair.
	//
	OldRev int64

	// OldVal is the value associated with the key prior to the change
	// that lead to the notification. For a create operation, this is
	// set to the empty string "" as there was no previous key, value
	// pair.
	//
	OldVal string
}

// Close is a method used to close the upstream source of the notification
// channel and should be called once the watchpoint is no longer required.
//
func (w *Watch) Close(ctx context.Context) error {
	return w.watch.Close(ctx)
}
