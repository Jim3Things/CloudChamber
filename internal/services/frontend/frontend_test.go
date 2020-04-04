// Unit tests for the web service frontend.
//
// Borrows heavily from the gorilla mux test package.
//
package frontend

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Jim3Things/CloudChamber/internal/config"
	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters"
	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters/unit_test"
	"github.com/Jim3Things/CloudChamber/internal/tracing/setup"
)

var (
	baseURI     string
	initialized bool
)

// commonSetup bears more than a passing resemblance to the primary package
// entry point StartService()
//
func commonSetup() {

	if initialized {
		log.Fatalf("Error initializing service for second or subsequent time")
	}

	setup.Init(exporters.UnitTest)
	if err := initService(&config.GlobalConfig{
		Controller: config.ControllerType{},
		Inventory:  config.InventoryType{},
		SimSupport: config.SimSupportType{},
		WebServer:  config.WebServerType{
			RootFilePath:  "C:\\CloudChamber",
			SystemAccount: "Admin",
			FE:            config.Endpoint{
				Hostname: "localhost",
				Port:     8080,
			},
			BE:            config.Endpoint{},
		},
	}); err != nil {
		log.Fatalf("Error initializing service: %v", err)
	}

	baseURI = fmt.Sprintf("http://localhost:%d", server.port)

	initialized = true
}

// Helper function to execute the http request/response sequence
func doHttp(req *http.Request) *http.Response {
	w := httptest.NewRecorder()

	server.handler.ServeHTTP(w, req)

	return w.Result()
}

func TestMain(m *testing.M) {

	commonSetup()

	os.Exit(m.Run())
}

func TestUsersList(t *testing.T) {
	route := "/api/users"

	unit_test.SetTesting(t)

	request := httptest.NewRequest("GET", fmt.Sprintf("%s%s", baseURI, route), nil)
	response := doHttp(request)

	body, err := ioutil.ReadAll(response.Body)
	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", route, err)

	t.Logf("[%s]: SC=%v, Content-Type='%v'\n", route, response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", err)
	assert.Equal(t, "Users (List)\nhttp://localhost:8080/api/users/Admin\n", string(body), "Handler returned unexpected response body: %v", string(body))
}

func TestUsersCreate(t *testing.T) {

}

func TestUsersRead(t *testing.T) {

	const route = "/api/users/Alice"

	unit_test.SetTesting(t)

	request := httptest.NewRequest("GET", fmt.Sprintf("%s%s", baseURI, route), nil)
	response := doHttp(request)

	body, err := ioutil.ReadAll(response.Body)
	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", route, err)

	t.Logf("[%s]: SC=%v, Content-Type='%v'\n", route, response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	// At present, all base handlers effectively echo the supplied username so all
	// we need to verify is that we get a successful return of the supplied username.
	//
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", err)
	assert.Equal(t, "User: Alice", string(body), "Handler returned unexpected response body: %v", string(body))
}

func TestUsersOperation(t *testing.T) {

	const route = "/api/users/Alice"

	unit_test.SetTesting(t)

	// First verify a bunch of failure cases. Specifically,
	// - that a naked op fails
	// - that an invalid op fails
	// Second, verify correct behavior
	// - that all of the allowed ops succeed

	// Case 1, check that naked op fails
	//
	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op", baseURI, route), nil)
	response := doHttp(request)

	body, err := ioutil.ReadAll(response.Body)
	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", route, err)

	t.Logf("[?op]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	// At present, all base handlers effectively echo the supplied username so all
	// we need to verify is that we get a successful return of the supplied username.
	//
	assert.Equal(t, http.StatusBadRequest, response.StatusCode, "Handler returned unexpected error: %v", err)
	assert.Equal(t, "Invalid user operation requested (?op=)\n", string(body), "Handler returned unexpected response body: %v", string(body))

	// Case 2, check that an invalid op fails
	//
	request = httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=testInvalid", baseURI, route), nil)
	response = doHttp(request)

	body, err = ioutil.ReadAll(response.Body)
	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", route, err)

	t.Logf("[?op=testInvalid]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	// At present, all base handlers effectively echo the supplied username so all
	// we need to verify is that we get a successful return of the supplied username.
	//
	assert.Equal(t, http.StatusBadRequest, response.StatusCode, "Handler returned unexpected error: %v", err)
	assert.Equal(t, "Invalid user operation requested (?op=testInvalid)\n", string(body), "Handler returned unexpected response body: %v", string(body))

	// Case 3, check that each of the valid ops succeed
	//
	// 3a, enable
	//
	request = httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=enable", baseURI, route), nil)
	response = doHttp(request)

	body, err = ioutil.ReadAll(response.Body)
	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", route, err)

	t.Logf("[?op=enable]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	// At present, all base handlers effectively echo the supplied username so all
	// we need to verify is that we get a successful return of the supplied username.
	//
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", err)
	assert.Equal(t, "User: Alice op: enable", string(body), "Handler returned unexpected response body: %v", string(body))

	// 3b, disable
	//
	request = httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=disable", baseURI, route), nil)
	response = doHttp(request)

	body, err = ioutil.ReadAll(response.Body)
	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", route, err)

	t.Logf("[?op=disable]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	// At present, all base handlers effectively echo the supplied username so all
	// we need to verify is that we get a successful return of the supplied username.
	//
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", err)
	assert.Equal(t, "User: Alice op: disable", string(body), "Handler returned unexpected response body: %v", string(body))

	// 3c, login
	//
	request = httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=login", baseURI, route), nil)
	response = doHttp(request)

	body, err = ioutil.ReadAll(response.Body)
	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", route, err)

	t.Logf("[?op=login]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", err)
	assert.Equal(t, 0, len(body), "Handler returned unexpected response body: %v", string(body))

	// 3d, logout
	//
	request = httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=logout", baseURI, route), nil)
	response = doHttp(request)

	body, err = ioutil.ReadAll(response.Body)
	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", route, err)

	t.Logf("[?op=logout]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	// At present, all base handlers effectively echo the supplied username so all
	// we need to verify is that we get a successful return of the supplied username.
	//
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", err)
	assert.Equal(t, "User: Alice op: logout", string(body), "Handler returned unexpected response body: %v", string(body))
}

func TestUsersUpdate(t *testing.T) {

}

func TestUsersDelete(t *testing.T) {

}

func TestWorkloadsFetch(t *testing.T) {

}

func TestWorkloadsCreate(t *testing.T) {

}

func TestWorkloadsRead(t *testing.T) {

}

func TestWorkloadsUpdate(t *testing.T) {

}

func TestWorkloadsDelete(t *testing.T) {

}
