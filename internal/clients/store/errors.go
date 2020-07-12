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

// ErrStoreKeyReadFailure indicates the read transaction failed.
//
type ErrStoreKeyReadFailure string

func (eskff ErrStoreKeyReadFailure) Error() string {
	return fmt.Sprintf("CloudChamber: transaction failed reading key %q", string(eskff))
}

// ErrStoreKeyWriteFailure indicates the read transaction failed.
//
type ErrStoreKeyWriteFailure string

func (eskwf ErrStoreKeyWriteFailure) Error() string {
	return fmt.Sprintf("CloudChamber: transaction failed deleting key %q", string(eskwf))
}

// ErrStoreKeyDeleteFailure indicates the read transaction failed.
//
type ErrStoreKeyDeleteFailure string

func (eskdf ErrStoreKeyDeleteFailure) Error() string {
	return fmt.Sprintf("CloudChamber: transaction failed deleting key %q", string(eskdf))
}

// ErrStoreBadRecordKey indicates the store has found a record with an unrecognized
// format. This generally means the key itself is not properly constructed.
//
type ErrStoreBadRecordKey string

func (esbrk ErrStoreBadRecordKey) Error() string {
	return fmt.Sprintf("CloudChamber: discovered key with an unrecognized format %q", string(esbrk))
}

// ErrStoreBadRecordContent indicates the store has found a record with some content
// that does not match the key. An example might be that the user name used for a key
// does not match the user name field in the record.
//
// There is little consistency checking of this nature in the store itself due to the
// limited knowledge the store component has about the content of records. There
// should be no expectation that the store is taking on the responsibility of any
// consistency checking and any that does occur should be treated as advisory.
//
type ErrStoreBadRecordContent string

func (esbrk ErrStoreBadRecordContent) Error() string {
	return fmt.Sprintf("CloudChamber: discovered found record where content does not match key %q", string(esbrk))
}

// ErrStoreBadRecordCount indicates the record count for the operation was not valid.
// This might mean that the store found more, or less, than the number of records expected.
//
type ErrStoreBadRecordCount struct {
	key      string
	expected int
	actual   int
}

func (esbrc ErrStoreBadRecordCount) Error() string {
	return fmt.Sprintf("CloudChamber: unexpected record count for key %q - expected: %v actual %v", esbrc.key, esbrc.expected, esbrc.actual)
}

// ErrStoreBadArgCondition indicates the condition argument specified for the update was not valid.
//
type ErrStoreBadArgCondition struct {
	key       string
	condition Condition
}

func (esbac ErrStoreBadArgCondition) Error() string {
	return fmt.Sprintf("CloudChamber: compare operator %q not valid for key %q", esbac.condition, esbac.key)
}

// ErrStoreBadArgRevision indicates the supplied revision argument was invalid.
//
type ErrStoreBadArgRevision struct {
	key       string
	requested int64
	actual    int64
}

func (esbar ErrStoreBadArgRevision) Error() string {
	return fmt.Sprintf("CloudChamber: invalid revision argument supplied on update for key %q - requested: %v actual: %v", esbar.key, esbar.requested, esbar.actual)
}

// ErrStoreConditionFail indicates the update transaction failed due to a
// failure in the requested condition. This is like a revision mismatch
// but other conditions may apply.
//
type ErrStoreConditionFail struct {
	key       string
	requested int64
	condition Condition
	actual    int64
}

func (esucf ErrStoreConditionFail) Error() string {
	return fmt.Sprintf("CloudChamber: condition failure on update for key %q - requested: %v condition: %v actual: %v", esucf.key, esucf.requested, esucf.condition, esucf.actual)
}
