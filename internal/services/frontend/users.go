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
    "go.opentelemetry.io/otel/api/trace"

    "golang.org/x/crypto/bcrypt"
)

const (
    Enable = "enable"
    Disable = "disable"
    Login = "login"
    Logout = "logout"
)

// User is a representation of an individual user
//
// TODO This is just a placeholder until we have proper user records held
//     in a persisted store (Etcd)
//
type User struct {
    Name         string
    PasswordHash []byte
    //  UserId   int64
    Enabled bool
}

// DbUsers is a container used to established synchronized access to
// the in-memory set of user records.
//
type DbUsers struct {
    Mutex sync.Mutex
    Users map[string]User
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
    session, _ := server.cookieStore.Get(r, "cookie-name")

    // Check if user is authenticated
    if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }

    // Print secret message
    fmt.Fprintln(w, "secret message")
}

func userCreate(name string, password []byte) (*User, error) {

    passwordHash, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)

    if err != nil {
        return nil, err
    }

    user := &User{Name: name, PasswordHash: passwordHash}

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
        return ErrUserAlreadyCreated
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
        return ErrUserNotFound
    }

    user.Enabled = enable
    return nil
}

func usersAddRoutes(routeBase *mux.Router) {

    const routeString = "/{username:[a-z,A-Z][a-z,A-Z,0-9]*}"

    dbUsers = DbUsers{
        Mutex: sync.Mutex{},
        Users: map[string]User{},
    }

    routeUsers := routeBase.PathPrefix("/users").Subrouter()

    routeUsers.HandleFunc("", handlerUsersList).Methods("GET")
    routeUsers.HandleFunc("/", handlerUsersList).Methods("GET")

    // In the following, the "GET" method is allowed just for the purposes of test and
    // evaluation. At somepoint, it will need to be removed, but in the meantime, leaving
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

func usersDisplayArguments(w http.ResponseWriter, r *http.Request) {
    op := r.FormValue("op")
    vars := mux.Vars(r)
    username := vars["username"]

    if op != "" {
        fmt.Fprintf(w, "User: %v op: %v", username, op)
    } else {
        fmt.Fprintf(w, "User: %v", username)
    }
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

    usersDisplayArguments(w, r)
}

func handlerUsersRead(w http.ResponseWriter, r *http.Request) {

    usersDisplayArguments(w, r)
}

func handlerUsersUpdate(w http.ResponseWriter, r *http.Request) {

    ctx, span := tr.Start(context.Background(), methodName())
    defer span.End()

    span.AddEvent(ctx, "Update hit")
    usersDisplayArguments(w, r)
}

func handlerUsersDelete(w http.ResponseWriter, r *http.Request) {

    usersDisplayArguments(w, r)
}

func handlerUsersOperation(w http.ResponseWriter, r *http.Request) {

    _ = tr.WithSpan(context.Background(), methodName(), func(ctx context.Context) error {
        span := trace.SpanFromContext(ctx)

        op := r.FormValue("op")
        vars := mux.Vars(r)
        username := vars["username"]

        span.AddEvent(ctx, fmt.Sprintf("Operation '%s', user '%s'", op, username))

        switch op {
        case Enable:
            usersDisplayArguments(w, r)

        case Disable:
            usersDisplayArguments(w, r)

        case Login:
            login(ctx, span, w, r)

        case Logout:
            usersDisplayArguments(w, r)

        default:
            httpError(ctx, span, w, fmt.Sprintf("Invalid user operation requested (?op=%s)", op), http.StatusBadRequest)
        }

        return nil
    })
}

// Process a login request
func login(ctx context.Context, span trace.Span, w http.ResponseWriter, r *http.Request) {
    session, _ := server.cookieStore.Get(r, "cookie-name")
    if auth, ok := session.Values["authenticated"].(bool); ok && auth {
        http.Error(w, "Session already established", http.StatusLocked)
        return
    }

    // Authentication goes here
    // ...

    // Set user as authenticated
    //
    session.Values["authenticated"] = true

    if err := session.Save(r, w); err != nil {
        httpError(ctx, span, w, err.Error(), http.StatusInternalServerError)
    }
}

func logout(w http.ResponseWriter, r *http.Request) {
    session, _ := server.cookieStore.Get(r, "cookie-name")

    // Revoke users authentication
    session.Values["authenticated"] = false
    session.Save(r, w)
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
