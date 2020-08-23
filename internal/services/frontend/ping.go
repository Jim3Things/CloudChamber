// This module contains the functions necessary to implement the session
// refresh ping message support.

package frontend

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/Jim3Things/CloudChamber/internal/tracing"
	st "github.com/Jim3Things/CloudChamber/internal/tracing/server"
)

func pingAddRoutes(routeBase *mux.Router) {
	router := routeBase.PathPrefix("/ping").Subrouter()

	router.HandleFunc("", handlerPing).Methods("GET")
	router.HandleFunc("/", handlerPing).Methods("GET")
}

// handlerPing verifies that there is an active session, updating its expiry
// time in the process.  If successful, it returns an OK status with the new
// expiry time in the http header.
func handlerPing(w http.ResponseWriter, r *http.Request) {
	var ccSession SessionState

	_ = st.WithSpan(
		context.Background(),
		tracing.MethodName(1),
		func(ctx context.Context) error {
			err := doSessionHeader(
				ctx, w, r,
				func(_ context.Context, session *sessions.Session) error {
					// We get the cloud chamber session state.  We can ignore the ok
					// flag, as we only look at it if the next call succeeds, which
					// can only happen if there is a session...
					ccSession, _ = getSession(session)
					return ensureEstablishedSession(session)
				})
			if err != nil {
				return httpError(ctx, w, err)
			}

			w.Header().Set("Expires", ccSession.timeout.Format(time.RFC3339))
			return nil
		})
}
