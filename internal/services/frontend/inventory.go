// This module containes the routines etc to implement the frontend handlers for the inventory operators
// part of the API
//

package frontend

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Jim3Things/CloudChamber/internal/tracing"
	st "github.com/Jim3Things/CloudChamber/internal/tracing/server"
	"github.com/golang/protobuf/jsonpb"
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

func inventoryAddRoutes(routeBase *mux.Router) {

	routeRacks := routeBase.PathPrefix("/racks").Subrouter() //api/racks/rackid/ declared the path prefix

	routeRacks.HandleFunc("", handlerRacksList).Methods("GET") //handler rack list function is called for the Get method
	routeRacks.HandleFunc("/", handlerRacksList).Methods("GET")

	// In the following, the "GET" method is allowed just for the purposes of test and
	// evaluation. At somepoint, it will need to be removed, but in the meantime, leaving
	// it there allows simple experimentation with just a browser.
	//
	// As a reminder,
	//	 PUT is idempotent so translates to UPDATE in the CRUD methodolgy
	//   POST is NOT idempotent so translates to CREATE in the CRUD methodolgy
	//
	//routeRacks.HandleFunc("/racks/{rackid}", handlerRacksCreate).Methods("POST", "GET") // May be only GET
	routeRacks.HandleFunc("/{rackid}", handlerRacksRead).Methods("GET")
	//routeRacks.HandleFunc("/racks/{rackid}", handlerRacksUpdate).Methods("PUT", "GET")
	//routeRacks.HandleFunc("/racks/{rackid}", handlerRacksDelete).Methods("DELETE", "GET")
}

func racksDisplayArguments(w http.ResponseWriter, r *http.Request, command string) {

	vars := mux.Vars(r)

	racks := vars["rackid"]

	fmt.Fprintf(w, "racks: %v command: %v", racks, command)
}

func handlerRacksList(w http.ResponseWriter, r *http.Request) {
	_ = st.WithSpan(context.Background(), tracing.MethodName(1), func(ctx context.Context) (err error) {

		if _, err := fmt.Fprintln(w, "Racks (List)"); err != nil {
			httpError(ctx, w, err)
			return err
		}

		b := r.URL.String()
		if !strings.HasSuffix(b, "/") {
			b += "/"
		}

		return dbInventory.Scan(func(name string) (err error) {
			target := fmt.Sprintf("%s%s", b, name)

			st.Infof(ctx, -1, "   Listing rack '%s' at '%s'", name, target)

			if _, err = fmt.Fprintln(w, target); err != nil {
				httpError(ctx, w, err)
			}

			return err
		})
	})

}

//func handlerracksCreate(w http.ResponseWriter, r *http.Request) {

//	racksDisplayArguments(w, r, "add")
//}

func handlerRacksRead(w http.ResponseWriter, r *http.Request) {
	_ = st.WithSpan(context.Background(), tracing.MethodName(1), func(ctx context.Context) (err error) {
		vars := mux.Vars(r)
		rackid := vars["rackid"]

		u, err := dbInventory.Get(rackid)
		if err != nil {
			httpError(ctx, w, err)
			return err
		}
		w.Header().Set("Content-Type", "application/json")

		st.Infof(ctx, -1, "Returning details for rack %q: %v", rackid, u)

		// Get the user entry, and serialize it to json
		// (export userPublic to json and return that as the body)
		p := jsonpb.Marshaler{}
		return p.Marshal(w, u)

	})
}

//func handlerracksUpdate(w http.ResponseWriter, r *http.Request) {

//	racksDisplayArguments(w, r, "update")
//}

//func handlerracksDelete(w http.ResponseWriter, r *http.Request) {

//	racksDisplayArguments(w, r, "remove")
//}
