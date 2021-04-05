package frontend

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/Jim3Things/CloudChamber/simulation/internal/common"
	pbc "github.com/Jim3Things/CloudChamber/simulation/pkg/protos/common"
	pb "github.com/Jim3Things/CloudChamber/simulation/pkg/protos/services"
)

type StepperTestSuite struct {
	testSuiteCore
}

func (ts *StepperTestSuite) stepperPath() string { return ts.baseURI + "/api/stepper" }

func (ts *StepperTestSuite) setManual(match int64, cookies []*http.Cookie) []*http.Cookie {
	assert := ts.Assert()

	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s", ts.stepperPath(), "?mode=manual"), nil)
	request.Header.Set("If-Match", formatAsEtag(match))

	response := ts.doHTTP(request, cookies)

	assert.Equal(http.StatusOK, response.StatusCode)
	assert.Equal(formatAsEtag(match+1), response.Header.Get("ETag"))

	return response.Cookies()
}

func (ts *StepperTestSuite) getNow(cookies []*http.Cookie) (*pbc.Timestamp, []*http.Cookie) {
	assert := ts.Assert()

	request := httptest.NewRequest("GET", fmt.Sprintf("%s%s", ts.stepperPath(), "/now"), nil)
	response := ts.doHTTP(request, cookies)

	assert.Equal(http.StatusOK, response.StatusCode)

	res := &pbc.Timestamp{}
	err := ts.getJSONBody(response, res)

	assert.NoError(err, "Unexpected error, err: %v", err)

	return res, response.Cookies()
}

func (ts *StepperTestSuite) getStatus(cookies []*http.Cookie) (*pb.StatusResponse, []*http.Cookie) {
	assert := ts.Assert()

	request := httptest.NewRequest("GET", ts.stepperPath(), nil)
	response := ts.doHTTP(request, cookies)

	assert.Equal(http.StatusOK, response.StatusCode)

	res := &pb.StatusResponse{}
	err := ts.getJSONBody(response, res)

	assert.NoError(err, "Unexpected error, err: %v", err)

	return res, response.Cookies()
}

func (ts *StepperTestSuite) advance(cookies []*http.Cookie) (*pbc.Timestamp, []*http.Cookie) {
	assert := ts.Assert()

	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s", ts.stepperPath(), "?advance"), nil)
	response := ts.doHTTP(request, cookies)
	assert.Equal(http.StatusOK, response.StatusCode)

	res := &pbc.Timestamp{}
	err := ts.getJSONBody(response, res)

	assert.NoError(err, "Unexpected error, err: %v", err)

	return res, response.Cookies()
}

func (ts *StepperTestSuite) after(after int64, cookies []*http.Cookie) (*pb.StatusResponse, []*http.Cookie) {
	require := ts.Require()

	request := httptest.NewRequest("GET", fmt.Sprintf("%s/now?after=%d", ts.stepperPath(), after), nil)
	response := ts.doHTTP(request, cookies)

	require.Equal(http.StatusOK, response.StatusCode)

	res := &pb.StatusResponse{}
	err := ts.getJSONBody(response, res)

	require.NoError(err)

	return res, response.Cookies()
}

func (ts *StepperTestSuite) TestGetStatus() {
	log := ts.T().Log

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	res, cookies := ts.getStatus(response.Cookies())
	log(res)

	ts.doLogout(ts.randomCase(ts.adminAccountName()), cookies)
}

func (ts *StepperTestSuite) TestSetManual() {
	assert := ts.Assert()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	stat, cookies := ts.getStatus(response.Cookies())
	cookies = ts.setManual(stat.Epoch, cookies)

	res, cookies := ts.getStatus(cookies)
	assert.Equal(pb.StepperPolicy_Manual, res.Policy, "Unexpected policy")
	assert.Equal(int64(0), res.MeasuredDelay.Seconds, "Unexpected delay")
	assert.Equal(int32(0), res.MeasuredDelay.Nanos, "Unexpected delay")
	assert.Equal(int64(0), res.Now, "Unexpected current time")
	assert.LessOrEqual(int64(0), res.WaiterCount, "Unexpected active waiter count")
	assert.Equal(stat.Epoch+1, res.Epoch, "Unexpected epoch value")

	ts.doLogout(ts.randomCase(ts.adminAccountName()), cookies)
}

func (ts *StepperTestSuite) TestSetManualNoPrivilege() {
	require := ts.Require()

	response := ts.doLogin(ts.adminAccountName(), ts.adminPassword(), nil)

	stat, cookies := ts.getStatus(response.Cookies())
	cookies = ts.setManual(stat.Epoch, cookies)

	_, cookies = ts.ensureAccount(ts.aliceName(), ts.aliceDef, cookies)
	response = ts.doLogout(ts.adminAccountName(), cookies)

	response = ts.doLogin(ts.aliceName(), ts.alicePassword(), response.Cookies())
	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s", ts.stepperPath(), "?mode=manual"), nil)
	response = ts.doHTTP(request, response.Cookies())
	_, err := ts.getBody(response)
	require.NoError(err)

	require.Equal(http.StatusForbidden, response.StatusCode)

	ts.doLogout(ts.aliceName(), response.Cookies())
}

func (ts *StepperTestSuite) TestSetModeInvalid() {
	assert := ts.Assert()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	res, cookies := ts.getStatus(response.Cookies())
	cookies = ts.setManual(res.Epoch, cookies)

	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s", ts.stepperPath(), "?mode=badChoice"), nil)
	request.Header.Set("If-Match", formatAsEtag(-1))

	response = ts.doHTTP(request, cookies)

	assert.Equal(http.StatusBadRequest, response.StatusCode)

	body, err := ts.getBody(response)
	assert.NoError(err)
	assert.Equal(
		"CloudChamber: mode \"badChoice\" is invalid.  Supported modes are 'manual' and 'automatic'\n",
		string(body))

	res, cookies = ts.getStatus(response.Cookies())
	assert.Equal(pb.StepperPolicy_Manual, res.Policy, "Unexpected policy")
	assert.Equal(int64(0), res.MeasuredDelay.Seconds, "Unexpected delay")
	assert.Equal(int32(0), res.MeasuredDelay.Nanos, "Unexpected delay")
	assert.Equal(int64(0), res.Now, "Unexpected current time")
	assert.LessOrEqual(int64(0), res.WaiterCount, "Unexpected active waiter count")

	ts.doLogout(ts.randomCase(ts.adminAccountName()), cookies)
}

func (ts *StepperTestSuite) TestSetModeBadEpoch() {
	assert := ts.Assert()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	stat, cookies := ts.getStatus(response.Cookies())
	cookies = ts.setManual(stat.Epoch, cookies)

	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s", ts.stepperPath(), "?mode=manual"), nil)
	request.Header.Set("If-Match", formatAsEtag(stat.Epoch))

	response = ts.doHTTP(request, cookies)

	assert.Equal(http.StatusBadRequest, response.StatusCode)

	body, err := ts.getBody(response)
	assert.NoError(err)
	assert.Equal("CloudChamber: Set simulated time policy operation failed\n", string(body))

	res, cookies := ts.getStatus(response.Cookies())
	assert.Equal(pb.StepperPolicy_Manual, res.Policy, "Unexpected policy")
	assert.Equal(int64(0), res.MeasuredDelay.Seconds, "Unexpected delay")
	assert.Equal(int32(0), res.MeasuredDelay.Nanos, "Unexpected delay")
	assert.Equal(int64(0), res.Now, "Unexpected current time")
	assert.LessOrEqual(int64(0), res.WaiterCount, "Unexpected active waiter count")
	assert.Equal(stat.Epoch+1, res.Epoch, "Unexpected epoch")

	ts.doLogout(ts.randomCase(ts.adminAccountName()), cookies)
}

func (ts *StepperTestSuite) TestAdvanceOne() {
	assert := ts.Assert()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	stat, cookies := ts.getStatus(response.Cookies())
	cookies = ts.setManual(stat.Epoch, cookies)

	res, cookies := ts.getNow(cookies)
	assert.Equal(int64(0), res.Ticks, "Time expected to be reset, was %d", res.Ticks)

	res2, cookies := ts.advance(cookies)

	assert.Equal(
		int64(1), res2.Ticks-res.Ticks,
		"time expected to advance by 1, old: %d, new: %d",
		res.Ticks, res2.Ticks)

	ts.doLogout(ts.randomCase(ts.adminAccountName()), cookies)
}

func (ts *StepperTestSuite) TestAdvanceTwo() {
	assert := ts.Assert()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	status, cookies := ts.getStatus(response.Cookies())

	cookies = ts.setManual(status.Epoch, cookies)
	res, cookies := ts.getNow(cookies)

	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s", ts.stepperPath(), "?advance=2"), nil)
	response = ts.doHTTP(request, cookies)
	assert.Equal(http.StatusOK, response.StatusCode)

	res2 := &pbc.Timestamp{}
	err := ts.getJSONBody(response, res2)

	assert.NoError(err, "Unexpected error, err: %v", err)

	assert.Equal(
		int64(2), res2.Ticks-res.Ticks,
		"time expected to advance by 2, old: %d, new: %d",
		res.Ticks, res2.Ticks)

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *StepperTestSuite) TestAdvanceNotANumber() {
	assert := ts.Assert()
	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	stat, cookies := ts.getStatus(response.Cookies())
	cookies = ts.setManual(stat.Epoch, cookies)

	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s", ts.stepperPath(), "?advance=two"), nil)
	response = ts.doHTTP(request, cookies)
	body, err := ts.getBody(response)

	assert.Equal(http.StatusBadRequest, response.StatusCode)

	assert.NoError(err)
	assert.Equal(
		"CloudChamber: requested rate \"two\" could not be parsed as a positive decimal number\n",
		string(body))

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *StepperTestSuite) TestAdvanceMinusOne() {
	assert := ts.Assert()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	stat, cookies := ts.getStatus(response.Cookies())
	cookies = ts.setManual(stat.Epoch, cookies)

	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s", ts.stepperPath(), "?advance=-1"), nil)
	response = ts.doHTTP(request, cookies)

	assert.Equal(http.StatusBadRequest, response.StatusCode)

	body, err := ts.getBody(response)

	assert.NoError(err)
	assert.Equal(
		"CloudChamber: requested rate \"-1\" could not be parsed as a positive decimal number\n",
		string(body))

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *StepperTestSuite) TestAdvanceNoPrivilege() {
	require := ts.Require()

	response := ts.doLogin(ts.adminAccountName(), ts.adminPassword(), nil)

	stat, cookies := ts.getStatus(response.Cookies())
	cookies = ts.setManual(stat.Epoch, cookies)

	_, cookies = ts.ensureAccount(ts.aliceName(), ts.aliceDef, cookies)
	response = ts.doLogout(ts.adminAccountName(), cookies)

	response = ts.doLogin(ts.aliceName(), ts.alicePassword(), response.Cookies())
	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s", ts.stepperPath(), "?advance"), nil)
	response = ts.doHTTP(request, response.Cookies())
	_, err := ts.getBody(response)
	require.NoError(err)

	require.Equal(http.StatusForbidden, response.StatusCode)

	ts.doLogout(ts.aliceName(), response.Cookies())
}

func (ts *StepperTestSuite) TestSetManualBadRate() {
	assert := ts.Assert()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s", ts.stepperPath(), "?mode=manual=10"), nil)
	response = ts.doHTTP(request, response.Cookies())

	assert.Equal(http.StatusBadRequest, response.StatusCode)

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *StepperTestSuite) TestAfter() {
	assert := ts.Assert()

	var cookies2 []*http.Cookie

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	stat, cookies := ts.getStatus(response.Cookies())
	cookies = ts.setManual(stat.Epoch, cookies)

	res, cookies := ts.getNow(cookies)

	ch := make(chan bool)

	go func(ch chan<- bool, after int64, cookies []*http.Cookie) {
		var waiter *pb.StatusResponse
		waiter, cookies2 = ts.after(after, cookies)

		assert.Less(after, waiter.Now)
		ch <- true
	}(ch, res.Ticks, cookies)

	_, _ = ts.advance(cookies)

	assert.True(common.CompleteWithin(ch, time.Duration(2)*time.Second))

	ts.doLogout(ts.randomCase(ts.adminAccountName()), cookies2)
}

func (ts *StepperTestSuite) TestAfterBadPastTick() {
	assert := ts.Assert()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	stat, cookies := ts.getStatus(response.Cookies())
	cookies = ts.setManual(stat.Epoch, cookies)

	res, cookies := ts.getNow(cookies)
	_, cookies = ts.advance(cookies)

	res2, cookies := ts.after(res.Ticks, cookies)
	assert.Less(res.Ticks, res2.Now)

	ts.doLogout(ts.randomCase(ts.adminAccountName()), cookies)
}

func (ts *StepperTestSuite) TestAfterBadTick() {
	assert := ts.Assert()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	stat, cookies := ts.getStatus(response.Cookies())
	cookies = ts.setManual(stat.Epoch, cookies)

	request := httptest.NewRequest("GET", fmt.Sprintf("%s/now?after=-1", ts.stepperPath()), nil)
	response = ts.doHTTP(request, cookies)

	assert.Equal(http.StatusBadRequest, response.StatusCode)

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *StepperTestSuite) TestGetNow() {
	assert := ts.Assert()
	log := ts.T().Log

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	res, cookies := ts.getNow(response.Cookies())
	log(res.Ticks)

	res2, cookies := ts.getNow(cookies)

	assert.Equal(res.Ticks, res2.Ticks, "times with no advance do not match, %v != %v", res.Ticks, res2.Ticks)

	ts.doLogout(ts.randomCase(ts.adminAccountName()), cookies)
}

func TestStepperTestSuite(t *testing.T) {
	suite.Run(t, new(StepperTestSuite))
}
