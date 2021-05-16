package frontend

import (
	"context"
	"fmt"
	"net/http"

	"github.com/golang/protobuf/jsonpb"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/timestamp"
	tsc "github.com/Jim3Things/CloudChamber/simulation/internal/clients/trace_sink"
	"github.com/Jim3Things/CloudChamber/simulation/internal/tracing"
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
		tracing.WithContextValue(timestamp.EnsureTickInContext),
		tracing.AsInternal())
	defer span.End()

	vars := mux.Vars(r)
	from := vars["from"]
	count := vars["for"]

	err := doSessionHeader(
		ctx, w, r,
		func(ctx context.Context, session *sessions.Session) error {
			return server.sessions.ensureEstablishedSession(session)
		})

	if err != nil {
		postHTTPError(ctx, w, err)
		return
	}

	fromID, err := ensureNumber("from", from)
	if err != nil {
		postHTTPError(ctx, w, err)
		return
	}

	maxSize, err := ensurePositiveNumber("for", count)
	if err != nil {
		postHTTPError(ctx, w, err)
		return
	}

	tracing.UpdateSpanName(
		ctx,
		fmt.Sprintf(
			"Get up to %d trace entries, starting from entry #%d",
			maxSize,
			fromID))

	ch := tsc.GetTraces(ctx, fromID, maxSize)

	data := <-ch
	if data.Err != nil {
		postHTTPError(ctx, w, data.Err)
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
		tracing.WithName("Get tracing policy"),
		tracing.WithContextValue(timestamp.EnsureTickInContext),
		tracing.AsInternal())
	defer span.End()

	err := doSessionHeader(
		ctx, w, r,
		func(ctx context.Context, session *sessions.Session) error {
			return server.sessions.ensureEstablishedSession(session)
		})

	if err != nil {
		postHTTPError(ctx, w, err)
		return
	}

	policy, err := tsc.GetPolicy(ctx)
	if err != nil {
		postHTTPError(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	p := jsonpb.Marshaler{}
	err = p.Marshal(w, policy)

	httpErrorIf(ctx, w, err)
}
