package frontend

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"

    "github.com/Jim3Things/CloudChamber/internal/tracing/exporters/unit_test"
)

const (
    pingURI    = "/api/ping"
)

func testPingPath() string { return baseURI + pingURI }

func TestPing(t *testing.T) {
    unit_test.SetTesting(t)
    defer unit_test.SetTesting(nil)

    response := doLogin(t, randomCase(adminAccountName), adminPassword, nil)

    startTime := time.Now()
    time.Sleep(time.Duration(1) * time.Second)

    request := httptest.NewRequest("GET", testPingPath(), nil)
    response = doHTTP(request, response.Cookies())

    expiry := response.Header.Get("Expires")
    expTime, err := time.Parse(time.RFC3339, expiry)
    assert.Nil(t, err)

    assert.True(t, startTime.Before(expTime))

    startTime = expTime
    time.Sleep(time.Duration(1) * time.Second)

    request = httptest.NewRequest("GET", testPingPath() +"/", nil)
    response = doHTTP(request, response.Cookies())

    expiry = response.Header.Get("Expires")
    expTime, err = time.Parse(time.RFC3339, expiry)
    assert.Nil(t, err)

    assert.True(t, startTime.Before(expTime))

    doLogout(t, randomCase(adminAccountName), response.Cookies())
}

func TestPingNoSession(t *testing.T) {
    unit_test.SetTesting(t)
    defer unit_test.SetTesting(nil)

    request := httptest.NewRequest("GET", testPingPath(), nil)
    response := doHTTP(request, nil)

    assert.Equal(t, http.StatusForbidden, response.StatusCode)
}