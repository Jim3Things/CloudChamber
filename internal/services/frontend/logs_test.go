package frontend

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	tsc "github.com/Jim3Things/CloudChamber/internal/clients/trace_sink"
	"github.com/Jim3Things/CloudChamber/internal/common/channels"
	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters/unit_test"
	"github.com/Jim3Things/CloudChamber/pkg/protos/log"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/services"
)

const (
	logsURI = "/api/logs"
)

func testLogsPath() string { return baseURI + logsURI }

func testLogsGetPolicy(t *testing.T, cookies []*http.Cookie) (*pb.GetPolicyResponse, []*http.Cookie) {
	request := httptest.NewRequest("GET", fmt.Sprintf("%s%s", testLogsPath(), "/policy"), nil)
	response := doHTTP(request, cookies)

	assert.Equal(t, http.StatusOK, response.StatusCode)

	res := &pb.GetPolicyResponse{}
	err := getJSONBody(response, res)

	assert.Nilf(t, err, "Unexpected error, err: %v", err)

	return res, response.Cookies()
}

func testLogsGetAfter(
	t *testing.T,
	start int64,
	maxCount int64,
	cookies []*http.Cookie) (*pb.GetAfterResponse, []*http.Cookie) {
		path := fmt.Sprintf("%s?from=%d&for=%d", testLogsPath(), start, maxCount)
		request := httptest.NewRequest("GET", path, nil)
		response := doHTTP(request, cookies)

		assert.Equal(t, http.StatusOK, response.StatusCode)

		res := &pb.GetAfterResponse{}
		err := getJSONBody(response, res)

		assert.Nilf(t, err, "Unexpected error, err: %v", err)

		return res, response.Cookies()
}

func TestLogsGetPolicy(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	err := tsc.Reset()
	require.Nil(t, err)

	response := doLogin(t, randomCase(adminAccountName), adminPassword, nil)

	res, cookies := testLogsGetPolicy(t, response.Cookies())
	assert.Equal(t, int64(100), res.MaxEntriesHeld)
	assert.Equal(t, int64(-1), res.FirstId)

	doLogout(t, randomCase(adminAccountName), cookies)
}

func TestLogsGetPolicyNoSession(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	err := tsc.Reset()
	require.Nil(t, err)

	request := httptest.NewRequest("GET", fmt.Sprintf("%s%s", testLogsPath(), "/policy"), nil)
	response := doHTTP(request, nil)

	assert.Equal(t, http.StatusForbidden, response.StatusCode)
	assert.Nil(t, err)
}

func TestLogsGetAfterNoSession(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	err := tsc.Reset()
	require.Nil(t, err)

	path := fmt.Sprintf("%s?from=0&for=100", testLogsPath())
	request := httptest.NewRequest("GET", path, nil)
	response := doHTTP(request, nil)

	assert.Equal(t, http.StatusForbidden, response.StatusCode)
	assert.Nil(t, err)
}

func TestLogsGetAfterBadStart(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	err := tsc.Reset()
	require.Nil(t, err)

	response := doLogin(t, randomCase(adminAccountName), adminPassword, nil)

	path := fmt.Sprintf("%s?from=%s&for=%d", testLogsPath(), "bogus", 100)
	request := httptest.NewRequest("GET", path, nil)
	response = doHTTP(request, response.Cookies())

	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	doLogout(t, randomCase(adminAccountName), response.Cookies())
}

func TestLogsGetAfterBadCount(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	err := tsc.Reset()
	require.Nil(t, err)

	response := doLogin(t, randomCase(adminAccountName), adminPassword, nil)

	path := fmt.Sprintf("%s?from=%d&for=%s", testLogsPath(), -1, "bogus")
	request := httptest.NewRequest("GET", path, nil)
	response = doHTTP(request, response.Cookies())

	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	doLogout(t, randomCase(adminAccountName), response.Cookies())
}

func TestLogsGetAfter(t *testing.T) {
	var cookies2 []*http.Cookie

	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	entry := &log.Entry{
		Name:           "test",
		SpanID:         "0000",
		ParentID:       "1111",
		Status:         "ok",
		StackTrace:     "yyyy",
		Event:          []*log.Event {
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
	}

	err := tsc.Reset()
	require.Nil(t, err)

	response := doLogin(t, randomCase(adminAccountName), adminPassword, nil)

	ch := make(chan bool)

	go func(ch chan<- bool, cookies []*http.Cookie) {
		var res *pb.GetAfterResponse

		res, cookies2 = testLogsGetAfter(t, -1, 10, response.Cookies())

		require.Equal(t, 1, len(res.Entries))

		assert.Equal(t, int64(0), res.Entries[0].Id)

		resEntry := res.Entries[0].Entry
		assert.Equal(t, entry.Name, resEntry.Name)
		assert.Equal(t, entry.SpanID, resEntry.SpanID)
		assert.Equal(t, entry.ParentID, resEntry.ParentID)
		assert.Equal(t, entry.Infrastructure, resEntry.Infrastructure)
		assert.Equal(t, entry.Status, resEntry.Status)
		assert.Equal(t, entry.StackTrace, resEntry.StackTrace)

		require.Equal(t, len(entry.Event), len(resEntry.Event))

		resEvent := resEntry.Event[0]
		event := entry.Event[0]

		assert.Equal(t, event.Name, resEvent.Name)
		assert.Equal(t, event.StackTrace, resEvent.StackTrace)
		assert.Equal(t, event.Impacted, resEvent.Impacted)
		assert.Equal(t, event.Text, resEvent.Text)
		assert.Equal(t, event.Tick, resEvent.Tick)
		assert.Equal(t, event.Severity, resEvent.Severity)

		ch <- true
	}(ch, response.Cookies())

	assert.True(t, channels.DoNotCompleteWithin(ch, time.Duration(100) * time.Millisecond))

	err = tsc.Append(entry)
	require.Nil(t, err)

	assert.True(t, channels.CompleteWithin(ch, time.Duration(1) * time.Second))

	doLogout(t, randomCase(adminAccountName), cookies2)
}
