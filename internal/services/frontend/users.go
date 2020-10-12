// This module contains the routines etc to implement the frontend handlers for the users
// part of the API
//

package frontend

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang/protobuf/jsonpb"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"

	"github.com/Jim3Things/CloudChamber/internal/clients/timestamp"
	"github.com/Jim3Things/CloudChamber/internal/common"
	"github.com/Jim3Things/CloudChamber/internal/tracing"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/admin"
)

const (
	// InvalidRev is a value that can never be a valid version number
	InvalidRev = -1

	// Login is a string used to select and identify the login operation
	//
	Login = "login"

	// Logout is a string used to select and identify the logout operation
	//
	Logout = "logout"
)

// +++ Route handling methods

// This section contains the route definitions, and the top level handlers for
// each route.

func usersAddRoutes(routeBase *mux.Router) {

	const routeString = "/{username:[a-z,A-Z][a-z,A-Z,0-9]*}"

	routeUsers := routeBase.PathPrefix("/users").Subrouter()

	routeUsers.HandleFunc("", handlerUsersList).Methods("GET")
	routeUsers.HandleFunc("/", handlerUsersList).Methods("GET")

	// As a reminder,
	//   PUT is idempotent so translates to UPDATE in the CRUD methodology
	//   POST is NOT idempotent so translates to CREATE in the CRUD methodology
	//

	// Routes for routes with query tags.  Note that this must come before the
	// route handlers without query tags, as query tags are optional and will
	// be matched to the first handler encountered with a matching route.
	routeUsers.HandleFunc(routeString, handlerUserOperation).Queries("op", "{op}").Methods("PUT")
	routeUsers.HandleFunc(routeString, handlerUserSetPassword).Queries("password", "{pwd}").Methods("PUT")

	// Routes for individual user operations
	routeUsers.HandleFunc(routeString, handlerUserCreate).Methods("POST")
	routeUsers.HandleFunc(routeString, handlerUserRead).Methods("GET")
	routeUsers.HandleFunc(routeString, handlerUserUpdate).Methods("PUT")
	routeUsers.HandleFunc(routeString, handlerUserDelete).Methods("DELETE")
}

// Process an http request for the list of users.  Response should contain a document of links to the
// details URI for each known user.
func handlerUsersList(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithName("Get User List"),
		tracing.WithContextValue(timestamp.EnsureTickInContext),
		tracing.AsInternal())
	defer span.End()

	err := doSessionHeader(
		ctx, w, r,
		func(ctx context.Context, session *sessions.Session) error {
			return canManageAccounts(ctx, session, "")
		})

	if err != nil {
		postHTTPError(ctx, w, err)
		return
	}

	b := common.URLPrefix(r)

	w.Header().Set("Content-Type", "application/json")

	users := &pb.UserList{}

	err = dbUsers.Scan(ctx, func(entry *pb.User) error {
		target := fmt.Sprintf("%s%s", b, entry.Name)

		protected := ""
		if entry.NeverDelete {
			protected = " (protected)"
		}

		tracing.Info(ctx,
			tracing.WithImpactRead("/users"),
			"   Listing user %q: %q%s", entry.Name, target, protected)

		users.Users = append(users.Users, &pb.UserListEntry{
			Name:      entry.Name,
			Uri:       target,
			Protected: entry.NeverDelete,
		})

		return nil
	})

	if err != nil {
		postHTTPError(ctx, w, err)
		return
	}

	p := jsonpb.Marshaler{}
	err = p.Marshal(w, users)

	httpErrorIf(ctx, w, err)
}

func handlerUserCreate(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithName("Create New User"),
		tracing.WithContextValue(timestamp.EnsureTickInContext),
		tracing.AsInternal())
	defer span.End()

	vars := mux.Vars(r)
	username := vars["username"]

	err := doSessionHeader(
		ctx, w, r,
		func(ctx context.Context, session *sessions.Session) error {
			return canManageAccounts(ctx, session, "")
		})

	if err != nil {
		postHTTPError(ctx, w, err)
		return
	}

	tracing.UpdateSpanName(ctx, "Creating user %q", username)

	u := &pb.UserDefinition{}
	if err = jsonpb.Unmarshal(r.Body, u); err != nil {
		postHTTPError(ctx, w, &HTTPError{SC: http.StatusBadRequest, Base: err})
		return
	}

	var rev int64

	if rev, err = userAdd(ctx, username, u.Password, u.CanManageAccounts, u.Enabled, false); err != nil {
		postHTTPError(ctx, w, err)
		return
	}

	w.Header().Set("ETag", fmt.Sprintf("%v", rev))

	tracing.Info(
		ctx,
		tracing.WithImpactCreate("/users/"+username),
		"Created user %q, pwd: <redacted>, enabled: %v, accountManager: %v", username, u.Enabled, u.CanManageAccounts)

	_, err = fmt.Fprintf(
		w,
		"User %q created.  enabled: %v, can manage accounts: %v",
		username, u.Enabled, u.CanManageAccounts)

	httpErrorIf(ctx, w, err)
}

func handlerUserRead(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithName("Get User Details"),
		tracing.WithContextValue(timestamp.EnsureTickInContext),
		tracing.AsInternal())
	defer span.End()

	vars := mux.Vars(r)
	username := vars["username"]

	err := doSessionHeader(
		ctx, w, r,
		func(ctx context.Context, session *sessions.Session) error {
			return canManageAccounts(ctx, session, username)
		})

	if err != nil {
		postHTTPError(ctx, w, err)
		return
	}

	tracing.UpdateSpanName(ctx, "Getting details for user %q", username)

	u, rev, err := userRead(ctx, username)

	if err != nil {
		postHTTPError(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("ETag", fmt.Sprintf("%v", rev))

	ext := &pb.UserPublic{
		Enabled:           u.Enabled,
		CanManageAccounts: u.CanManageAccounts,
		NeverDelete:       u.NeverDelete,
	}

	tracing.Info(
		ctx,
		tracing.WithImpactRead("/users/"+username),
		"Returning details for user %s", formatUser(username, ext))

	// Get the user entry, and serialize it to json
	// (export userPublic to json and return that as the body)
	p := jsonpb.Marshaler{}
	err = p.Marshal(w, ext)

	httpErrorIf(ctx, w, err)
}

// Update the user entry
func handlerUserUpdate(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithName("Update User Details"),
		tracing.WithContextValue(timestamp.EnsureTickInContext),
		tracing.AsInternal())
	defer span.End()

	vars := mux.Vars(r)
	username := vars["username"]

	var caller *pb.User

	err := doSessionHeader(
		ctx, w, r,
		func(ctx context.Context, session *sessions.Session) (err error) {
			caller, err = getLoggedInUser(ctx, session)
			if err != nil {
				return err
			}

			return canManageAccounts(ctx, session, username)
		})

	if err != nil {
		postHTTPError(ctx, w, err)
		return
	}

	tracing.UpdateSpanName(ctx, "Updating details on user %q", username)

	// All updates are qualified by an ETag match.  The ETag comes from the database
	// revision number.  So, first we get the 'if-match' tag to determine the revision
	// that must be current for the update to proceed.

	var match int64

	matchString := r.Header.Get("If-Match")
	match, err = strconv.ParseInt(matchString, 10, 64)
	if err != nil {
		postHTTPError(ctx, w, NewErrBadMatchType(matchString))
		return
	}

	// Next, get the new definition values, and make sure that they are valid.
	upd := &pb.UserUpdate{}
	if err = jsonpb.Unmarshal(r.Body, upd); err != nil {
		postHTTPError(ctx, w, &HTTPError{SC: http.StatusBadRequest, Base: err})
		return
	}

	// Finally, check that no rights are being added that the logged in user does
	// not have.  Since a user can modify their own entries, the canManageAccounts
	// check is insufficient.
	if err = verifyRightsAvailable(caller, upd); err != nil {
		postHTTPError(ctx, w, err)
		return
	}

	// All the prep is done.  Proceed with the update.  This may get a version
	// mismatch, or the user may have been deleted.  Given the check above, these
	// can all be considered version conflicts.
	var rev int64
	var newVer *pb.User

	if newVer, rev, err = userUpdate(ctx, username, upd, match); err != nil {
		postHTTPError(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("ETag", fmt.Sprintf("%v", rev))

	ext := &pb.UserPublic{
		Enabled:           newVer.Enabled,
		CanManageAccounts: newVer.CanManageAccounts,
		NeverDelete:       newVer.NeverDelete,
	}

	tracing.Info(
		ctx,
		tracing.WithImpactModify("/users/"+username),
		"Returning details for user %s", formatUser(username, ext))

	p := jsonpb.Marshaler{}
	err = p.Marshal(w, ext)

	httpErrorIf(ctx, w, err)
}

// Delete the user entry
func handlerUserDelete(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithName("Delete User"),
		tracing.WithContextValue(timestamp.EnsureTickInContext),
		tracing.AsInternal())
	defer span.End()

	vars := mux.Vars(r)
	username := vars["username"]

	err := doSessionHeader(
		ctx, w, r,
		func(ctx context.Context, session *sessions.Session) error {
			return canManageAccounts(ctx, session, username)
		})

	if err != nil {
		postHTTPError(ctx, w, err)
		return
	}

	tracing.UpdateSpanName(ctx, "Deleting user %q", username)

	if err = userRemove(ctx, username); err != nil {
		postHTTPError(ctx, w, err)
		return
	}

	_, err = fmt.Fprintf(w, "User %q deleted.", username)

	httpErrorIf(ctx, w, err)
}

// Perform an admin operation (login, logout, enable, disable) on an account
func handlerUserOperation(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithName("Perform User Operation"),
		tracing.WithContextValue(timestamp.EnsureTickInContext),
		tracing.AsInternal())
	defer span.End()

	var s string

	err := doSessionHeader(ctx, w, r, func(ctx context.Context, session *sessions.Session) (err error) {
		op := r.FormValue("op")
		vars := mux.Vars(r)
		username := vars["username"]

		switch op {
		case Login:
			tracing.UpdateSpanName(ctx, "Logging in user %q", username)
			s, err = login(ctx, session, r)

		case Logout:
			tracing.UpdateSpanName(ctx, "Logging out user %q", username)
			s, err = logout(ctx, session, r)

		default:
			err = NewErrUserInvalidOperation(op, username)
		}

		if err != nil {
			_ = tracing.Error(ctx, dumpSessionState(session))
		}
		return err
	})

	if err != nil {
		postHTTPError(ctx, w, err)
		return
	}

	_, err = fmt.Fprintln(w, s)

	httpErrorIf(ctx, w, err)
}

// Set a new password on the specified account
func handlerUserSetPassword(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithName("Perform User Operation"),
		tracing.WithContextValue(timestamp.EnsureTickInContext),
		tracing.AsInternal())
	defer span.End()

	vars := mux.Vars(r)
	username := vars["username"]

	err := doSessionHeader(
		ctx, w, r,
		func(ctx context.Context, session *sessions.Session) (err error) {
			return canManageAccounts(ctx, session, username)
		})

	if err != nil {
		postHTTPError(ctx, w, err)
		return
	}

	tracing.UpdateSpanName(ctx, "Updating the password for user %q", username)

	// All updates are qualified by an ETag match.  The ETag comes from the database
	// revision number.  So, first we get the 'if-match' tag to determine the revision
	// that must be current for the update to proceed.

	var match int64

	matchString := r.Header.Get("If-Match")
	match, err = strconv.ParseInt(matchString, 10, 64)
	if err != nil {
		postHTTPError(ctx, w, NewErrBadMatchType(matchString))
		return
	}

	// Next, get the new password values, and make sure that they are valid.
	upd := &pb.UserPassword{}
	if err = jsonpb.Unmarshal(r.Body, upd); err != nil {
		postHTTPError(ctx, w, &HTTPError{SC: http.StatusBadRequest, Base: err})
		return
	}

	// All the prep is done.  Proceed to try to set.  This may get a version
	// mismatch, the user may have been deleted, or the old password may not
	// match.
	var rev int64

	if rev, err = userSetPassword(ctx, username, upd, match); err != nil {
		postHTTPError(ctx, w, err)
		return
	}

	w.Header().Set("ETag", fmt.Sprintf("%v", rev))

	tracing.Info(
		ctx,
		tracing.WithImpactModify("/users/"+username),
		"Password changed for user %q", username)
	_, err = fmt.Fprintf(w, "Password changed for user %q", username)

	httpErrorIf(ctx, w, err)
}

// --- Route handling methods

// +++ Query tag implementations

// Query operations differ from the others inasmuch as the processing occurs
// while handling the header, and may cause updates to that header.  Therefore,
// these methods return both an error status and a string that can be later
// applied to the response body.

// Process a login request (?op=login)
func login(ctx context.Context, session *sessions.Session, r *http.Request) (_ string, err error) {
	var pwd []byte

	vars := mux.Vars(r)
	username := vars["username"]

	// Verify that there is no logged in user
	if _, ok := getSession(session); ok {
		return "", &HTTPError{
			SC:   http.StatusBadRequest,
			Base: ErrUserAlreadyLoggedIn,
		}
	}

	// We have a session that could support a login.  Let's verify that this
	// is a good login

	// First, verify that this is an actual user account, and that account is
	// enabled for login operations.
	if u, _, err := userRead(ctx, username); err != nil || !u.Enabled {
		return "", &HTTPError{
			SC:   http.StatusNotFound,
			Base: ErrUserAuthFailed,
		}
	}

	// .. next, let's get the password, which is the body of the request
	if pwd, err = ioutil.ReadAll(r.Body); err != nil {
		return "", &HTTPError{
			SC:   http.StatusBadRequest,
			Base: ErrUserAuthFailed,
		}
	}

	// .. finally, let's confirm that this password matches the one for the
	// designated user account.
	if userVerifyPassword(ctx, username, pwd) != nil {
		return "", &HTTPError{
			SC:   http.StatusForbidden,
			Base: ErrUserAuthFailed,
		}
	}

	// .. all passed.  So finally mark the session as logged in
	//
	if err = newSession(session, sessionState{name: username}); err != nil {
		return "", &HTTPError{
			SC:   http.StatusBadRequest,
			Base: err,
		}
	}

	return fmt.Sprintf("User %q logged in", username), nil
}

// Process a logout request (?op=logout)
func logout(_ context.Context, session *sessions.Session, r *http.Request) (_ string, err error) {
	vars := mux.Vars(r)
	username := vars["username"]

	// Verify that there is a logged in user on this session
	// .. and that it is the user we're trying to logout
	if state, ok := getSession(session); !ok || !strings.EqualFold(state.name, username) {
		return "", NewErrNoLoginActive(username)
	}

	// .. and now log the user out
	removeSession(session)

	return fmt.Sprintf("User %q logged out", username), nil
}

// --- Query tag implementations

// +++ Mid-level support methods

// This section contains the methods that translate from the logical user
// attributes that are understood by the route handlers to the internal user
// attributes understood by the storage system.

func userAdd(
	ctx context.Context,
	name string,
	password string,
	accountManager bool,
	enabled bool,
	neverDelete bool) (int64, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return InvalidRev, err
	}

	revision, err := dbUsers.Create(ctx, &pb.User{
		Name:              name,
		PasswordHash:      passwordHash,
		Enabled:           enabled,
		CanManageAccounts: accountManager,
		NeverDelete:       neverDelete})

	if err == ErrUserAlreadyExists(name) {
		return InvalidRev, NewErrUserAlreadyExists(name)
	}

	if err != nil {
		return InvalidRev, err
	}

	return revision, nil
}

func userUpdate(ctx context.Context, name string, u *pb.UserUpdate, rev int64) (*pb.User, int64, error) {
	upd, revision, err := dbUsers.Update(ctx, name, u, rev)

	if err == ErrUserNotFound(name) {
		return nil, InvalidRev, NewErrUserNotFound(name)
	}

	if err == ErrUserStaleVersion(name) {
		return nil, InvalidRev, NewErrUserStaleVersion(name)
	}

	if err != nil {
		return nil, InvalidRev, err
	}

	return upd, revision, nil
}

func userRead(ctx context.Context, name string) (*pb.User, int64, error) {

	u, rev, err := dbUsers.Read(ctx, name)

	if err == ErrUserNotFound(name) {
		return nil, InvalidRev, NewErrUserNotFound(name)
	}

	if err != nil {
		return nil, InvalidRev, err
	}

	return u, rev, nil
}

func userRemove(ctx context.Context, name string) error {
	err := dbUsers.Delete(ctx, name, InvalidRev)

	if err == ErrUserProtected(name) {
		return NewErrUserProtected(name)
	}

	if err == ErrUserStaleVersion(name) {
		return NewErrUserStaleVersion(name)
	}

	if err == ErrUserNotFound(name) {
		return NewErrUserNotFound(name)
	}

	return err
}

// TODO: Figure out how to better protect leakage of the password in memory.

// Verify that the password matches the user's current hashed password
func userVerifyPassword(ctx context.Context, name string, password []byte) error {

	entry, _, err := userRead(ctx, name)

	if err != nil {
		return err
	}

	return bcrypt.CompareHashAndPassword(entry.PasswordHash, password)
}

// Set the password for a given user account, after first verifying that
// the current password was correctly provided (or an override was in place)
func userSetPassword(ctx context.Context, name string, changes *pb.UserPassword, rev int64) (int64, error) {
	if !changes.Force {
		if err := userVerifyPassword(ctx, name, []byte(changes.OldPassword)); err != nil {
			return InvalidRev, NewErrUserPermissionDenied()
		}
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(changes.NewPassword), bcrypt.DefaultCost)

	if err != nil {
		return InvalidRev, err
	}

	_, revision, err := dbUsers.UpdatePassword(ctx, name, passwordHash, rev)

	if err == ErrUserNotFound(name) {
		return InvalidRev, NewErrUserNotFound(name)
	}

	if err == ErrUserStaleVersion(name) {
		return InvalidRev, NewErrUserStaleVersion(name)
	}

	if err != nil {
		return InvalidRev, err
	}

	return revision, nil

}

// --- Mid-level support methods

// +++ Helper functions

// Determine if this session's active login has permission to change or
// manage the targeted account.  Note that any account may manage itself.
func canManageAccounts(ctx context.Context, session *sessions.Session, username string) error {
	user, err := getLoggedInUser(ctx, session)
	if err != nil {
		return NewErrUserPermissionDenied()
	}

	if !user.CanManageAccounts && !strings.EqualFold(user.Name, username) {
		return NewErrUserPermissionDenied()
	}

	return nil
}

func verifyRightsAvailable(current *pb.User, upd *pb.UserUpdate) error {
	if current.CanManageAccounts {
		return nil
	}

	if upd.CanManageAccounts {
		return NewErrUserPermissionDenied()
	}

	return nil
}

func formatUser(name string, user *pb.UserPublic) string {
	var attrs []string

	if user.NeverDelete {
		attrs = append(attrs, "protected")
	}

	if user.Enabled {
		attrs = append(attrs, "enabled")
	} else {
		attrs = append(attrs, "disabled")
	}

	if user.CanManageAccounts {
		attrs = append(attrs, "can manage accounts")
	}

	desc := strings.Join(attrs, ", ")

	return fmt.Sprintf("%q: %s", name, desc)
}

// Get the secret associated with this session
// func secret(w http.ResponseWriter, r *http.Request) {
//     session, _ := server.cookieStore.Get(r, SessionCookieName)
//
//     // Check if user is authenticated
//     if state, ok := getSession(session); !ok || !state.authenticated {
//         http.Error(w, "Forbidden", http.StatusForbidden)
//         return
//     }
//
//     // Print secret message
//     _, _ = fmt.Fprintln(w, "secret message")
// }

// --- Helper functions
