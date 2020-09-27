package clients

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// gprcConnect initiates a connection to target client, with the expected
// CloudChamber metadata.
func gprcConnect(ctx context.Context, dialName string, dialOpts []grpc.DialOption) (context.Context, *grpc.ClientConn, error) {
	conn, err := grpc.Dial(dialName, dialOpts...)

	if err != nil {
		return nil, nil, err
	}

	// TODO: These are placeholder metadata items.  Need to provide the actual ones
	//       we intend to use.
	md := metadata.Pairs(
		"timestamp", time.Now().Format(time.StampNano),
		"client-id", "web-api-client-us-east-1",
		"user-id", "some-test-user-id",
	)

	return metadata.NewOutgoingContext(ctx, md), conn, nil
}


