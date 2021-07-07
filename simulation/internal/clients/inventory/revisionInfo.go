package inventory

import (
	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/store"
)

type revisionProvider interface {
	// GetRevision returns the revision of the details field within the object.
	// This will be either the revision of the object in the store after a
	// Create(), Read() or Update() call or be store.RevisionInvalid if the
	// details have been set or no Create(), Read() or Update() call has been
	// executed.
	//
	GetRevision() int64

	// GetRevisionRecord returns the revision of the underlying store object as
	// determined at the time of the last Create(), Read() or Update() for the
	// object. The record revision is not reset by a SetDetails() call and is
	// used when performing either a conditional update or conditional delete
	// using the object.
	//
	GetRevisionRecord() int64

	// GetRevisionStore returns the revision of the underlying store itself as
	// determined at the time of the last Create() Read() for the object. The
	// store revision is not reset by a SetDetails() call and is provided
	// for information only.
	//
	GetRevisionStore() int64

	// GetRevisionForRequest returns the appropriate revision for the update
	// for either a conditional update based upon the revision of the most
	// recently read record, or an unconditional update.
	//
	GetRevisionForRequest(unconditional bool) int64

	// resetRevision resets the revision for the details field within the object.
	// Subsequent calls to GetRevision() will return store.RevisionInvalid until
	// a successful call is made to one of the routines which invoke the store
	//
	resetRevision() int64

	// updateRevision is used to set/update the current revision information
	// as part of a successful invocation of a store routine.
	//
	updateRevisionInfo(rev int64) int64
}

type revisionInfo struct {
	revision       int64
	revisionRecord int64
	revisionStore  int64
}

// GetRevision returns the revision of the details field within the object.
// This will be either the revision of the object in the store after a
// Create(), Read() or Update() call or be store.RevisionInvalid if the
// details have been set or no Create(), Read() or Update() call has been
// executed.
//
func (r *revisionInfo) GetRevision() int64 {
	return r.revision
}

// GetRevisionRecord returns the revision of the underlying store object as
// determined at the time of the last Create(), Read() or Update() for the
// object. The record revision is not reset by a SetDetails() call and is
// used when performing either a conditional update or conditional delete
// using the object.
//
func (r *revisionInfo) GetRevisionRecord() int64 {
	return r.revisionRecord
}

// GetRevisionStore returns the revision of the underlying store itself as
// determined at the time of the last Create() Read() for the object. The
// store revision is not reset by a SetDetails() call and is provided
// for information only.
//
func (r *revisionInfo) GetRevisionStore() int64 {
	return r.revisionStore
}

// GetRevisionForRequest returns the appropriate revision for the update
// for either a conditional update based upon the revision of the most
// recently read record, or an unconditional update.
//
func (r *revisionInfo) GetRevisionForRequest(unconditional bool) int64 {

	if unconditional {
		return store.RevisionInvalid
	}

	return r.revisionRecord
}

// resetRevision resets the revision for the details field within the object.
// Subsequent calls to GetRevision() will return store.RevisionInvalid until
// a successful call is made to one of the routines which invoke the store
//
func (r *revisionInfo) resetRevision() int64 {
	r.revision = store.RevisionInvalid

	return store.RevisionInvalid
}

// updateRevision is used to set/update the current revision information
// as part of a successful invocation of a store routine.
//
func (r *revisionInfo) updateRevisionInfo(rev int64) int64 {
	r.revision = rev
	r.revisionRecord = rev
	r.revisionStore = rev

	return rev
}

