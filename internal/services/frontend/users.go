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

// DbUsers is a container used to established synchronized access to
// the in-memory set of user records.
//
type DbUsers struct {
    Mutex sync.Mutex
    Users map[string]pb.User
}

var (
    dbUsers DbUsers
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

func userCreate(name string, password []byte) (*pb.User, error) {

    passwordHash, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)

    if err != nil {
        return nil, err
    }

    user := &pb.User{Name: name, PasswordHash: passwordHash}

    return user, nil
}

func UserAdd(name string, password []byte) error {

    newUser, err := userCreate(name, password)

    if nil != err {
        return err
    }

    dbUsers.Mutex.Lock()
    defer dbUsers.Mutex.Unlock()

    _, found := dbUsers.Users[name]

    if found {
        return ErrUserAlreadyCreated(name)
    }

    dbUsers.Users[name] = *newUser
    return nil
}

func userRemove(name string) error {
    dbUsers.Mutex.Lock()
    delete(dbUsers.Users, name)
    dbUsers.Mutex.Unlock()
    return nil
}

func userVerifyPassword(name string, password []byte) error {

    dbUsers.Mutex.Lock()
    defer dbUsers.Mutex.Unlock()

    return bcrypt.CompareHashAndPassword(dbUsers.Users[name].PasswordHash, password)
}

func userEnable(name string, enable bool) error {

    dbUsers.Mutex.Lock()
    defer dbUsers.Mutex.Unlock()

    user, found := dbUsers.Users[name]
    if !found {
        return ErrUserNotFound(name)
    }

    user.Enabled = enable
    return nil
}

func usersAddRoutes(routeBase *mux.Router) {

    const routeString = "/{username:[a-z,A-Z][a-z,A-Z,0-9]*}"

    dbUsers = DbUsers{
        Mutex: sync.Mutex{},
        Users: map[string]pb.User{},
    }

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
        span := trace.SpanFromContext(ctx)

        _, err := fmt.Fprintf(w, "Users (List)\n")
        if err != nil {
            httpError(ctx, span, w, err.Error(), http.StatusInternalServerError)
            return err
        }

        b := r.URL.String()
        if !strings.HasSuffix(b, "/") {
            b += "/"
        }

        for _, user := range dbUsers.Users {
            target := fmt.Sprintf("%s%s", b, user.Name)

            span.AddEvent(ctx, fmt.Sprintf("   Listing user '%s' at '%s'", user.Name, target))

            _, err = fmt.Fprintln(w, target)
            if err != nil {
                httpError(ctx, span, w, err.Error(), http.StatusInternalServerError)
                return err
            }
        }

        return nil
    })
}

func handlerUsersCreate(w http.ResponseWriter, r *http.Request) {
    _ = tr.WithSpan(context.Background(), methodName(), func(ctx context.Context) (err error) {
        return WithSession(ctx, w, r, func(ctx context.Context, span trace.Span, session *sessions.Session) error {
            span.AddEvent(ctx, "TBD: user creation")
            return usersDisplayArguments(w, r)
        })
    })
}

func handlerUsersRead(w http.ResponseWriter, r *http.Request) {
    _ = tr.WithSpan(context.Background(), methodName(), func(ctx context.Context) (err error) {
        return WithSession(ctx, w, r, func(ctx context.Context, span trace.Span, session *sessions.Session) error {
            span.AddEvent(ctx, "TBD: user details read")
            return usersDisplayArguments(w, r)
        })
    })
}

func handlerUsersUpdate(w http.ResponseWriter, r *http.Request) {

    _ = tr.WithSpan(context.Background(), methodName(), func(ctx context.Context) (err error) {
        return WithSession(ctx, w, r, func(ctx context.Context, span trace.Span, session *sessions.Session) error {
            span.AddEvent(ctx, "TBD: user update")
            return usersDisplayArguments(w, r)
        })
    })
}

func handlerUsersDelete(w http.ResponseWriter, r *http.Request) {
    _ = tr.WithSpan(context.Background(), methodName(), func(ctx context.Context) (err error) {
        return WithSession(ctx, w, r, func(ctx context.Context, span trace.Span, session *sessions.Session) error {
            span.AddEvent(ctx, "TBD: user deletion")
            return usersDisplayArguments(w, r)
        })
    })
}

func handlerUsersOperation(w http.ResponseWriter, r *http.Request) {
    _ = tr.WithSpan(context.Background(), methodName(), func(ctx context.Context) (err error) {
        return WithSession(ctx, w, r, func(ctx context.Context, span trace.Span, session *sessions.Session) error {
            op := r.FormValue("op")
            vars := mux.Vars(r)
            username := vars["username"]

            span.AddEvent(ctx, fmt.Sprintf("Operation %q, user %q", op, username))

            switch op {
            case Enable:
                err = usersDisplayArguments(w, r)

            case Disable:
                err = usersDisplayArguments(w, r)

            case Login:
                err = login(ctx, span, session, w, r)

            case Logout:
                err = logout(ctx, span, session, w, r)

            default:
                err = fmt.Errorf("invalid user operation requested (?op=%s)", op)
            }

            return err
        })
    })
}

// Process a login request
func login(_ context.Context, _ trace.Span, session *sessions.Session, _ http.ResponseWriter, r *http.Request) error {
    vars := mux.Vars(r)
    username := vars["username"]

    // Verify that there is no logged in user
    if auth, ok := session.Values[AuthStateKey].(bool); ok && auth {
        return ErrUserAlreadyLoggedIn
    }

    // Authentication goes here
    // ...

    // .. and finally mark the session as logged in
    //
    session.Values[AuthStateKey] = true
    session.Values[UserNameKey] = username
    return nil
}

// Process a logout request
func logout(_ context.Context, _ trace.Span, session *sessions.Session, _ http.ResponseWriter, r *http.Request) error {
    vars := mux.Vars(r)
    username := vars["username"]

    // Verify that there is a logged in user on this session
    if auth, ok := session.Values[AuthStateKey].(bool); !ok || !auth {
        return ErrNoLoginActive(username)
    }

    // .. and that it is the user we're trying to logout
    if name, ok := session.Values[UserNameKey].(string); !ok || name != username {
        return ErrNoLoginActive(username)
    }

    // .. and now log the user out
    session.Values[AuthStateKey] = false
    delete(session.Values, UserNameKey)
    return nil
}

// func handlerAuthenticateSession(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
//     return func(w http.ResponseWriter, r *http.Request) {
//     m := validPath.FindStringSubmatch(r.URL.Path)
//     if m == nil {
//         http.NotFound(w, r)
//         return
//     }
//     fn(w, r, m[2])
//     }
// }
