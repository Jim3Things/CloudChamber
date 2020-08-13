// This module contains the routines etc to implement the frontend handlers for the inventory operators
// part of the API
//

package frontend

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/sessions"

	"github.com/golang/protobuf/jsonpb"
	"github.com/gorilla/mux"

	"github.com/Jim3Things/CloudChamber/internal/tracing"
	st "github.com/Jim3Things/CloudChamber/internal/tracing/server"
	"github.com/Jim3Things/CloudChamber/pkg/protos/common"
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
	_ = st.WithSpan(context.Background(), tracing.MethodName(1), func(ctx context.Context) (err error) {

		err = doSessionHeader(ctx, w, r, func(_ context.Context, session *sessions.Session) error {
			return ensureEstablishedSession(ctx, session)
		})
		if err != nil {
			return httpError(ctx, w, err)
		}

		w.Header().Set("Content-Type", "application/json")

		maxBlades, maxCapacity := dbInventory.GetMemoData()
		res := &pb.ExternalZoneSummary{
			Racks:         make(map[string]*pb.ExternalRackSummary),
			MaxBladeCount: maxBlades,
			MaxCapacity:   maxCapacity,
		}

		st.Infof(
			ctx,
			tick(),
			"Listing all %d racks, max blades/rack=%d, max blade capacity=%v",
			len(res.Racks),
			res.MaxBladeCount,
			res.MaxCapacity)

		b := r.URL.String()
		if !strings.HasSuffix(b, "/") {
			b += "/"
		}

		err = dbInventory.Scan(func(name string, memo *common.BladeCapacity) (err error) {
			target := fmt.Sprintf("%s%s", b, name)

			res.Racks[name] = &pb.ExternalRackSummary{
				Name: name,
				Uri:  target,
			}

			st.Infof(ctx, tick(), "   Listing rack %q at %q", name, target)

			return nil
		})
		if err != nil {
			return httpError(ctx, w, err)
		}

		p := jsonpb.Marshaler{}
		return p.Marshal(w, res)
	})
}

func handlerRackRead(w http.ResponseWriter, r *http.Request) {
	_ = st.WithSpan(context.Background(), tracing.MethodName(1), func(ctx context.Context) (err error) {
		err = doSessionHeader(ctx, w, r, func(_ context.Context, session *sessions.Session) error {
			return ensureEstablishedSession(ctx, session)
		})
		if err != nil {
			return httpError(ctx, w, err)
		}

		vars := mux.Vars(r)
		rackID := vars["rackID"]

		u, err := dbInventory.Get(rackID)
		if err != nil {
			return httpError(ctx, w, err)
		}

		w.Header().Set("Content-Type", "application/json")

		st.Infof(ctx, tick(), "Returning details for rack %q: %v", rackID, u)

		// Get the user entry, and serialize it to json
		// (export userPublic to json and return that as the body)
		p := jsonpb.Marshaler{}
		return p.Marshal(w, u)
	})
}

func handlerBladesList(w http.ResponseWriter, r *http.Request) {
	_ = st.WithSpan(context.Background(), tracing.MethodName(1), func(ctx context.Context) (err error) {
		err = doSessionHeader(ctx, w, r, func(_ context.Context, session *sessions.Session) error {
			return ensureEstablishedSession(ctx, session)
		})
		if err != nil {
			return httpError(ctx, w, err)
		}

		vars := mux.Vars(r)
		rackID := vars["rackID"] // captured the key value in rackID variable

		if _, err = fmt.Fprintf(w, "Blades in %q (List)\n", rackID); err != nil {
			return httpError(ctx, w, err)
		}
		b := r.URL.String()
		if !strings.HasSuffix(b, "/") {
			b += "/"
		}
		return dbInventory.ScanBladesInRack(rackID, func(bladeID int64) error {

			target := fmt.Sprintf("%s%d", b, bladeID)
			st.Infof(ctx, tick(), " Listing blades '%d' at %q", bladeID, target)

			if _, err = fmt.Fprintln(w, target); err != nil {
				return httpError(ctx, w, err)
			}
			return nil
		})
	})
}

func handlerBladeRead(w http.ResponseWriter, r *http.Request) {
	_ = st.WithSpan(context.Background(), tracing.MethodName(1), func(ctx context.Context) (err error) {
		err = doSessionHeader(ctx, w, r, func(_ context.Context, session *sessions.Session) error {
			return ensureEstablishedSession(ctx, session)
		})
		if err != nil {
			return httpError(ctx, w, err)
		}

		vars := mux.Vars(r)
		rackID := vars["rackID"]
		bladeName := vars["bladeID"]

		w.Header().Set("Content-Type", "application/json")

		bladeID, err := strconv.ParseInt(bladeName, 10, 64)
		if err != nil {
			return httpError(ctx, w, &HTTPError{
				SC:   http.StatusBadRequest,
				Base: err,
			})
		}
		blade, err := dbInventory.GetBlade(rackID, bladeID)
		if err != nil {
			return httpError(ctx, w, err)
		}
		st.Infof(ctx, tick(), "Returning details for blade %d  in rack %q:  %v", bladeID, rackID, blade)

		p := jsonpb.Marshaler{}
		return p.Marshal(w, blade)
	})
}

func ReadInventoryDefinition ()