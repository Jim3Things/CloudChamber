// Package store - Errors contains the store package specific errors.
//
package store

import (
	"errors"
	"fmt"
)

var (

	// ErrStoreUnableToCreateClient indicates that it is not currently possible
	// to create a client.
	//
	ErrStoreUnableToCreateClient = errors.New("CloudChamber: unable to create a new client")

	// ErrStoreNotConnected indicates the store instance does not have a
	// currently active client. The Connect() method can be used to establist a client.
	//
	ErrStoreNotConnected = errors.New("CloudChamber: client not currently connected")

	// ErrStoreConnected indicates the request failed as the store is currently
	// connected and the request is not possible in that condition.
	//
	ErrStoreConnected = errors.New("CloudChamber: client currently connected")
)

// ErrStoreBadResultSize indicates the size of the result set does not match
// expectations. There may be either too many, or too few. Typically a single
// result way anticipated and more that that was received.
//
type ErrStoreBadResultSize struct {
	expected int
	actual   int
}

func (esbrs ErrStoreBadResultSize) Error() string {
	return fmt.Sprintf("CloudChamber: unexpected size for result set - got %v expected %v", esbrs.actual, esbrs.expected)
}

// ErrStoreKeyNotFound indicates the request key was not found when the store
// lookup/fetch was attempted.
//
type ErrStoreKeyNotFound string

func (esknf ErrStoreKeyNotFound) Error() string {
	return fmt.Sprintf("CloudChamber: key %q not found", string(esknf))
}

// ErrStoreKeyTypeMismatch indicates the request key was not found when the store
// lookup/fetch was attempted.
//
type ErrStoreKeyTypeMismatch string

func (esktm ErrStoreKeyTypeMismatch) Error() string {
	return fmt.Sprintf("CloudChamber: key %q not the requested type", string(esktm))
}

// ErrStoreNotImplemented indicated the called method does not yet have an
// implementation
//
type ErrStoreNotImplemented string

func (esni ErrStoreNotImplemented) Error() string {
	return fmt.Sprintf("CloudChamber: method %v not currently implemented", string(esni))
}

// ErrStoreKeyFetchFailure indicates the read transaction failed.
//
type ErrStoreKeyFetchFailure string

func (eskff ErrStoreKeyFetchFailure) Error() string {
	return fmt.Sprintf("CloudChamber: fetch txn failed reading key %q", string(eskff))
}

// ErrStoreKeyDeleteFailure indicates the read transaction failed.
//
type ErrStoreKeyDeleteFailure string

func (eskdf ErrStoreKeyDeleteFailure) Error() string {
	return fmt.Sprintf("CloudChamber: fetch txn failed deleting key %q", string(eskdf))
}

// ErrStoreWriteConditionFail indicates the update transaction failed due to a revision mismatch.
//
type ErrStoreWriteConditionFail string

func (eswcf ErrStoreWriteConditionFail) Error() string {
	return fmt.Sprintf("CloudChamber: condition fail/mismatch on update for key %q", string(eswcf))
}

// ErrStoreBadArgRevision indicates the supplied revision argument was invalid.
//
type ErrStoreBadArgRevision string

func (esbar ErrStoreBadArgRevision) Error() string {
	return fmt.Sprintf("CloudChamber: invalid revision argument supplied on update for key %q", string(esbar))
}

// ErrStoreBadArgCompare indicates the compare argument for the update was not valid.
//
type ErrStoreBadArgCompare string

func (esbac ErrStoreBadArgCompare) Error() string {
	return fmt.Sprintf("CloudChamber: compare operator not valid for key %q", string(esbac))
}

// ErrStoreBadRecordCount indicates the record count for the update was not valid.
//
type ErrStoreBadRecordCount string

func (esbrc ErrStoreBadRecordCount) Error() string {
	return fmt.Sprintf("CloudChamber: did not get the number of records expected %q", string(esbrc))
}

// ErrStoreBadRecordKey indicates the store has found a record with an unrecognized format
//
type ErrStoreBadRecordKey string

func (esbrk ErrStoreBadRecordKey) Error() string {
	return fmt.Sprintf("CloudChamber: discovered key with an unrecognized format %q", string(esbrk))
}

// ErrStoreBadRecordContent indicates the store has found a record with some content that does not match the key
//
type ErrStoreBadRecordContent string

func (esbrk ErrStoreBadRecordContent) Error() string {
	return fmt.Sprintf("CloudChamber: discovered found record where content does not match key %q", string(esbrk))
}
