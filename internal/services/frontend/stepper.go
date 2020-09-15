package frontend

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	clients "github.com/Jim3Things/CloudChamber/internal/clients/timestamp"
	"github.com/Jim3Things/CloudChamber/internal/common"
	"github.com/Jim3Things/CloudChamber/internal/tracing"
	ct "github.com/Jim3Things/CloudChamber/pkg/protos/common"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/services"
)

const (
	nanosPerSecond = 1_000_000_000
)

func stepperAddRoutes(routeBase *mux.Router) {
	routeStepper := routeBase.PathPrefix("/stepper").Subrouter()

	routeStepper.HandleFunc("", handleGetStatus).Methods("GET")
	routeStepper.HandleFunc("/", handleGetStatus).Methods("GET")

	routeStepper.HandleFunc("", handleAdvance).Queries("advance", "{num}").Methods("PUT")
	routeStepper.HandleFunc("", handleSetMode).Queries("mode", "{type}").Methods("PUT")

	routeStepper.HandleFunc("/now", handleWaitFor).Queries("after", "{after}").Methods("GET")
	routeStepper.HandleFunc("/now", handleGetNow).Methods("GET")
}

// Process an http request for the current Stepper service status.
func handleGetStatus(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithName("Get Simulated Time Service Status"),
		tracing.WithContextValue(clients.EnsureTickInContext),
		tracing.AsInternal())
	defer span.End()

	err := doSessionHeader(ctx, w, r, func(_ context.Context, session *sessions.Session) error {
		return ensureEstablishedSession(session)
	})

	if err != nil {
		postHttpError(ctx, w, err)
		return
	}

	stat, err := clients.Status(ctx)

	if err != nil {
		postHttpError(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("ETag", fmt.Sprintf("%v", stat.Epoch))

	p := jsonpb.Marshaler{}
	err = p.Marshal(w, stat)

	httpErrorIf(ctx, w, err)
}

// Process an http request to advance the simulated time by a specified number
// of ticks.  The number defaults to 1.
func handleAdvance(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithName("Advance Simulated Time"),
		tracing.WithContextValue(clients.EnsureTickInContext))
	defer span.End()

	var count int

	err := doSessionHeader(ctx, w, r, func(_ context.Context, session *sessions.Session) error {
		return ensureEstablishedSession(session)
	})

	if err != nil {
		postHttpError(ctx, w, err)
		return
	}

	// Get the optional count of ticks to advance, and validate it
	vars := mux.Vars(r)
	arg, ok := vars["num"]

	if !ok || len(arg) == 0 {
		count = 1
	} else {
		count, err = strconv.Atoi(arg)
		if err != nil || count <= 0 {
			postHttpError(ctx, w, NewErrInvalidStepperRate(arg))
			return
		}
	}

	// Advance the time the request number of ticks
	for i := 0; i < count; i++ {
		if err = clients.Advance(ctx); err != nil {
			postHttpError(ctx, w, err)
			return
		}
	}

	// .. and get the current time to return in the body of the response
	ctx = common.ContextWithTick(ctx, clients.Tick(ctx))
	w.Header().Set("Content-Type", "application/json")

	p := jsonpb.Marshaler{}
	err = p.Marshal(w, &ct.Timestamp{Ticks: common.TickFromContext(ctx)})

	httpErrorIf(ctx, w, err)
}

// Process an http request to change the simulated time service's policy.  The
// policy may be manual, which only advances in response to an explicit
// request, or it may advance at some number of ticks per second.  If the
// latter, the default rate is 1 tick per second.
func handleSetMode(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithName("Set Simulated Time Policy"),
		tracing.WithContextValue(clients.EnsureTickInContext))
	defer span.End()

	vars := mux.Vars(r)
	args := strings.Split(strings.Replace(vars["type"], "=", ":", 1), ":")

	var delay *duration.Duration
	var policy pb.StepperPolicy

	err := doSessionHeader(ctx, w, r, func(_ context.Context, session *sessions.Session) error {
		return ensureEstablishedSession(session)
	})

	if err != nil {
		postHttpError(ctx, w, err)
		return
	}

	matchString := r.Header.Get("If-Match")
	match, err := strconv.ParseInt(matchString, 10, 64)

	if err != nil {
		postHttpError(ctx, w, NewErrBadMatchType(matchString))
		return
	}

	switch args[0] {
	case "manual":
		if len(args) != 1 {
			postHttpError(ctx, w, NewErrInvalidRateRequest())
			return
		}

		delay = &duration.Duration{Seconds: 0, Nanos: 0}
		policy = pb.StepperPolicy_Manual

	case "automatic":
		delay = &duration.Duration{Seconds: 1, Nanos: 0}
		if len(args) == 2 {
			tps, err := strconv.Atoi(args[1])
			if err != nil || tps < 1 {
				postHttpError(ctx, w, NewErrInvalidStepperRate(args[1]))
				return
			}

			if tps > 1 {
				delay.Seconds = 0
				delay.Nanos = int32(nanosPerSecond / tps)
			}
		}

		policy = pb.StepperPolicy_Measured

	default:
		postHttpError(ctx, w, NewErrInvalidStepperMode(args[0]))
		return
	}

	if err = clients.SetPolicy(ctx, policy, delay, match); err != nil {
		postHttpError(ctx, w, NewErrStepperFailedToSetPolicy())
		return
	}

	w.Header().Set("ETag", fmt.Sprintf("%v", match+1))

	httpErrorIf(ctx, w, err)
}

// Process an http request to wait for the first tick after the one specified.
// This can be used to be notified of time changes when the simulated time is
// in the 'automatic' state.
func handleWaitFor(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	after := vars["after"]

	var data clients.TimeData

	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithName("Wait Until Simulate Time After..."),
		tracing.WithContextValue(clients.EnsureTickInContext))
	defer span.End()

	err := doSessionHeader(
		ctx, w, r,
		func(_ context.Context, session *sessions.Session) error {

			if err := ensureEstablishedSession(session); err != nil {
				return err
			}

			afterTick, err := ensurePositiveNumber("after", after)
			if err != nil {
				return err
			}

			data = <-clients.After(ctx, &ct.Timestamp{Ticks: afterTick + 1})
			if data.Err != nil {
				return data.Err
			}

			return nil
		})

	if err != nil {
		postHttpError(ctx, w, err)
		return
	}

	ctx = common.ContextWithTick(ctx, data.Time.Ticks)

	w.Header().Set("Content-Type", "application/json")

	p := jsonpb.Marshaler{}
	err = p.Marshal(w, data.Time)

	httpErrorIf(ctx, w, err)
}

// Process an http request to get the current simulated time.
func handleGetNow(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithName("Get Current Simulated Time"),
		tracing.WithContextValue(clients.EnsureTickInContext),
		tracing.AsInternal())
	defer span.End()

	err := doSessionHeader(ctx, w, r, func(_ context.Context, session *sessions.Session) error {
		return ensureEstablishedSession(session)
	})

	if err != nil {
		postHttpError(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	p := jsonpb.Marshaler{}
	err = p.Marshal(w, &ct.Timestamp{Ticks: common.TickFromContext(ctx)})

	httpErrorIf(ctx, w, err)
}
