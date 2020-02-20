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

//package frontend
package main

import (
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
)

// Computer is a represention an individual Computer
type Computer struct {
	Name       string
	Processors uint32
	Memory     uint64
}

type DbComputers struct {
	Lock      sync.Mutex
	Computers map[string]Computer
}

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

	ErrUserUnableToCreate   = errors.New("CloudChamber: unable to create a user account at this time")
	ErrUserAlreadyExists    = errors.New("CloudChamber: user already exists")
	ErrUserNotFound         = errors.New("CloudChamber: user not found")
	ErrUserAuthFailed       = errors.New("CloudChamber: authentication failed, invalid user name or password")
	ErrUserPermissionDenied = errors.New("CloudChamber: permission denied")

	dbComputers DbComputers

	rootFilePath = "ROOT_PATH"
)

// StartService is the primary entry point to start the front-end web service.
func StartService() error {

	// Really need some error handling here since continuation from this
	// point in the face of failure represents a compromise in security
	//
	if nil == keyAuthentication || nil == keyEncryption {
		return ErrNotInitialized
	}

	store = sessions.NewCookieStore(keyAuthentication, keyEncryption)

	routeBase := mux.NewRouter()

	routeAPI := routeBase.PathPrefix("/api").Subrouter()

	// Now add the routes for the API
	//
	filesAddRoutes(routeBase)
	usersAddRoutes(routeAPI)
	workloadsAddRoutes(routeAPI)

	routeAPI.HandleFunc("/logs", handlerLogsRoot).Methods("GET")
	routeAPI.HandleFunc("/stepper", handlerStepperRoot).Methods("GET")
	routeAPI.HandleFunc("/injector", handlerInjectorRoot).Methods("GET")

	http.ListenAndServe(":8080", routeBase)
	return nil
}

func main() {

	StartService()
}

//func handlerRoot(w http.ResponseWriter, r *http.Request) {
//
//	fmt.Fprintf(w, "Cloudchamber")
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
