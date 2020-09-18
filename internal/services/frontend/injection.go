package frontend

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func injectionAddRoutes(routeBase *mux.Router) {
	router := routeBase.PathPrefix("/inject").Subrouter()

	router.HandleFunc("", handlerInjectionRoot).Methods("GET")
	router.HandleFunc("/", handlerInjectionRoot).Methods("GET")
}

func handlerInjectionRoot(w http.ResponseWriter, _ *http.Request) {

	_, _ = fmt.Fprintf(w, "Injection (Root)")
}
