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

    clients "github.com/Jim3Things/CloudChamber/internal/clients/timestamp"
    "github.com/Jim3Things/CloudChamber/internal/tracing"
    st "github.com/Jim3Things/CloudChamber/internal/tracing/server"
    pb "github.com/Jim3Things/CloudChamber/pkg/protos/Stepper"
)

func stepperAddRoutes(routeBase *mux.Router) {
    routeStepper := routeBase.PathPrefix("/stepper").Subrouter()

    routeStepper.HandleFunc("", handleGetStatus).Methods("GET")
    routeStepper.HandleFunc("/", handleGetStatus).Methods("GET")

    routeStepper.HandleFunc("", handleAdvance).Queries("advance", "{num}").Methods("PUT")
    routeStepper.HandleFunc("", handleSetMode).Queries("mode", "{type}").Methods("PUT")
    routeStepper.HandleFunc("/now", handleGetNow).Methods("GET")
}

func handleGetStatus(w http.ResponseWriter, _ *http.Request) {
    _ = st.WithSpan(context.Background(), tracing.MethodName(1), func(ctx context.Context) error {
        stat, err := clients.Status()
        if err != nil {
            httpError(ctx, w, err)
            return err
        }

        w.Header().Set("Content-Type", "application/json")
        w.Header().Set("ETag", fmt.Sprintf("%v", stat.Epoch))

        p := jsonpb.Marshaler{}
        return p.Marshal(w, stat)
    })
}

func handleAdvance(w http.ResponseWriter, r *http.Request) {
    _ = st.WithSpan(context.Background(), tracing.MethodName(1), func(ctx context.Context) (err error) {
        var count int

        vars := mux.Vars(r)
        arg, ok := vars["num"]
        if !ok || len(arg) == 0 {
            count = 1
        } else {
            count, err = strconv.Atoi(arg)
            if err != nil || count <= 0 {
                err = NewErrInvalidStepperRate(arg)
                httpError(ctx, w, err)
                return err
            }
        }

        for i := 0; i < count; i++ {
            if err := clients.Advance(); err != nil {
                httpError(ctx, w, err)
                return err
            }
        }

        now, err := clients.Now()
        if err != nil {
            httpError(ctx, w, err)
            return err
        }

        w.Header().Set("Content-Type", "application/json")

        p := jsonpb.Marshaler{}
        return p.Marshal(w, now)
    })
}

func handleSetMode(w http.ResponseWriter, r *http.Request) {
    _ = st.WithSpan(context.Background(), tracing.MethodName(1), func(ctx context.Context) (err error) {
        vars := mux.Vars(r)
        args := strings.Split(strings.Replace(vars["type"], "=", ":", 1), ":")

        var delay *duration.Duration
        var policy pb.StepperPolicy

        var match int64

        matchString := r.Header.Get("If-Match")
        match, err = strconv.ParseInt(matchString, 10, 64)
        if err != nil {
            httpError(ctx, w, NewErrBadMatchType(matchString))
            return err
        }

        switch args[0] {
        case "manual":
            if len(args) != 1 {
                err := NewErrInvalidRateRequest()
                httpError(ctx, w, err)
                return err
            }

            delay = &duration.Duration{Seconds: 0, Nanos: 0}
            policy = pb.StepperPolicy_Manual

        case "automatic":
            delay = &duration.Duration{Seconds: 1, Nanos: 0}
            if len(args) == 2 {
                tps, err := strconv.Atoi(args[1])
                if err != nil {
                    err = NewErrInvalidStepperRate(args[1])
                    httpError(ctx, w, err)
                    return err
                }

                if tps > 1 {
                    delay.Seconds = 0
                    delay.Nanos = int32(1_000_000_000 / tps)
                }
            }

            policy = pb.StepperPolicy_Measured

        default:
            err := NewErrInvalidStepperMode(args[0])
            httpError(ctx, w, err)
            return err
        }

        if err := clients.SetPolicy(policy, delay, match); err != nil {
            err = NewErrStepperFailedToSetPolicy()
            httpError(ctx, w, err)
            return err
        }

        w.Header().Set("ETag", fmt.Sprintf("%v", match + 1))

        return nil
    })
}

func handleGetNow(w http.ResponseWriter, _ *http.Request) {
    _ = st.WithSpan(context.Background(), tracing.MethodName(1), func(ctx context.Context) error {
        now, err := clients.Now()
        if err != nil {
            httpError(ctx, w, err)
            return err
        }

        w.Header().Set("Content-Type", "application/json")

        p := jsonpb.Marshaler{}
        return p.Marshal(w, now)
    })
}
