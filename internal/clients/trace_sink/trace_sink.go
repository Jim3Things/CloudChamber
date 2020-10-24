package trace_sink

import (
	"context"

	"google.golang.org/grpc"

	"github.com/Jim3Things/CloudChamber/pkg/protos/log"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/services"
)

var (
	dialName string
	dialOpts []grpc.DialOption
)

// TraceData contains the 'GetAfter' response, or an error.
type TraceData struct {
	Traces *pb.GetAfterResponse
	Err    error
}

// InitSinkClient stores the information needed to be able to connect to the Stepper service.
func InitSinkClient(name string, opts ...grpc.DialOption) {
	dialName = name

	dialOpts = append([]grpc.DialOption{}, opts...)
}

// Reset forcibly resets the trace sink to its initial state.  This is intended
// to support unit tests
func Reset(ctx context.Context) error {
	conn, err := grpc.Dial(dialName, dialOpts...)
	if err != nil {
		return err
	}

	defer func() { _ = conn.Close() }()

	client := pb.NewTraceSinkClient(conn)

	_, err = client.Reset(ctx, &pb.ResetRequest{})

	return err
}

// Append adds a log entry to the trace sink
func Append(ctx context.Context, entry *log.Entry) error {
	conn, err := grpc.Dial(dialName, dialOpts...)
	if err != nil {
		return err
	}

	defer func() { _ = conn.Close() }()

	client := pb.NewTraceSinkClient(conn)

	_, err = client.Append(ctx, &pb.AppendRequest{Entry: entry})
	return err
}

// GetPolicy obtains the current trace sink policy and returns it
func GetPolicy(ctx context.Context) (*pb.GetPolicyResponse, error) {
	conn, err := grpc.Dial(dialName, dialOpts...)
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
func GetTraces(ctx context.Context, start int64, maxCount int64) <-chan TraceData {
	ch := make(chan TraceData)

	go func(res chan<- TraceData) {
		conn, err := grpc.Dial(dialName, dialOpts...)
		if err != nil {
			res <- TraceData{
				Traces: nil,
				Err:    err,
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
