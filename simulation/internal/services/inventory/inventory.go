package inventory

import (
	"context"

	"google.golang.org/grpc"

	ic "github.com/Jim3Things/CloudChamber/simulation/internal/clients/inventory"
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

	if err := s.initializeRacks(cfg.Inventory.InventoryDefinition); err != nil {
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

func (s *server) initializeRacks(path string) error {
	ctx, span := tracing.StartSpan(
		context.Background(),
		tracing.WithName("Initializing the simulated inventory"),
		tracing.WithContextValue(ts.EnsureTickInContext))
	defer span.End()

	zone, err := ic.ReadInventoryDefinition(ctx, path)
	if err != nil {
		return err
	}

	for name, r := range zone.Racks {
		// For each rack, create a rack item, supplying the tor, pdu, and blade
		tracing.Info(ctx, "Adding rack %q", name)
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