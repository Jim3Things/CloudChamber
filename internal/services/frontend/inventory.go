// This module contains the routines etc to implement the frontend handlers for the inventory operators
// part of the API
//

package frontend

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/sessions"

	"github.com/golang/protobuf/jsonpb"
	"github.com/gorilla/mux"

	"github.com/Jim3Things/CloudChamber/internal/clients/timestamp"
	"github.com/Jim3Things/CloudChamber/internal/common"
	"github.com/Jim3Things/CloudChamber/internal/tracing"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/inventory"
)

func inventoryAddRoutes(routeBase *mux.Router) {

	routeRacks := routeBase.PathPrefix("/racks").Subrouter()

	routeRacks.HandleFunc("", handlerRacksList).Methods("GET")
	routeRacks.HandleFunc("/", handlerRacksList).Methods("GET")

	// In the following, the "GET" method is allowed just for the purposes of test and
	// evaluation. At some point, it will need to be removed, but in the meantime, leaving
	// it there allows simple experimentation with just a browser.
	//
	// As a reminder,
	//	 PUT is idempotent so translates to UPDATE in the CRUD methodology
	//   POST is NOT idempotent so translates to CREATE in the CRUD methodology
	//
	// routeRacks.HandleFunc("/racks/{rackID}", handlerRacksCreate).Methods("POST", "GET") // May be only GET
	routeRacks.HandleFunc("/{rackID}", handlerRackRead).Methods("GET")
	routeRacks.HandleFunc("/{rackID}/blades", handlerBladesList).Methods("GET")
	routeRacks.HandleFunc("/{rackID}/blades/{bladeID}", handlerBladeRead).Methods("GET")
	// routeRacks.HandleFunc("/racks/{rackID}", handlerRacksDelete).Methods("DELETE", "GET")
}

func handlerRacksList(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithName("Get Cluster Inventory List of Racks"),
		tracing.AsInternal())
	defer span.End()

	// Pick up the current time to avoid repeatedly fetching the same value
	tick := clients.Tick(ctx)

	err := doSessionHeader(
		ctx, w, r,
		func(_ context.Context, session *sessions.Session) error {
			return ensureEstablishedSession(session)
		})

	if err != nil {
		postHttpError(ctx, tick, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	rackCount, maxBlades, maxCapacity := dbInventory.GetMemoData()
	res := &pb.ExternalZoneSummary{
		Racks:         make(map[string]*pb.ExternalRackSummary, rackCount),
		MaxBladeCount: maxBlades,
		MaxCapacity:   maxCapacity,
	}

	tracing.Infof(
		ctx,
		tick,
		"Listing all %d racks, max blades/rack=%d, max blade capacity=%v",
		rackCount,
		res.MaxBladeCount,
		res.MaxCapacity)

	b := common.URLPrefix(r)

	err = dbInventory.Scan(func(name string) error {
		target := fmt.Sprintf("%s%s", b, name)

		res.Racks[name] = &pb.ExternalRackSummary{Uri: target}

		tracing.Infof(ctx, tick, "   Listing rack %q at %q", name, target)

		return nil
	})

	if err != nil {
		postHttpError(ctx, tick, w, err)
		return
	}

	p := jsonpb.Marshaler{}
	err = p.Marshal(w, res)

	httpErrorIf(ctx, tick, w, err)
}

func handlerRackRead(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithName("Get Rack Details"),
		tracing.AsInternal())
	defer span.End()

	// Pick up the current time to avoid repeatedly fetching the same value
	tick := clients.Tick(ctx)

	vars := mux.Vars(r)
	rackID := vars["rackID"]

	err := doSessionHeader(
		ctx, w, r,
		func(_ context.Context, session *sessions.Session) error {
			return ensureEstablishedSession(session)
		})

	if err != nil {
		postHttpError(ctx, tick, w, err)
		return
	}

	rack, err := dbInventory.Get(rackID)
	if err != nil {
		postHttpError(ctx, tick, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	tracing.Infof(ctx, tick, "Returning details for rack %q: %v", rackID, rack)

	p := jsonpb.Marshaler{}
	err = p.Marshal(w, rack)

	httpErrorIf(ctx, tick, w, err)
}

func handlerBladesList(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithName("Get List of Blades in Selected Rack"),
		tracing.AsInternal())
	defer span.End()

	// Pick up the current time to avoid repeatedly fetching the same value
	tick := clients.Tick(ctx)

	err := doSessionHeader(
		ctx, w, r,
		func(_ context.Context, session *sessions.Session) error {
			return ensureEstablishedSession(session)
		})

	if err != nil {
		postHttpError(ctx, tick, w, err)
		return
	}

	vars := mux.Vars(r)
	rackID := vars["rackID"] // captured the key value in rackID variable

	if _, err = fmt.Fprintf(w, "Blades in %q (List)\n", rackID); err != nil {
		postHttpError(ctx, tick, w, err)
		return
	}

	b := common.URLPrefix(r)

	err = dbInventory.ScanBladesInRack(rackID, func(bladeID int64) error {

		target := fmt.Sprintf("%s%d", b, bladeID)
		tracing.Infof(ctx, tick, " Listing blades '%d' at %q", bladeID, target)

		if _, err = fmt.Fprintln(w, target); err != nil {
			return httpError(ctx, w, err)
		}

		return nil
	})

	httpErrorIf(ctx, tick, w, err)
}

func handlerBladeRead(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithName("Get Blade Details"),
		tracing.AsInternal())
	defer span.End()

	// Pick up the current time to avoid repeatedly fetching the same value
	tick := clients.Tick(ctx)

	err := doSessionHeader(
		ctx, w, r,
		func(_ context.Context, session *sessions.Session) error {
			return ensureEstablishedSession(session)
		})

	if err != nil {
		postHttpError(ctx, tick, w, err)
		return
	}

	vars := mux.Vars(r)
	rackID := vars["rackID"]
	bladeName := vars["bladeID"]

	w.Header().Set("Content-Type", "application/json")

	bladeID, err := strconv.ParseInt(bladeName, 10, 64)
	if err != nil {
		postHttpError(ctx, tick, w, &HTTPError{
			SC:   http.StatusBadRequest,
			Base: err,
		})

		return
	}

		blade, err := dbInventory.GetBlade(rackID, bladeID)

	if err != nil {
		postHttpError(ctx, tick, w, err)
		return
	}

	tracing.Infof(ctx, tick, "Returning details for blade %d  in rack %q:  %v", bladeID, rackID, blade)

	p := jsonpb.Marshaler{}
	err = p.Marshal(w, blade)

	httpErrorIf(ctx, tick, w, err)
}
