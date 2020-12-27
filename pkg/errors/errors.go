package errors

import (
	"errors"
	"fmt"
)

var (
	// ErrStoreUnableToCreateClient indicates that it is not currently possible
	// to create a client.
	//
	ErrStoreUnableToCreateClient = errors.New("CloudChamber: unable to create a new client")
)

// ErrClientAlreadyInitialized indicates that the attempt to initialize the
// specified internal client library used to mediate access to a microservice
// has failed because it was already initialized.
type ErrClientAlreadyInitialized string

func (eai ErrClientAlreadyInitialized) Error() string {
	return fmt.Sprintf("the %s client has already been initialized", string(eai))
}

// ErrClientNotReady indicates that an attempt to call an internal microservice
// through a client library has failed, as that library has not yet been
// initialized.
type ErrClientNotReady string

func (enr ErrClientNotReady) Error() string {
	return fmt.Sprintf("the %s client is not ready to process requests", string(enr))
}

// ErrTimerNotFound indicates that the specified timer ID was not found when
// attempting to look up an active timer.
type ErrTimerNotFound int

func (etnf ErrTimerNotFound) Error() string {
	return fmt.Sprintf("time ID %d was not found", int(etnf))
}

// ErrStoreNotConnected indicates the store instance does not have a
// currently active client. The Connect() method can be used to establist a client.
//
type ErrStoreNotConnected string

func (esnc ErrStoreNotConnected) Error() string {
	return fmt.Sprintf("CloudChamber: client not currently connected - %s", string(esnc))
}

// ErrStoreConnected indicates the request failed as the store is currently
// connected and the request is not possible in that condition.
//
type ErrStoreConnected string

func (esc ErrStoreConnected) Error() string {
	return fmt.Sprintf("CloudChamber: client currently connected - %s", string(esc))
}

// ErrStoreConnectionFailed indicates that the attempt to extablish a connection
// from this client to the store failed
//
type ErrStoreConnectionFailed struct {
	Endpoints []string
	Reason    error
}

func (escf ErrStoreConnectionFailed) Error() string {
	return fmt.Sprintf(
		"CloudChamber: failed to establish connection to store - Endpoints: %q Reason: %q",
		escf.Endpoints, escf.Reason)
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

// ErrStoreKeyReadFailure indicates the read transaction failed.
//
type ErrStoreKeyReadFailure string

func (eskff ErrStoreKeyReadFailure) Error() string {
	return fmt.Sprintf("CloudChamber: transaction failed reading key %q", string(eskff))
}

// ErrStoreKeyWriteFailure indicates the write transaction failed as a result of a pre-condition failure.
//
type ErrStoreKeyWriteFailure string

func (eskwf ErrStoreKeyWriteFailure) Error() string {
	return fmt.Sprintf("CloudChamber: transaction failed writing key %q", string(eskwf))
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
	return fmt.Sprintf("CloudChamber: discovered record where content does not match key %q", string(esbrk))
}

// ErrStoreBadRecordCount indicates the record count for the operation was not valid.
// This might mean that the store found more, or less, than the number of records expected.
//
type ErrStoreBadRecordCount struct {
	Key      string
	Expected int
	Actual   int
}

func (esbrc ErrStoreBadRecordCount) Error() string {
	return fmt.Sprintf(
		"CloudChamber: unexpected record count for key %q - expected: %v actual %v",
		esbrc.Key, esbrc.Expected, esbrc.Actual)
}

// ErrStoreBadArgCondition indicates the condition argument specified for the update was not valid.
//
type ErrStoreBadArgCondition struct {
	Key       string
	Condition string
}

func (esbac ErrStoreBadArgCondition) Error() string {
	return fmt.Sprintf("CloudChamber: compare operator %q not valid for key %q", esbac.Condition, esbac.Key)
}

// ErrStoreBadArgRevision indicates the supplied revision argument was invalid.
//
type ErrStoreBadArgRevision struct {
	Key       string
	Requested int64
	Actual    int64
}

func (esbar ErrStoreBadArgRevision) Error() string {
	return fmt.Sprintf(
		"CloudChamber: invalid revision argument supplied on update for key %q - requested: %v actual: %v",
		esbar.Key, esbar.Requested, esbar.Actual)
}

// ErrStoreConditionFail indicates the update transaction failed due to a
// failure in the requested condition. This is like a revision mismatch
// but other conditions may apply.
//
type ErrStoreConditionFail struct {
	Key       string
	Requested int64
	Condition string
	Actual    int64
}

func (esucf ErrStoreConditionFail) Error() string {
	return fmt.Sprintf(
		"CloudChamber: condition failure on update for key %q - requested: %v condition: %v actual: %v",
		esucf.Key, esucf.Requested, esucf.Condition, esucf.Actual)
}

// ErrStoreAlreadyExists indicates the key, value pair being created already exists
//
type ErrStoreAlreadyExists string

func (esae ErrStoreAlreadyExists) Error() string {
	return fmt.Sprintf("CloudChamber: condition failure (already exists) on create for key %q", string(esae))
}

// ErrStoreInvalidConfiguration indicates the key, value pair being created already exists
//
type ErrStoreInvalidConfiguration string

func (esic ErrStoreInvalidConfiguration) Error() string {
	return fmt.Sprintf("CloudChamber: invalid store configuration - %v", string(esic))
}

// ErrDuplicateRack indicates duplicates rack names found
type ErrDuplicateRack string

func (edr ErrDuplicateRack) Error() string {
	return fmt.Sprintf("Duplicate rack %q detected", string(edr))
}

// ErrDuplicateBlade indicates duplicates blade indexes found
type ErrDuplicateBlade struct {
	Rack  string
	Blade int64
}

func (edb ErrDuplicateBlade) Error() string {
	return fmt.Sprintf("Duplicate Blade %d in Rack %q detected", edb.Blade, edb.Rack)
}

// ErrRackValidationFailure indicates validation failure in the attributes associated
// with a rack.
type ErrRackValidationFailure struct {
	Rack string
	Err  error
}

func (evf ErrRackValidationFailure) Error() string {
	return fmt.Sprintf("In rack %q: %v", evf.Rack, evf.Err)
}

