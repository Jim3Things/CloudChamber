package inventory

import (
	"context"

	"google.golang.org/grpc"

	"github.com/Jim3Things/CloudChamber/internal/config"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/services"
)

type server struct {
	pb.UnimplementedInventoryServer

	racks map[string]rack
}

func Register(svc *grpc.Server, cfg *config.GlobalConfig) error {
	s := &server {

	}

	// Read the configuration to get the inventory link

	// Read the inventory

	// For each rack, create a rack item, supplying the tor, pdu, and blade

	// Start each rack (this gives us a channel and a goroutine)

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

