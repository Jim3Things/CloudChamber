// This module containes the routines etc to implement the frontend handlers for the users
// part of the API
//

package frontend

import (
    "context"
    "fmt"
    "io/ioutil"
    "net/http"
    "runtime"
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

// DbUsers is a container used to established synchronized access to
// the in-memory set of user records.
//
type DbUsers struct {
    Mutex sync.Mutex
    Users map[string]pb.UserInternal
}

// Initialize the users store.  For now this is only a map in memory.
func InitDbUsers(cfg *config.GlobalConfig) error {
    if dbUsers == nil {
        dbUsers = &DbUsers{
            Mutex: sync.Mutex{},
            Users: make(map[string]pb.UserInternal),
        }
    }

    return UserAdd(cfg.WebServer.SystemAccount, cfg.WebServer.SystemAccountPassword, true, true)
}

// Create a new user entry in the store.
func (m *DbUsers) Create(u *pb.User) error {
    m.Mutex.Lock()
    defer m.Mutex.Unlock()

    key := strings.ToLower(u.Name)

    if _, ok := m.Users[key]; ok {
        return ErrUserAlreadyCreated(u.Name)
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

    return nil
}

// Get the specified user from the store.
func (m *DbUsers) Get(name string) (*pb.User, int64, error) {
    m.Mutex.Lock()
    defer m.Mutex.Unlock()

    key := strings.ToLower(name)

    entry, ok := m.Users[key]
    if !ok {
        return nil, -1, ErrUserNotFound(name)
    }

    return entry.User, entry.Revision, nil
}

// Scan the set of known users in the store, invoking the supplied
// function with each entry.
func (m *DbUsers) Scan(action func(entry *pb.User) error) error {
    m.Mutex.Lock()
    defer m.Mutex.Unlock()

    for _, user := range dbUsers.Users {
        if err := action(user.User); err != nil {
            return err
        }
    }

    return nil
}


func (m *DbUsers) Update(u *pb.User, match int64) error {
    m.Mutex.Lock()
    defer m.Mutex.Unlock()

    key := strings.ToLower(u.Name)

    old, ok := m.Users[key]
    if !ok || old.Revision != match {
        return ErrUserUpdateFailed(u.Name)
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

    return nil
}

func (m *DbUsers) Remove(name string) error {
    m.Mutex.Lock()
    defer m.Mutex.Unlock()

    key := strings.ToLower(name)

    _, ok := m.Users[key]
    if !ok {
        return ErrUserNotFound(name)
    }

    delete(m.Users, key)
    return nil
}

// --- End User collection access mechanism


var (
    dbUsers *DbUsers
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
    _ = tr.WithSpan(context.Background(), methodName(), func(ctx context.Context) error {
        if _, err := doSessionHeader(ctx, w, r, func(ctx context.Context, span trace.Span, session *sessions.Session) error {
            err, sc := canManageAccounts(session, "")
            if err != nil {
                httpError(ctx, span, w, err.Error(), sc)
            }

            return err }); err != nil {
            return err
        }

        span := trace.SpanFromContext(ctx)

        if _, err := fmt.Fprintf(w, "Users (List)\n"); err != nil {
            httpError(ctx, span, w, err.Error(), http.StatusInternalServerError)
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
                httpError(ctx, span, w, err.Error(), http.StatusInternalServerError)
            }

            return err
        })
    })
}

func handlerUsersCreate(w http.ResponseWriter, r *http.Request) {
    _ = tr.WithSpan(context.Background(), methodName(), func(ctx context.Context) (err error) {
        var sc int

        vars := mux.Vars(r)
        username := vars["username"]

        _, err = doSessionHeader(ctx, w, r, func(ctx context.Context, span trace.Span, session *sessions.Session) error {
            err, sc = canManageAccounts(session, "")
            return err })

        span := trace.SpanFromContext(ctx)
        if err != nil {
            httpError(ctx, span, w, err.Error(), sc)
            return err
        }

        span.AddEvent(ctx, fmt.Sprintf("Creating user %q", username))

        u := &pb.UserDefinition{}
        if err = jsonpb.Unmarshal(r.Body, u); err != nil {
            httpError(ctx, span, w, err.Error(), http.StatusBadRequest)
            return err
        }

        if err = UserAdd(username, u.Password, u.AccountManager, u.Enabled); err != nil {
            httpError(ctx, span, w, err.Error(), http.StatusBadRequest)
            return err
        }

        span.AddEvent(ctx, fmt.Sprintf("Created user %q, pwd: %q, enabled: %v, accountManager: %v", username, u.Password, u.Enabled, u.AccountManager))
        _, err = fmt.Fprintf(w, "User %q created.  enabled: %v, can manage accounts: %v", username, u.Enabled, u.AccountManager)
        return err
    })
}

func handlerUsersRead(w http.ResponseWriter, r *http.Request) {
    _ = tr.WithSpan(context.Background(), methodName(), func(ctx context.Context) (err error) {
        var sc int

        vars := mux.Vars(r)
        username := vars["username"]

        _, err = doSessionHeader(ctx, w, r, func(ctx context.Context, span trace.Span, session *sessions.Session) error {
            err, sc = canManageAccounts(session, username)

            w.Header().Set("Content-Type", "application/json")

            return err })

        span := trace.SpanFromContext(ctx)
        if err != nil {
            httpError(ctx, span, w, err.Error(), sc)
            return err
        }

        u, _, err := dbUsers.Get(username)
        if err != nil {
            httpError(ctx, span, w, err.Error(), http.StatusBadRequest)
            return err
        }

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

// TBD
func handlerUsersUpdate(w http.ResponseWriter, r *http.Request) {
    _ = tr.WithSpan(context.Background(), methodName(), func(ctx context.Context) (err error) {
        _, err = doSessionHeader(ctx, w, r, func(ctx context.Context, span trace.Span, session *sessions.Session) error {
            span.AddEvent(ctx, "TBD: user update")
            return usersDisplayArguments(w, r)
        })

        return err
    })
}

// TBD
func handlerUsersDelete(w http.ResponseWriter, r *http.Request) {
    _ = tr.WithSpan(context.Background(), methodName(), func(ctx context.Context) (err error) {
        _, err = doSessionHeader(ctx, w, r, func(ctx context.Context, span trace.Span, session *sessions.Session) error {
            span.AddEvent(ctx, "TBD: user deletion")
            return usersDisplayArguments(w, r)
        })

        return err
    })
}

func handlerUsersOperation(w http.ResponseWriter, r *http.Request) {
    _ = tr.WithSpan(context.Background(), methodName(), func(ctx context.Context) (err error) {
        var s string
        var sc int

        _, err = doSessionHeader(ctx, w, r, func(ctx context.Context, span trace.Span, session *sessions.Session) error {
            op := r.FormValue("op")
            vars := mux.Vars(r)
            username := vars["username"]

            span.AddEvent(ctx, fmt.Sprintf("Operation %q, user %q, session %v", op, username, session))

            switch op {
            case Enable:
                s, err, sc = enableDisable(session, r, true, "enable")

            case Disable:
                s, err, sc = enableDisable(session, r, false, "disable")

            case Login:
                s, err, sc = login(session, r)

            case Logout:
                s, err, sc = logout(session, r)

            default:
                err = fmt.Errorf("invalid user operation requested (?op=%s)", op)
                sc = http.StatusBadRequest
            }

            return err
        })

        if err != nil {
            span := trace.SpanFromContext(ctx)
            httpError(ctx, span, w, err.Error(), sc)
        }

        if err == nil {
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
func enableDisable(session *sessions.Session, r *http.Request, v bool, s string) (string, error, int) {
    vars := mux.Vars(r)
    username := vars["username"]

    if err, sc := canManageAccounts(session, username); err != nil {
        return "", err, sc
    }

    if err := userEnable(username, v); err != nil {
        return "", err, http.StatusBadRequest
    }

    return fmt.Sprintf("User %q %sd", username, s), nil, http.StatusOK
}

// Process a login request (?op=login)
func login(session *sessions.Session, r *http.Request) (_ string, err error, sc int) {
    var pwd []byte

    vars := mux.Vars(r)
    username := vars["username"]

    // Verify that there is no logged in user
    if auth, ok := session.Values[AuthStateKey].(bool); ok && auth {
        return "", ErrUserAlreadyLoggedIn, http.StatusBadRequest
    }

    // We have a session that could support a login.  Let's verify that this
    // is a good login

    // First, verify that this is an actual user account, and that account is
    // enabled for login operations.
    if u, _, err := dbUsers.Get(username); err != nil || !u.Enabled {
        return "", ErrUserAuthFailed, http.StatusNotFound
    }

    // .. next, let's get the password, which is the body of the request
    if pwd, err = ioutil.ReadAll(r.Body); err != nil {
        return "", ErrUserAuthFailed, http.StatusBadRequest
    }

    // .. finally, let's confirm that this password matches the one for the
    // designated user account.
    if userVerifyPassword(username, pwd) != nil {
        return "", ErrUserAuthFailed, http.StatusForbidden
    }

    // .. all passed.  So finally mark the session as logged in
    //
    session.Values[AuthStateKey] = true
    session.Values[UserNameKey] = username

    return fmt.Sprintf("User %q logged in", username), nil, http.StatusOK
}

// Process a logout request (?op=logout)
func logout(session *sessions.Session, r *http.Request) (_ string, err error, sc int) {
    vars := mux.Vars(r)
    username := vars["username"]

    // Verify that there is a logged in user on this session
    if auth, ok := session.Values[AuthStateKey].(bool); !ok || !auth {
        return "", ErrNoLoginActive(username), http.StatusBadRequest
    }

    // .. and that it is the user we're trying to logout
    if name, ok := session.Values[UserNameKey].(string); !ok || !strings.EqualFold(name, username) {
        return "", ErrNoLoginActive(username), http.StatusBadRequest
    }

    // .. and now log the user out
    session.Values[AuthStateKey] = false
    delete(session.Values, UserNameKey)

    return fmt.Sprintf("User %q logged out", username), nil, http.StatusOK
}

// --- Query tag implementations

// This is a temporary method used for tracing prior to actual implementations.  Will be removed.
func usersDisplayArguments(w http.ResponseWriter, r *http.Request) (err error) {
    op := r.FormValue("op")
    vars := mux.Vars(r)
    username := vars["username"]

    if op != "" {
        _, err = fmt.Fprintf(w, "User: %v op: %v", username, op)
    } else {
        _, err = fmt.Fprintf(w, "User: %v", username)
    }

    return err
}

// +++ Mid-level support methods

// This section contains the methods that translate from the logical user
// attributes that are understood by the route handlers to the internal user
// attributes understood by the storage system.

func UserAdd(name string, password string, accountManager bool, enabled bool) error {

    passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

    if err != nil {
        return err
    }

    return dbUsers.Create(&pb.User{
        Name: name,
        PasswordHash: passwordHash,
        Enabled: enabled,
        AccountManager: accountManager})
}

func userRemove(name string) error {
    return dbUsers.Remove(name)
}

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
    return dbUsers.Update(entry, rev)
}

// --- Mid-level support methods

// +++ Helper functions

// Set an http error, and log it to the tracing system.
func httpError(ctx context.Context, span trace.Span, w http.ResponseWriter, msg string, sc int) {
    span.AddEvent(ctx, fmt.Sprintf("http error %v: %s", sc, msg))
    http.Error(w, msg, sc)
}

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
func canManageAccounts(session *sessions.Session, username string) (error, int) {
    key, ok := session.Values[UserNameKey].(string)
    if !ok {
        return http.ErrNoCookie, http.StatusBadRequest
    }

    user, _, err := dbUsers.Get(key)
    if err != nil {
        return ErrUserPermissionDenied, http.StatusForbidden
    }

    if !user.AccountManager && !strings.EqualFold(user.Name, username) {
        return ErrUserPermissionDenied, http.StatusForbidden
    }

    return nil, http.StatusOK
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
