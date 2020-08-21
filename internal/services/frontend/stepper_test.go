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
	stepperURI = "/api/stepper"
)

func testStepperPath() string { return baseURI + stepperURI }

func testStepperReset(t *testing.T) {
	err := ts.Reset()
	assert.Nilf(t, err, "Unexpected error when resetting the stepper service)")
}

func testStepperSetManual(t *testing.T, match int64, cookies []*http.Cookie) []*http.Cookie {
	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s", testStepperPath(), "?mode=manual"), nil)
	request.Header.Set("If-Match", fmt.Sprintf("%v", match))

	response := doHTTP(request, cookies)

	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, fmt.Sprintf("%v", match + 1), response.Header.Get("ETag"))

	return response.Cookies()
}

func testStepperGetNow(t *testing.T, cookies []*http.Cookie) (*common.Timestamp, []*http.Cookie) {
	request := httptest.NewRequest("GET", fmt.Sprintf("%s%s", testStepperPath(), "/now"), nil)
	response := doHTTP(request, cookies)

	assert.Equal(t, http.StatusOK, response.StatusCode)

	res := &common.Timestamp{}
	err := getJSONBody(response, res)

	assert.Nilf(t, err, "Unexpected error, err: %v", err)

	return res, response.Cookies()
}

func testStepperGetStatus(t *testing.T, cookies []*http.Cookie) (*pb.StatusResponse, []*http.Cookie) {
	request := httptest.NewRequest("GET", testStepperPath(), nil)
	response := doHTTP(request, cookies)

	assert.Equal(t, http.StatusOK, response.StatusCode)

	res := &pb.StatusResponse{}
	err := getJSONBody(response, res)

	assert.Nilf(t, err, "Unexpected error, err: %v", err)

	return res, response.Cookies()
}

func testStepperAdvance(t *testing.T, cookies []*http.Cookie) (*common.Timestamp, []*http.Cookie) {
	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s", testStepperPath(), "?advance"), nil)
	response := doHTTP(request, cookies)
	assert.Equal(t, http.StatusOK, response.StatusCode)

	res := &common.Timestamp{}
	err := getJSONBody(response, res)

	assert.Nilf(t, err, "Unexpected error, err: %v", err)

	return res, response.Cookies()
}

func testStepperAfter(t *testing.T, after int64, cookies []*http.Cookie) (*common.Timestamp, []*http.Cookie) {
	request := httptest.NewRequest("GET", fmt.Sprintf("%s/now?after=%d", testStepperPath(), after), nil)
	response := doHTTP(request, cookies)

	assert.Equal(t, http.StatusOK, response.StatusCode)

	res := &common.Timestamp{}
	err := getJSONBody(response, res)

	assert.Nilf(t, err, "Unexpected error, err: %v", err)

	return res, response.Cookies()
}

func TestStepperGetStatus(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	response := doLogin(t, randomCase(adminAccountName), adminPassword, nil)

	res, cookies := testStepperGetStatus(t, response.Cookies())
	t.Log(res)

	doLogout(t, randomCase(adminAccountName), cookies)
}

func TestStepperSetManual(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	testStepperReset(t)
	response := doLogin(t, randomCase(adminAccountName), adminPassword, nil)

	stat, cookies := testStepperGetStatus(t, response.Cookies())
	cookies = testStepperSetManual(t, stat.Epoch, cookies)

	res, cookies := testStepperGetStatus(t, cookies)
	assert.Equal(t, pb.StepperPolicy_Manual, res.Policy, "Unexpected policy")
	assert.Equal(t, int64(0), res.MeasuredDelay.Seconds, "Unexpected delay")
	assert.Equal(t, int32(0), res.MeasuredDelay.Nanos, "Unexpected delay")
	assert.Equal(t, int64(0), res.Now.Ticks, "Unexpected current time")
	assert.Equal(t, int64(0), res.WaiterCount, "Unexpected active waiter count")
	assert.Equal(t, stat.Epoch + 1, res.Epoch, "Unexpected epoch value")

	doLogout(t, randomCase(adminAccountName), cookies)
}

func TestStepperSetModeInvalid(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	testStepperReset(t)
	response := doLogin(t, randomCase(adminAccountName), adminPassword, nil)

	res, cookies := testStepperGetStatus(t, response.Cookies())
	cookies = testStepperSetManual(t, res.Epoch, cookies)

	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s", testStepperPath(), "?mode=badChoice"), nil)
	request.Header.Set("If-Match", "-1")

	response = doHTTP(request, cookies)

	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	body, err := getBody(response)
	assert.Nil(t, err)
	assert.Equal(t, "CloudChamber: mode \"badChoice\" is invalid.  Supported modes are 'manual' and 'automatic'\n", string(body))

	res, cookies = testStepperGetStatus(t, response.Cookies())
	assert.Equal(t, pb.StepperPolicy_Manual, res.Policy, "Unexpected policy")
	assert.Equal(t, int64(0), res.MeasuredDelay.Seconds, "Unexpected delay")
	assert.Equal(t, int32(0), res.MeasuredDelay.Nanos, "Unexpected delay")
	assert.Equal(t, int64(0), res.Now.Ticks, "Unexpected current time")
	assert.Equal(t, int64(0), res.WaiterCount, "Unexpected active waiter count")

	doLogout(t, randomCase(adminAccountName), cookies)
}

func TestStepperSetModeBadEpoch(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	testStepperReset(t)
	response := doLogin(t, randomCase(adminAccountName), adminPassword, nil)

	stat, cookies := testStepperGetStatus(t, response.Cookies())
	cookies = testStepperSetManual(t, stat.Epoch, cookies)

	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s", testStepperPath(), "?mode=manual"), nil)
	request.Header.Set("If-Match", fmt.Sprintf("%v", stat.Epoch))

	response = doHTTP(request, cookies)

	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	body, err := getBody(response)
	assert.Nil(t, err)
	assert.Equal(t, "CloudChamber: Set simulated time policy operation failed\n", string(body))

	res, cookies := testStepperGetStatus(t, response.Cookies())
	assert.Equal(t, pb.StepperPolicy_Manual, res.Policy, "Unexpected policy")
	assert.Equal(t, int64(0), res.MeasuredDelay.Seconds, "Unexpected delay")
	assert.Equal(t, int32(0), res.MeasuredDelay.Nanos, "Unexpected delay")
	assert.Equal(t, int64(0), res.Now.Ticks, "Unexpected current time")
	assert.Equal(t, int64(0), res.WaiterCount, "Unexpected active waiter count")
	assert.Equal(t, stat.Epoch + 1, res.Epoch, "Unexpected epoch")

	doLogout(t, randomCase(adminAccountName), cookies)
}

func TestStepperAdvanceOne(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	testStepperReset(t)
	response := doLogin(t, randomCase(adminAccountName), adminPassword, nil)

	stat, cookies := testStepperGetStatus(t, response.Cookies())
	cookies = testStepperSetManual(t, stat.Epoch, cookies)

	res, cookies := testStepperGetNow(t, cookies)
	assert.Equal(t, int64(0), res.Ticks, "Time expected to be reset, was %d", res.Ticks)

	res2, cookies := testStepperAdvance(t, cookies)

	assert.Equal(t, int64(1), res2.Ticks-res.Ticks, "time expected to advance by 1, old: %d, new: %d", res.Ticks, res2.Ticks)

	doLogout(t, randomCase(adminAccountName), cookies)
}

func TestStepperAdvanceTwo(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	response := doLogin(t, randomCase(adminAccountName), adminPassword, nil)

	status, cookies := testStepperGetStatus(t, response.Cookies())

	cookies = testStepperSetManual(t, status.Epoch, cookies)
	res, cookies := testStepperGetNow(t, cookies)

	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s", testStepperPath(), "?advance=2"), nil)
	response = doHTTP(request, cookies)
	assert.Equal(t, http.StatusOK, response.StatusCode)

	res2 := &common.Timestamp{}
	err := getJSONBody(response, res2)

	assert.Nilf(t, err, "Unexpected error, err: %v", err)

	assert.Equal(t, int64(2), res2.Ticks-res.Ticks, "time expected to advance by 1, old: %d, new: %d", res.Ticks, res2.Ticks)

	doLogout(t, randomCase(adminAccountName), response.Cookies())
}

func TestStepperAdvanceNotANumber(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	response := doLogin(t, randomCase(adminAccountName), adminPassword, nil)

	stat, cookies := testStepperGetStatus(t, response.Cookies())
	cookies = testStepperSetManual(t, stat.Epoch, cookies)

	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s", testStepperPath(), "?advance=two"), nil)
	response = doHTTP(request, cookies)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	body, err := getBody(response)

	assert.Nil(t, err)
	assert.Equal(t, "CloudChamber: requested rate \"two\" could not be parsed as a positive decimal number\n", string(body))

	doLogout(t, randomCase(adminAccountName), response.Cookies())
}

func TestStepperAdvanceMinusOne(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	response := doLogin(t, randomCase(adminAccountName), adminPassword, nil)

	stat, cookies := testStepperGetStatus(t, response.Cookies())
	cookies = testStepperSetManual(t, stat.Epoch, cookies)

	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s", testStepperPath(), "?advance=-1"), nil)
	response = doHTTP(request, cookies)

	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	body, err := getBody(response)

	assert.Nil(t, err)
	assert.Equal(t, "CloudChamber: requested rate \"-1\" could not be parsed as a positive decimal number\n", string(body))

	doLogout(t, randomCase(adminAccountName), response.Cookies())
}

func TestStepperSetManualBadRate(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	response := doLogin(t, randomCase(adminAccountName), adminPassword, nil)

	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s", testStepperPath(), "?mode=manual=10"), nil)
	response = doHTTP(request, response.Cookies())

	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	doLogout(t, randomCase(adminAccountName), response.Cookies())
}

func TestStepperAfter(t *testing.T) {
	var afterHit = false
	var cookies2 []*http.Cookie

	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	response := doLogin(t, randomCase(adminAccountName), adminPassword, nil)

	stat, cookies := testStepperGetStatus(t, response.Cookies())
	cookies = testStepperSetManual(t, stat.Epoch, cookies)

	res, cookies := testStepperGetNow(t, cookies)

	go func(after int64, cookies []*http.Cookie) {
		res, cookies2 = testStepperAfter(t, after, cookies)

		assert.Less(t, after, res.Ticks)
		afterHit = true
	}(res.Ticks, cookies)

	_, _ = testStepperAdvance(t, cookies)
	time.Sleep(time.Duration(2) * time.Second)
	assert.True(t, afterHit)

	doLogout(t, randomCase(adminAccountName), cookies2)
}

func TestStepperAfterBadPastTick(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	response := doLogin(t, randomCase(adminAccountName), adminPassword, nil)

	stat, cookies := testStepperGetStatus(t, response.Cookies())
	cookies = testStepperSetManual(t, stat.Epoch, cookies)

	res, cookies := testStepperGetNow(t, cookies)
	_, cookies = testStepperAdvance(t, cookies)

	res2, cookies := testStepperAfter(t, res.Ticks, cookies)
	assert.Less(t, res.Ticks, res2.Ticks)

	doLogout(t, randomCase(adminAccountName), cookies)
}

func TestStepperAfterBadTick(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	response := doLogin(t, randomCase(adminAccountName), adminPassword, nil)

	stat, cookies := testStepperGetStatus(t, response.Cookies())
	cookies = testStepperSetManual(t, stat.Epoch, cookies)

	request := httptest.NewRequest("GET", fmt.Sprintf("%s/now?after=-1", testStepperPath()), nil)
	response = doHTTP(request, cookies)

	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	doLogout(t, randomCase(adminAccountName), response.Cookies())
}

func TestStepperGetNow(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	response := doLogin(t, randomCase(adminAccountName), adminPassword, nil)

	res, cookies := testStepperGetNow(t, response.Cookies())
	t.Log(res.Ticks)

	res2, cookies := testStepperGetNow(t, cookies)

	assert.Equal(t, res.Ticks, res2.Ticks, "times with no advance do not match, %v != %v", res.Ticks, res2.Ticks)

	doLogout(t, randomCase(adminAccountName), cookies)
}
