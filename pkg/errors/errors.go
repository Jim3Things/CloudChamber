package errors

import (
	"errors"
	"fmt"
)

var (
	// ErrNotInitialized is a new error to indicate initialization failures.
	ErrNotInitialized = errors.New("CloudChamber: initialization failure")

	// ErrWorkloadNotEnabled indicates the specified workload is not enabled
	// for the purposes of deployment or execution.
	ErrWorkloadNotEnabled = errors.New("CloudChamber: workload not enabled")

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
// currently active client. The Connect() method can be used to establish a client.
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

// ErrStoreConnectionFailed indicates that the attempt to establish a connection
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
	return fmt.Sprintf("CloudChamber: method %s not currently implemented", string(esni))
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
		"CloudChamber: unexpected record count for key %q - expected: %d actual %d",
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
		"CloudChamber: invalid revision argument supplied on update for key %q - requested: %d actual: %d",
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
		"CloudChamber: condition failure on update for key %q - requested: %d condition: %s actual: %d",
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
	return fmt.Sprintf("CloudChamber: invalid store configuration - %s", string(esic))
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

var (
	// ErrUserUnableToCreate indicates the specified user account cannot be
	// created at this time
	ErrUserUnableToCreate = errors.New("CloudChamber: unable to create a user account at this time")

	// ErrUserAlreadyLoggedIn indicates that the session is currently logged in
	// and a new log in cannot be processed
	ErrUserAlreadyLoggedIn = errors.New("CloudChamber: session already has a logged in user")

	// ErrUserAuthFailed indicates the supplied username and password combination
	// is not valid.
	ErrUserAuthFailed = errors.New("CloudChamber: authentication failed, invalid user name or password")
)

// ErrUserAlreadyExists indicates the attempt to create a new user account
// failed as that user already exists.
//
type ErrUserAlreadyExists string

func (euae ErrUserAlreadyExists) Error() string {
	return fmt.Sprintf("CloudChamber: user %q already exists", string(euae))
}

// ErrUserNotFound indicates the attempt to locate a user account failed as that
// user does not exist.
//
type ErrUserNotFound string

func (eunf ErrUserNotFound) Error() string {
	return fmt.Sprintf("CloudChamber: user %q not found", string(eunf))
}

// ErrUserStaleVersion indicates the attempt to locate a user account failed as that
// user does not exist.
//
type ErrUserStaleVersion string

func (eusv ErrUserStaleVersion) Error() string {
	return fmt.Sprintf("CloudChamber: user %q has a newer version than expected", string(eusv))
}

// ErrUserProtected indicates the attempt to locate a user account failed as that
// user does not exist.
//
type ErrUserProtected string

func (eup ErrUserProtected) Error() string {
	return fmt.Sprintf("CloudChamber: user %q is protected and cannot be deleted", string(eup))
}

// ErrUserBadRecordContent indicates the user record retrieved from the store store
// has some content that does not match the key. An example might be that the user
// name used for a key does not match the user name field in the record.
//
type ErrUserBadRecordContent struct {
	Name  string
	Value string
}

func (eubrc ErrUserBadRecordContent) Error() string {
	return fmt.Sprintf(
		"CloudChamber: discovered record for user %q where the content does not match key %q",
		eubrc.Name, eubrc.Value)
}

// ErrUnableToVerifySystemAccount indicates that the system account has changed
// from what is defined in the configuration
type ErrUnableToVerifySystemAccount struct {
	Name string
	Err  error
}

func (eutvsa ErrUnableToVerifySystemAccount) Error() string {
	return fmt.Sprintf(
		"CloudChamber: unable to verify the standard %q account is using configured password - error %v",
		eutvsa.Name, eutvsa.Err)
}

// ErrZoneAlreadyExists indicates the attempt to create a new zone record
// failed as that zone already exists.
//
type ErrZoneAlreadyExists string

func (ezae ErrZoneAlreadyExists) Error() string {
	return fmt.Sprintf("CloudChamber: zone %q already exists", string(ezae))
}

// ErrZoneNotFound indicates the attempt to locate a zone record failed as that
// zone does not exist.
//
type ErrZoneNotFound string

func (eznf ErrZoneNotFound) Error() string {
	return fmt.Sprintf("CloudChamber: zone %q not found", string(eznf))
}

// ErrZoneStaleVersion indicates the attempt to locate a specific version of a
// zone record failed as either that zone does not exist, or the specific
// version is no longer present in the store.
//
type ErrZoneStaleVersion string

func (ezsv ErrZoneStaleVersion) Error() string {
	return fmt.Sprintf("CloudChamber: zone %q has a newer version than expected", string(ezsv))
}

// ErrRackAlreadyExists indicates the attempt to create a new rack record
// failed as that rack already exists.
//
type ErrRackAlreadyExists struct {
	Zone string
	Rack string
}

func (erae ErrRackAlreadyExists) Error() string {
	return fmt.Sprintf("CloudChamber: rack %q in zone %q already exists", erae.Rack, erae.Zone)
}

// ErrRackNotFound indicates the attempt to operate on a rack record failed
// as that record cannot be found.
//
type ErrRackNotFound struct {
	Zone string
	Rack string
}

func (ernf ErrRackNotFound) Error() string {
	return fmt.Sprintf("CloudChamber: rack %q in zone %q was not found", ernf.Rack, ernf.Zone)
}

// ErrPduAlreadyExists indicates the attempt to create a new pdu record
// failed as that pdu already exists.
//
type ErrPduAlreadyExists struct {
	Zone string
	Rack string
	Pdu  int64
}

func (epae ErrPduAlreadyExists) Error() string {
	return fmt.Sprintf(
		"CloudChamber: %s already exists",
		pduAddress(epae.Zone, epae.Rack, epae.Pdu))
}

// ErrPduNotFound indicates the attempt to operate on a pdu record
// failed as that record cannot be found.
//
type ErrPduNotFound struct {
	Zone string
	Rack string
	Pdu  int64
}

func (epae ErrPduNotFound) Error() string {
	return fmt.Sprintf(
		"CloudChamber: %s was not found",
		pduAddress(epae.Zone, epae.Rack, epae.Pdu))
}

// ErrTorAlreadyExists indicates the attempt to create a new zone record
// failed as that zone already exists.
//
type ErrTorAlreadyExists struct {
	Zone string
	Rack string
	Tor  int64
}

func (etae ErrTorAlreadyExists) Error() string {
	return fmt.Sprintf(
		"CloudChamber: %s already exists",
		torAddress(etae.Zone, etae.Rack, etae.Tor))
}

// ErrTorNotFound indicates the attempt to operate on a tor record
// failed as that record cannot be found.
//
type ErrTorNotFound struct {
	Zone string
	Rack string
	Tor  int64
}

func (etnf ErrTorNotFound) Error() string {
	return fmt.Sprintf(
		"CloudChamber: %s was not found",
		torAddress(etnf.Zone, etnf.Rack, etnf.Tor))
}

// ErrBladeAlreadyExists indicates the attempt to create a new blade record
// failed as that blade already exists.
//
type ErrBladeAlreadyExists struct {
	Zone  string
	Rack  string
	Blade int64
}

func (ebae ErrBladeAlreadyExists) Error() string {
	return fmt.Sprintf(
		"CloudChamber: %s already exists",
		bladeAddress(ebae.Zone, ebae.Rack, ebae.Blade))
}

// ErrBladeNotFound indicates the attempt to operate on a blade record
// failed as that record cannot be found.
//
type ErrBladeNotFound struct {
	Zone  string
	Rack  string
	Blade int64
}

func (ebnf ErrBladeNotFound) Error() string {
	return fmt.Sprintf(
		"CloudChamber: %s was not found",
		bladeAddress(ebnf.Zone, ebnf.Rack, ebnf.Blade))
}
