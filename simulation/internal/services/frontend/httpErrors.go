// This module contains the defined errors, and extended error types, that are
// specific to the frontend package

package frontend

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Jim3Things/CloudChamber/simulation/internal/tracing"
	err2 "github.com/Jim3Things/CloudChamber/simulation/pkg/errors"
)

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

// NewErrSessionNotFound creates an HTTP error indicating that the requested
// session ID was not found amongst the active sessions.
func NewErrSessionNotFound(id int64) *HTTPError {
	return &HTTPError{
		SC:   http.StatusNotFound,
		Base: err2.ErrSessionNotFound(id),
	}
}

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

// NewErrRegionNotFound indicates the specified region does not exist and the http
// request (http.statusNotFound) determines to be the request was made against
// a non-existing region.
func NewErrRegionNotFound(name string) *HTTPError {
	return &HTTPError{
		SC:   http.StatusNotFound,
		Base: fmt.Errorf("CloudChamber: region %q not found", name),
	}
}

// NewErrZoneNotFound indicates the specified zone does not exist and the http
// request (http.statusNotFound) determines to be the request was made against
// a non-existing zone.
func NewErrZoneNotFound(name string) *HTTPError {
	return &HTTPError{
		SC:   http.StatusNotFound,
		Base: fmt.Errorf("CloudChamber: zone %q not found", name),
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

// NewErrPduNotFound indicates that the rack was found but no pdu was found
//
func NewErrPduNotFound(rackid string, pduID int64) *HTTPError {
	return &HTTPError{
		SC:   http.StatusNotFound,
		Base: fmt.Errorf("CloudChamber: pdu %d not found in rack %q", pduID, rackid),
	}
}

// NewErrTorNotFound indicates that the rack was found but no tor was found
//
func NewErrTorNotFound(rackid string, torID int64) *HTTPError {
	return &HTTPError{
		SC:   http.StatusNotFound,
		Base: fmt.Errorf("CloudChamber: blade %d not found in rack %q", torID, rackid),
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
