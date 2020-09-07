package frontend

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	pingURI = "/api/ping"
)

func testPingPath() string { return baseURI + pingURI }

func TestPing(t *testing.T) {
	_ = utf.Open(t)
	defer utf.Close()

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

	request = httptest.NewRequest("GET", testPingPath()+"/", nil)
	response = doHTTP(request, response.Cookies())

	expiry = response.Header.Get("Expires")
	expTime, err = time.Parse(time.RFC3339, expiry)
	assert.Nil(t, err)

	assert.True(t, startTime.Before(expTime))

	doLogout(t, randomCase(adminAccountName), response.Cookies())
}

func TestPingNoSession(t *testing.T) {
	_ = utf.Open(t)
	defer utf.Close()

	request := httptest.NewRequest("GET", testPingPath(), nil)
	response := doHTTP(request, nil)

	assert.Equal(t, http.StatusForbidden, response.StatusCode)
}
