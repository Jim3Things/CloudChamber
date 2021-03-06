package inventory

import (
	"context"

	"google.golang.org/grpc"

	ic "github.com/Jim3Things/CloudChamber/simulation/internal/clients/inventory"
	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/limits"
	ns "github.com/Jim3Things/CloudChamber/simulation/internal/clients/namespace"
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

	store     *st.Store
	inventory *ic.Inventory
}

func Register(svc *grpc.Server, cfg *config.GlobalConfig) error {
	if err := ts.InitTimestamp(
		cfg.SimSupport.EP.String(),
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(ct.Interceptor),
		grpc.WithConnectParams(limits.BackoffSettings),
	); err != nil {
		return err
	}

	if err := tsc.InitSinkClient(
		cfg.SimSupport.EP.String(),
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(ct.Interceptor),
		grpc.WithConnectParams(limits.BackoffSettings),
	); err != nil {
		return err
	}

	timers := ts.NewTimers(
		cfg.SimSupport.EP.String(),
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(ct.Interceptor),
		grpc.WithConnectParams(limits.BackoffSettings),
	)

	s := &server{
		racks:  make(map[string]*Rack),
		timers: timers,
	}

	if err := s.initializeInventory(cfg); err != nil {
		return err
	}

	if err := s.initializeRacks(cfg); err != nil {
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

func (s *server) initializeInventory(cfg *config.GlobalConfig) error {
	ctx, span := tracing.StartSpan(
		context.Background(),
		tracing.WithName("Initializing the simulated inventory"),
		tracing.WithContextValue(ts.EnsureTickInContext))
	defer span.End()

	// Initialize the underlying store
	//
	st.Initialize(ctx, cfg)
	s.store = st.NewStore()
	s.inventory = ic.NewInventory(cfg, s.store)

	if err := s.inventory.Start(ctx); err != nil {
		return err
	}

	// Load/update the store inventory definitions from the configured file.
	// Once complete, all queries for the current definitions should be
	// performed against the store.
	//
	if err := s.inventory.UpdateInventoryDefinition(ctx, cfg.Inventory.InventoryDefinition); err != nil {
		return err
	}

	return nil
}

func (s *server) initializeRacks(cfg *config.GlobalConfig) error {
	ctx, span := tracing.StartSpan(
		context.Background(),
		tracing.WithName("Initializing the simulated inventory racks"),
		tracing.WithContextValue(ts.EnsureTickInContext))
	defer span.End()

	const (
		// These will match whatever is defined in the inventory. Or they can be
		// discovered by scanning for regions etc from the root of the definition
		// table.
		//
		// If the inventory service is intended to provide service for a single
		// zone as defined in (say) the configuration, then these values should
		// be updated to reflect that configuration information.
		//
		regionName = "standard"
		zoneName   = "standard"
	)

	zone, err := s.inventory.NewZone(ns.DefinitionTable, regionName, zoneName)

	if err != nil {
		return err
	}

	_, racks, err := zone.FetchChildren(ctx)

	if err != nil {
		return err
	}

	for _, rack := range *racks {
		// For each rack, create a rack item, supplying the tor, pdu, and blade
		tracing.Info(ctx, "Adding rack %s", rack.Key)

		r, err := rack.GetDefinitionRackWithChildren(ctx)

		if err != nil {
			return err
		}

		s.racks[rack.Key] = newRack(
			ctx,
			rack.Key,
			r,
			cfg,
			rack.KeyIndexPdu,
			rack.KeyIndexTor,
			rack.KeyIndexBlade,
			s.timers)

		// Start each rack (this gives us a channel and a goroutine)
		if err = s.racks[rack.Key].start(ctx); err != nil {
			return err
		}
	}

	return nil
}

// startBlades begins the mocked repair actions to move from an initial rack
// appearing to the point where the blades are booted.  This first step involves
// powering on each blade.  It then proceeds to connectBladeAfterPower.
//
// As this is a purely temporary mock, errors are currently ignored in the power
// on operation.
func (s *server) startBlades() error {
	ctx, span := tracing.StartSpan(
		context.Background(),
		tracing.WithName("Booting the simulated inventory, step 1: powering on"),
		tracing.WithContextValue(ts.EnsureTickInContext))
	defer span.End()

	for name, r := range s.racks {
		tracing.Info(ctx, "Applying power to the blades in rack %q", name)
		for _, b := range r.blades {
			rsp := make(chan *sm.Response)

			r.Receive(
				messages.NewSetPower(ctx, b.me(), common.TickFromContext(ctx), true, rsp))

			go s.connectBladeAfterPower(r, b, rsp)
		}
	}

	return nil
}

// connectBladeAfterPower waits for the power operation to complete, and then
// initiates a network connection operation.  It exits the goroutine when the
// network connection operation completes.  As with startBlades, errors from the
// network operation are ignored.
func (s *server) connectBladeAfterPower(r *Rack, b *blade, rsp chan *sm.Response) {
	<- rsp

	ctx, span := tracing.StartSpan(
		context.Background(),
		tracing.WithName("Connecting the network for %q", b.me().Describe()),
		tracing.WithContextValue(ts.EnsureTickInContext))
	defer span.End()

	rsp = make(chan *sm.Response)
	r.Receive(
		messages.NewSetConnection(ctx, b.me(), common.TickFromContext(ctx), true, rsp))

	<- rsp
}
