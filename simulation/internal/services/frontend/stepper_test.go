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
	require := ts.Require()

	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s", ts.stepperPath(), "?mode=manual"), nil)
	request.Header.Set("If-Match", formatAsEtag(match))

	response := ts.doHTTP(request, cookies)

	require.HTTPRSuccess(response)
	require.Equal(formatAsEtag(match+1), response.Header.Get("ETag"))

	return response.Cookies()
}

func (ts *StepperTestSuite) getNow(cookies []*http.Cookie) (*pbc.Timestamp, []*http.Cookie) {
	require := ts.Require()

	request := httptest.NewRequest("GET", fmt.Sprintf("%s%s", ts.stepperPath(), "/now"), nil)
	response := ts.doHTTP(request, cookies)

	require.HTTPRSuccess(response)

	res := &pbc.Timestamp{}
	require.NoError(ts.getJSONBody(response, res))

	return res, response.Cookies()
}

func (ts *StepperTestSuite) getStatus(cookies []*http.Cookie) (*pb.StatusResponse, []*http.Cookie) {
	require := ts.Require()

	request := httptest.NewRequest("GET", ts.stepperPath(), nil)
	response := ts.doHTTP(request, cookies)

	require.HTTPRSuccess(response)

	res := &pb.StatusResponse{}
	require.NoError(ts.getJSONBody(response, res))

	return res, response.Cookies()
}

func (ts *StepperTestSuite) advance(cookies []*http.Cookie) (*pbc.Timestamp, []*http.Cookie) {
	require := ts.Require()

	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s", ts.stepperPath(), "?advance"), nil)
	response := ts.doHTTP(request, cookies)
	require.HTTPRSuccess(response)

	res := &pbc.Timestamp{}
	require.NoError(ts.getJSONBody(response, res))

	return res, response.Cookies()
}

func (ts *StepperTestSuite) after(after int64, cookies []*http.Cookie) (*pb.StatusResponse, []*http.Cookie) {
	require := ts.Require()

	request := httptest.NewRequest("GET", fmt.Sprintf("%s/now?after=%d", ts.stepperPath(), after), nil)
	response := ts.doHTTP(request, cookies)

	require.HTTPRSuccess(response)

	res := &pb.StatusResponse{}
	require.NoError(ts.getJSONBody(response, res))

	return res, response.Cookies()
}

func (ts *StepperTestSuite) TestGetStatus() {
	require := ts.Require()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	res, cookies := ts.getStatus(response.Cookies())
	require.Equal(pb.StepperPolicy_Manual, res.Policy)
	require.EqualValues(0, res.MeasuredDelay.Seconds)
	require.EqualValues(0, res.MeasuredDelay.Nanos)

	ts.doLogout(ts.randomCase(ts.adminAccountName()), cookies)
}

func (ts *StepperTestSuite) TestSetManual() {
	require := ts.Require()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	stat, cookies := ts.getStatus(response.Cookies())
	cookies = ts.setManual(stat.Epoch, cookies)

	res, cookies := ts.getStatus(cookies)
	require.Equal(pb.StepperPolicy_Manual, res.Policy, "Unexpected policy")
	require.Equal(int64(0), res.MeasuredDelay.Seconds, "Unexpected delay")
	require.Equal(int32(0), res.MeasuredDelay.Nanos, "Unexpected delay")
	require.Equal(int64(0), res.Now, "Unexpected current time")
	require.LessOrEqual(int64(0), res.WaiterCount, "Unexpected active waiter count")
	require.Equal(stat.Epoch+1, res.Epoch, "Unexpected epoch value")

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
	require.HTTPRStatusEqual(http.StatusForbidden, response)

	_ = ts.getBody(response)

	ts.doLogout(ts.aliceName(), response.Cookies())
}

func (ts *StepperTestSuite) TestSetModeInvalid() {
	require := ts.Require()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	res, cookies := ts.getStatus(response.Cookies())
	cookies = ts.setManual(res.Epoch, cookies)

	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s", ts.stepperPath(), "?mode=badChoice"), nil)
	request.Header.Set("If-Match", formatAsEtag(-1))

	response = ts.doHTTP(request, cookies)
	require.HTTPRStatusEqual(http.StatusBadRequest, response)

	body := ts.getBody(response)
	require.Equal(
		"CloudChamber: mode \"badChoice\" is invalid.  Supported modes are 'manual' and 'automatic'\n",
		string(body))

	res, cookies = ts.getStatus(response.Cookies())
	require.Equal(pb.StepperPolicy_Manual, res.Policy, "Unexpected policy")
	require.Equal(int64(0), res.MeasuredDelay.Seconds, "Unexpected delay")
	require.Equal(int32(0), res.MeasuredDelay.Nanos, "Unexpected delay")
	require.Equal(int64(0), res.Now, "Unexpected current time")
	require.LessOrEqual(int64(0), res.WaiterCount, "Unexpected active waiter count")

	ts.doLogout(ts.randomCase(ts.adminAccountName()), cookies)
}

func (ts *StepperTestSuite) TestSetModeBadEpoch() {
	require := ts.Require()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	stat, cookies := ts.getStatus(response.Cookies())
	cookies = ts.setManual(stat.Epoch, cookies)

	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s", ts.stepperPath(), "?mode=manual"), nil)
	request.Header.Set("If-Match", formatAsEtag(stat.Epoch))

	response = ts.doHTTP(request, cookies)
	require.HTTPRStatusEqual(http.StatusBadRequest, response)

	body := ts.getBody(response)
	require.Equal("CloudChamber: Set simulated time policy operation failed\n", string(body))

	res, cookies := ts.getStatus(response.Cookies())
	require.Equal(pb.StepperPolicy_Manual, res.Policy, "Unexpected policy")
	require.Equal(int64(0), res.MeasuredDelay.Seconds, "Unexpected delay")
	require.Equal(int32(0), res.MeasuredDelay.Nanos, "Unexpected delay")
	require.Equal(int64(0), res.Now, "Unexpected current time")
	require.LessOrEqual(int64(0), res.WaiterCount, "Unexpected active waiter count")
	require.Equal(stat.Epoch+1, res.Epoch, "Unexpected epoch")

	ts.doLogout(ts.randomCase(ts.adminAccountName()), cookies)
}

func (ts *StepperTestSuite) TestAdvanceOne() {
	require := ts.Require()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	stat, cookies := ts.getStatus(response.Cookies())
	cookies = ts.setManual(stat.Epoch, cookies)

	res, cookies := ts.getNow(cookies)
	require.Equal(int64(0), res.Ticks, "Time expected to be reset, was %d", res.Ticks)

	res2, cookies := ts.advance(cookies)

	require.Equal(
		int64(1), res2.Ticks-res.Ticks,
		"time expected to advance by 1, old: %d, new: %d",
		res.Ticks, res2.Ticks)

	ts.doLogout(ts.randomCase(ts.adminAccountName()), cookies)
}

func (ts *StepperTestSuite) TestAdvanceTwo() {
	require := ts.Require()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	status, cookies := ts.getStatus(response.Cookies())

	cookies = ts.setManual(status.Epoch, cookies)
	res, cookies := ts.getNow(cookies)

	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s", ts.stepperPath(), "?advance=2"), nil)
	response = ts.doHTTP(request, cookies)
	require.HTTPRSuccess(response)

	res2 := &pbc.Timestamp{}
	require.NoError(ts.getJSONBody(response, res2))

	require.Equal(
		int64(2), res2.Ticks-res.Ticks,
		"time expected to advance by 2, old: %d, new: %d",
		res.Ticks, res2.Ticks)

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *StepperTestSuite) TestAdvanceNotANumber() {
	require := ts.Require()
	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	stat, cookies := ts.getStatus(response.Cookies())
	cookies = ts.setManual(stat.Epoch, cookies)

	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s", ts.stepperPath(), "?advance=two"), nil)
	response = ts.doHTTP(request, cookies)
	require.HTTPRStatusEqual(http.StatusBadRequest, response)

	body := ts.getBody(response)
	require.Equal(
		"CloudChamber: requested rate \"two\" could not be parsed as a positive decimal number\n",
		string(body))

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *StepperTestSuite) TestAdvanceMinusOne() {
	require := ts.Require()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	stat, cookies := ts.getStatus(response.Cookies())
	cookies = ts.setManual(stat.Epoch, cookies)

	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s", ts.stepperPath(), "?advance=-1"), nil)

	response = ts.doHTTP(request, cookies)
	require.HTTPRStatusEqual(http.StatusBadRequest, response)

	body := ts.getBody(response)
	require.Equal(
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
	require.HTTPRStatusEqual(http.StatusForbidden, response)

	_ = ts.getBody(response)

	ts.doLogout(ts.aliceName(), response.Cookies())
}

func (ts *StepperTestSuite) TestSetManualBadRate() {
	require := ts.Require()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s", ts.stepperPath(), "?mode=manual=10"), nil)

	response = ts.doHTTP(request, response.Cookies())
	require.HTTPRStatusEqual(http.StatusBadRequest, response)

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *StepperTestSuite) TestAfter() {
	require := ts.Require()

	var cookies2 []*http.Cookie

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	stat, cookies := ts.getStatus(response.Cookies())
	cookies = ts.setManual(stat.Epoch, cookies)

	res, cookies := ts.getNow(cookies)

	ch := make(chan bool)

	go func(ch chan<- bool, after int64, cookies []*http.Cookie) {
		var waiter *pb.StatusResponse
		waiter, cookies2 = ts.after(after, cookies)

		require.Less(after, waiter.Now)
		ch <- true
	}(ch, res.Ticks, cookies)

	_, _ = ts.advance(cookies)

	require.True(common.CompleteWithin(ch, time.Duration(2)*time.Second))

	ts.doLogout(ts.randomCase(ts.adminAccountName()), cookies2)
}

func (ts *StepperTestSuite) TestAfterBadPastTick() {
	require := ts.Require()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	stat, cookies := ts.getStatus(response.Cookies())
	cookies = ts.setManual(stat.Epoch, cookies)

	res, cookies := ts.getNow(cookies)
	_, cookies = ts.advance(cookies)

	res2, cookies := ts.after(res.Ticks, cookies)
	require.Less(res.Ticks, res2.Now)

	ts.doLogout(ts.randomCase(ts.adminAccountName()), cookies)
}

func (ts *StepperTestSuite) TestAfterBadTick() {
	require := ts.Require()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	stat, cookies := ts.getStatus(response.Cookies())
	cookies = ts.setManual(stat.Epoch, cookies)

	request := httptest.NewRequest("GET", fmt.Sprintf("%s/now?after=-1", ts.stepperPath()), nil)

	response = ts.doHTTP(request, cookies)
	require.HTTPRStatusEqual(http.StatusBadRequest, response)

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *StepperTestSuite) TestGetNow() {
	require := ts.Require()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	res, cookies := ts.getNow(response.Cookies())
	res2, cookies := ts.getNow(cookies)

	require.Equal(res.Ticks, res2.Ticks)

	ts.doLogout(ts.randomCase(ts.adminAccountName()), cookies)
}

func TestStepperTestSuite(t *testing.T) {
	suite.Run(t, new(StepperTestSuite))
}
