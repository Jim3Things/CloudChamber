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
//TODO This is just a placeholder until we have proper workload definitions
//     held in a persisted store (Etcd)
//
type Workload struct {
	Name            string
	TimeCreation    time.Time
	TimeLasteUpdate time.Time
	TimeLastStep    time.Time
	StepCount       uint64
	Enabled         bool
}

// DbWorkloads is a container used to established synchronized access to
// the in-memory set of workload records.
//
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

	// In the following, the "GET" method is allowed just for the purposes of test and
	// evaluation. At somepoint, it will need to be removed, but in the meantime, leaving
	// it there allows simple experimentation with just a browser.
	//
	// As a reminder,
	//	 PUT is idempotent so translates to UPDATE in the CRUD methodolgy
	//   POST is NOT idempotent so translates to CREATE in the CRUD methodolgy
	//
	routeWorkloads.HandleFunc("/{workloadname}", handlerWorkloadsCreate).Methods("POST", "GET")
	routeWorkloads.HandleFunc("/{workloadname}", handlerWorkloadsRead).Methods("GET")
	routeWorkloads.HandleFunc("/{workloadname}", handlerWorkloadsUpdate).Methods("PUT", "GET")
	routeWorkloads.HandleFunc("/{workloadname}", handlerWorkloadsDelete).Methods("DELETE", "GET")
}

func workloadsDisplayArguments(w http.ResponseWriter, r *http.Request, command string) {

	vars := mux.Vars(r)

	workload := vars["workloadname"]

	fmt.Fprintf(w, "Workload: %s command: %s", workload, command)
}

func handlerWorkloadsList(w http.ResponseWriter, r *http.Request) {

	fmt.Fprintf(w, "Workloads (List)")
}

func handlerWorkloadsCreate(w http.ResponseWriter, r *http.Request) {

	workloadsDisplayArguments(w, r, "add")
}

func handlerWorkloadsRead(w http.ResponseWriter, r *http.Request) {

	workloadsDisplayArguments(w, r, "fetch")
}

func handlerWorkloadsUpdate(w http.ResponseWriter, r *http.Request) {

	workloadsDisplayArguments(w, r, "update")
}

func handlerWorkloadsDelete(w http.ResponseWriter, r *http.Request) {

	workloadsDisplayArguments(w, r, "remove")
}
