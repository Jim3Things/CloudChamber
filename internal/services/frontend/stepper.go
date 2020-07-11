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
    "github.com/Jim3Things/CloudChamber/internal/tracing"
    st "github.com/Jim3Things/CloudChamber/internal/tracing/server"
    pb "github.com/Jim3Things/CloudChamber/pkg/protos/Stepper"
    ct "github.com/Jim3Things/CloudChamber/pkg/protos/common"
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
    _ = st.WithSpan(context.Background(), tracing.MethodName(1), func(ctx context.Context) error {

        err := doSessionHeader(ctx, w, r, func(_ context.Context, session *sessions.Session) error {
            st.Infof(ctx, -1, "Retrieved session, isNew=%v, value=%s", session.IsNew, dumpSessionState(session))
            return nil
        })
        if err != nil {
            return httpError(ctx, w, err)
        }

        stat, err := clients.Status()
        if err != nil {
            return httpError(ctx, w, err)
        }

        w.Header().Set("Content-Type", "application/json")
        w.Header().Set("ETag", fmt.Sprintf("%v", stat.Epoch))

        p := jsonpb.Marshaler{}
        return p.Marshal(w, stat)
    })
}

// Process an http request to advance the simulated time by a specified number
// of ticks.  The number defaults to 1.
func handleAdvance(w http.ResponseWriter, r *http.Request) {
    _ = st.WithSpan(context.Background(), tracing.MethodName(1), func(ctx context.Context) (err error) {
        var count int

        err = doSessionHeader(ctx, w, r, func(_ context.Context, session *sessions.Session) error {
            st.Infof(ctx, -1, "Retrieved session, isNew=%v, value=%s", session.IsNew, dumpSessionState(session))
            return nil
        })
        if err != nil {
            return httpError(ctx, w, err)
        }

        // Get the optional count of ticks to advance, and validate it
        vars := mux.Vars(r)
        arg, ok := vars["num"]
        if !ok || len(arg) == 0 {
            count = 1
        } else {
            count, err = strconv.Atoi(arg)
            if err != nil || count <= 0 {
                return httpError(ctx, w, NewErrInvalidStepperRate(arg))
            }
        }

        // Advance the time the request number of ticks
        for i := 0; i < count; i++ {
            if err := clients.Advance(); err != nil {
                return httpError(ctx, w, err)
            }
        }

        // .. and get the current time to return in the body of the response
        now, err := clients.Now()
        if err != nil {
            return httpError(ctx, w, err)
        }

        w.Header().Set("Content-Type", "application/json")

        p := jsonpb.Marshaler{}
        return p.Marshal(w, now)
    })
}

// Process an http request to change the simulated time service's policy.  The
// policy may be manual, which only advances in response to an explicit
// request, or it may advance at some number of ticks per second.  If the
// latter, the default rate is 1 tick per second.
func handleSetMode(w http.ResponseWriter, r *http.Request) {
    _ = st.WithSpan(context.Background(), tracing.MethodName(1), func(ctx context.Context) error {
        vars := mux.Vars(r)
        args := strings.Split(strings.Replace(vars["type"], "=", ":", 1), ":")

        var delay *duration.Duration
        var policy pb.StepperPolicy

        err := doSessionHeader(ctx, w, r, func(_ context.Context, session *sessions.Session) error {
            st.Infof(ctx, -1, "Retrieved session, isNew=%v, value=%s", session.IsNew, dumpSessionState(session))
            return nil
        })
        if err != nil {
            return httpError(ctx, w, err)
        }

        matchString := r.Header.Get("If-Match")
        match, err := strconv.ParseInt(matchString, 10, 64)
        if err != nil {
            return httpError(ctx, w, NewErrBadMatchType(matchString))
        }

        switch args[0] {
        case "manual":
            if len(args) != 1 {
                return httpError(ctx, w, NewErrInvalidRateRequest())
            }

            delay = &duration.Duration{Seconds: 0, Nanos: 0}
            policy = pb.StepperPolicy_Manual

        case "automatic":
            delay = &duration.Duration{Seconds: 1, Nanos: 0}
            if len(args) == 2 {
                tps, err := strconv.Atoi(args[1])
                if err != nil {
                    return httpError(ctx, w, NewErrInvalidStepperRate(args[1]))
                }

                if tps > 1 {
                    delay.Seconds = 0
                    delay.Nanos = int32(1_000_000_000 / tps)
                }
            }

            policy = pb.StepperPolicy_Measured

        default:
            return httpError(ctx, w, NewErrInvalidStepperMode(args[0]))
        }

        if err := clients.SetPolicy(policy, delay, match); err != nil {
            return httpError(ctx, w, NewErrStepperFailedToSetPolicy())
        }

        w.Header().Set("ETag", fmt.Sprintf("%v", match + 1))

        return nil
    })
}

// Process an http request to wait for the first tick after the one specified.
// This can be used to be notified of time changes when the simulated time is
// in the 'automatic' state.
func handleWaitFor(w http.ResponseWriter, r *http.Request) {
    _ = st.WithSpan(context.Background(), tracing.MethodName(1), func(ctx context.Context) error {
        vars := mux.Vars(r)
        after := vars["after"]

        err := doSessionHeader(ctx, w, r, func(_ context.Context, session *sessions.Session) error {
            st.Infof(ctx, -1, "Retrieved session, isNew=%v, value=%s", session.IsNew, dumpSessionState(session))
            return nil
        })
        if err != nil {
            return httpError(ctx, w, err)
        }

        afterTick, err := strconv.ParseInt(after, 10, 64)
        if err != nil || afterTick < 0 {
            return httpError(ctx, w, NewErrInvalidStepperAfter(after))
        }

        ch, err := clients.After(&ct.Timestamp{Ticks: afterTick + 1})
        if err != nil {
            return httpError(ctx, w, err)
        }

        data := <-ch
        if data.Err != nil {
            return httpError(ctx, w, err)
        }

        w.Header().Set("Content-Type", "application/json")

        p := jsonpb.Marshaler{}
        return p.Marshal(w, data.Time)
    })
}

// Process an http request to get the current simulated time.
func handleGetNow(w http.ResponseWriter, r *http.Request) {
    _ = st.WithSpan(context.Background(), tracing.MethodName(1), func(ctx context.Context) error {
        err := doSessionHeader(ctx, w, r, func(_ context.Context, session *sessions.Session) error {
            st.Infof(ctx, -1, "Retrieved session, isNew=%v, value=%s", session.IsNew, dumpSessionState(session))
            return nil
        })
        if err != nil {
            return httpError(ctx, w, err)
        }

        now, err := clients.Now()
        if err != nil {
            return httpError(ctx, w, err)
        }

        w.Header().Set("Content-Type", "application/json")

        p := jsonpb.Marshaler{}
        return p.Marshal(w, now)
    })
}
