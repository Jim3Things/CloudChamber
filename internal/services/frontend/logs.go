package frontend

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func logsAddRoutes(routeBase *mux.Router) {
	router := routeBase.PathPrefix("/logs").Subrouter()

	router.HandleFunc("", handlerLogsRoot).Methods("GET")
	router.HandleFunc("/", handlerLogsRoot).Methods("GET")
}

func handlerLogsRoot(w http.ResponseWriter, _ *http.Request) {

	_, _ = fmt.Fprintf(w, "Logs (Root)")
}


