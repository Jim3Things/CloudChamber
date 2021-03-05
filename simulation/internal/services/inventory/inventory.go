package inventory

import (
	"context"

	"google.golang.org/grpc"

	ic "github.com/Jim3Things/CloudChamber/simulation/internal/clients/inventory"
	st "github.com/Jim3Things/CloudChamber/simulation/internal/clients/store"
	ts "github.com/Jim3Things/CloudChamber/simulation/internal/clients/timestamp"
	tsc "github.com/Jim3Things/CloudChamber/simulation/internal/clients/trace_sink"
	"github.com/Jim3Things/CloudChamber/simulation/internal/common"
	"github.com/Jim3Things/CloudChamber/simulation/internal/config"
	"github.com/Jim3Things/CloudChamber/simulation/internal/services/inventory/messages"
	"github.com/Jim3Things/CloudChamber/simulation/internal/sm"
	"github.com/Jim3Things/CloudChamber/simulation/internal/tracing"
	ct "github.com/Jim3Things/CloudChamber/simulation/internal/tracing/client"
	pb "github.com/Jim3Things/CloudChamber/simulation/pkg/protos/services"
)

type server struct {
	pb.UnimplementedInventoryServer

	racks map[string]*Rack

	timers *ts.Timers

	store *st.Store
	inventory *ic.Inventory
	inventoryRoot *ic.Root
	inventoryRegion *ic.Region
	inventoryZone *ic.Zone
}

func Register(svc *grpc.Server, cfg *config.GlobalConfig) error {
	if err := ts.InitTimestamp(
		cfg.SimSupport.EP.String(),
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(ct.Interceptor)); err != nil {
		return err
	}

	if err := tsc.InitSinkClient(
		cfg.SimSupport.EP.String(),
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(ct.Interceptor)); err != nil {
		return err
	}

	timers := ts.NewTimers(
		cfg.SimSupport.EP.String(),
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(ct.Interceptor))

	s := &server{
		racks:  make(map[string]*Rack),
		timers: timers,
	}

	if err := s.initialiseInventory(cfg); err != nil {
		return err
	}

	if err := s.initializeRacks(); err != nil {
		return err
	}

	if err := s.startBlades(); err != nil {
		return err
	}

	pb.RegisterInventoryServer(svc, s)

	return nil
}

func (s *server) Repair(context.Context, *pb.InventoryRepairMsg) (*pb.InventoryRepairResp, error) {
	// Figure out the rack to send it to

	// -- fail if it is not found.

	// Forward to the channel

	// Respond with success

	return nil, nil
}

func (s *server) GetCurrentStatus(context.Context, *pb.InventoryStatusMsg) (*pb.InventoryStatusResp, error) {
	return nil, nil
}

func (s *server) initialiseInventory(cfg *config.GlobalConfig) error {
	// Initialize the underlying store
	//
	ctx := context.Background()

	st.Initialize(cfg)
	s.store = st.NewStore()
	s.inventory = ic.NewInventory(cfg, s.store)
	
	// NOTE TO Jim
	//
	// We need to decide who gets to write the defined inventory to the
	// store. At present this is done by the frontend process, but it may
	// well be that the inventoryd process should be doing that. 
	//
	//
	// NOTE TO Mike
	//
	// Need to not have Start() also do the load. That should/is a separate call.
	//
	if err := s.inventory.Start(ctx); err != nil {
		return err
	}

	if err := s.inventory.UpdateInventoryDefinition(ctx, cfg.Inventory.InventoryDefinition); err != nil {
		return err
	}

	// At this point the inventory is loaded from the configured file
	// into the store. All queries for the current defnitions should
	// now be performed against the store.
	//
	return nil
}

func (s *server) initializeRacks() error {
	ctx, span := tracing.StartSpan(
		context.Background(),
		tracing.WithName("Initializing the simulated inventory"),
		tracing.WithContextValue(ts.EnsureTickInContext))
	defer span.End()

	const (
		// These will be whatever is defined in the inventory. Or they can be
		// discovered by scanning for regions etc from the root of the definition
		//
		regionName = "region1"
		zoneName   = "zone1"
	)

	zone, err := s.inventory.NewZone(ic.DefinitionTable, regionName, zoneName)

	_, racks, err := zone.FetchChildren(ctx)

	if err != nil {
		return err
	}

	s.inventoryZone = zone

	for name, rack := range *racks {
		// For each rack, create a rack item, supplying the tor, pdu, and blade
		tracing.Info(ctx, "Adding rack %q", name)

		r, err := rack.GetDefinitionRackWithChildren(ctx)

		if err != nil {
			return err
		}

		s.racks[name] = newRack(ctx, name, r, s.timers)

		// Start each rack (this gives us a channel and a goroutine)
		if err = s.racks[name].start(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (s *server) startBlades() error {
	ctx, span := tracing.StartSpan(
		context.Background(),
		tracing.WithName("Booting the simulated inventory"),
		tracing.WithContextValue(ts.EnsureTickInContext))
	defer span.End()

	for name, r := range s.racks {
		tracing.Info(ctx, "Booting the blades in rack %q", name)
		for _, b := range r.blades {
			rsp := make(chan *sm.Response)

			r.Receive(
				messages.NewSetPower(ctx, b.me(), common.TickFromContext(ctx), true, rsp))

			<-rsp
		}
	}

	return nil
}
