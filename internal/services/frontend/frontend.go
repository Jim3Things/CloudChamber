// This module contains the implementation of the service front end which establishes the http[s] endpoint to receive and handle incomging requests.
//

// Useful references
//
// https://golang.org/doc/articles/wiki/
// https://tutorialedge.net/golang/creating-simple-web-server-with-golang/
// https://gist.github.com/enricofoltran/10b4a980cd07cb02836f70a4ab3e72d7
// https://github.com/icza/session/blob/master/_session_demo/session_demo.go
// https://www.alexedwards.net/blog/serving-static-sites-with-go
// https://gowebexamples.com/sessions/
// https://gowebexamples.com/http-server/
// https://gowebexamples.com/routes-using-gorilla-mux/
//

package frontend

import (
    "errors"
    "flag"
    "fmt"
    "log"
    "net/http"
    "sync"

    "github.com/gorilla/mux"
    "github.com/gorilla/securecookie"
    "github.com/gorilla/sessions"
)

// Computer is a represention an individual Computer
//
//TODO This is just a placeholder until we have proper inventory items
//     held in a persisted store (Etcd)
//
type Computer struct {
    Name       string
    Processors uint32
    Memory     uint64
}

// DbComputers is a container used to established synchronized access to
// the in-memory set of inventory records.
//
//TODO This is just a placeholder until we have proper inventory items
//     held in the persisted store (Etcd)
//
type DbComputers struct {
    Lock      sync.Mutex
    Computers map[string]Computer
}

// Server is the context structure for the frontend web service. It is used to
// provide a convenient place to store all the long-lived server/service global data fields.
//
type Server struct {
    port         int
    rootFilePath string

    handler     *mux.Router
    cookieStore *sessions.CookieStore
}

const (
    defaultPort     = 8080
    defaultRootFilePath = "C:\\Chamber"
)

var (
    // key must be 16, 24 or 32 bytes long (AES-128, AES-192 or AES-256)
    keyAuthentication = securecookie.GenerateRandomKey(64)
    keyEncryption     = securecookie.GenerateRandomKey(64)

    store *sessions.CookieStore

    // ErrNotInitialized is a new error to indicate initialization failures.
    ErrNotInitialized = errors.New("CloudChamber: initialization failure")

    // ErrWorkloadNotEnabled indicates the specified workload is not enabled
    // for the purposes of deployment or execution.
    ErrWorkloadNotEnabled = errors.New("CloudChamber: workload not enabled")

    // ErrUserUnableToCreate indicates the specified user account cannot be
    // created at this time
    //
    ErrUserUnableToCreate = errors.New("CloudChamber: unable to create a user account at this time")

    // ErrUserAlreadyCreated indicates the specified user account was previously
    // created and the request was determined to be a duplicate Create request.
    //
    ErrUserAlreadyCreated = errors.New("CloudChamber: user already exists")

    // ErrUserNotFound indicates the specified user account was determined to
    // not exist (i.e. the search succeeded but no record was found)
    //
    ErrUserNotFound = errors.New("CloudChamber: user not found")

    // ErrUserAuthFailed indicates the supplied username and password combination
    // is not valid.
    //
    ErrUserAuthFailed = errors.New("CloudChamber: authentication failed, invalid user name or password")

    // ErrUserPermissionDenied indicates the user does not have the appropriate
    // permissions for the requested operation.
    //
    ErrUserPermissionDenied = errors.New("CloudChamber: permission denied")

    // ErrUserInvalidOperation indicates the operation requested for the supplied
    // user account is invalid in some way, likely a non-existent operation code.
    //
    ErrUserInvalidOperation = errors.New("CloudChamber: invalid operation")

    dbComputers DbComputers

    server Server
)

func initHandlers() error {

    server.handler = mux.NewRouter()

    routeAPI := server.handler.PathPrefix("/api").Subrouter()

    // Now add the routes for the API
    //
    filesAddRoutes(server.rootFilePath, server.handler)
    usersAddRoutes(routeAPI)
    workloadsAddRoutes(routeAPI)

    // TODO the following handler definitions are just temporary placeholders and 
    // should at some point be converted to follow the same pattern as for files,
    // users and workloads, namely moved to a separate file and defined/handler
    // there.
    //
    routeAPI.HandleFunc("/logs", handlerLogsRoot).Methods("GET")
    routeAPI.HandleFunc("/stepper", handlerStepperRoot).Methods("GET")
    routeAPI.HandleFunc("/injector", handlerInjectorRoot).Methods("GET")

    return nil
}

func initArguments() error {

    if !flag.Parsed() {
        port := flag.Int(
            "port",
            defaultPort,
            "port used by the web service")

        rootFilePath := flag.String(
            "path",
            defaultRootFilePath,
            "directory path holding the cloud chamber web service data")

        flag.Parse()

        server.rootFilePath = *rootFilePath
        server.port = *port
    }

    return nil
}

func initService() error {

    // A failure to generate a random key is most likely a result of a failure of the
    // system supplied random number generator mechanism. Although not known for sure
    // at this point, this is likely a result of the system not yet having been running
    // long enough to gather enough random timing events (aka entropy) to allow for a
    // sufficiently random number to be properly generated. If this is the underlying
    // cause, a suitable delay (say 60 seconds) and then restarting the service will
    // likely resolve the issue.
    //
    if nil == keyAuthentication {
        log.Fatalf("Failed to generate required authentication key (Check system Random Number Generator and restart the service after 60s). Error: %v", ErrNotInitialized)
    } else if nil == keyEncryption {
        log.Fatalf("Failed to generate required encryption key (Check system Random Number Generator and restart the service after 60s). Error: %v", ErrNotInitialized)
    }

    server.cookieStore = sessions.NewCookieStore(keyAuthentication, keyEncryption)

    if err := initArguments(); err != nil {
        log.Fatalf("Error initializing arguments: %v", err)
    }

    if err := initHandlers(); err != nil {
        log.Fatalf("Error initializing handlers: %v", err)
    }

    return nil
}

// StartService is the primary entry point to start the front-end web service.
func StartService() error {

    if err := initService(); err != nil {
        log.Fatalf("Error initializing service: %v", err)
    }

    http.ListenAndServe(fmt.Sprintf(":%d", server.port), server.handler)

    return nil
}

//func handlerRoot(w http.ResponseWriter, r *http.Request) {
//
//  fmt.Fprintf(w, "Cloudchamber")
//}

func handlerLogsRoot(w http.ResponseWriter, r *http.Request) {

    fmt.Fprintf(w, "Logs (Root)")
}

func handlerStepperRoot(w http.ResponseWriter, r *http.Request) {

    fmt.Fprintf(w, "Stepper (Root)")
}

func handlerInjectorRoot(w http.ResponseWriter, r *http.Request) {

    fmt.Fprintf(w, "Injector (Root)")
}
