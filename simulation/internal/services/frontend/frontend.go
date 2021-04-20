// This module contains the implementation of the service front end which
// establishes the http[s] endpoint to receive and handle incoming requests.
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
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"google.golang.org/grpc"

	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/limits"
	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/store"
	ts "github.com/Jim3Things/CloudChamber/simulation/internal/clients/timestamp"
	tsc "github.com/Jim3Things/CloudChamber/simulation/internal/clients/trace_sink"
	"github.com/Jim3Things/CloudChamber/simulation/internal/config"
	"github.com/Jim3Things/CloudChamber/simulation/internal/tracing"
	ct "github.com/Jim3Things/CloudChamber/simulation/internal/tracing/client"
	"github.com/Jim3Things/CloudChamber/simulation/pkg/errors"
)

// Server is the context structure for the frontend web service. It is used to
// provide a convenient place to store all the long-lived server/service global
// data fields.
//
type Server struct {
	port         uint16
	rootFilePath string

	handler     http.Handler
	cookieStore *sessions.CookieStore

	startTime time.Time

	sessions managedSessions

	watchTimeout time.Duration
}

const (
	maxWatchTimeout = 5
)

var (
	// key must be 16, 24 or 32 bytes long (AES-128, AES-192 or AES-256)
	keyAuthentication = securecookie.GenerateRandomKey(32)
	keyEncryption     = securecookie.GenerateRandomKey(32)

	server Server
)

// normalizeURL standardizes the URL string before handing them to the normal
// route processing.
//
// The rules for processing URLs are:
//	a) any POST operation is lower-cased, except for the last segment.  That
//	   segment has the case retained.
//
//	b) any other operation is lower-cased fully.
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

// traceRequest is an interception handler that traces the request as it
// arrives.  This is used, for example, to trace file access requests.
func traceRequest(spanName string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, span := tracing.StartSpan(context.Background(),
			tracing.WithName(fmt.Sprintf("%s: %s %q", spanName, r.Method, r.URL.String())),
			tracing.AsInternal(),
			tracing.WithContextValue(ts.OutsideTime))
		defer span.End()

		next.ServeHTTP(w, r)
	})
}

func initHandlers() error {

	handler := mux.NewRouter()

	// Add the routes for the API
	//
	routeAPI := handler.PathPrefix("/api").Subrouter()
	injectionAddRoutes(routeAPI)
	inventoryAddRoutes(routeAPI)
	logsAddRoutes(routeAPI)
	simulationAddRoutes(routeAPI)
	stepperAddRoutes(routeAPI)
	usersAddRoutes(routeAPI)
	watchAddRoutes(routeAPI)
	workloadsAddRoutes(routeAPI)

	// Add the file handling for any other paths.
	handler.PathPrefix("/").
		Handler(
			traceRequest(
				"File Request",
				http.StripPrefix(
					"/",
					http.FileServer(http.Dir(server.rootFilePath)))))

	server.handler = normalizeURL(handler)

	return nil
}

// initClients sets up the internal service clients used by the frontend
// handlers themselves.
func initClients(cfg *config.GlobalConfig) error {
	err := ts.InitTimestamp(
		cfg.SimSupport.EP.String(),
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(ct.Interceptor),
		grpc.WithConnectParams(limits.BackoffSettings),
	)

	if err != nil {
		return err
	}

	err = tsc.InitSinkClient(
		cfg.SimSupport.EP.String(),
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(ct.Interceptor),
		grpc.WithConnectParams(limits.BackoffSettings),
	)

	return err
}

func initService(cfg *config.GlobalConfig) error {
	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithName("Initialize web service"),
		tracing.WithContextValue(ts.EnsureTickInContext),
		tracing.AsInternal())
	defer span.End()

	// A failure to generate a random key is most likely a result of a failure of the
	// system supplied random number generator mechanism. Although not known for sure
	// at this point, this is likely a result of the system not yet having been running
	// long enough to gather enough random timing events (aka entropy) to allow for a
	// sufficiently random number to be properly generated. If this is the underlying
	// cause, a suitable delay (say 60 seconds) and then restarting the service will
	// likely resolve the issue.
	//
	if nil == keyAuthentication {
		log.Fatalf(
			"Failed to generate required authentication key (Check system "+
				"Random Number Generator and restart the service after 60s). Error: %v",
			errors.ErrNotInitialized)
	} else if nil == keyEncryption {
		log.Fatalf(
			"Failed to generate required encryption key (Check system Random "+
				"Number Generator and restart the service after 60s). Error: %v",
			errors.ErrNotInitialized)
	}

	server.rootFilePath = cfg.WebServer.RootFilePath
	server.port = cfg.WebServer.FE.Port
	server.cookieStore = sessions.NewCookieStore(keyAuthentication, keyEncryption)
	server.cookieStore.Options.SameSite = http.SameSiteStrictMode

	// TODO: These are here only because we've not gotten https working yet.
	//       Once it is, these need to be removed.
	server.cookieStore.Options.Secure = false
	server.cookieStore.Options.HttpOnly = false

	server.sessions = newSessionTable(
		cfg.WebServer.ActiveSessionLimit,
		time.Duration(cfg.WebServer.SessionInactivity)*time.Second)

	server.watchTimeout = maxWatchTimeout * time.Second
	if cfg.WebServer.SessionInactivity < (maxWatchTimeout * 2) {
		server.watchTimeout = time.Duration(cfg.WebServer.SessionInactivity/2) * time.Second
	}

	server.startTime = time.Now()

	if err := initHandlers(); err != nil {
		return err
	}

	// Initialize the underlying store
	//
	store.Initialize(ctx, cfg)

	// initialize the inventory store and apply any updates from the configuration.
	//
	if err := InitDBInventory(ctx, cfg, store.NewStore()); err != nil {
		return err
	}

	// Finally, initialize the user store
	return InitDBUsers(ctx, cfg)
}

// StartService is the primary entry point to start the front-end web service.
func StartService(cfg *config.GlobalConfig) error {
	if err := initClients(cfg); err != nil {
		log.Fatalf("Error initializing local clients: %v", err)
	}

	if err := initService(cfg); err != nil {
		log.Fatalf("Error initializing service: %v", err)
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", server.port),
		Handler: server.handler,

		ErrorLog: log.New(
			&WriterToLog{},
			"HTTP Listener",
			log.LstdFlags|log.Lshortfile),
	}

	return srv.ListenAndServe()
}

// WriterToLog acts as a transducer between the http server's logging functions
// and the CloudChamber tracing support.  The http server expects an io.Writer,
// which this structure provides.
type WriterToLog struct {
}

// Write is the io writer compliant function that writes the incoming message
// into the trace log as an informational message in its own minimal span.
func (w WriterToLog) Write(p []byte) (n int, err error) {
	ctx, span := tracing.StartSpan(
		context.Background(),
		tracing.WithName("HTTP intercept event"),
		tracing.WithNewRoot(),
		tracing.WithContextValue(ts.OutsideTime))
	defer span.End()

	tracing.Info(ctx, string(p))
	return fmt.Fprintln(os.Stderr, string(p))
}
