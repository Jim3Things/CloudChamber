package frontend

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type PingTestSuite struct {
	testSuiteCore
}

func (ts *PingTestSuite) pingPath() string { return ts.baseURI + "/api/ping" }

func (ts *PingTestSuite) TestSimple() {
	assert := ts.Assert()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	startTime := time.Now()
	time.Sleep(time.Duration(1) * time.Second)

	request := httptest.NewRequest("GET", ts.pingPath(), nil)
	response = ts.doHTTP(request, response.Cookies())

	expiry := response.Header.Get("Expires")
	expTime, err := time.Parse(time.RFC3339, expiry)
	assert.NoError(err)

	assert.True(startTime.Before(expTime))

	startTime = expTime
	time.Sleep(time.Duration(1) * time.Second)

	request = httptest.NewRequest("GET", ts.pingPath()+"/", nil)
	response = ts.doHTTP(request, response.Cookies())

	expiry = response.Header.Get("Expires")
	expTime, err = time.Parse(time.RFC3339, expiry)
	assert.NoError(err)

	assert.True(startTime.Before(expTime))

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *PingTestSuite) TestNoSession() {
	assert := ts.Assert()

	request := httptest.NewRequest("GET", ts.pingPath(), nil)
	response := ts.doHTTP(request, nil)

	assert.Equal(http.StatusForbidden, response.StatusCode)
}

func TestPingTestSuite(t *testing.T) {
	suite.Run(t, new(PingTestSuite))
}
