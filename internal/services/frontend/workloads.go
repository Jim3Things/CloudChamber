// This module containes the routines etc to implement the frontend handlers for the workloads
// part of the API
//

//package frontend
package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

// Workload is a representation of a specific workload
//
type Workload struct {
	Name            string
	TimeCreation    time.Time
	TimeLasteUpdate time.Time
	TimeLastStep    time.Time
	StepCount       uint64
	Enabled         bool
}

type DbWorkloads struct {
	Mutex     sync.Mutex
	Workloads map[string]Workload
}

var (
	dbWorkloads DbWorkloads
)

func workloadsAddRoutes(routeBase *mux.Router) {

	routeWorkloads := routeBase.PathPrefix("/workloads").Subrouter()

	routeWorkloads.HandleFunc("", handlerWorkloadsList).Methods("GET")
	routeWorkloads.HandleFunc("/", handlerWorkloadsList).Methods("GET")

	routeWorkloads.HandleFunc("/{workloadname}", handlerWorkloadsFetch).Methods("GET")
	routeWorkloads.HandleFunc("/{workloadname}/", handlerWorkloadsFetch).Methods("GET")

	// In the following, the "GET" method is allowed just for the purposes of test and
	// evaluation. At somepoint, it will need to be removed, but in the meantime, leaving
	// it there allows simple experimentation with just a browser.
	//
	routeWorkloads.HandleFunc("/{workloadname}/add", handlerWorkloadsAdd).Methods("PUT", "GET")
	routeWorkloads.HandleFunc("/{workloadname}/remove", handlerWorkloadsRemove).Methods("DELETE", "GET")
	routeWorkloads.HandleFunc("/{workloadname}/enable", handlerWorkloadsEnable).Methods("PUT", "GET")
	routeWorkloads.HandleFunc("/{workloadname}/disable", handlerWorkloadsDisable).Methods("PUT", "GET")
	routeWorkloads.HandleFunc("/{workloadname}/update", handlerWorkloadsUpdate).Methods("PUT", "GET")
	routeWorkloads.HandleFunc("/{workloadname}/step", handlerWorkloadsSingleStep).Methods("PUT", "GET")
	routeWorkloads.HandleFunc("/{workloadname}/step/{count}", handlerWorkloadsStep).Methods("PUT", "GET")
}

func workloadsDisplayArguments(w http.ResponseWriter, r *http.Request, command string) {

	vars := mux.Vars(r)

	workload := vars["workloadname"]

	fmt.Fprintf(w, "Workload: %s command: %s", workload, command)
}

func handlerWorkloadsList(w http.ResponseWriter, r *http.Request) {

	fmt.Fprintf(w, "Workloads (List)")
}

func handlerWorkloadsFetch(w http.ResponseWriter, r *http.Request) {

	workloadsDisplayArguments(w, r, "fetch")
}

func handlerWorkloadsFetch2(w http.ResponseWriter, r *http.Request) {

	workloadsDisplayArguments(w, r, "fetch2")
}

func handlerWorkloadsFetch2B(w http.ResponseWriter, r *http.Request) {

	workloadsDisplayArguments(w, r, "fetch2B")
}

func handlerWorkloadsAdd(w http.ResponseWriter, r *http.Request) {

	workloadsDisplayArguments(w, r, "add")
}

func handlerWorkloadsRemove(w http.ResponseWriter, r *http.Request) {

	workloadsDisplayArguments(w, r, "remove")
}

func handlerWorkloadsEnable(w http.ResponseWriter, r *http.Request) {

	workloadsDisplayArguments(w, r, "enable")
}

func handlerWorkloadsDisable(w http.ResponseWriter, r *http.Request) {

	workloadsDisplayArguments(w, r, "disable")
}

func handlerWorkloadsUpdate(w http.ResponseWriter, r *http.Request) {

	workloadsDisplayArguments(w, r, "update")
}

func handlerWorkloadsSingleStep(w http.ResponseWriter, r *http.Request) {

	workloadsDisplayArguments(w, r, "singlestep")
}

func handlerWorkloadsStep(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	workload := vars["workloadname"]
	count := vars["count"]

	fmt.Fprintf(w, "Workload: %s - stepcount: %s", workload, count)
}
