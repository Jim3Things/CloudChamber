// Unit tests for the stepper service.
package stepper

import (
	"context"
	"log"
	"net"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"

	clienttrace "github.com/Jim3Things/CloudChamber/internal/tracing/client"
	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters"
	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters/unit_test"
	srvtrace "github.com/Jim3Things/CloudChamber/internal/tracing/server"
	"github.com/Jim3Things/CloudChamber/internal/tracing/setup"

	pb "github.com/Jim3Things/CloudChamber/pkg/protos/Stepper"

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
	Register(s)

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()
}

func bufDialer(_ context.Context, _ string) (net.Conn, error) {
	return lis.Dial()
}

func checkForEarlyCompletion(t *testing.T, ch <-chan bool, delay int, name string) {
	select {
	case <-ch:
		assert.Failf(t, "%s completed early", name)
	case <-time.After(time.Duration(delay) * time.Second):
	}
}

func checkForLateCompletion(t *testing.T, ch <-chan bool, delay int, name string) {
	select {
	case <-ch:
	case <-time.After(time.Duration(delay) * time.Second):
		assert.Failf(t, "%s did not complete on time", name)
	}
}

func testSetPolicy(t *testing.T, ctx context.Context, policy pb.StepperPolicy, badPolicy pb.StepperPolicy, seconds int64) {
	_, err := client.SetPolicy(ctx, &pb.PolicyRequest{Policy: policy, MeasuredDelay: &duration.Duration{Seconds: seconds}})
	assert.Nilf(t, err, "SetPolicy failed: %v", err)

	_, err = client.SetPolicy(ctx, &pb.PolicyRequest{Policy: policy, MeasuredDelay: &duration.Duration{Seconds: seconds}})
	assert.Nilf(t, err, "SetPolicy failed: %v", err)

	_, err = client.SetPolicy(ctx, &pb.PolicyRequest{Policy: policy, MeasuredDelay: &duration.Duration{Seconds: seconds + 1}})
	assert.NotNil(t, err, "SetPolicy succeeded, should have failed")

	_, err = client.SetPolicy(ctx, &pb.PolicyRequest{Policy: badPolicy, MeasuredDelay: &duration.Duration{Seconds: seconds}})
	assert.NotNil(t, err, "SetPolicy succeeded, should have failed")

	t.Log("SetPolicy subtest complete")
}

func callNow(t *testing.T, ctx context.Context) int64 {
	resp, err := client.Now(ctx, &empty.Empty{})
	assert.Nilf(t, err, "Now failed: %v", err)

	return resp.GetTicks()
}

func callNowVerify(t *testing.T, ctx context.Context, expected int64) {
	current := callNow(t, ctx)
	assert.Equalf(t, current, expected, "Now returned an invalid value: %d, should be %d", current, expected)
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

func callStep(t *testing.T, ctx context.Context, expected int64) {
	_, err := client.Step(ctx, &empty.Empty{})
	assert.Nilf(t, err, "Step failed: %v", err)
	callNowVerify(t, ctx, expected)
}

func testStep(t *testing.T, ctx context.Context, expected int64) {
	callStep(t, ctx, expected)
	t.Log("Step subtest complete")
}

func testDelay(t *testing.T, ctx context.Context, atLeast int64, jitter int64) {
	start := callNow(t, ctx)

	resp, err := client.Delay(ctx, &pb.DelayRequest{AtLeast: atLeast, Jitter: jitter})
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
	Reset()
	unit_test.SetTesting(t)

	conn, err := grpc.Dial("bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure(), grpc.WithUnaryInterceptor(clienttrace.Interceptor))
	assert.Nilf(t, err, "Failed to dial bufnet: %v", err)

	md := metadata.Pairs(
		"timestamp", time.Now().Format(time.StampNano),
		"client-id", "web-api-client-us-east-1",
		"user-id", "some-test-user-id",
	)
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	return ctx, conn
}

func TestInvalidSetPolicyType(t *testing.T) {
	ctx, conn := commonSetup(t)
	defer func() { _ = conn.Close() }()

	client = pb.NewStepperClient(conn)

	_, err := client.SetPolicy(ctx, &pb.PolicyRequest{Policy: pb.StepperPolicy_Invalid, MeasuredDelay: &duration.Duration{Seconds: 0}})
	assert.NotNilf(t, err, "SetPolicy unexpectedly succeeded with an invalid policy type")
}

func TestInvalidSetPolicyManual(t *testing.T) {
	ctx, conn := commonSetup(t)
	defer func() { _ = conn.Close() }()

	client = pb.NewStepperClient(conn)

	_, err := client.SetPolicy(ctx, &pb.PolicyRequest{Policy: pb.StepperPolicy_Manual, MeasuredDelay: &duration.Duration{Seconds: 1}})
	assert.NotNilf(t, err, "SetPolicy unexpectedly succeeded with an invalid duration: %v", err)

	_, err = client.SetPolicy(ctx, &pb.PolicyRequest{Policy: pb.StepperPolicy_Manual, MeasuredDelay: &duration.Duration{Seconds: 0}})
	assert.Nilf(t, err, "SetPolicy unexpectedly failed: %v", err)

	_, err = client.SetPolicy(ctx, &pb.PolicyRequest{Policy: pb.StepperPolicy_Manual, MeasuredDelay: &duration.Duration{Seconds: 2}})
	assert.NotNilf(t, err, "SetPolicy unexpectedly succeeded with an invalid duration: %v", err)
}

func TestInvalidSetPolicyMeasured(t *testing.T) {
	ctx, conn := commonSetup(t)
	defer func() { _ = conn.Close() }()

	client = pb.NewStepperClient(conn)

	_, err := client.SetPolicy(ctx, &pb.PolicyRequest{Policy: pb.StepperPolicy_Measured, MeasuredDelay: &duration.Duration{Seconds: 0}})
	assert.NotNilf(t, err, "SetPolicy unexpectedly succeeded with an invalid duration: %v", err)

	_, err = client.SetPolicy(ctx, &pb.PolicyRequest{Policy: pb.StepperPolicy_Measured, MeasuredDelay: &duration.Duration{Seconds: 1}})
	assert.Nilf(t, err, "SetPolicy unexpectedly failed: %v", err)

	_, err = client.SetPolicy(ctx, &pb.PolicyRequest{Policy: pb.StepperPolicy_Measured, MeasuredDelay: &duration.Duration{Seconds: 2}})
	assert.NotNilf(t, err, "SetPolicy unexpectedly succeeded with an invalid duration: %v", err)

	// Now force a reset to clear out the free-running autostepper
	Reset()
	time.Sleep(time.Duration(2) * time.Second)
}

func TestInvalidSetPolicyNoWait(t *testing.T) {
	ctx, conn := commonSetup(t)
	defer func() { _ = conn.Close() }()

	client = pb.NewStepperClient(conn)

	_, err := client.SetPolicy(ctx, &pb.PolicyRequest{Policy: pb.StepperPolicy_NoWait, MeasuredDelay: &duration.Duration{Seconds: 1}})
	assert.NotNilf(t, err, "SetPolicy unexpectedly succeeded with an invalid duration: %v", err)

	_, err = client.SetPolicy(ctx, &pb.PolicyRequest{Policy: pb.StepperPolicy_NoWait, MeasuredDelay: &duration.Duration{Seconds: 0}})
	assert.Nilf(t, err, "SetPolicy unexpectedly failed: %v", err)

	_, err = client.SetPolicy(ctx, &pb.PolicyRequest{Policy: pb.StepperPolicy_NoWait, MeasuredDelay: &duration.Duration{Seconds: 2}})
	assert.NotNilf(t, err, "SetPolicy unexpectedly succeeded with an invalid duration: %v", err)
}

func TestInvalidDelay(t *testing.T) {
	ctx, conn := commonSetup(t)
	defer func() { _ = conn.Close() }()

	client = pb.NewStepperClient(conn)

	_, err := client.SetPolicy(ctx, &pb.PolicyRequest{Policy: pb.StepperPolicy_NoWait, MeasuredDelay: &duration.Duration{Seconds: 0}})
	assert.Nilf(t, err, "SetPolicy unexpectedly failed: %v", err)

	_, err = client.Delay(ctx, &pb.DelayRequest{AtLeast: -1, Jitter: 0})
	assert.NotNil(t, err, "Delay unexpectedly succeeded with an invalid base delay time")

	_, err = client.Delay(ctx, &pb.DelayRequest{AtLeast: 1, Jitter: -1})
	assert.NotNil(t, err, "Delay unexpectedly succeeded with an invalid jitter")
}

func TestStepper_NoWait(t *testing.T) {
	ctx, conn := commonSetup(t)
	defer func() { _ = conn.Close() }()

	client = pb.NewStepperClient(conn)

	// These need to execute in a particular order, so we're calling them as
	// included subtests in this unit test

	testSetPolicy(t, ctx, pb.StepperPolicy_NoWait, pb.StepperPolicy_Manual, 0)
	testNow(t, ctx, 0)
	testDelay(t, ctx, 1, 2)
}

func TestStepper_Measured(t *testing.T) {
	ctx, conn := commonSetup(t)
	defer func() { _ = conn.Close() }()

	client = pb.NewStepperClient(conn)

	// These need to execute in a particular order, so we're calling them as
	// included subtests in this unit test

	testSetPolicy(t, ctx, pb.StepperPolicy_Measured, pb.StepperPolicy_Manual, 1)
	testNow(t, ctx, 0)
	time.Sleep(time.Duration(2) * time.Second)

	callNowAtLeast(t, ctx, 1)

	t.Log("Now subtest complete")
	testDelay(t, ctx, 3, 2)
}

func TestStepper_Manual(t *testing.T) {
	ctx, conn := commonSetup(t)
	defer func() { _ = conn.Close() }()

	client = pb.NewStepperClient(conn)

	// These need to execute in a particular order, so we're calling them as
	// included subtests in this unit test

	testSetPolicy(t, ctx, pb.StepperPolicy_Manual, pb.StepperPolicy_Measured, 0)
	testNow(t, ctx, 0)
	testStep(t, ctx, 1)

	ch := make(chan bool)

	go func(res chan<- bool) {
		rsp, err := client.Delay(ctx, &pb.DelayRequest{AtLeast: 3, Jitter: 0})
		assert.Nilf(t, err, "Delay called failed, returned %v", err)
		assert.Equal(t, rsp.Ticks, int64(3), "Delay returned an invalid time.  Should be 3, but was %d", rsp.Ticks)

		res <- true
	}(ch)

	checkForEarlyCompletion(t, ch, 1, "Delay")

	callStep(t, ctx, 2)
	checkForEarlyCompletion(t, ch, 1, "Delay")

	callStep(t, ctx, 3)
	checkForLateCompletion(t, ch, 1, "Delay")

	t.Log("DelayManual subtest complete")
}
