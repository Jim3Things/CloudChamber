// This module provides the basic client methods for using the simulated time
// service.

package clients

import (
	"context"
	"time"

	"github.com/golang/protobuf/ptypes/duration"
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
	Time *ct.Timestamp
	Err  error
}

// Store the information needed to be able to connect to the Stepper service.
func InitTimestamp(name string, opts ...grpc.DialOption) {
	dialName = name

	dialOpts = append(dialOpts, opts...)
}

// Set the stepper policy
func SetPolicy(policy pb.StepperPolicy, delay *duration.Duration, match int64) error {
	ctx, conn, err := connect()
	if err != nil {
		return err
	}

	defer func() { _ = conn.Close() }()

	client := pb.NewStepperClient(conn)

	_, err = client.SetPolicy(
		ctx,
		&pb.PolicyRequest{
			Policy:        policy,
			MeasuredDelay: delay,
			MatchEpoch:    match,
		})

	return err
}

// Advance the simulated time, assuming that the policy mode is manual
func Advance() error {
	ctx, conn, err := connect()
	if err != nil {
		return err
	}

	defer func() { _ = conn.Close() }()

	client := pb.NewStepperClient(conn)

	_, err = client.Step(ctx, &pb.StepRequest{})

	return err
}

// Get the current simulated time.
func Now() (*ct.Timestamp, error) {
	ctx, conn, err := connect()
	if err != nil {
		return nil, err
	}

	defer func() { _ = conn.Close() }()

	client := pb.NewStepperClient(conn)

	return client.Now(ctx, &pb.NowRequest{})
}

// Delay until the simulated time meets or exceeds the specified deadline.
// Completion is asynchronous, even if no delay is required.
func After(deadline *ct.Timestamp) (<-chan TimeData, error) {
	ch := make(chan TimeData)

	go func(res chan<- TimeData) {
		ctx, conn, err := connect()
		if err != nil {
			res <- TimeData{
				Time: nil,
				Err:  err,
			}
			return
		}

		defer func() { _ = conn.Close() }()

		client := pb.NewStepperClient(conn)

		rsp, err := client.Delay(ctx, &pb.DelayRequest{AtLeast: deadline, Jitter: 0})

		if err != nil {
			res <- TimeData{Time: nil, Err: err}
			return
		}
		res <- TimeData{Time: rsp, Err: nil}
	}(ch)

	return ch, nil
}

func Status() (*pb.StatusResponse, error) {
	ctx, conn, err := connect()
	if err != nil {
		return nil, err
	}

	defer func() { _ = conn.Close() }()

	client := pb.NewStepperClient(conn)

	return client.GetStatus(ctx, &pb.GetStatusRequest{})
}

// Reset the simulated time back to its starting state, including reverting all
// policies back to their default.  This is used by unit tests to ensure a well
// known starting state for a test.
func Reset() error {
	ctx, conn, err := connect()
	if err != nil {
		return err
	}

	defer func() { _ = conn.Close() }()

	client := pb.NewStepperClient(conn)

	_, err = client.Reset(ctx, &pb.ResetRequest{})
	return err
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
