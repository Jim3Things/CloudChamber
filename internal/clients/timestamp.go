// This module provides the basic client methods for using the simulated time
// service.

package clients

import (
    "context"
    "time"

    "github.com/golang/protobuf/ptypes/empty"
    "google.golang.org/grpc/metadata"

    pb "github.com/Jim3Things/CloudChamber/pkg/protos/Stepper"
    ct "github.com/Jim3Things/CloudChamber/pkg/protos/common"

    "google.golang.org/grpc"
)

var dialName string
var dialOpts []grpc.DialOption

// Defines the value returned from a delay wait.  This is more than the
// simple timestamp inasmuch as the delay call can fail asynchronously.
type TimeData struct {
    time *ct.Timestamp
    err error
}

// Store the information needed to be able to connect to the Stepper service.
func InitTimestamp(name string, opts ...grpc.DialOption) {
    dialName = name
    dialOpts = append(dialOpts, opts...)
}

// Get the current simulated time.
func Now() (*ct.Timestamp, error) {
    ctx, conn, err := connect()
    if err != nil {
        return nil, err
    }

    defer func() { _ = conn.Close() }()

    client := pb.NewStepperClient(conn)

    return client.Now(ctx, &empty.Empty{})
}

// Delay until the simulated time meets or exceeds the specified deadline.
// Completion is asynchronous, even if no delay is required.
func After(deadline *ct.Timestamp) (<-chan TimeData, error) {
    ch := make(chan TimeData)

    go func(res chan<- TimeData) {
        ctx, conn, err := connect()
        if err != nil {
            res <- TimeData {
                time: nil,
                err:  err,
            }
            return
        }

        defer func() { _ = conn.Close() }()

        client := pb.NewStepperClient(conn)

        rsp, err := client.Delay(ctx, &pb.DelayRequest{AtLeast: deadline, Jitter: 0})

        if err != nil {
            res <- TimeData{time: nil, err: err }
            return
        }
        res <- TimeData{time: rsp, err: nil}
    }(ch)

    return ch, nil
}

// Helper function to connect to the stepper client.
func connect() (context.Context, *grpc.ClientConn, error) {
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
    ctx := metadata.NewOutgoingContext(context.Background(), md)

    return ctx, conn, nil
}