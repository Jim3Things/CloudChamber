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

    "github.com/Jim3Things/CloudChamber/internal/tracing"
    st "github.com/Jim3Things/CloudChamber/internal/tracing/server"
    pb "github.com/Jim3Things/CloudChamber/pkg/protos/admin"
)

const (
    // Value that can never be a valid version number
    InvalidRev = -1

    Login = "login"
    Logout = "logout"

    SessionCookieName = "CC-Session"
    AuthStateKey      = "authenticated"
    UserNameKey       = "username"
)

var (
    dbUsers *DBUsers
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
    routeUsers.HandleFunc(routeString, handlerUsersOperation).Queries("op", "{op}").Methods("PUT")

    // Routes for individual user operations
    routeUsers.HandleFunc(routeString, handlerUsersCreate).Methods("POST")
    routeUsers.HandleFunc(routeString, handlerUsersRead).Methods("GET")
    routeUsers.HandleFunc(routeString, handlerUsersUpdate).Methods("PUT")
    routeUsers.HandleFunc(routeString, handlerUsersDelete).Methods("DELETE")
}

// Process an http request for the list of users.  Response should contain a document of links to the
// details URI for each known user.
func handlerUsersList(w http.ResponseWriter, r *http.Request) {
    _ = st.WithSpan(context.Background(), tracing.MethodName(1), func(ctx context.Context) (err error) {
        err = doSessionHeader(
            ctx, w, r,
            func(ctx context.Context, session *sessions.Session) error {
                return canManageAccounts(session, "")
            })

        if err != nil {
            httpError(ctx, w, err)
            return err
        }

        if _, err := fmt.Fprintln(w, "Users (List)"); err != nil {
            httpError(ctx, w, err)
            return err
        }

        b := r.URL.String()
        if !strings.HasSuffix(b, "/") {
            b += "/"
        }

        return dbUsers.Scan(func(entry *pb.User) (err error) {
            target := fmt.Sprintf("%s%s", b, entry.Name)

            st.Infof(ctx, -1, "   Listing user '%s' at '%s'", entry.Name, target)

            if _, err = fmt.Fprintln(w, target); err != nil {
                httpError(ctx, w, err)
            }

            return err
        })
    })
}

func handlerUsersCreate(w http.ResponseWriter, r *http.Request) {
    _ = st.WithSpan(context.Background(), tracing.MethodName(1), func(ctx context.Context) (err error) {
        vars := mux.Vars(r)
        username := vars["username"]

        err = doSessionHeader(
            ctx, w, r,
            func(ctx context.Context, session *sessions.Session) error {
                return canManageAccounts(session, "")
            })

        if err != nil {
            httpError(ctx, w, err)
            return err
        }

        st.Infof(ctx, -1, "Creating user %q", username)

        u := &pb.UserDefinition{}
        if err = jsonpb.Unmarshal(r.Body, u); err != nil {
            httpError(ctx, w, &HTTPError{SC: http.StatusBadRequest, Base: err})
            return err
        }

        var rev int64

        if rev, err = UserAdd(username, u.Password, u.ManageAccounts, u.Enabled, false); err != nil {
            httpError(ctx, w, err)
            return err
        }

        w.Header().Set("ETag", fmt.Sprintf("%v", rev))

        st.Infof(ctx, -1, "Created user %q, pwd: %q, enabled: %v, accountManager: %v", username, u.Password, u.Enabled, u.ManageAccounts)
        _, err = fmt.Fprintf(w, "User %q created.  enabled: %v, can manage accounts: %v", username, u.Enabled, u.ManageAccounts)
        return err
    })
}

func handlerUsersRead(w http.ResponseWriter, r *http.Request) {
    _ = st.WithSpan(context.Background(), tracing.MethodName(1), func(ctx context.Context) (err error) {
        vars := mux.Vars(r)
        username := vars["username"]

        err = doSessionHeader(
            ctx, w, r,
            func(ctx context.Context, session *sessions.Session) error {
                return canManageAccounts(session, username)
            })

        if err != nil {
            httpError(ctx, w, err)
            return err
        }

        u, rev, err := dbUsers.Get(username)
        if err != nil {
            httpError(ctx, w, err)
            return err
        }

        w.Header().Set("Content-Type", "application/json")
        w.Header().Set("ETag", fmt.Sprintf("%v", rev))

        ext := &pb.UserPublic{
            Enabled:        u.Enabled,
            AccountManager: u.AccountManager,
            NeverDelete:    u.NeverDelete,
        }

        st.Infof(ctx, -1, "Returning details for user %q: %v", username, u)

        // Get the user entry, and serialize it to json
        // (export userPublic to json and return that as the body)
        p := jsonpb.Marshaler{}
        return p.Marshal(w, ext)
    })
}

// Update the user entry
func handlerUsersUpdate(w http.ResponseWriter, r *http.Request) {
    _ = st.WithSpan(context.Background(), tracing.MethodName(1), func(ctx context.Context) (err error) {
        vars := mux.Vars(r)
        username := vars["username"]

        err = doSessionHeader(
            ctx, w, r,
            func(ctx context.Context, session *sessions.Session) error {
                return canManageAccounts(session, username)
            })

        if err != nil {
            httpError(ctx, w, err)
            return err
        }

        // All updates are qualified by an ETag match.  The ETag comes from the database
        // revision number.  So, first we get the 'if-match' tag to determine the revision
        // that must be current for the update to proceed.

        var match int64

        matchString := r.Header.Get("If-Match")
        match, err = strconv.ParseInt(matchString, 10, 64)
        if err != nil {
            httpError(ctx, w, NewErrBadMatchType(matchString))
            return err
        }

        // Next, get the new definition values, and make sure that they are valid.
        upd := &pb.UserDefinition{}
        if err = jsonpb.Unmarshal(r.Body, upd); err != nil {
            httpError(ctx, w, &HTTPError{SC: http.StatusBadRequest, Base: err})
            return err
        }

        // All the prep is done.  Proceed with the update.  This may get a version
        // mismatch, or the user may have been deleted.  Given the check above, these
        // can all be considered version conflicts.
        var rev int64
        if rev, err = userUpdate(username, upd.Password, upd.ManageAccounts, upd.Enabled, match); err != nil {
            httpError(ctx, w, err)
            return err
        }

        w.Header().Set("Content-Type", "application/json")
        w.Header().Set("ETag", fmt.Sprintf("%v", rev))

        ext := &pb.UserPublic{
            Enabled:        upd.Enabled,
            AccountManager: upd.ManageAccounts,
        }

        st.Infof(ctx, -1, "Returning details for user %q: %v", username, upd)

        p := jsonpb.Marshaler{}
        return p.Marshal(w, ext)
    })
}

// Delete the user entry
func handlerUsersDelete(w http.ResponseWriter, r *http.Request) {
    _ = st.WithSpan(context.Background(), tracing.MethodName(1), func(ctx context.Context) (err error) {
        vars := mux.Vars(r)
        username := vars["username"]

        err = doSessionHeader(
            ctx, w, r,
            func(ctx context.Context, session *sessions.Session) error {
                return canManageAccounts(session, username)
            })

        if err != nil {
            httpError(ctx, w, err)
            return err
        }

        if err = userRemove(username); err != nil {
            httpError(ctx, w, err)
            return err
        }

        _, err = fmt.Fprintf(w, "User %q deleted.", username)
        return err
    })
}

// Perform an admin operation (login, logout, enable, disable) on an account
func handlerUsersOperation(w http.ResponseWriter, r *http.Request) {
    _ = st.WithSpan(context.Background(), tracing.MethodName(1), func(ctx context.Context) (err error) {
        var s string

        err = doSessionHeader(ctx, w, r, func(ctx context.Context, session *sessions.Session) (err error) {
            op := r.FormValue("op")
            vars := mux.Vars(r)
            username := vars["username"]

            st.Infof(ctx, -1, "Operation %q, user %q, session %v", op, username, session)

            switch op {
            case Login:
                s, err = login(session, r)

            case Logout:
                s, err = logout(session, r)

            default:
                err = NewErrUserInvalidOperation(op)
            }

            return err
        })

        if err != nil {
            httpError(ctx, w, err)
        } else {
            _, err = fmt.Fprintln(w, s)
        }

        return err
    })
}

// --- Route handling methods

// +++ Query tag implementations

// Query operations differ from the others inasmuch as the processing occurs
// while handling the header, and may cause updates to that header.  Therefore,
// these methods return both an error status and a string that can be later
// applied to the response body.

// Process a login request (?op=login)
func login(session *sessions.Session, r *http.Request) (_ string, err error) {
    var pwd []byte

    vars := mux.Vars(r)
    username := vars["username"]

    // Verify that there is no logged in user
    if auth, ok := session.Values[AuthStateKey].(bool); ok && auth {
        return "", &HTTPError{
            SC:   http.StatusBadRequest,
            Base: ErrUserAlreadyLoggedIn,
        }
    }

    // We have a session that could support a login.  Let's verify that this
    // is a good login

    // First, verify that this is an actual user account, and that account is
    // enabled for login operations.
    if u, _, err := dbUsers.Get(username); err != nil || !u.Enabled {
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
    if userVerifyPassword(username, pwd) != nil {
        return "", &HTTPError{
            SC:   http.StatusForbidden,
            Base: ErrUserAuthFailed,
        }
    }

    // .. all passed.  So finally mark the session as logged in
    //
    session.Values[AuthStateKey] = true
    session.Values[UserNameKey] = username

    return fmt.Sprintf("User %q logged in", username), nil
}

// Process a logout request (?op=logout)
func logout(session *sessions.Session, r *http.Request) (_ string, err error) {
    vars := mux.Vars(r)
    username := vars["username"]

    // Verify that there is a logged in user on this session
    if auth, ok := session.Values[AuthStateKey].(bool); !ok || !auth {
        return "", NewErrNoLoginActive(username)
    }

    // .. and that it is the user we're trying to logout
    if name, ok := session.Values[UserNameKey].(string); !ok || !strings.EqualFold(name, username) {
        return "", NewErrNoLoginActive(username)
    }

    // .. and now log the user out
    session.Values[AuthStateKey] = false
    delete(session.Values, UserNameKey)

    return fmt.Sprintf("User %q logged out", username), nil
}

// --- Query tag implementations

// +++ Mid-level support methods

// This section contains the methods that translate from the logical user
// attributes that are understood by the route handlers to the internal user
// attributes understood by the storage system.

func UserAdd(name string, password string, accountManager bool, enabled bool, neverDelete bool) (int64, error) {
    passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

    if err != nil {
        return InvalidRev, err
    }

    return dbUsers.Create(&pb.User{
        Name: name,
        PasswordHash: passwordHash,
        Enabled: enabled,
        AccountManager: accountManager,
        NeverDelete: neverDelete})
}

func userUpdate(name string, password string, accountManager bool, enabled bool, rev int64) (int64, error) {
    passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

    if err != nil {
        return InvalidRev, err
    }

    return dbUsers.Update(&pb.User{
        Name: name,
        PasswordHash: passwordHash,
        Enabled: enabled,
        AccountManager: accountManager}, rev)
}

func userRemove(name string) error {
    return dbUsers.Remove(name)
}

// TODO: Figure out how to better protect leakage of the password in memory.
func userVerifyPassword(name string, password []byte) error {

    entry, _, err := dbUsers.Get(name)

    if err != nil {
        return err
    }

    return bcrypt.CompareHashAndPassword(entry.PasswordHash, password)
}

// --- Mid-level support methods

// +++ Helper functions

// Determine if this session's active login has permission to change or
// manage the targeted account.  Note that any account may manage itself.
func canManageAccounts(session *sessions.Session, username string) error {
    key, ok := session.Values[UserNameKey].(string)
    if !ok {
        return &HTTPError{
            SC:   http.StatusBadRequest,
            Base: http.ErrNoCookie,
        }
    }

    user, _, err := dbUsers.Get(key)
    if err != nil {
        return NewErrUserPermissionDenied()
    }

    if !user.AccountManager && !strings.EqualFold(user.Name, username) {
        return NewErrUserPermissionDenied()
    }

    return nil
}

// Get the secret associated with this session
func secret(w http.ResponseWriter, r *http.Request) {
    session, _ := server.cookieStore.Get(r, SessionCookieName)

    // Check if user is authenticated
    if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }

    // Print secret message
    _, _ = fmt.Fprintln(w, "secret message")
}

// --- Helper functions
