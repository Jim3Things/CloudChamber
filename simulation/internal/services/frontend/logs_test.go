package frontend

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	tsc "github.com/Jim3Things/CloudChamber/simulation/internal/clients/trace_sink"
	"github.com/Jim3Things/CloudChamber/simulation/internal/common"
	"github.com/Jim3Things/CloudChamber/simulation/pkg/protos/log"
	pb "github.com/Jim3Things/CloudChamber/simulation/pkg/protos/services"
)

type LogTestSuite struct {
	testSuiteCore
}

func (ts *LogTestSuite) logsPath() string { return ts.baseURI + "/api/logs" }

func (ts *LogTestSuite) getPolicy(cookies []*http.Cookie) (*pb.GetPolicyResponse, []*http.Cookie) {
	assert := ts.Assert()

	request := httptest.NewRequest("GET", fmt.Sprintf("%s%s", ts.logsPath(), "/policy"), nil)
	response := ts.doHTTP(request, cookies)

	assert.Equal(http.StatusOK, response.StatusCode)

	res := &pb.GetPolicyResponse{}
	err := ts.getJSONBody(response, res)

	assert.NoError(err, "Unexpected error, err: %v", err)

	return res, response.Cookies()
}

func (ts *LogTestSuite) getAfter(
	start int64,
	maxCount int64,
	cookies []*http.Cookie) (*pb.GetAfterResponse, []*http.Cookie) {
	assert := ts.Assert()

	path := fmt.Sprintf("%s?from=%d&for=%d", ts.logsPath(), start, maxCount)
	request := httptest.NewRequest("GET", path, nil)
	response := ts.doHTTP(request, cookies)

	assert.Equal(http.StatusOK, response.StatusCode)

	res := &pb.GetAfterResponse{}
	err := ts.getJSONBody(response, res)

	assert.NoError(err, "Unexpected error, err: %v", err)

	return res, response.Cookies()
}

func (ts *LogTestSuite) TestGetPolicy() {
	require := ts.Require()
	assert := ts.Assert()

	err := tsc.Reset(context.Background())
	require.NoError(err)

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	res, cookies := ts.getPolicy(response.Cookies())
	assert.Equal(int64(200), res.MaxEntriesHeld)
	assert.Equal(int64(-1), res.FirstId)

	ts.doLogout(ts.randomCase(ts.adminAccountName()), cookies)
}

func (ts *LogTestSuite) TestGetPolicyNoSession() {
	require := ts.Require()
	assert := ts.Assert()

	err := tsc.Reset(context.Background())
	require.NoError(err)

	request := httptest.NewRequest("GET", fmt.Sprintf("%s%s", ts.logsPath(), "/policy"), nil)
	response := ts.doHTTP(request, nil)

	assert.Equal(http.StatusForbidden, response.StatusCode)
	assert.NoError(err)
}

func (ts *LogTestSuite) TestGetAfterNoSession() {
	require := ts.Require()
	assert := ts.Assert()

	err := tsc.Reset(context.Background())
	require.NoError(err)

	path := fmt.Sprintf("%s?from=0&for=100", ts.logsPath())
	request := httptest.NewRequest("GET", path, nil)
	response := ts.doHTTP(request, nil)

	assert.Equal(http.StatusForbidden, response.StatusCode)
	assert.NoError(err)
}

func (ts *LogTestSuite) TestGetAfterBadStart() {
	require := ts.Require()
	assert := ts.Assert()

	err := tsc.Reset(context.Background())
	require.NoError(err)

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	path := fmt.Sprintf("%s?from=%s&for=%d", ts.logsPath(), "bogus", 100)
	request := httptest.NewRequest("GET", path, nil)
	response = ts.doHTTP(request, response.Cookies())

	assert.Equal(http.StatusBadRequest, response.StatusCode)

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *LogTestSuite) TestGetAfterBadCount() {
	require := ts.Require()
	assert := ts.Assert()

	err := tsc.Reset(context.Background())
	require.NoError(err)

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	path := fmt.Sprintf("%s?from=%d&for=%s", ts.logsPath(), -1, "bogus")
	request := httptest.NewRequest("GET", path, nil)
	response = ts.doHTTP(request, response.Cookies())

	assert.Equal(http.StatusBadRequest, response.StatusCode)

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *LogTestSuite) TestGetAfter() {
	require := ts.Require()
	assert := ts.Assert()

	var cookies2 []*http.Cookie

	entry := &log.Entry{
		Name:       "test",
		SpanID:     "1102030405060708",
		ParentID:   "1111111111111111",
		TraceID:    "11020304050607081102030405060708",
		Status:     "ok",
		StackTrace: "yyyy",
		Event: []*log.Event{
			{
				Tick:       0,
				Severity:   1,
				Name:       "testEvent",
				Text:       "xyzzy",
				StackTrace: "zzzz",
				Impacted:   nil,
			},
		},
		Infrastructure: false,
		Reason:         "My Reason",
	}

	err := tsc.Reset(context.Background())
	require.NoError(err)

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	ch := make(chan bool)

	go func(ch chan<- bool, cookies []*http.Cookie) {
		var res *pb.GetAfterResponse

		res, cookies2 = ts.getAfter(-1, 10, response.Cookies())

		require.Equal(1, len(res.Entries))

		assert.Equal(int64(0), res.Entries[0].Id)

		resEntry := res.Entries[0].Entry
		assert.Equal(entry.Name, resEntry.Name)
		assert.Equal(entry.SpanID, resEntry.SpanID)
		assert.Equal(entry.ParentID, resEntry.ParentID)
		assert.Equal(entry.TraceID, resEntry.TraceID)
		assert.Equal(entry.Infrastructure, resEntry.Infrastructure)
		assert.Equal(entry.Status, resEntry.Status)
		assert.Equal(entry.StackTrace, resEntry.StackTrace)
		assert.Equal(entry.Reason, resEntry.Reason)

		require.Equal(len(entry.Event), len(resEntry.Event))

		resEvent := resEntry.Event[0]
		event := entry.Event[0]

		assert.Equal(event.Name, resEvent.Name)
		assert.Equal(event.StackTrace, resEvent.StackTrace)
		assert.Equal(event.Impacted, resEvent.Impacted)
		assert.Equal(event.Text, resEvent.Text)
		assert.Equal(event.Tick, resEvent.Tick)
		assert.Equal(event.Severity, resEvent.Severity)

		ch <- true
	}(ch, response.Cookies())

	assert.True(common.DoNotCompleteWithin(ch, time.Duration(100)*time.Millisecond))

	err = tsc.Append(context.Background(), entry)
	require.NoError(err)

	assert.True(common.CompleteWithin(ch, time.Duration(1)*time.Second))

	ts.doLogout(ts.randomCase(ts.adminAccountName()), cookies2)
}

func TestLogTestSuite(t *testing.T) {
	suite.Run(t, new(LogTestSuite))
}
