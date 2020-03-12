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
)

var (
	baseURI     string
	initialized bool
)

// commonSetup bears more than a passing resemblance to the primary package
// entry point StartService()
//
func commonSetup() error {

	if initialized {
		log.Fatalf("Error initializing service for second or subsequent time")
	}

	if err := initService(); err != nil {
		log.Fatalf("Error initializing service: %v", err)
	}

	baseURI = fmt.Sprintf("http://localhost:%d", server.port)

	initialized = true

	return nil
}

func TestMain(m *testing.M) {

	commonSetup()

	os.Exit(m.Run())
}

func TestUsersList(t *testing.T) {

	route := "/api/users"

	request := httptest.NewRequest("GET", fmt.Sprintf("%s%s", baseURI, route), nil)

	w := httptest.NewRecorder()

	server.handler.ServeHTTP(w, request)

	response := w.Result()

	body, err := ioutil.ReadAll(response.Body)
	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", route, err)

	fmt.Println(response.StatusCode)
	fmt.Println(response.Header.Get("Content-Type"))
	fmt.Println(string(body))

	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", err)
	assert.Equal(t, "Users (List)", string(body), "Handler returned unexpected response body: %v", string(body))
}

func TestUsersCreate(t *testing.T) {

}

func TestUsersRead(t *testing.T) {

	const route = "/api/users/Alice"

	request := httptest.NewRequest("GET", fmt.Sprintf("%s%s", baseURI, route), nil)

	w := httptest.NewRecorder()

	server.handler.ServeHTTP(w, request)

	response := w.Result()

	body, err := ioutil.ReadAll(response.Body)
	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", route, err)

	fmt.Println(response.StatusCode)
	fmt.Println(response.Header.Get("Content-Type"))
	fmt.Println(string(body))

	// At present, all base handlers effectively echo the supplied username so all
	// we need to verify is that we get a successful return of the supplied username.
	//
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", err)
	assert.Equal(t, "User: Alice", string(body), "Handler returned unexpected response body: %v", string(body))
}

func TestUsersOperation(t *testing.T) {

	const route = "/api/users/Alice"

	// First verify a bunch of failuer cases. Specifically,
	// - that a trailing / char fails
	// - that a naked op fails
	// - that an invalid op fails
	// - that all of the allowed ops succeed

	// Case 1, check that trailing / fails
	//
	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s/", baseURI, route), nil)

	w := httptest.NewRecorder()

	server.handler.ServeHTTP(w, request)

	response := w.Result()

	body, err := ioutil.ReadAll(response.Body)
	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", route, err)

	fmt.Println(response.StatusCode)
	fmt.Println(response.Header.Get("Content-Type"))
	fmt.Println(string(body))

	assert.Equal(t, http.StatusBadRequest, response.StatusCode, "Handler returned unexpected error: %v", err)
	assert.Equal(t, "InvalidOp\n", string(body), "Handler returned unexpected response body: %v", string(body))

	// Case 2, check that naked op fails
	//
	request = httptest.NewRequest("PUT", fmt.Sprintf("%s%s/?op", baseURI, route), nil)

	w = httptest.NewRecorder()

	server.handler.ServeHTTP(w, request)

	response = w.Result()

	body, err = ioutil.ReadAll(response.Body)
	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", route, err)

	fmt.Println(response.StatusCode)
	fmt.Println(response.Header.Get("Content-Type"))
	fmt.Println(string(body))

	// At present, all base handlers effectively echo the supplied username so all
	// we need to verify is that we get a successful return of the supplied username.
	//
	assert.Equal(t, http.StatusBadRequest, response.StatusCode, "Handler returned unexpected error: %v", err)
	assert.Equal(t, "InvalidOp\n", string(body), "Handler returned unexpected response body: %v", string(body))

	// Case 3, check that an invalid op fails
	//
	request = httptest.NewRequest("PUT", fmt.Sprintf("%s%s/?op=testInvalid", baseURI, route), nil)

	w = httptest.NewRecorder()

	server.handler.ServeHTTP(w, request)

	response = w.Result()

	body, err = ioutil.ReadAll(response.Body)
	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", route, err)

	fmt.Println(response.StatusCode)
	fmt.Println(response.Header.Get("Content-Type"))
	fmt.Println(string(body))

	// At present, all base handlers effectively echo the supplied username so all
	// we need to verify is that we get a successful return of the supplied username.
	//
	assert.Equal(t, http.StatusBadRequest, response.StatusCode, "Handler returned unexpected error: %v", err)
	assert.Equal(t, "InvalidOp\n", string(body), "Handler returned unexpected response body: %v", string(body))

	// Case 4, check that each of the valid ops succeed
	//
	// 4a, enable
	//
	request = httptest.NewRequest("PUT", fmt.Sprintf("%s%s/?op=enable", baseURI, route), nil)

	w = httptest.NewRecorder()

	server.handler.ServeHTTP(w, request)

	response = w.Result()

	body, err = ioutil.ReadAll(response.Body)
	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", route, err)

	fmt.Println(response.StatusCode)
	fmt.Println(response.Header.Get("Content-Type"))
	fmt.Println(string(body))

	// At present, all base handlers effectively echo the supplied username so all
	// we need to verify is that we get a successful return of the supplied username.
	//
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", err)
	assert.Equal(t, "User: Alice op: enable", string(body), "Handler returned unexpected response body: %v", string(body))

	// 4b, disable
	//
	request = httptest.NewRequest("PUT", fmt.Sprintf("%s%s/?op=disable", baseURI, route), nil)

	w = httptest.NewRecorder()

	server.handler.ServeHTTP(w, request)

	response = w.Result()

	body, err = ioutil.ReadAll(response.Body)
	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", route, err)

	fmt.Println(response.StatusCode)
	fmt.Println(response.Header.Get("Content-Type"))
	fmt.Println(string(body))

	// At present, all base handlers effectively echo the supplied username so all
	// we need to verify is that we get a successful return of the supplied username.
	//
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", err)
	assert.Equal(t, "User: Alice op: disable", string(body), "Handler returned unexpected response body: %v", string(body))

	// 4c, login
	//
	request = httptest.NewRequest("PUT", fmt.Sprintf("%s%s/?op=login", baseURI, route), nil)

	w = httptest.NewRecorder()

	server.handler.ServeHTTP(w, request)

	response = w.Result()

	body, err = ioutil.ReadAll(response.Body)
	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", route, err)

	fmt.Println(response.StatusCode)
	fmt.Println(response.Header.Get("Content-Type"))
	fmt.Println(string(body))

	// At present, all base handlers effectively echo the supplied username so all
	// we need to verify is that we get a successful return of the supplied username.
	//
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", err)
	assert.Equal(t, "User: Alice op: login", string(body), "Handler returned unexpected response body: %v", string(body))

	// 4d, logout
	//
	request = httptest.NewRequest("PUT", fmt.Sprintf("%s%s/?op=logout", baseURI, route), nil)

	w = httptest.NewRecorder()

	server.handler.ServeHTTP(w, request)

	response = w.Result()

	body, err = ioutil.ReadAll(response.Body)
	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", route, err)

	fmt.Println(response.StatusCode)
	fmt.Println(response.Header.Get("Content-Type"))
	fmt.Println(string(body))

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
