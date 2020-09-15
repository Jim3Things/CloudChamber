package frontend

import (
	"context"
	"net/http"

	"github.com/golang/protobuf/jsonpb"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	clients "github.com/Jim3Things/CloudChamber/internal/clients/timestamp"
	tsc "github.com/Jim3Things/CloudChamber/internal/clients/trace_sink"
	"github.com/Jim3Things/CloudChamber/internal/tracing"
)

func logsAddRoutes(routeBase *mux.Router) {
	router := routeBase.PathPrefix("/logs").Subrouter()

	router.HandleFunc("", handlerLogsGetAfter).Queries("from", "{from}", "for", "{for}").Methods("GET")
	router.HandleFunc("/policy", handlerLogsGetPolicy).Methods("GET")
}

// handlerLogsGetAfter processes an incoming REST request to retrieve trace
// entries within the specified range.
func handlerLogsGetAfter(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithName("Get Traces After..."),
		tracing.WithContextValue(clients.EnsureTickInContext),
		tracing.AsInternal())
	defer span.End()

	vars := mux.Vars(r)
	from := vars["from"]
	count := vars["for"]

	err := doSessionHeader(
		ctx, w, r,
		func(ctx context.Context, session *sessions.Session) error {
			return ensureEstablishedSession(session)
		})

	if err != nil {
		postHttpError(ctx, w, err)
		return
	}

	fromId, err := ensureNumber("from", from)
	if err != nil {
		postHttpError(ctx, w, err)
		return
	}

	maxSize, err := ensurePositiveNumber("for", count)
	if err != nil {
		postHttpError(ctx, w, err)
		return
	}

	ch := tsc.GetTraces(ctx, fromId, maxSize)

	data := <-ch
	if data.Err != nil {
		postHttpError(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	p := jsonpb.Marshaler{}
	err = p.Marshal(w, data.Traces)

	httpErrorIf(ctx, w, err)
}

// handlerLogsGetPolicy processes an incoming REST request to obtain the
// current trace_sink policy.
func handlerLogsGetPolicy(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithName("Get Traces After..."),
		tracing.WithContextValue(clients.EnsureTickInContext),
		tracing.AsInternal())
	defer span.End()

	err := doSessionHeader(
		ctx, w, r,
		func(ctx context.Context, session *sessions.Session) error {
			return ensureEstablishedSession(session)
		})

	if err != nil {
		postHttpError(ctx, w, err)
		return
	}

	policy, err := tsc.GetPolicy(ctx)
	if err != nil {
		postHttpError(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	p := jsonpb.Marshaler{}
	err = p.Marshal(w, policy)

	httpErrorIf(ctx, w, err)
}
