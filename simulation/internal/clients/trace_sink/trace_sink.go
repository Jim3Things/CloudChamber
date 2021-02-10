package trace_sink

import (
	"context"
	"sync"

	"google.golang.org/grpc"

	"github.com/Jim3Things/CloudChamber/simulation/pkg/errors"
	"github.com/Jim3Things/CloudChamber/simulation/pkg/protos/log"
	pb "github.com/Jim3Things/CloudChamber/simulation/pkg/protos/services"
)

type traceClient interface {
	initialize(name string, opts ...grpc.DialOption) error
	reset(ctx context.Context) error
	append(ctx context.Context, entry *log.Entry) error
	getPolicy(ctx context.Context) (*pb.GetPolicyResponse, error)
	getTraces(ctx context.Context, start int64, maxCount int64) <-chan traceData
}

var (
	tsc traceClient = &notReady{}
)

// traceData contains the 'GetAfter' response, or an error.
type traceData struct {
	Traces *pb.GetAfterResponse
	Err    error
}

type notReady struct{}

func (nrc notReady) initialize(name string, opts ...grpc.DialOption) error {
	tsc = newTraceClient(name, opts...)
	return nil
}

func (nrc notReady) reset(_ context.Context) error {
	return errors.ErrClientNotReady("tracing")
}

func (nrc notReady) append(_ context.Context, _ *log.Entry) error {
	return errors.ErrClientNotReady("tracing")
}

func (nrc notReady) getPolicy(_ context.Context) (*pb.GetPolicyResponse, error) {
	return nil, errors.ErrClientNotReady("tracing")
}

func (nrc notReady) getTraces(_ context.Context, _ int64, _ int64) <-chan traceData {
	ch := make(chan traceData)
	ch <- traceData{
		Traces: nil,
		Err:    errors.ErrClientNotReady("tracing"),
	}

	return ch
}

type activeClient struct {
	dialName string
	dialOpts []grpc.DialOption

	conn   *grpc.ClientConn
	client pb.TraceSinkClient
	m      *sync.Mutex
}

func newTraceClient(dialName string, opts ...grpc.DialOption) traceClient {
	return &activeClient{
		dialName: dialName,
		dialOpts: append([]grpc.DialOption{}, opts...),
		conn:     nil,
		client:   nil,
		m:        &sync.Mutex{},
	}
}

func (acl *activeClient) initialize(_ string, _ ...grpc.DialOption) error {
	return errors.ErrClientAlreadyInitialized("stepper")
}

func (acl *activeClient) reset(ctx context.Context) error {
	client, err := acl.dial()
	if err != nil {
		return err
	}

	_, err = client.Reset(ctx, &pb.ResetRequest{})

	return acl.cleanup(client, err)
}

func (acl *activeClient) append(ctx context.Context, entry *log.Entry) error {
	client, err := acl.dial()
	if err != nil {
		return err
	}

	_, err = client.Append(ctx, &pb.AppendRequest{Entry: entry})
	return acl.cleanup(client, err)
}

func (acl *activeClient) getPolicy(ctx context.Context) (*pb.GetPolicyResponse, error) {
	client, err := acl.dial()
	if err != nil {
		return nil, err
	}

	policy, err := client.GetPolicy(ctx, &pb.GetPolicyRequest{})

	return policy, acl.cleanup(client, err)
}

func (acl *activeClient) getTraces(ctx context.Context, start int64, maxCount int64) <-chan traceData {
	ch := make(chan traceData)

	go func(res chan<- traceData) {
		client, err := acl.dial()
		if err != nil {
			res <- traceData{
				Traces: nil,
				Err:    err,
			}
			return
		}

		rsp, err := client.GetAfter(ctx, &pb.GetAfterRequest{
			Id:         start,
			MaxEntries: maxCount,
			Wait:       true,
		})

		res <- traceData{Traces: rsp, Err: acl.cleanup(client, err)}
	}(ch)

	return ch
}

// dial abstracts the connection logic to the trace sink service.  It caches
// the connection for use in later calls in order to avoid excess transport and
// grpc connection operations.
func (acl *activeClient) dial() (pb.TraceSinkClient, error) {
	acl.m.Lock()
	defer acl.m.Unlock()

	if acl.conn == nil {
		conn, err := grpc.Dial(acl.dialName, acl.dialOpts...)
		if err != nil {
			return nil, err
		}

		acl.conn = conn
		acl.client = pb.NewTraceSinkClient(conn)
	}

	return acl.client, nil
}

// cleanup ensures that if an error has occurred the cached connection is
// cleared.
func (acl *activeClient) cleanup(client pb.TraceSinkClient, err error) error {
	acl.m.Lock()
	defer acl.m.Unlock()

	// Clear the connection iff we had an error, the connection has not been
	// cleaned up already, and we had an error against the current client.
	if err != nil && acl.conn != nil && client == acl.client {
		_ = acl.conn.Close()

		acl.conn = nil
		acl.client = nil
	}

	return err
}

// InitSinkClient stores the information needed to be able to connect to the Stepper service.
func InitSinkClient(name string, opts ...grpc.DialOption) error {
	return tsc.initialize(name, opts...)
}

// Reset forcibly resets the trace sink to its initial state.  This is intended
// to support unit tests
func Reset(ctx context.Context) error {
	return tsc.reset(ctx)
}

// Append adds a log entry to the trace sink
func Append(ctx context.Context, entry *log.Entry) error {
	return tsc.append(ctx, entry)
}

// GetPolicy obtains the current trace sink policy and returns it
func GetPolicy(ctx context.Context) (*pb.GetPolicyResponse, error) {
	return tsc.getPolicy(ctx)
}

// GetTraces retrieves up to the specified limit of trace entries, from the
// specified starting point.  It will always wait for at least one
// non-internal entry before returning.
func GetTraces(ctx context.Context, start int64, maxCount int64) <-chan traceData {
	return tsc.getTraces(ctx, start, maxCount)
}
