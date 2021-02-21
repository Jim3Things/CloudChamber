// Unit tests for the web service frontend.
//
// Borrows heavily from the gorilla mux test package.
//
package frontend

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Jim3Things/CloudChamber/simulation/pkg/errors"
	pb "github.com/Jim3Things/CloudChamber/simulation/pkg/protos/admin"
)

type UserTestSuite struct {
	testSuiteCore
}

func (ts *UserTestSuite) SetupSuite() {
	ts.testSuiteCore.SetupSuite()
}

func (ts *UserTestSuite) userRead(path string, cookies []*http.Cookie) (*http.Response, *pb.UserPublic) {
	assert := ts.Assert()

	request := httptest.NewRequest("GET", path, nil)
	request.Header.Set("Content-Type", "application/json")

	response := ts.doHTTP(request, cookies)

	assert.Equal(http.StatusOK, response.StatusCode)

	user := &pb.UserPublic{}
	err := ts.getJSONBody(response, user)
	assert.NoError(err, "Failed to convert body to valid json.  err: %v", err)

	assert.Equal("application/json", strings.ToLower(response.Header.Get("Content-Type")))

	return response, user
}

func (ts *UserTestSuite) setPassword(
	name string,
	upd *pb.UserPassword,
	rev int64,
	cookies []*http.Cookie) (*http.Response, int64) {
	assert := ts.Assert()

	r, err := ts.toJSONReader(upd)
	assert.NoError(err)

	path := ts.userPath() + name + "?password"

	request := httptest.NewRequest("PUT", path, r)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("If-Match", formatAsEtag(rev))

	response := ts.doHTTP(request, cookies)

	match, err := parseETag(response.Header.Get("ETag"))
	assert.NoError(err)

	return response, match
}

func (ts *UserTestSuite) userUpdate(
	path string,
	upd *pb.UserUpdate,
	rev int64,
	cookies []*http.Cookie) (*http.Response, int64) {
	assert := ts.Assert()

	r, err := ts.toJSONReader(upd)
	assert.NoError(err, "Failed to format UserUpdate, err = %v", err)

	request := httptest.NewRequest("PUT", path, r)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("If-Match", formatAsEtag(rev))

	response := ts.doHTTP(request, cookies)
	assert.Equal(http.StatusOK, response.StatusCode)

	tag, err := parseETag(response.Header.Get("ETag"))
	assert.NoError(err)

	return response, tag
}

// --- Helper functions

// The individual unit tests follow here.  They are grouped by the operation
// they are testing, starting with a simple happy path case, followed by
// repeating sequences (optionally), and then by failure cases.
//
// Each group is demarcated by comment lines.  The group starts with one that
// that has "+++", and ends with one that has "==="

// +++ Login tests

func (ts *UserTestSuite) TestLoginSessionSimple() {
	assert := ts.Assert()
	logf := ts.T().Logf
	log := ts.T().Log

	// login for the first time, should succeed
	request := httptest.NewRequest("PUT", fmt.Sprintf("%s?op=login", ts.admin()), strings.NewReader(ts.adminPassword()))
	response := ts.doHTTP(request, nil)
	body, err := ts.getBody(response)

	logf("[?op=login]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	log(string(body))

	assert.Equal(1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	// ... and logout, which should succeed
	//     (note that this also checks that the username match is case insensitive)
	//
	request = httptest.NewRequest("PUT", fmt.Sprintf("%s?op=logout", strings.ToUpper(ts.admin())), nil)
	response = ts.doHTTP(request, response.Cookies())
	body, err = ts.getBody(response)

	assert.NoError(
		err,
		"Failed to read body returned from call to handler for route %v: %v", ts.admin(), err)

	logf("[?op=logout]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	log(string(body))

	assert.Equal(1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)
}

func (ts *UserTestSuite) TestLoginSessionRepeat() {
	assert := ts.Assert()

	// login for the first time, should succeed
	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	assert.Equal(1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	// ... and logout, which should succeed
	//
	response = ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())

	assert.Equal(1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	// login for the second iteration, should succeed
	response = ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), response.Cookies())

	assert.Equal(1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	// ... and logout, which should succeed
	//
	response = ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())

	assert.Equal(1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)
}

func (ts *UserTestSuite) TestLoginDupLogins() {
	assert := ts.Assert()
	logf := ts.T().Logf

	// login for the first time, should succeed
	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	assert.Equal(1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	// now repeat the attempt to login again, which should fail
	request := httptest.NewRequest("PUT", fmt.Sprintf("%s?op=login", ts.admin()), strings.NewReader(ts.adminPassword()))
	response = ts.doHTTP(request, response.Cookies())
	body, err := ts.getBody(response)

	assert.NoError(err, "Failed to read body returned from call to handler for route %v: %v", ts.admin(), err)

	logf("[?op=login]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))

	assert.Equal(1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(http.StatusBadRequest, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)
	assert.Equal(
		fmt.Sprintf("%s\n", errors.ErrUserAlreadyLoggedIn.Error()), string(body),
		"Handler returned unexpected response body: %v", string(body))

	// .. and let's just try with another user, which should also fail
	request = httptest.NewRequest("PUT", fmt.Sprintf("%s?op=login", ts.bob()), strings.NewReader("test2"))
	response = ts.doHTTP(request, response.Cookies())
	body, err = ts.getBody(response)

	assert.NoError(err, "Failed to read body returned from call to handler for route %v: %v", ts.bob(), err)

	logf("[?op=login]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))

	assert.Equal(1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(http.StatusBadRequest, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)
	assert.Equal(
		fmt.Sprintf("%s\n", errors.ErrUserAlreadyLoggedIn.Error()), string(body),
		"Handler returned unexpected response body: %v", string(body))

	// ... and logout, which should succeed
	//
	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())

	assert.Equal(1, len(response.Cookies()), "Unexpected number of cookies found")
}

func (ts *UserTestSuite) TestLoginLogoutDiffAccounts() {
	assert := ts.Assert()
	logf := ts.T().Logf
	log := ts.T().Log

	// login for the first time, should succeed
	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	assert.Equal(1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	// ... next we need a second account that we're sure exists
	_, cookies := ts.ensureAccount(ts.aliceName(), ts.aliceDef, response.Cookies())

	// ... and now try to logout from it, which should not succeed
	//
	request := httptest.NewRequest("PUT", fmt.Sprintf("%s?op=logout", ts.alice()), nil)
	response = ts.doHTTP(request, cookies)
	body, err := ts.getBody(response)

	assert.NoError(err, "Failed to read body returned from call to handler for route %v: %v", ts.alice(), err)

	logf("[?op=logout]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	log(string(body))

	assert.Equal(1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(http.StatusBadRequest, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	// ... and logout, which should succeed
	//
	response = ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())

	assert.Equal(1, len(response.Cookies()), "Unexpected number of cookies found")
}

func (ts *UserTestSuite) TestDoubleLogout() {
	assert := ts.Assert()
	logf := ts.T().Logf
	log := ts.T().Log

	// login for the first time, should succeed
	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	assert.Equal(1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	// ... logout, which should succeed
	//
	response = ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())

	assert.Equal(1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	// ... logout again, which should fail
	//
	request := httptest.NewRequest("PUT", fmt.Sprintf("%s?op=logout", ts.admin()), nil)
	response = ts.doHTTP(request, response.Cookies())
	body, err := ts.getBody(response)

	assert.NoError(err, "Failed to read body returned from call to handler for route %v: %v", ts.admin(), err)

	logf("[?op=logout]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	log(string(body))

	assert.Equal(1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(http.StatusBadRequest, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)
}

func (ts *UserTestSuite) TestLoginSessionBadPassword() {
	assert := ts.Assert()
	logf := ts.T().Logf
	log := ts.T().Log

	// login for the first time, should succeed
	request := httptest.NewRequest(
		"PUT",
		fmt.Sprintf("%s?op=login", ts.admin()),
		strings.NewReader(ts.adminPassword()+"rubbish"))
	response := ts.doHTTP(request, nil)
	body, err := ts.getBody(response)

	assert.NoError(err, "Failed to read body returned from call to handler for route %v: %v", ts.admin(), err)

	logf("[?op=login]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	log(string(body))

	assert.Equal(1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(http.StatusForbidden, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	// Now just validate that there really isn't an active session here.
	response = ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), response.Cookies())

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *UserTestSuite) TestLoginSessionNoUser() {
	assert := ts.Assert()
	logf := ts.T().Logf
	log := ts.T().Log

	// login for the first time, the http call should succeed, but fail the login
	request := httptest.NewRequest(
		"PUT",
		fmt.Sprintf("%s%s?op=login", ts.admin(), "Bogus"),
		strings.NewReader(ts.adminPassword()))
	response := ts.doHTTP(request, nil)
	body, err := ts.getBody(response)

	assert.NoError(err, "Failed to read body returned from call to handler for route %v: %v", ts.admin(), err)

	logf("[?op=login]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	log(string(body))

	assert.Equal(1, len(response.Cookies()), "Unexpected number of cookies found")
	assert.Equal(http.StatusNotFound, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	// Now just validate that there really isn't an active session here.
	response = ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), response.Cookies())

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

// --- Login tests

// +++ User creation tests

func (ts *UserTestSuite) TestCreate() {
	assert := ts.Assert()
	logf := ts.T().Logf
	log := ts.T().Log

	path := ts.alice() + "2"

	r, err := ts.toJSONReader(ts.aliceDef)
	assert.NoError(err, "Failed to format UserDefinition, err = %v", err)

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	request := httptest.NewRequest("POST", path, r)
	request.Header.Set("Content-Type", "application/json")

	response = ts.doHTTP(request, response.Cookies())
	body, err := ts.getBody(response)

	assert.NoError(err, "Failed to read body returned from call to handler for route %v: %v", path, err)

	logf("[%s]: SC=%v, Content-Type='%v'\n", path, response.StatusCode, response.Header.Get("Content-Type"))
	log(string(body))

	assert.Equal(http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)
	assert.Equal(
		"User \"Alice2\" created, enabled: true, rights: ", string(body),
		"Handler returned unexpected response body: %v", string(body))

	ts.knownNames[path] = path
	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *UserTestSuite) TestCreateDup() {
	assert := ts.Assert()
	logf := ts.T().Logf
	log := ts.T().Log

	r, err := ts.toJSONReader(ts.aliceDef)
	assert.NoError(err, "Failed to format UserDefinition, err = %v", err)

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	_, cookies := ts.ensureAccount(ts.aliceName(), ts.aliceDef, response.Cookies())

	request := httptest.NewRequest("POST", ts.alice(), r)
	request.Header.Set("Content-Type", "application/json")

	response = ts.doHTTP(request, cookies)
	body, err := ts.getBody(response)

	assert.NoError(err, "Failed to read body returned from call to handler for route %v: %v", ts.userPath(), err)

	logf("[%s]: SC=%v, Content-Type='%v'\n", ts.userPath(), response.StatusCode, response.Header.Get("Content-Type"))
	log(string(body))

	assert.Equal(http.StatusBadRequest, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)
	assert.Equal(
		"CloudChamber: user \"Alice\" already exists\n", string(body),
		"Handler returned unexpected response body: %v", string(body))

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *UserTestSuite) TestCreateBadData() {
	assert := ts.Assert()
	logf := ts.T().Logf
	log := ts.T().Log

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	request := httptest.NewRequest(
		"POST",
		ts.alice()+"2",
		strings.NewReader("{\"enabled\":123,\"manageAccounts\":false, \"password\":\"test\"}"))
	request.Header.Set("Content-Type", "application/json")

	response = ts.doHTTP(request, response.Cookies())
	body, err := ts.getBody(response)

	assert.NoError(err, "Failed to read body returned from call to handler for route %v: %v", ts.userPath(), err)

	logf("[%s]: SC=%v, Content-Type='%v'\n", ts.userPath(), response.StatusCode, response.Header.Get("Content-Type"))
	log(string(body))

	assert.Equal(http.StatusBadRequest, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)
	assert.Equal(
		"json: cannot unmarshal number into Go value of type bool\n", string(body),
		"Handler returned unexpected response body: %v", string(body))

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *UserTestSuite) TestCreateNoPrivilege() {
	assert := ts.Assert()
	logf := ts.T().Logf
	log := ts.T().Log

	r, err := ts.toJSONReader(ts.bobDef)
	assert.NoError(err, "Failed to format UserDefinition, err = %v", err)

	response := ts.doLogin(ts.aliceName(), ts.alicePassword(), nil)

	request := httptest.NewRequest("POST", ts.bob(), r)
	request.Header.Set("Content-Type", "application/json")

	response = ts.doHTTP(request, response.Cookies())
	body, err := ts.getBody(response)

	assert.NoError(err, "Failed to read body returned from call to handler for route %v: %v", ts.userPath(), err)

	logf("[%s]: SC=%v, Content-Type='%v'\n", ts.userPath(), response.StatusCode, response.Header.Get("Content-Type"))
	log(string(body))

	assert.Equal(http.StatusForbidden, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)
	assert.Equal(
		"CloudChamber: permission denied\n", string(body),
		"Handler returned unexpected response body: %v", string(body))

	ts.doLogout(ts.aliceName(), response.Cookies())
}

func (ts *UserTestSuite) TestCreateNoSession() {
	assert := ts.Assert()
	logf := ts.T().Logf
	log := ts.T().Log

	path := ts.alice() + "2"

	r, err := ts.toJSONReader(ts.aliceDef)
	assert.NoError(err, "Failed to format UserDefinition, err = %v", err)

	request := httptest.NewRequest("POST", path, r)
	request.Header.Set("Content-Type", "application/json")

	response := ts.doHTTP(request, nil)
	body, err := ts.getBody(response)

	assert.NoError(err, "Failed to read body returned from call to handler for route %v: %v", ts.userPath(), err)

	logf("[%s]: SC=%v, Content-Type='%v'\n", path, response.StatusCode, response.Header.Get("Content-Type"))
	log(string(body))

	assert.Equal(http.StatusForbidden, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)
	assert.Equal(
		"CloudChamber: permission denied\n", string(body),
		"Handler returned unexpected response body: %v", string(body))
}

// --- User creation tests

// +++ Known users list tests

func (ts *UserTestSuite) TestList() {
	assert := ts.Assert()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	request := httptest.NewRequest("GET", ts.userPath(), nil)
	request.Header.Set("Content-Type", "application/json")

	response = ts.doHTTP(request, response.Cookies())
	users := &pb.UserList{}
	err := ts.getJSONBody(response, users)
	assert.NoError(err, "Failed to convert body to valid json.  err: %v", err)

	assert.Equal("application/json", strings.ToLower(response.Header.Get("Content-Type")))
	assert.Equal(http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	// Now verify that the list of names matches our expectations.
	// First, form an array of names from the returned structure
	addresses := make([]string, 0, len(users.Users))
	for _, entry := range users.Users {
		assert.True(strings.HasSuffix(entry.Uri, entry.Name))
		if strings.EqualFold(entry.Name, ts.adminAccountName()) {
			assert.True(entry.Protected)
		}

		addresses = append(addresses, entry.Uri)
	}

	// .. and then verify that all following lines correctly consist of all the expected names
	match := ts.knownNames
	match[ts.admin()] = ts.admin()

	// .. this involves converting the set of keys to an array for matching
	keys := make([]string, 0, len(match))
	for k := range match {
		keys = append(keys, k)
	}

	assert.ElementsMatchf(keys, addresses, "elements did not match\nReturned Value: %s\nMatch Values: %v", addresses, keys)

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *UserTestSuite) TestListNoPrivilege() {
	assert := ts.Assert()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)
	_, cookies := ts.ensureAccount(ts.aliceName(), ts.aliceDef, response.Cookies())
	response = ts.doLogout(ts.randomCase(ts.adminAccountName()), cookies)

	response = ts.doLogin(ts.aliceName(), ts.alicePassword(), response.Cookies())

	request := httptest.NewRequest("GET", ts.userPath(), nil)

	response = ts.doHTTP(request, response.Cookies())
	_, err := ts.getBody(response)

	assert.NoError(err, "Failed to read body returned from call to handler for route %v: %v", ts.userPath(), err)
	assert.Equal(http.StatusForbidden, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	ts.doLogout("alice", response.Cookies())
}

// --- Known user list tests

// +++ Get user details tests

func (ts *UserTestSuite) TestRead() {
	assert := ts.Assert()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	request := httptest.NewRequest("GET", fmt.Sprintf("%s%s", ts.userPath(), ts.randomCase(ts.adminAccountName())), nil)
	request.Header.Set("Content-Type", "application/json")

	response = ts.doHTTP(request, response.Cookies())

	assert.Equal(http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	user := &pb.UserPublic{}
	err := ts.getJSONBody(response, user)
	assert.NoError(err, "Failed to convert body to valid json.  err: %v", err)

	assert.Equal("application/json", strings.ToLower(response.Header.Get("Content-Type")))

	match, err := parseETag(response.Header.Get("ETag"))
	assert.NoError(err, "failed to convert the ETag to valid int64")
	assert.Less(int64(1), match)

	assert.True(user.Enabled)
	assert.True(user.Rights.CanManageAccounts)
	assert.True(user.NeverDelete)

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *UserTestSuite) TestReadUnknownUser() {
	assert := ts.Assert()
	log := ts.T().Log

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	request := httptest.NewRequest("GET", fmt.Sprintf("%s%s", ts.userPath(), "BadUser"), nil)
	request.Header.Set("Content-Type", "application/json")

	response = ts.doHTTP(request, response.Cookies())
	body, err := ts.getBody(response)
	log(string(body))

	assert.NoError(err, "Expected no error in getting body.  err=%v", err)
	assert.NotEqual("application/json", strings.ToLower(response.Header.Get("Content-Type")))

	assert.Equal(http.StatusNotFound, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *UserTestSuite) TestReadNoPrivilege() {
	assert := ts.Assert()
	log := ts.T().Log

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)
	_, cookies := ts.ensureAccount(ts.aliceName(), ts.aliceDef, response.Cookies())
	response = ts.doLogout(ts.randomCase(ts.adminAccountName()), cookies)

	response = ts.doLogin(ts.aliceName(), ts.alicePassword(), response.Cookies())

	request := httptest.NewRequest("GET", fmt.Sprintf("%s%s", ts.userPath(), "BadUser"), nil)
	request.Header.Set("Content-Type", "application/json")

	response = ts.doHTTP(request, response.Cookies())
	body, err := ts.getBody(response)
	log(string(body))

	assert.NoError(err, "Expected no error in getting body.  err=%v", err)
	assert.NotEqual("application/json", strings.ToLower(response.Header.Get("Content-Type")))

	assert.Equal(http.StatusForbidden, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	ts.doLogout("alice", response.Cookies())
}

// --- Get user details tests

// +++ User operation (?op=) tests

func (ts *UserTestSuite) TestOperationIllegal() {
	assert := ts.Assert()
	logf := ts.T().Logf
	log := ts.T().Log

	// Verify a bunch of failure cases. Specifically,
	// - that a naked op fails
	// - that an invalid op fails

	// Case 1, check that naked op fails
	//
	request := httptest.NewRequest("PUT", fmt.Sprintf("%s?op", ts.alice()), nil)
	response := ts.doHTTP(request, nil)

	response = ts.doHTTP(request, response.Cookies())
	body, err := ts.getBody(response)

	assert.NoError(err, "Failed to read body returned from call to handler for route %v: %v", ts.alice(), err)

	logf("[?op]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	log(string(body))

	assert.Equal(http.StatusBadRequest, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)
	assert.Equal(
		"CloudChamber: invalid user operation requested (?op=) for user \"alice\"\n", string(body),
		"Handler returned unexpected response body: %v", string(body))

	// Case 2, check that an invalid op fails
	//
	request = httptest.NewRequest("PUT", fmt.Sprintf("%s?op=testInvalid", ts.alice()), nil)
	response = ts.doHTTP(request, response.Cookies())
	body, err = ts.getBody(response)

	assert.NoError(err, "Failed to read body returned from call to handler for route %v: %v", ts.alice(), err)

	logf("[?op=testInvalid]: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	log(string(body))

	assert.Equal(http.StatusBadRequest, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)
	assert.Equal(
		"CloudChamber: invalid user operation requested (?op=testInvalid) for user \"alice\"\n", string(body),
		"Handler returned unexpected response body: %v", string(body))
}

// --- User operation (?op=) tests

// +++ Update user tests

func (ts *UserTestSuite) TestUpdateSuccess() {
	assert := ts.Assert()

	aliceUpd := &pb.UserUpdate{
		Enabled: true,
		Rights: &pb.Rights{
			CanManageAccounts:  true,
			CanStepTime:        false,
			CanModifyWorkloads: false,
			CanModifyInventory: false,
			CanInjectFaults:    false,
			CanPerformRepairs:  false,
		},
	}

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	r, err := ts.toJSONReader(aliceUpd)
	assert.NoError(err, "Failed to format UserDefinition, err = %v", err)

	rev, cookies := ts.ensureAccount(ts.aliceName(), ts.aliceDef, response.Cookies())
	request := httptest.NewRequest("PUT", ts.alice(), r)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("If-Match", formatAsEtag(rev))

	response = ts.doHTTP(request, cookies)
	assert.Equal("application/json", strings.ToLower(response.Header.Get("Content-Type")))

	user := &pb.UserPublic{}
	err = ts.getJSONBody(response, user)
	assert.NoError(err, "Failed to convert body to valid json.  err: %v", err)

	match, err := parseETag(response.Header.Get("ETag"))
	assert.NoError(err, "failed to convert the ETag to valid int64")

	// Note: since ensureAccount() will attempt to re-use an existing account, all we know is
	// that by the time it returns there will be an account, and the returned revision is the
	// revision at the time the account was created, whether then, or earlier. Since for the
	// store, the revision is per-store, and NOT per-key, we cannot assume anything about the
	// exact relationship or "distance" between revisions that are not equal.
	//
	// So a "rev + 1" style test is not appropriate.
	//
	assert.Less(rev, match)

	assert.True(user.Enabled)
	assert.Equal(aliceUpd.Rights, user.Rights)
	assert.False(user.NeverDelete)

	assert.Equal(http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *UserTestSuite) TestUpdateBadData() {
	assert := ts.Assert()
	logf := ts.T().Logf
	log := ts.T().Log

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	rev, cookies := ts.ensureAccount(ts.aliceName(), ts.aliceDef, response.Cookies())

	request := httptest.NewRequest(
		"PUT",
		ts.alice(),
		strings.NewReader("{\"enabled\":123,\"manageAccounts\":false, \"password\":\"test\"}"))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("If-Match", formatAsEtag(rev))

	response = ts.doHTTP(request, cookies)
	body, err := ts.getBody(response)

	assert.NoError(err, "Failed to read body returned from call to handler for route %v: %v", ts.userPath(), err)

	logf("[%s]: SC=%v, Content-Type='%v'\n", ts.userPath(), response.StatusCode, response.Header.Get("Content-Type"))
	log(string(body))

	assert.Equal(http.StatusBadRequest, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)
	assert.Equal(
		"json: cannot unmarshal number into Go value of type bool\n", string(body),
		"Handler returned unexpected response body: %v", string(body))

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *UserTestSuite) TestUpdateBadMatch() {
	assert := ts.Assert()
	log := ts.T().Log

	aliceUpd := &pb.UserUpdate{
		Enabled: true,
		Rights:  &pb.Rights{CanManageAccounts: true},
	}

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	r, err := ts.toJSONReader(aliceUpd)
	assert.NoError(err, "Failed to format UserDefinition, err = %v", err)

	rev, cookies := ts.ensureAccount(ts.aliceName(), ts.aliceDef, response.Cookies())
	request := httptest.NewRequest("PUT", ts.alice(), r)

	// Poison the revision
	rev += 10

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("If-Match", formatAsEtag(rev))

	response = ts.doHTTP(request, cookies)
	body, err := ts.getBody(response)

	log(string(body))
	assert.NoError(err, "Failed to get response body.  err: %v", err)

	assert.Equal(http.StatusConflict, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *UserTestSuite) TestUpdateBadMatchSyntax() {
	assert := ts.Assert()
	log := ts.T().Log

	aliceUpd := &pb.UserDefinition{
		Password: ts.alicePassword(),
		Enabled:  true,
		Rights:   &pb.Rights{CanManageAccounts: true},
	}

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	r, err := ts.toJSONReader(aliceUpd)
	assert.NoError(err, "Failed to format UserDefinition, err = %v", err)

	_, cookies := ts.ensureAccount(ts.aliceName(), ts.aliceDef, response.Cookies())
	request := httptest.NewRequest("PUT", ts.alice(), r)

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("If-Match", "\"abc\"")

	response = ts.doHTTP(request, cookies)
	body, err := ts.getBody(response)

	log(string(body))
	assert.NoError(err, "Failed to get response body.  err: %v", err)

	assert.Equal(http.StatusBadRequest, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *UserTestSuite) TestUpdateNoUser() {
	assert := ts.Assert()
	logf := ts.T().Logf

	upd := &pb.UserUpdate{
		Enabled: true,
		Rights:  &pb.Rights{CanManageAccounts: true},
	}

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	r, err := ts.toJSONReader(upd)
	assert.NoError(err, "Failed to format UserDefinition, err = %v", err)

	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s", ts.userPath(), "BadUser"), r)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("If-Match", formatAsEtag(1))

	response = ts.doHTTP(request, response.Cookies())
	body, err := ts.getBody(response)
	logf(string(body))
	assert.NoError(err, "Error reading body, err = %v", err)

	assert.Equal(http.StatusNotFound, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *UserTestSuite) TestUpdateNoPrivilege() {
	assert := ts.Assert()
	logf := ts.T().Logf

	aliceUpd := &pb.UserUpdate{
		Enabled: true,
		Rights:  &pb.Rights{CanManageAccounts: true},
	}

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)
	_, cookies := ts.ensureAccount(ts.aliceName(), ts.aliceDef, response.Cookies())
	rev, cookies := ts.ensureAccount("Bob", ts.bobDef, cookies)
	response = ts.doLogout(ts.randomCase(ts.adminAccountName()), cookies)

	response = ts.doLogin("Bob", ts.bobPassword(), response.Cookies())

	r, err := ts.toJSONReader(aliceUpd)
	assert.NoError(err, "Failed to format UserDefinition, err = %v", err)

	request := httptest.NewRequest("PUT", ts.alice(), r)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("If-Match", formatAsEtag(rev))

	response = ts.doHTTP(request, response.Cookies())
	body, err := ts.getBody(response)

	logf(string(body))

	assert.Equal(http.StatusForbidden, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	ts.doLogout("bob", response.Cookies())
}

func (ts *UserTestSuite) TestUpdateExpandRights() {
	assert := ts.Assert()

	aliceUpd := &pb.UserUpdate{
		Enabled: true,
		Rights: &pb.Rights{
			CanStepTime: true,
		},
	}

	aliceOriginal := &pb.UserUpdate{
		Enabled: true,
		Rights: &pb.Rights{
			CanManageAccounts:  false,
			CanModifyInventory: true,
		},
	}

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)
	rev, cookies := ts.ensureAccount(ts.aliceName(), ts.aliceDef, response.Cookies())

	response, rev = ts.userUpdate(ts.alice(), aliceOriginal, rev, cookies)
	_, err := ts.getBody(response)
	assert.NoError(err)

	response = ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())

	response = ts.doLogin(ts.aliceName(), ts.alicePassword(), response.Cookies())

	r, err := ts.toJSONReader(aliceUpd)
	assert.NoError(err, "Failed to format UserUpdate, err = %v", err)

	request := httptest.NewRequest("PUT", ts.alice(), r)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("If-Match", formatAsEtag(rev))

	response = ts.doHTTP(request, response.Cookies())
	body, err := ts.getBody(response)
	assert.Equal("CloudChamber: permission denied\n", string(body))

	assert.Equal(http.StatusForbidden, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	// Now verify that the entry has not been changed
	response, user := ts.userRead(ts.alice(), response.Cookies())
	assert.Equal(aliceOriginal.Rights, user.Rights)
	assert.True(user.Enabled)
	assert.False(user.NeverDelete)

	ts.doLogout(ts.aliceName(), response.Cookies())
}

// --- Update user tests

// +++ Delete user tests

func (ts *UserTestSuite) TestDelete() {
	assert := ts.Assert()
	log := ts.T().Log

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	_, cookies := ts.ensureAccount(ts.aliceName(), ts.aliceDef, response.Cookies())
	request := httptest.NewRequest("DELETE", ts.alice(), nil)
	request.Header.Set("Content-Type", "application/json")

	response = ts.doHTTP(request, cookies)
	body, err := ts.getBody(response)

	assert.NoError(err, "Unable to retrieve response body, err = %v", err)
	log(string(body))

	assert.Equal(http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)
	delete(ts.knownNames, ts.aliceName())

	// Now verify the deletion by trying to get the user

	request = httptest.NewRequest("GET", ts.alice(), nil)

	response = ts.doHTTP(request, response.Cookies())
	body, err = ts.getBody(response)

	assert.NoError(err, "Unable to retrieve response body, err = %v", err)
	log(string(body))

	assert.Equal(http.StatusNotFound, response.StatusCode, "Found deleted user %q", ts.alice())

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *UserTestSuite) TestDeleteNoUser() {
	assert := ts.Assert()
	log := ts.T().Log

	path := ts.alice() + "Bogus"

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	request := httptest.NewRequest("DELETE", path, nil)
	request.Header.Set("Content-Type", "application/json")

	response = ts.doHTTP(request, response.Cookies())
	body, err := ts.getBody(response)

	assert.NoError(err, "Unable to retrieve response body, err = %v", err)
	log(string(body))

	assert.Equal(http.StatusNotFound, response.StatusCode, "Found deleted user %q", path)

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *UserTestSuite) TestDeleteNoPrivilege() {
	assert := ts.Assert()
	log := ts.T().Log

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)
	_, cookies := ts.ensureAccount(ts.aliceName(), ts.aliceDef, response.Cookies())
	_, cookies = ts.ensureAccount("Bob", ts.bobDef, cookies)
	response = ts.doLogout(ts.randomCase(ts.adminAccountName()), cookies)

	response = ts.doLogin("Bob", ts.bobPassword(), response.Cookies())

	request := httptest.NewRequest("DELETE", ts.alice(), nil)

	response = ts.doHTTP(request, response.Cookies())
	body, err := ts.getBody(response)

	assert.NoError(err, "Unable to retrieve response body, err = %v", err)
	log(string(body))

	assert.Equal(http.StatusForbidden, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	ts.doLogout("bob", response.Cookies())
}

func (ts *UserTestSuite) TestDeleteProtected() {
	assert := ts.Assert()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	request := httptest.NewRequest("DELETE", ts.admin(), nil)

	response = ts.doHTTP(request, response.Cookies())
	body, err := ts.getBody(response)

	assert.NoError(err, "Unable to retrieve response body, err = %v", err)
	assert.Equal("CloudChamber: user \"admin\" is protected and cannot be deleted\n", string(body))
	assert.Equalf(http.StatusForbidden, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	ts.doLogout(ts.adminAccountName(), response.Cookies())
}

// --- Delete user tests

// +++ SetPassword user tests

func (ts *UserTestSuite) TestSetPassword() {
	assert := ts.Assert()

	aliceNewPassword := ts.alicePassword() + "xxx"

	aliceUpd := &pb.UserPassword{
		OldPassword: ts.alicePassword(),
		NewPassword: aliceNewPassword,
		Force:       false,
	}

	aliceRevert := &pb.UserPassword{
		OldPassword: aliceNewPassword,
		NewPassword: ts.alicePassword(),
		Force:       true,
	}

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	rev, cookies := ts.ensureAccount(ts.aliceName(), ts.aliceDef, response.Cookies())
	response, match := ts.setPassword(ts.aliceName(), aliceUpd, rev, cookies)
	assert.Equal(http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	// Note: since ensureAccount() will attempt to re-use an existing account, all we know is
	// that by the time it returns there will be an account, and the returned revision is the
	// revision at the time the account was created, whether then, or earlier. Since for the
	// store, the revision is per-store, and NOT per-key, we cannot assume anything about the
	// exact relationship or "distance" between revisions that are not equal.
	//
	// So a "rev + 1" style test is not appropriate.
	//
	assert.Less(rev, match)

	response = ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())

	// Now verify that the password was changed, by trying to log in again
	response = ts.doLogin(ts.aliceName(), aliceNewPassword, response.Cookies())
	assert.Equal(http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	// Now set the password back
	response, _ = ts.setPassword(ts.aliceName(), aliceRevert, match, response.Cookies())
	assert.Equal(http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	ts.doLogout(ts.aliceName(), response.Cookies())
}

func (ts *UserTestSuite) TestSetPasswordForce() {
	assert := ts.Assert()

	aliceNewPassword := ts.alicePassword() + "xxx"

	aliceUpd := &pb.UserPassword{
		OldPassword: "bogus",
		NewPassword: aliceNewPassword,
		Force:       true,
	}

	aliceRevert := &pb.UserPassword{
		OldPassword: aliceNewPassword,
		NewPassword: ts.alicePassword(),
		Force:       true,
	}

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	rev, cookies := ts.ensureAccount(ts.aliceName(), ts.aliceDef, response.Cookies())
	response, match := ts.setPassword(ts.aliceName(), aliceUpd, rev, cookies)
	assert.Equal(http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	// Note: since ensureAccount() will attempt to re-use an existing account, all we know is
	// that by the time it returns there will be an account, and the returned revision is the
	// revision at the time the account was created, whether then, or earlier. Since for the
	// store, the revision is per-store, and NOT per-key, we cannot assume anything about the
	// exact relationship or "distance" between revisions that are not equal.
	//
	// So a "rev + 1" style test is not appropriate.
	//
	assert.Less(rev, match)

	response = ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())

	// Now verify that the password was changed, by trying to log in again
	response = ts.doLogin(ts.aliceName(), aliceNewPassword, response.Cookies())
	assert.Equal(http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	// Now set the password back
	response, _ = ts.setPassword(ts.aliceName(), aliceRevert, match, response.Cookies())
	assert.Equal(http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	ts.doLogout(ts.aliceName(), response.Cookies())
}

func (ts *UserTestSuite) TestSetPasswordBadPassword() {
	assert := ts.Assert()

	aliceNewPassword := ts.alicePassword() + "xxx"

	aliceUpd := &pb.UserPassword{
		OldPassword: "bogus",
		NewPassword: aliceNewPassword,
		Force:       false,
	}

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	rev, cookies := ts.ensureAccount(ts.aliceName(), ts.aliceDef, response.Cookies())

	r, err := ts.toJSONReader(aliceUpd)
	assert.NoError(err, "Failed to format UserPassword, err = %v", err)

	path := ts.alice() + "?password"

	request := httptest.NewRequest("PUT", path, r)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("If-Match", formatAsEtag(rev))

	response = ts.doHTTP(request, cookies)

	assert.Equal(http.StatusForbidden, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	response = ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())

	// Now verify that the password was not changed, by trying to log in again
	response = ts.doLogin(ts.aliceName(), ts.alicePassword(), response.Cookies())
	assert.Equal(http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	ts.doLogout(ts.aliceName(), response.Cookies())
}

func (ts *UserTestSuite) TestSetPasswordNoPrivilege() {
	assert := ts.Assert()

	adminNewPassword := ts.adminPassword() + "xxx"

	adminUpd := &pb.UserPassword{
		OldPassword: ts.adminPassword(),
		NewPassword: adminNewPassword,
		Force:       false,
	}

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)
	_, cookies := ts.ensureAccount(ts.aliceName(), ts.aliceDef, response.Cookies())
	response = ts.doLogout("Admin", cookies)

	response = ts.doLogin(ts.aliceName(), ts.alicePassword(), response.Cookies())

	r, err := ts.toJSONReader(adminUpd)
	assert.NoError(err, "Failed to format UserPassword, err = %v", err)

	path := ts.admin() + "?password"

	request := httptest.NewRequest("PUT", path, r)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("If-Match", formatAsEtag(-1))

	response = ts.doHTTP(request, response.Cookies())

	assert.Equal(http.StatusForbidden, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	response = ts.doLogout(ts.aliceName(), response.Cookies())

	// Now verify that the password was not changed, by trying to log in again
	response = ts.doLogin("Admin", ts.adminPassword(), response.Cookies())
	assert.Equal(http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	ts.doLogout("Admin", response.Cookies())
}

// --- SetPassword user tests

func (ts *UserTestSuite) TestSetRights() {
	assert := ts.Assert()
	require := ts.Require()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)
	_, cookies := ts.ensureAccount(ts.aliceName(), ts.aliceDef, response.Cookies())

	response, user := ts.userRead(ts.alice(), cookies)

	rev, err := parseETag(response.Header.Get("ETag"))
	require.NoError(err)

	newRights := &pb.Rights{
		CanManageAccounts:  true,
		CanStepTime:        false,
		CanModifyWorkloads: true,
		CanModifyInventory: false,
		CanInjectFaults:    true,
		CanPerformRepairs:  false,
	}

	upd := &pb.UserUpdate{
		Enabled: true,
		Rights:  newRights,
	}

	response, match := ts.userUpdate(ts.alice(), upd, rev, response.Cookies())

	user = &pb.UserPublic{}
	err = ts.getJSONBody(response, user)
	assert.NoError(err)

	require.Less(rev, match)
	require.Equal(newRights, user.Rights)

	rev = match

	newRights = &pb.Rights{
		CanManageAccounts:  false,
		CanStepTime:        true,
		CanModifyWorkloads: false,
		CanModifyInventory: true,
		CanInjectFaults:    false,
		CanPerformRepairs:  true,
	}

	upd.Rights = newRights

	response, match = ts.userUpdate(ts.alice(), upd, rev, response.Cookies())

	user = &pb.UserPublic{}
	err = ts.getJSONBody(response, user)
	assert.NoError(err)

	require.Less(rev, match)
	require.Equal(newRights, user.Rights)

	ts.doLogout(ts.adminAccountName(), response.Cookies())
}

func TestUserTestSuite(t *testing.T) {
	suite.Run(t, new(UserTestSuite))
}
