package inventory

import (
	"context"

	"google.golang.org/grpc"

	"github.com/Jim3Things/CloudChamber/internal/config"
	"github.com/Jim3Things/CloudChamber/internal/tracing"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/services"
)

type server struct {
	pb.UnimplementedInventoryServer

	racks map[string]*rack
}

func Register(svc *grpc.Server, cfg *config.GlobalConfig) error {
	s := &server {
		racks: make(map[string]*rack),
	}

	if err := s.initializeRacks(cfg.Inventory.InventoryDefinition); err != nil {
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
	ctx, span := tracing.StartSpan(context.Background(), tracing.WithName("Initializing the simulated inventory"))
	defer span.End()

	zone, err := config.ReadInventoryDefinition(path)
	if err != nil {
		return err
	}

	for name, r := range zone.Racks {
		// For each rack, create a rack item, supplying the tor, pdu, and blade
		tracing.Info(ctx, "Adding rack %q", name)
		s.racks[name] = newRack(ctx, name, r)

		// Start each rack (this gives us a channel and a goroutine)
		if err = s.racks[name].start(ctx); err != nil {
			return err
		}
	}

	return nil
}
