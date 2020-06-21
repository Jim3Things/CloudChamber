package frontend

import (
    "context"
    "net/http"

    "github.com/golang/protobuf/jsonpb"
    "github.com/gorilla/mux"

    clients "github.com/Jim3Things/CloudChamber/internal/clients/timestamp"
    "github.com/Jim3Things/CloudChamber/internal/tracing"
    st "github.com/Jim3Things/CloudChamber/internal/tracing/server"
)

func stepperAddRoutes(routeBase *mux.Router) {
    routeStepper := routeBase.PathPrefix("/stepper").Subrouter()

    routeStepper.HandleFunc("", handleGetStatus).Methods("GET")
    routeStepper.HandleFunc("/", handleGetStatus).Methods("GET")

    routeStepper.HandleFunc("", handleAdvance).Queries("advance", "{num}").Methods("PUT")
    routeStepper.HandleFunc("", handleSetMode).Queries("mode", "{type}").Methods("PUT")
    routeStepper.HandleFunc("/now", handleGetNow).Methods("GET")
}

func handleGetStatus(w http.ResponseWriter, r *http.Request) {
    _ = st.WithSpan(context.Background(), tracing.MethodName(1), func(ctx context.Context) (err error) {
        return nil
    })
}

func handleAdvance(w http.ResponseWriter, r *http.Request) {
    _ = st.WithSpan(context.Background(), tracing.MethodName(1), func(ctx context.Context) (err error) {
        return nil
    })
}

func handleSetMode(w http.ResponseWriter, r *http.Request) {
    _ = st.WithSpan(context.Background(), tracing.MethodName(1), func(ctx context.Context) (err error) {
        return nil
    })
}

func handleGetNow(w http.ResponseWriter, _ *http.Request) {
    _ = st.WithSpan(context.Background(), tracing.MethodName(1), func(ctx context.Context) (err error) {
        now, err := clients.Now()
        if err != nil {
            httpError(ctx, w, err)
            return err
        }

        p := jsonpb.Marshaler{}
        return p.Marshal(w, now)
    })
}
