package monitor

import (
    "google.golang.org/grpc"
)

// Register this with grpc as the inventory monitor service.
func Register(s *grpc.Server) {
    // TBD

    // We need to register the inventory update notification service
    // (note that there is also a client that will be used to talk to
    // the inventory)
}

