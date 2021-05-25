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
	require := ts.Require()

	request := httptest.NewRequest("GET", fmt.Sprintf("%s%s", ts.logsPath(), "/policy"), nil)
	response := ts.doHTTP(request, cookies)

	require.HTTPRSuccess(response)

	res := &pb.GetPolicyResponse{}
	require.NoError(ts.getJSONBody(response, res))

	return res, response.Cookies()
}

func (ts *LogTestSuite) getAfter(
	start int64,
	maxCount int64,
	cookies []*http.Cookie) (*pb.GetAfterResponse, []*http.Cookie) {
	require := ts.Require()

	path := fmt.Sprintf("%s?from=%d&for=%d", ts.logsPath(), start, maxCount)
	request := httptest.NewRequest("GET", path, nil)
	response := ts.doHTTP(request, cookies)

	require.HTTPRSuccess(response)

	res := &pb.GetAfterResponse{}
	require.NoError(ts.getJSONBody(response, res))

	return res, response.Cookies()
}

func (ts *LogTestSuite) TestGetPolicy() {
	require := ts.Require()

	require.NoError(tsc.Reset(context.Background()))

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	res, cookies := ts.getPolicy(response.Cookies())
	require.Equal(int64(200), res.MaxEntriesHeld)
	require.Equal(int64(-1), res.FirstId)

	ts.doLogout(ts.randomCase(ts.adminAccountName()), cookies)
}

func (ts *LogTestSuite) TestGetPolicyNoSession() {
	require := ts.Require()

	require.NoError(tsc.Reset(context.Background()))

	request := httptest.NewRequest("GET", fmt.Sprintf("%s%s", ts.logsPath(), "/policy"), nil)
	response := ts.doHTTP(request, nil)

	require.HTTPRStatusEqual(http.StatusForbidden, response)
}

func (ts *LogTestSuite) TestGetAfterNoSession() {
	require := ts.Require()

	require.NoError(tsc.Reset(context.Background()))

	path := fmt.Sprintf("%s?from=0&for=100", ts.logsPath())
	request := httptest.NewRequest("GET", path, nil)
	response := ts.doHTTP(request, nil)

	require.HTTPRStatusEqual(http.StatusForbidden, response)
}

func (ts *LogTestSuite) TestGetAfterBadStart() {
	require := ts.Require()

	require.NoError(tsc.Reset(context.Background()))

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	path := fmt.Sprintf("%s?from=%s&for=%d", ts.logsPath(), "bogus", 100)
	request := httptest.NewRequest("GET", path, nil)
	response = ts.doHTTP(request, response.Cookies())

	require.HTTPRStatusEqual(http.StatusBadRequest, response)

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *LogTestSuite) TestGetAfterBadCount() {
	require := ts.Require()

	require.NoError(tsc.Reset(context.Background()))

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	path := fmt.Sprintf("%s?from=%d&for=%s", ts.logsPath(), -1, "bogus")
	request := httptest.NewRequest("GET", path, nil)
	response = ts.doHTTP(request, response.Cookies())

	require.HTTPRStatusEqual(http.StatusBadRequest, response)

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *LogTestSuite) TestGetAfter() {
	require := ts.Require()

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
			},
		},
		Infrastructure: false,
		Reason:         "My Reason",
	}

	require.NoError(tsc.Reset(context.Background()))

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	ch := make(chan bool)

	go func(ch chan<- bool, cookies []*http.Cookie) {
		var res *pb.GetAfterResponse

		res, cookies2 = ts.getAfter(-1, 10, response.Cookies())

		require.Equal(1, len(res.Entries))

		require.Equal(int64(0), res.Entries[0].Id)

		resEntry := res.Entries[0].Entry
		require.Equal(entry.Name, resEntry.Name)
		require.Equal(entry.SpanID, resEntry.SpanID)
		require.Equal(entry.ParentID, resEntry.ParentID)
		require.Equal(entry.TraceID, resEntry.TraceID)
		require.Equal(entry.Infrastructure, resEntry.Infrastructure)
		require.Equal(entry.Status, resEntry.Status)
		require.Equal(entry.StackTrace, resEntry.StackTrace)
		require.Equal(entry.Reason, resEntry.Reason)
		require.Equal(entry.Impacted, resEntry.Impacted)

		require.Equal(len(entry.Event), len(resEntry.Event))

		resEvent := resEntry.Event[0]
		event := entry.Event[0]

		require.Equal(event.Name, resEvent.Name)
		require.Equal(event.StackTrace, resEvent.StackTrace)
		require.Equal(event.Text, resEvent.Text)
		require.Equal(event.Tick, resEvent.Tick)
		require.Equal(event.Severity, resEvent.Severity)

		ch <- true
	}(ch, response.Cookies())

	require.True(common.DoNotCompleteWithin(ch, time.Duration(100)*time.Millisecond))

	require.NoError(tsc.Append(context.Background(), entry))

	require.True(common.CompleteWithin(ch, time.Second))

	ts.doLogout(ts.randomCase(ts.adminAccountName()), cookies2)
}

func TestLogTestSuite(t *testing.T) {
	suite.Run(t, new(LogTestSuite))
}
