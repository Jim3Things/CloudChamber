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

    pb "../../pkg/protos/Stepper"

    "google.golang.org/grpc"
    "google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener
var client pb.StepperClient

func init() {
	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()
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

func testSetPolicy(t *testing.T, ctx context.Context, policy pb.StepperPolicy, badPolicy pb.StepperPolicy, seconds int64) {
    _, err := client.SetPolicy(ctx, &pb.PolicyRequest{Policy: policy, MeasuredDelay: &duration.Duration{ Seconds: seconds}} )
    assert.Nilf(t, err, "SetPolicy failed: %v", err)

    _, err = client.SetPolicy(ctx, &pb.PolicyRequest{Policy: policy, MeasuredDelay: &duration.Duration{ Seconds: seconds}} )
    assert.Nilf(t, err, "SetPolicy failed: %v", err)

    _, err = client.SetPolicy(ctx, &pb.PolicyRequest{Policy: policy, MeasuredDelay: &duration.Duration{ Seconds: seconds + 1}} )
    assert.NotNil(t, err, "SetPolicy succeeded, should have failed")

    _, err = client.SetPolicy(ctx, &pb.PolicyRequest{Policy: badPolicy, MeasuredDelay: &duration.Duration{ Seconds: seconds}} )
    assert.NotNil(t, err, "SetPolicy succeeded, should have failed")

    t.Log("SetPolicy subtest complete")
}

func callNow(t *testing.T, ctx context.Context) int64{
    resp, err := client.Now(ctx, &empty.Empty{})
    assert.Nilf(t, err, "Now failed: %v", err)

    return resp.Current
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

    resp, err := client.Delay(ctx, &pb.DelayRequest{ AtLeast: atLeast, Jitter: jitter })
    assert.Nilf(t, err, "Delay failed: %v, err")

    minLegal := atLeast
    maxLegal := start + atLeast + jitter

    assert.True(
        t,
        resp.Current >= minLegal && resp.Current <= maxLegal,
        "Delay out of range, should be %d - %d, is %d",
        minLegal,
        maxLegal,
        resp.Current)

    t.Log("Delay subtest complete")
}
func commonSetup(t *testing.T) (context.Context, *grpc.ClientConn) {
    Reset()
    ctx := context.Background()
    conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
    assert.Nilf(t, err, "Failed to dial bufnet: %v", err)

    return ctx, conn
}

func TestInvalidSetPolicyType(t *testing.T) {
    ctx, conn := commonSetup(t)
    defer conn.Close()

    client = pb.NewStepperClient(conn)

    _, err := client.SetPolicy(ctx, &pb.PolicyRequest{Policy: pb.StepperPolicy_Invalid, MeasuredDelay: &duration.Duration{ Seconds: 0}} )
    assert.NotNilf(t, err, "SetPolicy unexpectedly succeeded with an invalid policy type")
}

func TestInvalidSetPolicyManual(t *testing.T) {
    ctx, conn := commonSetup(t)
    defer conn.Close()

    client = pb.NewStepperClient(conn)

    _, err := client.SetPolicy(ctx, &pb.PolicyRequest{Policy: pb.StepperPolicy_Manual, MeasuredDelay: &duration.Duration{ Seconds: 1}} )
    assert.NotNilf(t, err, "SetPolicy unexpectedly succeeded with an invalid duration: %v", err)

    _, err = client.SetPolicy(ctx, &pb.PolicyRequest{Policy: pb.StepperPolicy_Manual, MeasuredDelay: &duration.Duration{ Seconds: 0}} )
    assert.Nilf(t, err, "SetPolicy unexpectedly failed: %v", err)

    _, err = client.SetPolicy(ctx, &pb.PolicyRequest{Policy: pb.StepperPolicy_Manual, MeasuredDelay: &duration.Duration{ Seconds: 2}} )
    assert.NotNilf(t, err, "SetPolicy unexpectedly succeeded with an invalid duration: %v", err)
}

func TestInvalidSetPolicyMeasured(t *testing.T) {
    ctx, conn := commonSetup(t)
    defer conn.Close()

    client = pb.NewStepperClient(conn)

    _, err := client.SetPolicy(ctx, &pb.PolicyRequest{Policy: pb.StepperPolicy_Measured, MeasuredDelay: &duration.Duration{ Seconds: 0}} )
    assert.NotNilf(t, err, "SetPolicy unexpectedly succeeded with an invalid duration: %v", err)

    _, err = client.SetPolicy(ctx, &pb.PolicyRequest{Policy: pb.StepperPolicy_Measured, MeasuredDelay: &duration.Duration{ Seconds: 1}} )
    assert.Nilf(t, err, "SetPolicy unexpectedly failed: %v", err)

    _, err = client.SetPolicy(ctx, &pb.PolicyRequest{Policy: pb.StepperPolicy_Measured, MeasuredDelay: &duration.Duration{ Seconds: 2}} )
    assert.NotNilf(t, err, "SetPolicy unexpectedly succeeded with an invalid duration: %v", err)

    // Now force a reset to clear out the free-running autostepper
    Reset()
    time.Sleep(time.Duration(2) * time.Second)
}

func TestInvalidSetPolicyNoWait(t *testing.T) {
    ctx, conn := commonSetup(t)
    defer conn.Close()

    client = pb.NewStepperClient(conn)

    _, err := client.SetPolicy(ctx, &pb.PolicyRequest{Policy: pb.StepperPolicy_NoWait, MeasuredDelay: &duration.Duration{ Seconds: 1}} )
    assert.NotNilf(t, err, "SetPolicy unexpectedly succeeded with an invalid duration: %v", err)

    _, err = client.SetPolicy(ctx, &pb.PolicyRequest{Policy: pb.StepperPolicy_NoWait, MeasuredDelay: &duration.Duration{ Seconds: 0}} )
    assert.Nilf(t, err, "SetPolicy unexpectedly failed: %v", err)

    _, err = client.SetPolicy(ctx, &pb.PolicyRequest{Policy: pb.StepperPolicy_NoWait, MeasuredDelay: &duration.Duration{ Seconds: 2}} )
    assert.NotNilf(t, err, "SetPolicy unexpectedly succeeded with an invalid duration: %v", err)
}

func TestInvalidDelay(t *testing.T) {
    ctx, conn := commonSetup(t)
    defer conn.Close()

    client = pb.NewStepperClient(conn)

    _, err := client.SetPolicy(ctx, &pb.PolicyRequest{Policy: pb.StepperPolicy_NoWait, MeasuredDelay: &duration.Duration{ Seconds: 0}} )
    assert.Nilf(t, err, "SetPolicy unexpectedly failed: %v", err)

    _, err = client.Delay(ctx, &pb.DelayRequest{AtLeast: -1, Jitter: 0 })
    assert.NotNil(t, err, "Delay unexpectedly succeeded with an invalid base delay time")

    _, err = client.Delay(ctx, &pb.DelayRequest{AtLeast: 1, Jitter: -1 })
    assert.NotNil(t, err, "Delay unexpectedly succeeded with an invalid jitter")
}

func TestInvalidSetToLatest(t *testing.T) {
    ctx, conn := commonSetup(t)
    defer conn.Close()

    client = pb.NewStepperClient(conn)

    _, err := client.SetPolicy(ctx, &pb.PolicyRequest{Policy: pb.StepperPolicy_NoWait, MeasuredDelay: &duration.Duration{ Seconds: 0}} )
    assert.Nilf(t, err, "SetPolicy unexpectedly failed: %v", err)

    _, err = client.SetToLatest(ctx, &pb.SetToLatestRequest{FirstTicks: -1, SecondTicks: 0})
    assert.NotNil(t, err, "SetToLatest unexpectedly succeeded with an invalid first ticks parameter")

    _, err = client.SetToLatest(ctx, &pb.SetToLatestRequest{FirstTicks: 0, SecondTicks: -1})
    assert.NotNil(t, err, "SetToLatest unexpectedly succeeded with an invalid second ticks parameter")
}

func TestStepper_NoWait(t *testing.T) {
    ctx, conn := commonSetup(t)
	defer conn.Close()

	client = pb.NewStepperClient(conn)

	// These need to execute in a particular order, so we're calling them as
	// included subtests in this unit test

	testSetPolicy(t, ctx, pb.StepperPolicy_NoWait, pb.StepperPolicy_Manual, 0)
    testNow(t, ctx, 0)
	testDelay(t, ctx, 1, 2)
}

func TestStepper_Measured(t *testing.T) {
    ctx, conn := commonSetup(t)
    defer conn.Close()

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
    defer conn.Close()

    client = pb.NewStepperClient(conn)

    // These need to execute in a particular order, so we're calling them as
    // included subtests in this unit test

    testSetPolicy(t, ctx, pb.StepperPolicy_Manual, pb.StepperPolicy_Measured, 0)
    testNow(t, ctx, 0)
    testStep(t, ctx, 1)

    var done = false
    go func(flag *bool) {
        rsp, err := client.Delay(ctx, &pb.DelayRequest{AtLeast: 3, Jitter: 0})
        assert.Nilf(t, err, "Delay called failed, returned %v", err)
        assert.Equal(t, rsp.Current, int64(3), "Delay returned an invalid time.  Should be 3, but was %d", rsp.Current)

        done = true
    }(&done)

    assert.False(t, done, "Delay completed early")

    callStep(t, ctx, 2)
    assert.False(t, done, "Delay completed early")

    callStep(t, ctx, 3)
    assert.True(t, done, "Delay did not complete on time")

    t.Log("DelayManual subtest complete")
}

// From here on we use the proof above that all the policies have working
// delay and sync mechanisms.  From here down we only use the manual policy,
// which allows the tests to run independently

func callSetToLatest(t *testing.T, ctx context.Context, first int64, second int64, expected int64, done *bool) {
    resp, err := client.SetToLatest(ctx, &pb.SetToLatestRequest{FirstTicks: first, SecondTicks: second })
    assert.Nilf(t, err, "SetToLatest failed unexpectedly: %v", err)
    assert.Equal(t, resp.Current, expected, "Invalid time returned: expected %d, was %d", expected, resp.Current)
    *done = true
}

func TestSetToLatest(t *testing.T) {
    ctx, conn := commonSetup(t)
    defer conn.Close()

    client = pb.NewStepperClient(conn)

    testSetPolicy(t, ctx, pb.StepperPolicy_Manual, pb.StepperPolicy_Measured, 0)

    callStep(t, ctx, 1)
    callStep(t, ctx, 2)

    done := false
    go callSetToLatest(t, ctx, 1, 0, 2, &done)

    time.Sleep(time.Duration(1) * time.Second)
    assert.True(t, done, "SetToLatest with a past deadline still waiting")

    done = false
    go callSetToLatest(t, ctx, 3, 0, 3, &done)

    time.Sleep(time.Duration(1) * time.Second)
    assert.False(t, done, "SetToLatest with a future deadline completed early")

    callStep(t, ctx, 3)
    time.Sleep(time.Duration(1) * time.Second)
    assert.True(t, done, "SetToLatest with a past deadline still waiting")
}

func callWaitForSync(t *testing.T, ctx context.Context, atLeast int64, expected int64, done *bool) {
    resp, err := client.WaitForSync(ctx, &pb.WaitForSyncRequest{AtLeast: atLeast})
    assert.Nilf(t, err, "WaitForSync failed unexpectedly: %v", err)
    assert.Equal(t, resp.Current, expected, "Invalid time returned: expected %d, was %d", expected, resp.Current)
    *done = true
}

func TestWaitForSync(t *testing.T) {
    ctx, conn := commonSetup(t)
    defer conn.Close()

    client = pb.NewStepperClient(conn)

    testSetPolicy(t, ctx, pb.StepperPolicy_Manual, pb.StepperPolicy_Measured, 0)

    callStep(t, ctx, 1)

    // Wait for a past time
    var done = false
    go callWaitForSync(t, ctx, 0, 1, &done)
    time.Sleep(time.Duration(1) * time.Second)
    assert.True(t, done, "WaitForStep with a past deadline still waiting")

    // Now wait for a future time
    done = false
    go callWaitForSync(t, ctx, 3, 3, &done)
    time.Sleep(time.Duration(1) * time.Second)
    assert.False(t, done, "WaitForStep with a future deadline completed early")

    callStep(t, ctx, 2)
    time.Sleep(time.Duration(1) * time.Second)
    assert.False(t, done, "WaitForStep with a future deadline completed early")

    callStep(t, ctx, 3)
    time.Sleep(time.Duration(1) * time.Second)
    assert.True(t, done, "WaitForStep with a past deadline still waiting")
}