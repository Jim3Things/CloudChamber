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
	alice = userURI + "Alice"
	admin = userURI + "Admin"
	bob = userURI + "Bob"
	adminPassword = "AdminPassword"
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
			SystemAccountPassword: adminPassword,
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
func doHTTP(req *http.Request, cookies []*http.Cookie) *http.Response {
	for _, c := range cookies {
		req.AddCookie(c)
	}

	w := httptest.NewRecorder()

	server.handler.ServeHTTP(w, req)

	return w.Result()
}

func getBody(resp *http.Response) ([]byte, error) {
	defer func() { _ = resp.Body.Close() }()
	return ioutil.ReadAll(resp.Body)
}

func doLogin(t *testing.T, user string, password string, cookies []*http.Cookie) *http.Response{
	path := fmt.Sprintf("%s%s%s?op=login", baseURI, userURI, user)
	t.Logf("[login as %q (%q)]", user, path)

	request := httptest.NewRequest("PUT", path, strings.NewReader(password))
	response := doHTTP(request, cookies)
	_, err := getBody(response)

	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %q: %v", path, err)
	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	return response
}

func doLogout(t *testing.T, user string, cookies []*http.Cookie) *http.Response {
	path := fmt.Sprintf("%s%s%s?op=logout", baseURI, userURI, user)
	t.Logf("[logout from %q (%q)]", user, path)

	request := httptest.NewRequest("PUT", path, nil)
	response := doHTTP(request, cookies)
	_, err := getBody(response)

	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", alice, err)
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	return response
}

func TestMain(m *testing.M) {

	commonSetup()

	os.Exit(m.Run())
}

func TestLoginSessionSimple(t *testing.T)  {
	unit_test.SetTesting(t)

	// login for the first time, should succeed
	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=login", baseURI, admin), strings.NewReader(adminPassword))
	response := doHTTP(request, nil)
	body, err := getBody(response)

	t.Logf("[?op=login]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)
	assert.Nilf(t, response.Body.Close(), "Failed to successfully close the response")

	// ... and logout, which should succeed
	//
	request = httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=logout", baseURI, admin), nil)
	response = doHTTP(request, response.Cookies())
	body, err = getBody(response)

	assert.Nilf(t, response.Body.Close(), "Failed to successfully close the response")
	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", admin, err)

	t.Logf("[?op=logout]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)
}

func TestLogingSessionBadPassword(t *testing.T) {
	unit_test.SetTesting(t)

	// login for the first time, should succeed
	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=login", baseURI, admin), strings.NewReader(adminPassword + "rubbish"))
	response := doHTTP(request, nil)
	body, err := getBody(response)

	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", baseURI + admin, err)

	t.Logf("[?op=login]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusForbidden, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	// Now just validate that there really isn't an active session here.
	response = doLogin(t, "Admin", adminPassword, response.Cookies())
	doLogout(t, "admin", response.Cookies())
}

func TestUsersCreate(t *testing.T) {
	unit_test.SetTesting(t)

	response := doLogin(t, "Admin", adminPassword, nil)

	request := httptest.NewRequest(
		"POST",
		fmt.Sprintf("%s%s%s", baseURI, userURI, "Alice"),
		strings.NewReader("{\"enabled\":true,\"manageAccounts\":false, \"password\":\"test\"}"))
	request.Header.Set("Content-Type", "application/json")

	response = doHTTP(request, response.Cookies())
	body, err := getBody(response)

	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", userURI, err)

	t.Logf("[%s]: SC=%v, Content-Type='%v'\n", userURI, response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)
	assert.Equal(t, "User \"Alice\" created.  enabled: true, can manage accounts: false", string(body), "Handler returned unexpected response body: %v", string(body))
	doLogout(t, "admin", response.Cookies())
}

func TestUsersCreateDup(t *testing.T) {
	unit_test.SetTesting(t)

	response := doLogin(t, "Admin", adminPassword, nil)

	request := httptest.NewRequest(
		"POST",
		fmt.Sprintf("%s%s%s", baseURI, userURI, "Alice"),
		strings.NewReader("{\"enabled\":true,\"manageAccounts\":false, \"password\":\"test\"}"))
	request.Header.Set("Content-Type", "application/json")

	response = doHTTP(request, response.Cookies())
	body, err := getBody(response)

	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", userURI, err)

	t.Logf("[%s]: SC=%v, Content-Type='%v'\n", userURI, response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	assert.Equal(t, http.StatusBadRequest, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)
	assert.Equal(t, "CloudChamber: user \"Alice\" already exists\n", string(body), "Handler returned unexpected response body: %v", string(body))
	doLogout(t, "admin", response.Cookies())
}

func TestUsersCreateNoPriv(t *testing.T) {
	unit_test.SetTesting(t)

	response := doLogin(t, "Alice", "test", nil)

	request := httptest.NewRequest(
		"POST",
		fmt.Sprintf("%s%s%s", baseURI, userURI, "Bob"),
		strings.NewReader("{\"enabled\":true,\"manageAccounts\":false, \"password\":\"test\"}"))
	request.Header.Set("Content-Type", "application/json")

	response = doHTTP(request, response.Cookies())
	body, err := getBody(response)

	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", userURI, err)

	t.Logf("[%s]: SC=%v, Content-Type='%v'\n", userURI, response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	assert.Equal(t, http.StatusForbidden, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)
	assert.Equal(t, "CloudChamber: permission denied\n", string(body), "Handler returned unexpected response body: %v", string(body))
	doLogout(t, "Alice", response.Cookies())
}

func TestUsersList(t *testing.T) {
	unit_test.SetTesting(t)

	response := doLogin(t, "Admin", adminPassword, nil)

	request := httptest.NewRequest("GET", fmt.Sprintf("%s%s", baseURI, userURI), nil)

	response = doHTTP(request, response.Cookies())
	body, err := getBody(response)

	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", userURI, err)

	t.Logf("[%s]: SC=%v, Content-Type='%v'\n", userURI, response.StatusCode, response.Header.Get("Content-Type"))
	s := string(body)

	t.Log(s)
	ok := s == "Users (List)\nhttp://localhost:8080/api/users/Admin\nhttp://localhost:8080/api/users/Alice\n" ||
		  s == "Users (List)\nhttp://localhost:8080/api/users/Alice\nhttp://localhost:8080/api/users/Admin\n"

	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)
	assert.True(t, ok, "Handler returned unexpected response body: %v", s)
	doLogout(t, "admin", response.Cookies())
}

func TestUsersRead(t *testing.T) {
	unit_test.SetTesting(t)

	response := doLogin(t, "Admin", adminPassword, nil)

	request := httptest.NewRequest("GET", fmt.Sprintf("%s%s%s", baseURI, userURI, "admin"), nil)
	request.Header.Set("Content-Type", "application/json")

	response = doHTTP(request, response.Cookies())

	user := &pb.UserPublic{}
	err := jsonpb.Unmarshal(response.Body, user)
	assert.Nilf(t, err, "Failed to convert body to valid json.  err: %v", err)

	assert.Nilf(t, response.Body.Close(), "Failed to successfully close the response")
	assert.Equal(t, "application/json", strings.ToLower(response.Header.Get("Content-Type")))
	assert.Equal(t, true, user.Enabled)
	assert.Equal(t, true, user.AccountManager)

	// At present, all base handlers effectively echo the supplied username so all
	// we need to verify is that we get a successful return of the supplied username.
	//
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)
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
	response := doHTTP(request, nil)

	response = doHTTP(request, response.Cookies())
	body, err := getBody(response)

	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", alice, err)

	t.Logf("[?op]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	// At present, all base handlers effectively echo the supplied username so all
	// we need to verify is that we get a successful return of the supplied username.
	//
	assert.Equal(t, http.StatusBadRequest, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)
	assert.Equal(t, "invalid user operation requested (?op=)\n", string(body), "Handler returned unexpected response body: %v", string(body))

	// Case 2, check that an invalid op fails
	//
	request = httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=testInvalid", baseURI, alice), nil)
	response = doHTTP(request, response.Cookies())
	body, err = getBody(response)

	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", alice, err)

	t.Logf("[?op=testInvalid]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	// At present, all base handlers effectively echo the supplied username so all
	// we need to verify is that we get a successful return of the supplied username.
	//
	assert.Equal(t, http.StatusBadRequest, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)
	assert.Equal(t, "invalid user operation requested (?op=testInvalid)\n", string(body), "Handler returned unexpected response body: %v", string(body))
}

func TestUserOperationsDisable(t *testing.T) {
	response := doLogin(t, "Admin", adminPassword, nil)

	// 3b, disable
	//
	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=disable", baseURI, alice), nil)
	response = doHTTP(request, response.Cookies())
	body, err := getBody(response)

	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", alice, err)

	t.Logf("[?op=disable]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	// At present, all base handlers effectively echo the supplied username so all
	// we need to verify is that we get a successful return of the supplied username.
	//
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)
	assert.Equal(t, "User \"Alice\" disabled\n", string(body), "Handler returned unexpected response body: %v", string(body))
	doLogout(t, "admin", response.Cookies())
}

func TestUsersOperationEnable(t *testing.T) {
	response := doLogin(t, "Admin", adminPassword, nil)

	// Case 3, check that each of the valid ops succeed
	//
	// 3a, enable
	//
	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=enable", baseURI, alice), nil)
	response = doHTTP(request, response.Cookies())
	body, err := getBody(response)

	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", alice, err)

	t.Logf("[?op=enable]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	// At present, all base handlers effectively echo the supplied username so all
	// we need to verify is that we get a successful return of the supplied username.
	//
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)
	assert.Equal(t, "User \"Alice\" enabled\n", string(body), "Handler returned unexpected response body: %v", string(body))
	doLogout(t, "admin", response.Cookies())
}

func TestLoginSessionRepeat(t *testing.T)  {
	unit_test.SetTesting(t)

	// login for the first time, should succeed
	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=login", baseURI, alice), strings.NewReader("test"))
	response := doHTTP(request, nil)
	body, err := getBody(response)

	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", alice, err)

	t.Logf("[?op=login]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	// ... and logout, which should succeed
	//
	request = httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=logout", baseURI, alice), nil)
	response = doHTTP(request, response.Cookies())
	body, err = getBody(response)

	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", alice, err)

	t.Logf("[?op=logout]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	// login for the second iteration, should succeed
	request = httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=login", baseURI, alice), strings.NewReader("test"))
	response = doHTTP(request, response.Cookies())
	body, err = getBody(response)

	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", alice, err)

	t.Logf("[?op=login]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	// ... and logout, which should succeed
	//
	request = httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=logout", baseURI, alice), nil)
	response = doHTTP(request, response.Cookies())
	body, err = getBody(response)

	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", alice, err)

	t.Logf("[?op=logout]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)
}

func TestLoginDupLogins(t *testing.T) {
	unit_test.SetTesting(t)

	// login for the first time, should succeed
	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=login", baseURI, alice), strings.NewReader("test"))
	response := doHTTP(request, nil)
	body, err := getBody(response)

	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", alice, err)

	t.Logf("[?op=login]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	// now repeat the attempt to login again, which should fail
	request = httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=login", baseURI, alice), strings.NewReader("test"))
	response = doHTTP(request, response.Cookies())
	body, err = getBody(response)

	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", alice, err)

	t.Logf("[?op=login]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusBadRequest, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)
	assert.Equal(t, fmt.Sprintf("%s\n", ErrUserAlreadyLoggedIn.Error()), string(body), "Handler returned unexpected response body: %v", string(body))

	// .. and let's just try with another user, which should also fail
	request = httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=login", baseURI, bob), strings.NewReader("test2"))
	response = doHTTP(request, response.Cookies())
	body, err = getBody(response)

	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", bob, err)

	t.Logf("[?op=login]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusBadRequest, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)
	assert.Equal(t, fmt.Sprintf("%s\n", ErrUserAlreadyLoggedIn.Error()), string(body), "Handler returned unexpected response body: %v", string(body))

	// ... and logout, which should succeed
	//
	request = httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=logout", baseURI, alice), nil)
	response = doHTTP(request, response.Cookies())
	body, err = getBody(response)

	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", alice, err)

	t.Logf("[?op=logout]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)
}

func TestLoginLogoutDiffAccounts(t *testing.T) {
	unit_test.SetTesting(t)

	// login for the first time, should succeed
	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=login", baseURI, alice), strings.NewReader("test"))
	response := doHTTP(request, nil)
	body, err := getBody(response)

	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", alice, err)

	t.Logf("[?op=login]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	// ... and logout, which should not succeed
	//
	request = httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=logout", baseURI, bob), nil)
	response = doHTTP(request, response.Cookies())
	body, err = getBody(response)

	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", bob, err)

	t.Logf("[?op=logout]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusBadRequest, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	// ... and logout, which should succeed
	//
	request = httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=logout", baseURI, alice), nil)
	response = doHTTP(request, response.Cookies())
	body, err = getBody(response)

	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", alice, err)

	t.Logf("[?op=logout]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)
}

func TestDoubleLogout(t *testing.T) {
	unit_test.SetTesting(t)

	// login for the first time, should succeed
	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=login", baseURI, alice), strings.NewReader("test"))
	response := doHTTP(request, nil)
	body, err := getBody(response)

	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", alice, err)

	t.Logf("[?op=login]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	// ... logout as 'bob', which should fail
	//
	request = httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=logout", baseURI, bob), nil)
	response = doHTTP(request, response.Cookies())
	body, err = getBody(response)

	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", bob, err)

	t.Logf("[?op=logout]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusBadRequest, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	// ... logout, which should succeed
	//
	request = httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=logout", baseURI, alice), nil)
	response = doHTTP(request, response.Cookies())
	body, err = getBody(response)

	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", alice, err)

	t.Logf("[?op=logout]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	// ... logout again, which should fail
	//
	request = httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=logout", baseURI, alice), nil)
	response = doHTTP(request, response.Cookies())
	body, err = getBody(response)

	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", alice, err)

	t.Logf("[?op=logout]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusBadRequest, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)
}

func TestUsersUpdate(t *testing.T) {

}

func TestUsersDelete(t *testing.T) {

}
