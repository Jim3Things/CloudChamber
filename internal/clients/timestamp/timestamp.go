// This module provides the basic client methods for using the simulated time
// service.

package timestamp

import (
	"context"

	"github.com/golang/protobuf/ptypes/duration"

	"github.com/Jim3Things/CloudChamber/internal/common"
	ct "github.com/Jim3Things/CloudChamber/pkg/protos/common"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/services"

	"google.golang.org/grpc"
)

var (
	dialName string
	dialOpts []grpc.DialOption
)

// TimeData defines the value returned from a delay wait.  This is more than
// the simple timestamp inasmuch as the delay call can fail asynchronously.
type TimeData struct {
	Time *ct.Timestamp
	Err  error
}

// InitTimestamp stores the information needed to be able to GrpcConnect to the
// Stepper service.
func InitTimestamp(name string, opts ...grpc.DialOption) {
	dialName = name

	dialOpts = append(dialOpts, opts...)
}

// SetPolicy sets the stepper policy
func SetPolicy(ctx context.Context, policy pb.StepperPolicy, delay *duration.Duration, match int64) error {
	conn, err := grpc.Dial(dialName, dialOpts...)
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
func Advance(ctx context.Context) error {
	conn, err := grpc.Dial(dialName, dialOpts...)
	if err != nil {
		return err
	}

	defer func() { _ = conn.Close() }()

	client := pb.NewStepperClient(conn)

	_, err = client.Step(ctx, &pb.StepRequest{})

	return err
}

// Now gets the current simulated time.
func Now(ctx context.Context) (*ct.Timestamp, error) {
	conn, err := grpc.Dial(dialName, dialOpts...)
	if err != nil {
		return nil, err
	}

	defer func() { _ = conn.Close() }()

	client := pb.NewStepperClient(conn)

	return client.Now(ctx, &pb.NowRequest{})
}

// After delays execution until the simulated time meets or exceeds the
// specified deadline.  Completion is asynchronous, even if no delay is
// required.
func After(ctx context.Context, deadline *ct.Timestamp) <-chan TimeData {
	ch := make(chan TimeData)

	go func(ctx context.Context, res chan<- TimeData) {
		conn, err := grpc.Dial(dialName, dialOpts...)
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
	}(ctx, ch)

	return ch
}

// Status retrieves the status of the Stepper service
func Status(ctx context.Context) (*pb.StatusResponse, error) {
	conn, err := grpc.Dial(dialName, dialOpts...)
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
func Reset(ctx context.Context) error {
	conn, err := grpc.Dial(dialName, dialOpts...)
	if err != nil {
		return err
	}

	defer func() { _ = conn.Close() }()

	client := pb.NewStepperClient(conn)

	_, err = client.Reset(ctx, &pb.ResetRequest{})
	return err
}

// Tick provides the current simulated time Tick, or '-1' if the simulated time
// cannot be retrieved (e.g. during startup)
func Tick(ctx context.Context) int64 {
	now, err := Now(ctx)
	if err != nil {
		return -1
	}

	return now.Ticks
}

// EnsureTickInContext checks if a simulated time tick is already present in
// the context.  If not, it stores the current simulated time.
func EnsureTickInContext(ctx context.Context) context.Context {
	if common.ContextHasTick(ctx) {
		return ctx
	}

	return common.ContextWithTick(ctx, Tick(ctx))
}

// OutsideTime forces the simulated time tick in the context to be '-1', which
// is the designator for an operation that is outside the simulated time flow.
func OutsideTime(ctx context.Context) context.Context {
	return common.ContextWithTick(ctx, -1)
}
