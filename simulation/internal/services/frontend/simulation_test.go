package frontend

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/suite"

	pb "github.com/Jim3Things/CloudChamber/simulation/pkg/protos/admin"
)

type SimulationTestSuite struct {
	testSuiteCore
}

func (ts *SimulationTestSuite) SetupSuite() {
	ts.testSuiteCore.SetupSuite()
}

func (ts *SimulationTestSuite) simulationPath() string  { return ts.baseURI + "/api/simulation" }
func (ts *SimulationTestSuite) sessionListPath() string { return ts.simulationPath() + "/sessions" }

func (ts *SimulationTestSuite) TestSimulationSummary() {
	require := ts.Require()

	response := ts.doLogin(ts.adminAccountName(), ts.adminPassword(), nil)

	request := httptest.NewRequest("GET", ts.simulationPath(), nil)
	request.Header.Set("Content-Type", "application/json")

	response = ts.doHTTP(request, response.Cookies())

	status := &pb.SimulationStatus{}
	require.NoError(ts.getJSONBody(response, status))

	startTime, err := ptypes.Timestamp(status.FrontEndStartedAt)
	require.NoError(err)
	nanos := startTime.UnixNano()

	require.NotEqual(int64(0), nanos)
	require.Greater(time.Now().UnixNano(), nanos)

	inactivity, err := ptypes.Duration(status.InactivityTimeout)
	require.NoError(err)
	require.Equal(time.Hour, inactivity)

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *SimulationTestSuite) TestActiveSessionList() {
	require := ts.Require()

	response := ts.doLogin(ts.adminAccountName(), ts.adminPassword(), nil)

	request := httptest.NewRequest("GET", ts.sessionListPath(), nil)
	request.Header.Set("Content-Type", "application/json")

	response = ts.doHTTP(request, response.Cookies())

	status := &pb.SessionSummary{}
	require.NoError(ts.getJSONBody(response, status))

	require.Len(status.Sessions, 1)

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *SimulationTestSuite) TestListNoPrivilege() {
	require := ts.Require()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	// ... next we need a second account that we're sure exists
	_, cookies := ts.ensureAccount("Alice", ts.aliceDef, response.Cookies())

	ts.doLogout(ts.adminAccountName(), cookies)

	response = ts.doLogin("alice", ts.alicePassword(), nil)

	request := httptest.NewRequest("GET", ts.sessionListPath(), nil)
	request.Header.Set("Content-Type", "application/json")

	response = ts.doHTTP(request, response.Cookies())
	require.HTTPRStatusEqual(http.StatusForbidden, response)

	ts.doLogout("alice", response.Cookies())
}

func (ts *SimulationTestSuite) TestSessionStatus() {
	require := ts.Require()

	response := ts.doLogin(ts.adminAccountName(), ts.adminPassword(), nil)

	request := httptest.NewRequest("GET", ts.sessionListPath(), nil)
	request.Header.Set("Content-Type", "application/json")

	response = ts.doHTTP(request, response.Cookies())
	status := &pb.SessionSummary{}
	require.NoError(ts.getJSONBody(response, status))

	request = httptest.NewRequest("GET", status.Sessions[0].Uri, nil)
	request.Header.Set("Content-Type", "application/json")

	response = ts.doHTTP(request, response.Cookies())
	require.HTTPRSuccess(response)

	entry := &pb.SessionStatus{}
	require.NoError(ts.getJSONBody(response, entry))

	tmo, err := ptypes.Timestamp(entry.Timeout)
	require.NoError(err)
	require.Less(time.Now().UnixNano(), tmo.UnixNano())

	require.Equal(strings.ToLower(ts.adminAccountName()), strings.ToLower(entry.UserName))

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *SimulationTestSuite) TestSessionStatusNoPrivilege() {
	require := ts.Require()

	response := ts.doLogin(ts.adminAccountName(), ts.adminPassword(), nil)

	request := httptest.NewRequest("GET", ts.sessionListPath(), nil)
	request.Header.Set("Content-Type", "application/json")

	response = ts.doHTTP(request, response.Cookies())

	status := &pb.SessionSummary{}
	require.NoError(ts.getJSONBody(response, status))

	_, cookies := ts.ensureAccount("alice", ts.aliceDef, response.Cookies())

	ts.doLogout(ts.randomCase(ts.adminAccountName()), cookies)
	response = ts.doLogin("alice", ts.alicePassword(), nil)

	request = httptest.NewRequest("GET", status.Sessions[0].Uri, nil)
	request.Header.Set("Content-Type", "application/json")

	response = ts.doHTTP(request, response.Cookies())
	require.HTTPRStatusEqual(http.StatusForbidden, response)

	ts.doLogout("alice", response.Cookies())
}

func (ts *SimulationTestSuite) TestDeleteSession() {
	require := ts.Require()

	response := ts.doLogin(ts.adminAccountName(), ts.adminPassword(), nil)

	request := httptest.NewRequest("GET", ts.sessionListPath(), nil)
	request.Header.Set("Content-Type", "application/json")

	response = ts.doHTTP(request, response.Cookies())
	status := &pb.SessionSummary{}
	require.NoError(ts.getJSONBody(response, status))

	request = httptest.NewRequest("GET", status.Sessions[0].Uri, nil)
	request.Header.Set("Content-Type", "application/json")

	response = ts.doHTTP(request, response.Cookies())
	require.HTTPRSuccess(response)

	entry := &pb.SessionStatus{}
	require.NoError(ts.getJSONBody(response, entry))

	require.Equal(strings.ToLower(ts.adminAccountName()), strings.ToLower(entry.UserName))

	request = httptest.NewRequest("DELETE", status.Sessions[0].Uri, nil)
	response = ts.doHTTP(request, response.Cookies())
	require.HTTPRSuccess(response)

	request = httptest.NewRequest("GET", status.Sessions[0].Uri, nil)
	request.Header.Set("Content-Type", "application/json")

	response = ts.doHTTP(request, response.Cookies())
	require.HTTPRStatusEqual(http.StatusForbidden, response)
}

func (ts *SimulationTestSuite) TestDeleteNoPrivilege() {
	require := ts.Require()

	response := ts.doLogin(ts.adminAccountName(), ts.adminPassword(), nil)

	request := httptest.NewRequest("GET", ts.sessionListPath(), nil)
	request.Header.Set("Content-Type", "application/json")

	response = ts.doHTTP(request, response.Cookies())

	status := &pb.SessionSummary{}
	require.NoError(ts.getJSONBody(response, status))

	_, cookies := ts.ensureAccount("alice", ts.aliceDef, response.Cookies())

	ts.doLogout(ts.randomCase(ts.adminAccountName()), cookies)
	response = ts.doLogin("alice", ts.alicePassword(), nil)

	request = httptest.NewRequest("DELETE", status.Sessions[0].Uri, nil)
	response = ts.doHTTP(request, response.Cookies())
	require.HTTPRStatusEqual(http.StatusForbidden, response)

	ts.doLogout("alice", response.Cookies())
}

func TestSimulationTestSuite(t *testing.T) {
	suite.Run(t, new(SimulationTestSuite))
}
