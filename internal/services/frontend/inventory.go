// This module containes the routines etc to implement the frontend handlers for the inventory operators
// part of the API
//

package frontend

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

// Inventory is a representation of a specific inventory operators
//
//TODO This is just a placeholder until we have proper inventory definitions
//     held in a persisted store (Etcd)
//
type Inventory struct {
	Name            string
	TimeCreation    time.Time
	TimeLasteUpdate time.Time
	TimeLastStep    time.Time
	StepCount       uint64
	Enabled         bool
}

// DbInventory is a container used to established synchronized access to
// the in-memory set of Inventory records.
//
type DbInventory struct {
	Mutex     sync.Mutex
	Inventory map[string]Inventory
}

var (
	dbInventory DbInventory
)

func inventoryAddRoutes(routeBase *mux.Router) {

	routeInventory := routeBase.PathPrefix("/Inventory").Subrouter()

	routeInventory.HandleFunc("", handlerInventoryList).Methods("GET")
	routeInventory.HandleFunc("/", handlerInventoryList).Methods("GET")

	// In the following, the "GET" method is allowed just for the purposes of test and
	// evaluation. At somepoint, it will need to be removed, but in the meantime, leaving
	// it there allows simple experimentation with just a browser.
	//
	// As a reminder,
	//	 PUT is idempotent so translates to UPDATE in the CRUD methodolgy
	//   POST is NOT idempotent so translates to CREATE in the CRUD methodolgy
	//
	routeInventory.HandleFunc("/{rackid}", handlerInventoryCreate).Methods("POST", "GET") // May be only GET
	routeInventory.HandleFunc("/{rackid}", handlerInventoryRead).Methods("GET")
	routeInventory.HandleFunc("/{rackid}", handlerInventoryUpdate).Methods("PUT", "GET")
	routeInventory.HandleFunc("/{rackid}", handlerInventoryDelete).Methods("DELETE", "GET")
}

func inventoryDisplayArguments(w http.ResponseWriter, r *http.Request, command string) {

	vars := mux.Vars(r)

	inventory := vars["rackid"]

	fmt.Fprintf(w, "Inventory: %s command: %s", inventory, command)
}

func handlerInventoryList(w http.ResponseWriter, r *http.Request) {

	fmt.Fprintf(w, "Inventory (List)")
}

func handlerInventoryCreate(w http.ResponseWriter, r *http.Request) {

	inventoryDisplayArguments(w, r, "add")
}

func handlerInventoryRead(w http.ResponseWriter, r *http.Request) {

	inventoryDisplayArguments(w, r, "fetch")
}

func handlerInventoryUpdate(w http.ResponseWriter, r *http.Request) {

	inventoryDisplayArguments(w, r, "update")
}

func handlerInventoryDelete(w http.ResponseWriter, r *http.Request) {









	inventoryDisplayArguments(w, r, "remove")
}
