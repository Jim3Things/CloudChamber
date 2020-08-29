package clients

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/Jim3Things/CloudChamber/pkg/protos/log"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/trace_sink"
)

var (
	dialName string
	dialOpts []grpc.DialOption
)

// TraceData contains the 'GetAfter' response, or an error.
type TraceData struct {
	Traces *pb.GetAfterResponse
	Err error
}

// InitSinkClient stores the information needed to be able to connect to the Stepper service.
func InitSinkClient(name string, opts ...grpc.DialOption) {
	dialName = name

	dialOpts = append(dialOpts, opts...)
}

// Reset forcibly resets the trace sink to its initial state.  This is intended
// to support unit tests
func Reset() error {
	ctx, conn, err := connect()
	if err != nil {
		return err
	}

	defer func() { _ = conn.Close() }()

	client := pb.NewTraceSinkClient(conn)

	_, err = client.Reset(ctx, &pb.ResetRequest{})

	return err
}

// Append adds a log entry to the trace sink
func Append(entry *log.Entry) error {
	ctx, conn, err := connect()
	if err != nil {
		return err
	}

	defer func() { _ = conn.Close() }()

	client := pb.NewTraceSinkClient(conn)

	_, err = client.Append(ctx, &pb.AppendRequest{Entry: entry})
	return err
}

// GetPolicy obtains the current trace sink policy and returns it
func GetPolicy() (*pb.GetPolicyResponse, error) {
	ctx, conn, err := connect()
	if err != nil {
		return nil, err
	}

	defer func() { _ = conn.Close() }()

	client := pb.NewTraceSinkClient(conn)

	policy, err := client.GetPolicy(ctx, &pb.GetPolicyRequest{})

	return policy, err
}

// GetTraces retrieves up to the specified limit of trace entries, from the
// specified starting point.  It will always wait for at least one
// non-internal entry before returning.
func GetTraces(start int64, maxCount int64) <-chan TraceData {
	ch := make(chan TraceData)

	go func(res chan<- TraceData) {
		ctx, conn, err := connect()
		if err != nil {
			res <- TraceData{
				Traces: nil,
				Err:  err,
			}
			return
		}

		defer func() { _ = conn.Close() }()

		client := pb.NewTraceSinkClient(conn)

		rsp, err := client.GetAfter(ctx, &pb.GetAfterRequest{
			Id:         start,
			MaxEntries: maxCount,
			Wait:       true,
		})

		res <- TraceData{Traces: rsp, Err: nil}
	}(ch)

	return ch
}

// connect is a helper function that sets up the communication context for
// the grpc client.
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
