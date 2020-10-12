// This module contains the defined errors, and extended error types, that are
// specific to the frontend package

package frontend

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Jim3Things/CloudChamber/internal/tracing"
)

var (
	// ErrNotInitialized is a new error to indicate initialization failures.
	ErrNotInitialized = errors.New("CloudChamber: initialization failure")

	// ErrWorkloadNotEnabled indicates the specified workload is not enabled
	// for the purposes of deployment or execution.
	ErrWorkloadNotEnabled = errors.New("CloudChamber: workload not enabled")

	// ErrUserUnableToCreate indicates the specified user account cannot be
	// created at this time
	//
	ErrUserUnableToCreate = errors.New("CloudChamber: unable to create a user account at this time")

	// ErrUserAlreadyLoggedIn indicates that the session is currently logged in
	// and a new log in cannot be processed
	ErrUserAlreadyLoggedIn = errors.New("CloudChamber: session already has a logged in user")

	// ErrUserAuthFailed indicates the supplied username and password combination
	// is not valid.
	//
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
	name  string
	value string
}

func (eubrc ErrUserBadRecordContent) Error() string {
	return fmt.Sprintf(
		"CloudChamber: discovered record for user %q where the content does not match key %q",
		eubrc.name, eubrc.value)
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
	zone string
	rack string
}

func (erae ErrRackAlreadyExists) Error() string {
	return fmt.Sprintf("CloudChamber: rack %q in zone %q already exists", erae.rack, erae.zone)
}

// ErrRackNotFound indicates the attempt to operate on a rack record failed
// as that record cannot be found.
//
type ErrRackNotFound struct {
	zone string
	rack string
}

func (ernf ErrRackNotFound) Error() string {
	return fmt.Sprintf("CloudChamber: rack %q in zone %q was not found", ernf.rack, ernf.zone)
}

// ErrPduAlreadyExists indicates the attempt to create a new pdu record
// failed as that pdu already exists.
//
type ErrPduAlreadyExists struct {
	zone string
	rack string
	pdu int64
}

func (epae ErrPduAlreadyExists) Error() string {
	return fmt.Sprintf("CloudChamber: pdu %v in zone %q, rack %q already exists", epae.pdu, epae.zone, epae.rack)
}

// ErrPduNotFound indicates the attempt to operate on a pdu record
// failed as that record cannot be found.
//
type ErrPduNotFound struct {
	zone string
	rack string
	pdu int64
}

func (epae ErrPduNotFound) Error() string {
	return fmt.Sprintf("CloudChamber: pdu %v in zone %q, rack %q was not found", epae.pdu, epae.zone, epae.rack)
}

// ErrTorAlreadyExists indicates the attempt to create a new zone record
// failed as that zone already exists.
//
type ErrTorAlreadyExists  struct {
	zone string
	rack string
	tor int64
}

func (etae ErrTorAlreadyExists) Error() string {
	return fmt.Sprintf("CloudChamber: tor %v in zone %q, rack %q already exists", etae.tor, etae.zone, etae.rack)
}

// ErrTorNotFound indicates the attempt to operate on a tor record
// failed as that record cannot be found.
//
type ErrTorNotFound struct {
	zone string
	rack string
	tor int64
}

func (etnf ErrTorNotFound) Error() string {
	return fmt.Sprintf("CloudChamber: tor %v in zone %q, rack %q was not found", etnf.tor, etnf.zone, etnf.rack)
}

// ErrBladeAlreadyExists indicates the attempt to create a new blade record
// failed as that blade already exists.
//
type ErrBladeAlreadyExists struct {
	zone string
	rack string
	blade int64
}

func (ebae ErrBladeAlreadyExists) Error() string {
	return fmt.Sprintf("CloudChamber: blade %v in zone %q, rack %q already exists", ebae.rack, ebae.zone, ebae.rack)
}

// ErrBladeNotFound indicates the attempt to operate on a blade record
// failed as that record cannot be found.
//
type ErrBladeNotFound struct {
	zone string
	rack string
	blade int64
}

func (ebnf ErrBladeNotFound) Error() string {
	return fmt.Sprintf("CloudChamber: blade %v in zone %q, rack %q was not found", ebnf.blade, ebnf.zone, ebnf.rack)
}

// HTTPError is a custom common HTTP error type that includes the status code
// to use in a response.
type HTTPError struct {
	// HTTP status code
	SC int

	// Underlying Go error
	Base error
}

// StatusCode is used to extract a status from a standard HTTPError
//
func (he *HTTPError) StatusCode() int {
	// We should not need this, but if we're called with no error at all,
	// then the status should be success...
	if he == nil {
		return http.StatusOK
	}

	return he.SC
}

func (he *HTTPError) Error() string {
	return he.Base.Error()
}

// postHTTPError sets an http error, and log it to the tracing system.
func postHTTPError(ctx context.Context, w http.ResponseWriter, err error) {
	// We're hoping this is an HTTPError form of error, which would have the
	// preferred HTTP status code included.
	//
	// If it isn't, then the error originated in some support or library logic,
	// rather than the web server's business logic.  In that case we assume a
	// status code of internal server error as the most likely correct value.
	he, ok := err.(*HTTPError)
	if !ok {
		he = &HTTPError{
			SC:   http.StatusInternalServerError,
			Base: err,
		}
	}

	_ = tracing.Error(ctx, "http error %v: %s", he.StatusCode(), he.Error())
	http.Error(w, he.Error(), he.StatusCode())
}

// httpErrorIf sets and traces an http error, if there is one.
func httpErrorIf(ctx context.Context, w http.ResponseWriter, err error) {
	if err != nil {
		postHTTPError(ctx, w, err)
	}
}

// ensurePositiveNumber is a helper function that takes a query value as a
// string, converts it, and range checks it.  If any of those operations fail
// it returns the appropriate HTTPError.
func ensurePositiveNumber(field string, value string) (int64, error) {
	res, err := strconv.ParseInt(value, 10, 64)
	if err != nil || res < 0 {
		return 0, NewErrInvalidPositiveNumber(field, value)
	}

	return res, nil
}

// ensurePositiveNumber is a helper function that takes a query value as a
// string, and converts it.  If this fails, it returns the appropriate
// HTTPError.
func ensureNumber(field string, value string) (int64, error) {
	res, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, NewErrInvalidNumber(field, value)
	}

	return res, nil

}

// +++ HTTPError specializations

// NewErrNoSessionActive indicates that the request was made without first
// establishing a logged in session
func NewErrNoSessionActive() *HTTPError {
	return &HTTPError{
		SC:   http.StatusForbidden,
		Base: errors.New("CloudChamber: no session active"),
	}
}

// NewErrNoLoginActive indicates that the specified user is not logged into this session
func NewErrNoLoginActive(name string) *HTTPError {
	return &HTTPError{
		SC:   http.StatusBadRequest,
		Base: fmt.Errorf("CloudChamber: user %q not logged into this session", name),
	}
}

// NewErrUserNotFound indicates the specified user account was determined to
// not exist (i.e. the search succeeded but no record was found)
//
func NewErrUserNotFound(name string) *HTTPError {
	return &HTTPError{
		SC:   http.StatusNotFound,
		Base: fmt.Errorf("CloudChamber: user %q not found", name),
	}
}

// NewErrUserAlreadyExists indicates the specified user account was previously
// created and the request was determined to be a duplicate Create request.
//
func NewErrUserAlreadyExists(name string) *HTTPError {
	return &HTTPError{
		SC:   http.StatusBadRequest,
		Base: fmt.Errorf("CloudChamber: user %q already exists", name),
	}
}

// NewErrUserPermissionDenied indicates the user does not have the appropriate
// permissions for the requested operation.
//
func NewErrUserPermissionDenied() *HTTPError {
	return &HTTPError{
		SC:   http.StatusForbidden,
		Base: errors.New("CloudChamber: permission denied"),
	}
}

// NewErrUserStaleVersion indicates that an operation against the specified user
// expected a different revision number than was found
//
func NewErrUserStaleVersion(name string) *HTTPError {
	return &HTTPError{
		SC:   http.StatusConflict,
		Base: fmt.Errorf("CloudChamber: user %q has a newer version than expected", name),
	}
}

// NewErrBadMatchType indicates that the If-Match value was syntactically incorrect,
// and could not be processed
func NewErrBadMatchType(match string) *HTTPError {
	return &HTTPError{
		SC:   http.StatusBadRequest,
		Base: fmt.Errorf("CloudChamber: match value %q is not recognized as a valid integer", match),
	}
}

// NewErrUserInvalidOperation indicates the operation requested for the supplied
// user account is invalid in some way, likely a non-existent operation code.
//
func NewErrUserInvalidOperation(op string, user string) *HTTPError {
	return &HTTPError{
		SC:   http.StatusBadRequest,
		Base: fmt.Errorf("CloudChamber: invalid user operation requested (?op=%s) for user %q", op, user),
	}
}

// NewErrUserProtected indicates that the user entry may not be deleted.
func NewErrUserProtected(name string) *HTTPError {
	return &HTTPError{
		SC:   http.StatusForbidden,
		Base: fmt.Errorf("CloudChamber: user %q is protected and cannot be deleted", name),
	}
}

// NewErrRackNotFound indicates the specified rack do not exist and the http
// request (http.statusNotFound) determines to be the request was made against
// a non-existing rack.
func NewErrRackNotFound(name string) *HTTPError {
	return &HTTPError{
		SC:   http.StatusNotFound,
		Base: fmt.Errorf("CloudChamber: rack %q not found", name),
	}
}

// NewErrBladeNotFound indicates that the rack was found but no blade was found
//
func NewErrBladeNotFound(rackid string, bladeid int64) *HTTPError {
	return &HTTPError{
		SC:   http.StatusNotFound,
		Base: fmt.Errorf("CloudChamber: blade %d not found in rack %q", bladeid, rackid),
	}
}

// NewErrInvalidStepperMode indicates that an unrecognized simulated policy mode
// was requested.
func NewErrInvalidStepperMode(mode string) *HTTPError {
	return &HTTPError{
		SC:   http.StatusBadRequest,
		Base: fmt.Errorf("CloudChamber: mode %q is invalid.  Supported modes are 'manual' and 'automatic'", mode),
	}
}

// NewErrInvalidRateRequest indicates that the automatic ticks-per-second rate was
// present, but the selected policy mode was not 'automatic'.
func NewErrInvalidRateRequest() *HTTPError {
	return &HTTPError{
		SC:   http.StatusBadRequest,
		Base: errors.New("CloudChamber: manual mode does not accept additional arguments"),
	}
}

// NewErrInvalidStepperRate indicates that the supplied ticks-per-second rate was
// not recognized as a valid number.
func NewErrInvalidStepperRate(rate string) *HTTPError {
	return &HTTPError{
		SC:   http.StatusBadRequest,
		Base: fmt.Errorf("CloudChamber: requested rate %q could not be parsed as a positive decimal number", rate),
	}
}

// NewErrStepperFailedToSetPolicy indicates that an error occurred while setting
// the new policy.  This most likely is due to an ETag mismatch.
func NewErrStepperFailedToSetPolicy() *HTTPError {
	return &HTTPError{
		SC:   http.StatusBadRequest,
		Base: errors.New("CloudChamber: Set simulated time policy operation failed"),
	}
}

// NewErrInvalidNumber indicates that the specified value could not be
// processed as a number.
func NewErrInvalidNumber(field string, value string) *HTTPError {
	return &HTTPError{
		SC: http.StatusBadRequest,
		Base: fmt.Errorf(
			"CloudChamber: the %q field's value %q could not be parsed as a decimal number",
			field, value),
	}
}

// NewErrInvalidPositiveNumber indicates that the specified value could not be
// processed as a positive number.
func NewErrInvalidPositiveNumber(field string, value string) *HTTPError {
	return &HTTPError{
		SC: http.StatusBadRequest,
		Base: fmt.Errorf(
			"CloudChamber: the %q field's value %q could not be parsed as a positive decimal number",
			field, value),
	}
}

// --- HTTPError specializations
