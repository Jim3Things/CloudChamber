package frontend

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	pbc "github.com/Jim3Things/CloudChamber/simulation/pkg/protos/common"
	pb "github.com/Jim3Things/CloudChamber/simulation/pkg/protos/services"
)

type WatchTestSuite struct {
	testSuiteCore

	cookies []*http.Cookie
}

func (ts *WatchTestSuite) SetupTest() {
	ts.testSuiteCore.SetupTest()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)
	ts.cookies = response.Cookies()
}

func (ts *WatchTestSuite) TearDownTest() {
	ts.doLogout(ts.randomCase(ts.adminAccountName()), ts.cookies)

	ts.testSuiteCore.TearDownTest()
}

func (ts *WatchTestSuite) stepperPath() string { return ts.baseURI + "/api/stepper" }
func (ts *WatchTestSuite) watchPath() string   { return ts.baseURI + "/api/watch" }

func (ts *WatchTestSuite) advance() *pbc.Timestamp {
	require := ts.Require()

	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s", ts.stepperPath(), "?advance"), nil)

	response := ts.doHTTP(request, ts.cookies)
	require.HTTPRSuccess(response)

	ts.cookies = response.Cookies()

	res := &pbc.Timestamp{}
	require.NoError(ts.getJSONBody(response, res))

	return res

}
func (ts *WatchTestSuite) getStatus() *pb.StatusResponse {
	require := ts.Require()

	request := httptest.NewRequest("GET", ts.stepperPath(), nil)

	response := ts.doHTTP(request, ts.cookies)
	require.HTTPRSuccess(response)

	ts.cookies = response.Cookies()

	res := &pb.StatusResponse{}
	require.NoError(ts.getJSONBody(response, res))

	return res
}

func (ts *WatchTestSuite) setManual(match int64) {
	require := ts.Require()

	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s", ts.stepperPath(), "?mode=manual"), nil)
	request.Header.Set("If-Match", formatAsEtag(match))

	response := ts.doHTTP(request, ts.cookies)
	require.HTTPRSuccess(response)

	ts.cookies = response.Cookies()

	require.Equal(formatAsEtag(match+1), response.Header.Get("ETag"))
}

func (ts *WatchTestSuite) doWatch(
	ch chan bool,
	tick int64,
	epoch int64,
	res *pb.WatchResponse,
	cookies []*http.Cookie) {

	require := ts.Require()

	request := httptest.NewRequest(
		"GET",
		fmt.Sprintf(
			"%s?tick=%d&epoch=%d",
			ts.watchPath(),
			tick,
			epoch),
		nil)

	response := ts.doHTTP(request, cookies)
	require.HTTPRSuccess(response)

	require.NoError(ts.getJSONBody(response, res))

	ch <- true
	close(ch)
}

func (ts *WatchTestSuite) TestEpoch() {
	res := &pb.WatchResponse{}

	require := ts.Require()

	ch := make(chan bool)
	status := ts.getStatus()

	go ts.doWatch(ch, status.Now, status.Epoch, res, ts.cookies)

	ts.setManual(-1)
	<-ch

	require.False(res.GetExpired())
	sr := res.GetStatusResponse()
	require.NotNil(sr)
	require.Less(status.Epoch, sr.Epoch)
}

func (ts *WatchTestSuite) TestAdvance() {
	res := &pb.WatchResponse{}

	require := ts.Require()

	ch := make(chan bool)
	status := ts.getStatus()

	go ts.doWatch(ch, status.Now, status.Epoch, res, ts.cookies)

	tick := ts.advance()
	<-ch

	require.False(res.GetExpired())
	sr := res.GetStatusResponse()
	require.NotNil(sr)
	require.Less(status.Now, sr.Now)
	require.LessOrEqual(tick.Ticks, sr.Now)
}

func (ts *WatchTestSuite) TestExpire() {
	res := &pb.WatchResponse{}

	require := ts.Require()

	ch := make(chan bool)
	status := ts.getStatus()

	go ts.doWatch(ch, status.Now, status.Epoch, res, ts.cookies)

	<-ch

	require.True(res.GetExpired())
}

func (ts *WatchTestSuite) TestNoSession() {
	require := ts.Require()

	request := httptest.NewRequest(
		"GET",
		fmt.Sprintf(
			"%s?tick=%d&epoch=%d",
			ts.watchPath(),
			0,
			0),
		nil)

	response := ts.doHTTP(request, nil)
	require.HTTPRStatusEqual(http.StatusForbidden, response)
}

func (ts *WatchTestSuite) TestBadValue() {
	require := ts.Require()

	request := httptest.NewRequest(
		"GET",
		fmt.Sprintf(
			"%s?tick=0&epoch=foo",
			ts.watchPath()),
		nil)

	response := ts.doHTTP(request, ts.cookies)
	require.HTTPRStatusEqual(http.StatusBadRequest, response)

	body := ts.getBody(response)

	require.Equal(
		"CloudChamber: the \"epoch\" field's value \"foo\" could not be parsed as a decimal number\n",
		string(body))
}

func TestWatchTestSuite(t *testing.T) {
	suite.Run(t, new(WatchTestSuite))
}
