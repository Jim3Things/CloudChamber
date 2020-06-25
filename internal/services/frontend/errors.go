// This module contains the defined errors, and extended error types, that are
// specific to the frontend package

package frontend

import (
    "context"
    "errors"
    "fmt"
    "net/http"

    st "github.com/Jim3Things/CloudChamber/internal/tracing/server"
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

// Custom common HTTP error type that includes the status code to use in
// the response.
type HTTPError struct {
	// HTTP status code
	SC int

	// Underlying Go error
	Base error
}

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

// Set an http error, and log it to the tracing system.
func httpError(ctx context.Context, w http.ResponseWriter, err error) {
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

    _ = st.Errorf(ctx, -1, "http error %v: %s", he.StatusCode(), he.Error())
    http.Error(w, he.Error(), he.StatusCode())
}

// +++ HTTPError specializations

// ErrNoLoginActive indicates that the specified user is not logged into this session
func NewErrNoLoginActive(name string) *HTTPError {
	return &HTTPError{
		SC:   http.StatusBadRequest,
		Base: fmt.Errorf("CloudChamber: user %q not logged into this session", name),
	}
}

// ErrUserNotFound indicates the specified user account was determined to
// not exist (i.e. the search succeeded but no record was found)
//
func NewErrUserNotFound(name string) *HTTPError {
	return &HTTPError{
		SC:   http.StatusNotFound,
		Base: fmt.Errorf("CloudChamber: user %q not found", name),
	}
}

// ErrUserAlreadyCreated indicates the specified user account was previously
// created and the request was determined to be a duplicate Create request.
//
func NewErrUserAlreadyCreated(name string) *HTTPError {
	return &HTTPError{
		SC:   http.StatusBadRequest,
		Base: fmt.Errorf("CloudChamber: user %q already exists", name),
	}
}

// ErrUserPermissionDenied indicates the user does not have the appropriate
// permissions for the requested operation.
//
func NewErrUserPermissionDenied() *HTTPError {
	return &HTTPError{
		SC:   http.StatusForbidden,
		Base: errors.New("CloudChamber: permission denied"),
	}
}

// ErrUserStaleVersion indicates that an operation against the specified user
// expected a different revision number than was found
//
func NewErrUserStaleVersion(name string) *HTTPError {
	return &HTTPError{
		SC:   http.StatusConflict,
		Base: fmt.Errorf("CloudChamber: user %q has a newer version than expected", name),
	}
}

// ErrBadMatchType indicates that the If-Match value was syntactically incorrect,
// and could not be processed
func NewErrBadMatchType(match string) *HTTPError {
	return &HTTPError{
		SC:   http.StatusBadRequest,
		Base: fmt.Errorf("CloudChamber: match value %q is not recognized as a valid integer", match),
	}
}

// ErrUserInvalidOperation indicates the operation requested for the supplied
// user account is invalid in some way, likely a non-existent operation code.
//
func NewErrUserInvalidOperation(op string) *HTTPError {
	return &HTTPError{
		SC:   http.StatusBadRequest,
		Base: fmt.Errorf("CloudChamber: invalid user operation requested (?op=%s)", op),
	}
}

func NewErrUserProtected(name string) *HTTPError {
	return &HTTPError{
		SC:   http.StatusForbidden,
		Base: fmt.Errorf("CloudChamber: user %q is protected and cannot be deleted", name),
	}
}

// ErrRackNotFound indicates the specified rack do not exist and the http
// request (http.statusNotFound) determines to be the request was made against
// a non-existing rack.
func NewErrRackNotFound(name string) *HTTPError {
	return &HTTPError{
		SC:   http.StatusNotFound,
		Base: fmt.Errorf("CloudChamber: rack %q not found", name),
	}
}

func NewErrInvalidStepperMode(mode string) *HTTPError {
    return &HTTPError{
        SC:   http.StatusBadRequest,
        Base: fmt.Errorf("CloudChamber: mode %q is invalid.  Supported modes are 'manual' and 'automatic'", mode),
    }
}

func NewErrInvalidRateRequest() *HTTPError {
    return &HTTPError{
        SC:   http.StatusBadRequest,
        Base: errors.New("CloudChamber: manual mode does not accept additional arguments"),
    }
}

func NewErrInvalidStepperRate(rate string) *HTTPError {
    return &HTTPError{
        SC:   http.StatusBadRequest,
        Base: fmt.Errorf("CloudChamber: requested rate %q could not be parsed as a positive decimal number", rate),
    }
}

func NewErrStepperFailedToSetPolicy() *HTTPError {
    return &HTTPError{
        SC:   http.StatusBadRequest,
        Base: errors.New("CloudChamber: Set simulated time policy operation failed"),
    }
}

// --- HTTPError specializations
