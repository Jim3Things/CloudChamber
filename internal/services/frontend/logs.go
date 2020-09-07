package frontend

import (
	"context"
	"net/http"

	"github.com/golang/protobuf/jsonpb"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	tsc "github.com/Jim3Things/CloudChamber/internal/clients/trace_sink"
	st "github.com/Jim3Things/CloudChamber/internal/tracing/server"
)

func logsAddRoutes(routeBase *mux.Router) {
	router := routeBase.PathPrefix("/logs").Subrouter()

	router.HandleFunc("", handlerLogsGetAfter).Queries("from", "{from}", "for", "{for}").Methods("GET")
	router.HandleFunc("/policy", handlerLogsGetPolicy).Methods("GET")
}

// handlerLogsGetAfter processes an incoming REST request to retrieve trace
// entries within the specified range.
func handlerLogsGetAfter(w http.ResponseWriter, r *http.Request) {
	_ = st.WithInfraSpan(context.Background(), func(ctx context.Context) error {

		vars := mux.Vars(r)
		from := vars["from"]
		count := vars["for"]

		err := doSessionHeader(
			ctx, w, r,
			func(ctx context.Context, session *sessions.Session) error {
				return ensureEstablishedSession(session)
			})

		if err != nil {
			return httpError(ctx, w, err)
		}

		fromId, err := ensureNumber("from", from)
		if err != nil {
			return httpError(ctx, w, err)
		}

		maxSize, err := ensurePositiveNumber("for", count)
		if err != nil {
			return httpError(ctx, w, err)
		}

		ch := tsc.GetTraces(ctx, fromId, maxSize)

		data := <-ch
		if data.Err != nil {
			return httpError(ctx, w, data.Err)
		}

		w.Header().Set("Content-Type", "application/json")

		p := jsonpb.Marshaler{}
		return p.Marshal(w, data.Traces)
	})
}

// handlerLogsGetPolicy processes an incoming REST request to obtain the
// current trace_sink policy.
func handlerLogsGetPolicy(w http.ResponseWriter, r *http.Request) {
	_ = st.WithInfraSpan(context.Background(), func(ctx context.Context) error {

		err := doSessionHeader(
			ctx, w, r,
			func(ctx context.Context, session *sessions.Session) error {
				return ensureEstablishedSession(session)
			})

		if err != nil {
			return httpError(ctx, w, err)
		}

		policy, err := tsc.GetPolicy(ctx)
		if err != nil {
			return httpError(ctx, w, err)
		}

		w.Header().Set("Content-Type", "application/json")

		p := jsonpb.Marshaler{}
		return p.Marshal(w, policy)
	})
}
