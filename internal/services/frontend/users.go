// This module containes the routines etc to implement the frontend handlers for the users
// part of the API
//

package frontend

import (
    "context"
    "fmt"
    "net/http"
    "runtime"
    "strings"
    "sync"

    "github.com/golang/protobuf/jsonpb"
    "github.com/gorilla/mux"
    "github.com/gorilla/sessions"
    "go.opentelemetry.io/otel/api/trace"

    "golang.org/x/crypto/bcrypt"

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

// DbUsers is a container used to established synchronized access to
// the in-memory set of user records.
//
type DbUsers struct {
    Mutex sync.Mutex
    Users map[string]pb.UserInternal
}

func InitDbUsers() {
    if dbUsers == nil {
        dbUsers = &DbUsers{
            Mutex: sync.Mutex{},
            Users: make(map[string]pb.UserInternal),
        }
    }
}

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

// Helper functions

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

func userCreate(name string, password []byte, accountManager bool, enabled bool) (*pb.User, error) {

    passwordHash, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)

    if err != nil {
        return nil, err
    }

    user := &pb.User{
        Name: name,
        PasswordHash: passwordHash,
        Enabled: enabled,
        AccountManager: accountManager}

    return user, nil
}

func UserAdd(name string, password []byte, accountManager bool, enabled bool) error {

    newUser, err := userCreate(name, password, accountManager, enabled)

    if nil != err {
        return err
    }

    return dbUsers.Create(newUser)
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

func usersAddRoutes(routeBase *mux.Router) {

    const routeString = "/{username:[a-z,A-Z][a-z,A-Z,0-9]*}"

    InitDbUsers()

    routeUsers := routeBase.PathPrefix("/users").Subrouter()

    routeUsers.HandleFunc("", handlerUsersList).Methods("GET")
    routeUsers.HandleFunc("/", handlerUsersList).Methods("GET")

    // In the following, the "GET" method is allowed just for the purposes of test and
    // evaluation. At some point, it will need to be removed, but in the meantime, leaving
    // it there allows simple experimentation with just a browser.
    //
    // As a reminder,
    //   PUT is idempotent so translates to UPDATE in the CRUD methodolgy
    //   POST is NOT idempotent so translates to CREATE in the CRUD methodolgy
    //
    routeUsers.HandleFunc(routeString, handlerUsersOperation).Queries("op", "{op}").Methods("PUT")
    routeUsers.HandleFunc(routeString, handlerUsersCreate).Methods("POST")
    routeUsers.HandleFunc(routeString, handlerUsersRead).Methods("GET")
    routeUsers.HandleFunc(routeString, handlerUsersUpdate).Methods("PUT")
    routeUsers.HandleFunc(routeString, handlerUsersDelete).Methods("DELETE")
}

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

// Process an http request for the list of users.  Response should contain a document of links to the
// details URI for each known user.
func handlerUsersList(w http.ResponseWriter, r *http.Request) {
    _ = tr.WithSpan(context.Background(), methodName(), func(ctx context.Context) error {
        _, err := doSessionHeader(ctx, w, r, func(ctx context.Context, span trace.Span, session *sessions.Session) error {
            err, sc := canManageAccounts(session, "")
            if err != nil {
                httpError(ctx, span, w, err.Error(), sc)
            }

            return err
        })

        if err != nil {
            return err
        }

        span := trace.SpanFromContext(ctx)
        span.AddEvent(ctx, "Beginning user list")

        _, err = fmt.Fprintf(w, "Users (List)\n")
        if err != nil {
            httpError(ctx, span, w, err.Error(), http.StatusInternalServerError)
           return err
        }

        b := r.URL.String()
        if !strings.HasSuffix(b, "/") {
            b += "/"
        }

        err = dbUsers.Scan(func(entry *pb.User) error {
            target := fmt.Sprintf("%s%s", b, entry.Name)

            span.AddEvent(ctx, fmt.Sprintf("   Listing user '%s' at '%s'", entry.Name, target))

            _, err = fmt.Fprintln(w, target)
            if err != nil {
                httpError(ctx, span, w, err.Error(), http.StatusInternalServerError)
            }

            // return err
            return nil
        })

        span.AddEvent(ctx, fmt.Sprintf("Ending user list, err = %v", err))
        return err
    })
}

func handlerUsersCreate(w http.ResponseWriter, r *http.Request) {
    _ = tr.WithSpan(context.Background(), methodName(), func(ctx context.Context) (err error) {
        vars := mux.Vars(r)
        username := vars["username"]

        _, err = doSessionHeader(ctx, w, r, func(ctx context.Context, span trace.Span, session *sessions.Session) error {
            err, sc := canManageAccounts(session, "")
            if err != nil {
                httpError(ctx, span, w, err.Error(), sc)
            }

            return err
        })

        if err != nil {
            return err
        }

        span := trace.SpanFromContext(ctx)
        span.AddEvent(ctx, fmt.Sprintf("Creating user %q", username))

        u := &pb.UserDefinition{}
        err = jsonpb.Unmarshal(r.Body, u)
        if err != nil {
            httpError(ctx, span, w, err.Error(), http.StatusBadRequest)
            return err
        }

        if err = UserAdd(username, []byte(u.Password), u.AccountManager, u.Enabled); err != nil {
            httpError(ctx, span, w, err.Error(), http.StatusInternalServerError)
            return err
        }

        span.AddEvent(ctx, fmt.Sprintf("Created user %q, pwd: %q, enabled: %v, accountManager: %v", username, u.Password, u.Enabled, u.AccountManager))
        _, err = fmt.Fprintf(w, "User %q created.  enabled: %v, can manage accounts: %v", username, u.Enabled, u.AccountManager)
        return err
    })
}

func handlerUsersRead(w http.ResponseWriter, r *http.Request) {
    _ = tr.WithSpan(context.Background(), methodName(), func(ctx context.Context) (err error) {
        vars := mux.Vars(r)
        username := vars["username"]

        _, err = doSessionHeader(ctx, w, r, func(ctx context.Context, span trace.Span, session *sessions.Session) error {
            err, sc := canManageAccounts(session, username)
            if err != nil {
                httpError(ctx, span, w, err.Error(), sc)
                return err
            }

            w.Header().Set("Content-Type", "application/json")
            return nil
        })

        span := trace.SpanFromContext(ctx)
        if err != nil {
            httpError(ctx, span, w, err.Error(), http.StatusInternalServerError)
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

func handlerUsersUpdate(w http.ResponseWriter, r *http.Request) {

    _ = tr.WithSpan(context.Background(), methodName(), func(ctx context.Context) (err error) {
        _, err = doSessionHeader(ctx, w, r, func(ctx context.Context, span trace.Span, session *sessions.Session) error {
            span.AddEvent(ctx, "TBD: user update")
            return usersDisplayArguments(w, r)
        })

        return err
    })
}

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

        _, err = doSessionHeader(ctx, w, r, func(ctx context.Context, span trace.Span, session *sessions.Session) error {
            op := r.FormValue("op")
            vars := mux.Vars(r)
            username := vars["username"]

            span.AddEvent(ctx, fmt.Sprintf("Operation %q, user %q", op, username))

            switch op {
            case Enable:
                s, err = enable(ctx, span, session, w, r)

            case Disable:
                s, err = disable(ctx, span, session, w, r)

            case Login:
                s, err = login(ctx, span, session, w, r)

            case Logout:
                s, err = logout(ctx, span, session, w, r)

            default:
                err = fmt.Errorf("invalid user operation requested (?op=%s)", op)
            }

            return err
        })

        if err == nil {
            _, err = fmt.Fprintln(w, s)
        }
        return err
    })
}

// Process a request to enable the account
func enable(ctx context.Context, span trace.Span, session *sessions.Session, w http.ResponseWriter, r *http.Request) (string, error) {
    vars := mux.Vars(r)
    username := strings.ToLower(vars["username"])

    err, sc := canManageAccounts(session, username)
    if err != nil {
        httpError(ctx, span, w, err.Error(), sc)
        return "", err
    }

    err = userEnable(username, true)
    if err != nil {
        httpError(ctx, span, w, err.Error(), http.StatusBadRequest)
        return "", err
    }

    return fmt.Sprintf("User %q enabled", username), nil
}

// Process a request to disable the account
func disable(ctx context.Context, span trace.Span, session *sessions.Session, w http.ResponseWriter, r *http.Request) (string, error) {
    vars := mux.Vars(r)
    username := strings.ToLower(vars["username"])

    err, sc := canManageAccounts(session, username)
    if err != nil {
        httpError(ctx, span, w, err.Error(), sc)
        return "", err
    }

    err = userEnable(username, false)
    if err != nil {
        httpError(ctx, span, w, err.Error(), http.StatusBadRequest)
        return "", err
    }

    return fmt.Sprintf("User %q disabled", username), nil
}

// Process a login request
func login(_ context.Context, _ trace.Span, session *sessions.Session, _ http.ResponseWriter, r *http.Request) (string, error) {
    vars := mux.Vars(r)
    username := strings.ToLower(vars["username"])

    // Verify that there is no logged in user
    if auth, ok := session.Values[AuthStateKey].(bool); ok && auth {
        return "", ErrUserAlreadyLoggedIn
    }

    // Authentication goes here
    // ...
    u, _, err := dbUsers.Get(username)
    if err != nil || !u.Enabled {
        return "", ErrUserAuthFailed
    }

    // .. and finally mark the session as logged in
    //
    session.Values[AuthStateKey] = true
    session.Values[UserNameKey] = username

    return fmt.Sprintf("User %q logged in", username), nil
}

// Process a logout request
func logout(_ context.Context, _ trace.Span, session *sessions.Session, _ http.ResponseWriter, r *http.Request) (string, error) {
    vars := mux.Vars(r)
    username := strings.ToLower(vars["username"])

    // Verify that there is a logged in user on this session
    if auth, ok := session.Values[AuthStateKey].(bool); !ok || !auth {
        return "", ErrNoLoginActive(username)
    }

    // .. and that it is the user we're trying to logout
    if name, ok := session.Values[UserNameKey].(string); !ok || name != username {
        return "", ErrNoLoginActive(username)
    }

    // .. and now log the user out
    session.Values[AuthStateKey] = false
    delete(session.Values, UserNameKey)
    return fmt.Sprintf("User %q logged out", username), nil
}

func getSessionAccount(session *sessions.Session) (*pb.User, error) {
    key, ok := session.Values[UserNameKey].(string)
    if !ok {
        return nil, http.ErrNoCookie
    }

    user, _, err := dbUsers.Get(key)

    return user, err
}

func canManageAccounts(session *sessions.Session, username string) (error, int) {
    user, err := getSessionAccount(session)
    if err != nil {
        return err, http.StatusBadRequest
    }

    if !user.AccountManager && !strings.EqualFold(user.Name, username) {
        return ErrUserPermissionDenied, http.StatusForbidden
    }

    return nil, http.StatusOK
}
