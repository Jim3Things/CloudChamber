package frontend

import (
    "fmt"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"

    ts "github.com/Jim3Things/CloudChamber/internal/clients/timestamp"
    "github.com/Jim3Things/CloudChamber/internal/tracing/exporters/unit_test"
    pb "github.com/Jim3Things/CloudChamber/pkg/protos/Stepper"
    "github.com/Jim3Things/CloudChamber/pkg/protos/common"
)

const (
	stepperURI    = "/api/stepper"
)

func testStepperPath() string { return baseURI + stepperURI }

func testStepperReset(t *testing.T) {
    err := ts.Reset()
    assert.Nilf(t, err, "Unexpected error when resetting the stepper service)")
}

func testStepperSetManual(t *testing.T, match int64) {
    request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s", testStepperPath(), "?mode=manual"), nil)
    request.Header.Set("If-Match", fmt.Sprintf("%v", match))

    response := doHTTP(request, nil)

    assert.Equal(t, http.StatusOK, response.StatusCode)
    assert.Equal(t, fmt.Sprintf("%v", match + 1), response.Header.Get("ETag"))
}

func testStepperGetNow(t *testing.T) *common.Timestamp {
    request := httptest.NewRequest("GET", fmt.Sprintf("%s%s", testStepperPath(), "/now"), nil)
    response := doHTTP(request, nil)

    assert.Equal(t, http.StatusOK, response.StatusCode)

    res := &common.Timestamp{}
    err := getJsonBody(response, res)

    assert.Nilf(t, err, "Unexpected error, err: %v", err)

    return res
}

func testStepperGetStatus(t *testing.T) *pb.StatusResponse {
    request := httptest.NewRequest("GET", testStepperPath(), nil)
    response := doHTTP(request, nil)

    assert.Equal(t, http.StatusOK, response.StatusCode)

    res := &pb.StatusResponse{}
    err := getJsonBody(response, res)

    assert.Nilf(t, err, "Unexpected error, err: %v", err)

    return res
}

func testStepperAdvance(t *testing.T) *common.Timestamp {
    request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s", testStepperPath(), "?advance"), nil)
    response := doHTTP(request, nil)
    assert.Equal(t, http.StatusOK, response.StatusCode)

    res := &common.Timestamp{}
    err := getJsonBody(response, res)

    assert.Nilf(t, err, "Unexpected error, err: %v", err)

    return res
}

func testStepperAfter(t *testing.T, after int64) *common.Timestamp  {
    request := httptest.NewRequest("GET", fmt.Sprintf("%s/now?after=%d", testStepperPath(), after), nil)
    response := doHTTP(request, nil)

    assert.Equal(t, http.StatusOK, response.StatusCode)

    res := &common.Timestamp{}
    err := getJsonBody(response, res)

    assert.Nilf(t, err, "Unexpected error, err: %v", err)

    return res
}

func TestStepperGetStatus(t *testing.T) {
    unit_test.SetTesting(t)
    defer unit_test.SetTesting(nil)

    res := testStepperGetStatus(t)
    t.Log(res)
}

func TestStepperSetManual(t *testing.T) {
    unit_test.SetTesting(t)
    defer unit_test.SetTesting(nil)

    testStepperReset(t)
    stat := testStepperGetStatus(t)
    testStepperSetManual(t, stat.Epoch)

    res := testStepperGetStatus(t)
    assert.Equal(t, pb.StepperPolicy_Manual, res.Policy, "Unexpected policy")
    assert.Equal(t, int64(0), res.MeasuredDelay.Seconds, "Unexpected delay")
    assert.Equal(t, int32(0), res.MeasuredDelay.Nanos, "Unexpected delay")
    assert.Equal(t, int64(0), res.Now.Ticks, "Unexpected current time")
    assert.Equal(t, int64(0), res.WaiterCount, "Unexpected active waiter count")
    assert.Equal(t, stat.Epoch + 1, res.Epoch, "Unexpected epoch value")
}

func TestStepperSetModeInvalid(t *testing.T) {
    unit_test.SetTesting(t)
    defer unit_test.SetTesting(nil)

    testStepperReset(t)
    res := testStepperGetStatus(t)
    testStepperSetManual(t, res.Epoch)

    request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s", testStepperPath(), "?mode=badChoice"), nil)
    request.Header.Set("If-Match", "-1")

    response := doHTTP(request, nil)

    assert.Equal(t, http.StatusBadRequest, response.StatusCode)

    body, err := getBody(response)
    assert.Nil(t, err)
    assert.Equal(t, "CloudChamber: mode \"badChoice\" is invalid.  Supported modes are 'manual' and 'automatic'\n", string(body))

    res = testStepperGetStatus(t)
    assert.Equal(t, pb.StepperPolicy_Manual, res.Policy, "Unexpected policy")
    assert.Equal(t, int64(0), res.MeasuredDelay.Seconds, "Unexpected delay")
    assert.Equal(t, int32(0), res.MeasuredDelay.Nanos, "Unexpected delay")
    assert.Equal(t, int64(0), res.Now.Ticks, "Unexpected current time")
    assert.Equal(t, int64(0), res.WaiterCount, "Unexpected active waiter count")
}

func TestStepperSetModeBadEpoch(t *testing.T) {
    unit_test.SetTesting(t)
    defer unit_test.SetTesting(nil)

    testStepperReset(t)
    stat := testStepperGetStatus(t)
    testStepperSetManual(t, stat.Epoch)

    request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s", testStepperPath(), "?mode=manual"), nil)
    request.Header.Set("If-Match", fmt.Sprintf("%v", stat.Epoch))

    response := doHTTP(request, nil)

    assert.Equal(t, http.StatusBadRequest, response.StatusCode)

    body, err := getBody(response)
    assert.Nil(t, err)
    assert.Equal(t, "CloudChamber: Set simulated time policy operation failed\n", string(body))

    res := testStepperGetStatus(t)
    assert.Equal(t, pb.StepperPolicy_Manual, res.Policy, "Unexpected policy")
    assert.Equal(t, int64(0), res.MeasuredDelay.Seconds, "Unexpected delay")
    assert.Equal(t, int32(0), res.MeasuredDelay.Nanos, "Unexpected delay")
    assert.Equal(t, int64(0), res.Now.Ticks, "Unexpected current time")
    assert.Equal(t, int64(0), res.WaiterCount, "Unexpected active waiter count")
    assert.Equal(t, stat.Epoch + 1, res.Epoch, "Unexpected epoch")
}

func TestStepperAdvanceOne(t *testing.T) {
    unit_test.SetTesting(t)
    defer unit_test.SetTesting(nil)

    testStepperReset(t)
    stat := testStepperGetStatus(t)
    testStepperSetManual(t, stat.Epoch)

    res := testStepperGetNow(t)
    assert.Equal(t, int64(0), res.Ticks, "Time expected to be reset, was %d", res.Ticks)

    res2 := testStepperAdvance(t)

    assert.Equal(t, int64(1), res2.Ticks - res.Ticks, "time expected to advance by 1, old: %d, new: %d", res.Ticks, res2.Ticks)
}

func TestStepperAdvanceTwo(t *testing.T) {
    unit_test.SetTesting(t)
    defer unit_test.SetTesting(nil)

    status := testStepperGetStatus(t)

    testStepperSetManual(t, status.Epoch)
    res := testStepperGetNow(t)

    request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s", testStepperPath(), "?advance=2"), nil)
    response := doHTTP(request, nil)
    assert.Equal(t, http.StatusOK, response.StatusCode)

    res2 := &common.Timestamp{}
    err := getJsonBody(response, res2)

    assert.Nilf(t, err, "Unexpected error, err: %v", err)

    assert.Equal(t, int64(2), res2.Ticks - res.Ticks, "time expected to advance by 1, old: %d, new: %d", res.Ticks, res2.Ticks)
}

func TestStepperAdvanceNotANumber(t *testing.T) {
    unit_test.SetTesting(t)
    defer unit_test.SetTesting(nil)

    stat := testStepperGetStatus(t)
    testStepperSetManual(t, stat.Epoch)

    request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s", testStepperPath(), "?advance=two"), nil)
    response := doHTTP(request, nil)
    assert.Equal(t, http.StatusBadRequest, response.StatusCode)
    body, err := getBody(response)
    assert.Nil(t, err)
    assert.Equal(t, "CloudChamber: requested rate \"two\" could not be parsed as a positive decimal number\n", string(body))
}

func TestStepperAdvanceMinusOne(t *testing.T) {
    unit_test.SetTesting(t)
    defer unit_test.SetTesting(nil)

    stat := testStepperGetStatus(t)
    testStepperSetManual(t, stat.Epoch)

    request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s", testStepperPath(), "?advance=-1"), nil)
    response := doHTTP(request, nil)
    assert.Equal(t, http.StatusBadRequest, response.StatusCode)
    body, err := getBody(response)
    assert.Nil(t, err)
    assert.Equal(t, "CloudChamber: requested rate \"-1\" could not be parsed as a positive decimal number\n", string(body))
}

func TestStepperSetManualBadRate(t *testing.T) {
    unit_test.SetTesting(t)
    defer unit_test.SetTesting(nil)

    request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s", testStepperPath(), "?mode=manual=10"), nil)
    response := doHTTP(request, nil)

    assert.Equal(t, http.StatusBadRequest, response.StatusCode)
}

func TestStepperAfter(t *testing.T) {
    var afterHit = false

    unit_test.SetTesting(t)
    defer unit_test.SetTesting(nil)

    stat := testStepperGetStatus(t)
    testStepperSetManual(t, stat.Epoch)

    res := testStepperGetNow(t)
    go func(after int64) {
        res := testStepperAfter(t, after)

        assert.Less(t, after, res.Ticks)
        afterHit = true
    }(res.Ticks)

    _ = testStepperAdvance(t)
    time.Sleep(time.Duration(2) * time.Second)
    assert.True(t, afterHit)
}

func TestStepperAfterBadPastTick(t *testing.T) {
    unit_test.SetTesting(t)
    defer unit_test.SetTesting(nil)

    stat := testStepperGetStatus(t)
    testStepperSetManual(t, stat.Epoch)

    res := testStepperGetNow(t)
    _ = testStepperAdvance(t)

    res2 := testStepperAfter(t, res.Ticks)
    assert.Less(t, res.Ticks, res2.Ticks)
}

func TestStepperAfterBadTick(t *testing.T) {
    unit_test.SetTesting(t)
    defer unit_test.SetTesting(nil)

    stat := testStepperGetStatus(t)
    testStepperSetManual(t, stat.Epoch)

    request := httptest.NewRequest("GET", fmt.Sprintf("%s/now?after=-1", testStepperPath()), nil)
    response := doHTTP(request, nil)

    assert.Equal(t, http.StatusBadRequest, response.StatusCode)
}

func TestStepperGetNow(t *testing.T) {
    unit_test.SetTesting(t)
    defer unit_test.SetTesting(nil)

    res := testStepperGetNow(t)
    t.Log(res.Ticks)

    res2 := testStepperGetNow(t)

    assert.Equal(t, res.Ticks, res2.Ticks, "times with no advance do not match, %v != %v", res.Ticks, res2.Ticks)
}
