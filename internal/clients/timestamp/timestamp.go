// This module provides the basic client methods for using the simulated time
// service.

package timestamp

import (
	"context"
	"errors"
	"sync"

	"github.com/golang/protobuf/ptypes/duration"

	"github.com/Jim3Things/CloudChamber/internal/common"
	ct "github.com/Jim3Things/CloudChamber/pkg/protos/common"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/services"

	"google.golang.org/grpc"
)

var (
	errAlreadyInitialized = errors.New("timestamp client has already been initialized")
	errNotReady           = errors.New("client is not ready")

	tsc timeClient = &notReady{}
)

type timeClient interface {
	initialize(name string, opts ...grpc.DialOption) error
	setPolicy(ctx context.Context, policy pb.StepperPolicy, delay *duration.Duration, match int64) error
	advance(ctx context.Context) error
	now(ctx context.Context) (*ct.Timestamp, error)
	after(ctx context.Context, deadline *ct.Timestamp) <-chan TimeData
	status(ctx context.Context) (*pb.StatusResponse, error)
	reset(ctx context.Context) error
}

// TimeData defines the value returned from a delay wait.  This is more than
// the simple timestamp inasmuch as the delay call can fail asynchronously.
type TimeData struct {
	Time *ct.Timestamp
	Err  error
}

// notReady is the uninitialized timestamp client, which only returns an error
type notReady struct{}

func (n notReady) initialize(name string, opts ...grpc.DialOption) error {
	tsc = newTimeClient(name, opts...)
	return nil
}

// SetPolicy sets the stepper policy
func (n *notReady) setPolicy(_ context.Context, _ pb.StepperPolicy, _ *duration.Duration, _ int64) error {
	return errNotReady
}

func (n *notReady) advance(_ context.Context) error {
	return errNotReady
}

func (n *notReady) now(_ context.Context) (*ct.Timestamp, error) {
	return nil, errNotReady
}

func (n *notReady) after(_ context.Context, _ *ct.Timestamp) <-chan TimeData {
	ch := make(chan TimeData)
	ch <- TimeData{
		Time: nil,
		Err:  errNotReady,
	}

	return ch
}

func (n *notReady) status(_ context.Context) (*pb.StatusResponse, error) {
	return nil, errNotReady
}

func (n *notReady) reset(_ context.Context) error {
	return errNotReady
}

// activeClient is the timestamp client that has been configured to connect to
// the stepper service.
type activeClient struct {
	dialName string
	dialOpts []grpc.DialOption

	conn   *grpc.ClientConn
	client pb.StepperClient
	m      *sync.Mutex
}

func newTimeClient(dialName string, opts ...grpc.DialOption) timeClient {
	return &activeClient{
		dialName: dialName,
		dialOpts: append([]grpc.DialOption{}, opts...),
		conn:     nil,
		client:   nil,
		m:        &sync.Mutex{},
	}
}

func (t *activeClient) setPolicy(
	ctx context.Context,
	policy pb.StepperPolicy,
	delay *duration.Duration,
	match int64) error {
	client, err := t.dial()
	if err != nil {
		return err
	}

	_, err = client.SetPolicy(
		ctx,
		&pb.PolicyRequest{
			Policy:        policy,
			MeasuredDelay: delay,
			MatchEpoch:    match,
		})

	return err
}

func (t *activeClient) initialize(_ string, _ ...grpc.DialOption) error {
	return errAlreadyInitialized
}

func (t *activeClient) advance(ctx context.Context) error {
	client, err := t.dial()
	if err != nil {
		return err
	}

	_, err = client.Step(ctx, &pb.StepRequest{})

	return t.cleanup(client, err)
}

func (t *activeClient) now(ctx context.Context) (*ct.Timestamp, error) {
	client, err := t.dial()
	if err != nil {
		return nil, err
	}

	stamp, err := client.Now(ctx, &pb.NowRequest{})
	return stamp, t.cleanup(client, err)
}

func (t *activeClient) after(ctx context.Context, deadline *ct.Timestamp) <-chan TimeData {
	ch := make(chan TimeData)

	go func(ctx context.Context, res chan<- TimeData) {
		client, err := t.dial()
		if err != nil {
			res <- TimeData{
				Time: nil,
				Err:  err,
			}
			return
		}

		rsp, err := client.Delay(ctx, &pb.DelayRequest{AtLeast: deadline, Jitter: 0})

		if err != nil {
			res <- TimeData{Time: nil, Err: t.cleanup(client, err)}
			return
		}
		res <- TimeData{Time: rsp, Err: nil}
	}(ctx, ch)

	return ch
}

func (t *activeClient) status(ctx context.Context) (*pb.StatusResponse, error) {
	client, err := t.dial()
	if err != nil {
		return nil, err
	}

	rsp, err := client.GetStatus(ctx, &pb.GetStatusRequest{})
	return rsp, t.cleanup(client, err)
}

func (t *activeClient) reset(ctx context.Context) error {
	client, err := t.dial()
	if err != nil {
		return err
	}

	_, err = client.Reset(ctx, &pb.ResetRequest{})
	return t.cleanup(client, err)
}

// dial abstracts the connection logic to the trace sink service.  It caches
// the connection for use in later calls in order to avoid excess transport and
// grpc connection operations.
func (t *activeClient) dial() (pb.StepperClient, error) {
	t.m.Lock()
	defer t.m.Unlock()

	if t.conn == nil {
		conn, err := grpc.Dial(t.dialName, t.dialOpts...)
		if err != nil {
			return nil, err
		}

		t.conn = conn
		t.client = pb.NewStepperClient(conn)
	}

	return t.client, nil
}

// cleanup ensures that if an error has occurred the cached connection is
// cleared.
func (t *activeClient) cleanup(client pb.StepperClient, err error) error {
	t.m.Lock()
	defer t.m.Unlock()

	// Clear the connection iff we had an error, the connection has not been
	// cleaned up already, and we had an error against the current client.
	if err != nil && t.conn != nil && client == t.client {
		_ = t.conn.Close()

		t.conn = nil
		t.client = nil
	}

	return err
}

// InitTimestamp stores the information needed to be able to connect to the
// Stepper service.
func InitTimestamp(name string, opts ...grpc.DialOption) error {
	return tsc.initialize(name, opts...)
}

// SetPolicy sets the stepper policy
func SetPolicy(ctx context.Context, policy pb.StepperPolicy, delay *duration.Duration, match int64) error {
	return tsc.setPolicy(ctx, policy, delay, match)
}

// Advance the simulated time, assuming that the policy mode is manual
func Advance(ctx context.Context) error {
	return tsc.advance(ctx)
}

// Now gets the current simulated time.
func Now(ctx context.Context) (*ct.Timestamp, error) {
	return tsc.now(ctx)
}

// After delays execution until the simulated time meets or exceeds the
// specified deadline.  Completion is asynchronous, even if no delay is
// required.
func After(ctx context.Context, deadline *ct.Timestamp) <-chan TimeData {
	return tsc.after(ctx, deadline)
}

// Status retrieves the status of the Stepper service
func Status(ctx context.Context) (*pb.StatusResponse, error) {
	return tsc.status(ctx)
}

// Reset the simulated time back to its starting state, including reverting all
// policies back to their default.  This is used by unit tests to ensure a well
// known starting state for a test.
func Reset(ctx context.Context) error {
	return tsc.reset(ctx)
}

// Tick provides the current simulated time Tick, or '-1' if the simulated time
// cannot be retrieved (e.g. during startup)
func Tick(ctx context.Context) int64 {
	now, err := tsc.now(ctx)
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
