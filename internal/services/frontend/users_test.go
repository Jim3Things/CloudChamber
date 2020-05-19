// Unit tests for the web service frontend.
//
// Borrows heavily from the gorilla mux test package.
//
package frontend

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"

	"github.com/Jim3Things/CloudChamber/internal/config"
	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters"
	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters/unit_test"
	"github.com/Jim3Things/CloudChamber/internal/tracing/setup"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/admin"
)

const (
	userURI       = "/api/users/"
	alice         = userURI + "Alice"
	admin         = userURI + "Admin"
	bob           = userURI + "Bob"
	adminPassword = "AdminPassword"
	alicePassword = "test"
	bobPassword   = "test2"
)

var (
	baseURI     string
	initialized bool
	aliceDef    = &pb.UserDefinition{
		Password:       alicePassword,
		Enabled:        true,
		ManageAccounts: false,
	}
	bobDef = &pb.UserDefinition{
		Password:       bobPassword,
		Enabled:        true,
		ManageAccounts: false,
	}
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
		WebServer: config.WebServerType{
			RootFilePath:          "C:\\CloudChamber",
			SystemAccount:         "Admin",
			SystemAccountPassword: adminPassword,
			FE: config.Endpoint{
				Hostname: "localhost",
				Port:     8080,
			},
			BE: config.Endpoint{},
		},
	}); err != nil {
		log.Fatalf("Error initializing service: %v", err)
	}

	baseURI = fmt.Sprintf("http://localhost:%d", server.port)

	initialized = true
}

// +++ Helper functions

// Convert a proto message into a reader with json-formatted contents
func toJsonReader(v proto.Message) (io.Reader, error) {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	p := jsonpb.Marshaler{}

	if err := p.Marshal(w, v); err != nil {
		return nil, err
	}

	if err := w.Flush(); err != nil {
		return nil, err
	}

	return bufio.NewReader(&buf), nil
}

// Execute an http request/response sequence
func doHTTP(req *http.Request, cookies []*http.Cookie) *http.Response {
	for _, c := range cookies {
		req.AddCookie(c)
	}

	w := httptest.NewRecorder()

	server.handler.ServeHTTP(w, req)

	return w.Result()
}

// Get the body of a response, and close it
func getBody(resp *http.Response) ([]byte, error) {
	defer func() { _ = resp.Body.Close() }()
	return ioutil.ReadAll(resp.Body)
}

// Get the body of a response, unmarshaled into the supplied message structure
func getJsonBody(resp *http.Response, v proto.Message) error {
	defer func() { _ = resp.Body.Close() }()
	return jsonpb.Unmarshal(resp.Body, v)
}

// Log the specified user into CloudChamber
func doLogin(t *testing.T, user string, password string, cookies []*http.Cookie) *http.Response {
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

// Log the specified user out of CloudChamber
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

// Ensure that the specified account exists.  This function first checks if it
// is already known, returning that account's current revision if it is.  If it
// is not, then the account is created using the supplied definition, again
// returning the revision number.
//
// Note that this is mostly used by unit tests in order to support running any
// unit test in isolation from the overall flow.
func ensureAccount(t *testing.T, user string, u *pb.UserDefinition, cookies []*http.Cookie) (int64, []*http.Cookie) {
	path := baseURI + userURI + user

	req := httptest.NewRequest("GET", path, nil)
	req.Header.Set("Content-Type", "application/json")
	response := doHTTP(req, cookies)
	_ = response.Body.Close()

	// If we found the user, just return the existing revision and cookies
	if response.StatusCode == http.StatusOK {
		t.Logf("Found existing user %q.", user)

		var rev int64
		rev, err := strconv.ParseInt(response.Header.Get("ETag"), 10, 64)
		assert.Nilf(t, err, fmt.Sprintf("Error parsing ETag. Value received is : %q", err))

		return rev, response.Cookies()
	}

	// Didn't find the user, create a new incarnation of it.
	t.Logf("Did not find user %q.  Creating it from scratch.", user)

	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)

	req = httptest.NewRequest("POST", path, nil)
	p := jsonpb.Marshaler{}
	err := p.Marshal(w, u)
	assert.Nilf(t, err, fmt.Sprintf("Error formatting the new user definition. err = %v", err))
	_ = w.Flush()
	r := bufio.NewReader(&buf)

	req = httptest.NewRequest("POST", path, r)
	req.Header.Set("Content-Type", "application/json")

	response = doHTTP(req, response.Cookies())
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	tagString := response.Header.Get("ETag")
	tag, err := strconv.ParseInt(tagString, 10, 64)
	assert.Nilf(t, err, fmt.Sprintf("Error parsing ETag. tag = %q, err = %v", tagString, err))

	return tag, response.Cookies()
}

// --- Helper functions

func TestMain(m *testing.M) {

	commonSetup()

	os.Exit(m.Run())
}

// The individual unit tests follow here.  They are grouped by the operation
// they are testing, starting with a simple happy path case, followed by
// repeating sequences (optionally), and then by failure cases.
//
// Each group is demarcated by comment lines.  The group starts with one that
// that has "+++", and ends with one that has "==="

// +++ Login tests

func TestLoginSessionSimple(t *testing.T) {
	unit_test.SetTesting(t)

	// login for the first time, should succeed
	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=login", baseURI, admin), strings.NewReader(adminPassword))
	response := doHTTP(request, nil)
	body, err := getBody(response)

	t.Logf("[?op=login]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	// ... and logout, which should succeed
	//
	request = httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=logout", baseURI, admin), nil)
	response = doHTTP(request, response.Cookies())
	body, err = getBody(response)

	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", admin, err)

	t.Logf("[?op=logout]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)
}

func TestLoginSessionRepeat(t *testing.T) {
	unit_test.SetTesting(t)

	// login for the first time, should succeed
	response := doLogin(t, "Admin", adminPassword, nil)

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	// ... and logout, which should succeed
	//
	response = doLogout(t, "admin", response.Cookies())

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	// login for the second iteration, should succeed
	response = doLogin(t, "Admin", adminPassword, response.Cookies())

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	// ... and logout, which should succeed
	//
	response = doLogout(t, "Admin", response.Cookies())

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)
}

func TestLoginDupLogins(t *testing.T) {
	unit_test.SetTesting(t)

	// login for the first time, should succeed
	response := doLogin(t, "Admin", adminPassword, nil)

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	// now repeat the attempt to login again, which should fail
	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=login", baseURI, admin), strings.NewReader(adminPassword))
	response = doHTTP(request, response.Cookies())
	body, err := getBody(response)

	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", admin, err)

	t.Logf("[?op=login]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusBadRequest, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)
	assert.Equal(t,
		fmt.Sprintf("%s\n", ErrUserAlreadyLoggedIn.Error()), string(body),
		"Handler returned unexpected response body: %v", string(body))

	// .. and let's just try with another user, which should also fail
	request = httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=login", baseURI, bob), strings.NewReader("test2"))
	response = doHTTP(request, response.Cookies())
	body, err = getBody(response)

	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", bob, err)

	t.Logf("[?op=login]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusBadRequest, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)
	assert.Equal(t,
		fmt.Sprintf("%s\n", ErrUserAlreadyLoggedIn.Error()), string(body),
		"Handler returned unexpected response body: %v", string(body))

	// ... and logout, which should succeed
	//
	doLogout(t, "admin", response.Cookies())

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
}

func TestLoginLogoutDiffAccounts(t *testing.T) {
	unit_test.SetTesting(t)

	// login for the first time, should succeed
	response := doLogin(t, "admin", adminPassword, nil)

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	// ... next we need a second account that we're sure exists
	_, cookies := ensureAccount(t, "Alice", aliceDef, response.Cookies())

	// ... and now try to logout from it, which should not succeed
	//
	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=logout", baseURI, alice), nil)
	response = doHTTP(request, cookies)
	body, err := getBody(response)

	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", alice, err)

	t.Logf("[?op=logout]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusBadRequest, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	// ... and logout, which should succeed
	//
	response = doLogout(t, "admin", response.Cookies())

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
}

func TestDoubleLogout(t *testing.T) {
	unit_test.SetTesting(t)

	// login for the first time, should succeed
	response := doLogin(t, "Admin", adminPassword, nil)

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	// ... logout, which should succeed
	//
	response = doLogout(t, "admin", response.Cookies())

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	// ... logout again, which should fail
	//
	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=logout", baseURI, admin), nil)
	response = doHTTP(request, response.Cookies())
	body, err := getBody(response)

	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", admin, err)

	t.Logf("[?op=logout]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusBadRequest, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)
}

func TestLoginSessionBadPassword(t *testing.T) {
	unit_test.SetTesting(t)

	// login for the first time, should succeed
	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=login", baseURI, admin), strings.NewReader(adminPassword+"rubbish"))
	response := doHTTP(request, nil)
	body, err := getBody(response)

	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", baseURI+admin, err)

	t.Logf("[?op=login]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusForbidden, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	// Now just validate that there really isn't an active session here.
	response = doLogin(t, "Admin", adminPassword, response.Cookies())

	doLogout(t, "admin", response.Cookies())
}

func TestLoginSessionNoUser(t *testing.T) {
	unit_test.SetTesting(t)

	// login for the first time, should succeed
	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=login", baseURI, admin+"Bogus"), strings.NewReader(adminPassword))
	response := doHTTP(request, nil)
	body, err := getBody(response)

	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", baseURI+admin, err)

	t.Logf("[?op=login]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusNotFound, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	// Now just validate that there really isn't an active session here.
	response = doLogin(t, "Admin", adminPassword, response.Cookies())

	doLogout(t, "admin", response.Cookies())
}

// --- Login tests

// +++ User creation tests

func TestUsersCreate(t *testing.T) {
	unit_test.SetTesting(t)

	path := baseURI + alice + "2"

	r, err := toJsonReader(aliceDef)
	assert.Nilf(t, err, "Failed to format UserDefinition, err = %v", err)

	response := doLogin(t, "Admin", adminPassword, nil)

	request := httptest.NewRequest("POST", path, r)
	request.Header.Set("Content-Type", "application/json")

	response = doHTTP(request, response.Cookies())
	body, err := getBody(response)

	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", userURI, err)

	t.Logf("[%s]: SC=%v, Content-Type='%v'\n", path, response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)
	assert.Equal(t,
		"User \"Alice2\" created.  enabled: true, can manage accounts: false", string(body),
		"Handler returned unexpected response body: %v", string(body))

	doLogout(t, "admin", response.Cookies())
}

func TestUsersCreateDup(t *testing.T) {
	unit_test.SetTesting(t)

	r, err := toJsonReader(aliceDef)
	assert.Nilf(t, err, "Failed to format UserDefinition, err = %v", err)

	response := doLogin(t, "Admin", adminPassword, nil)

	_, cookies := ensureAccount(t, "Alice", aliceDef, response.Cookies())

	request := httptest.NewRequest("POST", fmt.Sprintf("%s%s", baseURI, alice), r)
	request.Header.Set("Content-Type", "application/json")

	response = doHTTP(request, cookies)
	body, err := getBody(response)

	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", userURI, err)

	t.Logf("[%s]: SC=%v, Content-Type='%v'\n", userURI, response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	assert.Equal(t, http.StatusBadRequest, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)
	assert.Equal(t,
		"CloudChamber: user \"Alice\" already exists\n", string(body),
		"Handler returned unexpected response body: %v", string(body))

	doLogout(t, "admin", response.Cookies())
}

func TestUsersCreateBadData(t *testing.T) {
	unit_test.SetTesting(t)

	response := doLogin(t, "Admin", adminPassword, nil)

	request := httptest.NewRequest(
		"POST",
		fmt.Sprintf("%s%s%s", baseURI, userURI, "Alice2"),
		strings.NewReader("{\"enabled\":123,\"manageAccounts\":false, \"password\":\"test\"}"))
	request.Header.Set("Content-Type", "application/json")

	response = doHTTP(request, response.Cookies())
	body, err := getBody(response)

	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", userURI, err)

	t.Logf("[%s]: SC=%v, Content-Type='%v'\n", userURI, response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	assert.Equal(t, http.StatusBadRequest, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)
	assert.Equal(t,
		"json: cannot unmarshal number into Go value of type bool\n", string(body),
		"Handler returned unexpected response body: %v", string(body))

	doLogout(t, "admin", response.Cookies())
}

func TestUsersCreateNoPriv(t *testing.T) {
	unit_test.SetTesting(t)

	r, err := toJsonReader(bobDef)
	assert.Nilf(t, err, "Failed to format UserDefinition, err = %v", err)

	response := doLogin(t, "Alice", alicePassword, nil)

	request := httptest.NewRequest("POST", fmt.Sprintf("%s%s", baseURI, bob), r)
	request.Header.Set("Content-Type", "application/json")

	response = doHTTP(request, response.Cookies())
	body, err := getBody(response)

	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", userURI, err)

	t.Logf("[%s]: SC=%v, Content-Type='%v'\n", userURI, response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	assert.Equal(t, http.StatusForbidden, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)
	assert.Equal(t,
		"CloudChamber: permission denied\n", string(body),
		"Handler returned unexpected response body: %v", string(body))

	doLogout(t, "Alice", response.Cookies())
}

func TestUsersCreateNoSession(t *testing.T) {
	unit_test.SetTesting(t)

	path := baseURI + alice + "2"

	r, err := toJsonReader(aliceDef)
	assert.Nilf(t, err, "Failed to format UserDefinition, err = %v", err)

	request := httptest.NewRequest("POST", path, r)
	request.Header.Set("Content-Type", "application/json")

	response := doHTTP(request, nil)
	body, err := getBody(response)

	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", userURI, err)

	t.Logf("[%s]: SC=%v, Content-Type='%v'\n", path, response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	assert.Equal(t, http.StatusBadRequest, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)
	assert.Equal(t,
		"http: named cookie not present\n", string(body),
		"Handler returned unexpected response body: %v", string(body))
}

// --- User creation tests

// +++ Known users list tests

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

	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	doLogout(t, "admin", response.Cookies())
}

func TestUsersListNoPriv(t *testing.T) {
	unit_test.SetTesting(t)

	response := doLogin(t, "Admin", adminPassword, nil)
	_, cookies := ensureAccount(t, "Alice", aliceDef, response.Cookies())
	response = doLogout(t, "Admin", cookies)

	response = doLogin(t, "Alice", alicePassword, response.Cookies())

	request := httptest.NewRequest("GET", fmt.Sprintf("%s%s", baseURI, userURI), nil)

	response = doHTTP(request, response.Cookies())
	body, err := getBody(response)

	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", userURI, err)

	t.Logf("[%s]: SC=%v, Content-Type='%v'\n", userURI, response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	assert.Equal(t, http.StatusForbidden, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	doLogout(t, "alice", response.Cookies())
}

// --- Known user list tests

// +++ Get user details tests

func TestUsersRead(t *testing.T) {
	unit_test.SetTesting(t)

	response := doLogin(t, "Admin", adminPassword, nil)

	request := httptest.NewRequest("GET", fmt.Sprintf("%s%s%s", baseURI, userURI, "admin"), nil)
	request.Header.Set("Content-Type", "application/json")

	response = doHTTP(request, response.Cookies())

	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	user := &pb.UserPublic{}
	err := getJsonBody(response, user)
	assert.Nilf(t, err, "Failed to convert body to valid json.  err: %v", err)

	assert.Equal(t, "application/json", strings.ToLower(response.Header.Get("Content-Type")))
	assert.Equal(t, "1", response.Header.Get("ETag"))
	assert.Equal(t, true, user.Enabled)
	assert.Equal(t, true, user.AccountManager)

	doLogout(t, "admin", response.Cookies())
}

func TestUsersReadUnknownUser(t *testing.T) {
	unit_test.SetTesting(t)

	response := doLogin(t, "Admin", adminPassword, nil)

	request := httptest.NewRequest("GET", fmt.Sprintf("%s%s%s", baseURI, userURI, "BadUser"), nil)
	request.Header.Set("Content-Type", "application/json")

	response = doHTTP(request, response.Cookies())
	body, err := getBody(response)
	t.Log(string(body))

	assert.Nilf(t, err, "Expected no error in getting body.  err=%v", err)
	assert.NotEqual(t, "application/json", strings.ToLower(response.Header.Get("Content-Type")))

	assert.Equal(t, http.StatusNotFound, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	doLogout(t, "admin", response.Cookies())
}

func TestUsersReadNoPriv(t *testing.T) {
	unit_test.SetTesting(t)

	response := doLogin(t, "Admin", adminPassword, nil)
	_, cookies := ensureAccount(t, "Alice", aliceDef, response.Cookies())
	response = doLogout(t, "Admin", cookies)

	response = doLogin(t, "Alice", alicePassword, response.Cookies())

	request := httptest.NewRequest("GET", fmt.Sprintf("%s%s%s", baseURI, userURI, "BadUser"), nil)
	request.Header.Set("Content-Type", "application/json")

	response = doHTTP(request, response.Cookies())
	body, err := getBody(response)
	t.Log(string(body))

	assert.Nilf(t, err, "Expected no error in getting body.  err=%v", err)
	assert.NotEqual(t, "application/json", strings.ToLower(response.Header.Get("Content-Type")))

	assert.Equal(t, http.StatusForbidden, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	doLogout(t, "alice", response.Cookies())
}

// --- Get user details tests

// +++ User operation (?op=) tests

func TestUserOperationsDisable(t *testing.T) {
	unit_test.SetTesting(t)

	response := doLogin(t, "Admin", adminPassword, nil)

	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=disable", baseURI, alice), nil)
	response = doHTTP(request, response.Cookies())
	body, err := getBody(response)

	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", alice, err)

	t.Logf("[?op=disable]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)
	assert.Equal(t, "User \"Alice\" disabled\n", string(body), "Handler returned unexpected response body: %v", string(body))

	doLogout(t, "admin", response.Cookies())
}

func TestUsersOperationEnable(t *testing.T) {
	unit_test.SetTesting(t)

	response := doLogin(t, "Admin", adminPassword, nil)

	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=enable", baseURI, alice), nil)
	response = doHTTP(request, response.Cookies())
	body, err := getBody(response)

	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", alice, err)

	t.Logf("[?op=enable]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)
	assert.Equal(t,
		"User \"Alice\" enabled\n", string(body),
		"Handler returned unexpected response body: %v", string(body))

	doLogout(t, "admin", response.Cookies())
}

func TestUsersOperationEnableNoUser(t *testing.T) {
	unit_test.SetTesting(t)

	response := doLogin(t, "Admin", adminPassword, nil)

	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=enable", baseURI, alice+"Bogus"), nil)
	response = doHTTP(request, response.Cookies())
	body, err := getBody(response)

	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", alice, err)

	t.Logf("[?op=enable]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	assert.Equal(t, http.StatusNotFound, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)
	assert.Equal(t,
		"CloudChamber: user \"AliceBogus\" not found\n", string(body),
		"Handler returned unexpected response body: %v", string(body))

	doLogout(t, "admin", response.Cookies())
}

func TestUsersOperationNoPriv(t *testing.T) {
	unit_test.SetTesting(t)

	response := doLogin(t, "Admin", adminPassword, nil)
	_, cookies := ensureAccount(t, "Alice", aliceDef, response.Cookies())
	_, cookies = ensureAccount(t, "Bob", bobDef, cookies)
	response = doLogout(t, "Admin", cookies)

	response = doLogin(t, "Bob", bobPassword, response.Cookies())

	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=enable", baseURI, alice), nil)
	response = doHTTP(request, response.Cookies())
	body, err := getBody(response)

	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", alice, err)

	t.Logf("[?op=enable]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	assert.Equal(t, http.StatusForbidden, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)
	assert.Equal(t,
		"CloudChamber: permission denied\n", string(body),
		"Handler returned unexpected response body: %v", string(body))

	doLogout(t, "Bob", response.Cookies())
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

	assert.Equal(t, http.StatusBadRequest, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)
	assert.Equal(t,
		"invalid user operation requested (?op=)\n", string(body),
		"Handler returned unexpected response body: %v", string(body))

	// Case 2, check that an invalid op fails
	//
	request = httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=testInvalid", baseURI, alice), nil)
	response = doHTTP(request, response.Cookies())
	body, err = getBody(response)

	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", alice, err)

	t.Logf("[?op=testInvalid]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	assert.Equal(t, http.StatusBadRequest, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)
	assert.Equal(t,
		"invalid user operation requested (?op=testInvalid)\n", string(body),
		"Handler returned unexpected response body: %v", string(body))
}

// --- User operation (?op=) tests

// +++ Update user tests

func TestUsersUpdate(t *testing.T) {
	aliceUpd := &pb.UserDefinition{
		Password:       alicePassword,
		Enabled:        true,
		ManageAccounts: true,
	}

	unit_test.SetTesting(t)

	response := doLogin(t, "Admin", adminPassword, nil)

	r, err := toJsonReader(aliceUpd)
	assert.Nilf(t, err, "Failed to format UserDefinition, err = %v", err)

	rev, cookies := ensureAccount(t, "Alice", aliceDef, response.Cookies())
	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s", baseURI, alice), r)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("If-Match", fmt.Sprintf("%v", rev))

	response = doHTTP(request, cookies)
	assert.Equal(t, "application/json", strings.ToLower(response.Header.Get("Content-Type")))

	user := &pb.UserPublic{}
	err = getJsonBody(response, user)
	assert.Nilf(t, err, "Failed to convert body to valid json.  err: %v", err)

	assert.Equal(t, fmt.Sprintf("%v", rev+1), response.Header.Get("ETag"))
	assert.Equal(t, true, user.Enabled)
	assert.Equal(t, true, user.AccountManager)

	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	doLogout(t, "admin", response.Cookies())
}

func TestUsersUpdateBadData(t *testing.T) {
	unit_test.SetTesting(t)

	response := doLogin(t, "Admin", adminPassword, nil)

	rev, cookies := ensureAccount(t, "Alice", aliceDef, response.Cookies())

	request := httptest.NewRequest(
		"PUT",
		fmt.Sprintf("%s%s", baseURI, alice),
		strings.NewReader("{\"enabled\":123,\"manageAccounts\":false, \"password\":\"test\"}"))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("If-Match", fmt.Sprintf("%v", rev))

	response = doHTTP(request, cookies)
	body, err := getBody(response)

	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", userURI, err)

	t.Logf("[%s]: SC=%v, Content-Type='%v'\n", userURI, response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	assert.Equal(t, http.StatusBadRequest, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)
	assert.Equal(t,
		"json: cannot unmarshal number into Go value of type bool\n", string(body),
		"Handler returned unexpected response body: %v", string(body))

	doLogout(t, "admin", response.Cookies())
}

func TestUsersUpdateBadMatch(t *testing.T) {
	aliceUpd := &pb.UserDefinition{
		Password:       alicePassword,
		Enabled:        true,
		ManageAccounts: true,
	}

	unit_test.SetTesting(t)

	response := doLogin(t, "Admin", adminPassword, nil)

	r, err := toJsonReader(aliceUpd)
	assert.Nilf(t, err, "Failed to format UserDefinition, err = %v", err)

	rev, cookies := ensureAccount(t, "Alice", aliceDef, response.Cookies())
	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s", baseURI, alice), r)

	// Poison the revision
	rev += 10

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("If-Match", fmt.Sprintf("%v", rev))

	response = doHTTP(request, cookies)
	body, err := getBody(response)

	t.Log(string(body))
	assert.Nilf(t, err, "Failed to get response body.  err: %v", err)

	assert.Equal(t, http.StatusConflict, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	doLogout(t, "admin", response.Cookies())
}

func TestUsersUpdateBadMatchSyntax(t *testing.T) {
	aliceUpd := &pb.UserDefinition{
		Password:       alicePassword,
		Enabled:        true,
		ManageAccounts: true,
	}

	unit_test.SetTesting(t)

	response := doLogin(t, "Admin", adminPassword, nil)

	r, err := toJsonReader(aliceUpd)
	assert.Nilf(t, err, "Failed to format UserDefinition, err = %v", err)

	_, cookies := ensureAccount(t, "Alice", aliceDef, response.Cookies())
	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s", baseURI, alice), r)

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("If-Match", "abc")

	response = doHTTP(request, cookies)
	body, err := getBody(response)

	t.Log(string(body))
	assert.Nilf(t, err, "Failed to get response body.  err: %v", err)

	assert.Equal(t, http.StatusBadRequest, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	doLogout(t, "admin", response.Cookies())
}

func TestUsersUpdateNoUser(t *testing.T) {
	upd := &pb.UserDefinition{
		Password:       "bogus",
		Enabled:        true,
		ManageAccounts: true,
	}

	unit_test.SetTesting(t)

	response := doLogin(t, "Admin", adminPassword, nil)

	r, err := toJsonReader(upd)
	assert.Nilf(t, err, "Failed to format UserDefinition, err = %v", err)

	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s", baseURI, userURI+"BadUser"), r)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("If-Match", fmt.Sprintf("%v", "1"))

	response = doHTTP(request, response.Cookies())
	body, err := getBody(response)
	t.Logf(string(body))
	assert.Nilf(t, err, "Error reading body, err = %v", err)

	assert.Equal(t, http.StatusNotFound, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	doLogout(t, "admin", response.Cookies())
}

func TestUsersUpdateNoPriv(t *testing.T) {
	aliceUpd := &pb.UserDefinition{
		Password:       alicePassword,
		Enabled:        true,
		ManageAccounts: true,
	}

	unit_test.SetTesting(t)

	response := doLogin(t, "Admin", adminPassword, nil)
	_, cookies := ensureAccount(t, "Alice", aliceDef, response.Cookies())
	rev, cookies := ensureAccount(t, "Bob", bobDef, cookies)
	response = doLogout(t, "Admin", cookies)

	response = doLogin(t, "Bob", bobPassword, response.Cookies())

	r, err := toJsonReader(aliceUpd)
	assert.Nilf(t, err, "Failed to format UserDefinition, err = %v", err)

	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s", baseURI, alice), r)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("If-Match", fmt.Sprintf("%v", rev))

	response = doHTTP(request, response.Cookies())
	body, err := getBody(response)

	t.Logf(string(body))

	assert.Equal(t, http.StatusForbidden, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	doLogout(t, "bob", response.Cookies())
}

// --- Update user tests

// +++ Delete user tests

func TestUsersDelete(t *testing.T) {
	unit_test.SetTesting(t)

	path := fmt.Sprintf("%s%s", baseURI, alice)

	response := doLogin(t, "Admin", adminPassword, nil)

	_, cookies := ensureAccount(t, "Alice", aliceDef, response.Cookies())
	request := httptest.NewRequest("DELETE", path, nil)
	request.Header.Set("Content-Type", "application/json")

	response = doHTTP(request, cookies)
	body, err := getBody(response)

	assert.Nilf(t, err, "Unable to retrieve response body, err = %v", err)
	t.Log(string(body))

	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	// Now verify the deletion by trying to get the user

	request = httptest.NewRequest("GET", path, nil)

	response = doHTTP(request, response.Cookies())
	body, err = getBody(response)

	assert.Nilf(t, err, "Unable to retrieve response body, err = %v", err)
	t.Log(string(body))

	assert.Equal(t, http.StatusNotFound, response.StatusCode, "Found deleted user %q", path)

	doLogout(t, "admin", response.Cookies())
}

func TestUsersDeleteNoUser(t *testing.T) {
	unit_test.SetTesting(t)

	path := fmt.Sprintf("%s%s", baseURI, alice+"Bogus")

	response := doLogin(t, "Admin", adminPassword, nil)

	request := httptest.NewRequest("DELETE", path, nil)
	request.Header.Set("Content-Type", "application/json")

	response = doHTTP(request, response.Cookies())
	body, err := getBody(response)

	assert.Nilf(t, err, "Unable to retrieve response body, err = %v", err)
	t.Log(string(body))

	assert.Equal(t, http.StatusNotFound, response.StatusCode, "Found deleted user %q", path)

	doLogout(t, "admin", response.Cookies())
}

func TestUsersDeleteNoPriv(t *testing.T) {
	unit_test.SetTesting(t)

	response := doLogin(t, "Admin", adminPassword, nil)
	_, cookies := ensureAccount(t, "Alice", aliceDef, response.Cookies())
	_, cookies = ensureAccount(t, "Bob", bobDef, cookies)
	response = doLogout(t, "Admin", cookies)

	response = doLogin(t, "Bob", bobPassword, response.Cookies())

	request := httptest.NewRequest("DELETE", fmt.Sprintf("%s%s", baseURI, alice), nil)

	response = doHTTP(request, response.Cookies())
	body, err := getBody(response)

	assert.Nilf(t, err, "Unable to retrieve response body, err = %v", err)
	t.Logf(string(body))

	assert.Equal(t, http.StatusForbidden, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	doLogout(t, "bob", response.Cookies())
}

// --- Delete user tests
