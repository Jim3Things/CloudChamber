// Unit tests for the web service frontend.
//
// Borrows heavily from the gorilla mux test package.
//
package frontend

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/golang/protobuf/jsonpb"
	"github.com/stretchr/testify/assert"

	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters/unit_test"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/admin"
)

const (
	userURI       = "/api/users/"
	admin         = userURI + adminAccountName
	alice         = userURI + "Alice"
	bob           = userURI + "Bob"
	alicePassword = "test"
	bobPassword   = "test2"
)

var (
	aliceDef = &pb.UserDefinition{
		Password:          alicePassword,
		Enabled:           true,
		CanManageAccounts: false,
	}
	bobDef = &pb.UserDefinition{
		Password:          bobPassword,
		Enabled:           true,
		CanManageAccounts: false,
	}

	// The user URLs that have been added and not deleted during the test run.
	// Note that this does not include any predefined users, such as Admin.
	knownNames = make(map[string]string)
)

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

	p := jsonpb.Marshaler{}
	err := p.Marshal(w, u)
	assert.Nilf(t, err, fmt.Sprintf("Error formatting the new user definition. err = %v", err))
	_ = w.Flush()
	r := bufio.NewReader(&buf)

	req = httptest.NewRequest("POST", path, r)
	req.Header.Set("Content-Type", "application/json")

	response = doHTTP(req, response.Cookies())
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	knownNames[path] = path

	tagString := response.Header.Get("ETag")
	tag, err := strconv.ParseInt(tagString, 10, 64)
	assert.Nilf(t, err, fmt.Sprintf("Error parsing ETag. tag = %q, err = %v", tagString, err))

	return tag, response.Cookies()
}

// --- Helper functions

// The individual unit tests follow here.  They are grouped by the operation
// they are testing, starting with a simple happy path case, followed by
// repeating sequences (optionally), and then by failure cases.
//
// Each group is demarcated by comment lines.  The group starts with one that
// that has "+++", and ends with one that has "==="

// +++ Login tests

func TestUsersLoginSessionSimple(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	// login for the first time, should succeed
	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=login", baseURI, admin), strings.NewReader(adminPassword))
	response := doHTTP(request, nil)
	body, err := getBody(response)

	t.Logf("[?op=login]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	// ... and logout, which should succeed
	//     (note that this also checks that the username match is case insensitive)
	//
	request = httptest.NewRequest("PUT", fmt.Sprintf("%s%s?op=logout", baseURI, strings.ToUpper(admin)), nil)
	response = doHTTP(request, response.Cookies())
	body, err = getBody(response)

	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", admin, err)

	t.Logf("[?op=logout]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)
}

func TestUsersLoginSessionRepeat(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	// login for the first time, should succeed
	response := doLogin(t, randomCase(adminAccountName), adminPassword, nil)

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	// ... and logout, which should succeed
	//
	response = doLogout(t, randomCase(adminAccountName), response.Cookies())

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	// login for the second iteration, should succeed
	response = doLogin(t, randomCase(adminAccountName), adminPassword, response.Cookies())

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	// ... and logout, which should succeed
	//
	response = doLogout(t, randomCase(adminAccountName), response.Cookies())

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)
}

func TestUsersLoginDupLogins(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	// login for the first time, should succeed
	response := doLogin(t, randomCase(adminAccountName), adminPassword, nil)

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
	doLogout(t, randomCase(adminAccountName), response.Cookies())

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
}

func TestUsersLoginLogoutDiffAccounts(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	// login for the first time, should succeed
	response := doLogin(t, randomCase(adminAccountName), adminPassword, nil)

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
	response = doLogout(t, randomCase(adminAccountName), response.Cookies())

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
}

func TestUsersDoubleLogout(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	// login for the first time, should succeed
	response := doLogin(t, randomCase(adminAccountName), adminPassword, nil)

	assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	// ... logout, which should succeed
	//
	response = doLogout(t, randomCase(adminAccountName), response.Cookies())

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

func TestUsersLoginSessionBadPassword(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

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
	response = doLogin(t, randomCase(adminAccountName), adminPassword, response.Cookies())

	doLogout(t, randomCase(adminAccountName), response.Cookies())
}

func TestUsersLoginSessionNoUser(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

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
	response = doLogin(t, randomCase(adminAccountName), adminPassword, response.Cookies())

	doLogout(t, randomCase(adminAccountName), response.Cookies())
}

// --- Login tests

// +++ User creation tests

func TestUsersCreate(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	path := baseURI + alice + "2"

	r, err := toJSONReader(aliceDef)
	assert.Nilf(t, err, "Failed to format UserDefinition, err = %v", err)

	response := doLogin(t, randomCase(adminAccountName), adminPassword, nil)

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

	knownNames[path] = path
	doLogout(t, randomCase(adminAccountName), response.Cookies())
}

func TestUsersCreateDup(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	r, err := toJSONReader(aliceDef)
	assert.Nilf(t, err, "Failed to format UserDefinition, err = %v", err)

	response := doLogin(t, randomCase(adminAccountName), adminPassword, nil)

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

	doLogout(t, randomCase(adminAccountName), response.Cookies())
}

func TestUsersCreateBadData(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	response := doLogin(t, randomCase(adminAccountName), adminPassword, nil)

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

	doLogout(t, randomCase(adminAccountName), response.Cookies())
}

func TestUsersCreateNoPriv(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	r, err := toJSONReader(bobDef)
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
	defer unit_test.SetTesting(nil)

	path := baseURI + alice + "2"

	r, err := toJSONReader(aliceDef)
	assert.Nilf(t, err, "Failed to format UserDefinition, err = %v", err)

	request := httptest.NewRequest("POST", path, r)
	request.Header.Set("Content-Type", "application/json")

	response := doHTTP(request, nil)
	body, err := getBody(response)

	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", userURI, err)

	t.Logf("[%s]: SC=%v, Content-Type='%v'\n", path, response.StatusCode, response.Header.Get("Content-Type"))
	t.Log(string(body))

	assert.Equal(t, http.StatusForbidden, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)
	assert.Equal(t,
		"CloudChamber: permission denied\n", string(body),
		"Handler returned unexpected response body: %v", string(body))
}

// --- User creation tests

// +++ Known users list tests

func TestUsersList(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	response := doLogin(t, randomCase(adminAccountName), adminPassword, nil)

	request := httptest.NewRequest("GET", fmt.Sprintf("%s%s", baseURI, userURI), nil)

	response = doHTTP(request, response.Cookies())
	body, err := getBody(response)

	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", userURI, err)
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	// Now verify that the list of names matches our expectations.
	// First, split the return string into an array of lines, removing the possible trailing newline
	s := strings.TrimSuffix(string(body), "\n")
	names := strings.Split(s, "\n")

	// Next, verify the initial title line
	assert.Equal(t, "Users (List)", names[0])

	// .. and then verify that all following lines correctly consist of all the expected names
	match := knownNames
	match[baseURI+admin] = baseURI + admin

	// .. this involves converting the set of keys to an array for matching
	keys := make([]string, 0, len(match))
	for k := range match {
		keys = append(keys, k)
	}

	assert.ElementsMatchf(t, keys, names[1:], "elements did not match\nReturned Value: %s\nMatch Values: %v", s, keys)

	doLogout(t, randomCase(adminAccountName), response.Cookies())
}

func TestUsersListNoPriv(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	response := doLogin(t, randomCase(adminAccountName), adminPassword, nil)
	_, cookies := ensureAccount(t, "Alice", aliceDef, response.Cookies())
	response = doLogout(t, randomCase(adminAccountName), cookies)

	response = doLogin(t, "Alice", alicePassword, response.Cookies())

	request := httptest.NewRequest("GET", fmt.Sprintf("%s%s", baseURI, userURI), nil)

	response = doHTTP(request, response.Cookies())
	_, err := getBody(response)

	assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", userURI, err)
	assert.Equal(t, http.StatusForbidden, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	doLogout(t, "alice", response.Cookies())
}

// --- Known user list tests

// +++ Get user details tests

func TestUsersRead(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	response := doLogin(t, randomCase(adminAccountName), adminPassword, nil)

	request := httptest.NewRequest("GET", fmt.Sprintf("%s%s%s", baseURI, userURI, randomCase(adminAccountName)), nil)
	request.Header.Set("Content-Type", "application/json")

	response = doHTTP(request, response.Cookies())

	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	user := &pb.UserPublic{}
	err := getJSONBody(response, user)
	assert.Nilf(t, err, "Failed to convert body to valid json.  err: %v", err)

	assert.Equal(t, "application/json", strings.ToLower(response.Header.Get("Content-Type")))

	match, err := strconv.ParseInt(response.Header.Get("ETag"), 10, 64)
	assert.Nilf(t, err, "failed to convert the ETag to valid int64")
	assert.Less(t, int64(1), match)

	assert.True(t, user.Enabled)
	assert.True(t, user.CanManageAccounts)
	assert.True(t, user.NeverDelete)

	doLogout(t, randomCase(adminAccountName), response.Cookies())
}

func TestUsersReadUnknownUser(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	response := doLogin(t, randomCase(adminAccountName), adminPassword, nil)

	request := httptest.NewRequest("GET", fmt.Sprintf("%s%s%s", baseURI, userURI, "BadUser"), nil)
	request.Header.Set("Content-Type", "application/json")

	response = doHTTP(request, response.Cookies())
	body, err := getBody(response)
	t.Log(string(body))

	assert.Nilf(t, err, "Expected no error in getting body.  err=%v", err)
	assert.NotEqual(t, "application/json", strings.ToLower(response.Header.Get("Content-Type")))

	assert.Equal(t, http.StatusNotFound, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	doLogout(t, randomCase(adminAccountName), response.Cookies())
}

func TestUsersReadNoPriv(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	response := doLogin(t, randomCase(adminAccountName), adminPassword, nil)
	_, cookies := ensureAccount(t, "Alice", aliceDef, response.Cookies())
	response = doLogout(t, randomCase(adminAccountName), cookies)

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

func TestUsersOperationIllegal(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

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
		"CloudChamber: invalid user operation requested (?op=)\n", string(body),
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
		"CloudChamber: invalid user operation requested (?op=testInvalid)\n", string(body),
		"Handler returned unexpected response body: %v", string(body))
}

// --- User operation (?op=) tests

// +++ Update user tests

func TestUsersUpdate(t *testing.T) {
	aliceUpd := &pb.UserDefinition{
		Password:          alicePassword,
		Enabled:           true,
		CanManageAccounts: true,
	}

	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	response := doLogin(t, randomCase(adminAccountName), adminPassword, nil)

	r, err := toJSONReader(aliceUpd)
	assert.Nilf(t, err, "Failed to format UserDefinition, err = %v", err)

	rev, cookies := ensureAccount(t, "Alice", aliceDef, response.Cookies())
	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s", baseURI, alice), r)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("If-Match", fmt.Sprintf("%v", rev))

	response = doHTTP(request, cookies)
	assert.Equal(t, "application/json", strings.ToLower(response.Header.Get("Content-Type")))

	user := &pb.UserPublic{}
	err = getJSONBody(response, user)
	assert.Nilf(t, err, "Failed to convert body to valid json.  err: %v", err)

	match, err := strconv.ParseInt(response.Header.Get("ETag"), 10, 64)
	assert.Nilf(t, err, "failed to convert the ETag to valid int64")

	// Note: since ensureAccount() will attempt to re-use an existing account, all we know is
	// that by the time it returns there will be an account, and the returned revision is the
	// revision at the time the account was created, whether then, or earlier. Since for the
	// store, the revision is per-store, and NOT per-key, we cannot assume anything about the
	// exact relationship or "distance" between revisions that are not equal.
	//
	// So a "rev + 1" style test is not appropriate.
	//
	assert.Less(t, rev, match)

	assert.True(t, user.Enabled)
	assert.True(t, user.CanManageAccounts)
	assert.False(t, user.NeverDelete)

	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	doLogout(t, randomCase(adminAccountName), response.Cookies())
}

func TestUsersUpdateBadData(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	response := doLogin(t, randomCase(adminAccountName), adminPassword, nil)

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

	doLogout(t, randomCase(adminAccountName), response.Cookies())
}

func TestUsersUpdateBadMatch(t *testing.T) {
	aliceUpd := &pb.UserDefinition{
		Password:          alicePassword,
		Enabled:           true,
		CanManageAccounts: true,
	}

	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	response := doLogin(t, randomCase(adminAccountName), adminPassword, nil)

	r, err := toJSONReader(aliceUpd)
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

	doLogout(t, randomCase(adminAccountName), response.Cookies())
}

func TestUsersUpdateBadMatchSyntax(t *testing.T) {
	aliceUpd := &pb.UserDefinition{
		Password:          alicePassword,
		Enabled:           true,
		CanManageAccounts: true,
	}

	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	response := doLogin(t, randomCase(adminAccountName), adminPassword, nil)

	r, err := toJSONReader(aliceUpd)
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

	doLogout(t, randomCase(adminAccountName), response.Cookies())
}

func TestUsersUpdateNoUser(t *testing.T) {
	upd := &pb.UserDefinition{
		Password:          "bogus",
		Enabled:           true,
		CanManageAccounts: true,
	}

	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	response := doLogin(t, randomCase(adminAccountName), adminPassword, nil)

	r, err := toJSONReader(upd)
	assert.Nilf(t, err, "Failed to format UserDefinition, err = %v", err)

	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s", baseURI, userURI+"BadUser"), r)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("If-Match", fmt.Sprintf("%v", "1"))

	response = doHTTP(request, response.Cookies())
	body, err := getBody(response)
	t.Logf(string(body))
	assert.Nilf(t, err, "Error reading body, err = %v", err)

	assert.Equal(t, http.StatusNotFound, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	doLogout(t, randomCase(adminAccountName), response.Cookies())
}

func TestUsersUpdateNoPriv(t *testing.T) {
	aliceUpd := &pb.UserDefinition{
		Password:          alicePassword,
		Enabled:           true,
		CanManageAccounts: true,
	}

	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	response := doLogin(t, randomCase(adminAccountName), adminPassword, nil)
	_, cookies := ensureAccount(t, "Alice", aliceDef, response.Cookies())
	rev, cookies := ensureAccount(t, "Bob", bobDef, cookies)
	response = doLogout(t, randomCase(adminAccountName), cookies)

	response = doLogin(t, "Bob", bobPassword, response.Cookies())

	r, err := toJSONReader(aliceUpd)
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
	defer unit_test.SetTesting(nil)

	path := fmt.Sprintf("%s%s", baseURI, alice)

	response := doLogin(t, randomCase(adminAccountName), adminPassword, nil)

	_, cookies := ensureAccount(t, "Alice", aliceDef, response.Cookies())
	request := httptest.NewRequest("DELETE", path, nil)
	request.Header.Set("Content-Type", "application/json")

	response = doHTTP(request, cookies)
	body, err := getBody(response)

	assert.Nilf(t, err, "Unable to retrieve response body, err = %v", err)
	t.Log(string(body))

	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)
	delete(knownNames, "Alice")

	// Now verify the deletion by trying to get the user

	request = httptest.NewRequest("GET", path, nil)

	response = doHTTP(request, response.Cookies())
	body, err = getBody(response)

	assert.Nilf(t, err, "Unable to retrieve response body, err = %v", err)
	t.Log(string(body))

	assert.Equal(t, http.StatusNotFound, response.StatusCode, "Found deleted user %q", path)

	doLogout(t, randomCase(adminAccountName), response.Cookies())
}

func TestUsersDeleteNoUser(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	path := fmt.Sprintf("%s%s", baseURI, alice+"Bogus")

	response := doLogin(t, randomCase(adminAccountName), adminPassword, nil)

	request := httptest.NewRequest("DELETE", path, nil)
	request.Header.Set("Content-Type", "application/json")

	response = doHTTP(request, response.Cookies())
	body, err := getBody(response)

	assert.Nilf(t, err, "Unable to retrieve response body, err = %v", err)
	t.Log(string(body))

	assert.Equal(t, http.StatusNotFound, response.StatusCode, "Found deleted user %q", path)

	doLogout(t, randomCase(adminAccountName), response.Cookies())
}

func TestUsersDeleteNoPriv(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	response := doLogin(t, randomCase(adminAccountName), adminPassword, nil)
	_, cookies := ensureAccount(t, "Alice", aliceDef, response.Cookies())
	_, cookies = ensureAccount(t, "Bob", bobDef, cookies)
	response = doLogout(t, randomCase(adminAccountName), cookies)

	response = doLogin(t, "Bob", bobPassword, response.Cookies())

	request := httptest.NewRequest("DELETE", fmt.Sprintf("%s%s", baseURI, alice), nil)

	response = doHTTP(request, response.Cookies())
	body, err := getBody(response)

	assert.Nilf(t, err, "Unable to retrieve response body, err = %v", err)
	t.Logf(string(body))

	assert.Equal(t, http.StatusForbidden, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	doLogout(t, "bob", response.Cookies())
}

func TestUsersDeleteProtected(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	response := doLogin(t, randomCase(adminAccountName), adminPassword, nil)

	request := httptest.NewRequest("DELETE", fmt.Sprintf("%s%s", baseURI, admin), nil)

	response = doHTTP(request, response.Cookies())
	body, err := getBody(response)

	assert.Nilf(t, err, "Unable to retrieve response body, err = %v", err)
	assert.Equal(t, "CloudChamber: user \"admin\" is protected and cannot be deleted\n", string(body))
	assert.Equalf(t, http.StatusForbidden, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	doLogout(t, adminAccountName, response.Cookies())
}

// --- Delete user tests
