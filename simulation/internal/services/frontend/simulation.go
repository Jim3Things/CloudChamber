package frontend

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/ptypes"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/timestamp"
	"github.com/Jim3Things/CloudChamber/simulation/internal/common"
	"github.com/Jim3Things/CloudChamber/simulation/internal/tracing"
	pb "github.com/Jim3Things/CloudChamber/simulation/pkg/protos/admin"
)

func simulationAddRoutes(routeBase *mux.Router) {
	routeSim := routeBase.PathPrefix("/simulation").Subrouter()

	routeSim.HandleFunc("", handlerSimStatus).Methods("GET")
	routeSim.HandleFunc("/", handlerSimStatus).Methods("GET")

	routeSim.HandleFunc("/sessions", handlerSimSessionList).Methods("GET")
	routeSim.HandleFunc("/sessions/", handlerSimSessionList).Methods("GET")

	routeSim.HandleFunc("/sessions/{id}", handlerSimSessionDetail).Methods("GET")
	routeSim.HandleFunc("/sessions/{id}", handlerSimRemoveSession).Methods("DELETE")
}

// handlerSimStatus collects and returns the current status report for the
// simulation.  Any logged in user can request the simulation status.
func handlerSimStatus(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithName("Get Simulation Status"),
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

	ts, err := ptypes.TimestampProto(server.startTime)
	if err != nil {
		postHTTPError(ctx, w, err)
		return
	}

	status := &pb.SimulationStatus{
		FrontEndStartedAt: ts,
		InactivityTimeout: ptypes.DurationProto(server.sessions.inactivity()),
	}

	p := jsonpb.Marshaler{}
	err = p.Marshal(w, status)

	httpErrorIf(ctx, w, err)
}

// handlerSimSessionList returns the summary list of active sessions containing
// each session ID and the uri to use to get the details on it.  The caller
// must be a logged in user with account management rights.
func handlerSimSessionList(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithName("Get Simulation Active Session List"),
		tracing.WithContextValue(timestamp.EnsureTickInContext),
		tracing.AsInternal())
	defer span.End()

	err := doSessionHeader(
		ctx, w, r,
		func(ctx context.Context, session *sessions.Session) error {
			return canManageAccounts(ctx, session, "")
		})

	if err != nil {
		postHTTPError(ctx, w, err)
		return
	}

	list := &pb.SessionSummary{
		Sessions: []*pb.SessionSummary_Session{},
	}

	b := common.URLPrefix(r)

	for _, key := range server.sessions.knownIDs() {
		list.Sessions = append(list.Sessions, &pb.SessionSummary_Session{
			Id:  key,
			Uri: fmt.Sprintf("%s%d", b, key),
		})
	}

	p := jsonpb.Marshaler{}
	err = p.Marshal(w, list)

	httpErrorIf(ctx, w, err)
}

// handlerSimSessionDetails returns detail information about the active session
// specified by the id field of the uri.  The caller must be from a logged in
// session with account management rights.
func handlerSimSessionDetail(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithName("Get Active Session Details"),
		tracing.WithContextValue(timestamp.EnsureTickInContext),
		tracing.AsInternal())
	defer span.End()

	vars := mux.Vars(r)
	idStr := vars["id"]

	err := doSessionHeader(
		ctx, w, r,
		func(ctx context.Context, session *sessions.Session) error {
			return canManageAccounts(ctx, session, "")
		})

	if err != nil {
		postHTTPError(ctx, w, err)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		postHTTPError(ctx, w, &HTTPError{
			SC:   http.StatusBadRequest,
			Base: err,
		})
		return
	}

	session, ok := server.sessions.get(id)
	if !ok {
		postHTTPError(ctx, w, NewErrSessionNotFound(id))
		return
	}

	ts, err := ptypes.TimestampProto(session.timeout)
	if err != nil {
		postHTTPError(ctx, w, err)
		return
	}

	resp := &pb.SessionStatus{
		UserName: session.name,
		Timeout:  ts,
	}

	tracing.UpdateSpanName(
		ctx,
		"Getting Details for Session %d (user %q)",
		id,
		session.name)

	p := jsonpb.Marshaler{}
	err = p.Marshal(w, resp)

	httpErrorIf(ctx, w, err)
}

// handlerSimRemoveSession forcibly deletes the active session specified by the
// id field in the uri.  The caller must be from a logged in session with
// account management rights.
func handlerSimRemoveSession(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithName("Delete An Active Session"),
		tracing.WithContextValue(timestamp.EnsureTickInContext),
		tracing.AsInternal())
	defer span.End()

	vars := mux.Vars(r)
	idStr := vars["id"]

	err := doSessionHeader(
		ctx, w, r,
		func(ctx context.Context, session *sessions.Session) error {
			return canManageAccounts(ctx, session, "")
		})

	if err != nil {
		postHTTPError(ctx, w, err)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		postHTTPError(ctx, w, &HTTPError{
			SC:   http.StatusBadRequest,
			Base: err,
		})

		return
	}

	session, ok := server.sessions.delete(id)
	if !ok {
		postHTTPError(ctx, w, NewErrSessionNotFound(id))
		return
	}

	tracing.UpdateSpanName(
		ctx,
		"Delete Active Session %d (user %q)",
		id,
		session.name)
}
