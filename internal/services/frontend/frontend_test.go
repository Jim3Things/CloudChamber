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
	"strings"
	"testing"

	"github.com/golang/protobuf/jsonpb"
	"github.com/stretchr/testify/assert"

	"github.com/Jim3Things/CloudChamber/internal/config"
	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters"
	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters/unit_test"
	"github.com/Jim3Things/CloudChamber/internal/tracing/setup"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/admin"
)

var (
	baseURI     string
	initialized bool
)

const (
	userURI = "/api/users/"
	alice = "/api/users/Alice"
	admin = "/api/users/Admin"
	bob = "/api/users/Bob"
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

func doLogin(t *testing.T, user string, cookies []*http.Cookie) *http.Response{
	path := fmt.Sprintf("%s%s%s?op=login", baseURI, userURI, user)
	t.Logf("[login as %q (%q)]", user, path)

	request := httptest.NewRequest("PUT", path, nil)
	for _, c := range cookies {
		request.AddCookie(c)
	}
	response := doHttp(request)

	_, err := ioutil.ReadAll(response.Body)
	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %q: %v", path, err)
	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", err)

	return response
}

func doLogout(t *testing.T, user string, cookies []*http.Cookie) *http.Response {
	path := fmt.Sprintf("%s%s%s?op=logout", baseURI, userURI, user)
	t.Logf("[logout from %q (%q)]", user, path)

	request := httptest.NewRequest("PUT", path, nil)
	for _, c := range cookies {
		request.AddCookie(c)
	}
	response := doHttp(request)

	_, err := ioutil.ReadAll(response.Body)
	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", alice, err)
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", err)

	return response
}

func TestMain(m *testing.M) {

	commonSetup()

	os.Exit(m.Run())
}

func TestLoginSessionSimple(t *testing.T)  {
	unit_test.SetTesting(t)

	// login for the first time, should succeed
	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=login", baseURI, admin), nil)
	response := doHttp(request)

	body, err := ioutil.ReadAll(response.Body)
	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", admin, err)

	t.Logf("[?op=login]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", err)

	// ... and logout, which should succeed
	//
	request = httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=logout", baseURI, admin), nil)
	for _, c := range response.Cookies() {
		request.AddCookie(c)
	}
	response = doHttp(request)

	body, err = ioutil.ReadAll(response.Body)
	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", admin, err)

	t.Logf("[?op=logout]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", err)
}

func TestUsersList(t *testing.T) {
	unit_test.SetTesting(t)

	response := doLogin(t, "Admin", nil)

	request := httptest.NewRequest("GET", fmt.Sprintf("%s%s", baseURI, userURI), nil)
	for _, c := range response.Cookies() {
		request.AddCookie(c)
	}
	response = doHttp(request)

	body, err := ioutil.ReadAll(response.Body)
	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", userURI, err)

	t.Logf("[%s]: SC=%v, Content-Type='%v'\n", userURI, response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", err)
	assert.Equal(t, "Users (List)\nhttp://localhost:8080/api/users/Admin\n", string(body), "Handler returned unexpected response body: %v", string(body))
	doLogout(t, "admin", response.Cookies())
}

func TestUsersCreate(t *testing.T) {
	unit_test.SetTesting(t)

	response := doLogin(t, "Admin", nil)

	request := httptest.NewRequest(
		"POST",
		fmt.Sprintf("%s%s%s", baseURI, userURI, "Alice"),
		strings.NewReader("{\"enabled\":true,\"accountManager\":false, \"password\":\"test\"}"))
	for _, c := range response.Cookies() {
		request.AddCookie(c)
	}

	request.Header.Set("Content-Type", "application/json")
	response = doHttp(request)
	body, err := ioutil.ReadAll(response.Body)
	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", userURI, err)

	t.Logf("[%s]: SC=%v, Content-Type='%v'\n", userURI, response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))


	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", err)
	assert.Equal(t, "User \"Alice\" created.  enabled: true, can manage accounts: false", string(body), "Handler returned unexpected response body: %v", string(body))
	doLogout(t, "admin", response.Cookies())
}

func TestUsersRead(t *testing.T) {
	unit_test.SetTesting(t)

	response := doLogin(t, "Admin", nil)

	request := httptest.NewRequest("GET", fmt.Sprintf("%s%s%s", baseURI, userURI, "admin"), nil)
	for _, c := range response.Cookies() {
		request.AddCookie(c)
	}
	response = doHttp(request)

	user := &pb.UserPublic{}
	err := jsonpb.Unmarshal(response.Body, user)
	assert.Nilf(t, err, "Failed to convert body to valid json.  err: %v", err)

	assert.Equal(t, "application/json", strings.ToLower(response.Header.Get("Content-Type")))
	assert.Equal(t, true, user.Enabled)
	assert.Equal(t, true, user.AccountManager)

	// At present, all base handlers effectively echo the supplied username so all
	// we need to verify is that we get a successful return of the supplied username.
	//
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", err)
	doLogout(t, "admin", response.Cookies())
}

func TestUsersOperationIllegal(t *testing.T) {
	unit_test.SetTesting(t)

	// Verify a bunch of failure cases. Specifically,
	// - that a naked op fails
	// - that an invalid op fails

	// Case 1, check that naked op fails
	//
	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op", baseURI, alice), nil)
	response := doHttp(request)

	body, err := ioutil.ReadAll(response.Body)
	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", alice, err)

	t.Logf("[?op]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	// At present, all base handlers effectively echo the supplied username so all
	// we need to verify is that we get a successful return of the supplied username.
	//
	assert.Equal(t, http.StatusBadRequest, response.StatusCode, "Handler returned unexpected error: %v", err)
	assert.Equal(t, "invalid user operation requested (?op=)\n", string(body), "Handler returned unexpected response body: %v", string(body))

	// Case 2, check that an invalid op fails
	//
	request = httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=testInvalid", baseURI, alice), nil)
	response = doHttp(request)

	body, err = ioutil.ReadAll(response.Body)
	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", alice, err)

	t.Logf("[?op=testInvalid]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	// At present, all base handlers effectively echo the supplied username so all
	// we need to verify is that we get a successful return of the supplied username.
	//
	assert.Equal(t, http.StatusBadRequest, response.StatusCode, "Handler returned unexpected error: %v", err)
	assert.Equal(t, "invalid user operation requested (?op=testInvalid)\n", string(body), "Handler returned unexpected response body: %v", string(body))
}

func TestUserOperationsDisable(t *testing.T) {
	response := doLogin(t, "Admin", nil)

	// 3b, disable
	//
	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=disable", baseURI, alice), nil)
	for _, c := range response.Cookies() {
		request.AddCookie(c)
	}
	response = doHttp(request)

	body, err := ioutil.ReadAll(response.Body)
	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", alice, err)

	t.Logf("[?op=disable]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	// At present, all base handlers effectively echo the supplied username so all
	// we need to verify is that we get a successful return of the supplied username.
	//
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", err)
	assert.Equal(t, "User \"alice\" disabled\n", string(body), "Handler returned unexpected response body: %v", string(body))
	doLogout(t, "admin", response.Cookies())
}

func TestUsersOperationEnable(t *testing.T) {
	response := doLogin(t, "Admin", nil)

	// Case 3, check that each of the valid ops succeed
	//
	// 3a, enable
	//
	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=enable", baseURI, alice), nil)
	for _, c := range response.Cookies() {
		request.AddCookie(c)
	}
	response = doHttp(request)

	body, err := ioutil.ReadAll(response.Body)
	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", alice, err)

	t.Logf("[?op=enable]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	// At present, all base handlers effectively echo the supplied username so all
	// we need to verify is that we get a successful return of the supplied username.
	//
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", err)
	assert.Equal(t, "User \"alice\" enabled\n", string(body), "Handler returned unexpected response body: %v", string(body))
	doLogout(t, "admin", response.Cookies())
}

func TestLoginSessionRepeat(t *testing.T)  {
	unit_test.SetTesting(t)

	// login for the first time, should succeed
	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=login", baseURI, alice), nil)
	response := doHttp(request)

	body, err := ioutil.ReadAll(response.Body)
	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", alice, err)

	t.Logf("[?op=login]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", err)

	// ... and logout, which should succeed
	//
	request = httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=logout", baseURI, alice), nil)
	for _, c := range response.Cookies() {
		request.AddCookie(c)
	}
	response = doHttp(request)

	body, err = ioutil.ReadAll(response.Body)
	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", alice, err)

	t.Logf("[?op=logout]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", err)

	// login for the second iteration, should succeed
	request = httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=login", baseURI, alice), nil)
	response = doHttp(request)

	body, err = ioutil.ReadAll(response.Body)
	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", alice, err)

	t.Logf("[?op=login]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", err)

	// ... and logout, which should succeed
	//
	request = httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=logout", baseURI, alice), nil)
	for _, c := range response.Cookies() {
		request.AddCookie(c)
	}
	response = doHttp(request)

	body, err = ioutil.ReadAll(response.Body)
	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", alice, err)

	t.Logf("[?op=logout]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", err)
}

func TestLoginDupLogins(t *testing.T) {
	unit_test.SetTesting(t)

	// login for the first time, should succeed
	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=login", baseURI, alice), nil)
	response := doHttp(request)

	body, err := ioutil.ReadAll(response.Body)
	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", alice, err)

	t.Logf("[?op=login]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", err)

	// now repeat the attempt to login again, which should fail
	request = httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=login", baseURI, alice), nil)
	for _, c := range response.Cookies() {
		request.AddCookie(c)
	}
	response = doHttp(request)

	body, err = ioutil.ReadAll(response.Body)
	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", alice, err)

	t.Logf("[?op=login]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusBadRequest, response.StatusCode, "Handler returned unexpected error: %v", err)
	assert.Equal(t, fmt.Sprintf("%s\n", ErrUserAlreadyLoggedIn.Error()), string(body), "Handler returned unexpected response body: %v", string(body))

	// .. and let's just try with another user, which should also fail
	request = httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=login", baseURI, bob), nil)
	for _, c := range response.Cookies() {
		request.AddCookie(c)
	}
	response = doHttp(request)

	body, err = ioutil.ReadAll(response.Body)
	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", bob, err)

	t.Logf("[?op=login]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusBadRequest, response.StatusCode, "Handler returned unexpected error: %v", err)
	assert.Equal(t, fmt.Sprintf("%s\n", ErrUserAlreadyLoggedIn.Error()), string(body), "Handler returned unexpected response body: %v", string(body))

	// ... and logout, which should succeed
	//
	request = httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=logout", baseURI, alice), nil)
	for _, c := range response.Cookies() {
		request.AddCookie(c)
	}
	response = doHttp(request)

	body, err = ioutil.ReadAll(response.Body)
	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", alice, err)

	t.Logf("[?op=logout]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", err)
}

func TestLoginLogoutDiffAccounts(t *testing.T) {
	unit_test.SetTesting(t)

	// login for the first time, should succeed
	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=login", baseURI, alice), nil)
	response := doHttp(request)

	body, err := ioutil.ReadAll(response.Body)
	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", alice, err)

	t.Logf("[?op=login]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", err)

	// ... and logout, which should not succeed
	//
	request = httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=logout", baseURI, bob), nil)
	for _, c := range response.Cookies() {
		request.AddCookie(c)
	}
	response = doHttp(request)

	body, err = ioutil.ReadAll(response.Body)
	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", bob, err)

	t.Logf("[?op=logout]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusBadRequest, response.StatusCode, "Handler returned unexpected error: %v", err)

	// ... and logout, which should succeed
	//
	request = httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=logout", baseURI, alice), nil)
	for _, c := range response.Cookies() {
		request.AddCookie(c)
	}
	response = doHttp(request)

	body, err = ioutil.ReadAll(response.Body)
	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", alice, err)

	t.Logf("[?op=logout]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", err)
}

func TestDoubleLogout(t *testing.T) {
	unit_test.SetTesting(t)

	// login for the first time, should succeed
	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=login", baseURI, alice), nil)
	response := doHttp(request)

	body, err := ioutil.ReadAll(response.Body)
	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", alice, err)

	t.Logf("[?op=login]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", err)

	// ... logout as 'bob', which should fail
	//
	request = httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=logout", baseURI, bob), nil)
	for _, c := range response.Cookies() {
		request.AddCookie(c)
	}
	response = doHttp(request)

	body, err = ioutil.ReadAll(response.Body)
	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", bob, err)

	t.Logf("[?op=logout]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusBadRequest, response.StatusCode, "Handler returned unexpected error: %v", err)

	// ... logout, which should succeed
	//
	request = httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=logout", baseURI, alice), nil)
	for _, c := range response.Cookies() {
		request.AddCookie(c)
	}
	response = doHttp(request)

	body, err = ioutil.ReadAll(response.Body)
	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", alice, err)

	t.Logf("[?op=logout]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", err)

	// ... logout again, which should fail
	//
	request = httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=logout", baseURI, alice), nil)
	for _, c := range response.Cookies() {
		request.AddCookie(c)
	}
	response = doHttp(request)

	body, err = ioutil.ReadAll(response.Body)
	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", alice, err)

	t.Logf("[?op=logout]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusBadRequest, response.StatusCode, "Handler returned unexpected error: %v", err)
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
