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

	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/inventory"
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
		return
	}

	zoneDetails, err := dbInventory.GetZoneDetails(ctx, inventory.DefaultRegion, inventory.DefaultZone)
	if err != nil {
		postHTTPError(ctx, w, errors.ErrInventoryNotAvailable)
		return
	}

	res := &pb.External_ZoneSummary{
		Racks:         make(map[string]*pb.External_RackSummary, memoData.RackCount),
		MaxBladeCount: int64(memoData.MaxBladeCount),
		MaxCapacity:   memoData.MaxCapacity,
		MaxTorCount:   int64(memoData.MaxTorCount),
		MaxPduCount:   int64(memoData.MaxPduCount),
		MaxConnectors: int64(memoData.MaxConnectors),
		Name:          inventory.DefaultZone,
		Details:       zoneDetails,
	}

	tracing.Info(
		ctx,
		"Listing all %d racks, %s",
		memoData.RackCount,
		memoData.RackSizeSummary.String())

	b := common.URLPrefix(r)

	// We're going to group all racks listed together into a single trace entry.
	//
	// msg is the formatted text built so far.
	// nl is the connector between racks.  It starts as nothing and is changed
	// to a new-line once anything has been stored into the formatted text.
	msg := ""
	nl := ""

	err = dbInventory.ScanRacksInZone(
		ctx,
		inventory.DefaultRegion,
		inventory.DefaultZone,
		func(name string) error {
			target := fmt.Sprintf("%s%s/", b, name)

			res.Racks[name] = &pb.External_RackSummary{Uri: target}

			msg = fmt.Sprintf("%s%sListing rack %q at %q", msg, nl, name, target)
			nl = "\n"

			return nil
		})

	tracing.Info(ctx, msg)

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

	rd, err := dbInventory.GetRackInZone(ctx, inventory.DefaultRegion, inventory.DefaultZone, rackID)
	if err != nil {
		postHTTPError(ctx, w, err)
		return
	}

	rack, err := transformRack(rd)
	if err != nil {
		postHTTPError(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	tracing.UpdateSpanName(ctx, "Get Details for rack %q", rackID)
	tracing.Info(ctx, "Rack %q details: %v", rackID, rack)

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

	tracing.UpdateSpanName(ctx, "Get List of Blades in rack %q", rackID)

	if _, err = fmt.Fprintf(w, "Blades in %q (List)\n", rackID); err != nil {
		postHTTPError(ctx, w, err)
		return
	}

	b := common.URLPrefix(r)

	err = dbInventory.ScanBladesInRack(
		ctx,
		inventory.DefaultRegion,
		inventory.DefaultZone,
		rackID,
		func(bladeID int64) error {

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

	blade, err := dbInventory.GetBlade(
		ctx,
		inventory.DefaultRegion,
		inventory.DefaultZone,
		rackID,
		bladeID)
	if err != nil {
		postHTTPError(ctx, w, err)
		return
	}

	tracing.UpdateSpanName(ctx, "Get Details for blade %d in rack %q", bladeID, rackID)
	tracing.Info(ctx, "Returning details for blade %d in rack %q:  %v", bladeID, rackID, blade)

	p := jsonpb.Marshaler{}
	err = p.Marshal(w, blade)

	httpErrorIf(ctx, w, err)
}

// transformRack converts the internal rack states into an exportable external
// view.
func transformRack(rd *pb.Definition_Rack) (*pb.External_Rack, error) {
	rack := &pb.External_Rack{
		Details:    rd.GetDetails(),
		Pdu:        &pb.External_Pdu{},
		Tor:        &pb.External_Tor{},
		Blades:     make(map[int64]*pb.BladeCapacity),
		Tors:       make(map[int64]*pb.External_Tor),
		Pdus:       make(map[int64]*pb.External_Pdu),
		FullBlades: make(map[int64]*pb.External_Blade),
	}

	for i, blade := range rd.Blades {
		rack.Blades[i] = blade.GetCapacity()
		rack.FullBlades[i] = &pb.External_Blade{
			Details:       blade.GetDetails(),
			Capacity:      blade.GetCapacity(),
			BootOnPowerOn: blade.BootOnPowerOn,
			BootInfo:      blade.GetBootInfo(),
			Observed:      fakeBladeObserved(),
		}
	}

	for i, tor := range rd.Tors {
		 t := &pb.External_Tor{
			Details: tor.GetDetails(),
			Ports:   make(map[int64]*pb.External_Tor_Port),
			Observed: fakeTorObserved(),
		}

		for k, port := range tor.GetPorts() {
			t.Ports[k] = &pb.External_Tor_Port{
				Port: port,
				Observed: &pb.External_ObservedCable{
					At:        0,
					SmState:   fakeCableState(),
					EnteredAt: 0,
				}}
		}

		rack.Tors[i] = t
	}

	for i, pdu := range rd.Pdus {
		p := &pb.External_Pdu{
			Details: pdu.GetDetails(),
			Ports:   make(map[int64]*pb.External_Pdu_Port),
			Observed: fakePduObserved(),
		}

		for k, port := range pdu.GetPorts() {
			p.Ports[k] = &pb.External_Pdu_Port{
				Port:     port,
				Observed: &pb.External_ObservedCable{
					At:        0,
					SmState:   fakeCableState(),
					EnteredAt: 0,
				},
			}
		}

		rack.Pdus[i] = p
	}

	return rack, nil
}

// +++ Temporary observed state creation

// This temporary feature just cycles through the set of possible blade SM
// states, and fakes up an observed blade state from that.

var bladeState = pb.BladeState_start

func fakeBladeObserved() *pb.External_Blade_ObservedState {
	c := bladeState

	state := c
	c++
	if c > 11 {
		c = pb.BladeState_start
	}

	bladeState = c

	return &pb.External_Blade_ObservedState{
		At:        10,
		SmState:   state,
		EnteredAt: 5,
	}
}

var pduState = pb.PduState_working

func fakePduObserved() *pb.External_Pdu_ObservedState {
	c := pduState

	state := c
	c++
	if c > 3 {
		c = pb.PduState_working
	}

	pduState = c

	return &pb.External_Pdu_ObservedState{
		At:        10,
		SmState:   state,
		EnteredAt: 1,
	}
}

var torState = pb.TorState_working

func fakeTorObserved() *pb.External_Tor_ObservedState {
	c := torState

	state := c
	c++

	if c > 2 {
		c = pb.TorState_working
	}

	torState = c

	return &pb.External_Tor_ObservedState{
		At:        10,
		SmState:   state,
		EnteredAt: 1,
	}
}

var cableState = pb.CableState_on

func fakeCableState() pb.CableState_SM {
	c := cableState

	state := c
	c++

	if c > 2 {
		c = pb.CableState_on
	}

	cableState = c

	return state
}

// --- Temporary observed state creation
