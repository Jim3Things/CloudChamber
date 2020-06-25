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
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"google.golang.org/grpc"

	ts "github.com/Jim3Things/CloudChamber/internal/clients/timestamp"
	"github.com/Jim3Things/CloudChamber/internal/config"
	ctrc "github.com/Jim3Things/CloudChamber/internal/tracing/client"
)

// Server is the context structure for the frontend web service. It is used to
// provide a convenient place to store all the long-lived server/service global data fields.
//
type Server struct {
	port         int
	rootFilePath string

	handler     http.Handler
	cookieStore *sessions.CookieStore
}

var (
	// key must be 16, 24 or 32 bytes long (AES-128, AES-192 or AES-256)
	keyAuthentication = securecookie.GenerateRandomKey(32)
	keyEncryption     = securecookie.GenerateRandomKey(32)

	server Server
)

// normalize the URLs we process before handing them off to the normal route
// processing.
//
// The rules for processing URLs are:
//	a) any POST operation is lowercased, except for the last segment.  That
//	   segment has the case retained.
//
//	b) any other operation is lowercased fully.
//
// This allows user casing choice for values they determine, such as user
// names, while also allowing for case insensitive processing for all
// internal Cloud Chamber components, as well as case insensitive lookup
// for the user values.
//
func normalizeURL(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			// A logical create operation. Retain the case of the last segment,
			// such as the user name.  Lowercase everything else.
			//
			path := r.URL.Path
			toLower := path[:strings.LastIndex(path, "/")]
			r.URL.Path = strings.ToLower(toLower) + path[strings.LastIndex(path, "/"):]
		} else {
			// An action against an existing object, so lowercase everything.
			//
			r.URL.Path = strings.ToLower(r.URL.Path)
		}

		// Now invoke the actual handler with the modified URL.
		next.ServeHTTP(w, r)
	})
}

func initHandlers() error {

	handler := mux.NewRouter()
	filesAddRoutes(server.rootFilePath, handler)

	routeAPI := handler.PathPrefix("/api").Subrouter()

	// Now add the routes for the API
	//
	usersAddRoutes(routeAPI)
	workloadsAddRoutes(routeAPI)
	inventoryAddRoutes(routeAPI)
	stepperAddRoutes(routeAPI)

	// TODO the following handler definitions are just temporary placeholders and
	// should at some point be converted to follow the same pattern as for files,
	// users and workloads, namely moved to a separate file and defined/handler
	// there.
	//
	routeAPI.HandleFunc("/logs", handlerLogsRoot).Methods("GET")
	routeAPI.HandleFunc("/injector", handlerInjectorRoot).Methods("GET")

	server.handler = normalizeURL(handler)

	return nil
}

func initService(cfg *config.GlobalConfig) error {

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

	server.rootFilePath = cfg.WebServer.RootFilePath
	server.port = cfg.WebServer.FE.Port
	server.cookieStore = sessions.NewCookieStore(keyAuthentication, keyEncryption)

	// TODO: These are here only because we've not gotten https working yet.  Once it is, these need to be removed.
	server.cookieStore.Options.Secure = false
	server.cookieStore.Options.HttpOnly = false

	// Initialize the simulated time (stepper) service client
	ts.InitTimestamp(
		fmt.Sprintf(
			"%s:%d",
			cfg.SimSupport.EP.Hostname,
			cfg.SimSupport.EP.Port),
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(ctrc.Interceptor))

	if err := initHandlers(); err != nil {
		return err
	}

	if err := InitDBInventory(); err != nil {
		return err
	}

	// Finally, initialize the user store
	return InitDBUsers(cfg)
}

// StartService is the primary entry point to start the front-end web service.
func StartService(cfg *config.GlobalConfig) error {

	if err := initService(cfg); err != nil {
		log.Fatalf("Error initializing service: %v", err)
	}

	return http.ListenAndServe(
		fmt.Sprintf(":%d", server.port),
		server.handler)
}

// func handlerRoot(w http.ResponseWriter, r *http.Request) {
//
//   fmt.Fprintf(w, "Cloudchamber")
// }

func handlerLogsRoot(w http.ResponseWriter, r *http.Request) {

	fmt.Fprintf(w, "Logs (Root)")
}

func handlerInjectorRoot(w http.ResponseWriter, r *http.Request) {

	fmt.Fprintf(w, "Injector (Root)")
}
