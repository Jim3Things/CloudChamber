// This module contains the routines etc to implement the frontend handlers for the users
// part of the API
//

package frontend

import (
    "context"
    "fmt"
    "io/ioutil"
    "net/http"
    "runtime"
    "strconv"
    "strings"
    "sync"

    "github.com/golang/protobuf/jsonpb"
    "github.com/gorilla/mux"
    "github.com/gorilla/sessions"
    "go.opentelemetry.io/otel/api/trace"
    "golang.org/x/crypto/bcrypt"

    "github.com/Jim3Things/CloudChamber/internal/config"
    pb "github.com/Jim3Things/CloudChamber/pkg/protos/admin"
)

const (
    // Value that can never be a valid version number
    InvalidRev = -1

    Enable = "enable"
    Disable = "disable"
    Login = "login"
    Logout = "logout"

    SessionCookieName = "CC-Session"
    AuthStateKey      = "authenticated"
    UserNameKey       = "username"
)

// +++ User collection access mechanism
// This section encapsulates storage and retrieval of known users

// The full user entry contains attributes about what it can do, the password
// hash, and a revision number.  The password hash is never exposed outside of
// this module.  The revision number is returned, and used as a precondition
// on any update requests.

// Each user entry has an associated key which is the lowercased form of the
// username.  The supplied name is retained as an attribute in order to present
// the form that the caller originally used for display purposes.

// DBUsers is a container used to established synchronized access to
// the in-memory set of user records.
//
type DBUsers struct {
    Mutex sync.Mutex
    Users map[string]pb.UserInternal
}

// Initialize the users store.  For now this is only a map in memory.
func InitDBUsers(cfg *config.GlobalConfig) error {
    if dbUsers == nil {
        dbUsers = &DBUsers{
            Mutex: sync.Mutex{},
            Users: make(map[string]pb.UserInternal),
        }
    }

    _, err := UserAdd(cfg.WebServer.SystemAccount, cfg.WebServer.SystemAccountPassword, true, true)
    return err
}

// Create a new user entry in the store.
func (m *DBUsers) Create(u *pb.User) (int64, error) {
    m.Mutex.Lock()
    defer m.Mutex.Unlock()

    key := strings.ToLower(u.Name)

    if _, ok := m.Users[key]; ok {
        return InvalidRev, NewErrUserAlreadyCreated(u.Name)
    }

    entry := &pb.UserInternal{
        User:                 &pb.User{
            Name:                 u.Name,
            PasswordHash:         u.PasswordHash,
            UserId:               u.UserId,
            Enabled:              u.Enabled,
            AccountManager:       u.AccountManager,
        },
        Revision:             1,
    }
    m.Users[key] = *entry

    return 1, nil
}

// Get the specified user from the store.
func (m *DBUsers) Get(name string) (*pb.User, int64, error) {
    m.Mutex.Lock()
    defer m.Mutex.Unlock()

    key := strings.ToLower(name)

    entry, ok := m.Users[key]
    if !ok {
        return nil, InvalidRev, NewErrUserNotFound(name)
    }

    return entry.User, entry.Revision, nil
}

// Scan the set of known users in the store, invoking the supplied
// function with each entry.
func (m *DBUsers) Scan(action func(entry *pb.User) error) error {
    m.Mutex.Lock()
    defer m.Mutex.Unlock()

    for _, user := range dbUsers.Users {
        if err := action(user.User); err != nil {
            return err
        }
    }

    return nil
}


func (m *DBUsers) Update(u *pb.User, match int64) (int64, error) {
    m.Mutex.Lock()
    defer m.Mutex.Unlock()

    key := strings.ToLower(u.Name)

    old, ok := m.Users[key]
    if !ok {
        return InvalidRev, NewErrUserNotFound(u.Name)
    }

    if old.Revision != match {
        return InvalidRev, NewErrUserStaleVersion(u.Name)
    }

    entry := &pb.UserInternal{
        User:                 &pb.User{
            Name:                 u.Name,
            PasswordHash:         u.PasswordHash,
            UserId:               u.UserId,
            Enabled:              u.Enabled,
            AccountManager:       u.AccountManager,
        },
        Revision:             match + 1,
    }
    m.Users[key] = *entry

    return entry.Revision, nil
}

func (m *DBUsers) Remove(name string) error {
    m.Mutex.Lock()
    defer m.Mutex.Unlock()

    key := strings.ToLower(name)

    _, ok := m.Users[key]
    if !ok {
        return NewErrUserNotFound(name)
    }

    delete(m.Users, key)
    return nil
}

// --- End User collection access mechanism


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
    _ = tr.WithSpan(context.Background(), methodName(), func(ctx context.Context) (err error) {
        err = doSessionHeader(
            ctx, w, r,
            func(ctx context.Context, span trace.Span, session *sessions.Session) error {
                return canManageAccounts(session, "")
            })

        span := trace.SpanFromContext(ctx)
        if err != nil {
            httpError(ctx, span, w, err)
            return err
        }

        if _, err := fmt.Fprintln(w, "Users (List)"); err != nil {
            httpError(ctx, span, w, err)
           return err
        }

        b := r.URL.String()
        if !strings.HasSuffix(b, "/") {
            b += "/"
        }

        return dbUsers.Scan(func(entry *pb.User) (err error) {
            target := fmt.Sprintf("%s%s", b, entry.Name)

            span.AddEvent(ctx, fmt.Sprintf("   Listing user '%s' at '%s'", entry.Name, target))

            if _, err = fmt.Fprintln(w, target); err != nil {
                httpError(ctx, span, w, err)
            }

            return err
        })
    })
}

func handlerUsersCreate(w http.ResponseWriter, r *http.Request) {
    _ = tr.WithSpan(context.Background(), methodName(), func(ctx context.Context) (err error) {
        vars := mux.Vars(r)
        username := vars["username"]

        err = doSessionHeader(
            ctx, w, r,
            func(ctx context.Context, span trace.Span, session *sessions.Session) error {
                return canManageAccounts(session, "")
            })

        span := trace.SpanFromContext(ctx)
        if err != nil {
            httpError(ctx, span, w, err)
            return err
        }

        span.AddEvent(ctx, fmt.Sprintf("Creating user %q", username))

        u := &pb.UserDefinition{}
        if err = jsonpb.Unmarshal(r.Body, u); err != nil {
            httpError(ctx, span, w, &HTTPError{ SC: http.StatusBadRequest, Base: err })
            return err
        }

        var rev int64

        if rev, err = UserAdd(username, u.Password, u.ManageAccounts, u.Enabled); err != nil {
            httpError(ctx, span, w, err)
            return err
        }

        w.Header().Set("ETag", fmt.Sprintf("%v", rev))

        span.AddEvent(ctx, fmt.Sprintf("Created user %q, pwd: %q, enabled: %v, accountManager: %v", username, u.Password, u.Enabled, u.ManageAccounts))
        _, err = fmt.Fprintf(w, "User %q created.  enabled: %v, can manage accounts: %v", username, u.Enabled, u.ManageAccounts)
        return err
    })
}

func handlerUsersRead(w http.ResponseWriter, r *http.Request) {
    _ = tr.WithSpan(context.Background(), methodName(), func(ctx context.Context) (err error) {
        vars := mux.Vars(r)
        username := vars["username"]

        err = doSessionHeader(
            ctx, w, r,
            func(ctx context.Context, span trace.Span, session *sessions.Session) error {
                return canManageAccounts(session, username)
            })

        span := trace.SpanFromContext(ctx)
        if err != nil {
            httpError(ctx, span, w, err)
            return err
        }

        u, rev, err := dbUsers.Get(username)
        if err != nil {
            httpError(ctx, span, w, err)
            return err
        }

        w.Header().Set("Content-Type", "application/json")
        w.Header().Set("ETag", fmt.Sprintf("%v", rev))

        ext := &pb.UserPublic{
            Enabled:              u.Enabled,
            AccountManager:       u.AccountManager,
        }

        span.AddEvent(ctx, fmt.Sprintf("Returning details for user %q: %v", username, u))

        // Get the user entry, and serialize it to json
        // (export userPublic to json and return that as the body)
        p := jsonpb.Marshaler{}
        return p.Marshal(w, ext)
    })
}

// Update the user entry
func handlerUsersUpdate(w http.ResponseWriter, r *http.Request) {
    _ = tr.WithSpan(context.Background(), methodName(), func(ctx context.Context) (err error) {
        vars := mux.Vars(r)
        username := vars["username"]

        err = doSessionHeader(
            ctx, w, r,
            func(ctx context.Context, span trace.Span, session *sessions.Session) error {
                return canManageAccounts(session, username)
            })

        span := trace.SpanFromContext(ctx)
        if err != nil {
            httpError(ctx, span, w, err)
            return err
        }

        // All updates are qualified by an ETag match.  The ETag comes from the database
        // revision number.  So, first we get the 'if-match' tag to determine the revision
        // that must be current for the update to proceed.

        var match int64

        matchString := r.Header.Get("If-Match")
        match, err = strconv.ParseInt(matchString, 10, 64)
        if err != nil {
            httpError(ctx, span, w, NewErrBadMatchType(matchString))
            return err
        }

        // Next, get the new definition values, and make sure that they are valid.
        upd := &pb.UserDefinition{}
        if err = jsonpb.Unmarshal(r.Body, upd); err != nil {
            httpError(ctx, span, w, &HTTPError{ SC: http.StatusBadRequest, Base: err })
            return err
        }

        // Now with the input in hand, get the user (if it exists).
        _, rev, err := dbUsers.Get(username)
        if err != nil {
            httpError(ctx, span, w, err)
            return err
        }

        // All the prep is done.  Proceed with the update.  This may get a version
        // mismatch, or the user may have been deleted.  Given the check above, these
        // can all be considered version conflicts.
        if rev, err = userUpdate(username, upd.Password, upd.ManageAccounts, upd.Enabled, match); err != nil {
            httpError(ctx, span, w, err)
            return err
        }

        w.Header().Set("Content-Type", "application/json")
        w.Header().Set("ETag", fmt.Sprintf("%v", rev))

        ext := &pb.UserPublic{
            Enabled:              upd.Enabled,
            AccountManager:       upd.ManageAccounts,
        }

        span.AddEvent(ctx, fmt.Sprintf("Returning details for user %q: %v", username, upd))

        p := jsonpb.Marshaler{}
        return p.Marshal(w, ext)
    })
}

// Delete the user entry
func handlerUsersDelete(w http.ResponseWriter, r *http.Request) {
    _ = tr.WithSpan(context.Background(), methodName(), func(ctx context.Context) (err error) {
        vars := mux.Vars(r)
        username := vars["username"]

        err = doSessionHeader(
            ctx, w, r,
            func(ctx context.Context, span trace.Span, session *sessions.Session) error {
                return canManageAccounts(session, username)
            })

        span := trace.SpanFromContext(ctx)
        if err != nil {
            httpError(ctx, span, w, err)
            return err
        }

        if err = userRemove(username); err != nil {
            httpError(ctx, span, w, err)
            return err
        }

        _, err = fmt.Fprintf(w, "User %q deleted.", username)
        return err
    })
}

// Perform an admin operation (login, logout, enable, disable) on an account
func handlerUsersOperation(w http.ResponseWriter, r *http.Request) {
    _ = tr.WithSpan(context.Background(), methodName(), func(ctx context.Context) (err error) {
        var s string

        err = doSessionHeader(ctx, w, r, func(ctx context.Context, span trace.Span, session *sessions.Session) (err error) {
            op := r.FormValue("op")
            vars := mux.Vars(r)
            username := vars["username"]

            span.AddEvent(ctx, fmt.Sprintf("Operation %q, user %q, session %v", op, username, session))

            switch op {
            case Enable:
                s, err = enableDisable(session, r, true, "enable")

            case Disable:
                s, err = enableDisable(session, r, false, "disable")

            case Login:
                s, err = login(session, r)

            case Logout:
                s, err = logout(session, r)

            default:
                err = &HTTPError{
                    SC:   http.StatusBadRequest,
                    Base: fmt.Errorf("invalid user operation requested (?op=%s)", op),
                }
            }

            return err
        })

        if err != nil {
            span := trace.SpanFromContext(ctx)
            httpError(ctx, span, w, err)
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

// Process a request to enable (?op=enable), or disable (?op=disable) an account
func enableDisable(session *sessions.Session, r *http.Request, v bool, s string) (string, error) {
    vars := mux.Vars(r)
    username := vars["username"]

    if err := canManageAccounts(session, username); err != nil {
        return "", err
    }

    if err := userEnable(username, v); err != nil {
        return "", err
    }

    return fmt.Sprintf("User %q %sd", username, s), nil
}

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

func UserAdd(name string, password string, accountManager bool, enabled bool) (int64, error) {
    passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

    if err != nil {
        return InvalidRev, err
    }

    return dbUsers.Create(&pb.User{
        Name: name,
        PasswordHash: passwordHash,
        Enabled: enabled,
        AccountManager: accountManager})
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

func userEnable(name string, enable bool) error {

    entry, rev, err := dbUsers.Get(name)
    if err != nil {
        return err
    }

    entry.Enabled = enable
    _, err = dbUsers.Update(entry, rev)
    return err
}

// --- Mid-level support methods

// +++ Helper functions

// Return the caller's fully qualified method name
func methodName() string {
    fpcs := make([]uintptr, 1)

    // Get the caller's caller information (i.e. the caller of this method)
    if runtime.Callers(2, fpcs) == 0 {
        return "?"
    }

    caller := runtime.FuncForPC(fpcs[0] - 1)
    if caller == nil {
        return "?"
    }

    // ... and return the name
    return caller.Name()
}

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
