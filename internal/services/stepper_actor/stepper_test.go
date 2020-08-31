// Unit tests for the stepper service.
package stepper

import (
	"context"
	"log"
	"net"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes/duration"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"

	"github.com/Jim3Things/CloudChamber/internal/common/channels"
	clienttrace "github.com/Jim3Things/CloudChamber/internal/tracing/client"
	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters"
	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters/unit_test"
	srvtrace "github.com/Jim3Things/CloudChamber/internal/tracing/server"
	"github.com/Jim3Things/CloudChamber/internal/tracing/setup"
	ct "github.com/Jim3Things/CloudChamber/pkg/protos/common"

	pb "github.com/Jim3Things/CloudChamber/pkg/protos/services"

	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener
var client pb.StepperClient

func init() {
	setup.Init(exporters.UnitTest)

	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer(grpc.UnaryInterceptor(srvtrace.Interceptor))
	if err := Register(s, pb.StepperPolicy_Invalid); err != nil {
		log.Fatalf("Failed to register wither error: %v", err)
	}

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()
}

func bufDialer(_ context.Context, _ string) (net.Conn, error) {
	return lis.Dial()
}

func callNow(t *testing.T, ctx context.Context) int64 {
	resp, err := client.Now(ctx, &pb.NowRequest{})
	assert.Nilf(t, err, "Now failed: %v", err)

	return resp.GetTicks()
}

func callNowVerify(t *testing.T, ctx context.Context, expected int64) {
	current := callNow(t, ctx)
	assert.Equalf(t, expected, current, "Now returned an invalid value: %d, should be %d", current, expected)
}

func callNowAtLeast(t *testing.T, ctx context.Context, atLeast int64) {
	current := callNow(t, ctx)
	assert.Truef(t, current >= atLeast, "Now returned an invalid value: %d, should be at least %d", current, atLeast)
	t.Logf("Now returned %d", current)
}

func testNow(t *testing.T, ctx context.Context, expected int64) {
	callNowVerify(t, ctx, expected)

	t.Log("Now subtest complete")
}

func testGetStatus(
	t *testing.T,
	ctx context.Context,
	minTime int64,
	policy pb.StepperPolicy,
	duration *duration.Duration,
	waiters int64) {
	resp, err := client.GetStatus(ctx, &pb.GetStatusRequest{})
	assert.Nilf(t, err, "GetStatus failed: %v", err)

	assert.GreaterOrEqual(
		t,
		minTime, resp.Now.Ticks,
		"Time returned is less than expected, %d should be no more than %d", minTime, resp.Now.Ticks)

	assert.Equal(t, policy, resp.Policy, "Expected policy %v, got policy %v", policy, resp.Policy)
	assert.Equal(t, duration.Seconds, resp.MeasuredDelay.Seconds, "Expected delay to be %d, was %d", duration.Seconds, resp.MeasuredDelay.Seconds)
	assert.Equal(t, duration.Nanos, resp.MeasuredDelay.Nanos, "Expected delay to be %d, was %d", duration.Seconds, resp.MeasuredDelay.Seconds)
	assert.Equal(t, waiters, resp.WaiterCount, "Expected waiter count to be %d, was %d", waiters, resp.WaiterCount)
}

func callStep(t *testing.T, ctx context.Context, expected int64) {
	_, err := client.Step(ctx, &pb.StepRequest{})
	assert.Nilf(t, err, "Step failed: %v", err)
	callNowVerify(t, ctx, expected)
}

func testStep(t *testing.T, ctx context.Context, expected int64) {
	callStep(t, ctx, expected)
	t.Log("Step subtest complete")
}

func testDelay(t *testing.T, ctx context.Context, atLeast int64, jitter int64) {
	start := callNow(t, ctx)

	resp, err := client.Delay(ctx, &pb.DelayRequest{AtLeast: &ct.Timestamp{Ticks: atLeast}, Jitter: jitter})
	assert.Nilf(t, err, "Delay failed: %v, err")

	minLegal := atLeast
	maxLegal := start + atLeast + jitter

	assert.True(
		t,
		resp.Ticks >= minLegal && resp.Ticks <= maxLegal,
		"Delay out of range, should be %d - %d, is %d",
		minLegal,
		maxLegal,
		resp.Ticks)

	t.Log("Delay subtest complete")
}

func commonSetup(t *testing.T) (context.Context, *grpc.ClientConn) {
	conn, err := grpc.Dial(
		"bufnet",
		grpc.WithContextDialer(bufDialer),
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(clienttrace.Interceptor))
	assert.Nilf(t, err, "Failed to dial bufnet: %v", err)

	md := metadata.Pairs(
		"timestamp", time.Now().Format(time.StampNano),
		"client-id", "web-api-client-us-east-1",
		"user-id", "some-test-user-id",
	)
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	client = pb.NewStepperClient(conn)
	_, err = client.Reset(ctx, &pb.ResetRequest{})
	assert.Nilf(t, err, "Failed to reset: %v", err)

	return ctx, conn
}

func TestInvalidSetPolicyType(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	ctx, conn := commonSetup(t)
	defer func() { _ = conn.Close() }()

	client = pb.NewStepperClient(conn)

	_, err := client.SetPolicy(ctx, &pb.PolicyRequest{Policy: pb.StepperPolicy_Invalid, MeasuredDelay: &duration.Duration{Seconds: 0}, MatchEpoch: -1})
	assert.NotNilf(t, err, "SetPolicy unexpectedly succeeded with an invalid policy type")
}

func TestInvalidSetPolicyManual(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	ctx, conn := commonSetup(t)
	defer func() { _ = conn.Close() }()

	client = pb.NewStepperClient(conn)

	_, err := client.SetPolicy(ctx, &pb.PolicyRequest{Policy: pb.StepperPolicy_Manual, MeasuredDelay: &duration.Duration{Seconds: 1}, MatchEpoch: -1})
	assert.NotNilf(t, err, "SetPolicy unexpectedly succeeded with an invalid duration: %v", err)

	_, err = client.SetPolicy(ctx, &pb.PolicyRequest{Policy: pb.StepperPolicy_Manual, MeasuredDelay: &duration.Duration{Seconds: 0}, MatchEpoch: -1})
	assert.Nilf(t, err, "SetPolicy unexpectedly failed: %v", err)

	_, err = client.SetPolicy(ctx, &pb.PolicyRequest{Policy: pb.StepperPolicy_Manual, MeasuredDelay: &duration.Duration{Seconds: 2}, MatchEpoch: -1})
	assert.NotNilf(t, err, "SetPolicy unexpectedly succeeded with an invalid duration: %v", err)
}

func TestInvalidSetPolicyMeasured(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	ctx, conn := commonSetup(t)
	defer func() { _ = conn.Close() }()

	client = pb.NewStepperClient(conn)

	_, err := client.SetPolicy(ctx, &pb.PolicyRequest{Policy: pb.StepperPolicy_Measured, MeasuredDelay: &duration.Duration{Seconds: 0}, MatchEpoch: -1})
	assert.NotNilf(t, err, "SetPolicy unexpectedly succeeded with an invalid duration: %v", err)

	_, err = client.SetPolicy(ctx, &pb.PolicyRequest{Policy: pb.StepperPolicy_Measured, MeasuredDelay: &duration.Duration{Seconds: 1}, MatchEpoch: -1})
	assert.Nilf(t, err, "SetPolicy unexpectedly failed: %v", err)

	_, err = client.SetPolicy(ctx, &pb.PolicyRequest{Policy: pb.StepperPolicy_Measured, MeasuredDelay: &duration.Duration{Seconds: 2}, MatchEpoch: -1})
	assert.Nilf(t, err, "SetPolicy unexpectedly failed: %v", err)

	time.Sleep(time.Duration(2) * time.Second)
}

func TestInvalidSetPolicyNoWait(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	ctx, conn := commonSetup(t)
	defer func() { _ = conn.Close() }()

	client = pb.NewStepperClient(conn)

	_, err := client.SetPolicy(ctx, &pb.PolicyRequest{Policy: pb.StepperPolicy_NoWait, MeasuredDelay: &duration.Duration{Seconds: 1}, MatchEpoch: -1})
	assert.NotNilf(t, err, "SetPolicy unexpectedly succeeded with an invalid duration: %v", err)

	_, err = client.SetPolicy(ctx, &pb.PolicyRequest{Policy: pb.StepperPolicy_NoWait, MeasuredDelay: &duration.Duration{Seconds: 0}, MatchEpoch: -1})
	assert.Nilf(t, err, "SetPolicy unexpectedly failed: %v", err)

	_, err = client.SetPolicy(ctx, &pb.PolicyRequest{Policy: pb.StepperPolicy_NoWait, MeasuredDelay: &duration.Duration{Seconds: 2}, MatchEpoch: -1})
	assert.NotNilf(t, err, "SetPolicy unexpectedly succeeded with an invalid duration: %v", err)
}

func TestInvalidDelay(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	ctx, conn := commonSetup(t)
	defer func() { _ = conn.Close() }()

	client = pb.NewStepperClient(conn)

	_, err := client.SetPolicy(ctx, &pb.PolicyRequest{Policy: pb.StepperPolicy_NoWait, MeasuredDelay: &duration.Duration{Seconds: 0}, MatchEpoch: -1})
	assert.Nilf(t, err, "SetPolicy unexpectedly failed: %v", err)

	_, err = client.Delay(ctx, &pb.DelayRequest{AtLeast: &ct.Timestamp{Ticks: -1}, Jitter: 0})
	assert.NotNil(t, err, "Delay unexpectedly succeeded with an invalid base delay time")

	_, err = client.Delay(ctx, &pb.DelayRequest{AtLeast: &ct.Timestamp{Ticks: 1}, Jitter: -1})
	assert.NotNil(t, err, "Delay unexpectedly succeeded with an invalid jitter")
}

func TestStepper_NoWait(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	ctx, conn := commonSetup(t)
	defer func() { _ = conn.Close() }()

	client = pb.NewStepperClient(conn)

	// These need to execute in a particular order, so we're calling them as
	// included subtests in this unit test

	_, err := client.SetPolicy(ctx, &pb.PolicyRequest{Policy: pb.StepperPolicy_NoWait, MeasuredDelay: &duration.Duration{Seconds: 0}, MatchEpoch: -1})
	assert.Nilf(t, err, "SetPolicy failed: %v", err)

	testNow(t, ctx, 0)
	testGetStatus(t, ctx, 0, pb.StepperPolicy_NoWait, &duration.Duration{Seconds: 0}, 0)
	testDelay(t, ctx, 1, 2)
}

func TestStepper_Measured(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	ctx, conn := commonSetup(t)
	defer func() { _ = conn.Close() }()

	client = pb.NewStepperClient(conn)

	// These need to execute in a particular order, so we're calling them as
	// included subtests in this unit test

	_, err := client.SetPolicy(ctx, &pb.PolicyRequest{Policy: pb.StepperPolicy_Measured, MeasuredDelay: &duration.Duration{Seconds: 1}, MatchEpoch: -1})
	assert.Nilf(t, err, "SetPolicy failed: %v", err)

	testNow(t, ctx, 0)
	testGetStatus(t, ctx, 0, pb.StepperPolicy_Measured, &duration.Duration{Seconds: 1}, 0)
	time.Sleep(time.Duration(2) * time.Second)

	callNowAtLeast(t, ctx, 1)

	t.Log("Now subtest complete")
	testDelay(t, ctx, 3, 2)
}

func TestStepper_Manual(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	ctx, conn := commonSetup(t)
	defer func() { _ = conn.Close() }()

	client = pb.NewStepperClient(conn)

	// These need to execute in a particular order, so we're calling them as
	// included subtests in this unit test

	_, err := client.SetPolicy(ctx, &pb.PolicyRequest{Policy: pb.StepperPolicy_Manual, MeasuredDelay: &duration.Duration{Seconds: 0}, MatchEpoch: -1})
	assert.Nilf(t, err, "SetPolicy failed: %v", err)

	testNow(t, ctx, 0)
	testGetStatus(t, ctx, 0, pb.StepperPolicy_Manual, &duration.Duration{Seconds: 0}, 0)

	testStep(t, ctx, 1)
	testGetStatus(t, ctx, 1, pb.StepperPolicy_Manual, &duration.Duration{Seconds: 0}, 0)

	ch := make(chan bool)

	go func(res chan<- bool) {
		rsp, err := client.Delay(ctx, &pb.DelayRequest{AtLeast: &ct.Timestamp{Ticks: 3}, Jitter: 0})
		assert.Nilf(t, err, "Delay called failed, returned %v", err)
		assert.Equal(t, rsp.Ticks, int64(3), "Delay returned an invalid time.  Should be 3, but was %d", rsp.Ticks)

		res <- true
	}(ch)

	assert.True(t, channels.DoNotCompleteWithin(ch, time.Duration(1) * time.Second))
	testGetStatus(t, ctx, 1, pb.StepperPolicy_Manual, &duration.Duration{Seconds: 0}, 1)

	callStep(t, ctx, 2)
	assert.True(t, channels.DoNotCompleteWithin(ch, time.Duration(1) * time.Second))

	callStep(t, ctx, 3)
	assert.True(t, channels.CompleteWithin(ch, time.Duration(1) * time.Second))

	testGetStatus(t, ctx, 3, pb.StepperPolicy_Manual, &duration.Duration{Seconds: 0}, 0)

	t.Log("DelayManual subtest complete")
}
