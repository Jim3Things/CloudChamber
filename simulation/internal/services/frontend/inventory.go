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

	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/timestamp"
	"github.com/Jim3Things/CloudChamber/simulation/internal/common"
	"github.com/Jim3Things/CloudChamber/simulation/internal/tracing"
	"github.com/Jim3Things/CloudChamber/simulation/pkg/errors"
	pb "github.com/Jim3Things/CloudChamber/simulation/pkg/protos/inventory"
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
	routeRacks.HandleFunc("/{rackID}/", handlerRackRead).Methods("GET")
	routeRacks.HandleFunc("/{rackID}/blades", handlerBladesList).Methods("GET")
	routeRacks.HandleFunc("/{rackID}/blades/", handlerBladesList).Methods("GET")
	routeRacks.HandleFunc("/{rackID}/blades/{bladeID}", handlerBladeRead).Methods("GET")
	// routeRacks.HandleFunc("/racks/{rackID}", handlerRacksDelete).Methods("DELETE", "GET")
}

func handlerRacksList(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithName("Get Cluster Inventory List of Racks"),
		tracing.WithContextValue(timestamp.EnsureTickInContext),
		tracing.AsInternal())
	defer span.End()

	err := doSessionHeader(
		ctx, w, r,
		func(_ context.Context, session *sessions.Session) error {
			return server.sessions.ensureEstablishedSession(session)
		})

	if err != nil {
		postHTTPError(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	memoData := dbInventory.GetMemoData()
	if memoData == nil {
		postHTTPError(ctx, w, errors.ErrInventoryNotAvailable)
	}

	res := &pb.External_ZoneSummary{
		Racks:         make(map[string]*pb.External_RackSummary, memoData.RackCount),
		MaxBladeCount: int64(memoData.MaxBladeCount),
		MaxCapacity:   memoData.MaxCapacity,
	}

	tracing.Info(
		ctx,
		"Listing all %d racks, max blades/rack=%d, max blade capacity=%v",
		memoData.RackCount,
		memoData.MaxBladeCount,
		&memoData.MaxCapacity)

	b := common.URLPrefix(r)

	err = dbInventory.ScanRacksInZone(defaultRegion, defaultZone, func(name string) error {
		target := fmt.Sprintf("%s%s/", b, name)

		res.Racks[name] = &pb.External_RackSummary{Uri: target}

		tracing.Info(ctx, "   Listing rack %q at %q", name, target)

		return nil
	})

	if err != nil {
		postHTTPError(ctx, w, err)
		return
	}

	p := jsonpb.Marshaler{}
	err = p.Marshal(w, res)

	httpErrorIf(ctx, w, err)
}

func handlerRackRead(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithName("Get Rack Details"),
		tracing.WithContextValue(timestamp.EnsureTickInContext),
		tracing.AsInternal())
	defer span.End()

	vars := mux.Vars(r)
	rackID := vars["rackID"]

	err := doSessionHeader(
		ctx, w, r,
		func(_ context.Context, session *sessions.Session) error {
			return server.sessions.ensureEstablishedSession(session)
		})

	if err != nil {
		postHTTPError(ctx, w, err)
		return
	}

	rack, err := dbInventory.GetRack(rackID)
	if err != nil {
		postHTTPError(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	tracing.Info(ctx, "Returning details for rack %q: %v", rackID, rack)

	p := jsonpb.Marshaler{}
	err = p.Marshal(w, rack)

	httpErrorIf(ctx, w, err)
}

func handlerBladesList(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithName("Get List of Blades in Selected Rack"),
		tracing.WithContextValue(timestamp.EnsureTickInContext),
		tracing.AsInternal())
	defer span.End()

	err := doSessionHeader(
		ctx, w, r,
		func(_ context.Context, session *sessions.Session) error {
			return server.sessions.ensureEstablishedSession(session)
		})

	if err != nil {
		postHTTPError(ctx, w, err)
		return
	}

	vars := mux.Vars(r)
	rackID := vars["rackID"] // captured the key value in rackID variable

	if _, err = fmt.Fprintf(w, "Blades in %q (List)\n", rackID); err != nil {
		postHTTPError(ctx, w, err)
		return
	}

	b := common.URLPrefix(r)

	err = dbInventory.ScanBladesInRack(rackID, func(bladeID int64) error {

		target := fmt.Sprintf("%s%d", b, bladeID)
		tracing.Info(ctx, " Listing blades '%d' at %q", bladeID, target)

		_, err = fmt.Fprintln(w, target)
		return err
	})

	httpErrorIf(ctx, w, err)
}

func handlerBladeRead(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithName("Get Blade Details"),
		tracing.WithContextValue(timestamp.EnsureTickInContext),
		tracing.AsInternal())
	defer span.End()

	err := doSessionHeader(
		ctx, w, r,
		func(_ context.Context, session *sessions.Session) error {
			return server.sessions.ensureEstablishedSession(session)
		})

	if err != nil {
		postHTTPError(ctx, w, err)
		return
	}

	vars := mux.Vars(r)
	rackID := vars["rackID"]
	bladeName := vars["bladeID"]

	w.Header().Set("Content-Type", "application/json")

	bladeID, err := strconv.ParseInt(bladeName, 10, 64)
	if err != nil {
		postHTTPError(ctx, w, &HTTPError{
			SC:   http.StatusBadRequest,
			Base: err,
		})

		return
	}

	blade, err := dbInventory.GetBlade(rackID, bladeID)

	if err != nil {
		postHTTPError(ctx, w, err)
		return
	}

	tracing.Info(ctx, "Returning details for blade %d  in rack %q:  %v", bladeID, rackID, blade)

	p := jsonpb.Marshaler{}
	err = p.Marshal(w, blade)

	httpErrorIf(ctx, w, err)
}
