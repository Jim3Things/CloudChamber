// Unit tests for the stepper service.
package stepper

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes/duration"
	"github.com/stretchr/testify/suite"

	"github.com/Jim3Things/CloudChamber/simulation/internal/common"
	clienttrace "github.com/Jim3Things/CloudChamber/simulation/internal/tracing/client"
	"github.com/Jim3Things/CloudChamber/simulation/internal/tracing/exporters"
	srvtrace "github.com/Jim3Things/CloudChamber/simulation/internal/tracing/server"
	ct "github.com/Jim3Things/CloudChamber/simulation/pkg/protos/common"
	pb "github.com/Jim3Things/CloudChamber/simulation/pkg/protos/services"
	"github.com/Jim3Things/CloudChamber/simulation/test/utilities"

	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

type StepperTestSuite struct {
	suite.Suite

	utf    *exporters.Exporter
	lis    *bufconn.Listener
	client pb.StepperClient
	conn   *grpc.ClientConn
}

func (ts *StepperTestSuite) SetupSuite() {
	require := ts.Require()

	ts.utf = exporters.NewExporter(exporters.NewUTForwarder())
	exporters.ConnectToProvider(ts.utf)

	ts.lis = bufconn.Listen(bufSize)
	s := grpc.NewServer(grpc.UnaryInterceptor(srvtrace.Interceptor))
	err := Register(context.Background(), s, pb.StepperPolicy_Invalid)
	require.NoError(err)

	go func() {
		err2 := s.Serve(ts.lis)
		require.NoError(err2)
	}()
}

func (ts *StepperTestSuite) SetupTest() {
	require := ts.Require()

	_ = ts.utf.Open(ts.T())

	var err error

	ts.conn, err = grpc.Dial(
		"test_channel",
		grpc.WithContextDialer(ts.bufDialer),
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(clienttrace.Interceptor))
	require.NoError(err)

	ts.client = pb.NewStepperClient(ts.conn)

	_, err = ts.client.Reset(context.Background(), &pb.ResetRequest{})
	require.NoError(err)
}

func (ts *StepperTestSuite) TearDownTest() {
	_ = ts.conn.Close()
	ts.utf.Close()
}

func (ts *StepperTestSuite) bufDialer(_ context.Context, _ string) (net.Conn, error) {
	return ts.lis.Dial()
}

func (ts *StepperTestSuite) callNow(ctx context.Context) int64 {
	require := ts.Require()

	resp, err := ts.client.Now(ctx, &pb.NowRequest{})
	require.NoError(err)

	return resp.GetTicks()
}

func (ts *StepperTestSuite) callNowVerify(ctx context.Context, expected int64) {
	assert := ts.Assert()

	current := ts.callNow(ctx)
	assert.Equal(expected, current)
}

func (ts *StepperTestSuite) testNow(ctx context.Context, expected int64) {
	log := ts.T().Log

	ts.callNowVerify(ctx, expected)

	log("Now subtest complete")
}

func (ts *StepperTestSuite) testGetStatus(
	ctx context.Context,
	minTime int64,
	policy pb.StepperPolicy,
	duration *duration.Duration,
	waiters int64) {
	assert := ts.Assert()

	resp, err := ts.client.GetStatus(ctx, &pb.GetStatusRequest{})
	assert.NoError(err)

	assert.GreaterOrEqual(minTime, resp.Now.Ticks)

	assert.Equal(policy, resp.Policy)
	assert.Equal(duration.Seconds, resp.MeasuredDelay.Seconds)
	assert.Equal(duration.Nanos, resp.MeasuredDelay.Nanos)
	assert.Equal(waiters, resp.WaiterCount)
}

func (ts *StepperTestSuite) callStep(ctx context.Context, expected int64) {
	assert := ts.Assert()

	_, err := ts.client.Step(ctx, &pb.StepRequest{})
	assert.NoError(err)
	ts.callNowVerify(ctx, expected)
}

func (ts *StepperTestSuite) testStep(ctx context.Context, expected int64) {
	log := ts.T().Log

	ts.callStep(ctx, expected)

	log("Step subtest complete")
}

func (ts *StepperTestSuite) testDelay(ctx context.Context, atLeast int64, jitter int64) {
	assert := ts.Assert()
	log := ts.T().Log

	start := ts.callNow(ctx)

	resp, err := ts.client.Delay(ctx, &pb.DelayRequest{AtLeast: &ct.Timestamp{Ticks: atLeast}, Jitter: jitter})
	assert.NoError(err)

	minLegal := atLeast
	maxLegal := start + atLeast + jitter

	assert.True(
		resp.Ticks >= minLegal && resp.Ticks <= maxLegal,
		"Delay out of range, should be %d - %d, is %d",
		minLegal,
		maxLegal,
		resp.Ticks)

	log("Delay subtest complete")
}

// Verify that it is not legal to set the stepper to the "Invalid" policy.
func (ts *StepperTestSuite) TestInvalidSetPolicyType() {
	assert := ts.Assert()

	ctx := context.Background()

	_, err := ts.client.SetPolicy(
		ctx,
		&pb.PolicyRequest{
			Policy:        pb.StepperPolicy_Invalid,
			MeasuredDelay: &duration.Duration{Seconds: 0},
			MatchEpoch:    -1})

	assert.Error(err)
}

// Verify that a 'Manual' policy must specify a measured delay that
// is zero (this delay value is only valid for the 'Measured" policy).
func (ts *StepperTestSuite) TestInvalidSetPolicyManual() {
	assert := ts.Assert()

	ctx := context.Background()

	_, err := ts.client.SetPolicy(
		ctx,
		&pb.PolicyRequest{
			Policy:        pb.StepperPolicy_Manual,
			MeasuredDelay: &duration.Duration{Seconds: 1},
			MatchEpoch:    -1})
	assert.Error(err)

	_, err = ts.client.SetPolicy(
		ctx,
		&pb.PolicyRequest{
			Policy:        pb.StepperPolicy_Manual,
			MeasuredDelay: &duration.Duration{Seconds: 0},
			MatchEpoch:    -1})
	assert.NoError(err)

	_, err = ts.client.SetPolicy(
		ctx,
		&pb.PolicyRequest{
			Policy:        pb.StepperPolicy_Manual,
			MeasuredDelay: &duration.Duration{Seconds: 2},
			MatchEpoch:    -1})
	assert.Error(err)
}

// Verify that a 'Measured' policy must have a non-zero and positive
// measured delay value.
func (ts *StepperTestSuite) TestInvalidSetPolicyMeasured() {
	assert := ts.Assert()

	ctx := context.Background()

	_, err := ts.client.SetPolicy(
		ctx,
		&pb.PolicyRequest{
			Policy:        pb.StepperPolicy_Measured,
			MeasuredDelay: &duration.Duration{Seconds: -1},
			MatchEpoch:    -1})
	assert.Error(err)

	_, err = ts.client.SetPolicy(
		ctx,
		&pb.PolicyRequest{
			Policy:        pb.StepperPolicy_Measured,
			MeasuredDelay: &duration.Duration{Seconds: 0},
			MatchEpoch:    -1})
	assert.Error(err)

	_, err = ts.client.SetPolicy(
		ctx,
		&pb.PolicyRequest{
			Policy:        pb.StepperPolicy_Measured,
			MeasuredDelay: &duration.Duration{Seconds: 1},
			MatchEpoch:    -1})
	assert.NoError(err)

	_, err = ts.client.SetPolicy(
		ctx,
		&pb.PolicyRequest{
			Policy:        pb.StepperPolicy_Measured,
			MeasuredDelay: &duration.Duration{Seconds: 2},
			MatchEpoch:    -1})
	assert.NoError(err)
}

// Verify that a 'NoWait' policy has a zero measured delay value.
func (ts *StepperTestSuite) TestInvalidSetPolicyNoWait() {
	assert := ts.Assert()

	ctx := context.Background()

	_, err := ts.client.SetPolicy(
		ctx,
		&pb.PolicyRequest{
			Policy:        pb.StepperPolicy_NoWait,
			MeasuredDelay: &duration.Duration{Seconds: 1},
			MatchEpoch:    -1})
	assert.Error(err)

	_, err = ts.client.SetPolicy(
		ctx,
		&pb.PolicyRequest{
			Policy:        pb.StepperPolicy_NoWait,
			MeasuredDelay: &duration.Duration{Seconds: 0},
			MatchEpoch:    -1})
	assert.NoError(err)

	_, err = ts.client.SetPolicy(
		ctx,
		&pb.PolicyRequest{
			Policy:        pb.StepperPolicy_NoWait,
			MeasuredDelay: &duration.Duration{Seconds: 2},
			MatchEpoch:    -1})
	assert.Error(err)
}

// Verify that a delay request has a non-zero and positive delay, and a
// non-negative jitter value.
func (ts *StepperTestSuite) TestInvalidDelay() {
	assert := ts.Assert()

	ctx := context.Background()

	_, err := ts.client.SetPolicy(
		ctx,
		&pb.PolicyRequest{
			Policy:        pb.StepperPolicy_NoWait,
			MeasuredDelay: &duration.Duration{Seconds: 0},
			MatchEpoch:    -1})
	assert.NoError(err)

	_, err = ts.client.Delay(
		ctx,
		&pb.DelayRequest{
			AtLeast: &ct.Timestamp{Ticks: -1},
			Jitter:  0})
	assert.Error(err)

	_, err = ts.client.Delay(
		ctx,
		&pb.DelayRequest{
			AtLeast: &ct.Timestamp{Ticks: 1},
			Jitter:  -1})
	assert.Error(err)
}

// Verify the basic operations while under the 'NoWait' policy.
func (ts *StepperTestSuite) TestStepper_NoWait() {
	assert := ts.Assert()

	ctx := context.Background()

	// These need to execute in a particular order, so we're calling them as
	// included subtests in this unit test

	_, err := ts.client.SetPolicy(
		ctx,
		&pb.PolicyRequest{
			Policy:        pb.StepperPolicy_NoWait,
			MeasuredDelay: &duration.Duration{Seconds: 0},
			MatchEpoch:    -1})
	assert.NoError(err)

	ts.testNow(ctx, 0)
	ts.testGetStatus(ctx, 0, pb.StepperPolicy_NoWait, &duration.Duration{Seconds: 0}, 0)
	ts.testDelay(ctx, 1, 2)
}

// Verify the basic operations while under the 'Measured' policy.
func (ts *StepperTestSuite) TestStepper_Measured() {
	assert := ts.Assert()
	log := ts.T().Log

	ctx := context.Background()

	// These need to execute in a particular order, so we're calling them as
	// included subtests in this unit test

	_, err := ts.client.SetPolicy(
		ctx,
		&pb.PolicyRequest{
			Policy:        pb.StepperPolicy_Measured,
			MeasuredDelay: &duration.Duration{Seconds: 1},
			MatchEpoch:    -1})
	assert.NoError(err)

	ts.testNow(ctx, 0)
	ts.testGetStatus(ctx, 0, pb.StepperPolicy_Measured, &duration.Duration{Seconds: 1}, 0)

	// Verify that simulated time moves forward as a result of wall clock time.
	utilities.WaitForStateChange(2, func() bool {
		current := ts.callNow(ctx)
		return current > 1
	})

	log("Now subtest complete")
	ts.testDelay(ctx, 3, 2)
}

// Verify the basic operations while under the 'Manual' policy.
func (ts *StepperTestSuite) TestStepper_Manual() {
	assert := ts.Assert()

	ctx := context.Background()

	// These need to execute in a particular order, so we're calling them as
	// included subtests in this unit test

	_, err := ts.client.SetPolicy(
		ctx,
		&pb.PolicyRequest{
			Policy:        pb.StepperPolicy_Manual,
			MeasuredDelay: &duration.Duration{Seconds: 0},
			MatchEpoch:    -1})
	assert.NoError(err)

	ts.testNow(ctx, 0)
	ts.testGetStatus(ctx, 0, pb.StepperPolicy_Manual, &duration.Duration{Seconds: 0}, 0)

	ts.testStep(ctx, 1)
	ts.testGetStatus(ctx, 1, pb.StepperPolicy_Manual, &duration.Duration{Seconds: 0}, 0)

	ch := make(chan bool)

	go func(res chan<- bool) {
		rsp, err2 := ts.client.Delay(ctx, &pb.DelayRequest{AtLeast: &ct.Timestamp{Ticks: 3}, Jitter: 0})
		assert.NoError(err2)
		assert.Equal(rsp.Ticks, int64(3))

		res <- true
	}(ch)

	assert.True(common.DoNotCompleteWithin(ch, time.Duration(1)*time.Second))
	ts.testGetStatus(ctx, 1, pb.StepperPolicy_Manual, &duration.Duration{Seconds: 0}, 1)

	ts.callStep(ctx, 2)
	assert.True(common.DoNotCompleteWithin(ch, time.Duration(1)*time.Second))

	ts.callStep(ctx, 3)
	assert.True(common.CompleteWithin(ch, time.Duration(1)*time.Second))

	ts.testGetStatus(ctx, 3, pb.StepperPolicy_Manual, &duration.Duration{Seconds: 0}, 0)
}

func (ts *StepperTestSuite) TestStepperResetWithActiveDelays() {
	assert := ts.Assert()

	ctx := context.Background()

	_, err := ts.client.SetPolicy(
		ctx,
		&pb.PolicyRequest{
			Policy:        pb.StepperPolicy_Manual,
			MeasuredDelay: &duration.Duration{Seconds: 0},
			MatchEpoch:    -1})
	assert.NoError(err)

	ts.testGetStatus(ctx, 0, pb.StepperPolicy_Manual, &duration.Duration{Seconds: 0}, 0)

	ch := make(chan bool)

	go func(res chan<- bool) {
		rsp, err2 := ts.client.Delay(ctx, &pb.DelayRequest{AtLeast: &ct.Timestamp{Ticks: 3}, Jitter: 0})
		assert.Error(err2, "Delay called succeeded")
		assert.Nil(rsp)

		res <- true
	}(ch)

	utilities.WaitForStateChange(1, func() bool {
		require := ts.Require()

		resp, err2 := ts.client.GetStatus(ctx, &pb.GetStatusRequest{})
		require.NoError(err2)

		return resp.WaiterCount == 1
	})

	ts.testGetStatus(ctx, 0, pb.StepperPolicy_Manual, &duration.Duration{Seconds: 0}, 1)

	_, err = ts.client.Reset(ctx, &pb.ResetRequest{})
	assert.NoError(err)

	assert.True(common.CompleteWithin(ch, time.Duration(1)*time.Second))
	ts.testGetStatus(ctx, 0, pb.StepperPolicy_Invalid, &duration.Duration{Seconds: 0}, 0)
}

// Verify the basic operations while under the 'Invalid' policy fai.
func (ts *StepperTestSuite) TestInvalidOperations() {
	assert := ts.Assert()

	ctx := context.Background()

	nowRsp, err := ts.client.Now(ctx, &pb.NowRequest{})
	assert.Error(err)
	assert.Nil(nowRsp)

	stepRsp, err := ts.client.Step(ctx, &pb.StepRequest{})
	assert.Error(err)
	assert.Nil(stepRsp)

	delayRsp, err := ts.client.Delay(ctx, &pb.DelayRequest{
		AtLeast: &ct.Timestamp{Ticks: 1},
		Jitter:  0,
	})

	assert.Error(err)
	assert.Nil(delayRsp)
}

func TestStepperTestSuite(T *testing.T) {
	suite.Run(T, new(StepperTestSuite))
}
